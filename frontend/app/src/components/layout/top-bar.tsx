import { Button } from '@/components/ui/button'
import { useAuth } from '@/features/auth/auth-context'

export function TopBar() {
  const { logout, session } = useAuth()

  return (
    <header className="sticky top-0 z-20 flex items-center justify-between gap-4 border-b border-white/60 bg-background/80 px-6 py-4 backdrop-blur-md">
      <div>
        <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">
          Architectural Ledger
        </p>
        <h1 className="font-headline text-3xl font-extrabold tracking-tight text-on-surface">
          Financial Overview
        </h1>
      </div>
      <div className="flex items-center gap-3">
        <div className="flex items-center gap-3 rounded-full bg-surface-container-low px-4 py-2 text-sm text-on-surface-variant">
          <span className="font-medium text-on-surface">{session?.email ?? 'Ledger User'}</span>
          <span>•</span>
          <span>Self-hosted</span>
        </div>
        <Button variant="ghost" onClick={() => void logout()}>
          Logout
        </Button>
      </div>
    </header>
  )
}
