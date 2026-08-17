# 技术实现文档

## 1. 文档目的

本文档把已确认的业务需求转换为可实施的技术方案，用于第二阶段的项目框架搭建。框架搭建完成后，需要再次确认，再进入具体业务功能开发。

## 2. 技术目标

- 使用 Go 作为后端语言，代码以清晰分层和中文注释为学习重点。
- 提供桌面优先的网页管理后台。
- 本地开发可在不安装 Docker 的环境中运行。
- 支持未来切换到 PostgreSQL、对象存储和多管理员，而不需要推翻业务代码。
- 将敏感数据、合同附件、删除审计和到期提醒作为一等需求处理。

## 3. 技术选型

| 层级 | 选型 | 原因 |
| --- | --- | --- |
| 后端语言 | Go 1.26+ | 静态类型、标准库完善，适合作为学习和业务后端项目 |
| HTTP 框架 | Gin | 路由、中间件和请求绑定简洁，生态成熟，适合 REST API |
| ORM | GORM | 便于学习模型关联、迁移和事务；对 SQLite、PostgreSQL 都有支持 |
| 参数校验 | go-playground/validator（通过 Gin 集成） | 结构体标签校验直观，可减少手工校验代码 |
| 认证 | JWT + bcrypt | 无服务端会话依赖，密码安全存储，适合单管理员起步 |
| 日志 | Go slog | Go 标准库，结构化日志，减少额外依赖 |
| 本地数据库 | SQLite | 无需安装数据库服务，适合本地学习和快速启动 |
| 生产数据库 | PostgreSQL | 适合并发、检索和大数据量，便于后续部署 |
| 数据库迁移 | golang-migrate | 版本化 SQL 迁移，避免仅靠 ORM 自动迁移造成生产风险 |
| 文件存储抽象 | 本地磁盘实现起步 | 当前环境可直接运行；接口预留 S3/MinIO 实现 |
| 前端 | React + TypeScript + Vite | 组件化、类型安全、构建速度快，适合管理后台 |
| 前端组件 | Ant Design | 表格、表单、弹窗、上传、日期选择等后台组件完整 |
| 前端数据请求 | TanStack Query + Axios | 缓存、加载、错误和失效处理清晰 |
| 前端路由 | React Router | 管理后台多页面路由的成熟方案 |
| 图标 | lucide-react | 通用、清晰，符合前端界面规范 |
| 导出 | Excelize | Go 侧直接生成 XLSX，满足完整字段导出 |
| 测试 | Go testing + httptest；Vitest | 覆盖业务规则、接口和前端工具函数 |

### 3.1 当前本机环境

已检查的版本：

- Go：`go1.26.5`
- Node.js：`v26.7.0`
- npm：`11.19.0`
- pnpm：未安装
- Docker：未安装

因此本项目框架默认通过 `go run`、SQLite 和 `npm` 启动，不依赖 Docker。生产部署文档会提供 PostgreSQL 和对象存储配置说明，但不以容器作为本地开发前提。

## 4. 总体架构

```text
浏览器
  |
  | HTTPS / JSON API / 文件下载
  v
React + TypeScript 管理后台
  |
  | /api/v1
  v
Go API 服务（Gin）
  |
  +-- 认证中间件（JWT）
  +-- 请求校验与统一错误响应
  +-- 业务服务层
  +-- 审计日志服务
  +-- 提醒查询与状态计算
  |
  +-----------------------+------------------------+
  |                       |                        |
  v                       v                        v
PostgreSQL / SQLite   文件存储接口              后台定时任务
                      本地磁盘 / S3             到期状态刷新
```

### 4.1 分层约束

```text
HTTP Handler
  -> Service（业务编排、事务、权限）
    -> Repository（数据库读写）
      -> Model（数据模型）

Service
  -> Storage（附件存储接口）
  -> Audit（审计记录）
  -> Clock（时间抽象，便于测试）
```

- Handler 只负责解析 HTTP 请求、调用 Service、返回统一响应，不能写复杂业务规则。
- Service 负责身份证唯一校验、合同状态、到期提醒、物理删除、附件权限与审计等业务规则。
- Repository 只负责数据库查询和保存，不能感知 HTTP 细节。
- 每个会改变业务状态的 Service 方法必须在同一事务中处理主数据和审计日志。
- 业务时间通过 `Clock` 接口取得，测试中可传入固定时间，避免直接散落调用 `time.Now()`。

## 5. 仓库目录设计

```text
construction-hrms/
├── README.md
├── .gitignore
├── docs/
│   ├── requirements.md
│   ├── frontend-function-map.md
│   └── technical-design.md
├── backend/
│   ├── cmd/api/main.go                 # 程序入口与依赖组装
│   ├── internal/
│   │   ├── config/                     # 环境变量与配置加载
│   │   ├── domain/                     # 领域模型、枚举、业务错误
│   │   ├── repository/                 # GORM 数据访问实现
│   │   ├── service/                    # 核心业务服务
│   │   ├── transport/http/             # Gin 路由、Handler、DTO、中间件
│   │   ├── storage/                    # 附件存储接口和本地磁盘实现
│   │   ├── scheduler/                  # 状态刷新定时任务
│   │   └── testutil/                   # 测试数据库、固定时钟等工具
│   ├── migrations/                     # 版本化 SQL 迁移
│   ├── storage/                        # 本地开发附件目录，不提交到 Git
│   ├── go.mod
│   ├── go.sum
│   └── .env.example
├── frontend/
│   ├── src/
│   │   ├── api/                        # Axios 客户端与接口模块
│   │   ├── components/                 # 通用 UI 组件
│   │   ├── features/                   # 按业务域组织的页面和组件
│   │   ├── hooks/                      # 查询、表单等复用 Hook
│   │   ├── layouts/                    # 后台布局
│   │   ├── routes/                     # 路由与登录保护
│   │   ├── types/                      # 前端业务类型
│   │   ├── utils/                      # 脱敏、日期、错误转换等工具
│   │   └── main.tsx
│   ├── package.json
│   ├── vite.config.ts
│   └── .env.example
└── scripts/                            # 本地启动、初始化等辅助脚本
```

## 6. 数据库设计

### 6.1 命名和通用约定

- 表名使用复数英文蛇形命名，如 `talents`、`contracts`。
- 主键使用 UUID，避免暴露连续自增 ID；SQLite 和 PostgreSQL 都可存储 UUID 字符串。
- 所有时间字段存 UTC 时间；界面和提醒日期规则以 `Asia/Shanghai` 时区计算。
- 用户可编辑的日期字段使用日期类型或 ISO `YYYY-MM-DD` 字符串语义，不含具体时分秒。
- 敏感字段在数据库中明文存储以支持精确搜索和导出；生产部署依赖磁盘/数据库加密、最小访问权限和备份加密。后续可增加应用层加密。

### 6.2 主要表

#### `admins`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | UUID | 主键 |
| username | varchar(64) | 管理员账号，唯一 |
| password_hash | varchar(255) | bcrypt 哈希，绝不保存明文 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

第一版仅初始化一个管理员，表结构仍保留扩展多管理员的空间。

#### `talents`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | UUID | 主键 |
| name | varchar(50) | 姓名 |
| id_card_number | varchar(18) | 身份证号，唯一索引 |
| gender | varchar(16) | 性别枚举 |
| birth_date | date | 出生日期 |
| phone | varchar(32) | 手机号 |
| native_place | varchar(255) | 籍贯 |
| current_location | varchar(255) | 现居地 |
| education | varchar(32) | 学历枚举 |
| major | varchar(255) | 专业 |
| years_of_experience | integer | 从业年限 |
| cooperation_intentions | json/text | 多选合作意向 |
| expected_locations | json/text | 多选期望地区 |
| note | text | 备注 |
| status | varchar(32) | 在库、暂停合作、已归档 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

索引：`id_card_number` 唯一索引、`phone` 普通索引、`status` 索引、`updated_at DESC` 索引。

#### `certificate_catalogs`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | UUID | 主键 |
| name | varchar(100) | 证书名称，唯一且忽略首尾空白 |
| is_enabled | boolean | 是否可在新增录入时选择 |
| sort_order | integer | 显示排序 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

不物理删除已被引用的字典项；停用后保留历史引用。

#### `certificates`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | UUID | 主键 |
| talent_id | UUID | 外键，关联人才 |
| catalog_id | UUID nullable | 关联证书名称字典 |
| certificate_name_snapshot | varchar(100) | 保存当时名称，避免字典改名影响历史显示 |
| category | varchar(32) | 证书类别 |
| specialty | varchar(100) | 专业/方向 |
| certificate_number | varchar(100) | 证书编号 |
| issuer | varchar(255) | 发证机构 |
| issued_date | date | 发证日期 |
| expires_on | date | 有效期至 |
| registration_status | varchar(32) | 注册状态 |
| registered_company | varchar(255) | 注册单位 |
| is_available | boolean | 是否可用 |
| note | text | 备注 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

约束：同一 `talent_id` 下非空 `certificate_number` 唯一。索引：`talent_id`、`catalog_id`、`expires_on`、`category`。

#### `contracts`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | UUID | 主键 |
| talent_id | UUID | 外键，关联人才 |
| contract_number | varchar(100) | 全局唯一合同编号 |
| company_name | varchar(255) | 签约单位 |
| contract_type | varchar(32) | 合同类型 |
| start_date | date | 开始日期 |
| end_date | date | 结束日期 |
| status | varchar(32) | 履约中、已到期、已解除、已续约 |
| note | text | 签约备注 |
| terminated_on | date nullable | 解除日期 |
| termination_reason | text nullable | 解除原因 |
| renewed_from_contract_id | UUID nullable | 续约来源合同 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

约束和索引：`contract_number` 唯一；`talent_id`、`end_date`、`status` 索引。履约中合同唯一性使用业务事务校验；在 PostgreSQL 中进一步使用部分唯一索引约束 `status = 'active'`。

#### `contract_attachments`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | UUID | 主键 |
| contract_id | UUID | 外键，关联合同 |
| original_filename | varchar(255) | 用户上传的文件名 |
| storage_key | varchar(512) | 存储系统内部路径，不直接暴露 |
| content_type | varchar(100) | 服务端校验后的 MIME 类型 |
| size_bytes | bigint | 文件大小，最大 20 MB |
| uploaded_by_admin_id | UUID | 上传人 |
| created_at | timestamp | 上传时间 |

约束：每份合同最多 10 条附件；删除人才或合同通过事务协调数据库记录删除和文件存储删除。

#### `reminders`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | UUID | 主键 |
| reminder_type | varchar(32) | `contract_expiry`、`certificate_expiry` |
| source_id | UUID | 合同或证书 ID |
| talent_id | UUID | 人才 ID，方便查询 |
| due_date | date | 合同结束日期或证书有效期至 |
| status | varchar(32) | 待处理、已处理、已忽略 |
| handled_at | timestamp nullable | 处理时间 |
| handled_by_admin_id | UUID nullable | 处理人 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

唯一约束：`reminder_type + source_id`。提醒级别不存储，查询时按当前日期动态计算，避免每天批量写入。

#### `system_settings`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| key | varchar(100) | 主键，例如 `contract_reminder_days` |
| value | varchar(255) | 配置值 |
| updated_by_admin_id | UUID | 最后修改人 |
| updated_at | timestamp | 最后修改时间 |

初始值：`contract_reminder_days=30`、`certificate_reminder_days=30`。

#### `audit_logs`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | UUID | 主键 |
| admin_id | UUID nullable | 操作人 |
| action | varchar(64) | 例如 `talent.deleted` |
| resource_type | varchar(64) | 人才、合同、证书、提醒、附件等 |
| resource_id | UUID nullable | 业务对象 ID |
| display_name | varchar(255) | 对象显示名称或合同编号 |
| summary | text | 去敏变更摘要 |
| created_at | timestamp | 操作时间 |

物理删除人才时保留本表记录，但 `summary` 不能保存身份证号、手机号、附件路径等敏感信息。

### 6.3 关系图

```text
admins
  |-- audit_logs
  |-- contract_attachments
  |-- reminders (handled_by)
  |-- system_settings (updated_by)

talents
  |-- certificates -- certificate_catalogs
  |-- contracts -- contract_attachments
  |-- reminders
  |-- audit_logs（逻辑关联，不设外键以保存删除审计）

contracts/certificates
  |-- reminders（每种来源对象一条提醒）
```

## 7. 核心业务设计

### 7.1 人才创建、更新和删除

- 创建人才时验证身份证格式、手机号格式与身份证号全局唯一性。
- 创建和更新人才、其首批证书时使用单一数据库事务。
- 删除人才为不可恢复的物理删除。确认操作后，在事务中删除提醒、附件记录、合同、证书和人才记录；事务成功后删除对应的文件存储对象。
- 文件存储删除失败时记录错误并进入清理重试队列或由定期清理任务处理，避免接口将文件错误当作数据库删除已失败。
- 删除审计日志在删除主数据前写入，内容仅记录人才名称、对象 ID、操作者和“已删除”事实，不保存敏感字段。

### 7.2 证书名称字典与即时新增

- 录入接口接收 `catalog_id` 或 `new_certificate_name`，两者二选一。
- 传入新名称时，服务层去除首尾空白、检查长度与重复名称；不存在则创建启用状态的 `certificate_catalogs` 记录。
- 新增字典和证书记录在同一事务中完成，确保没有孤立字典项或失败证书引用。
- 已引用字典项禁止物理删除；停用只影响后续下拉选项。
- 已创建的证书存 `certificate_name_snapshot`，以保证历史内容可读。

### 7.3 合同状态和签约状态

- 创建履约中合同前，在事务内检查人才状态为“在库”，且不存在其他履约中合同。
- 合同开始日期必须早于结束日期。
- 合同续约会将原合同标记为“已续约”，创建新履约中合同，并建立来源关联。
- 合同解除要求解除日期和解除原因，状态变为“已解除”。
- 每次读取人才列表或详情时，根据当前中国标准日期计算签约状态：存在履约中且当前日期落在合同日期范围内即为已签约。
- 后台任务每天刷新已经结束的履约中合同为已到期，保证状态检索稳定；即使任务尚未运行，读取路径也应正确计算到期展示。

### 7.4 合同附件

- 上传采用 `multipart/form-data`，服务端读取前限制请求体与单文件最大 20 MB。
- 服务端不信任文件扩展名：同时校验扩展名白名单、MIME 类型及文件头特征。
- 每份合同最多 10 个附件；先检查现有数量与本次上传总数。
- 存储键使用随机 UUID 目录和文件名，不能使用原始文件名构造路径，防止路径穿越。
- 附件下载必须经过认证和合同存在校验；默认通过 API 流式返回，文件存储目录不公开暴露。
- PDF 和图片附件返回受认证保护的预览流；Office 文件只提供下载。
- 删除附件前由前端二次确认，后端仍执行存在性、归属和权限检查，并写审计日志。

### 7.5 到期提醒

提醒采用“记录 + 动态级别”的混合方式：

1. 合同或证书被创建、修改、续约、解除、删除时，同步创建、更新或删除对应提醒记录。
2. 只有履约中合同和填写有效期至的证书拥有有效提醒。
3. 到期提醒查询按 `Asia/Shanghai` 的当天日期和配置天数过滤：`due_date <= today + configured_days`。
4. 已处理和已忽略提醒保留记录，可通过状态筛选回查。
5. 当前级别动态计算：已到期、剩余不超过 7 天、提醒窗口内其他日期。
6. 系统设置变更不需要重写全部提醒记录，查询层按新配置重新筛选即可。
7. 每日后台任务刷新合同到期状态，同时检查并修复遗漏的提醒记录。

本期不发送短信、邮件、微信或浏览器推送。提醒通过仪表盘、导航徽标和提醒中心呈现。

### 7.6 导出

- 人才列表导出接口接收与列表一致的筛选条件。
- 前端须先弹出完整敏感字段风险确认框，再调用导出接口。
- API 返回 XLSX 文件，字段包括完整身份证号、完整手机号、人才信息、签约状态和证书摘要。
- 后端记录 `talent.exported` 审计日志，摘要记录筛选条件和导出行数，不记录导出文件内容。
- 首期同步导出设置安全上限，例如 10,000 行；超过上限提示缩小筛选范围。异步导出留作后续优化。

## 8. REST API 设计

统一前缀：`/api/v1`。

### 8.1 通用响应

成功：

```json
{
  "data": {},
  "meta": {}
}
```

失败：

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "请求参数不正确",
    "details": [
      {"field": "phone", "message": "请输入正确的手机号"}
    ]
  }
}
```

### 8.2 认证

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/auth/login` | 管理员登录，返回短期访问令牌 |
| POST | `/auth/logout` | 前端清除令牌并记录退出，可保留为无状态接口 |
| GET | `/auth/me` | 获取当前管理员信息 |

### 8.3 仪表盘与人才

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/dashboard` | 统计与最近数据 |
| GET | `/talents` | 分页、搜索、组合筛选 |
| POST | `/talents` | 创建人才和初始证书 |
| GET | `/talents/:id` | 获取人才详情和汇总 |
| PUT | `/talents/:id` | 更新人才基本信息 |
| POST | `/talents/:id/archive` | 归档人才 |
| POST | `/talents/:id/restore` | 恢复在库 |
| DELETE | `/talents/:id` | 物理删除，前端已二次确认，后端写审计 |
| GET | `/talents/export` | 导出当前筛选结果 XLSX |

### 8.4 证书和字典

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/talents/:id/certificates` | 创建人才证书，可即时新增字典名称 |
| PUT | `/talents/:id/certificates/:certificateId` | 更新证书 |
| DELETE | `/talents/:id/certificates/:certificateId` | 删除证书 |
| GET | `/certificate-catalogs` | 获取启用/全部字典项 |
| POST | `/certificate-catalogs` | 创建字典项 |
| PUT | `/certificate-catalogs/:id` | 编辑名称、排序、启用状态 |
| POST | `/certificate-catalogs/:id/disable` | 停用已引用项 |
| DELETE | `/certificate-catalogs/:id` | 仅未引用项允许删除 |

### 8.5 合同和附件

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/talents/:id/contracts` | 创建合同 |
| GET | `/talents/:id/contracts/:contractId` | 获取合同与附件 |
| PUT | `/talents/:id/contracts/:contractId` | 编辑允许修改的合同字段 |
| POST | `/talents/:id/contracts/:contractId/renew` | 基于原合同续约 |
| POST | `/talents/:id/contracts/:contractId/terminate` | 解除合同 |
| POST | `/contracts/:contractId/attachments` | 批量上传合同附件 |
| GET | `/contracts/:contractId/attachments/:attachmentId/download` | 下载附件 |
| GET | `/contracts/:contractId/attachments/:attachmentId/preview` | 预览 PDF 或图片 |
| DELETE | `/contracts/:contractId/attachments/:attachmentId` | 删除附件 |

### 8.6 提醒、设置和审计

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/reminders` | 合同/证书提醒分页筛选 |
| POST | `/reminders/:id/handle` | 标记已处理 |
| POST | `/reminders/:id/ignore` | 标记已忽略 |
| POST | `/reminders/batch-handle` | 批量标记已处理 |
| POST | `/reminders/batch-ignore` | 批量忽略 |
| GET | `/settings` | 获取提醒配置 |
| PUT | `/settings/reminders` | 更新合同和证书提醒天数 |
| GET | `/audit-logs` | 操作日志分页筛选 |

## 9. 认证、安全和错误处理

### 9.1 认证流程

1. 管理员提交账号密码。
2. 后端查询账号，使用 bcrypt 校验密码。
3. 校验成功后签发 JWT，包含管理员 ID、账号和过期时间。
4. 前端仅在内存或受控存储中保存令牌；每次 API 请求携带 `Authorization: Bearer <token>`。
5. 后端认证中间件校验签名、过期时间与管理员状态，并将管理员信息写入请求上下文。

首期可采用访问令牌方案。若后续需要更高安全性，可改为 HttpOnly Cookie + 刷新令牌，不改变业务 API。

### 9.2 安全措施

- bcrypt 密码哈希成本使用安全默认值。
- JWT 密钥从环境变量读取，开发和生产必须使用不同随机值。
- API 启用 CORS 白名单，只允许配置的前端来源。
- 登录接口限制失败次数或使用限流中间件，降低暴力破解风险。
- 所有 ID 使用 UUID 并验证资源归属关系。
- 附件上传进行文件大小、类型、文件头、数量、路径和鉴权校验。
- 列表默认对身份证号和手机号脱敏；详情、编辑和导出完整展示需登录。
- 物理删除、附件删除、忽略提醒等操作在前端二次确认，后端审计记录不可省略。
- 生产部署启用 HTTPS、数据库备份加密和附件目录最小权限。

### 9.3 错误处理

- 业务错误定义统一错误码，如 `TALENT_ID_CARD_EXISTS`、`ACTIVE_CONTRACT_EXISTS`、`ATTACHMENT_LIMIT_EXCEEDED`。
- Handler 统一转换错误为固定 JSON 格式，不向客户端泄露 SQL、文件路径或堆栈信息。
- 服务端使用 `slog` 输出请求 ID、错误码与上下文；日志中不得记录完整身份证号、手机号、密码或 JWT。

## 10. 前端实现设计

### 10.1 页面模块

| 路由 | 页面 | 核心组件 |
| --- | --- | --- |
| `/login` | 登录 | 登录表单 |
| `/dashboard` | 仪表盘 | 统计卡、提醒表、最近人才表 |
| `/talents` | 人才列表 | 搜索筛选栏、表格、导出确认弹窗 |
| `/talents/new` | 新增人才 | 人才表单、动态证书表单 |
| `/talents/:id` | 人才详情 | 基本资料、证书、合同、日志页签 |
| `/talents/:id/edit` | 编辑人才 | 人才表单 |
| `/reminders` | 到期提醒 | 类型筛选、状态筛选、批量处理表格 |
| `/audit-logs` | 操作日志 | 过滤表格 |
| `/settings` | 系统设置 | 提醒配置、证书名称配置表 |

### 10.2 前端状态策略

- TanStack Query 管理服务端数据：人才、证书、合同、提醒、设置和日志。
- 成功变更后精准失效相关查询，例如更新提醒设置后失效 `dashboard`、`reminders`、`settings`。
- Ant Design Form 负责表单临时状态与字段校验；不将所有表单数据放入全局状态。
- 使用 React Router 的路由守卫保护管理后台路由；遇到 401 时清理令牌并跳转登录。

### 10.3 敏感数据展示

- 人才列表和仪表盘列表调用后端返回的脱敏字段，前端不负责从完整值脱敏。
- 人才详情、编辑表单只在已认证请求下取得完整敏感字段。
- 导出按钮先展示确认对话框，明确说明导出文件包含完整身份证号和手机号。

### 10.4 文件上传与预览

- 选择文件时在浏览器端先验证文件数量、文件大小和扩展名，随后由服务端作最终校验。
- 上传列表显示文件名、大小、进度、失败原因和移除操作。
- PDF 和图片点击预览时请求受保护预览地址，并在抽屉或对话框中展示。
- Office 文档只提供下载图标按钮。

## 11. 后台任务和时间规则

### 11.1 定时任务

服务启动后使用 Go `time.Ticker` 或可替换的 scheduler，每天在 `Asia/Shanghai` 00:10 执行：

- 将结束日期早于当天的履约中合同标记为已到期。
- 查找应存在但缺失的合同/证书提醒并补齐。
- 删除或失效来源对象的提醒记录。
- 记录任务执行结果和异常。

任务实现必须可手动触发，便于集成测试和运维排查。由于提醒列表会动态计算到期等级与窗口，任务延迟不会导致当天页面展示错误。

### 11.2 日期示例

假设当前中国标准日期为 `2026-03-20`，合同提醒配置为 30 天：

- 合同结束日期 `2026-04-19`：进入提醒窗口，剩余 30 天，级别“正常提醒”。
- 合同结束日期 `2026-03-27`：剩余 7 天，级别“即将到期”。
- 合同结束日期 `2026-03-19`：已到期，级别“已到期”。

证书使用相同算法，但读取证书提醒天数配置。

## 12. 数据迁移和初始化

### 12.1 迁移顺序

1. 创建管理员、人才、证书名称字典和证书表。
2. 创建合同、附件、提醒表。
3. 创建系统设置和审计日志表。
4. 创建索引、唯一约束和 PostgreSQL 特有的部分唯一索引。
5. 写入默认设置和常用证书名称种子数据。

### 12.2 首次管理员初始化

- 使用环境变量 `INITIAL_ADMIN_USERNAME` 和 `INITIAL_ADMIN_PASSWORD` 初始化第一个管理员。
- 启动时若管理员表为空，则读取变量、使用 bcrypt 加密密码并创建管理员。
- 生产环境若变量缺失则服务拒绝启动；开发环境可提供明确的开发默认值，但必须打印安全警告。

## 13. 配置项

后端 `.env.example` 计划包含：

```dotenv
APP_ENV=development
HTTP_ADDR=:8080
DATABASE_DRIVER=sqlite
DATABASE_DSN=./data/hrms.db
JWT_SECRET=replace-with-a-long-random-secret
CORS_ALLOWED_ORIGINS=http://localhost:5173
FILE_STORAGE_DRIVER=local
FILE_STORAGE_LOCAL_DIR=./storage
INITIAL_ADMIN_USERNAME=admin
INITIAL_ADMIN_PASSWORD=change-me-now
TIMEZONE=Asia/Shanghai
MAX_ATTACHMENT_SIZE_MB=20
MAX_ATTACHMENTS_PER_CONTRACT=10
EXPORT_MAX_ROWS=10000
```

前端 `.env.example` 计划包含：

```dotenv
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

机密 `.env`、SQLite 数据库文件和本地附件目录必须写入 `.gitignore`。

## 14. 测试策略

### 14.1 后端

- 单元测试：身份证/手机号校验、日期等级计算、提醒窗口、脱敏函数、证书名称标准化。
- 服务测试：身份证唯一性、履约中合同唯一性、续约、解除、附件数量限制、人才物理删除和审计记录。
- Handler 测试：认证、权限、校验错误格式、分页筛选和文件上传下载。
- Repository 测试：SQLite 测试数据库上的查询和事务。
- 每个提醒规则用固定 `Clock` 测试，覆盖当天、7 天、配置边界和已到期情形。

### 14.2 前端

- 单元测试：脱敏展示、日期格式化、提醒等级标签转换。
- 组件测试：人才表单校验、证书名称“录入新证书名”交互、删除确认、导出风险确认。
- 核心流程冒烟测试：登录、创建人才、增加证书、创建合同、上传附件、处理提醒。

## 15. 实施阶段

### 阶段 A：框架搭建（下一步，待确认后执行）

1. 初始化 `backend` Go 模块和 `frontend` Vite React TypeScript 项目。
2. 建立目录分层、配置加载、健康检查、统一响应、日志和 CORS。
3. 建立 SQLite 本地连接、迁移骨架、管理员初始化和登录 API 骨架。
4. 建立 React 后台布局、路由、登录页、API 客户端和空页面。
5. 配置本地启动命令、示例环境变量、格式化和基础测试命令。

此阶段只构建可运行骨架，不实现人才、证书、合同、附件、提醒等具体业务。

### 阶段 B：核心人才与证书

1. 人才 CRUD、筛选、身份证校验和脱敏。
2. 证书 CRUD、字典配置和即时新增名称。
3. 人才列表、表单、详情页基础资料和证书页签。

### 阶段 C：合同与附件

1. 合同创建、状态、续约、解除与签约状态计算。
2. 本地文件存储、附件上传、下载、预览和删除。
3. 合同界面和附件交互。

### 阶段 D：提醒、导出与审计

1. 合同/证书提醒、定时刷新、仪表盘和提醒中心。
2. Excel 导出及完整敏感数据确认。
3. 审计日志、系统设置、测试完善和部署说明。

## 16. 框架搭建验收标准

进入具体业务实现前，框架阶段应满足：

1. 后端能通过 `go run ./cmd/api` 启动并返回健康检查结果。
2. 后端能加载环境变量、连接 SQLite、执行迁移，并在空库中创建管理员。
3. 后端有统一 JSON 成功/错误响应、中间件、请求日志和 CORS 配置。
4. 登录接口和 `GET /auth/me` 骨架可运行，受保护路由能拒绝未认证请求。
5. 前端能通过 `npm run dev` 启动，显示登录页和受保护的后台基础布局。
6. 前端 API 客户端能处理 401 并跳转登录。
7. 前后端的 README 提供准确启动命令、环境变量说明和基础测试命令。
8. Go 格式化、Go 测试与前端类型检查命令可执行。

## 17. 本阶段确认点

请确认本技术方案，尤其是以下实现决策：

1. 后端采用 `Gin + GORM`，本地 SQLite、生产 PostgreSQL。
2. 前端采用 `React + TypeScript + Vite + Ant Design`，使用 npm 管理依赖。
3. 附件在第一版本地磁盘存储，通过受认证 API 提供下载和预览，并预留对象存储接口。
4. 到期提醒采用数据库提醒记录与查询时动态计算级别的混合方案，每日执行一次状态修复任务。
5. 人才删除为不可恢复的物理删除，审计日志只保留去敏记录。
6. 本次确认后仅进入“框架搭建”阶段，不提前实现具体业务模块。
