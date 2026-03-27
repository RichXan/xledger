const SW_VERSION = 'xledger-v2'
const SHELL_CACHE = `${SW_VERSION}-shell`
const RUNTIME_CACHE = `${SW_VERSION}-runtime`
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

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches
      .open(SHELL_CACHE)
      .then((cache) => cache.addAll(PRECACHE_ASSETS))
      .then(() => self.skipWaiting()),
  )
})

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches
      .keys()
      .then((keys) =>
        Promise.all(
          keys.filter((key) => ![SHELL_CACHE, RUNTIME_CACHE].includes(key)).map((key) => caches.delete(key)),
        ),
      )
      .then(() => self.clients.claim()),
  )
})

self.addEventListener('message', (event) => {
  if (event.data && event.data.type === 'SKIP_WAITING') {
    self.skipWaiting()
  }
})

function isSameOrigin(requestUrl) {
  return new URL(requestUrl).origin === self.location.origin
}

self.addEventListener('fetch', (event) => {
  const request = event.request
  if (request.method !== 'GET') return
  if (!isSameOrigin(request.url)) return

  const requestUrl = new URL(request.url)
  if (requestUrl.pathname.startsWith('/api/')) {
    event.respondWith(fetch(request))
    return
  }

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

  event.respondWith(
    caches.match(request).then((cached) => {
      if (cached) return cached
      return fetch(request).then((response) => {
        const clone = response.clone()
        caches.open(RUNTIME_CACHE).then((cache) => cache.put(request, clone))
        return response
      })
    }),
  )
})
