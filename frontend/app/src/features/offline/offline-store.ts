// frontend/app/src/features/offline/offline-store.ts
import Dexie, { type Table } from 'dexie'

export interface CachedTransaction {
  id: string
  ledger_id: string
  account_id: string | null
  category_id: string | null
  type: 'income' | 'expense' | 'transfer'
  amount: number
  memo: string | null
  occurred_at: string
  created_at: string
}

export interface CachedAccount {
  id: string
  name: string
  type: string
  balance: number
}

export interface CachedCategory {
  id: string
  name: string
  parent_id: string | null
}

export interface OfflineOverview {
  id: string  // 总为 'latest'
  total_assets: number
  income: number
  expense: number
  cached_at: string
}

class XLedgerOfflineDB extends Dexie {
  transactions!: Table<CachedTransaction>
  accounts!: Table<CachedAccount>
  categories!: Table<CachedCategory>
  overviews!: Table<OfflineOverview>

  constructor() {
    super('xledger-offline')
    this.version(1).stores({
      transactions: 'id, ledger_id, account_id, category_id, occurred_at',
      accounts: 'id, type',
      categories: 'id, parent_id',
      overviews: 'id',
    })
  }
}

export const offlineDb = new XLedgerOfflineDB()

export async function cacheTransactions(txns: CachedTransaction[]) {
  await offlineDb.transactions.bulkPut(txns)
}

export async function cacheAccounts(accts: CachedAccount[]) {
  await offlineDb.accounts.bulkPut(accts)
}

export async function cacheCategories(cats: CachedCategory[]) {
  await offlineDb.categories.bulkPut(cats)
}

export async function cacheOverview(overview: Omit<OfflineOverview, 'id'>) {
  await offlineDb.overviews.put({ id: 'latest', ...overview })
}

export async function getCachedTransactions(limit = 100): Promise<CachedTransaction[]> {
  return offlineDb.transactions.orderBy('occurred_at').reverse().limit(limit).toArray()
}

export async function getCachedOverview(): Promise<OfflineOverview | undefined> {
  return offlineDb.overviews.get('latest')
}

export async function getCachedAccounts(): Promise<CachedAccount[]> {
  return offlineDb.accounts.toArray()
}

export async function clearOfflineCache() {
  await offlineDb.transactions.clear()
  await offlineDb.accounts.clear()
  await offlineDb.categories.clear()
  await offlineDb.overviews.clear()
}