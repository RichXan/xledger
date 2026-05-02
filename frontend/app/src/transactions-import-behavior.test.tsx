import type { AccountRecord, LedgerRecord, TransactionRecord } from '@/features/transactions/transactions-api'
import {
  buildImportIdempotencyKey,
  getTransactionAccountLabel,
  getTransactionLedgerLabel,
} from './pages/transactions-page'

describe('transactions import behavior', () => {
  it('uses a stable idempotency key for the same import file contents', async () => {
    const firstFile = new File(['date,amount,description\n2026-03-01,25,Lunch'], 'first.csv', {
      type: 'text/csv',
    })
    const secondFile = new File(['date,amount,description\n2026-03-01,25,Lunch'], 'second.csv', {
      type: 'text/csv',
    })

    await expect(buildImportIdempotencyKey(firstFile)).resolves.toBe(
      await buildImportIdempotencyKey(secondFile),
    )
  })

  it('resolves account and ledger labels from each transaction record', () => {
    const accounts: AccountRecord[] = [
      { id: 'acc-1', name: 'Cash Wallet', type: 'cash', initial_balance: 0 },
      { id: 'acc-2', name: 'Checking', type: 'bank', initial_balance: 0 },
    ]
    const ledgers: LedgerRecord[] = [{ id: 'ledger-1', name: 'Default Ledger', is_default: true }]
    const accountNameById = new Map(accounts.map((account) => [account.id, account.name]))
    const ledgerNameById = new Map(ledgers.map((ledger) => [ledger.id, ledger.name]))
    const tx: TransactionRecord = {
      id: 'txn-1',
      ledger_id: 'ledger-1',
      account_id: 'acc-2',
      type: 'expense',
      amount: 25,
      occurred_at: '2026-03-01T08:30:00Z',
    }

    expect(getTransactionAccountLabel(tx, accountNameById, 'No account')).toBe('Checking')
    expect(getTransactionLedgerLabel(tx, ledgerNameById, 'No ledger')).toBe('Default Ledger')
  })
})
