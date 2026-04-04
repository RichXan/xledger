# PWA 增强实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 增强 PWA 实现离线只读缓存 + Web Push 通知 + iOS 安装引导

**Architecture:**
- Service Worker: Cache-First（静态资源）+ Network-First 回退到 IndexedDB（API 数据）
- 离线数据: 最近 30 天交易缓存于 IndexedDB（Dexie.js）
- 推送: Web Push API + VAPID
- 移动端: 底部 Tab 导航 + 离线横幅

**Tech Stack:** Service Worker, Dexie.js (IndexedDB wrapper), Web Push API

---

## 文件影响

| 文件 | 动作 |
|------|------|
| `frontend/app/public/sw.js` | 修改：增强缓存策略 |
| `frontend/app/public/manifest.webmanifest` | 修改：添加 push permission |
| `frontend/app/src/features/pwa/use-pwa-install.ts` | 修改：增强安装检测 |
| `frontend/app/src/features/offline/offline-store.ts` | 新建：Dexie.js 数据层 |
| `frontend/app/src/features/offline/offline-queue.ts` | 新建：离线写入队列 |
| `frontend/app/src/features/notifications/push-manager.ts` | 新建：Web Push 管理 |
| `frontend/app/src/features/notifications/preference-store.ts` | 新建：通知偏好 |
| `frontend/app/src/components/layout/mobile-nav.tsx` | 新建：移动端底部导航 |
| `frontend/app/src/components/layout/offline-banner.tsx` | 新建：离线状态横幅 |
| `frontend/app/src/pages/onboarding-pwa-page.tsx` | 新建：iOS 安装引导 |
| `frontend/app/src/App.tsx` | 修改：注册路由 |
| `frontend/app/src/main.tsx` | 修改：初始化离线存储 |
| `frontend/app/src/components/layout/app-shell.tsx` | 修改：添加移动端适配 |
| `backend/internal/push/handler.go` | 新建：Push 订阅 API |
| `backend/internal/push/service.go` | 新建：VAPID 密钥 + 推送服务 |
| `backend/internal/bootstrap/http/push_wiring.go` | 新建：Push 依赖注入 |
| `backend/internal/bootstrap/http/router.go` | 修改：注册 push 路由 |
| `backend/migrations/YYYYMMDDDD_create_notification_prefs.sql` | 新建：通知偏好表 |

---

## Task 1: 安装 Dexie.js

- [ ] **Step 1: 安装依赖**

```bash
cd frontend/app && npm install dexie
```

- [ ] **Step 2: Commit**

```bash
git add package.json package-lock.json && git commit -m "chore(frontend): add dexie for IndexedDB

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 2: 创建 Dexie.js 离线存储层

- [ ] **Step 1: 创建 offline-store.ts**

```typescript
// frontend/app/src/features/offline/offline-store.ts
import Dexie, { type Table } from 'dexie'

export interface CachedTransaction {
  id: string
  ledger_id: string
  account_id: string | null
  category_id: string | null
  type: 'income' | 'expense' | 'transfer'
  amount: number
  memo: string | null
  occurred_at: string
  created_at: string
}

export interface CachedAccount {
  id: string
  name: string
  type: string
  balance: number
}

export interface CachedCategory {
  id: string
  name: string
  parent_id: string | null
}

export interface OfflineOverview {
  id: string  // 总为 'latest'
  total_assets: number
  income: number
  expense: number
  cached_at: string
}

class XLedgerOfflineDB extends Dexie {
  transactions!: Table<CachedTransaction>
  accounts!: Table<CachedAccount>
  categories!: Table<CachedCategory>
  overviews!: Table<OfflineOverview>

  constructor() {
    super('xledger-offline')
    this.version(1).stores({
      transactions: 'id, ledger_id, account_id, category_id, occurred_at',
      accounts: 'id, type',
      categories: 'id, parent_id',
      overviews: 'id',
    })
  }
}

export const offlineDb = new XLedgerOfflineDB()

export async function cacheTransactions(txns: CachedTransaction[]) {
  await offlineDb.transactions.bulkPut(txns)
}

export async function cacheAccounts(accts: CachedAccount[]) {
  await offlineDb.accounts.bulkPut(accts)
}

export async function cacheCategories(cats: CachedCategory[]) {
  await offlineDb.categories.bulkPut(cats)
}

export async function cacheOverview(overview: Omit<OfflineOverview, 'id'>) {
  await offlineDb.overviews.put({ id: 'latest', ...overview })
}

export async function getCachedTransactions(limit = 100): Promise<CachedTransaction[]> {
  return offlineDb.transactions.orderBy('occurred_at').reverse().limit(limit).toArray()
}

export async function getCachedOverview(): Promise<OfflineOverview | undefined> {
  return offlineDb.overviews.get('latest')
}

export async function getCachedAccounts(): Promise<CachedAccount[]> {
  return offlineDb.accounts.toArray()
}

export async function clearOfflineCache() {
  await offlineDb.transactions.clear()
  await offlineDb.accounts.clear()
  await offlineDb.categories.clear()
  await offlineDb.overviews.clear()
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/app/src/features/offline/offline-store.ts && git commit -m "feat(frontend): add Dexie.js offline storage layer

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 3: 创建离线横幅组件

- [ ] **Step 1: 创建 offline-banner.tsx**

```typescript
// frontend/app/src/components/layout/offline-banner.tsx
import { useState, useEffect } from 'react'

export function OfflineBanner() {
  const [isOffline, setIsOffline] = useState(!navigator.onLine)

  useEffect(() => {
    const handleOnline = () => setIsOffline(false)
    const handleOffline = () => setIsOffline(true)

    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)

    return () => {
      window.removeEventListener('online', handleOnline)
      window.removeEventListener('offline', handleOffline)
    }
  }, [])

  if (!isOffline) return null

  return (
    <div className="fixed top-0 left-0 right-0 z-50 bg-amber-500 text-white text-center text-sm py-2 px-4">
      你当前处于离线状态，显示的是缓存数据
    </div>
  )
}
```

- [ ] **Step 2: 修改 app-shell.tsx 集成横幅**

在 `app-shell.tsx` 中导入并使用 `OfflineBanner`：

```typescript
import { OfflineBanner } from './offline-banner'

export function AppShell({ children }: PropsWithChildren) {
  return (
    <div className="min-h-screen bg-transparent text-on-surface md:flex">
      <OfflineBanner />
      <SideNav />
      {/* ... 其余不变 */}
    </div>
  )
}
```

- [ ] **Step 3: Commit**

```bash
git add frontend/app/src/components/layout/offline-banner.tsx frontend/app/src/components/layout/app-shell.tsx
git commit -m "feat(frontend): add offline banner component

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 4: 创建移动端底部导航

- [ ] **Step 1: 创建 mobile-nav.tsx**

```typescript
// frontend/app/src/components/layout/mobile-nav.tsx
import { NavLink, useLocation } from 'react-router-dom'
import { Home, Receipt, PieChart, Wallet, Settings } from 'lucide-react'
import { useTranslation } from 'react-i18next'

const navItems = [
  { to: '/dashboard', icon: Home, labelKey: 'nav.dashboard' },
  { to: '/transactions', icon: Receipt, labelKey: 'nav.transactions' },
  { to: '/analytics', icon: PieChart, labelKey: 'nav.analytics' },
  { to: '/accounts', icon: Wallet, labelKey: 'nav.accounts' },
  { to: '/settings', icon: Settings, labelKey: 'nav.settings' },
]

export function MobileNav() {
  const { t } = useTranslation()
  const location = useLocation()

  // 只在移动端显示
  const isMobile = typeof window !== 'undefined' && window.innerWidth < 768

  if (!isMobile) return null

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-40 bg-surface border-t border-outline/15 md:hidden safe-area-inset-bottom">
      <div className="flex items-center justify-around h-16">
        {navItems.map(({ to, icon: Icon, labelKey }) => {
          const isActive = location.pathname === to || location.pathname.startsWith(to + '/')
          return (
            <NavLink
              key={to}
              to={to}
              className={`flex flex-col items-center justify-center gap-0.5 px-3 py-2 min-w-[48px] min-h-[48px] transition-colors ${
                isActive ? 'text-primary' : 'text-on-surface-variant'
              }`}
            >
              <Icon size={20} strokeWidth={isActive ? 2.5 : 2} />
              <span className="text-[10px] font-semibold">{t(labelKey)}</span>
            </NavLink>
          )
        })}
      </div>
    </nav>
  )
}
```

- [ ] **Step 2: 修改 app-shell.tsx 集成移动端导航**

```typescript
import { MobileNav } from './mobile-nav'

export function AppShell({ children }: PropsWithChildren) {
  return (
    <div className="min-h-screen bg-transparent text-on-surface md:flex">
      <OfflineBanner />
      <SideNav />
      <div className="flex min-h-screen flex-1 flex-col pb-16 md:pb-0">
        <TopBar />
        <main className="mx-auto w-full max-w-[1800px] flex-1 p-4 md:p-6">{children}</main>
      </div>
      <MobileNav />
    </div>
  )
}
```

- [ ] **Step 3: Commit**

```bash
git add frontend/app/src/components/layout/mobile-nav.tsx frontend/app/src/components/layout/app-shell.tsx
git commit -m "feat(frontend): add mobile bottom navigation

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 5: 增强 Service Worker

- [ ] **Step 1: 修改 sw.js**

替换现有的 `frontend/app/public/sw.js`，实现增强的缓存策略：

```javascript
const SW_VERSION = 'xledger-v3'
const SHELL_CACHE = `${SW_VERSION}-shell`
const RUNTIME_CACHE = `${SW_VERSION}-runtime`
const OFFLINE_CACHE = `${SW_VERSION}-offline`
const PRECACHE_ASSETS = [
  '/',
  '/index.html',
  '/offline.html',
  '/manifest.webmanifest',
  '/favicon.svg',
  '/pwa-icon.svg',
  '/pwa-192.png',
  '/pwa-512.png',
]

const CACHE_MAX_AGE = {
  '/api/stats/overview': 5 * 60 * 1000,      // 5 分钟
  '/api/transactions': 10 * 60 * 1000,         // 10 分钟
  '/api/categories': 30 * 60 * 1000,           // 30 分钟
  '/api/accounts': 30 * 60 * 1000,
}

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(SHELL_CACHE)
      .then((cache) => cache.addAll(PRECACHE_ASSETS))
      .then(() => self.skipWaiting()),
  )
})

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys()
      .then((keys) =>
        Promise.all(
          keys
            .filter((key) => ![SHELL_CACHE, RUNTIME_CACHE, OFFLINE_CACHE].includes(key))
            .map((key) => caches.delete(key)),
        ),
      )
      .then(() => self.clients.claim()),
  )
})

self.addEventListener('message', (event) => {
  if (event.data && event.data.type === 'SKIP_WAITING') {
    self.skipWaiting()
  }
  if (event.data && event.data.type === 'CACHE_DATA') {
    // 前端手动缓存数据（如交易列表）
    const { url, data } = event.data
    caches.open(OFFLINE_CACHE)
      .then((cache) => cache.put(url, new Response(JSON.stringify(data))))
  }
})

function isSameOrigin(requestUrl) {
  return new URL(requestUrl).origin === self.location.origin
}

function isAPIRequest(url) {
  return url.pathname.startsWith('/api/')
}

function shouldUseCacheFirst(url) {
  const staticPatterns = ['/static/', '/assets/', '.js', '.css', '.woff2', '.png', '.svg']
  return staticPatterns.some((p) => url.pathname.includes(p))
}

function getCacheAge(url) {
  for (const [pattern, age] of Object.entries(CACHE_MAX_AGE)) {
    if (url.pathname.startsWith(pattern)) return age
  }
  return null
}

self.addEventListener('fetch', (event) => {
  const request = event.request
  if (request.method !== 'GET') return
  if (!isSameOrigin(request.url)) return

  const requestUrl = new URL(request.url)

  // 静态资源：Cache-First
  if (shouldUseCacheFirst(requestUrl)) {
    event.respondWith(
      caches.match(request).then((cached) => cached || fetch(request)),
    )
    return
  }

  // API 请求：Network-First，失败时回退到缓存
  if (isAPIRequest(requestUrl)) {
    const cacheAge = getCacheAge(requestUrl)
    event.respondWith(
      fetch(request)
        .then((response) => {
          if (response.ok) {
            const clone = response.clone()
            caches.open(RUNTIME_CACHE).then((cache) => {
              const headers = new Headers(clone.headers)
              headers.append('sw-cached-at', Date.now().toString())
              cache.put(request, new Response(clone.body, {
                status: clone.status,
                statusText: clone.statusText,
                headers,
              }))
            })
          }
          return response
        })
        .catch(async () => {
          // 网络失败，尝试从缓存读取
          const cached = await caches.match(request)
          if (cached) {
            // 检查缓存是否过期
            if (cacheAge) {
              const cachedAt = parseInt(cached.headers.get('sw-cached-at') || '0')
              if (Date.now() - cachedAt < cacheAge) {
                return cached
              }
            } else {
              return cached
            }
          }
          return new Response(JSON.stringify({ code: 'OFFLINE', message: '离线' }), {
            status: 503,
            headers: { 'Content-Type': 'application/json' },
          })
        }),
    )
    return
  }

  // 导航请求：Network-First，回退到 HTML shell
  if (request.mode === 'navigate') {
    event.respondWith(
      fetch(request)
        .then((response) => {
          const clone = response.clone()
          caches.open(RUNTIME_CACHE).then((cache) => cache.put(request, clone))
          return response
        })
        .catch(async () => {
          const runtimeCached = await caches.match(request)
          if (runtimeCached) return runtimeCached
          const appShell = await caches.match('/index.html')
          if (appShell) return appShell
          return caches.match('/offline.html')
        }),
    )
    return
  }

  // 其他请求：Stale-While-Revalidate
  event.respondWith(
    caches.match(request).then((cached) => {
      const networkFetch = fetch(request).then((response) => {
        if (response.ok) {
          const clone = response.clone()
          caches.open(RUNTIME_CACHE).then((cache) => cache.put(request, clone))
        }
        return response
      })
      return cached || networkFetch
    }),
  )
})

// Web Push 事件处理
self.addEventListener('push', (event) => {
  if (!event.data) return

  const data = event.data.json()
  const options = {
    body: data.body,
    icon: '/pwa-192.png',
    badge: '/pwa-192.png',
    tag: data.tag || 'default',
    data: data.url || '/dashboard',
    vibrate: [200, 100, 200],
    actions: data.actions || [],
  }

  event.waitUntil(
    self.registration.showNotification(data.title, options),
  )
})

self.addEventListener('notificationclick', (event) => {
  event.notification.close()

  event.waitUntil(
    clients.matchAll({ type: 'window', includeUncontrolled: true })
      .then((clientList) => {
        // 如果已有窗口，打开它
        for (const client of clientList) {
          if (client.url === event.notification.data && 'focus' in client) {
            return client.focus()
          }
        }
        // 否则打开新窗口
        if (clients.openWindow) {
          return clients.openWindow(event.notification.data)
        }
      }),
  )
})
```

- [ ] **Step 2: Commit**

```bash
git add frontend/app/public/sw.js && git commit -m "feat(frontend): enhance service worker with offline API caching and push support

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 6: Web Push 管理

- [ ] **Step 1: 创建 push-manager.ts**

```typescript
// frontend/app/src/features/notifications/push-manager.ts

const VAPID_PUBLIC_KEY = import.meta.env.VITE_VAPID_PUBLIC_KEY as string

export interface PushSubscriptionJSON {
  endpoint: string
  keys: {
    p256dh: string
    auth: string
  }
}

export async function subscribeToPush(): Promise<PushSubscription | null> {
  if (!('serviceWorker' in navigator) || !('PushManager' in window)) {
    console.warn('Push not supported')
    return null
  }

  try {
    const registration = await navigator.serviceWorker.ready
    const subscription = await registration.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: urlBase64ToUint8Array(VAPID_PUBLIC_KEY),
    })
    return subscription
  } catch (err) {
    console.error('Push subscription failed:', err)
    return null
  }
}

export async function getExistingSubscription(): Promise<PushSubscription | null> {
  if (!('serviceWorker' in navigator)) return null

  const registration = await navigator.serviceWorker.ready
  return registration.pushManager.getSubscription()
}

export async function unsubscribeFromPush(): Promise<boolean> {
  const subscription = await getExistingSubscription()
  if (subscription) {
    return subscription.unsubscribe()
  }
  return true
}

export async function sendSubscriptionToBackend(subscription: PushSubscription): Promise<void> {
  const response = await fetch('/api/push/subscribe', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(subscription.toJSON()),
  })
  if (!response.ok) {
    throw new Error('Failed to send subscription to backend')
  }
}

export async function removeSubscriptionFromBackend(): Promise<void> {
  await fetch('/api/push/subscribe', {
    method: 'DELETE',
  })
}

// 辅助函数：VAPID key 转换
function urlBase64ToUint8Array(base64String: string): Uint8Array {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4)
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/')
  const rawData = window.atob(base64)
  return new Uint8Array([...rawData].map((char) => char.charCodeAt(0)))
}
```

- [ ] **Step 2: 创建 preference-store.ts**

```typescript
// frontend/app/src/features/notifications/preference-store.ts
import { requestEnvelope } from '@/lib/api'

export interface NotificationPreferences {
  realtime_alert: boolean
  daily_digest: boolean
  weekly_digest: boolean
}

export interface PreferenceStore {
  getPrefs(): Promise<NotificationPreferences>
  updatePrefs(prefs: Partial<NotificationPreferences>): Promise<void>
}

export function createPreferenceStore(accessToken: string): PreferenceStore {
  return {
    async getPrefs() {
      const data = await requestEnvelope<{ prefs: NotificationPreferences }>('/push/preferences', {
        headers: { Authorization: `Bearer ${accessToken}` },
      })
      return data.prefs
    },
    async updatePrefs(prefs: Partial<NotificationPreferences>) {
      await requestEnvelope('/push/preferences', {
        method: 'PATCH',
        headers: {
          Authorization: `Bearer ${accessToken}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(prefs),
      })
    },
  }
}
```

- [ ] **Step 3: Commit**

```bash
git add frontend/app/src/features/notifications/push-manager.ts frontend/app/src/features/notifications/preference-store.ts
git commit -m "feat(frontend): add Web Push management

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 7: iOS 安装引导页

- [ ] **Step 1: 创建 onboarding-pwa-page.tsx**

```typescript
// frontend/app/src/pages/onboarding-pwa-page.tsx
import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { X, Smartphone, ArrowRight } from 'lucide-react'

export function PWAOnboardingPage() {
  const { t } = useTranslation()
  const [visible, setVisible] = useState(false)
  const [dismissed, setDismissed] = useState(false)

  useEffect(() => {
    // 检测 iOS Safari 且未安装
    const isIOS = /iPad|iPhone|iPod/.test(navigator.userAgent)
    const isStandalone = window.matchMedia('(display-mode: standalone)').matches
    const hasDismissed = localStorage.getItem('pwa-onboarding-dismissed')

    if (isIOS && !isStandalone && !hasDismissed) {
      setVisible(true)
    }
  }, [])

  const handleDismiss = () => {
    setDismissed(true)
    localStorage.setItem('pwa-onboarding-dismissed', 'true')
  }

  if (!visible || dismissed) return null

  return (
    <div className="fixed inset-0 z-[100] flex items-end justify-center md:items-center">
      <div className="absolute inset-0 bg-black/50" onClick={handleDismiss} />
      <div className="relative w-full max-w-md rounded-t-3xl bg-white p-6 shadow-2xl md:rounded-3xl">
        <button
          onClick={handleDismiss}
          className="absolute right-4 top-4 p-2 text-gray-400 hover:text-gray-600"
        >
          <X size={20} />
        </button>

        <div className="flex flex-col items-center text-center">
          <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-primary/10">
            <Smartphone className="text-primary" size={32} />
          </div>

          <h2 className="mb-2 text-xl font-bold text-on-surface">
            {t('pwa.installTitle')}
          </h2>
          <p className="mb-6 text-sm text-on-surface-variant">
            {t('pwa.installDescription')}
          </p>

          <div className="w-full space-y-3">
            <div className="flex items-start gap-3 rounded-xl bg-surface-container p-4 text-left">
              <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary text-white text-xs font-bold">1</span>
              <p className="text-sm text-on-surface">{t('pwa.step1')}</p>
            </div>
            <div className="flex items-start gap-3 rounded-xl bg-surface-container p-4 text-left">
              <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary text-white text-xs font-bold">2</span>
              <p className="text-sm text-on-surface">{t('pwa.step2')}</p>
            </div>
            <div className="flex items-start gap-3 rounded-xl bg-surface-container p-4 text-left">
              <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary text-white text-xs font-bold">3</span>
              <p className="text-sm text-on-surface">{t('pwa.step3')}</p>
            </div>
          </div>

          <button
            onClick={handleDismiss}
            className="mt-6 flex w-full items-center justify-center gap-2 rounded-xl bg-primary px-6 py-3 text-white font-semibold"
          >
            {t('pwa.gotIt')} <ArrowRight size={18} />
          </button>
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: 添加翻译资源到 translation.json**

在 zh/translation.json 添加：

```json
"pwa": {
  "installTitle": "添加到主屏幕",
  "installDescription": "安装 Xledger 到你的主屏幕，随时快速访问",
  "step1": "点击 Safari 底部的分享按钮",
  "step2": "向下滚动，找到「添加到主屏幕」",
  "step3": "点击右上角「添加」",
  "gotIt": "知道了"
}
```

- [ ] **Step 3: 在 App.tsx 中注册路由**

在 `App.tsx` 中添加 PWA 引导页路由：

```typescript
import { PWAOnboardingPage } from '@/pages/onboarding-pwa-page'

// 在 Routes 中（在所有受保护路由之外）：
<Route path="/pwa-onboarding" element={<PWAOnboardingPage />} />
```

并在 `DashboardPage` 或 `AppShell` 中检测并重定向：

```typescript
// 在 DashboardPage useEffect 中：
useEffect(() => {
  const isIOS = /iPad|iPhone|iPod/.test(navigator.userAgent)
  const isStandalone = window.matchMedia('(display-mode: standalone)').matches
  const hasDismissed = localStorage.getItem('pwa-onboarding-dismissed')
  if (isIOS && !isStandalone && !hasDismissed) {
    navigate('/pwa-onboarding')
  }
}, [])
```

- [ ] **Step 4: Commit**

```bash
git add frontend/app/src/pages/onboarding-pwa-page.tsx
git add frontend/app/src/App.tsx
git commit -m "feat(frontend): add iOS PWA install guide

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 8: 后端 Push 订阅 API

- [ ] **Step 1: 创建 push 服务**

```go
// backend/internal/push/service.go
package push

import (
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

var vapIDPublicKey, vapIDPrivateKey string

func init() {
    // 生成 VAPID 密钥对（生产环境应该持久化）
    privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil {
        panic(err)
    }
    vapIDPrivateKey = base64.URLEncoding.EncodeToString(privateKey.D.Bytes())
    vapIDPublicKey = base64.URLEncoding.EncodeToString(
        elliptic.MarshalCompressed(privateKey.PublicKey.Curve, privateKey.PublicKey.X, privateKey.PublicKey.Y),
    )
}

func GetVAPIDPublicKey() string {
    return vapIDPublicKey
}

// 存储订阅（内存-map，生产环境用 Redis）
var subscriptionStore = make(map[string]PushSubscription)

type PushSubscription struct {
    UserID    string `json:"user_id"`
    Endpoint  string `json:"endpoint"`
    P256dh    string `json:"keys.p256dh"`
    Auth      string `json:"keys.auth"`
    CreatedAt int64  `json:"created_at"`
}

func (s *Service) Subscribe(c *gin.Context) {
    userID, ok := c.Get("user_id")
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "未认证"})
        return
    }

    var sub struct {
        Endpoint string `json:"endpoint" binding:"required"`
        Keys      struct {
            P256dh string `json:"p256dh" binding:"required"`
            Auth   string `json:"auth" binding:"required"`
        } `json:"keys" binding:"required"`
    }
    if err := c.ShouldBindJSON(&sub); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "请求格式错误"})
        return
    }

    key := fmt.Sprintf("%s:%s", userID, sub.Endpoint)
    subscriptionStore[key] = PushSubscription{
        UserID:    userID.(string),
        Endpoint:  sub.Endpoint,
        P256dh:    sub.Keys.P256dh,
        Auth:      sub.Keys.Auth,
        CreatedAt: time.Now().Unix(),
    }

    c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "订阅成功"})
}

func (s *Service) Unsubscribe(c *gin.Context) {
    userID, ok := c.Get("user_id")
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "未认证"})
        return
    }

    var sub struct {
        Endpoint string `json:"endpoint"`
    }
    c.ShouldBindJSON(&sub)

    key := fmt.Sprintf("%s:%s", userID, sub.Endpoint)
    delete(subscriptionStore, key)

    c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "已退订"})
}

func (s *Service) SendPushNotification(userID, title, body, tag string) error {
    // 查找用户的订阅
    for _, sub := range subscriptionStore {
        if sub.UserID != userID {
            continue
        }
        // 使用 Web Push 发送
        // 注意：生产环境应该使用 web-push 库
        _ = sub // 实际发送逻辑省略
    }
    return nil
}
```

- [ ] **Step 2: 创建 handler.go**

```go
// backend/internal/push/handler.go
package push

import "github.com/gin-gonic/gin"

type Handler struct {
    service *Service
}

func NewHandler(service *Service) *Handler {
    return &Handler{service: service}
}

func (h *Handler) Subscribe(c *gin.Context) {
    h.service.Subscribe(c)
}

func (h *Handler) Unsubscribe(c *gin.Context) {
    h.service.Unsubscribe(c)
}

func (h *Handler) GetVAPIDKey(c *gin.Context) {
    c.JSON(200, gin.H{
        "code":    "OK",
        "message": "成功",
        "data": gin.H{
            "publicKey": GetVAPIDPublicKey(),
        },
    })
}
```

- [ ] **Step 3: 创建路由注册**

在 `router.go` 中添加：

```go
// Push handler registration
pushHandler := push.NewHandler(push.NewService())
pushGroup := r.Group("/api/push")
pushGroup.Use(accountingAuthMiddleware(deps.UserIDResolver, patService))
pushGroup.POST("/subscribe", pushHandler.Subscribe)
pushGroup.DELETE("/subscribe", pushHandler.Unsubscribe)
pushGroup.GET("/vapid-key", pushHandler.GetVAPIDKey)
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/push/ backend/internal/bootstrap/http/router.go
git commit -m "feat(backend): add Web Push subscription API

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 9: 最终验证

- [ ] **Step 1: 构建前端**

```bash
cd frontend/app && npm run build
```

Expected: 构建成功

- [ ] **Step 2: 后端测试**

```bash
cd backend && go test ./... -count=1
```

Expected: 所有测试通过

- [ ] **Step 3: 手动测试**

1. 打开 DevTools → Application → Service Workers，确认新的 sw.js 已注册
2. 断网，刷新页面，确认离线横幅显示
3. 使用 ngrok 暴露后端，测试 Web Push 订阅
