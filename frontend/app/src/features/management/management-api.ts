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

export function getLedgers(accessToken: string) {
  return requestEnvelope<PaginatedResponse<LedgerItem>>('/ledgers', {
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
