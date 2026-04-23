import { expect, test } from './fixture'
import {
  bootstrapAuthedSession,
  createAccountOnAccountsPage,
  createLedgerOnAccountsPage,
  openAccounts,
  uniqueName,
} from './helpers/ui-flows'

test('account management flow: edit account and rename/delete ledger', async ({ page, recordToReport }, testInfo) => {
  await bootstrapAuthedSession(page, testInfo)
  await openAccounts(page)

  const accountName = uniqueName('E2E Editable Account')
  await createAccountOnAccountsPage(page, {
    name: accountName,
    initialBalance: '1500',
  })

  const accountCard = page
    .locator('div')
    .filter({ hasText: accountName })
    .filter({ has: page.getByRole('button', { name: 'Edit account' }) })
    .first()
  await accountCard.getByRole('button', { name: 'Edit account' }).click()

  const updatedAccountName = uniqueName('E2E Edited Account')
  await page.getByLabel('Account Name').fill(updatedAccountName)
  await page.getByLabel('Account Type').selectOption('bank')
  await page.getByRole('button', { name: 'Save Changes' }).click()
  await expect(page.getByText(updatedAccountName)).toBeVisible()
  await expect(page.getByText('bank')).toBeVisible()

  const ledgerName = uniqueName('E2E Editable Ledger')
  await createLedgerOnAccountsPage(page, ledgerName)
  const updatedLedgerName = uniqueName('E2E Renamed Ledger')

  const targetLedgerRow = page
    .getByText(ledgerName, { exact: true })
    .locator('xpath=ancestor::div[contains(@class,"justify-between")][1]')
  await targetLedgerRow.getByRole('button', { name: 'Edit' }).click()
  await page.getByLabel('Ledger Name').fill(updatedLedgerName)
  await page.getByRole('button', { name: 'Save Changes' }).click()
  await expect(page.getByText(updatedLedgerName)).toBeVisible()

  const renamedLedgerRow = page
    .getByText(updatedLedgerName, { exact: true })
    .locator('xpath=ancestor::div[contains(@class,"justify-between")][1]')
  await renamedLedgerRow.getByRole('button', { name: 'Delete' }).click()
  await expect(page.getByText(updatedLedgerName)).toHaveCount(0)

  await recordToReport('Account and ledger management verified', {
    content: `Account updated to ${updatedAccountName}, ledger deleted ${updatedLedgerName}`,
  })
})
