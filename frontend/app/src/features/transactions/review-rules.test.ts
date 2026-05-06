import { describe, expect, it } from 'vitest'
import type { TransactionRecord } from './transactions-api'
import {
  buildDuplicateIds,
  buildReviewSummary,
  getReviewReasonKeys,
  transactionNeedsReview,
} from './review-rules'

function tx(input: Partial<TransactionRecord> & Pick<TransactionRecord, 'id' | 'type' | 'amount' | 'occurred_at'>): TransactionRecord {
  return {
    ledger_id: 'ledger-1',
    ...input,
  }
}

describe('transaction review rules', () => {
  it('classifies uncategorized, duplicate, and large expense review reasons consistently', () => {
    const transactions = [
      tx({ id: 'uncategorized', type: 'expense', amount: 12, occurred_at: '2026-05-01T08:00:00Z', memo: 'cash' }),
      tx({ id: 'large', type: 'expense', amount: 1200, occurred_at: '2026-05-01T09:00:00Z', category_name: 'Travel' }),
      tx({ id: 'dup-a', type: 'expense', amount: 25, occurred_at: '2026-05-02T10:00:00Z', category_name: 'Food', memo: 'Lunch' }),
      tx({ id: 'dup-b', type: 'expense', amount: 25, occurred_at: '2026-05-02T12:00:00Z', category_name: 'Cafe', memo: 'Lunch' }),
      tx({ id: 'income', type: 'income', amount: 5000, occurred_at: '2026-05-02T12:00:00Z', category_name: 'Salary' }),
    ]

    const duplicateIds = buildDuplicateIds(transactions)

    expect(getReviewReasonKeys(transactions[0], duplicateIds)).toEqual(['uncategorized'])
    expect(getReviewReasonKeys(transactions[1], duplicateIds)).toEqual(['large'])
    expect(getReviewReasonKeys(transactions[2], duplicateIds)).toEqual(['duplicate'])
    expect(transactionNeedsReview(transactions[4], duplicateIds)).toBe(false)
    expect(buildReviewSummary(transactions)).toEqual({
      review: 4,
      uncategorized: 1,
      duplicates: 1,
      large: 1,
    })
  })
})
