// frontend/app/src/components/layout/mobile-nav.tsx
import { NavLink, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { mobileNavItems } from './nav-items'

export function MobileNav() {
  const { t } = useTranslation()
  const location = useLocation()

  // 只在移动端显示
  const isMobile = typeof window !== 'undefined' && window.innerWidth < 768

  if (!isMobile) return null

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-40 max-w-full overflow-hidden border-t border-outline/15 bg-surface md:hidden safe-area-inset-bottom">
      <div className="grid h-16 grid-cols-6 items-center">
        {mobileNavItems.map(({ to, icon: Icon, labelKey }) => {
          const isActive = location.pathname === to || location.pathname.startsWith(to + '/')
          return (
            <NavLink
              key={to}
              to={to}
              className={`flex min-h-[48px] min-w-0 flex-col items-center justify-center gap-0.5 px-1 py-2 transition-colors ${
                isActive ? 'text-primary' : 'text-on-surface-variant'
              }`}
            >
              <Icon size={19} strokeWidth={isActive ? 2.5 : 2} />
              <span className="max-w-full truncate text-[10px] font-semibold">{t(labelKey)}</span>
            </NavLink>
          )
        })}
      </div>
    </nav>
  )
}
