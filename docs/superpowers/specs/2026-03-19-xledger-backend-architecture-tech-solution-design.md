# Xledger 后端架构与技术方案设计说明

> 日期：2026-03-19  
> 状态：已完成设计评审（用户对话确认）  
> 范围：`v1 核心 + 可选增强（AI/Shortcuts）`

---

## 1. 背景与目标

基于 `PRD.md`，为 Xledger 输出两份后端文档并落盘到 `backend/`：

1. 系统架构设计文档（架构主文档）。
2. 技术方案文档（按业务域拆分）。

本轮目标是形成可评审、可作为后续计划输入的文档，不进入代码实现。

本文件定位为“父级规格（Parent Spec）”，只负责统一规则与拆分边界。  
进入实现计划时，必须按业务域拆分为独立计划单元，禁止把 A-F 全部内容合并成一个大计划。

---

## 2. 设计决策记录

### 2.1 范围决策

- 采用范围：`v1 核心 + 可选增强`。
- 不做全路线图实现细节（v1.1/v2 仅作为演进边界说明）。

### 2.2 深度决策

- 采用深度：设计评审级。
- 特征：聚焦边界、流程、约束、异常、验收要点；不写字段级实现细节。

### 2.3 文档组织决策

用户最终选择的组织方式：

- 架构主文档统一系统视图与跨域规则。
- 技术方案按业务域拆分，逐域描述落地方法。

该方式兼顾“全局一致性”和“模块实现可读性”。

---

## 3. 方案概览

### 3.1 架构主文档内容边界

架构主文档覆盖：

- 上下文架构与分层。
- 领域边界（Auth / Accounting / Classification / Reporting / Portability / Automation）。
- 核心业务口径统一（总资产、未绑定账户交易、转账对冲规则）。
- 跨域能力（安全、可观测、部署、降级策略）。
- 风险与演进边界。

### 3.2 业务域技术方案边界

技术方案文档按以下业务域展开：

- A 认证与会话。
- B 账本/账户/交易核心。
- C 分类/标签与查询。
- D 统计与 Dashboard。
- E 导入导出与 PAT。
- F 可选增强（n8n + AI + Shortcuts）。

每个域统一结构：目标、核心流程、接口范围、数据约束、异常降级、验收要点。

### 3.3 计划拆分规则（强制）

为避免形成不可执行的“超级计划”，后续实现计划必须按业务域拆分为独立计划单元：

1. Plan-A：认证与会话。
2. Plan-B：账本/账户/交易核心。
3. Plan-C：分类/标签与查询。
4. Plan-D：统计与 Dashboard。
5. Plan-E：导入导出与 PAT。
6. Plan-F（可选）：AI 增强（CSV 映射建议 + Shortcuts 快记）。

依赖顺序：`A -> B -> C -> D -> E -> F`。  
说明：`F` 只能在 `A` 与 `B` 可用后启动；`D` 依赖 `B/C` 的稳定数据语义。

强约束：

- 每次只为一个 Plan 单元生成实现计划与任务列表。
- 不允许跨两个以上 Plan 单元做并行落地，除非无共享写模型。
- 每个 Plan 必须有独立验收与回滚策略。

### 3.4 业务域接口目录（评审级）

| 业务域 | 端点 | Method | Auth | 幂等规则 | 成功语义 | 错误语义 |
|---|---|---|---|---|---|---|
| Auth | `/api/auth/send-code` | POST | 匿名 | 不要求 | 验证码发送触发 | `AUTH_CODE_RATE_LIMIT` / `AUTH_CODE_SEND_FAILED` |
| Auth | `/api/auth/verify-code` | POST | 匿名 | 不要求 | 返回 Access/Refresh | `AUTH_CODE_INVALID` / `AUTH_CODE_EXPIRED` |
| Auth | `/api/auth/google/callback` | GET | 匿名 | 不要求 | 返回 Access/Refresh | `AUTH_OAUTH_FAILED` |
| Auth | `/api/auth/refresh` | POST | Refresh | 不要求 | 刷新 Access（可轮转 Refresh） | `AUTH_REFRESH_REVOKED` / `AUTH_REFRESH_EXPIRED` |
| Auth | `/api/auth/logout` | POST | Access | 不要求 | Refresh 失效 | `AUTH_UNAUTHORIZED` |
| Accounting | `/api/transactions` | POST | Access/PAT | 业务请求不强制幂等 | 创建交易或转账对 | `TXN_VALIDATION_FAILED` / `TXN_INVARIANT_VIOLATION` |
| Accounting | `/api/transactions` | PUT | Access/PAT | 不要求 | 更新交易并回算 | `TXN_NOT_FOUND` / `TXN_CONFLICT` |
| Accounting | `/api/transactions` | DELETE | Access/PAT | 不要求 | 删除交易并回算 | `TXN_NOT_FOUND` / `TXN_CONFLICT` |
| Classification | `/api/categories/*` | GET/POST/PUT/DELETE | Access/PAT | 不要求 | 分类生命周期管理 | `CAT_IN_USE_ARCHIVED` / `CAT_INVALID_PARENT` |
| Reporting | `/api/stats/overview` | GET | Access/PAT | 不适用 | 返回统一口径汇总 | `STAT_QUERY_INVALID` / `STAT_TIMEOUT` |
| Portability | `/api/import/csv/confirm` | POST | Access/PAT | 必须 `X-Idempotency-Key` | 返回导入结果摘要 | `IMPORT_DUPLICATE_REQUEST` / `IMPORT_PARTIAL_FAILED` |
| Portability | `/api/export` | GET | Access/PAT | 不适用 | 返回 CSV/JSON 文件 | `EXPORT_TIMEOUT` / `EXPORT_INVALID_RANGE` |
| Portability | `/api/settings/tokens` | POST/DELETE | Access | 不要求 | PAT 生成/撤销 | `PAT_REVOKED` / `PAT_EXPIRED` |
| Automation(Optional) | `{n8n}/webhook/quick-entry` | POST | PAT | 必须 `X-Idempotency-Key` | 调用成功返回结构化记账结果 | `QE_PAT_INVALID` / `QE_LLM_UNAVAILABLE` / `QE_PARSE_FAILED` |

### 3.5 子规格清单（计划输入）

后续计划必须引用以下子规格（每个子规格只服务一个 Plan 单元）：

- `docs/superpowers/specs/2026-03-19-plan-a-auth-spec.md`
- `docs/superpowers/specs/2026-03-19-plan-b-accounting-spec.md`
- `docs/superpowers/specs/2026-03-19-plan-c-classification-spec.md`
- `docs/superpowers/specs/2026-03-19-plan-d-reporting-spec.md`
- `docs/superpowers/specs/2026-03-19-plan-e-portability-spec.md`
- `docs/superpowers/specs/2026-03-19-plan-f-automation-spec.md`

---

## 4. 核心设计原则

### 4.1 一致性优先

- 账务正确性高于可选能力可用性。
- 转账双边操作必须原子。
- 统计口径通过领域规则统一，不允许在接口层分叉解释。

### 4.2 可降级与可插拔

- 可选增强（n8n/LLM）不可阻断核心记账。
- 外部依赖失败必须返回明确错误，且不写入脏账务数据。

### 4.3 自托管友好

- 核心部署最小闭环可运行。
- 可选增强通过配置启停，默认关闭。

---

## 5. 关键规则（实现计划必须继承）

1. `Transaction.ledger_id` 必填，默认总账本。
2. `Transaction.account_id` 可空，可空时不影响总资产。
3. 总资产固定为所有账户余额总和，不受账本筛选影响。
4. 转账必须使用双边交易模型并成对维护。
5. 分类被引用后删除转归档，历史语义必须可追溯。
6. PAT 仅用于业务 API，哈希存储、可撤销、不可用于签发/刷新认证令牌。

### 5.1 账务不变量澄清表

| 规则 | 约束 | 表现语义 | 计划验收检查 |
|---|---|---|---|
| 未绑定账户交易 | `account_id = NULL` 合法（仅收入/支出） | 出现在交易列表、分类聚合、收支趋势；不进入总资产 | 明细与报表一致；总资产不受影响 |
| 总资产口径 | 总资产 = 全部账户余额和 | 即使携带 `ledger_id` 筛选，总资产值不变 | 同请求下比较有无 `ledger_id` 参数 |
| 转账双边 | `transfer_pair_id` 成对存在 | 默认收支统计排除对冲影响 | 创建/编辑/删除均双边一致 |
| 分类历史保留 | 已引用分类删除转归档 | 历史交易详情、统计、导出保留历史名称 | 删除后回查历史记录名称不丢失 |

### 5.2 并发与幂等规则

- `POST /api/import/csv/confirm` 必须携带 `X-Idempotency-Key`；去重窗口 24 小时（`user_id + path + key`）。
- 导入防重优先级：先做请求级幂等，再做业务级三元组去重；任一命中均不得重复写账。
- `POST {n8n}/webhook/quick-entry` 必须携带 `X-Idempotency-Key`；去重窗口 24 小时（`pat_id + normalized_text + key`）。
- 转账更新必须在同事务中锁定交易对，禁止单边更新提交。
- PAT 撤销传播 SLA：5 秒内全节点生效；超过 5 秒视为鉴权故障。
- 同一请求重放：若命中幂等窗口，返回首次成功结果或标准重复请求错误，不重复写账。

### 5.3 统一验收阈值

- 查询性能：1k 交易规模下，交易筛选与 Dashboard 聚合查询 P95 <= 1 秒。
- 导出性能：10k 交易规模导出耗时 <= 10 秒。
- SMTP/OAuth/n8n/LLM 外部调用超时默认 5 秒，最多 2 次重试（指数退避，仅读操作可重试）。
- 时区口径：默认按用户时区聚合；未设置时区时使用 UTC+8（可配置）。

---

## 6. 产出文件与完成定义

- `backend/系统架构设计文档.md`
- `backend/技术方案文档.md`

完成定义（DoD）：

1. 架构文档包含：分层、边界、数据流、跨域能力、可选增强隔离。
2. 技术方案包含 A-F 全部业务域章节，且每章都具备统一六段结构。
3. 所有规则可追溯到 PRD，不出现互相冲突表述。
4. 文档不含占位语（TODO/TBD/待补充）。
5. 已通过规格审查（spec review）并可进入计划阶段。

---

## 7. 失败模式矩阵（计划阶段必须覆盖）

| 场景 | 预期行为 | 数据写入策略 | 用户可见反馈 |
|---|---|---|---|
| SMTP 邮件发送失败 | 验证码流程失败并可重试 | 不创建会话 | 返回可重试错误 |
| OAuth Provider 超时 | Google 登录失败 | 不创建会话 | 返回外部依赖错误 |
| 转账写入单边失败 | 整体回滚 | 0 写入 | 返回事务失败 |
| 导入任务中断 | 已提交分片按事务边界处理 | 禁止半条交易写入 | 返回成功/失败分项统计 |
| LLM 解析失败（可选） | 快记失败但核心链路可用 | 不写入交易 | 返回明确错误，不做猜测落账 |
| `account_hint` 不唯一（可选） | 安全降级 | 创建 `account_id=NULL` 交易 | 返回“未匹配账户”提示 |

---

## 8. PRD 追溯矩阵（完整）

| PRD 条目 | 关键约束 | 本规格位置 | 后续计划验收项 |
|---|---|---|---|
| FR-Auth1 | 邮箱验证码登录、首登初始化总账本 | 3.4 Auth, 7 | Plan-A 登录链路验收 |
| FR-Auth2 | Google OAuth 登录 | 3.4 Auth, 7 | Plan-A OAuth 回调验收 |
| FR-Auth3 | 登出黑名单 | 5.2, 7 | Plan-A Token 失效验收 |
| FR-A0 | 无账户可记账 | 5.1 未绑定账户交易 | Plan-B 创建交易验收 |
| FR-A1 | 账本管理、总账本不可删 | 5, 6 | Plan-B 账本生命周期验收 |
| FR-A2 | 账户管理与余额语义 | 5.1 总资产口径 | Plan-B 账户余额验收 |
| FR-A3 | 收入/支出/转账 | 5.1 转账双边, 5.2 | Plan-B 转账事务验收 |
| FR-A4 | 交易编辑删除自动回算 | 5.2, 7 | Plan-B 回算一致性验收 |
| FR-B1 | 分类两级结构、归档语义 | 3.4 Classification, 5.1 | Plan-C 分类生命周期验收 |
| FR-B2 | 标签多选 | 3.4 Classification | Plan-C 标签管理验收 |
| FR-B3 | 多条件筛选性能 | 3.4 Classification, 9 | Plan-C 查询性能验收 |
| FR-C1 | Dashboard 总资产口径固定 | 5.1 总资产口径 | Plan-D 概览一致性验收 |
| FR-C2 | 趋势和分类统计 | 3.4 Reporting, 7 | Plan-D 聚合统计验收 |
| FR-D1 | CSV/JSON 导出与历史分类保留 | 3.4 Portability, 5.1 | Plan-E 导出正确性验收 |
| FR-D2 | CSV 导入映射与去重 | 3.4 Portability, 5.2 | Plan-E 导入幂等验收 |
| FR-D3 | 业务 REST API + OpenAPI | 3.4 全域接口目录 | Plan-E 接口文档验收 |
| FR-D4 | PAT 生成/撤销与作用范围 | 3.4 Portability, 5.2 | Plan-E PAT 权限验收 |
| FR-X1 | AI 映射建议可选 | 3.3, 7 | Plan-F 可选启停验收 |
| FR-X2 | Shortcuts AI 快记可选 | 3.4 Automation, 7 | Plan-F 快记降级验收 |

---

## 9. 验收标准（文档层）

- 两份文档覆盖 PRD 约束且无明显冲突。
- 明确区分核心链路与可选增强。
- 各业务域边界清晰，可独立理解和后续拆分计划。
- 不含 TODO/TBD/占位段落。

---

## 10. 后续阶段入口

完成本文档后，下一步进入 `writing-plans` 阶段，输出实现计划（任务分解、里程碑、验收与回滚步骤）。
