import { expect, test } from './fixture'
import {
  bootstrapAuthedSession,
  createAccountOnAccountsPage,
  createLedgerOnAccountsPage,
  openAccounts,
  openTransactions,
  uniqueName,
} from './helpers/ui-flows'

test('transactions flow: create transaction with second-level timestamp', async ({ page, recordToReport }, testInfo) => {
  const { apiClient, session } = await bootstrapAuthedSession(page, testInfo)
  const memo = uniqueName('E2E precise timestamp')
  const localDateTime = '2026-03-03T12:34:56'

  await openAccounts(page)
  await createLedgerOnAccountsPage(page, uniqueName('E2E Timestamp Ledger'))
  await createAccountOnAccountsPage(page, {
    name: uniqueName('E2E Timestamp Wallet'),
    initialBalance: '1000',
  })

  await openTransactions(page)
  await page.getByRole('button', { name: '+ Add Transaction' }).click()
  const form = page.locator('#add-transaction-form')
  await expect(form).toBeVisible()
  await form.getByLabel('Amount').fill('88.5')
  await form.getByLabel('Date & Time').fill(localDateTime)
  await form.getByLabel('Memo').fill(memo)

  const accountSelect = form.getByLabel('Account')
  const accountOptions = await accountSelect.locator('option').allTextContents()
  const firstAccount = accountOptions.find((option) => option !== 'Select account')
  expect(firstAccount).toBeTruthy()
  await accountSelect.selectOption({ label: firstAccount })

  const createResponsePromise = page.waitForResponse((response) => {
    return response.request().method() === 'POST' && response.url().includes('/api/transactions')
  })
  await page.getByRole('button', { name: 'Save Transaction' }).click()
  const createResponse = await createResponsePromise
  expect(createResponse.ok()).toBeTruthy()

  const transactions = await apiClient.listTransactions(session.authSession.accessToken, {
    page: 1,
    page_size: 50,
  })
  const created = transactions.items.find((transaction) => transaction.memo === memo)
  expect(created).toBeTruthy()
  expect(new Date(created!.occurred_at).getSeconds()).toBe(56)

  await recordToReport('Transaction second-level timestamp verification passed', {
    content: `${memo}: ${created?.occurred_at}`,
  })
})
