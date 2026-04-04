# i18n 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 为 Xledger 前端添加 react-i18next 国际化支持，首期支持简体中文。

**Architecture:** react-i18next + i18next-browser-languagedetector。语言偏好通过用户表 `language` 字段持久化，API 请求携带 `Accept-Language` header。

**Tech Stack:** react-i18next, i18next, i18next-browser-languagedetector

---

## 文件影响

| 文件 | 动作 |
|------|------|
| `frontend/app/package.json` | 添加依赖 |
| `frontend/app/src/i18n/index.ts` | 新建 |
| `frontend/app/src/i18n/resources.ts` | 新建 |
| `frontend/app/src/locales/zh/translation.json` | 新建 |
| `frontend/app/src/locales/en/translation.json` | 新建 |
| `frontend/app/src/main.tsx` | 修改：初始化 i18next |
| `frontend/app/src/App.tsx` | 修改：路由添加语言前缀 |
| `frontend/app/src/features/auth/auth-api.ts` | 修改：添加 language header |
| `frontend/app/src/features/auth/auth-context.tsx` | 修改：同步语言偏好 |
| `frontend/app/src/lib/api.ts` | 修改：携带 language header |
| `frontend/app/src/pages/dashboard-page.tsx` | 修改：使用 t() 替换硬编码文本 |
| `frontend/app/src/pages/transactions-page.tsx` | 修改：使用 t() 替换硬编码文本 |
| `frontend/app/src/pages/accounts-page.tsx` | 修改：使用 t() 替换硬编码文本 |
| `frontend/app/src/pages/analytics-page.tsx` | 修改：使用 t() 替换硬编码文本 |
| `frontend/app/src/pages/settings-page.tsx` | 修改：使用 t() 替换硬编码文本 |
| `frontend/app/src/components/layout/top-bar.tsx` | 修改：添加语言切换器 |
| `backend/internal/auth/handler.go` | 修改：/auth/me 响应添加 language 字段 |
| `backend/internal/models/user.go` | 修改：User 模型添加 Language 字段 |
| `migrations/YYYYMMDDDD_add_language_to_users.sql` | 新建 |

---

## Task 1: 安装 i18n 依赖

- [ ] **Step 1: 添加 npm 依赖**

```bash
cd frontend/app && npm install react-i18next i18next i18next-browser-languagedetector i18next-http-backend
```

Expected output: 包安装成功，无报错

- [ ] **Step 2: 验证安装**

```bash
npm list react-i18next i18next
```

Expected output: 显示已安装版本

- [ ] **Step 3: Commit**

```bash
git add package.json package-lock.json && git commit -m "chore(frontend): add i18n dependencies

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 2: 创建 i18n 配置

- [ ] **Step 1: 创建目录结构**

```bash
mkdir -p frontend/app/src/i18n frontend/app/src/locales/zh frontend/app/src/locales/en
```

- [ ] **Step 2: 创建中文翻译资源**

新建 `frontend/app/src/locales/zh/translation.json`:

```json
{
  "common": {
    "save": "保存",
    "cancel": "取消",
    "confirm": "确认",
    "delete": "删除",
    "edit": "编辑",
    "create": "创建",
    "loading": "加载中...",
    "error": "出错了",
    "retry": "重试",
    "success": "操作成功",
    "noData": "暂无数据"
  },
  "nav": {
    "dashboard": "首页",
    "transactions": "交易",
    "analytics": "统计",
    "accounts": "账户",
    "settings": "设置",
    "shortcuts": "快捷指令"
  },
  "auth": {
    "login": "登录",
    "register": "注册",
    "logout": "退出登录",
    "email": "邮箱",
    "code": "验证码",
    "sendCode": "发送验证码",
    "codeSent": "验证码已发送",
    "invalidCode": "验证码无效"
  },
  "dashboard": {
    "title": "财务概览",
    "income": "收入",
    "expense": "支出",
    "totalAssets": "总资产",
    "monthlyOverview": "月度概览",
    "recentTransactions": "最近交易",
    "lastSynced": "最后同步",
    "minutesAgo": "分钟前"
  },
  "transaction": {
    "add": "记一笔",
    "type": "类型",
    "typeExpense": "支出",
    "typeIncome": "收入",
    "typeTransfer": "转账",
    "amount": "金额",
    "category": "分类",
    "account": "账户",
    "tag": "标签",
    "date": "日期",
    "memo": "备注",
    "noAccount": "无账户",
    "createSuccess": "记账成功",
    "deleteSuccess": "删除成功",
    "deleteConfirm": "确认删除这笔交易？"
  },
  "account": {
    "title": "账户",
    "create": "新建账户",
    "edit": "编辑账户",
    "name": "账户名称",
    "type": "账户类型",
    "balance": "余额",
    "archive": "归档",
    "archived": "已归档"
  },
  "analytics": {
    "title": "统计分析",
    "trend": "收支趋势",
    "byCategory": "按分类",
    "selectPeriod": "选择周期"
  },
  "settings": {
    "title": "设置",
    "language": "语言",
    "notifications": "通知设置",
    "export": "导出数据",
    "about": "关于"
  },
  "budget": {
    "title": "预算",
    "create": "创建预算",
    "edit": "编辑预算",
    "monthly": "月度预算",
    "amount": "预算金额",
    "spent": "已花费",
    "remaining": "剩余",
    "overBudget": "已超支",
    "alertThreshold": "提醒阈值"
  },
  "import": {
    "title": "导入数据",
    "selectFile": "选择文件",
    "preview": "预览",
    "confirm": "确认导入",
    "success": "导入成功",
    "skipped": "跳过",
    "failed": "失败"
  },
  "errors": {
    "networkError": "网络错误，请检查连接",
    "serverError": "服务器错误，请稍后重试",
    "unauthorized": "未登录或登录已过期",
    "validationError": "输入数据不合法"
  }
}
```

- [ ] **Step 3: 创建英文翻译资源**

新建 `frontend/app/src/locales/en/translation.json`（英文版本，同结构）

- [ ] **Step 4: 创建 i18n 配置模块**

新建 `frontend/app/src/i18n/index.ts`:

```typescript
import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import LanguageDetector from 'i18next-browser-languagedetector'
import HttpBackend from 'i18next-http-backend'
import { en, zh } from './resources'

export const supportedLanguages = ['en', 'zh'] as const
export type SupportedLanguage = (typeof supportedLanguages)[number]

export const defaultLanguage: SupportedLanguage = 'en'

i18n
  .use(HttpBackend)
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources: {
      en: { translation: en },
      zh: { translation: zh },
    },
    fallbackLng: defaultLanguage,
    supportedLngs: supportedLanguages,
    detection: {
      order: ['querystring', 'localStorage', 'navigator'],
      lookupQuerystring: 'lang',
      caches: ['localStorage'],
    },
    interpolation: {
      escapeValue: false,
    },
  })

export default i18n

export function changeLanguage(lang: SupportedLanguage) {
  return i18n.changeLanguage(lang)
}

export function getCurrentLanguage(): SupportedLanguage {
  return i18n.language as SupportedLanguage
}
```

- [ ] **Step 5: 创建 resources.ts 简化内联资源**

新建 `frontend/app/src/i18n/resources.ts`:

```typescript
import { en } from '../locales/en/translation'
import { zh } from '../locales/zh/translation'

export { en, zh }
```

- [ ] **Step 6: Commit**

```bash
git add frontend/app/src/i18n frontend/app/src/locales && git commit -m "feat(frontend): add i18n configuration and translation resources

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 3: 集成 i18n 到 main.tsx

- [ ] **Step 1: 修改 main.tsx**

```typescript
// 在 import styles.css 之后、ReactDOM.createRoot 之前添加：
import './i18n'
```

完整 `frontend/app/src/main.tsx`:

```typescript
import React from 'react'
import ReactDOM from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'
import App from './App'
import './styles.css'
import './i18n'
import { AuthProvider } from './features/auth/auth-context'
import { registerServiceWorker } from './features/pwa/register-sw'

const queryClient = new QueryClient()
registerServiceWorker()

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <BrowserRouter>
          <App />
        </BrowserRouter>
      </AuthProvider>
    </QueryClientProvider>
  </React.StrictMode>,
)
```

- [ ] **Step 2: Commit**

```bash
git add frontend/app/src/main.tsx && git commit -m "feat(frontend): initialize i18next in main.tsx

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 4: 修改 API 层携带 language

- [ ] **Step 1: 修改 lib/api.ts**

在所有 fetch 调用中添加 `Accept-Language` header:

```typescript
// 在 requestEnvelope 函数中添加：
headers: {
  ...headers,
  'Accept-Language': i18n.language,
}
```

完整模式：

```typescript
export async function requestEnvelope<T>(path: string, options?: RequestInit): Promise<T> {
  const lang = typeof i18n !== 'undefined' ? i18n.language : 'en'
  const response = await fetch(`/api${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      'Accept-Language': lang,
      ...options?.headers,
    },
  })
  // ... 其余不变
}
```

注意：需要处理 `i18n` 在模块加载时可能未初始化的情况。

- [ ] **Step 2: Commit**

```bash
git add frontend/app/src/lib/api.ts && git commit -m "feat(frontend): add Accept-Language header to API requests

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 5: 添加语言切换器到 TopBar

- [ ] **Step 1: 修改 top-bar.tsx**

添加语言切换下拉菜单：

```typescript
import { useTranslation } from 'react-i18next'
import { changeLanguage, supportedLanguages, getCurrentLanguage } from '@/i18n'
import { Globe } from 'lucide-react'

// 在 TopBar 组件中添加：
const { t } = useTranslation()
const currentLang = getCurrentLanguage()

// 添加语言切换 UI（按钮打开下拉菜单）
<select
  value={currentLang}
  onChange={(e) => changeLanguage(e.target.value as 'en' | 'zh')}
  className="..."
>
  {supportedLanguages.map((lang) => (
    <option key={lang} value={lang}>
      {lang === 'zh' ? '中文' : 'English'}
    </option>
  ))}
</select>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/app/src/components/layout/top-bar.tsx && git commit -m "feat(frontend): add language switcher to TopBar

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 6: 翻译 Dashboard 页面（示例）

- [ ] **Step 1: 修改 dashboard-page.tsx**

将硬编码文本替换为 t() 调用：

```typescript
import { useTranslation } from 'react-i18next'

// 在组件内：
const { t } = useTranslation()

// 替换：
// "Financial Overview" → t('dashboard.title')
// "Real-time precision analytics..." → t('dashboard.subtitle')
// "Month Income" → t('dashboard.income')
// "Month Expenses" → t('dashboard.expense')
// "Total Assets" → t('dashboard.totalAssets')
// "12-Month Spending Trend" → t('analytics.trend')
// 等等...
```

- [ ] **Step 2: 验证页面正常运行**

```bash
npm run dev
```

打开浏览器，检查 Dashboard 文本是否正确显示（中文或英文取决于语言设置）。

- [ ] **Step 3: Commit**

```bash
git add frontend/app/src/pages/dashboard-page.tsx && git commit -m "feat(frontend): translate dashboard page with i18n

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 7: 后端添加 language 字段

- [ ] **Step 1: 创建迁移文件**

新建迁移文件（文件名用实际日期）：

```sql
-- migrations/YYYYMMDDDD_add_language_to_users.sql
ALTER TABLE users ADD COLUMN language TEXT NOT NULL DEFAULT 'en';
```

- [ ] **Step 2: 修改 User 模型**

在 `internal/models/user.go` 添加 `Language` 字段：

```go
type User struct {
    // ... 现有字段
    Language string `json:"language" gorm:"default:en"`
}
```

- [ ] **Step 3: 修改 /auth/me 响应**

在 `internal/auth/handler.go` 的 `Me` 方法中，确保返回 `language` 字段：

```go
// 在构建用户响应时添加：
c.JSON(http.StatusOK, gin.H{
    "user": gin.H{
        "id":       user.ID,
        "email":    user.Email,
        "language": user.Language,  // 添加这行
    },
})
```

- [ ] **Step 4: 运行迁移测试**

```bash
cd backend && go test ./migrations/... -count=1
```

- [ ] **Step 5: Commit**

```bash
git add backend/migrations/ backend/internal/models/ backend/internal/auth/handler.go
git commit -m "feat(backend): add language field to User model

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 8: auth-context 同步语言偏好

- [ ] **Step 1: 修改 auth-context.tsx**

当 `/auth/me` 返回用户信息时，同步语言偏好到 i18n：

```typescript
// 在获取用户信息后：
if (user.language && supportedLanguages.includes(user.language)) {
  changeLanguage(user.language)
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/app/src/features/auth/auth-context.tsx && git commit -m "feat(frontend): sync user language preference to i18n

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 9: 翻译其余页面

对以下页面重复 Task 6 的模式：

- [ ] transactions-page.tsx
- [ ] accounts-page.tsx
- [ ] analytics-page.tsx
- [ ] settings-page.tsx
- [ ] login-page.tsx

每个页面单独 commit：

```bash
git add frontend/app/src/pages/<page-name>.tsx
git commit -m "feat(frontend): translate <page-name> with i18n

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 10: 最终验证

- [ ] **Step 1: 完整构建测试**

```bash
cd frontend/app && npm run build
```

Expected: 构建成功，无 TypeScript 错误

- [ ] **Step 2: 语言切换测试**

本地启动前端，测试：
1. 打开 Dashboard → 切换语言 → 验证所有文本更新
2. 刷新页面 → 验证语言偏好保持
3. 切换到中文 → 刷新 → 验证仍是中文

- [ ] **Step 3: 后端测试**

```bash
cd backend && go test ./... -count=1
```

Expected: 所有测试通过
