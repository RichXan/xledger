# Xledger 功能增强实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现 5 个功能子系统：i18n、PWA 离线+推送、预算系统、AI 记账、导入工具完善

**Architecture:** 5 个独立子系统，通过统一 API 整合。前端 React + TypeScript，后端 Go + Gin + GORM + PostgreSQL。

**Tech Stack:** react-i18next, Dexie.js (IndexedDB), Web Push API, n8n, Apple Shortcuts

---

## 子系统文件影响总览

### 前端 (`frontend/app/`)

| 文件 | 改动 | 属于子系统 |
|------|------|-----------|
| `package.json` | 添加 i18n + dexie 依赖 | i18n + PWA |
| `src/main.tsx` | 初始化 i18next | i18n |
| `src/i18n/index.ts` | 新建 - i18n 配置 | i18n |
| `src/locales/zh/translation.json` | 新建 - 中文资源 | i18n |
| `src/locales/en/translation.json` | 新建 - 英文资源 | i18n |
| `src/lib/api.ts` | 添加 i18n language header | i18n |
| `src/App.tsx` | 添加 /settings/notifications, /budgets 路由 | 预算 + PWA |
| `src/components/layout/mobile-nav.tsx` | 新建 - 移动端底部导航 | PWA |
| `src/components/layout/offline-banner.tsx` | 新建 - 离线横幅 | PWA |
| `src/features/offline/offline-store.ts` | 新建 - IndexedDB 层 | PWA |
| `src/features/offline/offline-queue.ts` | 新建 - 离线队列 | PWA |
| `src/features/notifications/push-manager.ts` | 新建 - Web Push 管理 | PWA |
| `src/features/notifications/preference-store.ts` | 新建 - 通知偏好 | PWA |
| `src/features/budgets/budgets-api.ts` | 新建 - 预算 API | 预算 |
| `src/features/budgets/budgets-hooks.ts` | 新建 - 预算 React Query | 预算 |
| `src/pages/budgets-page.tsx` | 新建 - 预算列表页 | 预算 |
| `src/pages/budget-form-page.tsx` | 新建 - 创建/编辑预算页 | 预算 |
| `src/pages/budget-alerts-page.tsx` | 新建 - 告警历史页 | 预算 |
| `src/pages/notification-prefs-page.tsx` | 新建 - 通知偏好页 | PWA |
| `src/pages/onboarding-pwa-page.tsx` | 新建 - iOS 安装引导页 | PWA |
| `public/sw.js` | 增强 - 离线缓存策略 | PWA |
| `public/manifest.webmanifest` | 增强 - push permission | PWA |
| `src/features/transactions/transactions-api.ts` | 添加语言 header | i18n |

### 后端 (`backend/`)

| 文件 | 改动 | 属于子系统 |
|------|------|-----------|
| `internal/budget/handler.go` | 新建 - 预算 API handler | 预算 |
| `internal/budget/service.go` | 新建 - 预算业务逻辑 | 预算 |
| `internal/budget/repository.go` | 新建 - 预算数据访问接口 | 预算 |
| `internal/budget/repository_postgres.go` | 新建 - Postgres 实现 | 预算 |
| `internal/budget/models.go` | 新建 - 预算数据模型 | 预算 |
| `internal/budget/alert_service.go` | 新建 - 提醒触发逻辑 | 预算 |
| `internal/bootstrap/http/budget_wiring.go` | 新建 - 预算依赖注入 | 预算 |
| `internal/bootstrap/http/router.go` | 注册预算路由 | 预算 |
| `migrations/YYYYMMDDDD_create_budgets.sql` | 新建 - 预算表迁移 | 预算 |
| `migrations/YYYYMMDDDD_create_budget_alerts.sql` | 新建 - 告警表迁移 | 预算 |
| `migrations/YYYYMMDDDD_create_notification_prefs.sql` | 新建 - 通知偏好表 | PWA |
| `internal/auth/handler.go` | 添加 language 字段到响应 | i18n |
| `internal/bootstrap/http/router.go` | 添加 language 到 /auth/me | i18n |
| `internal/portability/import_preview_service.go` | 增强 - 支付宝格式检测 | 导入 |
| `internal/portability/csv.go` | 增强 - 支付宝字段映射 | 导入 |
| `internal/portability/handler.go` | 增强 - 导入 API 扩展 | 导入 |
| `internal/accounting/handler.go` | 添加 language 到用户上下文 | i18n |
| `internal/bootstrap/http/accounting_wiring.go` | 添加 BudgetService 依赖 | 预算 |

---

## Phase 1：基础设施（可并行执行）

### 子计划文件

- `docs/superpowers/plans/2026-04-04-01-i18n-implementation-plan.md`
- `docs/superpowers/plans/2026-04-04-02-pwa-enhancement-implementation-plan.md`

### 执行顺序

1. **i18n** (独立，可最先开始)
2. **PWA 增强** (manifest + service worker 基础)

---

## Phase 2：核心功能

### 子计划文件

- `docs/superpowers/plans/2026-04-04-03-budget-system-implementation-plan.md`
- `docs/superpowers/plans/2026-04-04-04-ai-bookkeeping-implementation-plan.md`

### 执行顺序

1. **预算系统** (后端先，再前端)
2. **AI 记账** (n8n workflow 先于 Go API)

---

## Phase 3：完善

### 子计划文件

- `docs/superpowers/plans/2026-04-04-05-import-tool-enhancement-plan.md`

### 执行顺序

1. **导入工具完善** (支付宝 CSV + ezbk CSV)

---

## 依赖关系图

```
Phase 1 (基础设施)
    │
    ├── i18n ─────────────────────────────┐
    │                                      │
    └── PWA (manifest + SW 基础) ─────────┤
                                           │
Phase 2 (核心功能)                          │
    │                                      │
    ├── 预算系统 ◄──────────────────────────┤
    │   (依赖 i18n 翻译资源)                │
    │                                      │
    └── AI 记账 Go API ◄───────────────────┤
        (依赖 预算系统的 category_id)       │
                                           │
Phase 3 (完善)                             │
    │                                      │
    └── 导入工具 ◄─────────────────────────┘
        (依赖 预算系统的 category_id)
```

---

## 测试策略

| 子系统 | 测试方式 |
|--------|----------|
| i18n | 手动切换语言，检查所有页面文本是否正确显示 |
| PWA | 浏览器 DevTools → Application → Service Workers；断网测试 |
| 预算系统 | 单元测试 + 集成测试（创建预算 → 记账 → 验证告警触发） |
| AI 记账 | n8n workflow 测试 + API 集成测试 |
| 导入工具 | 上传支付宝 CSV 文件，验证正确条数导入 |
