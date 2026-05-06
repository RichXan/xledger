import type { TransactionRecord } from './transactions-api'

export type ReviewReasonKey = 'uncategorized' | 'duplicate' | 'large'

function getTransactionDayKey(value: string) {
  const date = new Date(value)
  return `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}`
}

function getDuplicateKey(tx: TransactionRecord) {
  const memo = tx.memo?.trim().toLowerCase() ?? ''
  const category = tx.category_name?.trim().toLowerCase() ?? ''
  return [tx.type, Math.abs(tx.amount).toFixed(2), getTransactionDayKey(tx.occurred_at), memo || category].join('|')
}

export function buildDuplicateIds(transactions: TransactionRecord[]) {
  const groups = new Map<string, TransactionRecord[]>()
  transactions.forEach((tx) => {
    const key = getDuplicateKey(tx)
    const group = groups.get(key) ?? []
    groups.set(key, [...group, tx])
  })

  const ids = new Set<string>()
  groups.forEach((group) => {
    if (group.length > 1) {
      group.forEach((tx) => ids.add(tx.id))
    }
  })
  return ids
}

export function countDuplicateGroups(transactions: TransactionRecord[]) {
  const counts = new Map<string, number>()
  transactions.forEach((tx) => {
    const key = getDuplicateKey(tx)
    counts.set(key, (counts.get(key) ?? 0) + 1)
  })
  return Array.from(counts.values()).filter((count) => count > 1).length
}

export function getReviewReasonKeys(tx: TransactionRecord, duplicateIds: ReadonlySet<string>): ReviewReasonKey[] {
  const reasons: ReviewReasonKey[] = []
  if (!tx.category_name?.trim()) {
    reasons.push('uncategorized')
  }
  if (duplicateIds.has(tx.id)) {
    reasons.push('duplicate')
  }
  if (tx.type === 'expense' && Math.abs(tx.amount) >= 1000) {
    reasons.push('large')
  }
  return reasons
}

export function transactionNeedsReview(tx: TransactionRecord, duplicateIds: ReadonlySet<string>) {
  return getReviewReasonKeys(tx, duplicateIds).length > 0
}

export function buildReviewSummary(transactions: TransactionRecord[]) {
  const duplicateIds = buildDuplicateIds(transactions)
  return {
    review: transactions.filter((tx) => transactionNeedsReview(tx, duplicateIds)).length,
    uncategorized: transactions.filter((tx) => !tx.category_name?.trim()).length,
    duplicates: countDuplicateGroups(transactions),
    large: transactions.filter((tx) => tx.type === 'expense' && Math.abs(tx.amount) >= 1000).length,
  }
}
