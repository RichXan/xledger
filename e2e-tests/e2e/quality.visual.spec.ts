import { devices } from '@playwright/test'
import type { Page } from '@playwright/test'
import { expect, test } from './fixture'
import { appRoutes, installVisualStabilizers } from './helpers/quality'
import { bootstrapAuthedSession } from './helpers/ui-flows'

const { defaultBrowserType: _ignored, ...iphone13 } = devices['iPhone 13']

async function captureVisualBaselines(page: Page, viewportTag: string) {
  for (const route of appRoutes) {
    await page.goto(route.path, { waitUntil: 'networkidle' })
    await expect(page.getByRole('heading', { name: route.heading })).toBeVisible()
    await expect(page).toHaveScreenshot(`visual-${viewportTag}-${route.snapshotKey}.png`, {
      animations: 'disabled',
      fullPage: false,
      maxDiffPixelRatio: 0.02,
    })
  }
}

test.describe('visual regression @visual', () => {
  test.describe('desktop', () => {
    test.use({
      viewport: { width: 1440, height: 900 },
    })

    test('core pages visual baseline', async ({ page }, testInfo) => {
      await installVisualStabilizers(page)
      await bootstrapAuthedSession(page, testInfo)
      await captureVisualBaselines(page, 'desktop')
    })
  })

  test.describe('mobile', () => {
    test.use({
      ...iphone13,
    })

    test('core pages visual baseline', async ({ page }, testInfo) => {
      await installVisualStabilizers(page)
      await bootstrapAuthedSession(page, testInfo)
      await captureVisualBaselines(page, 'mobile')
    })
  })
})
