import { NavLink } from 'react-router-dom'
import { BarChart3, LayoutDashboard, ReceiptText, Settings, Wallet } from 'lucide-react'
import { cn } from '@/lib/utils'

const items = [
  { to: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { to: '/transactions', label: 'Transactions', icon: ReceiptText },
  { to: '/analytics', label: 'Analytics', icon: BarChart3 },
  { to: '/accounts', label: 'Accounts', icon: Wallet },
  { to: '/settings', label: 'Settings', icon: Settings },
]

export function SideNav() {
  return (
    <aside className="hidden w-64 shrink-0 border-r border-white/50 bg-background/90 p-4 md:flex md:flex-col">
      <div className="mb-8 px-4 py-3">
        <div className="flex items-center gap-3">
          <div className="grid h-10 w-10 place-items-center rounded-xl bg-primary text-white shadow-ambient">
            <svg viewBox="0 0 24 24" className="h-6 w-6" aria-hidden="true">
              <path
                d="M4 5h6v6H4V5Zm10 0h6v3h-6V5ZM14 10h6v9h-6v-9ZM4 13h6v6H4v-6Z"
                fill="currentColor"
              />
            </svg>
          </div>
          <div>
            <p className="font-headline text-2xl font-extrabold tracking-tight text-primary">xledger</p>
            <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">
              Financial Precision
            </p>
          </div>
        </div>
      </div>
      <nav aria-label="Primary" className="flex flex-1 flex-col gap-1">
        {items.map(({ to, label, icon: Icon }) => (
          <NavLink
            key={to}
            to={to}
            className={({ isActive }) =>
              cn(
                'flex items-center gap-3 rounded-2xl px-4 py-3 text-sm font-semibold text-on-surface-variant transition hover:bg-surface-container-low hover:text-primary',
                isActive && 'bg-surface-container-lowest text-primary shadow-ambient',
              )
            }
          >
            <Icon className="h-4 w-4" />
            <span className="font-label text-[11px] uppercase tracking-[0.14em]">{label}</span>
          </NavLink>
        ))}
      </nav>
    </aside>
  )
}
