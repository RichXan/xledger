# Xledger 功能增强设计方案

> 版本: v1.0 | 日期: 2026-04-04 | 状态: 已确认

---

## 1. 概述

本文档涵盖 Xledger 五个独立子系统的设计方案：

| 子系统 | 核心功能 |
|--------|----------|
| 移动端 PWA 增强 | 离线只读、推送通知、iOS 安装引导 |
| 预算系统 + 提醒 | 按分类月度预算、三类提醒触发 |
| AI 记账 | 自然语言解析 + OCR 截图（iOS Shortcuts + n8n） |
| 导入工具完善 | 支付宝 CSV 优先、ezbookkeeping 其次 |
| 多语言支持 (i18n) | 简体中文首期、react-i18next |

---

## 2. 子系统一：移动端 PWA 增强

### 2.1 离线只读策略

**数据范围：**
- 最近 30 天的交易（完整列表）
- 所有账户和分类元数据
- Dashboard 摘要数据

**缓存架构：**

| 层级 | 技术 | 策略 | 内容 |
|------|------|------|------|
| 静态资源 | Cache Storage | Cache-First | JS/CSS/图标/fonts |
| API 响应 | IndexedDB | Network-First → IndexedDB 回退 | 交易列表、账户、分类 |
| 用户操作 | 写入队列（IndexedDB） | 联网后重放 | 离线期间的记账操作 |

**Service Worker 缓存流程：**
```
请求进来
    │
    ├── 静态资源（/static/*, /assets/*）
    │   └── Cache-First，直接用缓存
    │
    ├── API 请求（/api/*）
    │   ├── 尝试网络请求
    │   │   ├── 成功 → 返回数据，同时更新 IndexedDB
    │   │   └── 失败（离线）→ 从 IndexedDB 读取
    │   └── 失败（离线）→ 从 IndexedDB 读取
    │
    └── 离线提示 UI
        └── 检测到离线时，顶部显示横幅
```

**离线写入队列：**
- 用户离线期间的记账请求先存入 IndexedDB 队列
- 网络恢复后，通过 Background Sync API 依次重放
- 重放失败时保留队列并提示用户手动重试

### 2.2 推送通知

**技术实现：** Web Push API + VAPID（ Voluntary Application Server Identity for Web Push）

**通知类型：**

| 类型 | 触发时机 | 内容示例 |
|------|----------|----------|
| 实时超支提醒 | 用户记账后立即检查该分类预算 | "餐饮分类已超支！本月 ¥2100 / 预算 ¥2000" |
| 每日起床推送 | 每天早上 9:00（用户本地时间） | "早安！本月已花 ¥8500，剩余 ¥1500" |
| 周报摘要推送 | 每周日晚上 20:00（用户本地时间） | "本周支出 ¥1200，餐饮占比最高 ¥450" |

**通知偏好设置（用户可关闭任一类型）：**
```typescript
interface NotificationPreference {
  realtime_alert: boolean   // 实时超支提醒，默认开启
  daily_digest: boolean     // 每日起床推送，默认关闭
  weekly_digest: boolean    // 周报摘要推送，默认开启
}
```

**后端定时任务：**
- 使用 Go 的 `robfig/cron` 或 Redis sorted set 实现定时触发
- `每日起床推送`：查询所有当日预算超支用户，推送
- `周报摘要推送`：每周日统计并推送

### 2.3 移动端 UX 优化

**响应式布局：**

| 断点 | 布局 |
|------|------|
| < 640px（手机） | 底部 Tab 导航、单列布局、记账入口突出 |
| 640-1024px（平板） | 侧边栏收起、底部 Tab 导航 |
| > 1024px（桌面） | 侧边栏固定、顶部导航 |

**iOS Safari "添加到主屏幕"引导：**
- 首次访问时检测 iOS Safari，弹出引导提示
- 说明如何通过分享按钮 → "添加到主屏幕"安装
- 引导只出现一次，已安装用户不再显示

**触摸优化：**
- 记账页分类网格按钮：最小 48x48px 触摸区域
- 滑动操作：左滑显示删除/编辑快捷操作
- 长按交易项：显示更多操作菜单

### 2.4 组件清单

| 文件路径 | 职责 |
|----------|------|
| `public/sw.js` | Service Worker，缓存策略、Push 监听 |
| `public/manifest.webmanifest` | 更新 display: standalone，添加相关权限 |
| `src/features/offline/offline-store.ts` | IndexedDB 读写层（Dexie.js） |
| `src/features/offline/offline-queue.ts` | 离线写入队列管理 |
| `src/features/notifications/push-manager.ts` | Web Push 订阅/退订、VAPID 密钥管理 |
| `src/features/notifications/preference-store.ts` | 用户通知偏好设置 UI 和 API 同步 |
| `src/components/layout/mobile-nav.tsx` | 底部 Tab 导航栏（移动端专用） |
| `src/components/layout/offline-banner.tsx` | 离线状态横幅 |
| `src/pages/onboarding-pwa-page.tsx` | iOS 安装引导页 |

---

## 3. 子系统二：预算系统 + 提醒

### 3.1 数据模型

**新增数据库表：**

```sql
CREATE TABLE budgets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    amount      DECIMAL(15, 2) NOT NULL CHECK (amount > 0),
    period      TEXT NOT NULL DEFAULT 'monthly',  -- 'monthly' for v1
    thresholds  INTEGER[] NOT NULL DEFAULT ARRAY[80, 100],  -- 百分比阈值
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, category_id, period)
);

CREATE TABLE budget_alerts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_id     UUID NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    threshold     INTEGER NOT NULL,   -- 触发阈值 80 或 100
    spent_amount  DECIMAL(15, 2) NOT NULL,
    period_start  DATE NOT NULL,     -- 预算周期开始日期
    period_end    DATE NOT NULL,     -- 预算周期结束日期
    sent_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_notification_prefs (
    user_id          UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    realtime_alert   BOOLEAN NOT NULL DEFAULT true,
    daily_digest     BOOLEAN NOT NULL DEFAULT false,
    weekly_digest    BOOLEAN NOT NULL DEFAULT true,
    push_subscription TEXT,  -- Web Push 订阅 JSON，null = 未订阅
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### 3.2 提醒触发逻辑

**实时超支提醒流程：**

```
用户提交交易 POST /api/transactions
    │
    ├── 事务写入成功
    │
    ├── 查询该分类的当月已用金额
    │   SELECT SUM(amount) FROM transactions
    │   WHERE user_id = ? AND category_id = ? AND type = 'expense'
    │   AND date >= period_start AND date <= period_end
    │
    ├── 查询该分类的预算
    │   SELECT * FROM budgets WHERE user_id = ? AND category_id = ?
    │
    ├── 计算使用率 = 当月已用 / 预算金额 * 100
    │
    ├── 若使用率 >= 任一阈值 且 未发送过该阈值提醒：
    │   ├── 记录 budget_alerts
    │   └── 发送 Web Push 通知
    │
    └── 返回交易创建响应
```

**定时推送任务（后台 goroutine 或 cron）：**

| 任务 | 频率 | 逻辑 |
|------|------|------|
| 每日起床推送 | 每天 9:00 UTC | 查询所有 `daily_digest=true` 用户，生成摘要并推送 |
| 周报摘要 | 每周日 20:00 UTC | 查询所有 `weekly_digest=true` 用户，生成周报并推送 |

### 3.3 API 扩展

| Method | Path | 说明 |
|--------|------|------|
| GET | `/api/budgets` | 获取用户所有预算（含已用金额和使用率） |
| POST | `/api/budgets` | 创建分类预算 |
| PUT | `/api/budgets/:id` | 编辑预算金额/阈值 |
| DELETE | `/api/budgets/:id` | 删除预算 |
| GET | `/api/budgets/alerts` | 获取超支告警历史 |
| PATCH | `/api/budgets/preferences` | 更新通知偏好（实时/每日/每周开关） |
| POST | `/api/budgets/subscribe-push` | 订阅 Web Push |
| DELETE | `/api/budgets/subscribe-push` | 退订 Web Push |

**GET /api/budgets 响应示例：**

```json
{
  "data": [
    {
      "id": "uuid",
      "category_id": "uuid",
      "category_name": "餐饮",
      "amount": 2000.00,
      "period": "monthly",
      "thresholds": [80, 100],
      "spent_amount": 2150.00,
      "usage_percent": 107.5,
      "remaining": -150.00,
      "period_start": "2026-04-01",
      "period_end": "2026-04-30",
      "is_over_budget": true
    }
  ]
}
```

### 3.4 前端页面

| 页面 | 路径 | 组件 |
|------|------|------|
| 预算列表 | `/budgets` | BudgetListPage |
| 创建/编辑预算 | `/budgets/new`, `/budgets/:id/edit` | BudgetFormPage |
| 预算告警历史 | `/budgets/alerts` | BudgetAlertsPage |
| 通知偏好设置 | `/settings/notifications` | NotificationPrefsPage |

---

## 4. 子系统三：AI 记账

### 4.1 整体架构

**两条通道，均通过 iOS Shortcuts 触发：**

```
通道一：自然语言
用户: "午饭25元微信"
    │
    ▼
iOS Shortcuts → n8n Webhook (/webhook/quick-entry)
    │
    ▼
n8n: 验证 PAT → 调用 LLM 解析 → POST /api/transactions
    │
    ▼
Go API → 返回结构化结果 → Shortcuts 显示确认

通道二：OCR 截图
用户截屏 → iOS Shortcuts "识别文本" → n8n Webhook
    │
    ▼
n8n: OCR 文本 → LLM 解析 → POST /api/transactions
```

### 4.2 自然语言解析

**LLM 解析 Prompt（n8n 中配置）：**

```
你是一个家庭记账助手。用户输入一段自然语言记账内容，你需要提取结构化信息。

输入格式示例：
- "午饭25元微信"
- "工资 15000 招行"
- "打车35 账户没指定"
- "给老婆转1000"

输出格式（仅返回 JSON，不要其他内容）：
{
  "type": "expense" | "income" | "transfer",
  "amount": 数字（正数）,
  "category_hint": "分类名称（字符串或null）",
  "account_hint": "账户名称（字符串或null）",
  "description": "交易描述（简短）",
  "date": "YYYY-MM-DD（今天）"
}

规则：
1. type 必须从内容判断（"午饭""打车""买"→expense，"工资""收入""奖金"→income，"转"→transfer）
2. amount 必须是正数，类型由 type 决定正负符号
3. category_hint 匹配用户已有分类，不确定则设为 null
4. account_hint 匹配用户已有账户，不确定则设为 null
5. date 默认为今天，格式必须为 YYYY-MM-DD
6. transfer 类型需要转出账户和转入账户，account_hint 字段改为 from_account_hint 和 to_account_hint
```

**n8n Workflow 节点：**

```
ON.webhook (receive text)
    │
    ├── HTTP Request (Validate PAT)
    │       │
    │       └── 401 → Return error to Shortcuts
    │
    ├── Code (LLM prompt)
    │
    ├── HTTP Request (Call LLM: OpenAI/Claude)
    │       │
    │       └── Error → Return error to Shortcuts
    │
    ├── Code (Parse LLM response JSON)
    │
    ├── HTTP Request (POST /api/transactions)
    │   Body: { type, amount, category_id (resolved), account_id (resolved), ... }
    │       │
    │       └── Success → Return "已记录：午饭 ¥25 → 餐饮" to Shortcuts
    │       └── Error → Return error to Shortcuts
    │
    └── Response to Shortcuts
```

**Go API 端点（n8n 调用）：**

```
POST /api/transactions (via PAT authentication)
X-PAT: {token}

Body:
{
  "type": "expense",
  "amount": 25.00,
  "category_id": "resolved-from-hint-or-null",
  "account_id": "resolved-from-hint-or-null",
  "description": "午饭",
  "date": "2026-04-04",
  "ledger_id": "user-default-ledger"
}
```

### 4.3 OCR 截图识别

**iOS Shortcuts 流程：**

```
用户截屏（或从照片选择）
    │
    ▼
"识别文本" 动作（iOS 原生 OCR）
    │
    ▼
提取文本内容
    │
    ▼
"获取剪贴板" 或 "文本" 动作传入 n8n webhook URL
    │
    ▼
n8n 接收文本，后续流程同自然语言解析
```

**备注：** iOS Shortcuts 的"识别文本"支持中英文，使用设备端 ML，不过网络请求。

### 4.4 准确性目标

| 指标 | 目标值 |
|------|--------|
| 交易类型识别准确率 | >= 95% |
| 金额提取准确率 | >= 98% |
| 分类推荐准确率（当用户有相关历史） | >= 80% |
| 账户匹配准确率（当用户有相关账户） | >= 85% |

---

## 5. 子系统四：导入工具完善

### 5.1 支持格式

**优先级 1：支付宝 CSV**

支付宝导出文件特点：
- 表头固定行：`交易号,交易对方,交易对方账号,商品说明,金额(元),收/支,交易状态,备注,资金状态,创建时间`
- `收/支` 列区分方向：收入显示"收入"，支出显示"支出"
- `金额(元)` 列为正数，绝对值形式
- 日期格式：`2024-01-15 12:30:45`
- `交易状态`：已完成、退款等

**优先级 2：ezbookkeeping CSV**

需要确认实际字段，对齐用户已有格式。

### 5.2 格式自动检测

```typescript
type CSVFormat = 'alipay' | 'ezbookkeeping' | 'generic'

function detectFormat(headers: string[]): CSVFormat {
  const headerStr = headers.join(',')
  if (headerStr.includes('交易号') && headerStr.includes('收/支') && headerStr.includes('金额(元)')) {
    return 'alipay'
  }
  if (headerStr.includes('create_time') || headerStr.includes('transaction_date')) {
    return 'ezbookkeeping'
  }
  return 'generic'
}
```

### 5.3 字段映射规则

**支付宝 → Xledger：**

| 支付宝字段 | Xledger 字段 | 说明 |
|------------|--------------|------|
| 商品说明 | description | 交易描述 |
| 金额(元) | amount | 收/支=收入→正数；收/支=支出→负数 |
| 收/支 | type | 区分 income / expense |
| 创建时间 | date | 解析为 YYYY-MM-DD |
| 备注 | - | 暂时忽略或作为额外描述 |

**ezbookkeeping → Xledger：** 待确认实际格式后补充映射。

### 5.4 导入流程

```
用户上传 CSV
    │
    ▼
前端解析前 5 行，检测格式
    │
    ├── alipay → 自动映射 → 预览页
    ├── ezbookkeeping → 自动映射 → 预览页
    └── generic → 手动映射 → 预览页
    │
    ▼
用户确认预览（可修改分类映射）
    │
    ▼
POST /api/import/csv/confirm
    │
    ▼
返回 { success: N, skipped: M, failed: K, errors: [...] }
```

**去重规则：** `(ABS(amount) + DATE(date) + description)` 三元组精确匹配，跳过已存在的记录。

### 5.5 API 扩展

| Method | Path | 说明 |
|--------|------|------|
| POST | `/api/import/csv` | 上传 CSV，AI 或规则检测格式和映射 |
| GET | `/api/import/csv/preview` | 获取预览结果（含映射建议） |
| POST | `/api/import/csv/confirm` | 确认导入（返回详细结果） |
| POST | `/api/import/csv/mapping` | 保存用户自定义映射模板 |
| GET | `/api/import/csv/templates` | 获取已保存的映射模板列表 |

---

## 6. 子系统五：多语言支持 (i18n)

### 6.1 技术方案

- 库：`react-i18next` + `i18next` + `i18next-browser-languagedetector`
- 资源文件结构：
  ```
  src/locales/
    zh/
      translation.json
    en/
      translation.json
  ```
- 语言检测优先级：URL 参数 > localStorage > 浏览器 Accept-Language
- 用户语言偏好存储：用户表 `language` 字段，API `/api/auth/me` 返回

### 6.2 实现范围

**首期支持：简体中文（zh）**

所有用户可见文本均通过 `t('key')` 调用：

```typescript
// 使用示例
<div>{t('nav.dashboard')}</div>
<button>{t('action.save')}</button>
<span>{t('budget.remaining', { amount: remaining })}</span>
```

**命名空间划分：**

| 命名空间 | 内容 |
|----------|------|
| `common` | 通用语：保存、取消、确认、加载中、错误消息 |
| `nav` | 导航菜单项 |
| `auth` | 登录、注册、验证码相关 |
| `dashboard` | Dashboard 页文本 |
| `transaction` | 交易创建/编辑/列表 |
| `budget` | 预算相关 |
| `import` | 导入导出 |
| `settings` | 设置页 |
| `notifications` | 通知相关 |
| `errors` | 错误消息 |
| `validation` | 表单校验消息 |

### 6.3 日期/货币本地化

```typescript
// 日期格式化
new Intl.DateTimeFormat('zh-CN', { year: 'numeric', month: 'long', day: 'numeric' }).format(date)

// 货币格式化
new Intl.NumberFormat('zh-CN', { style: 'currency', currency: 'CNY' }).format(amount)
```

### 6.4 前端改动

| 文件 | 改动 |
|------|------|
| `src/main.tsx` | 初始化 i18next |
| `src/i18n/index.ts` | i18n 配置模块 |
| `src/locales/zh/translation.json` | 中文翻译资源 |
| `src/locales/en/translation.json` | 英文翻译资源 |
| `src/App.tsx` | 路由增加语言前缀 |
| `src/components/layout/top-bar.tsx` | 语言切换器 |

---

## 7. 依赖关系与优先级

```
实现顺序：

Phase 1（基础设施）
├── i18n（其他功能都要用）
└── PWA 增强（manifest + service worker 基础）

Phase 2（核心功能）
├── 预算系统（后端模型 + API + 前端页面）
└── AI 记账后端（n8n workflow + Go API 适配）

Phase 3（完善）
├── AI 记账 iOS Shortcuts 配置文档
├── 导入工具完善（支付宝 + ezbk）
└── PWA 推送通知（Web Push 完整实现）
```

---

## 8. 非目标 / 明确排除

- 原生 iOS/Android App（仅 PWA）
- 离线可写（Phase 1 仅为只读离线）
- 预算按周/按年周期（v1 仅月度）
- 自动汇率换算
- 银行账户直连
