#!/usr/bin/env bash
set -euo pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
env_id="${CLOUDBASE_ENV_ID:-chrms-d9gywgbw57877b4fb}"
service_name="${CLOUDBASE_SERVICE_NAME:-construction-hrms-api}"
api_origin="${CLOUDBASE_API_ORIGIN:-https://construction-hrms-api-298690-10-1416107181.sh.run.tcloudbase.com}"
hosting_origin="${CLOUDBASE_HOSTING_ORIGIN:-https://chrms-d9gywgbw57877b4fb-1416107181.tcloudbaseapp.com}"
mode="release"
release_dir=""

usage() {
  cat <<'EOF'
Usage: ./scripts/release-cloudbase.sh [--dry-run]

Builds and verifies the API and frontend, creates a minimal Cloud Run source
directory, and deploys the API and static Hosting assets to CloudBase.

Environment overrides:
  CLOUDBASE_ENV_ID
  CLOUDBASE_SERVICE_NAME
  CLOUDBASE_API_ORIGIN
  CLOUDBASE_HOSTING_ORIGIN

--dry-run  Run all local checks and create the release directory, but do not
           log in to CloudBase or change any remote resource.
EOF
}

cleanup() {
  if [[ -n "$release_dir" && -d "$release_dir" ]]; then
    rm -rf "$release_dir"
  fi
}

fail_if_release_contains_local_state() {
  local forbidden
  forbidden="$(find "$release_dir" \( \
    -type d \( -name .cache -o -name node_modules -o -name dist -o -name .logs -o -name data -o -name storage -o -name uploads \) -o \
    -type f \( -name '.env' -o -name '.env.*' -o -name '*.db' -o -name '*.sqlite' -o -name '*.sqlite-wal' -o -name '*.sqlite-shm' -o -name api -o -name hrms-api \) \
  \) -print -quit)"
  if [[ -n "$forbidden" ]]; then
    echo "Refusing deployment: local runtime file found in release directory: ${forbidden#$release_dir/}" >&2
    exit 1
  fi
}

if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
  usage
  exit 0
fi
if [[ "${1:-}" == "--dry-run" ]]; then
  mode="dry-run"
elif [[ $# -ne 0 ]]; then
  usage >&2
  exit 2
fi

command -v go >/dev/null || { echo "Go 1.26+ is required." >&2; exit 1; }
command -v npm >/dev/null || { echo "Node.js and npm are required." >&2; exit 1; }
command -v rsync >/dev/null || { echo "rsync is required." >&2; exit 1; }
command -v curl >/dev/null || { echo "curl is required." >&2; exit 1; }
if [[ "$mode" == "release" ]]; then
  command -v tcb >/dev/null || { echo "CloudBase CLI (tcb) is required." >&2; exit 1; }
fi

trap cleanup EXIT INT TERM

echo "Running release quality gates..."
(
  cd "$project_dir/backend"
  go test ./...
  go vet ./...
  go build -o /dev/null ./cmd/api
)
npm --prefix "$project_dir/frontend" ci
npm --prefix "$project_dir/frontend" run typecheck
VITE_API_BASE_URL="$api_origin/api/v1" npm --prefix "$project_dir/frontend" run build
git -C "$project_dir" diff --check

release_dir="$(mktemp -d "$project_dir/.cloudbase-release.XXXXXX")"
mkdir -p "$release_dir/backend"
rsync -a \
  --exclude '/.env' \
  --exclude '/.env.*' \
  --exclude '/.cache/' \
  --exclude '/.logs/' \
  --exclude '/data/' \
  --exclude '/storage/' \
  --exclude '/uploads/' \
  --exclude '/api' \
  --exclude '/hrms-api' \
  --exclude '*.db' \
  --exclude '*.sqlite' \
  --exclude '*.sqlite-shm' \
  --exclude '*.sqlite-wal' \
  "$project_dir/backend/" "$release_dir/backend/"
cp "$project_dir/Dockerfile" "$project_dir/.dockerignore" "$release_dir/"

fail_if_release_contains_local_state
test -f "$release_dir/backend/cmd/api/main.go" || { echo "Release source is missing backend/cmd/api/main.go" >&2; exit 1; }

echo "Release source prepared ($(du -sh "$release_dir" | awk '{print $1}'))."
if [[ "$mode" == "dry-run" ]]; then
  echo "Dry run passed. No CloudBase resource was changed."
  exit 0
fi

echo "Authenticating for CloudBase environment $env_id..."
tcb login

echo "Checking CloudBase Hosting..."
tcb hosting detail --env-id "$env_id" >/dev/null

echo "Deploying Cloud Run service $service_name..."
tcb cloudrun deploy \
  --env-id "$env_id" \
  --service-name "$service_name" \
  --port 8080 \
  --source "$release_dir" \
  --wait \
  --force

echo "Deploying frontend assets..."
tcb hosting deploy "$project_dir/frontend/dist" --env-id "$env_id"

echo "Verifying public endpoints..."
curl -fsS -o /dev/null "$api_origin/healthz"
curl -fsS -o /dev/null "$hosting_origin/"
tcb cloudrun list --env-id "$env_id" --service-name "$service_name" >/dev/null

echo "CloudBase release completed successfully."
