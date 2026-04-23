import AxeBuilder from '@axe-core/playwright'
import { expect, test } from './fixture'
import { appRoutes, installVisualStabilizers } from './helpers/quality'
import { bootstrapAuthedSession } from './helpers/ui-flows'

test.describe('accessibility audit @a11y', () => {
  test('core pages have no critical wcag2a/aa violations', async ({ page, recordToReport }, testInfo) => {
    await installVisualStabilizers(page)
    await bootstrapAuthedSession(page, testInfo)

    const seriousSummary: Array<{ page: string; count: number }> = []

    for (const route of appRoutes) {
      await page.goto(route.path, { waitUntil: 'networkidle' })
      await expect(page.getByRole('heading', { name: route.heading })).toBeVisible()

      const result = await new AxeBuilder({ page })
        .withTags(['wcag2a', 'wcag2aa'])
        .analyze()

      const critical = result.violations.filter((item) => item.impact === 'critical')
      const serious = result.violations.filter((item) => item.impact === 'serious')
      seriousSummary.push({ page: route.path, count: serious.length })

      expect(
        critical,
        `Critical accessibility violations on ${route.path}: ${critical.map((item) => item.id).join(', ')}`,
      ).toEqual([])
    }

    await recordToReport('Accessibility serious-level summary', {
      content: seriousSummary.map((item) => `${item.page}: ${item.count}`).join('\n'),
    })
  })
})
