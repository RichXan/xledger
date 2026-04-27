import type { Page } from '@playwright/test'

const FIXED_TIME_ISO = '2026-01-15T12:00:00.000Z'

export const appRoutes: Array<{ path: string; heading: string; snapshotKey: string }> = [
  { path: '/dashboard?lang=en', heading: 'Financial Overview', snapshotKey: 'dashboard' },
  { path: '/transactions', heading: 'Transactions', snapshotKey: 'transactions' },
  { path: '/analytics', heading: 'Analytics', snapshotKey: 'analytics' },
  { path: '/accounts', heading: 'Accounts', snapshotKey: 'accounts' },
  { path: '/settings', heading: 'Settings', snapshotKey: 'settings' },
]

export async function installVisualStabilizers(page: Page) {
  await page.addInitScript(({ fixedTimeISO }) => {
    const fixedTime = new Date(fixedTimeISO).valueOf()
    const RealDate = Date
    type DateArgs = [] | [string | number | Date] | [number, number, number?, number?, number?, number?, number?]

    class MockDate extends RealDate {
      constructor(...args: DateArgs) {
        if (args.length === 0) {
          super(fixedTime)
          return
        }
        if (args.length === 1) {
          super(args[0])
          return
        }
        super(args[0], args[1], args[2], args[3], args[4], args[5], args[6])
      }

      static now() {
        return fixedTime
      }
    }

    Object.defineProperty(globalThis, 'Date', {
      configurable: true,
      writable: true,
      value: MockDate,
    })

    const style = document.createElement('style')
    style.setAttribute('data-e2e-visual-stable', 'true')
    style.textContent = `
      *, *::before, *::after {
        animation: none !important;
        transition: none !important;
        caret-color: transparent !important;
      }
      html { scroll-behavior: auto !important; }
    `
    document.head.appendChild(style)
  }, { fixedTimeISO: FIXED_TIME_ISO })
}
