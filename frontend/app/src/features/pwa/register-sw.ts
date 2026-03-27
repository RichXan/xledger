const UPDATE_EVENT = 'xledger:pwa-update-ready'

function notifyUpdate(worker: ServiceWorker) {
  window.dispatchEvent(new CustomEvent(UPDATE_EVENT, { detail: { worker } }))
}

export function registerServiceWorker() {
  if (!('serviceWorker' in navigator)) return

  window.addEventListener('load', () => {
    navigator.serviceWorker
      .register('/sw.js')
      .then((registration) => {
        if (registration.waiting) {
          notifyUpdate(registration.waiting)
        }

        registration.addEventListener('updatefound', () => {
          const installingWorker = registration.installing
          if (!installingWorker) return

          installingWorker.addEventListener('statechange', () => {
            if (installingWorker.state === 'installed' && navigator.serviceWorker.controller) {
              notifyUpdate(installingWorker)
            }
          })
        })
      })
      .catch(() => {
        // noop: failing to register SW should not block app bootstrap
      })
  })
}

export function subscribePwaUpdate(handler: (worker: ServiceWorker) => void) {
  const listener = (event: Event) => {
    const detail = (event as CustomEvent<{ worker?: ServiceWorker }>).detail
    if (detail?.worker) {
      handler(detail.worker)
    }
  }
  window.addEventListener(UPDATE_EVENT, listener)
  return () => window.removeEventListener(UPDATE_EVENT, listener)
}

export async function applyServiceWorkerUpdate(worker: ServiceWorker) {
  worker.postMessage({ type: 'SKIP_WAITING' })
  await new Promise<void>((resolve) => {
    const done = () => {
      navigator.serviceWorker.removeEventListener('controllerchange', done)
      resolve()
    }
    navigator.serviceWorker.addEventListener('controllerchange', done)
  })
  window.location.reload()
}
