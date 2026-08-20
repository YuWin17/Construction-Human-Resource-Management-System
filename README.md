# 建筑人力资源人才录入系统

面向建筑人力资源中介公司的管理员端网页系统。本仓库当前完成阶段 A：可运行的前后端基础框架、SQLite 数据库迁移、管理员初始化与认证骨架、前端登录和后台布局。

具体人才、证书、合同、附件和提醒业务尚未实现，将在下一阶段确认后开发。

## 前置条件

- Go 1.26+
- Node.js 22+ 与 npm

## 快速启动

本地测试环境通过脚本启动，前端固定为 `http://127.0.0.1:5173`，API 为 `http://127.0.0.1:8080`。脚本使用独立的 `backend/data/hrms.local-test.db`，不会读取或写入 CloudBase PostgreSQL。

```bash
./scripts/local-dev.sh
```

本地测试账号为 `admin`，密码为 `123456`。按 `Ctrl-C` 会同时停止前后端。若要清空**仅本地测试库**后重新初始化账号，运行：

```bash
./scripts/local-dev.sh --reset
```

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

CloudBase Hosting + CloudBase Run 的发布通过以下脚本执行；它会运行测试和构建、创建安全的最小化 Cloud Run 源码目录、部署 API 与前端并验证公网健康检查：

```bash
./scripts/release-cloudbase.sh
```

首次在新电脑发布时，脚本会通过 `tcb login` 请求 CloudBase 授权。只验证本地发布门禁、不修改任何云端资源时，运行 `./scripts/release-cloudbase.sh --dry-run`。完整发布约束与环境变量覆盖方式见 [docs/cloudbase-deploy.md](docs/cloudbase-deploy.md)。

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

- `backend/`：Go API 服务，采用 Gin、GORM；本地使用 SQLite，生产持久化到 CloudBase PostgreSQL。
- `frontend/`：React + TypeScript 管理后台。
- `docs/`：需求、界面功能图和技术实现文档。
