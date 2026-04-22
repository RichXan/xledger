# AGENTS.md（根目录）

本文件为 AI 编码代理在此仓库中工作时提供指引。

## 快速参考

```bash
# 完整部署（在 /deploy 目录执行）
docker compose up -d --build                      # 启动完整部署栈
docker compose ps                                 # 查看服务状态
docker compose down                               # 停止部署栈

# 后端（在 /backend 目录执行）
go test ./...                                    # 全量测试
go test ./internal/accounting -run TestCreate     # 单个测试
go fmt ./... && go vet ./...                      # 格式化 + 静态检查

# 前端（在 /frontend/app 目录执行）
pnpm run dev                                     # 开发服务器
pnpm run test                                    # 全量测试
pnpm run test -- src/App.test.tsx                # 单个测试
pnpm run check                                   # Lint + 测试

# 根目录 Makefile
make setup | make backend | make frontend | make migrate-up | make clean
```

## 层级说明

| 路径 | 领域 |
|------|------|
| `backend/AGENTS.md` | 后端概览、命令、约定 |
| `backend/internal/accounting/AGENTS.md` | 账本、账户、交易、转账 |
| `backend/internal/auth/AGENTS.md` | 认证、会话、OAuth |
| `backend/internal/bootstrap/AGENTS.md` | DI 装配、路由、基础设施 |
| `backend/internal/classification/AGENTS.md` | 分类、标签 |
| `backend/internal/portability/AGENTS.md` | 导入/导出、PAT |
| `backend/internal/reporting/AGENTS.md` | 统计、趋势 |
| `frontend/app/src/features/AGENTS.md` | Feature 模式、API/hooks 约定 |

## 代码风格

### TypeScript/React
- Imports：使用 `@/` 别名，按 external → internal → relative 分组
- Components：使用 PascalCase；函数使用 camelCase
- Types：对象优先使用 `interface`，联合类型使用 `type`
- Hooks：使用 `use` 前缀，并与功能代码同目录放在 `*-hooks.ts`
- Exports：优先使用命名导出，只有 page 允许默认导出

### Go
- 错误码：使用包内常量（如 `LEDGER_INVALID`）
- 错误类型：使用带 code 的 `contractError`
- Packages：按领域组织
- Testing：采用表驱动，测试文件与实现文件同目录

## 反模式

1. **不要使用 `as any` 或 `@ts-ignore`** — 应正确修复类型问题
2. **不要写空的 catch 块** — 必须显式处理错误
3. **迁移文件是追加式的** — 不要修改已有迁移
4. **`PRD.md` 是事实来源** — 改动记账逻辑前先阅读

## 关键文件

| 文件 | 用途 |
|------|------|
| `PRD.md` | 产品需求与记账规则 |
| `backend/openapi/openapi.yaml` | API 合约 |
| `frontend/app/src/lib/api.ts` | HTTP 客户端、响应类型 |
| `backend/internal/bootstrap/http/router.go` | 路由注册、DI |
| `deploy/docker-compose.yaml` | 完整部署栈入口 |
| `deploy/README.md` | 部署说明 |
| `Makefile` | 常用开发命令 |

## 环境变量

必需：`SMTP_HOST`、`AUTH_CODE_PEPPER`
可选：`DATABASE_URL`（留空则使用内存模式）、`REDIS_URL`
