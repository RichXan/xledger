// frontend/app/src/components/layout/mobile-nav.tsx
import { NavLink, useLocation } from 'react-router-dom'
import { Home, Receipt, PieChart, Wallet, Settings, Upload } from 'lucide-react'
import { useTranslation } from 'react-i18next'

const navItems = [
  { to: '/dashboard', icon: Home, labelKey: 'nav.dashboard' },
  { to: '/transactions', icon: Receipt, labelKey: 'nav.transactions' },
  { to: '/analytics', icon: PieChart, labelKey: 'nav.analytics' },
  { to: '/accounts', icon: Wallet, labelKey: 'nav.accounts' },
  { to: '/import', icon: Upload, labelKey: 'nav.import' },
  { to: '/settings', icon: Settings, labelKey: 'nav.settings' },
]

export function MobileNav() {
  const { t } = useTranslation()
  const location = useLocation()

  // 只在移动端显示
  const isMobile = typeof window !== 'undefined' && window.innerWidth < 768

  if (!isMobile) return null

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-40 bg-surface border-t border-outline/15 md:hidden safe-area-inset-bottom">
      <div className="flex items-center justify-around h-16">
        {navItems.map(({ to, icon: Icon, labelKey }) => {
          const isActive = location.pathname === to || location.pathname.startsWith(to + '/')
          return (
            <NavLink
              key={to}
              to={to}
              className={`flex flex-col items-center justify-center gap-0.5 px-3 py-2 min-w-[48px] min-h-[48px] transition-colors ${
                isActive ? 'text-primary' : 'text-on-surface-variant'
              }`}
            >
              <Icon size={20} strokeWidth={isActive ? 2.5 : 2} />
              <span className="text-[10px] font-semibold">{t(labelKey)}</span>
            </NavLink>
          )
        })}
      </div>
    </nav>
  )
}