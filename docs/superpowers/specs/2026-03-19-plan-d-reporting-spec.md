# Plan-D 规格：统计与 Dashboard

> 作用：仅用于统计聚合能力实现计划输入

## 1. 目标

- 提供统一口径的概览、趋势、分类统计。
- 保证报表与交易明细可对账。

## 2. 核心流程

1. `overview` 聚合总资产与月度收支。
2. `trend` 按时间粒度聚合收支趋势。
3. `category` 按分类聚合支出结构。

## 3. 接口范围

| Endpoint | Method | Auth | Success | Error |
|---|---|---|---|---|
| `/api/stats/overview` | GET | Access/PAT | 总资产+月度汇总 | `STAT_QUERY_INVALID`, `STAT_TIMEOUT` |
| `/api/stats/trend` | GET | Access/PAT | 趋势序列 | `STAT_QUERY_INVALID`, `STAT_TIMEOUT` |
| `/api/stats/category` | GET | Access/PAT | 分类聚合 | `STAT_QUERY_INVALID`, `STAT_TIMEOUT` |

## 4. 数据与约束

- `total_assets` 始终按全账户余额计算，不受账本筛选影响。
- `account_id=NULL` 交易计入收入/支出/分类/趋势，不计入总资产。
- 默认收支统计排除转账对冲；不提供“包含转账”开关。
- 时间边界按用户时区聚合，未设置时区使用 UTC+8。
- 空区间返回 0 值序列，不返回 null 占位。

## 5. 异常与降级

- 聚合超时返回 `STAT_TIMEOUT`，不返回部分结果。
- 参数非法返回 `STAT_QUERY_INVALID`。
- 下游 DB 压力过高时允许只读降级（拒绝复杂窗口查询）。

## 6. 验收要点

- Overview 与交易明细口径一致。
- 分类聚合对已归档/已删除但仍被引用分类保持历史名称语义。
- 性能阈值：1k 交易规模 P95 <= 1 秒。

### Out of Scope

- 交易写入、导入导出。
