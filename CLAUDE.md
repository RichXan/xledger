# CLAUDE.md

本文件为 Claude Code（claude.ai/code）在此仓库中工作时提供指引。

## 仓库结构

- `backend/` — Go API 服务、业务逻辑、SQL 迁移、OpenAPI 合约、集成测试、本地基础设施所需的 Docker Compose。
- `deploy/` — 面向部署的完整 Docker Compose 编排、后端部署配置、前端生产镜像与 Nginx 配置。
- `frontend/app/` — 使用 Vite + React + TypeScript 构建的 Web 客户端。
- `frontend/stitch/` — 导出的设计稿/产物目录，可作为 UI 参考。
- `docs/superpowers/specs/` 和 `docs/superpowers/plans/` — 早期规划阶段产出的架构/规格与实施计划文档。
- `PRD.md` — v1 的产品需求与核心领域规则。

仓库根目录没有统一的 workspace 脚本；后端和前端需要分别在各自目录中独立开发。

## 常用命令

### 完整部署（`/deploy`）

主要参考：`deploy/docker-compose.yaml` 与 `deploy/README.md`。

- 启动完整部署栈（Postgres + Redis + Backend + Frontend）：
  - `docker compose up -d --build`
- 查看服务状态：
  - `docker compose ps`
- 查看日志：
  - `docker compose logs -f`
- 停止并移除服务：
  - `docker compose down`
- 停止并移除服务与数据卷：
  - `docker compose down -v`

### 后端（`/backend`）

主要参考：`backend/README.md` 了解运行/测试流程，`backend/cmd/api/main.go` 了解 CLI 行为。

- 运行全部测试：
  - `go test ./...`
- 运行集成测试：
  - `go test ./tests/integration -count=1`
- 运行单个 Go 测试：
  - `go test ./internal/accounting -run TestLedgerService_Create -count=1`
- 运行迁移测试（需要 `TEST_DATABASE_URL`）：
  - `TEST_DATABASE_URL=postgres://... go test ./migrations -count=1`
- 启动本地开发 API：
  - 将 `config/config.yaml.example` 复制为 `config/config.yaml`
  - 在该配置中补齐必需的 auth 字段
  - `go run ./cmd/api`
- 运行迁移命令：
  - `go run ./cmd/api migrate up`
  - `go run ./cmd/api migrate down`
  - `go run ./cmd/api migrate status`
- 仅启动本地 Postgres + Redis：
  - `docker compose -f docker-compose.backend.yml up -d`
- 停止本地 Postgres + Redis：
  - `docker compose -f docker-compose.backend.yml down -v`
- 启动完整后端栈（API + Postgres + Redis）：
  - `docker compose -f docker-compose.yml up -d`
- 格式化 / vet：
  - `go fmt ./...`
  - `go vet ./...`

### 前端（`/frontend/app`）

主要参考：`frontend/app/package.json`。

- 安装依赖：
  - `npm install`
- 启动开发服务器：
  - `npm run dev`
- 构建生产包：
  - `npm run build`
- 预览生产构建：
  - `npm run preview`
- 运行全部测试：
  - `npm run test`
- 以 watch 模式运行测试：
  - `npm run test:watch`
- 运行单个 Vitest 文件：
  - `npm run test -- src/App.test.tsx`
- 类型检查：
  - `npm run lint`
- 运行前端主校验流程：
  - `npm run check`

## 产品/领域上下文

修改核心记账行为前，请先阅读 `PRD.md`。v1 最重要的规则如下：

- 账户是用户级全局资源，不隶属于某个账本。
- 每笔交易都属于某个账本；默认账本会被自动创建。
- `account_id = NULL` 的交易仍会影响账本统计，但不会影响账户余额或总资产。
- 仪表盘、报表和导出功能都必须保持与上述记账语义一致。

## 后端架构

### 入口与装配

- API 入口：`backend/cmd/api/main.go`
- 配置加载：`backend/internal/bootstrap/config/config.go`
- 基础设施连接：`backend/internal/bootstrap/infrastructure/connections.go`
- Router 组装：`backend/internal/bootstrap/http/router.go`

`main.go` 同时支持服务启动与迁移命令。它会加载配置、按需连接 Postgres 和 Redis，在启用 Postgres 时自动应用迁移，并根据数据库是否配置，在内存版装配和 `NewRouterWithPostgreSQL(...)` 之间做选择。

这个区分很重要：部分测试和本地开发是有意依赖内存模式运行的。

### HTTP/API 形态

- API 基础前缀：`/api`
- 认证路由：`/api/auth/*`
- 业务路由直接挂在 `/api` 下
- OpenAPI 合约：`backend/openapi/openapi.yaml`
- 业务 JSON 响应使用统一信封结构：`{ code, message, data }`
- 分页列表使用 `data.items` 和 `data.pagination`
- PAT 管理路径位于 `/api/personal-access-tokens`

Router 会组合各领域 handler，并通过共享的认证中间件包裹大多数业务路由。

### 后端主要领域

Go 代码按领域/上下文组织，而不是按技术分层组织：

- `internal/auth` — 邮箱验证码登录、密码登录、刷新/登出、Google OAuth、会话/令牌处理
- `internal/accounting` — 账本、账户、交易、转账、quick-add、核心余额逻辑
- `internal/classification` — 分类、标签、模板/历史
- `internal/reporting` — 总览、趋势、分类聚合、可选的 Redis 缓存
- `internal/portability` — CSV 导入/导出、个人访问令牌、快捷指令
- `internal/automation` — 围绕 quick-entry 流程的自动化契约/适配器
- `internal/push` — Web Push 订阅端点/服务
- `internal/common/httpx` — 共享 HTTP 响应/请求辅助工具

每个领域通常会把 handler、service 和 repository 实现放在一起。

### 持久化与基础设施

- Postgres 仓库在 `backend/internal/bootstrap/http/router.go` 中构建
- SQL Schema 位于 `backend/migrations/`
- 本地 Docker 基础设施位于 `backend/docker-compose.backend.yml`
- 完整容器化后端栈位于 `backend/docker-compose.yml`
- 面向整仓库的部署编排位于 `deploy/docker-compose.yaml`

需要注意的行为：

- 当 SMTP 配置只是占位值时，系统会回退到开发用邮件发送器，而不是真实 SMTP 投递
- 如果未显式配置，Google OAuth 回调地址和默认前端返回地址会在配置中自动补齐

### 测试

- 大多数单元测试与实现文件同目录，位于 `internal/...`
- API / 集成测试位于 `backend/tests/integration/`
- 迁移锁定测试位于 `backend/migrations/`

修改 API 行为时，需要同时校验实现、OpenAPI 合约以及相关集成测试预期。

## 前端架构

### 入口与应用外壳

- 前端入口：`frontend/app/src/main.tsx`
- 路由树：`frontend/app/src/App.tsx`
- 共享布局壳：`frontend/app/src/components/layout/app-shell.tsx`

该应用是一个纯客户端 React SPA，使用了：

- `react-router-dom` 处理路由
- `@tanstack/react-query` 管理服务端状态
- 自定义 auth context 处理会话初始化与登出
- `i18next` 处理中英文国际化
- service worker 与 PWA 安装/更新流程

已认证页面会在 `AppShell` 中渲染，并由 `RequireAuth` 进行保护。公开路由包括登录页、Google OAuth 回调页和 PWA 引导页。

### 前端功能划分

前端主要按功能拆分：

- `src/features/auth` — 认证 API、认证上下文、基于 localStorage 的会话持久化、路由守卫
- `src/features/transactions` — 交易列表/表单流程与 React Query hooks
- `src/features/management` — PAT、CSV 导入导出客户端调用、管理 hooks
- `src/features/reporting` — 总览/趋势/分类统计 hooks 与 API 调用
- `src/features/offline` — 基于 Dexie 的客户端离线队列/存储
- `src/features/pwa` — service worker 注册与更新辅助逻辑
- `src/features/notifications` — 推送通知偏好/存储辅助逻辑
- `src/pages` — 页面级组合
- `src/components` — 可复用 UI / 布局组件
- `src/lib` — API 工具、格式化与通用辅助函数

常见模式如下：

1. 功能内的 `*-api.ts` 负责封装后端端点
2. 功能内的 `*-hooks.ts` 负责用 React Query 包装这些调用
3. 页面直接消费这些 hooks

### 认证与 API 假设

- API 基地址在 `frontend/app/src/lib/api.ts` 中被硬编码为 `/api`
- 共享 API helper 假定后端响应使用统一信封，并在失败时抛出带有 `status` 和 `code` 的 `ApiError`
- 认证会话通过 `frontend/app/src/features/auth/auth-storage.ts` 存入 `localStorage`
- `AuthProvider` 会先请求 `/auth/me`，若需要则再通过 `/auth/refresh` 刷新
- 认证初始化逻辑明确防止 React 18 StrictMode 下出现重复刷新
- 受保护页面依赖前端会话状态中的 bearer token
- `frontend/app/vite.config.ts` 中将 Vite 开发服务器的 `/api` 代理到 `http://127.0.0.1:8080`

### 需要了解的前端行为

- 导入是多步骤 UI：上传 → 预览/映射 → 确认/结果
- 语言检测顺序为 querystring → localStorage → 浏览器语言
- service worker 注册是非阻塞的；更新可用性通过自定义浏览器事件向外暴露
- 离线队列通过 Dexie 使用 IndexedDB，而不是 localStorage

### 前端测试

- 测试运行器：Vitest + jsdom
- 配置文件：`frontend/app/vitest.config.ts`
- 初始化文件：`frontend/app/src/test/setup.ts`
- 测试大多与源码并列放在 `src/*.test.tsx`

## 重要文件

- `PRD.md` — 产品范围和记账规则的唯一事实来源。
- `backend/README.md` — 后端运行/测试/API 说明。
- `deploy/README.md` — 完整部署栈的启动与停止说明。
- `deploy/docker-compose.yaml` — 部署用 Compose 入口。
- `backend/cmd/api/main.go` — 进程启动、迁移 CLI 与模式选择。
- `backend/internal/bootstrap/http/router.go` — 路由注册与依赖装配。
- `backend/openapi/openapi.yaml` — API 合约。
- `backend/migrations/` — 数据库 schema 与迁移测试。
- `frontend/app/package.json` — 前端脚本。
- `frontend/app/src/main.tsx` — React 根组件、路由、query client、auth provider、service worker 初始化。
- `frontend/app/src/App.tsx` — 路由结构。
- `frontend/app/src/lib/api.ts` — 共享 HTTP 请求辅助逻辑与响应信封假设。
- `frontend/app/src/features/auth/auth-context.tsx` — 会话初始化与刷新流程。
- `frontend/app/src/features/transactions/transactions-api.ts` — 面向交易的核心客户端端点。
- `frontend/app/src/features/management/management-api.ts` — 导入/导出与 PAT 客户端端点。
- `frontend/app/src/features/reporting/reporting-api.ts` — 仪表盘/分析数据端点。

## 发生重大修改前值得查看的现有文档

- `docs/superpowers/specs/2026-03-19-xledger-backend-architecture-tech-solution-design.md`
- `docs/superpowers/specs/2026-03-19-plan-a-auth-spec.md`
- `docs/superpowers/specs/2026-03-19-plan-b-accounting-spec.md`
- `docs/superpowers/specs/2026-03-19-plan-c-classification-spec.md`
- `docs/superpowers/specs/2026-03-19-plan-d-reporting-spec.md`
- `docs/superpowers/specs/2026-03-19-plan-e-portability-spec.md`
- `docs/superpowers/specs/2026-03-19-plan-f-automation-spec.md`

当某项改动会跨多个子系统影响领域规则或 API 形态时，请使用这些文档作为参考。
