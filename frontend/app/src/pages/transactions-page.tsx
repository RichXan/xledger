import { ChevronLeft, ChevronRight, Download, Search, Trash2, Undo2, Upload } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { DialogShell } from '@/components/ui/dialog-shell'
import { SelectField } from '@/components/ui/select-field'
import { TextField } from '@/components/ui/text-field'
import type { TransactionRecord } from '@/features/transactions/transactions-api'
import {
  useCreateTransaction,
  useDeleteTransaction,
  useExportTransactions,
  useImportConfirm,
  useImportPreview,
  useTransactionFormOptions,
  useTransactionsWithOptions,
} from '@/features/transactions/transactions-hooks'
import { formatCurrency } from '@/lib/format'

function toLocalDateKey(value: Date) {
  return `${value.getFullYear()}-${String(value.getMonth() + 1).padStart(2, '0')}-${String(value.getDate()).padStart(2, '0')}`
}

function toLocalDateTimeInputValue(value: Date) {
  const year = value.getFullYear()
  const month = String(value.getMonth() + 1).padStart(2, '0')
  const day = String(value.getDate()).padStart(2, '0')
  const hours = String(value.getHours()).padStart(2, '0')
  const minutes = String(value.getMinutes()).padStart(2, '0')
  const seconds = String(value.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day}T${hours}:${minutes}:${seconds}`
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

type QuickFilter = 'all' | 'income' | 'expense' | 'uncategorized' | 'week' | 'large'

function bytesToHex(bytes: Uint8Array) {
  return Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('')
}

async function readFileBuffer(file: File): Promise<ArrayBuffer> {
  const readableFile = file as File & {
    arrayBuffer?: () => Promise<ArrayBuffer>
    text?: () => Promise<string>
  }

  if (typeof readableFile.arrayBuffer === 'function') {
    return readableFile.arrayBuffer()
  }
  if (typeof readableFile.text === 'function') {
    const bytes = new TextEncoder().encode(await readableFile.text())
    return bytes.buffer.slice(bytes.byteOffset, bytes.byteOffset + bytes.byteLength)
  }

  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      if (reader.result instanceof ArrayBuffer) {
        resolve(reader.result)
        return
      }
      reject(new Error('Unsupported file reader result'))
    }
    reader.onerror = () => reject(reader.error ?? new Error('Unable to read import file'))
    reader.readAsArrayBuffer(file)
  })
}

export async function buildImportIdempotencyKey(file: File) {
  const buffer = await readFileBuffer(file)
  if (globalThis.crypto?.subtle) {
    const digest = await globalThis.crypto.subtle.digest('SHA-256', buffer)
    return `ui-file-${bytesToHex(new Uint8Array(digest))}`
  }

  const bytes = new Uint8Array(buffer)
  let hash = 0
  bytes.forEach((byte) => {
    hash = (hash * 31 + byte) % 4294967295
  })
  return `ui-file-${file.size}-${hash.toString(16).padStart(8, '0')}`
}

export function getTransactionAccountLabel(
  tx: TransactionRecord,
  accountNameById: ReadonlyMap<string, string>,
  fallback: string,
) {
  const accountID = tx.account_id ?? tx.from_account_id ?? tx.to_account_id
  if (!accountID) return fallback
  return accountNameById.get(accountID) ?? accountID
}

export function getTransactionLedgerLabel(
  tx: TransactionRecord,
  ledgerNameById: ReadonlyMap<string, string>,
  fallback: string,
) {
  if (!tx.ledger_id) return fallback
  return ledgerNameById.get(tx.ledger_id) ?? tx.ledger_id
}

export function TransactionsPage() {
  const { t, i18n } = useTranslation()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const searchParamQ = (searchParams.get('q') ?? '').trim()

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
  const [date, setDate] = useState(() => toLocalDateTimeInputValue(new Date()))
  const [importFile, setImportFile] = useState<File | null>(null)
  const [dateRangePreset, setDateRangePreset] = useState<'7' | '30' | '120' | '365'>('120')
  const [listSearchQuery, setListSearchQuery] = useState(searchParamQ)
  const [selectedAccountFilter, setSelectedAccountFilter] = useState('')
  const [selectedLedgerFilter, setSelectedLedgerFilter] = useState('')
  const [quickFilter, setQuickFilter] = useState<QuickFilter>('all')
  const [deletedTransactionIds, setDeletedTransactionIds] = useState<Set<string>>(() => new Set())
  const [pendingUndo, setPendingUndo] = useState<TransactionRecord | null>(null)

  useEffect(() => {
    setListSearchQuery(searchParamQ)
  }, [searchParamQ])

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
    accountId: selectedAccountFilter || undefined,
    ledgerId: selectedLedgerFilter || undefined,
  })
  const options = useTransactionFormOptions()
  const createTransactionMutation = useCreateTransaction()
  const importPreviewMutation = useImportPreview()
  const importConfirmMutation = useImportConfirm()
  const exportTransactionsMutation = useExportTransactions()
  const deleteTransactionMutation = useDeleteTransaction()

  const calendarTransactions = calendarTransactionsQuery.data?.items ?? []
  const listTransactions = listTransactionsQuery.data?.items ?? []
  const categories = options.categoriesQuery.data?.items ?? []
  const accounts = options.accountsQuery.data?.items ?? []
  const ledgers = options.ledgersQuery.data?.items ?? []
  const tags = options.tagsQuery.data?.items ?? []

  const accountNameById = useMemo(() => new Map(accounts.map((account) => [account.id, account.name])), [accounts])
  const ledgerNameById = useMemo(() => new Map(ledgers.map((ledger) => [ledger.id, ledger.name])), [ledgers])
  const normalizedListSearchQuery = listSearchQuery.trim().toLowerCase()

  const filteredListTransactions = useMemo(() => {
    const now = new Date()
    const weekStart = new Date(now)
    weekStart.setDate(now.getDate() - 7)
    return listTransactions.filter((tx) => {
      if (deletedTransactionIds.has(tx.id)) {
        return false
      }
      if (normalizedListSearchQuery) {
        const candidates = [tx.id, tx.category_name ?? '', tx.memo ?? '', tx.type, tx.occurred_at]
        if (!candidates.join(' ').toLowerCase().includes(normalizedListSearchQuery)) {
          return false
        }
      }
      if (quickFilter === 'income') return tx.type === 'income'
      if (quickFilter === 'expense') return tx.type === 'expense'
      if (quickFilter === 'uncategorized') return !tx.category_name?.trim()
      if (quickFilter === 'week') return new Date(tx.occurred_at) >= weekStart
      if (quickFilter === 'large') return Math.abs(tx.amount) >= 1000
      return true
    })
  }, [deletedTransactionIds, listTransactions, normalizedListSearchQuery, quickFilter])
  const hasActiveListFilters = Boolean(
    normalizedListSearchQuery || selectedAccountFilter || selectedLedgerFilter || dateRangePreset !== '120' || quickFilter !== 'all',
  )
  const isInitialEmptyState = filteredListTransactions.length === 0 && !hasActiveListFilters

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

    const parsedDate = date ? new Date(date) : null
    const occurredAt =
      parsedDate && Number.isFinite(parsedDate.getTime()) ? parsedDate.toISOString() : undefined

    await createTransactionMutation.mutateAsync({
      ledger_id: ledger.id,
      account_id: accountId || undefined,
      category_id: categoryId || undefined,
      type: transactionType,
      amount: Number(amount),
      memo: memo || undefined,
      occurred_at: occurredAt,
    })
    setShowAddDialog(false)
    setAmount('')
    setCategoryId('')
    setAccountId('')
    setMemo('')
    setDate(toLocalDateTimeInputValue(new Date()))
  }

  async function handlePreviewImport(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!importFile) return
    await importPreviewMutation.mutateAsync(importFile)
  }

  async function handleConfirmImport() {
    if (!importFile) return
    const idempotencyKey = await buildImportIdempotencyKey(importFile)
    await importConfirmMutation.mutateAsync({ file: importFile, idempotencyKey })
    setShowImportDialog(false)
    setImportFile(null)
  }

  async function handleExportTransactions() {
    const exportRange = view === 'calendar' ? monthRange : listRange
    const content = await exportTransactionsMutation.mutateAsync({
      format: 'csv',
      dateFrom: exportRange.from,
      dateTo: exportRange.to,
      accountId: selectedAccountFilter || undefined,
      ledgerId: selectedLedgerFilter || undefined,
    })
    const blob = new Blob([content], { type: 'text/csv;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const anchor = document.createElement('a')
    const fromDate = exportRange.from.slice(0, 10)
    const toDate = exportRange.to.slice(0, 10)
    anchor.href = url
    anchor.download = `xledger-transactions-${fromDate}-to-${toDate}.csv`
    document.body.appendChild(anchor)
    anchor.click()
    anchor.remove()
    URL.revokeObjectURL(url)
  }

  async function handleDeleteTransaction(tx: TransactionRecord) {
    await deleteTransactionMutation.mutateAsync(tx.id)
    setDeletedTransactionIds((current) => new Set(current).add(tx.id))
    setPendingUndo(tx)
  }

  async function handleUndoDelete() {
    if (!pendingUndo || pendingUndo.type === 'transfer') return
    const ledger = ledgers[0]
    if (!ledger?.id) return

    await createTransactionMutation.mutateAsync({
      ledger_id: ledger.id,
      type: pendingUndo.type,
      amount: Math.abs(pendingUndo.amount),
      memo: pendingUndo.memo || undefined,
      occurred_at: pendingUndo.occurred_at,
    })
    setDeletedTransactionIds((current) => {
      const next = new Set(current)
      next.delete(pendingUndo.id)
      return next
    })
    setPendingUndo(null)
  }

  const locale = i18n.language === 'zh' ? 'zh-CN' : 'en-US'
  const quickFilters: Array<{ id: QuickFilter; label: string }> = [
    { id: 'all', label: t('transactionsPage.quickFilters.all') },
    { id: 'income', label: t('transactionsPage.quickFilters.income') },
    { id: 'expense', label: t('transactionsPage.quickFilters.expense') },
    { id: 'uncategorized', label: t('transactionsPage.quickFilters.uncategorized') },
    { id: 'week', label: t('transactionsPage.quickFilters.week') },
    { id: 'large', label: t('transactionsPage.quickFilters.large') },
  ]
  const weekdayLabels = [
    t('transactionsPage.calendar.weekdays.sun'),
    t('transactionsPage.calendar.weekdays.mon'),
    t('transactionsPage.calendar.weekdays.tue'),
    t('transactionsPage.calendar.weekdays.wed'),
    t('transactionsPage.calendar.weekdays.thu'),
    t('transactionsPage.calendar.weekdays.fri'),
    t('transactionsPage.calendar.weekdays.sat'),
  ]

  return (
    <div className="space-y-5">
      <section className="rounded-[28px] border border-outline/15 bg-surface-container-lowest p-6 shadow-ambient md:p-7">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div className="flex items-center gap-4">
            <h2 className="font-headline text-[48px] font-extrabold leading-none tracking-tight text-primary">
              {t('transactionsPage.title')}
            </h2>
            <div className="inline-flex rounded-xl border border-outline/15 bg-surface-container p-1">
              <button
                type="button"
                className={`rounded-lg px-4 py-2 text-xs font-semibold ${view === 'list' ? 'bg-white text-primary shadow-sm' : 'text-on-surface-variant hover:text-primary'}`}
                onClick={() => setView('list')}
              >
                {t('transactionsPage.listView')}
              </button>
              <button
                type="button"
                className={`rounded-lg px-4 py-2 text-xs font-semibold ${view === 'calendar' ? 'bg-white text-primary shadow-sm' : 'text-on-surface-variant hover:text-primary'}`}
                onClick={() => setView('calendar')}
              >
                {t('transactionsPage.calendarView')}
              </button>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <Button
              variant="secondary"
              onClick={() => void handleExportTransactions()}
              disabled={exportTransactionsMutation.isPending}
            >
              <Download className="h-4 w-4" />
              {exportTransactionsMutation.isPending ? t('transactionsPage.exporting') : t('transactionsPage.export')}
            </Button>
            <Button variant="secondary" onClick={() => setShowImportDialog(true)}>
              <Upload className="h-4 w-4" />
              {t('transactionsPage.import')}
            </Button>
            <Button onClick={() => setShowAddDialog(true)}>{t('transactionsPage.addTransaction')}</Button>
          </div>
        </div>
        {exportTransactionsMutation.isError ? (
          <p className="mt-3 text-sm text-error">
            {t('transactionsPage.exportFailed', {
              message: (exportTransactionsMutation.error as Error)?.message ?? t('common.unknownError'),
            })}
          </p>
        ) : null}

        {view === 'list' ? (
          <div className="mt-6 space-y-4">
            <article className="rounded-2xl border border-outline/10 bg-surface-container-low p-5">
              <div className="mb-4 flex flex-wrap items-center gap-2">
                {quickFilters.map((filter) => (
                  <button
                    key={filter.id}
                    type="button"
                    aria-pressed={quickFilter === filter.id}
                    className={`rounded-full border px-3 py-1.5 text-xs font-bold transition ${
                      quickFilter === filter.id
                        ? 'border-primary bg-primary text-white shadow-sm'
                        : 'border-outline/15 bg-white text-on-surface-variant hover:border-primary/30 hover:text-primary'
                    }`}
                    onClick={() => setQuickFilter(filter.id)}
                  >
                    {filter.label}
                  </button>
                ))}
              </div>
              <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
                <label className="space-y-2">
                  <span className="text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">
                    {t('transactionsPage.filters.searchLedger')}
                  </span>
                  <div className="relative">
                    <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-on-surface-variant" />
                    <input
                      className="h-11 w-full rounded-xl border border-outline/20 bg-white pl-9 pr-3 text-sm"
                      placeholder={t('transactionsPage.filters.searchPlaceholder')}
                      value={listSearchQuery}
                      onChange={(event) => setListSearchQuery(event.target.value)}
                    />
                  </div>
                </label>
                <label className="space-y-2">
                  <span className="text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">
                    {t('transactionsPage.filters.source')}
                  </span>
                  <select
                    className="h-11 w-full rounded-xl border border-outline/20 bg-white px-3 text-sm"
                    value={selectedAccountFilter}
                    onChange={(event) => setSelectedAccountFilter(event.target.value)}
                  >
                    <option value="">{t('transactionsPage.filters.allAccounts')}</option>
                    {accounts.map((account) => (
                      <option key={account.id} value={account.id}>
                        {account.name}
                      </option>
                    ))}
                  </select>
                </label>
                <label className="space-y-2">
                  <span className="text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">
                    {t('transactionsPage.filters.ledger')}
                  </span>
                  <select
                    className="h-11 w-full rounded-xl border border-outline/20 bg-white px-3 text-sm"
                    value={selectedLedgerFilter}
                    onChange={(event) => setSelectedLedgerFilter(event.target.value)}
                  >
                    <option value="">{t('transactionsPage.filters.allLedgers')}</option>
                    {ledgers.map((ledger) => (
                      <option key={ledger.id} value={ledger.id}>
                        {ledger.name}
                      </option>
                    ))}
                  </select>
                </label>
                <label className="space-y-2">
                  <span className="text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">
                    {t('transactionsPage.filters.dateRange')}
                  </span>
                  <select
                    className="h-11 w-full rounded-xl border border-outline/20 bg-white px-3 text-sm"
                    value={dateRangePreset}
                    onChange={(event) => setDateRangePreset(event.target.value as '7' | '30' | '120' | '365')}
                  >
                    <option value="7">{t('transactionsPage.filters.last7Days')}</option>
                    <option value="30">{t('transactionsPage.filters.last30Days')}</option>
                    <option value="120">{t('transactionsPage.filters.last120Days')}</option>
                    <option value="365">{t('transactionsPage.filters.last365Days')}</option>
                  </select>
                </label>
              </div>
            </article>

            <article className="overflow-hidden rounded-2xl border border-outline/15 bg-white">
              <div className="grid grid-cols-[1.6fr_1fr_1fr_1fr_0.75fr_1fr_0.55fr] bg-surface-container-low px-5 py-3 text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">
                <p>{t('transactionsPage.table.transactionCategory')}</p>
                <p>{t('transactionsPage.table.accountLedger')}</p>
                <p>{t('transactionsPage.table.dateTime')}</p>
                <p>{t('transactionsPage.table.note')}</p>
                <p>{t('transactionsPage.table.tags')}</p>
                <p className="text-right">{t('transactionsPage.table.amount')}</p>
                <p className="text-right">{t('transactionsPage.table.action')}</p>
              </div>
              {filteredListTransactions.map((tx) => (
                <div key={tx.id} className="grid grid-cols-[1.6fr_1fr_1fr_1fr_0.75fr_1fr_0.55fr] items-center gap-3 border-t border-outline/10 px-5 py-4">
                  <div>
                    <p className="font-semibold text-on-surface">{tx.category_name ?? t('transactionsPage.quickFilters.uncategorized')}</p>
                    <p className="text-xs text-on-surface-variant">
                      {t('transactionsPage.table.memoLabel')}{tx.memo?.trim() || t('transactionsPage.table.noMemo')}
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-on-surface">
                      {getTransactionAccountLabel(tx, accountNameById, t('transactionsPage.table.noAccount'))}
                    </p>
                    <p className="text-xs uppercase text-on-surface-variant">
                      {getTransactionLedgerLabel(tx, ledgerNameById, t('transactionsPage.table.noLedger'))}
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-on-surface">{new Date(tx.occurred_at).toLocaleDateString(locale)}</p>
                    <p className="text-xs text-on-surface-variant">{new Date(tx.occurred_at).toLocaleTimeString(locale)}</p>
                  </div>
                  <div>
                    <p className="text-xs font-semibold uppercase text-on-surface">
                      {tx.type === 'income' ? t('transaction.typeIncome') : tx.type === 'expense' ? t('transaction.typeExpense') : t('transaction.typeTransfer')}
                    </p>
                    <p className="mt-1 text-xs text-on-surface-variant">
                      {tx.memo?.trim() || (tx.type === 'income' ? t('transactionsPage.table.incomingFlow') : tx.type === 'expense' ? t('transactionsPage.table.outgoingPayment') : t('transactionsPage.table.internalTransfer'))}
                    </p>
                  </div>
                  <div>
                    <span className="rounded-full bg-surface-container-low px-2 py-1 text-[10px] font-semibold uppercase text-on-surface-variant">
                      {tx.type === 'income' ? t('transaction.typeIncome') : tx.type === 'expense' ? t('transaction.typeExpense') : t('transaction.typeTransfer')}
                    </span>
                  </div>
                  <p className={`text-right text-4xl font-extrabold ${tx.type === 'income' ? 'text-emerald-600' : 'text-rose-600'}`}>
                    {formatCurrency(Math.abs(tx.amount))}
                  </p>
                  <div className="flex justify-end">
                    <button
                      type="button"
                      aria-label={t('transactionsPage.table.deleteLabel', { name: tx.memo?.trim() || tx.category_name || tx.id })}
                      className="grid h-9 w-9 place-items-center rounded-lg border border-outline/15 text-on-surface-variant transition hover:border-rose-200 hover:bg-rose-50 hover:text-rose-700 disabled:opacity-50"
                      onClick={() => void handleDeleteTransaction(tx)}
                      disabled={deleteTransactionMutation.isPending}
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>
                </div>
              ))}
              {filteredListTransactions.length === 0 ? (
                <div className="border-t border-outline/10 px-5 py-8 text-center text-sm text-on-surface-variant">
                  <p>{t('transactionsPage.empty.noMatching')}</p>
                  {isInitialEmptyState ? (
                    <>
                      <p className="mt-2">{t('transactionsPage.empty.initial')}</p>
                      <div className="mt-4 flex flex-wrap items-center justify-center gap-2">
                        <Button className="px-3 py-1.5 text-xs" onClick={() => setShowAddDialog(true)}>
                          {t('transactionsPage.empty.createFirst')}
                        </Button>
                        <Button className="px-3 py-1.5 text-xs" variant="secondary" onClick={() => navigate('/shortcut')}>
                          {t('transactionsPage.empty.openQuickEntry')}
                        </Button>
                      </div>
                    </>
                  ) : null}
                </div>
              ) : null}
            </article>
            {pendingUndo ? (
              <div className="fixed bottom-6 left-1/2 z-30 flex w-[min(520px,calc(100vw-32px))] -translate-x-1/2 items-center justify-between gap-4 rounded-2xl border border-primary/20 bg-primary px-5 py-4 text-white shadow-ambient">
                <div>
                  <p className="text-sm font-bold">{t('transactionsPage.undo.deleted')}</p>
                  <p className="text-xs text-primary-fixed">{pendingUndo.category_name ?? pendingUndo.id}</p>
                </div>
                <button
                  type="button"
                  className="inline-flex items-center gap-2 rounded-xl bg-white px-4 py-2 text-sm font-bold text-primary transition hover:bg-primary-fixed disabled:opacity-60"
                  onClick={() => void handleUndoDelete()}
                  disabled={createTransactionMutation.isPending}
                >
                  <Undo2 className="h-4 w-4" />
                  {t('transactionsPage.undo.undo')}
                </button>
              </div>
            ) : null}
          </div>
        ) : (
          <div className="mt-6 grid gap-4 xl:grid-cols-[1.5fr_0.85fr]">
            <article className="rounded-2xl border border-outline/15 bg-white p-5">
              <div className="mb-4 flex items-center gap-3">
                <button type="button" onClick={() => setActiveMonth(new Date(activeMonth.getFullYear(), activeMonth.getMonth() - 1, 1))}>
                  <ChevronLeft className="h-4 w-4 text-primary" />
                </button>
                <h3 className="font-headline text-4xl font-bold text-on-surface">
                  {activeMonth.toLocaleString(locale, { month: 'long', year: 'numeric' })}
                </h3>
                <button type="button" onClick={() => setActiveMonth(new Date(activeMonth.getFullYear(), activeMonth.getMonth() + 1, 1))}>
                  <ChevronRight className="h-4 w-4 text-primary" />
                </button>
                <button type="button" className="ml-2 text-xs font-semibold text-primary" onClick={() => setActiveMonth(new Date())}>
                  {t('transactionsPage.calendar.today')}
                </button>
              </div>
              <div className="grid grid-cols-7 border border-outline/15">
                {weekdayLabels.map((d) => (
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
                <p className="text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">
                  {t('transactionsPage.calendar.dailySummary')}
                </p>
                <h4 className="mt-2 font-headline text-5xl font-bold text-on-surface">
                  {new Date(effectiveSelectedDay).toLocaleDateString(locale)}
                </h4>
                <div className="mt-4 grid grid-cols-2 gap-3">
                  <div className="rounded-xl bg-surface-container-low p-3">
                    <p className="text-[10px] font-bold uppercase tracking-[0.12em] text-on-surface-variant">
                      {t('transactionsPage.calendar.totalOut')}
                    </p>
                    <p className="mt-2 font-headline text-4xl font-extrabold text-rose-700">-{formatCurrency(selectedTotals.out)}</p>
                  </div>
                  <div className="rounded-xl bg-surface-container-low p-3">
                    <p className="text-[10px] font-bold uppercase tracking-[0.12em] text-on-surface-variant">
                      {t('transactionsPage.calendar.totalIn')}
                    </p>
                    <p className="mt-2 font-headline text-4xl font-extrabold text-emerald-700">+{formatCurrency(selectedTotals.in)}</p>
                  </div>
                </div>
                <div className="mt-5 space-y-3">
                  {selectedDayTx.slice(0, 6).map((tx) => (
                    <div key={tx.id} className="rounded-xl bg-surface-container-low p-3">
                      <div className="flex items-center justify-between">
                        <p className="font-semibold text-on-surface">{tx.category_name ?? t('transactionsPage.quickFilters.uncategorized')}</p>
                        <p className={tx.type === 'income' ? 'font-bold text-emerald-700' : 'font-bold text-rose-700'}>
                          {tx.type === 'income' ? '+' : '-'}
                          {formatCurrency(Math.abs(tx.amount))}
                        </p>
                      </div>
                      <p className="mt-1 text-xs text-on-surface-variant">
                        {tx.memo?.trim() || (tx.type === 'income' ? t('transaction.typeIncome') : tx.type === 'expense' ? t('transaction.typeExpense') : t('transaction.typeTransfer'))}
                      </p>
                    </div>
                  ))}
                  {selectedDayTx.length === 0 ? <p className="text-sm text-on-surface-variant">{t('transactionsPage.calendar.noTransactions')}</p> : null}
                </div>
              </article>
              <article className="rounded-2xl bg-primary p-5 text-white">
                <p className="text-[10px] font-bold uppercase tracking-[0.14em] text-primary-fixed">{t('transactionsPage.calendar.monthSummary')}</p>
                <div className="mt-3 grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-xs text-primary-fixed">{t('transactionsPage.calendar.income')}</p>
                    <p className="font-headline text-3xl font-extrabold">+{formatCurrency(monthTotals.in)}</p>
                  </div>
                  <div>
                    <p className="text-xs text-primary-fixed">{t('transactionsPage.calendar.expense')}</p>
                    <p className="font-headline text-3xl font-extrabold">-{formatCurrency(monthTotals.out)}</p>
                  </div>
                </div>
                <div className="mt-4 border-t border-white/20 pt-4">
                  <p className="text-xs text-primary-fixed">{t('transactionsPage.calendar.net')}</p>
                  <p className="font-headline text-4xl font-extrabold">{formatCurrency(monthTotals.in - monthTotals.out)}</p>
                </div>
              </article>
            </div>
          </div>
        )}
      </section>

      {showImportDialog ? (
        <DialogShell
          title={t('transactionsPage.importDialog.title')}
          description={t('transactionsPage.importDialog.description')}
          onClose={() => setShowImportDialog(false)}
          footer={
            <>
              <Button variant="ghost" onClick={() => setShowImportDialog(false)}>
                {t('common.cancel')}
              </Button>
              <Button type="submit" form="import-preview-form">
                {t('transactionsPage.importDialog.preview')}
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
              <p className="mt-4 text-3xl font-semibold text-on-surface">{t('transactionsPage.importDialog.dropTitle')}</p>
              <p className="mt-2 text-sm text-on-surface-variant">{t('transactionsPage.importDialog.supportedFormats')}</p>
              <span className="mt-6 rounded-lg border border-primary px-6 py-2 text-sm font-semibold text-primary">{t('transactionsPage.importDialog.selectFiles')}</span>
              <input id="csv-file" type="file" accept=".csv,text/csv" aria-label={t('transactionsPage.importDialog.csvFileLabel')} className="hidden" onChange={(event) => setImportFile(event.target.files?.[0] ?? null)} />
            </label>
            {importFile ? <p className="text-sm text-on-surface">{t('transactionsPage.importDialog.selected', { name: importFile.name })}</p> : null}
            {importPreviewMutation.isError ? (
              <div className="rounded-xl border border-error bg-error-container p-4">
                <p className="text-sm text-on-error-container">
                  {t('transactionsPage.importDialog.previewFailed', {
                    message: (importPreviewMutation.error as Error)?.message ?? t('common.unknownError'),
                  })}
                </p>
              </div>
            ) : null}
            {importPreviewMutation.data ? (
              <div className="rounded-xl border border-outline/10 bg-surface-container-low p-4">
                <p className="text-xs font-bold uppercase tracking-[0.12em] text-on-surface-variant">
                  {t('transactionsPage.importDialog.detectedColumns')}
                </p>
                <div className="mt-3 flex flex-wrap gap-2">
                  {importPreviewMutation.data.columns.map((column) => (
                    <span key={column} className="rounded-full bg-white px-3 py-1 text-sm text-on-surface">
                      {column}
                    </span>
                  ))}
                </div>
                <div className="mt-4 flex items-center gap-3">
                  <Button onClick={() => void handleConfirmImport()} disabled={importConfirmMutation.isPending}>
                    {importConfirmMutation.isPending ? t('transactionsPage.importDialog.importing') : t('transactionsPage.importDialog.confirm')}
                  </Button>
                  {importConfirmMutation.isError ? (
                    <span className="text-sm text-error">
                      {t('transactionsPage.importDialog.importFailed', {
                        message: (importConfirmMutation.error as Error)?.message ?? t('common.unknownError'),
                      })}
                    </span>
                  ) : null}
                  {importConfirmMutation.isSuccess ? (
                    <span className="text-sm text-emerald-600">{t('transactionsPage.importDialog.importSuccessful')}</span>
                  ) : null}
                </div>
              </div>
            ) : null}
          </form>
        </DialogShell>
      ) : null}

      {showAddDialog ? (
        <DialogShell
          title={t('transactionsPage.addDialog.title')}
          description={t('transactionsPage.addDialog.description')}
          onClose={() => setShowAddDialog(false)}
          footer={
            <>
              <Button variant="ghost" onClick={() => setShowAddDialog(false)}>
                {t('common.cancel')}
              </Button>
              <Button type="submit" form="add-transaction-form">
                {t('transactionsPage.addDialog.save')}
              </Button>
            </>
          }
        >
          <form id="add-transaction-form" className="grid gap-5 md:grid-cols-2" onSubmit={(event) => void handleCreateTransaction(event)}>
            <TextField label={t('transaction.amount')} type="number" step="0.01" value={amount} onChange={(event) => setAmount(event.target.value)} placeholder="0.00" />
            <TextField label={t('transaction.dateTime')} type="datetime-local" step="1" value={date} onChange={(event) => setDate(event.target.value)} />
            <SelectField label={t('transaction.category')} value={categoryId} onChange={(event) => setCategoryId(event.target.value)}>
              <option value="">{t('transactionsPage.addDialog.selectCategory')}</option>
              {categories.map((category) => (
                <option key={category.id} value={category.id}>
                  {category.name}
                </option>
              ))}
            </SelectField>
            <SelectField label={t('transaction.account')} value={accountId} onChange={(event) => setAccountId(event.target.value)}>
              <option value="">{t('transactionsPage.addDialog.selectAccount')}</option>
              {accounts.map((account) => (
                <option key={account.id} value={account.id}>
                  {account.name}
                </option>
              ))}
            </SelectField>
            <SelectField label={t('transactionsPage.filters.ledger')} defaultValue={ledgers[0]?.id ?? ''} disabled>
              {ledgers.map((ledger) => (
                <option key={ledger.id} value={ledger.id}>
                  {ledger.name}
                </option>
              ))}
            </SelectField>
            <TextField label={t('transaction.memo')} value={memo} onChange={(event) => setMemo(event.target.value)} placeholder={t('transactionsPage.addDialog.memoPlaceholder')} />
            <div className="md:col-span-2">
              <SelectField label={t('transaction.type')} value={transactionType} onChange={(event) => setTransactionType(event.target.value as 'income' | 'expense')}>
                <option value="expense">{t('transaction.typeExpense')}</option>
                <option value="income">{t('transaction.typeIncome')}</option>
              </SelectField>
            </div>
            <div className="md:col-span-2">
              <p className="mb-2 text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">{t('transactionsPage.table.tags')}</p>
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
