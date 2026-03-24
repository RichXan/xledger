import { Navigate, useLocation } from 'react-router-dom'
import { useAuth } from './auth-context'

export function RequireAuth({ children }: { children: JSX.Element }) {
  const location = useLocation()
  const { isAuthenticated, isBootstrapping } = useAuth()

  if (isBootstrapping) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background text-on-surface">
        <p className="font-headline text-2xl font-bold tracking-tight">Syncing your ledger…</p>
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />
  }

  return children
}
