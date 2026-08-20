#!/bin/zsh
set -euo pipefail

project_dir="/Users/nb110/work/My Project/Construction Human Resource Management System"
caddy_bin="/opt/homebrew/bin/caddy"

cd "$project_dir/backend"
export APP_ENV="production"
export HTTP_ADDR="127.0.0.1:8080"
export CORS_ALLOWED_ORIGINS="http://100.97.47.112:8081"
./hrms-api > "$project_dir/backend/.logs/hrms-api.log" 2>&1 &
api_pid=$!

cleanup() {
  kill "$api_pid" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

cd "$project_dir"
"$caddy_bin" run --config "$project_dir/Caddyfile" --adapter caddyfile \
  > "$project_dir/backend/.logs/caddy.log" 2>&1 &
caddy_pid=$!

wait "$caddy_pid"
