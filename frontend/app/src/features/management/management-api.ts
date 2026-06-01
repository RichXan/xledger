import { requestEnvelope } from '@/lib/api'
import type { PaginatedResponse } from '@/features/transactions/transactions-api'

export interface AccountItem {
  id: string
  name: string
  type: string
  initial_balance: number
  current_balance?: number
}

export interface LedgerItem {
  id: string
  name: string
  is_default?: boolean
}

export interface CategoryItem {
  id: string
  name: string
  parent_id?: string
  archived_at?: string
}

export interface TagItem {
  id: string
  name: string
}

export interface PATItem {
  id: string
  name: string
  expires_at?: string
}

export interface CreatePATResponse {
  token: string
  id: string
  expires_at?: string
}

export interface RevokePATResponse {
  revoked: boolean
}

export function getAccounts(accessToken: string) {
  return requestEnvelope<PaginatedResponse<AccountItem>>('/accounts', {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

export function createAccount(accessToken: string, input: { name: string; type: string; initial_balance: number }) {
  return requestEnvelope<AccountItem>('/accounts', {
    method: 'POST',
    headers: { Authorization: `Bearer ${accessToken}` },
    body: JSON.stringify(input),
  })
}

export function updateAccount(accessToken: string, id: string, input: { name?: string; type?: string }) {
  return requestEnvelope<AccountItem>(`/accounts/${id}`, {
    method: 'PATCH',
    headers: { Authorization: `Bearer ${accessToken}` },
    body: JSON.stringify(input),
  })
}

export function getLedgers(accessToken: string) {
  return requestEnvelope<PaginatedResponse<LedgerItem>>('/ledgers', {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

export function createLedger(accessToken: string, input: { name: string; is_default?: boolean }) {
  return requestEnvelope<LedgerItem>('/ledgers', {
    method: 'POST',
    headers: { Authorization: `Bearer ${accessToken}` },
    body: JSON.stringify(input),
  })
}

export function updateLedger(accessToken: string, id: string, input: { name: string }) {
  return requestEnvelope<LedgerItem>(`/ledgers/${id}`, {
    method: 'PATCH',
    headers: { Authorization: `Bearer ${accessToken}` },
    body: JSON.stringify(input),
  })
}

export function deleteLedger(accessToken: string, id: string) {
  return requestEnvelope<{ deleted: boolean }>(`/ledgers/${id}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

export function getCategories(accessToken: string) {
  return requestEnvelope<PaginatedResponse<CategoryItem>>('/categories', {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

export function createCategory(accessToken: string, input: { name: string; parent_id?: string }) {
  return requestEnvelope<CategoryItem>('/categories', {
    method: 'POST',
    headers: { Authorization: `Bearer ${accessToken}` },
    body: JSON.stringify(input),
  })
}

export function updateCategory(accessToken: string, id: string, input: { name?: string; archived?: boolean }) {
  return requestEnvelope<CategoryItem>(`/categories/${id}`, {
    method: 'PATCH',
    headers: { Authorization: `Bearer ${accessToken}` },
    body: JSON.stringify(input),
  })
}

export function deleteCategory(accessToken: string, id: string) {
  return requestEnvelope<{ deleted: boolean; archived?: boolean; category?: CategoryItem }>(`/categories/${id}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

export function getTags(accessToken: string) {
  return requestEnvelope<PaginatedResponse<TagItem>>('/tags', {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

export function getPATs(accessToken: string) {
  return requestEnvelope<PaginatedResponse<PATItem>>('/personal-access-tokens', {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

export function createPAT(accessToken: string) {
  return requestEnvelope<CreatePATResponse>('/personal-access-tokens', {
    method: 'POST',
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

export function revokePAT(accessToken: string, id: string) {
  return requestEnvelope<RevokePATResponse>(`/personal-access-tokens/${id}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

export async function exportCsv(accessToken: string) {
  const params = new URLSearchParams({ format: 'csv' })
  const response = await fetch(`/api/export?${params.toString()}`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  })

  if (!response.ok) {
    throw new Error('Unable to export CSV')
  }

  return response.text()
}

export interface GenerateShortcutResponse {
  shortcut_id?: string
  pat_token: string
  api_endpoint: string
  shortcut_url?: string
  install_url?: string
  qr_url?: string
  default_ledger_id?: string
  default_account_id?: string
  expires_at?: string
}

export function generateShortcut(accessToken: string, name: string = '快捷记账', expiresIn?: number) {
  const body: {
    name: string
    expires_in?: number
  } = { name }
  if (expiresIn) {
    body.expires_in = expiresIn
  }
  return requestEnvelope<GenerateShortcutResponse>('/shortcuts/generate', {
    method: 'POST',
    headers: { Authorization: `Bearer ${accessToken}` },
    body: JSON.stringify(body),
  })
}

export interface GenerateOCRShortcutInput {
  name?: string
  expiresIn?: number
  defaultLedgerId: string
  defaultAccountId?: string
}

export function generateOCRShortcut(accessToken: string, input: GenerateOCRShortcutInput) {
  const body: {
    name: string
    expires_in?: number
    default_ledger_id: string
    default_account_id?: string
    mode: string
  } = {
    name: input.name ?? 'Xledger OCR 记账',
    default_ledger_id: input.defaultLedgerId,
    mode: 'ocr_confirm',
  }
  if (input.expiresIn) body.expires_in = input.expiresIn
  if (input.defaultAccountId) body.default_account_id = input.defaultAccountId
  return requestEnvelope<GenerateShortcutResponse>('/shortcuts/generate', {
    method: 'POST',
    headers: { Authorization: `Bearer ${accessToken}` },
    body: JSON.stringify(body),
  })
}

export interface QuickAddSuggestion {
  id: string
  name: string
  reason: string
  confidence?: number
}

export interface QuickAddPreviewResponse {
  shortcut_id?: string
  amount: number
  type: 'expense' | 'income'
  occurred_at: string
  memo: string
  category_suggestion?: QuickAddSuggestion
  ledger_suggestions: QuickAddSuggestion[]
  account_suggestions: QuickAddSuggestion[]
  needs_review: boolean
}

export function previewQuickAdd(
  patToken: string,
  input: {
    shortcutId?: string
    rawText: string
    source?: string
    idempotencyKey: string
    defaultLedgerId?: string
    defaultAccountId?: string
  },
) {
  return requestEnvelope<QuickAddPreviewResponse>('/quick-add/preview', {
    method: 'POST',
    headers: { Authorization: `Bearer ${patToken}` },
    body: JSON.stringify({
      shortcut_id: input.shortcutId,
      raw_text: input.rawText,
      source: input.source,
      idempotency_key: input.idempotencyKey,
      default_ledger_id: input.defaultLedgerId,
      default_account_id: input.defaultAccountId,
    }),
  })
}

export function confirmQuickAdd(
  patToken: string,
  input: {
    shortcutId?: string
    idempotencyKey: string
    ledgerId: string
    accountId?: string
    categoryId?: string
    type: 'expense' | 'income'
    amount: number
    memo?: string
    occurredAt?: string
  },
) {
  return requestEnvelope<{ id: string; success: boolean }>('/quick-add/confirm', {
    method: 'POST',
    headers: { Authorization: `Bearer ${patToken}` },
    body: JSON.stringify({
      shortcut_id: input.shortcutId,
      idempotency_key: input.idempotencyKey,
      ledger_id: input.ledgerId,
      account_id: input.accountId,
      category_id: input.categoryId,
      type: input.type,
      amount: input.amount,
      memo: input.memo,
      occurred_at: input.occurredAt,
    }),
  })
}

// Import CSV
export interface ImportPreviewResponse {
  format: string
  columns: string[]
  sample_rows: string[][]
  mappingSlots: string[]
  mappingCandidates: Record<string, string[]>
  suggested_mapping?: Record<string, string>
}

export interface ImportConfirmResponse {
  success_count: number
  skip_count: number
  fail_count: number
}

export async function importCSVPreview(formData: FormData) {
  const response = await fetch('/api/import/csv', {
    method: 'POST',
    body: formData,
  })
  const json = await response.json()
  if (!response.ok) throw new Error(json.message)
  return json.data as ImportPreviewResponse
}

export async function importCSVConfirm(formData: FormData) {
  const response = await fetch('/api/import/csv/confirm', {
    method: 'POST',
    body: formData,
  })
  const json = await response.json()
  if (!response.ok) throw new Error(json.message)
  return json.data as ImportConfirmResponse
}
