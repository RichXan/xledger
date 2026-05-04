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
