import { expect, test } from './fixture'
import { bootstrapAuthedSession, openSettings } from './helpers/ui-flows'

test('pat flow: create personal access token from settings', async ({ page, recordToReport }, testInfo) => {
  const { apiClient, session } = await bootstrapAuthedSession(page, testInfo)

  const patsBefore = await apiClient.listPersonalAccessTokens(session.authSession.accessToken)

  await openSettings(page)
  await page.getByRole('button', { name: 'Create PAT' }).click()
  await expect(page.getByText('No PAT generated yet.')).toHaveCount(0)

  const patsAfter = await apiClient.listPersonalAccessTokens(session.authSession.accessToken)
  expect(patsAfter.items.length).toBeGreaterThan(patsBefore.items.length)

  await recordToReport('PAT verification passed', {
    content: `PAT count: ${patsBefore.items.length} -> ${patsAfter.items.length}`,
  })
})
