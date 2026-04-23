import { expect, test } from './fixture'
import {
  bootstrapAuthedSession,
  createAccountOnAccountsPage,
  createLedgerOnAccountsPage,
  openAccounts,
  uniqueName,
} from './helpers/ui-flows'

test('accounts flow: create account and ledger', async ({ page, recordToReport }, testInfo) => {
  const { apiClient, session } = await bootstrapAuthedSession(page, testInfo)

  await openAccounts(page)

  const accountName = uniqueName('E2E Cash')
  await createAccountOnAccountsPage(page, { name: accountName, initialBalance: '1000' })

  const ledgerName = uniqueName('E2E Ledger')
  await createLedgerOnAccountsPage(page, ledgerName)

  const [accounts, ledgers] = await Promise.all([
    apiClient.listAccounts(session.authSession.accessToken),
    apiClient.listLedgers(session.authSession.accessToken),
  ])

  expect(accounts.items.some((account) => account.name === accountName)).toBeTruthy()
  expect(ledgers.items.some((ledger) => ledger.name === ledgerName)).toBeTruthy()

  await recordToReport('Account and ledger created', {
    content: `Account=${accountName}, Ledger=${ledgerName}`,
  })
})
