# Plan-B 规格：账本/账户/交易核心

> 作用：仅用于核心账务能力实现计划输入

## 1. 目标

- 建立收入/支出/转账可持续运行的账务核心。
- 保证余额回算一致性与转账原子性。
- 支持“无账户也可记账”的低摩擦体验。

## 2. 核心流程

1. 普通交易：校验 -> 事务写入 -> 回算余额。
2. 转账交易：生成双边交易 -> 绑定 `transfer_pair_id` -> 同事务提交。
3. 编辑/删除：按交易类型重算账户余额与统计输入。

## 3. 接口范围

| Endpoint | Method | Auth | Success | Error |
|---|---|---|---|---|
| `/api/ledgers` | GET/POST/PUT/DELETE | Access/PAT | 账本结果 | `LEDGER_INVALID`, `LEDGER_DEFAULT_IMMUTABLE` |
| `/api/accounts` | GET/POST/PUT/DELETE | Access/PAT | 账户结果 | `ACCOUNT_INVALID`, `ACCOUNT_NOT_FOUND` |
| `/api/transactions` | GET/POST/PUT/DELETE | Access/PAT | 交易结果 | `TXN_VALIDATION_FAILED`, `TXN_CONFLICT`, `TXN_NOT_FOUND` |

## 4. 数据与约束

- `ledger_id` 必填；默认总账本不可删除。
- `account_id` 可空（仅收入/支出）；转账必须双账户。
- 允许跨账本转账，统计口径按交易所属账本分别聚合。
- `TXN_CONFLICT` 触发条件：版本号冲突、交易对锁冲突、双边不一致。

## 5. 异常与降级

- 任一转账边写入失败：整体回滚。
- 并发编辑冲突：返回 `TXN_CONFLICT`，客户端刷新后重试。
- 回算失败：拒绝提交并告警，不允许部分成功。

## 6. 验收要点

- 转账双边记录始终成对。
- 编辑/删除后余额与交易明细一致。
- 无账户也能创建第一笔交易。

### Out of Scope

- 统计图表渲染、导入导出、PAT 管理、AI 快记编排。
