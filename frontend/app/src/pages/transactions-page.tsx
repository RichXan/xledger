import { useMemo, useState } from 'react'
import { Button } from '@/components/ui/button'
import { DialogShell } from '@/components/ui/dialog-shell'
import { PageSection } from '@/components/ui/page-section'
import { SelectField } from '@/components/ui/select-field'
import { TextField } from '@/components/ui/text-field'
import { useCreateTransaction, useImportPreview, useTransactionFormOptions, useTransactions } from '@/features/transactions/transactions-hooks'
import { formatCurrency, formatShortDate } from '@/lib/format'

export function TransactionsPage() {
  const [view, setView] = useState<'list' | 'calendar'>('list')
  const [showAddDialog, setShowAddDialog] = useState(false)
  const [showImportDialog, setShowImportDialog] = useState(false)
  const [amount, setAmount] = useState('')
  const [categoryId, setCategoryId] = useState('')
  const [accountId, setAccountId] = useState('')
  const [importFile, setImportFile] = useState<File | null>(null)

  const transactionsQuery = useTransactions()
  const options = useTransactionFormOptions()
  const createTransactionMutation = useCreateTransaction()
  const importPreviewMutation = useImportPreview()

  const transactions = transactionsQuery.data?.items ?? []
  const categories = options.categoriesQuery.data?.items ?? []
  const accounts = options.accountsQuery.data?.items ?? []
  const ledgers = options.ledgersQuery.data?.items ?? []

  const groupedTransactions = useMemo(() => {
    const map = new Map<string, typeof transactions>()
    transactions.forEach((transaction) => {
      const key = formatShortDate(transaction.occurred_at)
      const existing = map.get(key) ?? []
      map.set(key, [...existing, transaction])
    })
    return Array.from(map.entries())
  }, [transactions])

  async function handleCreateTransaction(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const ledger = ledgers[0]
    await createTransactionMutation.mutateAsync({
      ledger_id: ledger?.id ?? '',
      account_id: accountId,
      category_id: categoryId,
      type: 'expense',
      amount: Number(amount),
    })
    setShowAddDialog(false)
    setAmount('')
    setCategoryId('')
    setAccountId('')
  }

  async function handlePreviewImport(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!importFile) {
      return
    }
    await importPreviewMutation.mutateAsync(importFile)
  }

  return (
    <div className="space-y-8">
      <PageSection
        eyebrow="Transaction ledger"
        title="Transactions"
        description="Review, classify, and import money movement across every connected account and ledger."
        actions={
          <>
            <Button variant="secondary" onClick={() => setShowImportDialog(true)}>
              Import
            </Button>
            <Button onClick={() => setShowAddDialog(true)}>Add Transaction</Button>
          </>
        }
      >
        <div className="flex flex-wrap gap-3">
          <Button variant={view === 'list' ? 'primary' : 'secondary'} onClick={() => setView('list')}>
            List View
          </Button>
          <Button variant={view === 'calendar' ? 'primary' : 'secondary'} onClick={() => setView('calendar')}>
            Calendar View
          </Button>
        </div>

        {view === 'list' ? (
          <div className="mt-8 space-y-4">
            {transactions.map((transaction) => (
              <article key={transaction.id} className="rounded-[24px] bg-surface-container-low p-5">
                <div className="flex items-center justify-between gap-4">
                  <div>
                    <p className="font-medium text-on-surface">{transaction.category_name ?? 'Uncategorized'}</p>
                    <p className="mt-1 text-xs text-on-surface-variant">{formatShortDate(transaction.occurred_at)} • {transaction.type}</p>
                  </div>
                  <p className="font-headline text-xl font-bold text-on-surface">{formatCurrency(transaction.amount)}</p>
                </div>
              </article>
            ))}
          </div>
        ) : (
          <div className="mt-8 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
            {groupedTransactions.map(([day, items]) => (
              <article key={day} className="rounded-[24px] bg-surface-container-low p-5">
                <p className="font-headline text-xl font-bold text-on-surface">{day}</p>
                <div className="mt-4 space-y-3">
                  {items.map((transaction) => (
                    <div key={transaction.id} className="rounded-2xl bg-surface-container-lowest p-3">
                      <p className="font-medium text-on-surface">{transaction.category_name ?? 'Uncategorized'}</p>
                      <p className="mt-1 text-sm text-on-surface-variant">{formatCurrency(transaction.amount)}</p>
                    </div>
                  ))}
                </div>
              </article>
            ))}
          </div>
        )}
      </PageSection>

      {showAddDialog ? (
        <DialogShell title="Add Transaction" description="Capture a new expense or income entry." footer={<Button type="submit" form="add-transaction-form">Save Transaction</Button>}>
          <form id="add-transaction-form" className="grid gap-6 md:grid-cols-2" onSubmit={(event) => void handleCreateTransaction(event)}>
            <TextField label="Amount" type="number" value={amount} onChange={(event) => setAmount(event.target.value)} />
            <SelectField label="Category" value={categoryId} onChange={(event) => setCategoryId(event.target.value)}>
              <option value="">Select category</option>
              {categories.map((category) => (
                <option key={category.id} value={category.id}>
                  {category.name}
                </option>
              ))}
            </SelectField>
            <SelectField label="Account" value={accountId} onChange={(event) => setAccountId(event.target.value)}>
              <option value="">Select account</option>
              {accounts.map((account) => (
                <option key={account.id} value={account.id}>
                  {account.name}
                </option>
              ))}
            </SelectField>
            <SelectField label="Ledger" defaultValue={ledgers[0]?.id ?? ''} disabled>
              {ledgers.map((ledger) => (
                <option key={ledger.id} value={ledger.id}>
                  {ledger.name}
                </option>
              ))}
            </SelectField>
          </form>
        </DialogShell>
      ) : null}

      {showImportDialog ? (
        <DialogShell title="Import Transactions" description="Upload a CSV file to preview column mappings before confirmation." footer={<Button type="submit" form="import-preview-form">Preview Import</Button>}>
          <form id="import-preview-form" className="space-y-6" onSubmit={(event) => void handlePreviewImport(event)}>
            <label className="block space-y-2" htmlFor="csv-file">
              <span className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">CSV File</span>
              <input id="csv-file" type="file" accept=".csv,text/csv" onChange={(event) => setImportFile(event.target.files?.[0] ?? null)} />
            </label>

            {importFile ? <p className="text-sm text-on-surface">{importFile.name}</p> : null}

            {importPreviewMutation.data ? (
              <div className="rounded-[24px] bg-surface-container-low p-5">
                <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">Detected columns</p>
                <div className="mt-4 flex flex-wrap gap-2">
                  {importPreviewMutation.data.columns.map((column) => (
                    <span key={column} className="rounded-full bg-surface-container-lowest px-3 py-1 text-sm text-on-surface">
                      {column}
                    </span>
                  ))}
                </div>
              </div>
            ) : null}
          </form>
        </DialogShell>
      ) : null}
    </div>
  )
}
