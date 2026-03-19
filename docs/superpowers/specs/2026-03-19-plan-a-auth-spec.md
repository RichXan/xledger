# Plan-A 规格：认证与会话

> 作用：仅用于认证会话能力实现计划输入

## 1. 目标

- 提供邮箱验证码与 Google OAuth 双登录入口。
- 统一会话模型（Access/Refresh）与登出失效机制。
- 首次登录自动完成用户初始化（默认总账本创建）。

## 2. 核心流程

1. `send-code` 触发验证码发送并写入 Redis（频控）。
2. `verify-code` 验证通过后签发 Token。
3. Google OAuth 回调完成身份确认并签发 Token。
4. `refresh` 刷新 Access（可轮转 Refresh）。
5. `logout` 将 Refresh 写入黑名单并立即失效。

## 3. 接口范围

| Endpoint | Method | Auth | Success | Error |
|---|---|---|---|---|
| `/api/auth/send-code` | POST | 匿名 | `code_sent=true` | `AUTH_CODE_RATE_LIMIT`, `AUTH_CODE_SEND_FAILED` |
| `/api/auth/verify-code` | POST | 匿名 | Access/Refresh | `AUTH_CODE_INVALID`, `AUTH_CODE_EXPIRED` |
| `/api/auth/google/callback` | GET | 匿名 | Access/Refresh | `AUTH_OAUTH_FAILED` |
| `/api/auth/refresh` | POST | Refresh | New Access | `AUTH_REFRESH_REVOKED`, `AUTH_REFRESH_EXPIRED` |
| `/api/auth/logout` | POST | Access | `revoked=true` | `AUTH_UNAUTHORIZED` |

## 4. 数据与约束

- 验证码：6 位数字，10 分钟有效，同邮箱 60 秒限发。
- OAuth：必须校验 `state`，启用 `nonce` 防重放。
- Refresh：有效期 7 天；轮换时旧 Refresh 立即失效。
- 黑名单传播 SLA <= 5 秒。

## 5. 异常与降级

- SMTP 邮件发送失败：返回可重试错误，不创建会话。
- OAuth 失败：仅影响 Google 登录，不影响验证码链路。
- 黑名单传播超 SLA：进入“严格模式”，临时拒绝 Refresh 并告警。

## 6. 验收要点

- 新用户首登后能直接进入核心业务（默认总账本存在）。
- 登出后 Refresh 无法再刷新 Access。
- 错误码与安全边界（state/nonce/重放）可被自动化测试覆盖。

### Out of Scope

- 交易、统计、导入导出、PAT 管理。
