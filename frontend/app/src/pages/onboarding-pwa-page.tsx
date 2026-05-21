import { ArrowLeft, CheckCircle2, Download, Home, Plus, Share2, Smartphone } from 'lucide-react'
import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { usePwaInstall } from '@/features/pwa/use-pwa-install'

function getInstallPlatform() {
  if (typeof navigator === 'undefined') return 'desktop'
  const ua = navigator.userAgent
  if (/iPad|iPhone|iPod/.test(ua)) return 'ios'
  if (/Android/i.test(ua)) return 'android'
  return 'desktop'
}

export function getStandaloneStatus() {
  if (typeof window === 'undefined') return false
  const displayModeStandalone = window.matchMedia?.('(display-mode: standalone)').matches ?? false
  const navigatorStandalone = 'standalone' in navigator && Boolean((navigator as Navigator & { standalone?: boolean }).standalone)
  return displayModeStandalone || navigatorStandalone
}

interface PWAOnboardingPageProps {
  mobileEntry?: boolean
}

export function PWAOnboardingPage({ mobileEntry = false }: PWAOnboardingPageProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { canInstall, install } = usePwaInstall()
  const platform = useMemo(() => getInstallPlatform(), [])
  const isStandalone = useMemo(() => getStandaloneStatus(), [])

  function leaveInstallGuide(target?: string) {
    window.localStorage.setItem('pwa-onboarding-dismissed', 'true')
    if (target) {
      navigate(target)
      return
    }
    navigate(-1)
  }

  const steps = platform === 'ios'
    ? [
        { icon: Share2, label: t('pwa.iosStep1') },
        { icon: Plus, label: t('pwa.iosStep2') },
        { icon: Home, label: t('pwa.iosStep3') },
      ]
    : platform === 'android'
      ? [
          { icon: Download, label: t('pwa.androidStep1') },
          { icon: Plus, label: t('pwa.androidStep2') },
          { icon: Home, label: t('pwa.androidStep3') },
        ]
      : [
          { icon: Download, label: t('pwa.desktopStep1') },
          { icon: Plus, label: t('pwa.desktopStep2') },
          { icon: Home, label: t('pwa.desktopStep3') },
        ]

  return (
    <div className="mx-auto max-w-5xl space-y-5">
      <section className="overflow-hidden rounded-2xl border border-outline/15 bg-surface-container-lowest shadow-ambient">
        <div className="grid gap-6 p-5 md:grid-cols-[1.05fr_0.95fr] md:p-7">
          <div>
            {mobileEntry ? null : (
              <button
                type="button"
                className="mb-5 inline-flex min-h-10 items-center gap-2 rounded-lg border border-outline/15 bg-surface-container-low px-3 text-sm font-bold text-on-surface-variant"
                onClick={() => leaveInstallGuide()}
              >
                <ArrowLeft className="h-4 w-4" />
                {t('common.back')}
              </button>
            )}
            <div className="inline-flex items-center gap-2 rounded-full bg-primary-fixed px-3 py-1.5 text-xs font-bold text-primary">
              <Smartphone className="h-4 w-4" />
              {mobileEntry ? t('pwa.mobileEntryEyebrow') : t('pwa.eyebrow')}
            </div>
            <h2 className="mt-4 font-headline text-4xl font-extrabold leading-tight text-primary md:text-[44px]">
              {mobileEntry ? t('pwa.mobileEntryTitle') : t('pwa.installTitle')}
            </h2>
            <p className="mt-3 max-w-2xl text-sm leading-6 text-on-surface-variant">
              {mobileEntry ? t('pwa.mobileEntryDescription') : t('pwa.installDescription')}
            </p>
            <div className="mt-5 flex flex-wrap gap-2">
              {canInstall ? (
                <Button onClick={() => void install()}>
                  <Download className="h-4 w-4" />
                  {t('pwa.installNow')}
                </Button>
              ) : null}
              <Button variant="secondary" onClick={() => leaveInstallGuide('/dashboard')}>
                {mobileEntry ? t('pwa.openApp') : t('pwa.backToApp')}
              </Button>
            </div>
          </div>

          <div className="rounded-2xl border border-primary/15 bg-primary-fixed p-4">
            <div className="flex items-start gap-3 rounded-xl bg-white p-4">
              <div className="grid h-10 w-10 shrink-0 place-items-center rounded-xl bg-primary text-white">
                {isStandalone ? <CheckCircle2 className="h-5 w-5" /> : <Smartphone className="h-5 w-5" />}
              </div>
              <div>
                <p className="font-bold text-on-surface">
                  {isStandalone ? t('pwa.installedStatus') : t('pwa.browserStatus')}
                </p>
                <p className="mt-1 text-sm leading-6 text-on-surface-variant">
                  {isStandalone ? t('pwa.installedDescription') : t('pwa.browserDescription')}
                </p>
              </div>
            </div>
            <div className="mt-3 grid gap-2 sm:grid-cols-3">
              {[t('pwa.benefitFast'), t('pwa.benefitHome'), t('pwa.benefitFocus')].map((benefit) => (
                <div key={benefit} className="rounded-xl bg-white/80 px-3 py-3 text-xs font-bold text-primary">
                  {benefit}
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>

      <section className="grid gap-3 md:grid-cols-3">
        {steps.map(({ icon: Icon, label }, index) => (
          <article key={label} className="rounded-2xl border border-outline/15 bg-white p-4">
            <div className="flex items-center gap-3">
              <div className="grid h-10 w-10 shrink-0 place-items-center rounded-xl bg-surface-container-low text-primary">
                <Icon className="h-5 w-5" />
              </div>
              <span className="text-xs font-black uppercase tracking-[0.14em] text-on-surface-variant">
                {t('pwa.stepLabel', { count: index + 1 })}
              </span>
            </div>
            <p className="mt-4 text-sm font-semibold leading-6 text-on-surface">{label}</p>
          </article>
        ))}
      </section>

      <section className="rounded-2xl border border-outline/15 bg-surface-container-low p-5">
        <p className="font-label text-[10px] font-bold uppercase tracking-[0.18em] text-on-surface-variant">
          {t('pwa.longTermTitle')}
        </p>
        <div className="mt-4 grid gap-3 md:grid-cols-3">
          {[t('pwa.longTermHttps'), t('pwa.longTermLogin'), t('pwa.longTermUpdate')].map((item) => (
            <div key={item} className="flex items-start gap-3 rounded-xl bg-white p-4 text-sm leading-6 text-on-surface">
              <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0 text-emerald-600" />
              <span>{item}</span>
            </div>
          ))}
        </div>
      </section>
    </div>
  )
}
