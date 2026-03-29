import { ChevronLeft, ChevronRight, Search, Upload } from 'lucide-react'
import { useMemo, useState } from 'react'
import { Button } from '@/components/ui/button'
import { DialogShell } from '@/components/ui/dialog-shell'
import { SelectField } from '@/components/ui/select-field'
import { TextField } from '@/components/ui/text-field'
import type { TransactionRecord } from '@/features/transactions/transactions-api'
import {
  useCreateTransaction,
  useImportConfirm,
  useImportPreview,
  useTransactionFormOptions,
  useTransactionsWithOptions,
} from '@/features/transactions/transactions-hooks'
import { formatCurrency } from '@/lib/format'

function toLocalDateKey(value: Date) {
  return `${value.getFullYear()}-${String(value.getMonth() + 1).padStart(2, '0')}-${String(value.getDate()).padStart(2, '0')}`
}

function getMonthGrid(baseDate: Date) {
  const year = baseDate.getFullYear()
  const month = baseDate.getMonth()
  const firstDay = new Date(year, month, 1)
  const firstWeekDay = firstDay.getDay()
  const gridStart = new Date(year, month, 1 - firstWeekDay)
  const cells: Date[] = []
  for (let i = 0; i < 42; i += 1) {
    const d = new Date(gridStart)
    d.setDate(gridStart.getDate() + i)
    cells.push(d)
  }
  return cells
}

export function TransactionsPage() {
  const [view, setView] = useState<'list' | 'calendar'>('list')
  const [showAddDialog, setShowAddDialog] = useState(false)
  const [showImportDialog, setShowImportDialog] = useState(false)
  const [activeMonth, setActiveMonth] = useState(() => new Date())
  const [selectedDay, setSelectedDay] = useState<string | null>(null)
  const [transactionType, setTransactionType] = useState<'income' | 'expense'>('expense')
  const [amount, setAmount] = useState('')
  const [categoryId, setCategoryId] = useState('')
  const [accountId, setAccountId] = useState('')
  const [memo, setMemo] = useState('')
  const [date, setDate] = useState(() => new Date().toISOString().slice(0, 10))
  const [importFile, setImportFile] = useState<File | null>(null)
  const [dateRangePreset, setDateRangePreset] = useState<'7' | '30' | '120' | '365'>('120')

  const monthCells = useMemo(() => getMonthGrid(activeMonth), [activeMonth])
  const monthRange = useMemo(() => {
    const start = monthCells[0]
    const end = monthCells[monthCells.length - 1]
    const startUtc = new Date(Date.UTC(start.getFullYear(), start.getMonth(), start.getDate(), 0, 0, 0))
    const endUtc = new Date(Date.UTC(end.getFullYear(), end.getMonth(), end.getDate(), 23, 59, 59))
    return { from: startUtc.toISOString(), to: endUtc.toISOString() }
  }, [monthCells])

  const listRange = useMemo(() => {
    const now = new Date()
    const start = new Date(now)
    const days = Number(dateRangePreset)
    start.setDate(now.getDate() - days)
    return {
      from: start.toISOString(),
      to: now.toISOString(),
    }
  }, [dateRangePreset])

  const calendarTransactionsQuery = useTransactionsWithOptions({
    page: 1,
    pageSize: 1000,
    dateFrom: monthRange.from,
    dateTo: monthRange.to,
  })
  const listTransactionsQuery = useTransactionsWithOptions({
    page: 1,
    pageSize: 200,
    dateFrom: listRange.from,
    dateTo: listRange.to,
  })
  const options = useTransactionFormOptions()
  const createTransactionMutation = useCreateTransaction()
  const importPreviewMutation = useImportPreview()
  const importConfirmMutation = useImportConfirm()

  const calendarTransactions = calendarTransactionsQuery.data?.items ?? []
  const listTransactions = listTransactionsQuery.data?.items ?? []
  const categories = options.categoriesQuery.data?.items ?? []
  const accounts = options.accountsQuery.data?.items ?? []
  const ledgers = options.ledgersQuery.data?.items ?? []
  const tags = options.tagsQuery.data?.items ?? []

  const txByDay = useMemo(() => {
    const map = new Map<string, TransactionRecord[]>()
    calendarTransactions.forEach((tx) => {
      const dt = new Date(tx.occurred_at)
      const key = toLocalDateKey(dt)
      const list = map.get(key) ?? []
      map.set(key, [...list, tx])
    })
    return map
  }, [calendarTransactions])

  const monthTotals = useMemo(() => {
    const currentMonth = activeMonth.getMonth()
    const currentYear = activeMonth.getFullYear()
    return calendarTransactions
      .filter((tx) => {
        const dt = new Date(tx.occurred_at)
        return dt.getMonth() === currentMonth && dt.getFullYear() === currentYear
      })
      .reduce(
        (acc, tx) => {
          if (tx.type === 'income') acc.in += tx.amount
          if (tx.type === 'expense') acc.out += tx.amount
          return acc
        },
        { in: 0, out: 0 },
      )
  }, [calendarTransactions, activeMonth])

  const fallbackSelectedDay = useMemo(() => {
    const first = calendarTransactions[0]
    if (!first) return toLocalDateKey(new Date())
    return toLocalDateKey(new Date(first.occurred_at))
  }, [calendarTransactions])

  const effectiveSelectedDay = selectedDay ?? fallbackSelectedDay
  const selectedDayTx = txByDay.get(effectiveSelectedDay) ?? []
  const selectedTotals = selectedDayTx.reduce(
    (acc, tx) => {
      if (tx.type === 'income') acc.in += tx.amount
      if (tx.type === 'expense') acc.out += tx.amount
      return acc
    },
    { in: 0, out: 0 },
  )

  async function handleCreateTransaction(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const ledger = ledgers[0]
    if (!ledger?.id) return
    await createTransactionMutation.mutateAsync({
      ledger_id: ledger.id,
      account_id: accountId || undefined,
      category_id: categoryId || undefined,
      type: transactionType,
      amount: Number(amount),
      memo: memo || undefined,
    })
    setShowAddDialog(false)
    setAmount('')
    setCategoryId('')
    setAccountId('')
    setMemo('')
  }

  async function handlePreviewImport(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!importFile) return
    await importPreviewMutation.mutateAsync(importFile)
  }

  async function handleConfirmImport() {
    if (!importFile) return
    const idempotencyKey = `ui-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
    await importConfirmMutation.mutateAsync({ file: importFile, idempotencyKey })
    setShowImportDialog(false)
    setImportFile(null)
  }

  return (
    <div className="space-y-5">
      <section className="rounded-[28px] border border-outline/15 bg-surface-container-lowest p-6 shadow-ambient md:p-7">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div className="flex items-center gap-4">
            <h2 className="font-headline text-[48px] font-extrabold leading-none tracking-tight text-primary">Transactions</h2>
            <div className="inline-flex rounded-xl border border-outline/15 bg-surface-container p-1">
              <button
                type="button"
                className={`rounded-lg px-4 py-2 text-xs font-semibold ${view === 'list' ? 'bg-white text-primary shadow-sm' : 'text-on-surface-variant hover:text-primary'}`}
                onClick={() => setView('list')}
              >
                List View
              </button>
              <button
                type="button"
                className={`rounded-lg px-4 py-2 text-xs font-semibold ${view === 'calendar' ? 'bg-white text-primary shadow-sm' : 'text-on-surface-variant hover:text-primary'}`}
                onClick={() => setView('calendar')}
              >
                Calendar View
              </button>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <Button variant="secondary" onClick={() => setShowImportDialog(true)}>
              <Upload className="h-4 w-4" />
              Import
            </Button>
            <Button onClick={() => setShowAddDialog(true)}>+ Add Transaction</Button>
          </div>
        </div>

        {view === 'list' ? (
          <div className="mt-6 space-y-4">
            <article className="rounded-2xl border border-outline/10 bg-surface-container-low p-5">
              <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
                <label className="space-y-2">
                  <span className="text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">Search Ledger</span>
                  <div className="relative">
                    <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-on-surface-variant" />
                    <input className="h-11 w-full rounded-xl border border-outline/20 bg-white pl-9 pr-3 text-sm" placeholder="Transaction ID, vendor, or note..." />
                  </div>
                </label>
                <label className="space-y-2">
                  <span className="text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">Account</span>
                  <select className="h-11 w-full rounded-xl border border-outline/20 bg-white px-3 text-sm">
                    <option>All Accounts</option>
                  </select>
                </label>
                <label className="space-y-2">
                  <span className="text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">Ledger</span>
                  <select className="h-11 w-full rounded-xl border border-outline/20 bg-white px-3 text-sm">
                    {ledgers.slice(0, 1).map((ledger) => (
                      <option key={ledger.id}>{ledger.name}</option>
                    ))}
                  </select>
                </label>
                <label className="space-y-2">
                  <span className="text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">Date Range</span>
                  <select
                    className="h-11 w-full rounded-xl border border-outline/20 bg-white px-3 text-sm"
                    value={dateRangePreset}
                    onChange={(event) => setDateRangePreset(event.target.value as '7' | '30' | '120' | '365')}
                  >
                    <option value="7">Last 7 Days</option>
                    <option value="30">Last 30 Days</option>
                    <option value="120">Last 120 Days</option>
                    <option value="365">Last 365 Days</option>
                  </select>
                </label>
              </div>
            </article>

            <article className="overflow-hidden rounded-2xl border border-outline/15 bg-white">
              <div className="grid grid-cols-[1.7fr_1fr_1fr_1fr_0.8fr_1fr] bg-surface-container-low px-5 py-3 text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">
                <p>Transaction & Category</p>
                <p>Account / Ledger</p>
                <p>Date & Time</p>
                <p>Income / Expense Note</p>
                <p>Tags</p>
                <p className="text-right">Amount</p>
              </div>
              {listTransactions.map((tx) => (
                <div key={tx.id} className="grid grid-cols-[1.7fr_1fr_1fr_1fr_0.8fr_1fr] items-center border-t border-outline/10 px-5 py-4">
                  <div>
                    <p className="font-semibold text-on-surface">{tx.category_name ?? 'Uncategorized'}</p>
                    <p className="text-xs text-on-surface-variant">备注：{tx.memo?.trim() || '暂无备注'}</p>
                  </div>
                  <div>
                    <p className="text-sm text-on-surface">{accounts[0]?.name ?? 'Account'}</p>
                    <p className="text-xs uppercase text-on-surface-variant">{ledgers[0]?.name ?? 'Ledger'}</p>
                  </div>
                  <div>
                    <p className="text-sm text-on-surface">{new Date(tx.occurred_at).toLocaleDateString()}</p>
                    <p className="text-xs text-on-surface-variant">{new Date(tx.occurred_at).toLocaleTimeString()}</p>
                  </div>
                  <div>
                    <p className="text-xs font-semibold uppercase text-on-surface">{tx.type === 'income' ? 'Income' : tx.type === 'expense' ? 'Expense' : 'Transfer'}</p>
                    <p className="mt-1 text-xs text-on-surface-variant">{tx.memo?.trim() || (tx.type === 'income' ? 'Incoming flow' : tx.type === 'expense' ? 'Outgoing payment' : 'Internal transfer')}</p>
                  </div>
                  <div>
                    <span className="rounded-full bg-surface-container-low px-2 py-1 text-[10px] font-semibold uppercase text-on-surface-variant">
                      {tx.type}
                    </span>
                  </div>
                  <p className={`text-right text-4xl font-extrabold ${tx.type === 'income' ? 'text-emerald-600' : 'text-rose-600'}`}>
                    {tx.type === 'income' ? '+' : '-'}
                    {formatCurrency(Math.abs(tx.amount))}
                  </p>
                </div>
              ))}
              {listTransactions.length === 0 ? (
                <div className="border-t border-outline/10 px-5 py-8 text-center text-sm text-on-surface-variant">
                  No transactions yet.
                </div>
              ) : null}
            </article>
          </div>
        ) : (
          <div className="mt-6 grid gap-4 xl:grid-cols-[1.5fr_0.85fr]">
            <article className="rounded-2xl border border-outline/15 bg-white p-5">
              <div className="mb-4 flex items-center gap-3">
                <button type="button" onClick={() => setActiveMonth(new Date(activeMonth.getFullYear(), activeMonth.getMonth() - 1, 1))}>
                  <ChevronLeft className="h-4 w-4 text-primary" />
                </button>
                <h3 className="font-headline text-4xl font-bold text-on-surface">
                  {activeMonth.toLocaleString('en-US', { month: 'long', year: 'numeric' })}
                </h3>
                <button type="button" onClick={() => setActiveMonth(new Date(activeMonth.getFullYear(), activeMonth.getMonth() + 1, 1))}>
                  <ChevronRight className="h-4 w-4 text-primary" />
                </button>
                <button type="button" className="ml-2 text-xs font-semibold text-primary" onClick={() => setActiveMonth(new Date())}>
                  Today
                </button>
              </div>
              <div className="grid grid-cols-7 border border-outline/15">
                {['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'].map((d) => (
                  <div key={d} className="border-b border-outline/10 bg-surface-container-low p-2 text-center text-[10px] font-bold uppercase tracking-[0.12em] text-on-surface-variant">
                    {d}
                  </div>
                ))}
                {monthCells.map((cell) => {
                  const key = toLocalDateKey(cell)
                  const items = txByDay.get(key) ?? []
                  const totals = items.reduce(
                    (acc, tx) => {
                      if (tx.type === 'income') acc.in += tx.amount
                      if (tx.type === 'expense') acc.out += tx.amount
                      return acc
                    },
                    { in: 0, out: 0 },
                  )
                  const inMonth = cell.getMonth() === activeMonth.getMonth()
                  const selected = effectiveSelectedDay === key
                  return (
                    <button
                      key={key}
                      type="button"
                      className={`min-h-[112px] border border-outline/10 p-2 text-left transition ${
                        selected ? 'bg-blue-50 ring-1 ring-primary' : inMonth ? 'bg-white hover:bg-surface-container-low' : 'bg-surface-container-low'
                      }`}
                      onClick={() => setSelectedDay(key)}
                    >
                      <p className={`text-xs font-semibold ${inMonth ? 'text-on-surface' : 'text-on-surface-variant'}`}>{cell.getDate()}</p>
                      <div className="mt-1 space-y-1">
                        {totals.out > 0 ? <p className="rounded bg-rose-100 px-1 py-0.5 text-[10px] font-semibold text-rose-700">-{formatCurrency(totals.out)}</p> : null}
                        {totals.in > 0 ? <p className="rounded bg-emerald-100 px-1 py-0.5 text-[10px] font-semibold text-emerald-700">+{formatCurrency(totals.in)}</p> : null}
                      </div>
                    </button>
                  )
                })}
              </div>
            </article>

            <div className="space-y-4">
              <article className="rounded-2xl border border-primary/45 bg-white p-5">
                <p className="text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">Daily Summary</p>
                <h4 className="mt-2 font-headline text-5xl font-bold text-on-surface">
                  {new Date(effectiveSelectedDay).toLocaleDateString()}
                </h4>
                <div className="mt-4 grid grid-cols-2 gap-3">
                  <div className="rounded-xl bg-surface-container-low p-3">
                    <p className="text-[10px] font-bold uppercase tracking-[0.12em] text-on-surface-variant">Total Out</p>
                    <p className="mt-2 font-headline text-4xl font-extrabold text-rose-700">-{formatCurrency(selectedTotals.out)}</p>
                  </div>
                  <div className="rounded-xl bg-surface-container-low p-3">
                    <p className="text-[10px] font-bold uppercase tracking-[0.12em] text-on-surface-variant">Total In</p>
                    <p className="mt-2 font-headline text-4xl font-extrabold text-emerald-700">+{formatCurrency(selectedTotals.in)}</p>
                  </div>
                </div>
                <div className="mt-5 space-y-3">
                  {selectedDayTx.slice(0, 6).map((tx) => (
                    <div key={tx.id} className="rounded-xl bg-surface-container-low p-3">
                      <div className="flex items-center justify-between">
                        <p className="font-semibold text-on-surface">{tx.category_name ?? 'Uncategorized'}</p>
                        <p className={tx.type === 'income' ? 'font-bold text-emerald-700' : 'font-bold text-rose-700'}>
                          {tx.type === 'income' ? '+' : '-'}
                          {formatCurrency(Math.abs(tx.amount))}
                        </p>
                      </div>
                      <p className="mt-1 text-xs text-on-surface-variant">{tx.memo?.trim() || tx.type}</p>
                    </div>
                  ))}
                  {selectedDayTx.length === 0 ? <p className="text-sm text-on-surface-variant">No transactions on this date.</p> : null}
                </div>
              </article>
              <article className="rounded-2xl bg-primary p-5 text-white">
                <p className="text-[10px] font-bold uppercase tracking-[0.14em] text-primary-fixed">Month Summary</p>
                <div className="mt-3 grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-xs text-primary-fixed">Income</p>
                    <p className="font-headline text-3xl font-extrabold">+{formatCurrency(monthTotals.in)}</p>
                  </div>
                  <div>
                    <p className="text-xs text-primary-fixed">Expense</p>
                    <p className="font-headline text-3xl font-extrabold">-{formatCurrency(monthTotals.out)}</p>
                  </div>
                </div>
                <div className="mt-4 border-t border-white/20 pt-4">
                  <p className="text-xs text-primary-fixed">Net</p>
                  <p className="font-headline text-4xl font-extrabold">{formatCurrency(monthTotals.in - monthTotals.out)}</p>
                </div>
              </article>
            </div>
          </div>
        )}
      </section>

      {showImportDialog ? (
        <DialogShell
          title="Import Ledger Data"
          description="Upload your financial statements to synchronize your ledger."
          onClose={() => setShowImportDialog(false)}
          footer={
            <>
              <Button variant="ghost" onClick={() => setShowImportDialog(false)}>
                Cancel
              </Button>
              <Button type="submit" form="import-preview-form">
                Preview Import
              </Button>
            </>
          }
        >
          <form id="import-preview-form" className="space-y-5" onSubmit={(event) => void handlePreviewImport(event)}>
            <label
              htmlFor="csv-file"
              className="flex min-h-60 cursor-pointer flex-col items-center justify-center rounded-2xl border-2 border-dashed border-outline-variant bg-surface-container-low p-8 text-center"
            >
              <Upload className="h-10 w-10 text-primary" />
              <p className="mt-4 text-3xl font-semibold text-on-surface">Click to upload or drag and drop</p>
              <p className="mt-2 text-sm text-on-surface-variant">Supported formats: .CSV (max file size 10MB)</p>
              <span className="mt-6 rounded-lg border border-primary px-6 py-2 text-sm font-semibold text-primary">Select Files</span>
              <input id="csv-file" type="file" accept=".csv,text/csv" className="hidden" onChange={(event) => setImportFile(event.target.files?.[0] ?? null)} />
            </label>
            {importFile ? <p className="text-sm text-on-surface">Selected: {importFile.name}</p> : null}
            {importPreviewMutation.isError ? (
              <div className="rounded-xl border border-error bg-error-container p-4">
                <p className="text-sm text-on-error-container">
                  Preview failed: {(importPreviewMutation.error as Error)?.message ?? 'Unknown error'}
                </p>
              </div>
            ) : null}
            {importPreviewMutation.data ? (
              <div className="rounded-xl border border-outline/10 bg-surface-container-low p-4">
                <p className="text-xs font-bold uppercase tracking-[0.12em] text-on-surface-variant">Detected Columns</p>
                <div className="mt-3 flex flex-wrap gap-2">
                  {importPreviewMutation.data.columns.map((column) => (
                    <span key={column} className="rounded-full bg-white px-3 py-1 text-sm text-on-surface">
                      {column}
                    </span>
                  ))}
                </div>
                <div className="mt-4 flex items-center gap-3">
                  <Button onClick={() => void handleConfirmImport()} disabled={importConfirmMutation.isPending}>
                    {importConfirmMutation.isPending ? 'Importing...' : 'Confirm Import'}
                  </Button>
                  {importConfirmMutation.isError ? (
                    <span className="text-sm text-error">
                      Import failed: {(importConfirmMutation.error as Error)?.message ?? 'Unknown error'}
                    </span>
                  ) : null}
                  {importConfirmMutation.isSuccess ? (
                    <span className="text-sm text-emerald-600">Import successful!</span>
                  ) : null}
                </div>
              </div>
            ) : null}
          </form>
        </DialogShell>
      ) : null}

      {showAddDialog ? (
        <DialogShell
          title="Add New Transaction"
          description="Record a new financial entry in your digital ledger."
          onClose={() => setShowAddDialog(false)}
          footer={
            <>
              <Button variant="ghost" onClick={() => setShowAddDialog(false)}>
                Cancel
              </Button>
              <Button type="submit" form="add-transaction-form">
                Save Transaction
              </Button>
            </>
          }
        >
          <form id="add-transaction-form" className="grid gap-5 md:grid-cols-2" onSubmit={(event) => void handleCreateTransaction(event)}>
            <TextField label="Amount" type="number" step="0.01" value={amount} onChange={(event) => setAmount(event.target.value)} placeholder="0.00" />
            <TextField label="Date" type="date" value={date} onChange={(event) => setDate(event.target.value)} />
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
            <TextField label="Memo" value={memo} onChange={(event) => setMemo(event.target.value)} placeholder="Internal reference or notes" />
            <div className="md:col-span-2">
              <SelectField label="Type" value={transactionType} onChange={(event) => setTransactionType(event.target.value as 'income' | 'expense')}>
                <option value="expense">Expense (支出)</option>
                <option value="income">Income (收入)</option>
              </SelectField>
            </div>
            <div className="md:col-span-2">
              <p className="mb-2 text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">Tags</p>
              <div className="flex flex-wrap gap-2">
                {tags.slice(0, 6).map((tag) => (
                  <span key={tag.id} className="rounded-full bg-surface-container-low px-3 py-1 text-sm font-semibold text-primary">
                    {tag.name}
                  </span>
                ))}
              </div>
            </div>
          </form>
        </DialogShell>
      ) : null}
    </div>
  )
}

