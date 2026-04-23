import { expect, test } from './fixture'
import { bootstrapAuthedSession, openTransactions, uniqueName } from './helpers/ui-flows'

test('transfer flow: create, edit, and delete transfer pair', async ({ page, recordToReport }, testInfo) => {
  const { apiClient, session } = await bootstrapAuthedSession(page, testInfo)
  const accessToken = session.authSession.accessToken

  const [fromAccount, toAccount] = await Promise.all([
    apiClient.createAccount(accessToken, {
      name: uniqueName('E2E Transfer From'),
      type: 'cash',
      initial_balance: 1000,
    }),
    apiClient.createAccount(accessToken, {
      name: uniqueName('E2E Transfer To'),
      type: 'bank',
      initial_balance: 200,
    }),
  ])

  const ledgers = await apiClient.listLedgers(accessToken)
  const defaultLedger = ledgers.items.find((ledger) => ledger.is_default) ?? ledgers.items[0]
  expect(defaultLedger).toBeTruthy()

  const createdTransfer = await apiClient.createTransaction(accessToken, {
    ledger_id: defaultLedger.id,
    type: 'transfer',
    from_account_id: fromAccount.id,
    to_account_id: toAccount.id,
    amount: 88,
  })
  expect(createdTransfer.type).toBe('transfer')
  expect(createdTransfer.transfer_pair_id).toBeTruthy()

  await openTransactions(page)
  await expect(page.getByText('Transfer', { exact: true }).first()).toBeVisible()

  const beforeEditList = await apiClient.listTransactions(accessToken, {
    ledger_id: defaultLedger.id,
    page_size: 200,
  })
  const beforeEditPair = (beforeEditList.items ?? []).filter(
    (item) => item.transfer_pair_id === createdTransfer.transfer_pair_id,
  )
  expect(beforeEditPair.length).toBeGreaterThanOrEqual(2)
  expect(beforeEditPair.every((item) => item.amount === 88)).toBeTruthy()

  const updatedTransfer = await apiClient.updateTransaction(accessToken, createdTransfer.id, {
    amount: 99,
    version: createdTransfer.version,
  })
  expect(updatedTransfer.amount).toBe(99)

  const afterEditList = await apiClient.listTransactions(accessToken, {
    ledger_id: defaultLedger.id,
    page_size: 200,
  })
  const afterEditPair = (afterEditList.items ?? []).filter(
    (item) => item.transfer_pair_id === createdTransfer.transfer_pair_id,
  )
  expect(afterEditPair.length).toBeGreaterThanOrEqual(2)
  expect(afterEditPair.every((item) => item.amount === 99)).toBeTruthy()

  await apiClient.deleteTransaction(accessToken, createdTransfer.id, { version: updatedTransfer.version })

  const afterDeleteList = await apiClient.listTransactions(accessToken, {
    ledger_id: defaultLedger.id,
    page_size: 200,
  })
  const afterDeletePair = (afterDeleteList.items ?? []).filter(
    (item) => item.transfer_pair_id === createdTransfer.transfer_pair_id,
  )
  expect(afterDeletePair.length).toBe(0)

  await page.reload({ waitUntil: 'networkidle' })
  await expect(page.getByText('No matching transactions.')).toBeVisible()

  await recordToReport('Transfer lifecycle verified', {
    content: `pair=${createdTransfer.transfer_pair_id}, amount: 88 -> 99 -> deleted`,
  })
})
