import { expect, test } from './fixture'
import {
  bootstrapAuthedSession,
  createAccountOnAccountsPage,
  createLedgerOnAccountsPage,
  createTransactionOnTransactionsPage,
  openAccounts,
  openAnalytics,
  openDashboard,
  openTransactions,
} from './helpers/ui-flows'

test('reporting flow: overview and analytics reflect transactions', async ({ page, recordToReport }, testInfo) => {
  const { apiClient, session } = await bootstrapAuthedSession(page, testInfo)

  await openAccounts(page)
  await createLedgerOnAccountsPage(page, 'E2E Reporting Ledger')
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
  await expect(page.getByRole('heading', { name: 'Expense Structure' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Spending Cloud' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Cashflow Rhythm' })).toBeVisible()

  await recordToReport('Reporting verification passed', {
    content: `assets=${overview.total_assets}, income=${overview.income}, expense=${overview.expense}, net=${overview.net}`,
  })
})

test('analytics keyword cloud reflects expense memo and category terms', async ({ page, recordToReport }, testInfo) => {
  const { apiClient, session } = await bootstrapAuthedSession(page, testInfo)
  const accessToken = session.authSession.accessToken
  const now = new Date()
  const unique = String(Date.now())
  const from = new Date(now.getFullYear(), now.getMonth(), 1)
  const to = new Date(now.getFullYear(), now.getMonth() + 1, 0, 23, 59, 59, 999)

  const ledger = await apiClient.createLedger(accessToken, {
    name: `Keyword Ledger ${unique}`,
  })

  const account = await apiClient.createAccount(accessToken, {
    name: `Keyword Wallet ${unique}`,
    type: 'cash',
    initial_balance: 500,
  })
  const category = await apiClient.createCategory(accessToken, {
    name: `Coffee Cloud ${unique}`,
  })

  await apiClient.createTransaction(accessToken, {
    ledger_id: ledger.id,
    account_id: account.id,
    category_id: category.id,
    type: 'expense',
    amount: 42,
    memo: `latte focus ${unique}`,
    occurred_at: now.toISOString(),
  })
  await apiClient.createTransaction(accessToken, {
    ledger_id: ledger.id,
    account_id: account.id,
    category_id: category.id,
    type: 'expense',
    amount: 18,
    memo: `latte snack ${unique}`,
    occurred_at: now.toISOString(),
  })
  await apiClient.createTransaction(accessToken, {
    ledger_id: ledger.id,
    account_id: account.id,
    type: 'income',
    amount: 900,
    memo: `salary ${unique}`,
    occurred_at: now.toISOString(),
  })

  const keywords = await apiClient.getKeywordStats(accessToken, {
    from: from.toISOString(),
    to: to.toISOString(),
    limit: 20,
  })
  const byText = new Map(keywords.items.map((item) => [item.text, item]))

  expect(byText.get('latte')?.amount).toBeCloseTo(60, 2)
  expect(byText.get('latte')?.count).toBe(2)
  expect(byText.get('Coffee')?.amount).toBeCloseTo(60, 2)
  expect(byText.get('salary')).toBeUndefined()

  await openAnalytics(page)
  await expect(page.getByRole('heading', { name: 'Spending Cloud' })).toBeVisible()
  await expect(page.getByRole('button', { name: /latte:/ })).toBeVisible()
  await expect(page.getByRole('button', { name: /Coffee:/ })).toBeVisible()
  await expect(page.getByText('Keyword spend')).toBeVisible()

  await recordToReport('Analytics keyword cloud verification passed', {
    content: `latte amount=${byText.get('latte')?.amount}, count=${byText.get('latte')?.count}`,
  })
})
