import { NavLink } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { primaryNavItems } from './nav-items'
import { cn } from '@/lib/utils'

function BrandMark() {
  return (
    <div className="grid h-10 w-10 place-items-center rounded-xl bg-primary text-white shadow-ambient">
      <svg viewBox="0 0 24 24" className="h-6 w-6" aria-hidden="true">
        <path d="M4 5h6v6H4V5Zm10 0h6v3h-6V5ZM14 10h6v9h-6v-9ZM4 13h6v6H4v-6Z" fill="currentColor" />
      </svg>
    </div>
  )
}

export function SideNav() {
  const { t } = useTranslation()

  return (
    <aside className="hidden w-[230px] shrink-0 border-r border-outline/15 bg-surface-container-lowest/90 px-4 py-5 md:flex md:flex-col md:sticky md:top-0 md:h-screen">
      <div className="mb-7 rounded-2xl px-2 py-3">
        <div className="flex items-center gap-3">
          <BrandMark />
          <div>
            <p className="font-headline text-[40px] font-extrabold leading-none tracking-tight text-primary">xledger</p>
            <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">Financial Precision</p>
          </div>
        </div>
      </div>

      <nav aria-label="Primary" className="flex flex-1 flex-col gap-1.5">
        {primaryNavItems.map(({ to, labelKey, icon: Icon }) => (
          <NavLink
            key={to}
            to={to}
            className={({ isActive }) =>
              cn(
                'group flex items-center gap-3 rounded-xl border px-3 py-3 text-sm font-semibold text-on-surface-variant transition',
                isActive
                  ? 'border-primary/15 bg-white text-primary shadow-sm'
                  : 'border-transparent hover:border-outline/20 hover:bg-white hover:text-primary',
              )
            }
          >
            <span className="grid h-6 w-6 place-items-center rounded-md bg-surface-container text-on-surface-variant group-hover:bg-primary-fixed group-hover:text-primary">
              <Icon className="h-3.5 w-3.5" />
            </span>
            <span className="font-label text-[11px] uppercase tracking-[0.13em]">{t(labelKey)}</span>
          </NavLink>
        ))}
      </nav>
    </aside>
  )
}
