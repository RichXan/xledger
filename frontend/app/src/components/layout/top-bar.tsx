import { Bell, CircleHelp, Search } from 'lucide-react'
import { useMemo } from 'react'
import { Button } from '@/components/ui/button'
import { useAuth } from '@/features/auth/auth-context'

export function TopBar() {
  const { logout, session } = useAuth()
  const displayName = useMemo(() => session?.name || session?.email || 'Ledger User', [session?.email, session?.name])

  return (
    <header className="sticky top-0 z-20 border-b border-outline/15 bg-background/95 px-6 py-4 backdrop-blur md:px-8">
      <div className="mx-auto flex w-full max-w-[1680px] items-center justify-between gap-4">
        <div className="relative hidden w-full max-w-sm lg:block">
          <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-on-surface-variant" />
          <input
            readOnly
            value=""
            placeholder="Search transactions..."
            className="h-10 w-full rounded-xl border border-outline/20 bg-surface-container-low pl-9 pr-3 text-sm text-on-surface placeholder:text-on-surface-variant/70"
          />
        </div>

        <div className="ml-auto flex items-center gap-3">
          <button
            type="button"
            className="grid h-9 w-9 place-items-center rounded-lg border border-outline/20 bg-surface-container-low text-on-surface-variant transition hover:bg-surface-container"
            aria-label="Notifications"
          >
            <Bell className="h-4 w-4" />
          </button>
          <button
            type="button"
            className="grid h-9 w-9 place-items-center rounded-lg border border-outline/20 bg-surface-container-low text-on-surface-variant transition hover:bg-surface-container"
            aria-label="Help"
          >
            <CircleHelp className="h-4 w-4" />
          </button>
          <div className="hidden items-center gap-3 rounded-xl bg-surface-container-low px-3 py-2 sm:flex">
            <div className="text-right">
              <p className="text-sm font-semibold text-on-surface">{displayName}</p>
            </div>
            <div className="grid h-8 w-8 place-items-center rounded-lg bg-primary text-xs font-bold text-white">
              {(displayName[0] ?? 'U').toUpperCase()}
            </div>
          </div>
          <Button variant="ghost" className="px-3 py-2" onClick={() => void logout()}>
            Logout
          </Button>
        </div>
      </div>
    </header>
  )
}
