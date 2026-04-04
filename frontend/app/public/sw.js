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