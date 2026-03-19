# Plan-E 规格：导入导出与 PAT

> 作用：仅用于可携带与开放能力实现计划输入

## 1. 目标

- 提供可迁移的数据导入导出能力。
- 提供可控的长期凭证访问能力（PAT）。

## 2. 核心流程

1. CSV 上传并生成映射预览。
2. 用户确认导入，执行防重检查后入库。
3. 按范围导出 CSV/JSON。
4. PAT 生成、列出、撤销。

## 3. 接口范围

| Endpoint | Method | Auth | 幂等 | Success | Error |
|---|---|---|---|---|---|
| `/api/import/csv` | POST | Access/PAT | 否 | 导入预览 | `IMPORT_INVALID_FILE` |
| `/api/import/csv/confirm` | POST | Access/PAT | 是（`X-Idempotency-Key`） | 导入摘要 | `IMPORT_DUPLICATE_REQUEST`, `IMPORT_PARTIAL_FAILED` |
| `/api/export` | GET | Access/PAT | 不适用 | 文件输出 | `EXPORT_INVALID_RANGE`, `EXPORT_TIMEOUT` |
| `/api/settings/tokens` | GET/POST/DELETE | Access | 不适用 | PAT 生命周期操作 | `PAT_EXPIRED`, `PAT_REVOKED` |

## 4. 数据与约束

- 导入防重采用双层规则：
  1) 请求级幂等：`X-Idempotency-Key`，24 小时窗口；
  2) 业务级去重：（金额 + 日期 + 描述）三元组。
- 双层规则优先级：先请求级，后业务级；任一命中均不重复写账。
- 导出保留历史分类名称语义。
- PAT 仅存哈希，默认有效期 90 天，可撤销。
- PAT 不可调用认证签发/刷新接口。

## 5. 异常与降级

- 导入部分失败返回分项结果（成功/跳过/失败 + 原因）。
- 导出超时返回 `EXPORT_TIMEOUT`，允许客户端重试。
- PAT 撤销传播超 5 秒触发告警并进入严格鉴权模式。

## 6. 验收要点

- 导入结果可解释且可追溯。
- 10k 交易导出耗时 <= 10 秒。
- 撤销 PAT 后业务 API 请求在 5 秒内失效。

### Out of Scope

- AI 映射建议、认证令牌签发。
