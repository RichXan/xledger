# Xledger AI E2E SOP (Midscene)

本文定义 AI Agent 在本仓库执行端到端业务验证的标准步骤。目标是让任意一次自动化执行都可复现、可追踪、可回归。

## 1. 目标范围

本 SOP 覆盖以下核心业务链路（v1 核心闭环）：

1. 用户会话建立（API 注册/登录 + 前端会话注入）
2. 账户与账本管理（创建账户、创建账本）
3. 交易录入（支出 + 收入）
4. 统计口径验证（`total_assets`、`income`、`expense`、`net`）
5. 设置能力验证（创建 PAT）

## 2. 前置条件

1. 已安装并可用：
   - Docker / Docker Compose
   - Node.js 20+
   - pnpm
2. 可用 Midscene 模型配置（二选一）：
   - 方式 A：常规模型网关（`MIDSCENE_MODEL_*`）
   - 方式 B：Codex App Server（推荐在 Codex 环境）
3. 端口未冲突：`4173`、`8080`、`5432`、`6379`

## 3. 目录与职责

- `e2e-tests/playwright.config.ts`：Playwright + Midscene 运行配置
- `e2e-tests/e2e/fixture.ts`：Midscene Playwright fixture 注入
- `e2e-tests/e2e/helpers/*`：会话与 API 校验辅助代码
- `e2e-tests/e2e/accounts.spec.ts`：账户与账本流程
- `e2e-tests/e2e/account-ledger-management.spec.ts`：账户与账本管理（编辑/删除）
- `e2e-tests/e2e/transactions.spec.ts`：交易录入流程
- `e2e-tests/e2e/transfer-lifecycle.spec.ts`：转账创建/编辑/删除流程
- `e2e-tests/e2e/reporting.spec.ts`：统计与分析流程
- `e2e-tests/e2e/classification-export.spec.ts`：分类标签与导出流程
- `e2e-tests/e2e/pat.spec.ts`：PAT 流程
- `e2e-tests/e2e/quality.visual.spec.ts`：视觉基线回归（`@visual`）
- `e2e-tests/e2e/quality.a11y.spec.ts`：可访问性审计（`@a11y`）

## 4. 执行步骤（严格顺序）

### Step 1: 启动服务栈

在仓库根目录执行：

```bash
cd /home/xan/paramita/xledger
cd deploy
docker compose up -d --build
docker compose ps
```

服务必须处于 `healthy` 或 `running`，否则不得进入下一步。

### Step 2: 健康检查

```bash
curl -fsS http://127.0.0.1:8080/healthz
curl -fsS http://127.0.0.1:4173/
```

检查失败时：

1. `docker compose logs --tail 200 xledger-backend`
2. `docker compose logs --tail 200 xledger-frontend`
3. 修复后重新执行 Step 1

### Step 3: 安装 E2E 依赖

```bash
cd /home/xan/paramita/xledger/e2e-tests
pnpm install
pnpm run e2e:install
```

### Step 4: 配置 Midscene 模型

#### 方案 A：Codex App Server（推荐）

```bash
export MIDSCENE_MODEL_BASE_URL="codex://app-server"
export MIDSCENE_MODEL_NAME="gpt-5.4"
export MIDSCENE_MODEL_FAMILY="gpt-5"
```

#### 方案 B：OpenAI-compatible 模型网关

```bash
export MIDSCENE_MODEL_BASE_URL="https://<your-endpoint>/v1"
export MIDSCENE_MODEL_API_KEY="<your-api-key>"
export MIDSCENE_MODEL_NAME="<your-model-name>"
export MIDSCENE_MODEL_FAMILY="<your-model-family>"
```

### Step 5: 执行端到端测试

```bash
cd /home/xan/paramita/xledger/e2e-tests
E2E_BASE_URL="http://127.0.0.1:4173" \
E2E_API_BASE_URL="http://127.0.0.1:8080/api" \
pnpm run e2e
```

CI 并行执行（默认配置）：

1. `CI=true` 时 `playwright.config.ts` 使用 `workers=4`
2. 当前 7 个业务 spec 可并行执行，无共享用户数据依赖
3. `@visual` 与 `@a11y` 质量测试默认不并入业务流水线，按需单独执行：
   - `pnpm run e2e:visual`
   - `pnpm run e2e:visual:update`
   - `pnpm run e2e:a11y`

### Step 6: 结果归档与判定

执行后必须检查并记录：

1. 命令退出码（0 为通过）
2. `e2e-tests/playwright-report/`
3. `e2e-tests/midscene_run/report/`
4. 失败时对应截图、trace、Midscene 报告片段

## 5. 业务验证断言基线

本测试默认创建：

- 1 个账户（初始余额 `1000`）
- 1 笔支出（`120`）
- 1 笔收入（`2500`）

统计接口期望：

- `total_assets = 1000`（交易未绑定账户，不影响资产）
- `income = 2500`
- `expense = 120`
- `net = 2380`

如果实际不满足上述口径，判定为业务回归或规则偏移，需阻断合并。

## 6. 失败处理规范

1. 先判断环境问题还是业务问题：
   - 环境问题：服务不可达、模型不可达、浏览器依赖缺失
   - 业务问题：断言失败、接口返回与预期不一致
2. 重试策略：
   - 同一失败允许重跑 1 次
   - 连续失败必须进入日志分析，不得无限重试
3. 产出最小诊断包：
   - 失败步骤
   - 关键错误信息
   - 对应报告文件路径
   - 建议修复方向

## 7. 清理步骤

如需回收环境：

```bash
cd /home/xan/paramita/xledger/deploy
docker compose down
```

如需彻底清空数据卷：

```bash
docker compose down -v
```
