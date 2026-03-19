# Xledger - 智能个人财务管理平台

## Product Requirements Document (PRD)

> 版本: v0.6 | 日期: 2026-03-17 | 状态: 评审修订版

---

## 1. Executive Summary

### 1.1 Problem Statement

现有个人记账产品普遍存在三类问题：

- **记账成本高**：用户需要手动录入、切换账户、补分类，长期坚持成本高。
- **数据主权弱**：商业记账 App 普遍限制导出、迁移和自托管。
- **跨端与分析能力不足**：移动端、桌面端和 Web 体验割裂，统计结果难以支撑复盘。

### 1.2 Proposed Solution

**Xledger** 是一款开源、自托管优先的个人财务管理平台。`v1` 聚焦“可持续手动记账”这一核心闭环，优先交付：

1. **低摩擦记账**：支持用户零配置开始记账，不强依赖先建账户。
2. **数据可携带**：支持 Docker 自托管、CSV/JSON 导出、REST API。
3. **统一体验**：通过 PWA 提供统一的 Web/桌面/移动端基础体验。

### 1.3 v1 Success Criteria

| 指标                   | 目标值     | 口径             |
| -------------------- | ------- | -------------- |
| 首次安装后 10 分钟内完成首笔记账比例 | >= 70%  | 产品埋点           |
| 手动记一笔账平均耗时           | <= 15 秒 | 从打开记账入口到提交成功   |
| 首周完成 10 笔以上记账的用户占比   | >= 40%  | 产品埋点           |
| Docker 安装文档一次成功率     | >= 90%  | 新环境安装测试 + 社区反馈 |
| CSV/JSON 导出成功率       | >= 99%  | 后端日志           |

---

## 2. User Experience & Functionality

### 2.1 User Personas

#### Persona 1: 普通个人用户

- 目标：快速记录日常收支，月底看懂钱花到了哪里。
- 关注点：录入速度、移动端体验、统计是否清晰。

#### Persona 2: 自由职业者 / 副业用户

- 目标：按账户、标签、时间维度追踪多个收入和支出场景。
- 关注点：多账户、标签、导出、基础统计。

#### Persona 3: 技术型自托管用户

- 目标：自己部署、自己持有数据、通过 API 扩展能力。
- 关注点：Docker 部署、数据导出、REST API、权限控制。

### 2.2 Product Principles

- **MVP 只解决基础记账闭环，不追求一次覆盖完整财务平台。**
- **AI 是增强项，不是 v1 核心链路依赖。**
- **账户不是记账前置条件，用户应能零配置开始。**
- **先保证单人账本体验成立，再扩展协作与共享。**
- **先统一核心口径，再扩展复杂多币种与自动化能力。**

### 2.3 Core Domain Model

#### 核心模型：全局账户 + 账本分流

```text
用户
├── 账户（全局，负责余额）
│   ├── 招商银行储蓄卡
│   ├── 微信零钱
│   └── 支付宝
│
├── 总账本（注册时自动创建，不可删除）
│   ├── 午饭 -25（账户=微信零钱）
│   ├── 工资 +15000（账户=招商银行）
│   └── 打车 -35（账户未指定）
│
├── 自由职业账本（用户自建）
└── 旅行基金账本（用户自建）
```

#### 核心规则

| 规则 | 说明 |
|------|------|
| 账户是全局的 | 一个账户属于用户，不属于某个账本 |
| 交易必须属于一个账本 | 默认写入“总账本” |
| 账户可选 | `account_id = NULL` 时仅影响账本统计，不影响任何账户余额 |
| 总账本不可删除 | 注册时自动创建，作为默认账本 |
| 分类为用户维度 | 系统提供默认分类模板，注册时复制到用户分类集中；用户可自行删除或调整，不影响其他用户 |
| 分类支持父子结构 | `v1` 支持一级分类 + 一级子分类的两级结构 |
| 统计可按账本筛选 | 支持单账本视图与全部账本汇总视图 |
| 总资产口径固定 | 总资产 = 所有账户当前余额之和，不受账本筛选影响；`account_id = NULL` 的交易计入账本统计，但不计入总资产 |

#### 关键业务口径

- **账户余额** = `initial_balance + SUM(该账户所有交易 amount)`。
- **总资产** = 所有已绑定账户余额按统一口径汇总后的结果。
- **未绑定账户的交易**：
  - 计入收入、支出、分类、账本统计；
  - 不计入总资产；
  - 不影响任何账户余额。

这一定义是 v1 的强约束。Dashboard、导入、导出、统计和验收测试必须使用同一口径。

### 2.4 Key User Flows

#### Flow 1: 首次使用

1. 用户通过邮箱验证码或 Google 登录。
2. 系统自动创建“总账本”。
3. 系统为该用户初始化一套可编辑的默认分类与子分类。
4. 用户无需先创建资金账户，即可记录第一笔交易。
5. 首页展示本月收支摘要和最近交易。
6. 用户按需补充账户、标签、自定义账本。

#### Flow 2: 日常记账

1. 用户点击首页快捷入口进入记账页。
2. 输入金额并选择交易类型：支出、收入、转账。
3. 选择分类；账户、标签、备注、日期、账本为可选或有默认值。
4. 提交后立即更新对应统计；若选择了账户，同时更新账户余额。

#### Flow 3: Apple Shortcuts AI 记账（可选增强）

1. 用户在设置页生成 Personal Access Token。
2. 导入预制的 Apple 快捷指令，填入服务器地址和 Token。
3. 用户通过 Siri、双击背面或手动运行快捷指令输入自然语言记账文本。
4. 快捷指令将文本 POST 到 `POST {n8n_url}/webhook/quick-entry`。
5. n8n 工作流验证 PAT，调用 LLM 解析文本，再调用 Go API 创建交易。
6. 创建交易时默认写入“总账本”；若解析结果包含 `account_hint` 且唯一匹配到用户已有账户，则写入 `account_id`；若未匹配成功或匹配不唯一，则 `account_id = NULL`。
7. 快捷指令显示结构化确认结果，例如“已记录：午饭 ¥25 → 餐饮”。

#### Flow 4: 数据导出

1. 用户选择导出格式与时间范围。
2. 系统生成导出文件。
3. 用户下载文件，在本地做备份或迁移。

### 2.5 User Stories & Acceptance Criteria

#### Epic Auth

| ID | User Story | Acceptance Criteria |
|----|------------|---------------------|
| FR-Auth1 | 作为用户，我想通过邮箱验证码注册和登录 | - 输入邮箱后发送 6 位数字验证码，有效期 10 分钟<br>- 同一邮箱 60 秒内不可重发<br>- Go 后端生成验证码并存储 Redis，直接通过 SMTP 账号密码发送邮件<br>- 验证通过后签发 JWT（Access Token 15min，Refresh Token 7d）<br>- 新邮箱首次验证后自动创建 `User` 和“总账本”，不自动创建任何资金账户 |
| FR-Auth2 | 作为用户，我想通过 Google 账号一键登录 | - 完整 OAuth 2.0 授权流程<br>- 仅请求 `email` 和 `profile` 权限<br>- 授权成功后自动创建或关联已有用户，并签发 JWT |
| FR-Auth3 | 作为用户，我想安全退出登录 | - 退出时 Refresh Token 加入 Redis 黑名单<br>- 客户端本地 Token 被清除 |

#### Epic A: 账本、账户与账务核心

| ID | User Story | Acceptance Criteria |
|----|------------|---------------------|
| FR-A0 | 作为新用户，我想在注册后立即开始记账，而不是先配置账户 | - 注册成功后自动创建“总账本”（`is_default=true`）<br>- 用户无需先创建资金账户即可创建第一笔交易<br>- `account_id = NULL` 的交易仅更新账本统计 |
| FR-A1 | 作为用户，我想管理多个账本，以便按场景隔离统计 | - 支持创建、编辑、删除自定义账本<br>- 总账本不可删除<br>- Dashboard 支持按账本筛选与全部账本汇总 |
| FR-A2 | 作为用户，我想管理多个资金账户，以便记录真实余额 | - 支持创建、编辑、归档账户<br>- 每个账户必须有名称、类型、初始余额<br>- 账户不绑定账本<br>- 账户列表默认展示当前余额 |
| FR-A3 | 作为用户，我想记录收入、支出和转账 | - 支持三种交易类型：收入、支出、转账<br>- 账本必填，默认“总账本”<br>- 普通收入/支出中账户可为空<br>- 转账必须同时选择转出账户和转入账户 |
| FR-A4 | 作为用户，我想查看和编辑历史交易，以便修正错误 | - 支持按时间倒序浏览交易<br>- 支持按账本、账户、分类、标签筛选<br>- 支持编辑和删除交易<br>- 编辑和删除后，相关统计与账户余额自动回算 |

#### Epic B: 分类、标签与查询

| ID | User Story | Acceptance Criteria |
|----|------------|---------------------|
| FR-B1 | 作为用户，我想给交易选择分类，以便分析支出用途 | - 系统为每个新用户提供一套默认一级/二级分类模板<br>- 默认分类在用户侧可编辑、可删除，删除后不影响其他用户的默认分类<br>- 支持用户新增、编辑、删除、归档自定义分类<br>- 分类支持父分类与一层子分类结构<br>- 分类被删除或归档后，历史交易仍保留原分类引用与名称展示；交易编辑时若原分类已不可用，用户需重新选择当前可用分类 |
| FR-B2 | 作为用户，我想给交易添加标签，以便按项目筛选 | - 一笔交易支持 0 到多个标签<br>- 标签支持新增、编辑、删除 |
| FR-B3 | 作为用户，我想按时间、账户、分类、标签筛选交易 | - 交易列表支持多条件筛选<br>- 1k 条数据下查询 P95 <= 1s |

#### Epic C: Dashboard 与基础报表

| ID | User Story | Acceptance Criteria |
|----|------------|---------------------|
| FR-C1 | 作为用户，我想在首页看到核心财务摘要 | - Dashboard 至少展示总资产、本月收入、本月支出、本月结余<br>- “总资产”仅基于已绑定资金账户计算，且不受账本筛选影响<br>- `account_id = NULL` 的交易计入收入/支出/结余统计，但不计入总资产<br>- 页面加载 P95 <= 1s |
| FR-C2 | 作为用户，我想查看按时间和分类聚合的统计图表 | - 支持按月查看收支趋势<br>- 支持按分类查看支出分布<br>- 支持按账本和时间范围筛选<br>- 已归档或已删除但仍被历史交易引用的分类，仍需在统计结果中按历史名称参与聚合展示 |

#### Epic D: 导入导出与开放能力

| ID | User Story | Acceptance Criteria |
|----|------------|---------------------|
| FR-D1 | 作为用户，我想导出我的数据，以便迁移或备份 | - 支持 CSV、JSON 导出<br>- 用户可选择时间范围<br>- 导出结果需保留交易创建时的分类名称；若分类后续已删除或归档，导出文件仍应输出历史分类名称<br>- 10k 笔交易以内导出时间 <= 10 秒 |
| FR-D2 | 作为用户，我想导入历史数据，以便从其他工具迁移 | - v1 支持 CSV 导入<br>- 用户可手动映射源列到系统字段<br>- 系统提供预览界面，导入前可确认和修正映射<br>- 去重规则基于（金额 + 日期 + 描述）三元组精确匹配<br>- 导入结果返回成功、跳过、失败条数及原因 |
| FR-D3 | 作为技术用户，我想通过 API 读取和写入账务数据 | - 提供账户、交易、分类、标签、统计、导出的 REST API<br>- 提供 OpenAPI 文档 |
| FR-D4 | 作为技术用户，我想使用长期凭证访问 API | - 支持生成、列出、撤销 Personal Access Token<br>- PAT 默认有效期 90 天，可自定义更短时间<br>- PAT 仅用于业务 API，不可用于创建或刷新认证令牌 |

#### Optional Enhancement: AI 导入映射

| ID | User Story | Acceptance Criteria |
|----|------------|---------------------|
| FR-X1 | 作为用户，我希望系统能自动推荐 CSV 字段映射 | - 当系统配置 LLM API Key 时，可调用 n8n 工作流 W3 分析表头和样本行<br>- 输出推荐映射供用户确认<br>- 若未配置 LLM，则系统仍可通过手动映射完成导入 |

#### Optional Enhancement: Apple Shortcuts AI 记账

| ID | User Story | Acceptance Criteria |
|----|------------|---------------------|
| FR-X2 | 作为用户，我想通过 Apple 快捷指令快速记账，以便在 iPhone 上低摩擦记录支出和收入 | - 通过 `POST {n8n_url}/webhook/quick-entry` 接收快捷指令文本输入<br>- n8n 工作流验证 PAT、调用 LLM 解析文本、再调用 Go API 创建交易<br>- 交易默认写入“总账本”；若 `account_hint` 唯一匹配成功则写入 `account_id`，否则允许 `account_id = NULL`<br>- 快捷指令返回结构化确认结果<br>- 未配置 LLM 或 n8n 时，不影响 Web/PWA 核心手动记账链路 |

### 2.6 Non-Goals

以下能力明确不进入 `v1` 核心交付范围：

- Web 端自然语言记账
- OCR 票据识别
- 预算提醒与推送
- 共享账本与 AA 分账
- AI 洞察与对话式查询
- 银行账户直连
- 投资、报税、加密货币管理
- 原生 iOS / Android App
- 自动汇率拉取与实时换算

### 2.7 UI/UX Design Principles

#### 设计方向

采用**简洁克制的功能型设计**，优先服务“快记账、快看懂、少配置”。

#### 原则

| 原则 | 说明 |
|------|------|
| 即时记账 | 打开 App 后 <= 2 步到达记账页 |
| 信息密度适中 | Dashboard 只展示核心四数与最近流水 |
| 色彩语义化 | 收入绿色、支出红色、转账蓝色 |
| 留白与呼吸感 | 列表项间距 >= 12px，卡片轻阴影 |
| 渐进披露 | 次要字段默认折叠，需要时展开 |

#### 关键页面

- **首页 / Dashboard**：月度摘要卡片 + 最近交易列表 + 固定底部导航。
- **记账页**：支出 / 收入 / 转账切换，分类网格优先，账户和标签折叠展示。
- **统计页**：分类图、趋势图、时间范围切换。
- **交易列表**：按日分组，支持编辑与删除。

---

## 3. AI System Requirements

本节仅适用于可选增强能力，不属于 `v1` 核心发布门禁。

### 3.1 Tool Requirements

- `n8n`：承载 AI 相关工作流编排。
- `LLM Provider`：OpenAI / Claude / Ollama 等，具体供应商在部署时配置。
- `Redis`：用于认证、限流、会话与异步流程辅助。

### 3.2 AI Scope in v1

`v1` 允许两类 AI 相关能力，但都不作为核心发布门禁：

- **CSV 字段映射推荐**
- **Apple Shortcuts + n8n 外部 AI 记账（可选增强）**

以下 AI 能力不属于 `v1` 核心交付：

- Web 端自然语言记账
- 智能分类
- 对话式查询

### 3.3 Evaluation Strategy

对于 CSV 字段映射推荐能力，采用以下评估：

- 使用 20 份代表性 CSV 样本做离线验证。
- Top-1 字段映射准确率目标 >= 80%。
- AI 建议错误时，用户必须可在 UI 中手动纠正并成功完成导入。
- 未配置 AI 时，手动导入链路必须 100% 可用。

对于 Apple Shortcuts AI 记账能力，采用以下评估：

- 使用至少 50 条代表性自然语言记账语料做离线验证，覆盖支出、收入、转账、含账户提示和无账户提示场景。
- 交易类型识别准确率目标 >= 95%，金额提取准确率目标 >= 98%。
- 当 `account_hint` 无法唯一匹配时，系统必须返回可接受结果并以 `account_id = NULL` 成功创建交易，不得错误绑定账户。
- LLM 或 n8n 不可用时，必须返回明确错误，且不得创建半成功或错误交易。

---

## 4. Technical Specifications

### 4.1 Architecture Overview

#### v1 核心链路

```text
React PWA
  -> Go API
     -> PostgreSQL
     -> Redis
     -> SMTP (验证码邮件发送)
     -> n8n (可选 AI 导入增强 / 可选快捷指令记账编排)
```

#### 职责划分

| 职责 | Go 后端 | n8n |
|------|---------|-----|
| 用户认证 | Yes | No |
| 账本 / 账户 / 交易 / 分类 / 标签 CRUD | Yes | No |
| 余额计算与转账事务 | Yes | No |
| Dashboard 统计查询 | Yes | No |
| PAT 管理 | Yes | No |
| 邮箱验证码发送（SMTP） | Yes | No |
| CSV AI 字段映射推荐 | No | Yes，可选 |
| Apple Shortcuts AI 记账编排 | No | Yes，可选 |

### 4.2 Proposed Stack

| 层级 | 技术选型 | 说明 |
|------|---------|------|
| 前端 | React 19 + TypeScript | PWA 主体 |
| UI | Shadcn/UI + Tailwind CSS | 快速搭建界面 |
| 后端 | Go + Gin | 核心账务逻辑与 API |
| 数据库 | PostgreSQL | 主数据存储 |
| 缓存 | Redis | 验证码、会话、黑名单、限流 |
| 邮件 | SMTP | 验证码邮件发送 |
| 工作流 | n8n | 可选 AI 编排 |
| ORM / Migration | GORM + golang-migrate | ORM 与迁移分离 |
| 部署 | Docker + Docker Compose | 标准自托管方式 |

### 4.3 Core Data Model

核心实体：

- `User`
- `Ledger`
- `Account`
- `Transaction`
- `Category`
- `Tag`
- `PersonalAccessToken`

其中 `Category` 采用用户级数据模型，支持以下字段语义：

- `user_id`：分类所属用户
- `parent_id`：为空表示一级分类；非空表示其子分类
- `source`：`system` 或 `custom`，用于标识是否来自默认分类模板
- `archived_at`：归档时间，可为空

#### 关键约束

- `Transaction.ledger_id` 为 `NOT NULL`。
- `Transaction.account_id` 可为空。
- `Account` 属于 `User`，不属于 `Ledger`。
- 每个 `User` 有且仅有一个默认账本。
- 每个 `User` 在注册时初始化一套默认分类模板，后续变更仅影响该用户。
- `Category.parent_id` 可为空；`v1` 限定为两级结构，不支持无限层级。
- 已被交易引用的分类不得物理删除；用户删除此类分类时，系统应转为归档，并保留历史交易引用。
- 已归档或已删除的分类在统计、交易详情、导出结果中仍展示历史分类名称，但在新建交易和编辑交易时不再作为可选项。
- 总资产 = 所有账户当前余额之和，不受账本筛选影响。
- `account_id = NULL` 的交易计入收入、支出、分类和趋势统计，但不计入总资产。
- `PersonalAccessToken` 仅保存哈希值，不保存明文；支持过期与撤销。

#### 转账模型

采用双边交易模型，通过 `transfer_pair_id` 关联一出一入两条记录。

```text
转账: 储蓄卡 -> 支付宝 500

Transaction A
- type = expense
- account_id = 储蓄卡
- amount = -500
- transfer_pair_id = uuid

Transaction B
- type = income
- account_id = 支付宝
- amount = +500
- transfer_pair_id = uuid
```

规则：

- 创建转账时，两条交易在同一数据库事务中写入。
- 编辑转账时，两条交易同步更新。
- 删除转账时，两条交易同时删除。
- 默认统计口径排除 `transfer_pair_id` 非空的记录。

### 4.4 Integration Points

#### 认证 API

| Method | Path | 说明 |
|--------|------|------|
| POST | `/api/auth/send-code` | 发送邮箱验证码 |
| POST | `/api/auth/verify-code` | 验证码登录 |
| GET | `/api/auth/google` | Google OAuth 跳转 |
| GET | `/api/auth/google/callback` | Google OAuth 回调 |
| POST | `/api/auth/refresh` | 刷新 Access Token |
| POST | `/api/auth/logout` | 退出登录 |
| GET | `/api/auth/me` | 获取当前用户信息 |

#### 业务 API

| 资源 | 接口 | 说明 |
|------|------|------|
| `/api/ledgers` | GET / POST / PUT / DELETE | 账本 CRUD |
| `/api/accounts` | GET / POST / PUT / DELETE | 账户 CRUD 与归档 |
| `/api/transactions` | GET / POST / PUT / DELETE | 交易 CRUD 与筛选 |
| `/api/categories` | GET / POST / PUT / DELETE | 分类 CRUD；支持父分类与子分类；默认分类模板在用户注册时初始化到用户空间，后续删除或编辑仅影响当前用户 |
| `/api/tags` | GET / POST / PUT / DELETE | 标签 CRUD |
| `/api/stats/overview` | GET | Dashboard 概览；`total_assets` 始终按全部账户余额返回，不受 `ledger_id` 筛选影响；收入、支出、结余可按账本筛选 |
| `/api/stats/trend` | GET | 收支趋势 |
| `/api/stats/category` | GET | 分类统计；已归档或已删除但仍被历史交易引用的分类，仍按历史分类名称参与聚合展示 |
| `/api/import/csv` | POST | 上传 CSV 并进入映射预览 |
| `/api/import/csv/confirm` | POST | 确认导入 |
| `/api/export` | GET | 导出 CSV / JSON；若分类后续已删除或归档，导出结果仍保留交易对应的历史分类名称 |
| `/api/settings/tokens` | GET / POST / DELETE | PAT 管理；默认有效期 90 天，仅可访问业务资源 API，不可用于认证签发/刷新接口 |

#### 明确不属于 v1 API

- `/api/quick-entry` — Apple 快捷指令不直接调用 Go API，而是调用 n8n Webhook
- `/api/ai/*` — Web 端 AI 洞察、自然语言记账
- `/api/webhooks` — 用户自定义 Webhook 管理
- `/api/ledgers/*/members`

### 4.5 Security & Privacy

- Access Token 有效期 15 分钟，Refresh Token 有效期 7 天。
- 验证码发送限制：同邮箱 60 秒，同 IP 每小时最多 20 次。
- Refresh Token 退出后进入 Redis 黑名单。
- PAT 默认有效期 90 天，支持手动撤销。
- PAT 仅用于业务 API，不可调用认证签发接口。
- PAT 在数据库中仅保存哈希，不保存明文。
- 所有敏感配置通过环境变量管理。
- 默认不向第三方发送财务数据；AI 相关能力必须显式启用。

### 4.6 Apple Shortcuts Optional Enhancement

#### 目标

Apple Shortcuts AI 记账属于 `v1` 的可选增强能力，不作为核心发布门禁。该能力用于在 iPhone 上通过 Siri 或快捷指令快速提交自然语言记账文本，但不改变核心 Web/PWA 手动记账链路。

#### 方案概述

- Apple Shortcuts 通过 `POST {n8n_url}/webhook/quick-entry` 调用 n8n。
- n8n 验证 PAT 后调用 LLM 解析文本。
- n8n 再调用 Go API `POST /api/transactions` 创建交易。
- 默认写入“总账本”；若解析结果包含 `account_hint` 且唯一匹配成功，则写入 `account_id`；否则 `account_id = NULL`。
- 未配置 n8n 或 LLM 时，该增强能力不可用，但不影响 v1 核心手动记账功能。

#### 验收标准

- 输入如“午饭25元 微信”时，可返回结构化结果并成功创建交易。
- 账户唯一匹配失败时，交易仍可落账到总账本，且 `account_id = NULL`。
- LLM 不可用时返回明确错误，不创建错误交易。
- 撤销 PAT 后再次调用返回 401。

### 4.7 Deployment

- 提供单机 Docker Compose 方案，包含 Go 后端、React 前端、PostgreSQL、Redis。
- n8n 不纳入主 Compose，可通过现有实例对接。
- 通过环境变量 `N8N_BASE_URL` 配置 n8n 地址。
- 通过环境变量 `SMTP_HOST`、`SMTP_PORT`、`SMTP_USER`、`SMTP_PASS`、`SMTP_FROM` 配置验证码邮件发送。
- 新用户应能在 30 分钟内按文档完成部署。

---

## 5. Risks & Roadmap

### 5.1 Phased Rollout

#### Milestone 1: Foundation（第 1-2 周）

交付：

- 项目脚手架、CI
- 邮箱验证码登录 + Google OAuth
- 默认总账本创建逻辑
- PostgreSQL + Redis + 前后端 Docker 开发环境
- SMTP 验证码邮件发送能力

发布标准：

- 新用户登录后自动拥有默认总账本
- 可在无账户情况下创建第一笔交易
- 验证码链路、Google 登录链路可用

#### Milestone 2: Core Ledger（第 3-5 周）

交付：

- 收入、支出、转账完整链路
- 分类、标签、交易列表、编辑、删除
- 余额回算

发布标准：

- 转账双边交易写入正确
- 编辑和删除后余额回算正确
- 交易筛选在 1k 数据量下 P95 <= 1s

#### Milestone 3: Reporting & Portability（第 6-8 周）

交付：

- Dashboard
- 基础统计图表
- CSV / JSON 导出
- CSV 手动映射导入
- OpenAPI 文档
- PAT 管理

发布标准：

- Dashboard 口径与账户余额一致
- 导入导出链路完整可用
- PAT 可生成、撤销并受权限限制

#### Milestone 4: Optional Enhancements（第 9-10 周，可选）

交付：

- CSV AI 字段映射推荐
- Apple Shortcuts + n8n 外部 AI 记账
- 对接 n8n W3 / quick-entry 工作流
- 相关部署文档

发布标准：

- AI 功能关闭时不影响核心 v1 链路
- AI 功能开启时可给出可编辑的映射建议
- 快捷指令增强链路可创建交易，但不阻塞核心 v1 发布

### 5.2 Technical Risks

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| “总资产”口径不统一 | Dashboard、导入导出、测试结果不一致 | 中 | 在 PRD 中固定口径，并体现在 API 与测试用例中 |
| 转账与余额回算逻辑复杂 | 容易出现账实不一致 | 中 | 使用事务 + 专项测试覆盖创建、编辑、删除 |
| 自托管环境差异大 | 安装失败率高 | 中 | 提供最小依赖 Compose 和安装验收脚本 |
| n8n 实例不可用 | AI 增强不可用 | 低 | 核心 CRUD 与验证码邮件发送不依赖 n8n；AI 增强必须可关闭 |
| PAT 泄漏 | 业务 API 被滥用 | 中 | 默认过期、仅存哈希、支持撤销、限制作用范围 |

### 5.3 Open Questions

1. v1 是否需要游客模式或本地只读演示模式？
2. CSV 导入首批需要兼容哪些外部模板？
3. PAT 是否需要进一步细化为只读 / 读写两种 scope？
4. Dashboard 是否需要展示“未分配交易金额”作为辅助信息？

### 5.4 Future Roadmap

#### v1.1

- CSV AI 字段映射稳定化
- 深色模式
- 离线记账与同步
- 多币种可见能力设计

#### v1.2

- Web 端自然语言记账
- 智能分类
- OCR 票据识别

#### v2.0

- 共享账本
- AA 分账
- Webhook
- AI 洞察与对话式查询

---

*Xledger - Your finances, your rules.*
