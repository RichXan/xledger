import { expect, type Page, type TestInfo } from '@playwright/test'
import { XledgerApiClient } from './api-client'
import { injectSession, prepareSession, type PreparedSession } from './session'

const API_BASE_URL = process.env.E2E_API_BASE_URL ?? 'http://127.0.0.1:8080/api'

export interface E2ESessionContext {
  apiClient: XledgerApiClient
  session: PreparedSession
}

export function uniqueName(prefix: string): string {
  const now = Date.now()
  const salt = Math.floor(Math.random() * 10_000)
  return `${prefix} ${now}-${salt}`
}

export async function bootstrapAuthedSession(page: Page, testInfo: TestInfo): Promise<E2ESessionContext> {
  const apiClient = new XledgerApiClient(API_BASE_URL)
  const runTag = `${testInfo.title}-${testInfo.retry}`
  const session = await prepareSession(apiClient, runTag)

  await injectSession(page, session.authSession)
  await page.goto('/dashboard?lang=en', { waitUntil: 'networkidle' })
  await expect(page.getByRole('heading', { name: 'Financial Overview' })).toBeVisible()

  return { apiClient, session }
}

export async function openAccounts(page: Page) {
  await page.getByRole('link', { name: 'Accounts' }).click()
  await expect(page.getByRole('heading', { name: 'Accounts' })).toBeVisible()
}

export async function openTransactions(page: Page) {
  await page.getByRole('link', { name: 'Transactions' }).click()
  await expect(page.getByRole('heading', { name: 'Transactions' })).toBeVisible()
}

export async function openDashboard(page: Page) {
  await page.getByRole('link', { name: 'Dashboard' }).click()
  await expect(page.getByRole('heading', { name: 'Financial Overview' })).toBeVisible()
}

export async function openAnalytics(page: Page) {
  await page.getByRole('link', { name: 'Analytics' }).click()
  await expect(page.getByRole('heading', { name: 'Analytics' })).toBeVisible()
}

export async function openSettings(page: Page) {
  await page.getByRole('link', { name: 'Settings' }).click()
  await expect(page.getByRole('heading', { name: 'Settings' })).toBeVisible()
}

export async function createAccountOnAccountsPage(page: Page, input: { name: string; initialBalance: string }) {
  await page.getByRole('button', { name: 'New Account' }).click()
  await page.getByLabel('Account Name').fill(input.name)
  await page.getByLabel('Initial Balance').fill(input.initialBalance)
  await page.getByRole('button', { name: 'Create Account' }).click()
  await expect(page.getByText(input.name)).toBeVisible()
}

export async function createLedgerOnAccountsPage(page: Page, ledgerName: string) {
  await page.getByRole('button', { name: 'New Ledger' }).click()
  await page.getByLabel('Ledger Name').fill(ledgerName)
  await page.getByRole('button', { name: 'Create Ledger' }).click()
  await expect(page.getByText(ledgerName)).toBeVisible()
}

export async function createTransactionOnTransactionsPage(
  page: Page,
  input: { amount: string; memo: string; type: 'income' | 'expense' },
) {
  await page.getByRole('button', { name: '+ Add Transaction' }).click()
  const form = page.locator('#add-transaction-form')
  await expect(form).toBeVisible()
  await form.getByLabel('Amount').fill(input.amount)
  await form.getByLabel('Memo').fill(input.memo)
  await form.getByLabel('Type').selectOption(input.type)
  const accountSelect = form.getByLabel('Account')
  const accountOptions = await accountSelect.locator('option').allTextContents()
  const firstAccount = accountOptions.find((option) => option !== 'Select account')
  if (firstAccount) {
    await accountSelect.selectOption({ label: firstAccount })
  }

  const createRequest = page.waitForResponse((response) => {
    return response.request().method() === 'POST' && response.url().includes('/api/transactions')
  })

  await page.getByRole('button', { name: 'Save Transaction' }).click()
  const createResponse = await createRequest
  if (!createResponse.ok()) {
    throw new Error(`Create transaction failed: ${createResponse.status()} ${await createResponse.text()}`)
  }

  await expect(form).toBeHidden()
}
