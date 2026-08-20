#!/usr/bin/env bash
set -euo pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
test_database="$project_dir/backend/data/hrms.local-test.db"
backend_log="$project_dir/backend/.logs/local-api.log"
frontend_log="$project_dir/frontend/.logs/local-web.log"
backend_pid=""
frontend_pid=""

usage() {
  cat <<'EOF'
Usage: ./scripts/local-dev.sh [--reset]

Starts the local API on http://127.0.0.1:8080 and the web app on
http://127.0.0.1:5173. The local test account is admin / 123456.

--reset  Deletes only backend/data/hrms.local-test.db before starting.
EOF
}

cleanup() {
  if [[ -n "$frontend_pid" ]]; then
    kill "$frontend_pid" 2>/dev/null || true
    wait "$frontend_pid" 2>/dev/null || true
  fi
  if [[ -n "$backend_pid" ]]; then
    kill "$backend_pid" 2>/dev/null || true
    wait "$backend_pid" 2>/dev/null || true
  fi
}

if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
  usage
  exit 0
fi

if [[ "${1:-}" == "--reset" ]]; then
  rm -f "$test_database" "$test_database-wal" "$test_database-shm"
elif [[ $# -ne 0 ]]; then
  usage >&2
  exit 2
fi

command -v go >/dev/null || { echo "Go 1.26+ is required." >&2; exit 1; }
command -v npm >/dev/null || { echo "Node.js and npm are required." >&2; exit 1; }
command -v curl >/dev/null || { echo "curl is required." >&2; exit 1; }

mkdir -p "$(dirname "$test_database")" "$(dirname "$backend_log")" "$(dirname "$frontend_log")"

if [[ ! -d "$project_dir/frontend/node_modules" ]]; then
  npm --prefix "$project_dir/frontend" ci
fi

trap cleanup EXIT INT TERM

(
  cd "$project_dir/backend"
  APP_ENV=development \
  HTTP_ADDR=127.0.0.1:8080 \
  DATABASE_DRIVER=sqlite \
  DATABASE_DSN=./data/hrms.local-test.db \
  JWT_SECRET=local-test-only-jwt-secret \
  JWT_TTL_HOURS=8 \
  CORS_ALLOWED_ORIGINS=http://127.0.0.1:5173,http://localhost:5173 \
  INITIAL_ADMIN_USERNAME=admin \
  INITIAL_ADMIN_PASSWORD=123456 \
  DAILY_REMINDER_TOKEN=local-test-only-reminder-token \
  TIMEZONE=Asia/Shanghai \
  go run ./cmd/api
) >"$backend_log" 2>&1 &
backend_pid=$!

for _ in $(seq 1 80); do
  if curl -fsS http://127.0.0.1:8080/healthz >/dev/null; then
    break
  fi
  if ! kill -0 "$backend_pid" 2>/dev/null; then
    cat "$backend_log" >&2
    exit 1
  fi
  sleep 0.25
done

if ! curl -fsS http://127.0.0.1:8080/healthz >/dev/null; then
  echo "Local API did not become ready. See $backend_log" >&2
  exit 1
fi

(
  cd "$project_dir/frontend"
  VITE_API_BASE_URL=http://127.0.0.1:8080/api/v1 \
  npm run dev -- --host 127.0.0.1 --port 5173 --strictPort
) >"$frontend_log" 2>&1 &
frontend_pid=$!

echo "Local test environment is running: http://127.0.0.1:5173"
echo "Login: admin / 123456"
echo "Press Ctrl-C to stop both services."
wait "$frontend_pid"
