import { expect, test } from './fixture'
import {
  bootstrapAuthedSession,
  createAccountOnAccountsPage,
  createLedgerOnAccountsPage,
  createTransactionOnTransactionsPage,
  openAccounts,
  openTransactions,
  uniqueName,
} from './helpers/ui-flows'

test('transactions flow: create expense and income records', async ({ page, recordToReport }, testInfo) => {
  await bootstrapAuthedSession(page, testInfo)

  await openAccounts(page)
  await createAccountOnAccountsPage(page, {
    name: uniqueName('E2E Wallet'),
    initialBalance: '1000',
  })
  await createLedgerOnAccountsPage(page, uniqueName('E2E Transaction Ledger'))

  await openTransactions(page)

  const expenseMemo = uniqueName('E2E expense coffee')
  await createTransactionOnTransactionsPage(page, {
    amount: '120',
    memo: expenseMemo,
    type: 'expense',
  })
  await expect(page.getByText(expenseMemo, { exact: true })).toBeVisible()

  const incomeMemo = uniqueName('E2E income salary')
  await createTransactionOnTransactionsPage(page, {
    amount: '2500',
    memo: incomeMemo,
    type: 'income',
  })
  await expect(page.getByText(incomeMemo, { exact: true })).toBeVisible()

  await recordToReport('Transactions created', {
    content: `Expense=120 (${expenseMemo}), Income=2500 (${incomeMemo})`,
  })
})

test('transactions flow: quick filters, delete, and undo', async ({ page, recordToReport }, testInfo) => {
  await bootstrapAuthedSession(page, testInfo)

  await openAccounts(page)
  await createAccountOnAccountsPage(page, {
    name: uniqueName('E2E Filter Wallet'),
    initialBalance: '1000',
  })
  await createLedgerOnAccountsPage(page, uniqueName('E2E Filter Ledger'))

  await openTransactions(page)

  const expenseMemo = uniqueName('E2E filtered expense')
  await createTransactionOnTransactionsPage(page, {
    amount: '42',
    memo: expenseMemo,
    type: 'expense',
  })
  await expect(page.getByText(expenseMemo, { exact: true })).toBeVisible()

  const incomeMemo = uniqueName('E2E filtered income')
  await createTransactionOnTransactionsPage(page, {
    amount: '4200',
    memo: incomeMemo,
    type: 'income',
  })
  await expect(page.getByText(incomeMemo, { exact: true })).toBeVisible()

  await page.getByRole('button', { name: 'Expense', exact: true }).click()
  await expect(page.getByText(expenseMemo, { exact: true })).toBeVisible()
  await expect(page.getByText(incomeMemo, { exact: true })).toBeHidden()

  await page.getByRole('button', { name: 'Income', exact: true }).click()
  await expect(page.getByText(expenseMemo, { exact: true })).toBeHidden()
  await expect(page.getByText(incomeMemo, { exact: true })).toBeVisible()

  await page.getByRole('button', { name: 'All', exact: true }).click()
  const deleteResponse = page.waitForResponse((response) => {
    return response.request().method() === 'DELETE' && response.url().includes('/api/transactions/')
  })
  await page.getByRole('button', { name: `Delete ${expenseMemo}` }).click()
  const response = await deleteResponse
  if (!response.ok()) {
    throw new Error(`Delete transaction failed: ${response.status()} ${await response.text()}`)
  }
  await expect(page.getByText(expenseMemo, { exact: true })).toBeHidden()

  const undoResponse = page.waitForResponse((response) => {
    return response.request().method() === 'POST' && response.url().includes('/api/transactions')
  })
  await page.getByRole('button', { name: 'Undo' }).click()
  const restoreResponse = await undoResponse
  if (!restoreResponse.ok()) {
    throw new Error(`Undo transaction failed: ${restoreResponse.status()} ${await restoreResponse.text()}`)
  }
  await expect(page.getByText(expenseMemo, { exact: true })).toBeVisible()

  await recordToReport('Transactions quick actions verified', {
    content: `Filtered by type, deleted, and restored ${expenseMemo}`,
  })
})

test('transactions flow: smart views surface items that need review', async ({ page, recordToReport }, testInfo) => {
  const { apiClient, session } = await bootstrapAuthedSession(page, testInfo)
  const accessToken = session.authSession.accessToken
  const unique = uniqueName('E2E smart')
  const occurredAt = new Date().toISOString()

  const ledger = await apiClient.createLedger(accessToken, {
    name: `${unique} Ledger`,
  })
  const account = await apiClient.createAccount(accessToken, {
    name: `${unique} Wallet`,
    type: 'cash',
    initial_balance: 1000,
  })
  const category = await apiClient.createCategory(accessToken, {
    name: `${unique} Category`,
  })

  await apiClient.createTransaction(accessToken, {
    ledger_id: ledger.id,
    account_id: account.id,
    type: 'expense',
    amount: 88,
    memo: `${unique} uncategorized`,
    occurred_at: occurredAt,
  })
  await apiClient.createTransaction(accessToken, {
    ledger_id: ledger.id,
    account_id: account.id,
    category_id: category.id,
    type: 'expense',
    amount: 1250,
    memo: `${unique} large expense`,
    occurred_at: occurredAt,
  })
  await apiClient.createTransaction(accessToken, {
    ledger_id: ledger.id,
    account_id: account.id,
    category_id: category.id,
    type: 'expense',
    amount: 25,
    memo: `${unique} duplicate lunch`,
    occurred_at: occurredAt,
  })
  await apiClient.createTransaction(accessToken, {
    ledger_id: ledger.id,
    account_id: account.id,
    category_id: category.id,
    type: 'expense',
    amount: 25,
    memo: `${unique} duplicate lunch`,
    occurred_at: occurredAt,
  })
  await apiClient.createTransaction(accessToken, {
    ledger_id: ledger.id,
    account_id: account.id,
    category_id: category.id,
    type: 'income',
    amount: 5000,
    memo: `${unique} normal income`,
    occurred_at: occurredAt,
  })

  await openTransactions(page)
  await expect(page.getByRole('heading', { name: 'Smart Views' })).toBeVisible()
  await expect(page.locator('span', { hasText: /\d+ uncategorized/ }).first()).toBeVisible()
  await expect(page.locator('span', { hasText: /\d+ possible duplicate/ }).first()).toBeVisible()
  await expect(page.locator('span', { hasText: /\d+ large transaction/ }).first()).toBeVisible()

  await page.getByPlaceholder('Transaction ID, category, or note...').fill(unique)
  await page.getByRole('button', { name: 'Needs Review', exact: true }).click()
  await expect(page.getByText(`${unique} uncategorized`, { exact: true })).toBeVisible()
  await expect(page.getByText(`${unique} large expense`, { exact: true })).toBeVisible()
  await expect(page.getByText(`${unique} duplicate lunch`, { exact: true }).first()).toBeVisible()
  await expect(page.getByText('Missing category', { exact: true })).toBeVisible()
  await expect(page.getByText('Large expense', { exact: true })).toBeVisible()
  await expect(page.getByText('Possible duplicate', { exact: true }).first()).toBeVisible()
  await expect(page.getByText(`${unique} normal income`, { exact: true })).toBeHidden()

  await page.getByRole('button', { name: 'Focus uncategorized review' }).click()
  await expect(page.getByText(`${unique} uncategorized`, { exact: true })).toBeVisible()
  await expect(page.getByText(`${unique} large expense`, { exact: true })).toBeHidden()
  await page.getByRole('button', { name: `Mark ${unique} uncategorized as reviewed` }).click()
  await expect(page.getByText(`${unique} uncategorized`, { exact: true })).toBeHidden()

  await recordToReport('Transactions smart views verified', {
    content: `Review queue surfaced uncategorized, duplicate, and large expense transactions for ${unique}, then dismissed one item`,
  })
})

test('transactions flow: export current transaction range as CSV', async ({ page, recordToReport }, testInfo) => {
  await bootstrapAuthedSession(page, testInfo)

  await openAccounts(page)
  await createAccountOnAccountsPage(page, {
    name: uniqueName('E2E Export Wallet'),
    initialBalance: '1000',
  })
  await createLedgerOnAccountsPage(page, uniqueName('E2E Export Ledger'))

  await openTransactions(page)

  const memo = uniqueName('E2E export lunch')
  await createTransactionOnTransactionsPage(page, {
    amount: '36.5',
    memo,
    type: 'expense',
  })
  await expect(page.getByText(memo, { exact: true })).toBeVisible()

  const exportResponse = page.waitForResponse((response) => {
    return response.request().method() === 'GET' && response.url().includes('/api/export?')
  })
  await page.getByRole('button', { name: 'Export', exact: true }).click()

  const response = await exportResponse
  if (!response.ok()) {
    throw new Error(`Export transactions failed: ${response.status()} ${await response.text()}`)
  }
  expect(response.headers()['content-type'] ?? '').toContain('text/csv')
  expect(await response.text()).toContain(memo)

  await recordToReport('Transactions export verified', {
    content: `Exported current transaction range containing ${memo}`,
  })
})
