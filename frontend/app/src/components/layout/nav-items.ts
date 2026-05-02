import {
  BarChart3,
  Home,
  LayoutDashboard,
  PieChart,
  Receipt,
  ReceiptText,
  Settings,
  Upload,
  Wallet,
  Zap,
  type LucideIcon,
} from 'lucide-react'

export interface NavItem {
  to: string
  labelKey: string
  icon: LucideIcon
}

export const primaryNavItems: NavItem[] = [
  { to: '/dashboard', labelKey: 'nav.dashboard', icon: LayoutDashboard },
  { to: '/transactions', labelKey: 'nav.transactions', icon: ReceiptText },
  { to: '/analytics', labelKey: 'nav.analytics', icon: BarChart3 },
  { to: '/accounts', labelKey: 'nav.accounts', icon: Wallet },
  { to: '/shortcut', labelKey: 'nav.quickEntry', icon: Zap },
  { to: '/settings', labelKey: 'nav.settings', icon: Settings },
]

export const mobileNavItems: NavItem[] = [
  { to: '/dashboard', labelKey: 'nav.dashboard', icon: Home },
  { to: '/transactions', labelKey: 'nav.transactions', icon: Receipt },
  { to: '/analytics', labelKey: 'nav.analytics', icon: PieChart },
  { to: '/accounts', labelKey: 'nav.accounts', icon: Wallet },
  { to: '/import', labelKey: 'nav.import', icon: Upload },
  { to: '/settings', labelKey: 'nav.settings', icon: Settings },
]
