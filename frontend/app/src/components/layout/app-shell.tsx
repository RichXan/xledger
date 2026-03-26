import type { PropsWithChildren } from 'react'
import { SideNav } from './side-nav'
import { TopBar } from './top-bar'

export function AppShell({ children }: PropsWithChildren) {
  return (
    <div className="min-h-screen bg-background text-on-surface md:flex">
      <SideNav />
      <div className="flex min-h-screen flex-1 flex-col">
        <TopBar />
        <main className="mx-auto w-full max-w-[1680px] flex-1 p-6 md:p-8">{children}</main>
      </div>
    </div>
  )
}
