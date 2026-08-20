# 腾讯云 CloudBase 生产发布

当前项目部署到 CloudBase Hosting（前端）和 CloudBase Run（Go API）。生产环境使用当前 CloudBase 环境的 PostgreSQL HTTP API；应用容器不保存数据库文件，因此 CloudBase Run 实例重建、扩缩容和发布不会清空业务档案。

目标环境：`chrms-d9gywgbw57877b4fb`。Cloud Run 服务：`construction-hrms-api`。服务监听 `8080`，健康检查为 `/healthz`，容量保持 `MinNum=1`、`MaxNum=1`。

## 0. PostgreSQL 表结构

环境 `chrms-d9gywgbw57877b4fb` 已启用 CloudBase PostgreSQL。业务表结构由 `cloudbase/migrations/` 的版本化迁移创建；发布应用不会执行远端 DDL，也不会清空表数据。

若需要导入本地数据，先创建 SQLite 备份，再通过 CloudBase PG 的受管 DML 或版本化导入脚本执行；不要把 API Key 写入 Git。

```sql
-- 远端业务表只通过 cloudbase/migrations/ 变更。
-- 不执行 DROP TABLE、TRUNCATE 或删除业务记录。
```

应用启动时只初始化容器内存中的工作集，再从 PostgreSQL 加载数据；不会执行远端建表、删表或清表。发布前在 CloudBase 控制台创建手工备份，并确认已启用自动备份和保留周期。

## 1. 一键发布

日常发布只执行项目根目录的脚本：

```bash
./scripts/release-cloudbase.sh
```

脚本会依次执行 Go 测试、`go vet`、前端依赖安装与类型检查、生产前端构建、最小化 Cloud Run 源码目录安全检查、Cloud Run 发布、Hosting 上传和公网健康检查。它不会修改 Cloud Run 的环境变量、密钥、CORS、数据库结构或业务数据，也不会执行删除、清表或迁移。

首次在一台电脑上执行时，脚本会调用 `tcb login` 完成授权。只验证本地质量门禁和发布目录、不改动任何 CloudBase 资源时，执行：

```bash
./scripts/release-cloudbase.sh --dry-run
```

默认目标是环境 `chrms-d9gywgbw57877b4fb`、服务 `construction-hrms-api` 与当前正式 API/Hosting 地址。需要切换到其他环境时，仅通过环境变量覆盖目标，不要编辑脚本或把凭据写入文件：

```bash
CLOUDBASE_ENV_ID="目标环境 ID" \
CLOUDBASE_SERVICE_NAME="目标服务名" \
CLOUDBASE_API_ORIGIN="https://目标 API 地址" \
CLOUDBASE_HOSTING_ORIGIN="https://目标 Hosting 地址" \
./scripts/release-cloudbase.sh
```

## 2. 发布目录安全机制

**CloudBase Cloud Run 的源码上传会先打包 ZIP，不能只依赖 `.dockerignore`。** `.dockerignore` 只约束后续 Docker 镜像层的 `COPY`，不会可靠地约束 CloudBase 上传 ZIP。因此，绝不能把项目根目录直接作为 `manageCloudRun(action="deploy")` 的 `targetPath`。

本机运行时文件和构建产物绝不能上传到 CloudBase 源码构建，包括：

- `backend/data/hrms.db`、所有 `*.db`、`*.sqlite`、`*.sqlite-wal`、`*.sqlite-shm`；
- `backend/.cache`、`frontend/.cache`、任意 `.cache`；
- `backend/api`、`backend/hrms-api` 等本地 Go 二进制文件；
- `node_modules`、`dist`、`.env`、日志、上传目录和本地存储目录。

根 `.dockerignore` 仍用于 Docker 镜像层防护。`scripts/release-cloudbase.sh` 已将以下准备与检查固化为强制步骤：每次发布仅复制允许的后端源码，不复制、不压缩、不提交本地运行时文件。脚本结束后会删除临时目录 `.cloudbase-release.*`。

```bash
release_dir="$(mktemp -d "$PWD/.cloudbase-release.XXXXXX")"
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
  backend/ "$release_dir/backend/"
cp Dockerfile .dockerignore "$release_dir/"

if find "$release_dir" -type f \
  \( -name '*.db' -o -name '*.sqlite' -o -name '*.sqlite-wal' -o -name '*.sqlite-shm' \) \
  -print -quit | grep -q .; then
  echo 'Refusing deployment: local database found in release directory.' >&2
  exit 1
fi

find "$release_dir" \
  \( -type d -name '.cache' -o -type f -name '.env' -o -type f -name '.env.*' \) \
  -print -quit | grep -q . && {
  echo 'Refusing deployment: cache or environment file found in release directory.' >&2
  exit 1
}

du -sh "$release_dir"
test -f "$release_dir/backend/cmd/api/main.go"
rg -n '^(\.cache|\*\*/\.cache|\*\*/node_modules|\*\*/dist|\*\*/\.env|backend/api|backend/hrms-api|backend/data|\*\.db|\*\.sqlite|\*\.sqlite-shm|\*\.sqlite-wal)$' .dockerignore
```

以上命令只有在临时发布目录中不存在本地数据库、缓存和 `.env` 时才继续。`du` 的结果应很小，不应出现数百 MB 级别的缓存或数据库；`rg` 必须列出全部 Docker 层防护规则。如任一检查失败，停止发布。CloudBase 构建日志的“检出 ZIP 包”阶段也不得出现 `backend/data/`、`*.db`、`.cache/`、`node_modules/`、`dist/` 或 `.env`。

生产环境必须保持 `DATABASE_DRIVER=cloudbase_pg`，不能设置 `DATABASE_DSN` 指向 SQLite 或其他本地文件。业务数据只保存在 CloudBase PostgreSQL；发布前执行 CloudBase PostgreSQL 手工备份，发布过程不执行远端 DDL、`TRUNCATE` 或数据清理。

## 3. 登录并选择环境

在项目根目录执行：

```bash
npx --yes -p @cloudbase/cli tcb login
npx --yes -p @cloudbase/cli tcb env list
```

记下目标环境 ID。下面以环境变量 `CLOUDBASE_ENV_ID` 表示它：

```bash
export CLOUDBASE_ENV_ID="你的环境 ID"
```

也可以把同样的信息填入 `cloudbase.example.json`，另存为本地的 `cloudbase.json`；不要把真实环境 ID、密钥提交到 Git。

## 4. 部署 Go API

首次部署前，请在 CloudBase 控制台进入环境 `chrms-d9gywgbw57877b4fb` 的“云托管/Cloud Run”，先开通云托管资源；否则 CLI 会返回 `云托管资源未开通`。

根目录的 `Dockerfile` 已准备好 CloudBase Run 所需的 Linux 容器。通过 CloudBase MCP 的 `manageCloudRun(action="deploy")` 部署时，`targetPath` 必须是上一步创建的临时发布目录的绝对路径，`Dockerfile` 为 `Dockerfile`，`BuildDir` 为 `.`。不要使用项目根目录、历史文档中的 `--source ./backend` 命令，或任何包含本地数据库/缓存的目录。

部署时保留现有的生产环境变量和密钥，并确认以下非敏感配置：

```text
APP_ENV=production
HTTP_ADDR=:8080
DATABASE_DRIVER=cloudbase_pg
CLOUDBASE_ENV_ID=chrms-d9gywgbw57877b4fb
CORS_ALLOWED_ORIGINS=https://chrms-d9gywgbw57877b4fb-1416107181.tcloudbaseapp.com
TIMEZONE=Asia/Shanghai
```

服务端 API Key、JWT 密钥、管理员初始凭据与定时提醒令牌只能保存在 Cloud Run 的运行时环境变量中，不能出现在仓库、构建参数、前端产物或发布记录中。

如果需要 CLI 作为 MCP 不可用时的后备手段，依然从项目根目录进行构建并先阅读 CloudBase CLI 的 Cloud Run 文档；不要使用 `tcb deploy` 或指定 `./backend` 作为源码目录。

未配置 `CLOUDBASE_API_KEY`、或生产环境误设为 SQLite 时，服务会拒绝启动，不会在 Cloud Run 容器中创建一个空数据库。API Key 只能通过 CloudBase Run 环境变量注入，不要提交到仓库或前端构建产物。

从 CloudBase Run 服务详情中复制公网访问地址，并确认：

```bash
curl https://你的 API 地址/healthz
```

应返回健康状态 JSON。

企业微信定时任务可读取以下纯文本 URL。将 `DAILY_REMINDER_TOKEN` 替换为上一步配置的实际值，勿将完整 URL 泄露或提交到 Git：

```text
https://你的 API 地址/api/v1/integrations/wecom/daily-reminder?token=DAILY_REMINDER_TOKEN
```

## 5. 构建并发布前端

将 API 地址写入构建时环境变量。前端变量只在构建时读取：

```bash
cd frontend
VITE_API_BASE_URL="https://你的 API 地址/api/v1" npm run build
cd ..
```

发布静态文件：

```bash
npx --yes -p @cloudbase/cli tcb hosting deploy \
  ./frontend/dist \
  --env-id "$CLOUDBASE_ENV_ID"
```

在 Hosting 控制台复制默认访问地址，再回到 CloudBase Run 更新 `CORS_ALLOWED_ORIGINS`，然后重启 API 服务。

## 6. 发布后验证与回滚

1. 在 Cloud Run 部署记录中确认新版本状态为 `normal` 且流量为 100%。
2. 请求 `https://construction-hrms-api-298690-10-1416107181.sh.run.tcloudbase.com/healthz`，应返回 HTTP 200 的固定健康状态 JSON，且不得回显请求头或环境变量。
3. 在 CloudBase PostgreSQL 中核对业务数据仍存在，例如查询 `public.admins` 与 `public.talents` 的记录数。不得用本机 SQLite 结果替代该验证。
4. 用正式 Hosting 地址完成登录、人才编辑和证书修改的冒烟测试，确认 CORS 仅允许正式前端来源。
5. 若健康检查或数据核对失败，立即在 Cloud Run 发布记录中回滚到上一稳定版本；不通过重新上传本地数据库来恢复数据。

发布完成后，确认管理员初始密码已修改；不要把 `.env`、数据库文件或 CloudBase 密钥提交到 Git。
