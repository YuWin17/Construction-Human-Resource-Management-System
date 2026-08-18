# 腾讯云 CloudBase 部署

当前项目可以部署到 CloudBase Hosting（前端）和 CloudBase Run（Go API）。后端当前使用 SQLite，因此这套步骤适合个人验证和低频使用；CloudBase Run 实例重建时本地 `/app/data/hrms.db` 可能丢失，重要数据请先不要只保存在该实例中。

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
JWT_SECRET=至少 32 位随机字符串
JWT_TTL_HOURS=8
INITIAL_ADMIN_USERNAME=你的管理员账号
INITIAL_ADMIN_PASSWORD=强密码
CORS_ALLOWED_ORIGINS=先填前端 Hosting 地址
TIMEZONE=Asia/Shanghai
```

从 CloudBase Run 服务详情中复制公网访问地址，并确认：

```bash
curl https://你的 API 地址/healthz
```

应返回健康状态 JSON。

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
- 当前 SQLite 仅适合验证。若需要可靠保存人才数据，应先迁移到托管 MySQL/PostgreSQL，并将 `database.OpenSQLite` 改为对应驱动。
