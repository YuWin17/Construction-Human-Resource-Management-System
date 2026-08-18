# 建筑人力资源人才录入系统

面向建筑人力资源中介公司的管理员端网页系统。本仓库当前完成阶段 A：可运行的前后端基础框架、SQLite 数据库迁移、管理员初始化与认证骨架、前端登录和后台布局。

具体人才、证书、合同、附件和提醒业务尚未实现，将在下一阶段确认后开发。

## 前置条件

- Go 1.26+
- Node.js 22+ 与 npm

## 快速启动

### 1. 启动后端

```bash
cd backend
cp .env.example .env
go mod tidy
go run ./cmd/api
```

后端默认运行在 `http://localhost:8080`，健康检查地址为 `http://localhost:8080/healthz`。

首次启动会自动创建 SQLite 数据库、执行迁移，并按 `.env` 中的 `INITIAL_ADMIN_USERNAME` 和 `INITIAL_ADMIN_PASSWORD` 初始化管理员账号。

### 2. 启动前端

另开一个终端：

```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

前端默认运行在 `http://localhost:5173`。

## 验证命令

```bash
cd backend
go test ./...

go vet ./...

cd ../frontend
npm run typecheck
npm run build
```

## 腾讯云 CloudBase 部署

CloudBase Hosting + CloudBase Run 的部署步骤见 [docs/cloudbase-deploy.md](docs/cloudbase-deploy.md)。当前后端使用 SQLite，CloudBase Run 本地磁盘不保证持久化，正式长期使用前应迁移到托管数据库。

## 受限缓存环境

若 Go 或 npm 因用户目录缓存不可写而失败，可将缓存限制在项目目录内：

```bash
cd backend
GOCACHE="$PWD/.cache/go-build" GOMODCACHE="$PWD/.cache/go-mod" GOPATH="$PWD/.cache/go-path" GONOSUMDB='*' go test ./...

cd ../frontend
npm install --cache "$PWD/.cache/npm"
```

`.cache/` 目录已被 Git 忽略。正常开发环境仍建议使用前文的标准命令。

## 目录概览

- `backend/`：Go API 服务，采用 Gin、GORM 和 SQLite。
- `frontend/`：React + TypeScript 管理后台。
- `docs/`：需求、界面功能图和技术实现文档。
