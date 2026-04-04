import { requestEnvelope } from '@/lib/api'
import type { PaginatedResponse } from '@/features/transactions/transactions-api'

export interface AccountItem {
  id: string
  name: string
  type: string
  initial_balance: number
}

export interface LedgerItem {
  id: string
  name: string
  is_default?: boolean
}

export interface CategoryItem {
  id: string
  name: string
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
  pat_token: string
  api_endpoint: string
  shortcut_url?: string
  expires_at?: string
}

export function generateShortcut(accessToken: string, name: string = '快捷记账', expiresIn?: number) {
  const body: { name: string; expires_in?: number } = { name }
  if (expiresIn) {
    body.expires_in = expiresIn
  }
  return requestEnvelope<GenerateShortcutResponse>('/shortcuts/generate', {
    method: 'POST',
    headers: { Authorization: `Bearer ${accessToken}` },
    body: JSON.stringify(body),
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
