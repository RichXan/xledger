import type { PropsWithChildren } from 'react'
import { SideNav } from './side-nav'
import { TopBar } from './top-bar'
import { OfflineBanner } from './offline-banner'
import { MobileNav } from './mobile-nav'

export function AppShell({ children }: PropsWithChildren) {
  return (
    <div className="min-h-screen bg-transparent text-on-surface md:flex">
      <OfflineBanner />
      <SideNav />
      <div className="flex min-h-screen flex-1 flex-col pb-16 md:pb-0">
        <TopBar />
        <main className="mx-auto w-full max-w-[1800px] flex-1 p-4 md:p-6">{children}</main>
      </div>
      <MobileNav />
    </div>
  )
}
