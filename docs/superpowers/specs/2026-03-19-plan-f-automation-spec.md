# Plan-F 规格：可选增强（AI + Shortcuts）

> 作用：仅用于可选增强能力实现计划输入

## 1. 目标

- 提供不阻塞核心链路的 AI 增强能力。
- 通过 n8n 编排实现 Shortcuts 快记与 CSV 映射建议。

## 2. 核心流程

1. Shortcuts 调用 n8n webhook，携带 PAT 与 `X-Idempotency-Key`。
2. n8n 校验 PAT，调用 LLM 解析意图。
3. n8n 调用 Go 业务 API 创建交易。
4. 解析异常或账户不唯一时按规则降级。

## 3. 接口范围

| Endpoint | Method | Auth | 幂等 | Success | Error |
|---|---|---|---|---|---|
| `{n8n}/webhook/quick-entry` | POST | PAT(Bearer) | 是（24h） | 结构化记账确认 | `QE_PAT_INVALID`, `QE_LLM_UNAVAILABLE`, `QE_PARSE_FAILED`, `QE_TIMEOUT` |

## 4. 数据与约束

- PAT 通过 `Authorization: Bearer <token>` 传递。
- 幂等键去重窗口 24 小时（`pat_id + normalized_text + key`）。
- n8n -> LLM 超时 5 秒，最多重试 2 次；超时后终止落账。
- `account_hint` 不唯一时，降级 `account_id=NULL`，并附带提示字段。

## 5. 异常与降级

- LLM 不可用：返回 `QE_LLM_UNAVAILABLE`，不写入交易。
- n8n 不可用：仅增强能力不可用，核心链路继续可用。
- Go API 返回验证错误：透传语义化错误，不做二次猜测。

## 6. 验收要点

- 可选增强关闭时，核心链路完全不受影响。
- 快记失败返回明确错误，数据不污染。
- 幂等重放不重复落账，成功返回可复现。

### Out of Scope

- 核心手动记账流程改造、Go 直连 `/api/quick-entry` 端点。
