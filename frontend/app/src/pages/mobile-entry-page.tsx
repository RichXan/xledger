import { Navigate } from 'react-router-dom'
import { PWAOnboardingPage, getStandaloneStatus } from './onboarding-pwa-page'

export function MobileEntryPage() {
  const onboardingDismissed = window.localStorage.getItem('pwa-onboarding-dismissed') === 'true'

  if (getStandaloneStatus() || onboardingDismissed) {
    return <Navigate to="/dashboard" replace />
  }

  return (
    <main className="min-h-screen bg-background px-4 py-5 text-on-surface sm:px-6">
      <PWAOnboardingPage mobileEntry />
    </main>
  )
}
