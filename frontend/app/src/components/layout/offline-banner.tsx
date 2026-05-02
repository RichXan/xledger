// frontend/app/src/components/layout/offline-banner.tsx
import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'

export function OfflineBanner() {
  const { t } = useTranslation()
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
      {t('layout.offline.message')}
    </div>
  )
}
