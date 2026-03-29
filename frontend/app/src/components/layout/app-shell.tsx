import type { PropsWithChildren } from 'react'
import { SideNav } from './side-nav'
import { TopBar } from './top-bar'

export function AppShell({ children }: PropsWithChildren) {
  return (
    <div className="min-h-screen bg-transparent text-on-surface md:flex">
      <SideNav />
      <div className="flex min-h-screen flex-1 flex-col">
        <TopBar />
        <main className="mx-auto w-full max-w-[1800px] flex-1 p-4 md:p-6">{children}</main>
      </div>
    </div>
  )
}
