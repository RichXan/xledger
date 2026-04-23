import { expect, test } from './fixture'
import {
  bootstrapAuthedSession,
  createAccountOnAccountsPage,
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
