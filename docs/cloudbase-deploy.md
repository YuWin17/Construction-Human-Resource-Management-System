# 腾讯云 CloudBase 部署

当前项目可以部署到 CloudBase Hosting（前端）和 CloudBase Run（Go API）。生产环境使用当前 CloudBase 环境的 PostgreSQL HTTP API；应用容器不保存数据库文件，因此 CloudBase Run 实例重建、扩缩容和发布不会清空业务档案。

## 0. PostgreSQL 表结构

环境 `chrms-d9gywgbw57877b4fb` 已启用 CloudBase PostgreSQL。业务表结构由 `cloudbase/migrations/` 的版本化迁移创建；发布应用不会执行远端 DDL，也不会清空表数据。

若需要导入本地数据，先创建 SQLite 备份，再通过 CloudBase PG 的受管 DML 或版本化导入脚本执行；不要把 API Key 写入 Git。

```sql
-- 远端业务表只通过 cloudbase/migrations/ 变更。
-- 不执行 DROP TABLE、TRUNCATE 或删除业务记录。
```

应用启动时只初始化容器内存中的工作集，再从 PostgreSQL 加载数据；不会执行远端建表、删表或清表。发布前在 CloudBase 控制台创建手工备份，并确认已启用自动备份和保留周期。

## 1. 登录并选择环境

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

## 2. 部署 Go API

首次部署前，请在 CloudBase 控制台进入环境 `chrms-d9gywgbw57877b4fb` 的“云托管/Cloud Run”，先开通云托管资源；否则 CLI 会返回 `云托管资源未开通`。

`backend/Dockerfile` 已准备好 CloudBase Run 所需的 Linux 容器：

```bash
npx --yes -p @cloudbase/cli tcb cloudrun deploy \
  --env-id "$CLOUDBASE_ENV_ID" \
  --service-name construction-hrms-api \
  --source ./backend \
  --port 8080 \
  --wait
```

在 CloudBase 控制台为该服务配置以下环境变量（请生成新的随机 JWT 密钥）：

```text
APP_ENV=production
HTTP_ADDR=:8080
DATABASE_DRIVER=cloudbase_pg
CLOUDBASE_ENV_ID=chrms-d9gywgbw57877b4fb
CLOUDBASE_API_KEY=CloudBase 创建的一次性服务端 API Key
JWT_SECRET=至少 32 位随机字符串
JWT_TTL_HOURS=8
INITIAL_ADMIN_USERNAME=你的管理员账号
INITIAL_ADMIN_PASSWORD=强密码
DAILY_REMINDER_TOKEN=随机生成的长令牌
CORS_ALLOWED_ORIGINS=先填前端 Hosting 地址
TIMEZONE=Asia/Shanghai
```

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

## 3. 构建并发布前端

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

## 4. 上线前检查

- 直接访问前端地址，确认登录、人才编辑和证书修改正常。
- 确认 API 的 CORS 只允许前端正式地址，不要长期使用 `*`。
- 确认管理员初始密码已修改；不要把 `.env`、数据库文件或 CloudBase 密钥提交到 Git。
- 确认 CloudBase Run 使用的是当前环境的 `cloudbase_pg` 配置，并在部署前完成一次 CloudBase PostgreSQL 手工备份。
- 发布后重复重启 API 服务，确认人才、证书、合同和提醒记录仍然存在。
