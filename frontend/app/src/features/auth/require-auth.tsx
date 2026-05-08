import { Navigate, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuth } from './auth-context'
import { readAuthSession } from './auth-storage'

function AuthSyncingMessage() {
  const { t } = useTranslation()

  return (
    <div className="flex min-h-screen items-center justify-center bg-background text-on-surface">
      <p className="font-headline text-2xl font-bold tracking-tight">{t('layout.authGuard.syncing')}</p>
    </div>
  )
}

export function RequireAuth({ children }: { children: JSX.Element }) {
  const location = useLocation()
  const { isAuthenticated, isBootstrapping } = useAuth()
  const hasStoredSession = Boolean(readAuthSession()?.accessToken)

  if (isBootstrapping) {
    return <AuthSyncingMessage />
  }

  if (!isAuthenticated && hasStoredSession) {
    return <AuthSyncingMessage />
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />
  }

  return children
}
