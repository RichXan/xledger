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