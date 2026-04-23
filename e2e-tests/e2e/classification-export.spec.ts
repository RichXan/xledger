import { expect, test } from './fixture'
import { bootstrapAuthedSession, openAccounts, openAnalytics, uniqueName } from './helpers/ui-flows'

test('classification flow: category/tag usage and CSV export', async ({ page, recordToReport }, testInfo) => {
  const { apiClient, session } = await bootstrapAuthedSession(page, testInfo)
  const accessToken = session.authSession.accessToken

  const categoryName = uniqueName('E2E Category')
  const tagName = uniqueName('E2E Tag')

  const [category, tag, ledgers] = await Promise.all([
    apiClient.createCategory(accessToken, { name: categoryName }),
    apiClient.createTag(accessToken, { name: tagName }),
    apiClient.listLedgers(accessToken),
  ])
  const defaultLedger = ledgers.items.find((ledger) => ledger.is_default) ?? ledgers.items[0]
  expect(defaultLedger).toBeTruthy()

  const memo = uniqueName('E2E category expense memo')
  const occurredAt = new Date().toISOString()
  const expense = await apiClient.createTransaction(accessToken, {
    ledger_id: defaultLedger.id,
    type: 'expense',
    amount: 333,
    category_id: category.id,
    tag_ids: [tag.id],
    memo,
    occurred_at: occurredAt,
  })
  expect(expense.category_id).toBe(category.id)

  const from = new Date()
  from.setDate(from.getDate() - 2)
  const to = new Date()
  to.setDate(to.getDate() + 1)
  const range = { from: from.toISOString(), to: to.toISOString() }

  const allTransactions = await apiClient.listTransactions(accessToken, {
    page_size: 200,
    date_from: range.from,
    date_to: range.to,
  })
  const categoryStats = await apiClient.getCategoryStats(accessToken, range)
  const exportedCsv = await apiClient.exportCsv(accessToken, range)

  expect(allTransactions.items.some((item) => item.id === expense.id)).toBeTruthy()
  const stat = categoryStats.items.find((item) => item.category_name === categoryName)
  expect(stat).toBeTruthy()
  expect(stat?.amount).toBeCloseTo(333, 2)
  expect(exportedCsv).toContain('occurred_at,amount,type,category_name')
  expect(exportedCsv).toContain(',333,expense,')
  expect(exportedCsv).toContain(categoryName)

  const deleteCategoryResult = await apiClient.deleteCategory(accessToken, category.id)
  expect(deleteCategoryResult.deleted).toBeTruthy()
  expect(deleteCategoryResult.archived).toBeTruthy()

  await openAccounts(page)
  await expect(page.getByText(categoryName)).toBeVisible()
  await expect(page.getByText(tagName)).toBeVisible()

  await openAnalytics(page)
  await expect(page.getByText('Expense Categories')).toBeVisible()
  await expect(page.getByRole('button', { name: new RegExp(categoryName) }).first()).toBeVisible()

  await recordToReport('Classification and export verified', {
    content: `category=${categoryName}, tag=${tagName}, memo=${memo}`,
  })
})
