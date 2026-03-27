import { useEffect, useState } from 'react'
import { applyServiceWorkerUpdate, subscribePwaUpdate } from './register-sw'

export function usePwaUpdate() {
  const [waitingWorker, setWaitingWorker] = useState<ServiceWorker | null>(null)
  const [updating, setUpdating] = useState(false)

  useEffect(() => {
    return subscribePwaUpdate((worker) => setWaitingWorker(worker))
  }, [])

  async function updateNow() {
    if (!waitingWorker) return
    setUpdating(true)
    try {
      await applyServiceWorkerUpdate(waitingWorker)
    } finally {
      setUpdating(false)
    }
  }

  return {
    updateAvailable: waitingWorker !== null,
    updating,
    updateNow,
  }
}
