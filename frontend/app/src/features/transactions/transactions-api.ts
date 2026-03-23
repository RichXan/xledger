import { requestEnvelope } from '@/lib/api'

export interface PaginatedResponse<T> {
  items: T[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
}

export interface TransactionRecord {
  id: string
  type: 'income' | 'expense' | 'transfer'
  amount: number
  category_name?: string
  occurred_at: string
}

export interface AccountRecord {
  id: string
  name: string
  type: string
  initial_balance: number
}

export interface LedgerRecord {
  id: string
  name: string
  is_default?: boolean
}

export interface CategoryRecord {
  id: string
  name: string
}

export interface TagRecord {
  id: string
  name: string
}

export interface CreateTransactionInput {
  ledger_id: string
  account_id?: string
  category_id?: string
  tag_ids?: string[]
  type: 'income' | 'expense' | 'transfer'
  amount: number
}

export interface ImportPreviewResponse {
  columns: string[]
  sample_rows: string[][]
  mappingSlots: string[]
  mappingCandidates: string[]
}

export function getTransactions(accessToken: string) {
  const params = new URLSearchParams({ page: '1', page_size: '20' })
  return requestEnvelope<PaginatedResponse<TransactionRecord>>(`/transactions?${params.toString()}`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

export function getAccounts(accessToken: string) {
  return requestEnvelope<PaginatedResponse<AccountRecord>>('/accounts', {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

export function getLedgers(accessToken: string) {
  return requestEnvelope<PaginatedResponse<LedgerRecord>>('/ledgers', {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

export function getCategories(accessToken: string) {
  return requestEnvelope<PaginatedResponse<CategoryRecord>>('/categories', {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

export function getTags(accessToken: string) {
  return requestEnvelope<PaginatedResponse<TagRecord>>('/tags', {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

export function createTransaction(accessToken: string, input: CreateTransactionInput) {
  return requestEnvelope<TransactionRecord>('/transactions', {
    method: 'POST',
    headers: { Authorization: `Bearer ${accessToken}` },
    body: JSON.stringify(input),
  })
}

export function previewImport(accessToken: string, file: File) {
  const formData = new FormData()
  formData.append('file', file)

  return requestEnvelope<ImportPreviewResponse>('/import/csv', {
    method: 'POST',
    headers: { Authorization: `Bearer ${accessToken}` },
    body: formData,
  })
}
