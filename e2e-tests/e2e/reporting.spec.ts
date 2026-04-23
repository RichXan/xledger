import { expect, test } from './fixture'
import {
  bootstrapAuthedSession,
  createAccountOnAccountsPage,
  createTransactionOnTransactionsPage,
  openAccounts,
  openAnalytics,
  openDashboard,
  openTransactions,
} from './helpers/ui-flows'

test('reporting flow: overview and analytics reflect transactions', async ({ page, recordToReport }, testInfo) => {
  const { apiClient, session } = await bootstrapAuthedSession(page, testInfo)

  await openAccounts(page)
  await createAccountOnAccountsPage(page, {
    name: 'E2E Reporting Account',
    initialBalance: '1000',
  })

  await openTransactions(page)
  await createTransactionOnTransactionsPage(page, {
    amount: '120',
    memo: 'E2E reporting expense',
    type: 'expense',
  })
  await createTransactionOnTransactionsPage(page, {
    amount: '2500',
    memo: 'E2E reporting income',
    type: 'income',
  })

  const from = new Date()
  from.setDate(from.getDate() - 7)
  const to = new Date()
  to.setDate(to.getDate() + 1)

  const overview = await apiClient.getOverview(session.authSession.accessToken, {
    from: from.toISOString(),
    to: to.toISOString(),
  })

  expect(overview.total_assets).toBeCloseTo(1000, 2)
  expect(overview.income).toBeCloseTo(2500, 2)
  expect(overview.expense).toBeCloseTo(120, 2)
  expect(overview.net).toBeCloseTo(2380, 2)

  await openDashboard(page)
  await expect(page.getByText('Total Assets')).toBeVisible()
  await expect(page.getByText('Income: ¥', { exact: false })).toBeVisible()
  await expect(page.getByText('Expense: ¥', { exact: false })).toBeVisible()

  await openAnalytics(page)
  await expect(page.getByText('Expense Categories')).toBeVisible()
  await expect(page.getByText('Revenue vs Burn Rate')).toBeVisible()

  await recordToReport('Reporting verification passed', {
    content: `assets=${overview.total_assets}, income=${overview.income}, expense=${overview.expense}, net=${overview.net}`,
  })
})
