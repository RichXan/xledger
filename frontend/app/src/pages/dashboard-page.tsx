import { useMemo, useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useOverviewStats, useTrendStatsRange } from '@/features/reporting/reporting-hooks'
import type { TransactionRecord } from '@/features/transactions/transactions-api'
import { useTransactionsWithOptions } from '@/features/transactions/transactions-hooks'
import { buildReviewSummary } from '@/features/transactions/review-rules'
import { formatCurrency } from '@/lib/format'

const periods = ['today', 'week', 'month', 'year'] as const

type Period = (typeof periods)[number]

type TrendBar = {
  key: string
  label: string
  income: number
  expense: number
  total: number
}

type DateRange = {
  from: string
  to: string
}

function startOfDay(date: Date) {
  const d = new Date(date)
  d.setHours(0, 0, 0, 0)
  return d
}

function endOfDay(date: Date) {
  const d = new Date(date)
  d.setHours(23, 59, 59, 999)
  return d
}

function startOfWeek(date: Date) {
  const d = startOfDay(date)
  d.setDate(d.getDate() - d.getDay())
  return d
}

function getMonthKey(date: Date) {
  return `${date.getFullYear()}-${date.getMonth()}`
}

function clampDay(year: number, month: number, day: number) {
  return Math.min(day, new Date(year, month + 1, 0).getDate())
}

function getCurrentPeriodRange(period: Period, anchor = new Date()): DateRange {
  if (period === 'today') {
    return { from: startOfDay(anchor).toISOString(), to: endOfDay(anchor).toISOString() }
  }
  if (period === 'week') {
    return { from: startOfWeek(anchor).toISOString(), to: endOfDay(anchor).toISOString() }
  }
  if (period === 'month') {
    return {
      from: startOfDay(new Date(anchor.getFullYear(), anchor.getMonth(), 1)).toISOString(),
      to: endOfDay(anchor).toISOString(),
    }
  }
  return {
    from: startOfDay(new Date(anchor.getFullYear(), 0, 1)).toISOString(),
    to: endOfDay(anchor).toISOString(),
  }
}

function getPreviousPeriodRange(period: Period, anchor = new Date()): DateRange {
  if (period === 'today') {
    const previous = new Date(anchor)
    previous.setDate(previous.getDate() - 1)
    return { from: startOfDay(previous).toISOString(), to: endOfDay(previous).toISOString() }
  }

  if (period === 'week') {
    const start = startOfWeek(anchor)
    const previousStart = new Date(start)
    previousStart.setDate(previousStart.getDate() - 7)
    const previousEnd = new Date(anchor)
    previousEnd.setDate(previousEnd.getDate() - 7)
    return { from: previousStart.toISOString(), to: endOfDay(previousEnd).toISOString() }
  }

  if (period === 'month') {
    const year = anchor.getFullYear()
    const month = anchor.getMonth() - 1
    const day = clampDay(year, month, anchor.getDate())
    return {
      from: startOfDay(new Date(year, month, 1)).toISOString(),
      to: endOfDay(new Date(year, month, day)).toISOString(),
    }
  }

  const previousYear = anchor.getFullYear() - 1
  const month = anchor.getMonth()
  const day = clampDay(previousYear, month, anchor.getDate())
  return {
    from: startOfDay(new Date(previousYear, 0, 1)).toISOString(),
    to: endOfDay(new Date(previousYear, month, day)).toISOString(),
  }
}

function pctLabel(current: number, previous: number) {
  if (!Number.isFinite(current) || !Number.isFinite(previous) || previous === 0) return '0.0%'
  const percent = ((current - previous) / Math.abs(previous)) * 100
  const sign = percent >= 0 ? '+' : ''
  return `${sign}${percent.toFixed(1)}%`
}

function buildActionCenter(transactions: TransactionRecord[]) {
  const reviewSummary = buildReviewSummary(transactions)
  const topExpense = transactions
    .filter((tx) => tx.type === 'expense')
    .sort((a, b) => Math.abs(b.amount) - Math.abs(a.amount))[0]

  return {
    reviewCount: reviewSummary.review,
    duplicateCount: reviewSummary.duplicates,
    uncategorizedCount: reviewSummary.uncategorized,
    topExpense,
  }
}

function buildLast12MonthBars(points: Array<{ bucket_start: string; income: number; expense: number }>, locale: string): TrendBar[] {
  const now = new Date()
  const monthKeys: Array<{ key: string; label: string }> = []
  for (let i = 11; i >= 0; i -= 1) {
    const d = new Date(now.getFullYear(), now.getMonth() - i, 1)
    monthKeys.push({
      key: `${d.getFullYear()}-${d.getMonth()}`,
      label: d.toLocaleString(locale, { month: 'short' }).toUpperCase(),
    })
  }

  const map = new Map<string, TrendBar>()
  monthKeys.forEach((item) => {
    map.set(item.key, { key: item.key, label: item.label, income: 0, expense: 0, total: 0 })
  })

  points.forEach((point) => {
    const dt = new Date(point.bucket_start)
    const key = getMonthKey(dt)
    const bucket = map.get(key)
    if (!bucket) return
    bucket.income += point.income
    bucket.expense += point.expense
    bucket.total = bucket.income + bucket.expense
  })

  return monthKeys.map((item) => map.get(item.key) as TrendBar)
}

export function DashboardPage() {
  const { t, i18n } = useTranslation()
  const [period, setPeriod] = useState<Period>('month')
  const [activeBarKey, setActiveBarKey] = useState<string | null>(null)
  const [nowTick, setNowTick] = useState(() => Date.now())
  const navigate = useNavigate()

  useEffect(() => {
    const isIOS = /iPad|iPhone|iPod/.test(navigator.userAgent)
    const isStandalone = window.matchMedia?.('(display-mode: standalone)').matches ?? false
    const hasDismissed = localStorage.getItem('pwa-onboarding-dismissed')
    if (isIOS && !isStandalone && !hasDismissed) {
      navigate('/pwa-onboarding')
    }
  }, [navigate])

  useEffect(() => {
    const timer = window.setInterval(() => {
      setNowTick(Date.now())
    }, 30_000)
    return () => window.clearInterval(timer)
  }, [])

  const rangeAnchor = useMemo(() => new Date(nowTick), [nowTick])
  const currentMonthKey = useMemo(() => getMonthKey(rangeAnchor), [rangeAnchor])
  const currentRange = useMemo(() => getCurrentPeriodRange(period, rangeAnchor), [period, rangeAnchor])
  const previousRange = useMemo(() => getPreviousPeriodRange(period, rangeAnchor), [period, rangeAnchor])

  const currentOverviewQuery = useOverviewStats({ from: currentRange.from, to: currentRange.to })
  const previousOverviewQuery = useOverviewStats({ from: previousRange.from, to: previousRange.to })
  const totalOverviewQuery = useOverviewStats()
  const trend12MonthsQuery = useTrendStatsRange({ days: 365, granularity: 'month' })
  const reviewTransactionsQuery = useTransactionsWithOptions({
    page: 1,
    pageSize: 12,
    dateFrom: currentRange.from,
    dateTo: currentRange.to,
  })

  const currentOverview = currentOverviewQuery.data
  const previousOverview = previousOverviewQuery.data
  const reviewTransactions = reviewTransactionsQuery.data?.items ?? []
  const locale = i18n.language === 'zh' ? 'zh-CN' : 'en-US'
  const bars = useMemo(() => buildLast12MonthBars(trend12MonthsQuery.data?.points ?? [], locale), [locale, trend12MonthsQuery.data?.points])
  const actionCenter = useMemo(() => buildActionCenter(reviewTransactions), [reviewTransactions])

  const derived = useMemo(() => {
    const currentIncome = currentOverview?.income ?? totalOverviewQuery.data?.income ?? 0
    const currentExpense = currentOverview?.expense ?? totalOverviewQuery.data?.expense ?? 0
    const previousIncome = previousOverview?.income ?? 0
    const previousExpense = previousOverview?.expense ?? 0
    return {
      income: currentIncome,
      expense: currentExpense,
      net: currentIncome - currentExpense,
      incomeDelta: pctLabel(currentIncome, previousIncome),
      expenseDelta: pctLabel(currentExpense, previousExpense),
    }
  }, [currentOverview, previousOverview, totalOverviewQuery.data?.expense, totalOverviewQuery.data?.income])

  useEffect(() => {
    if (bars.length === 0) {
      setActiveBarKey(null)
      return
    }
    const currentMonthBar = bars.find((bar) => bar.key === currentMonthKey)
    const hasSelected = activeBarKey ? bars.some((bar) => bar.key === activeBarKey) : false
    if (!hasSelected) {
      setActiveBarKey(currentMonthBar?.key ?? bars[bars.length - 1].key)
    }
  }, [activeBarKey, bars, currentMonthKey])

  const maxTotal = Math.max(1, ...bars.map((bar) => bar.total))
  const activeBar = bars.find((bar) => bar.key === activeBarKey) ?? bars[bars.length - 1] ?? null
  const syncedMinutesAgo = totalOverviewQuery.dataUpdatedAt
    ? Math.max(0, Math.floor((nowTick - totalOverviewQuery.dataUpdatedAt) / 60_000))
    : null
  const syncedLabel =
    syncedMinutesAgo === null
      ? t('common.loading')
      : syncedMinutesAgo === 0
        ? t('dashboard.justNow')
        : `${syncedMinutesAgo} ${t('dashboard.minutesAgo')}`

  return (
    <div className="space-y-5">
      <section className="rounded-2xl border border-outline/15 bg-surface-container-lowest p-5 shadow-ambient md:p-6">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <h2 className="font-headline text-4xl font-extrabold leading-tight text-on-surface md:text-[44px]">
              {t('dashboard.title')}
            </h2>
            <p className="mt-2 text-sm text-on-surface-variant">{t('dashboard.subtitle')}</p>
          </div>
          <div className="inline-flex rounded-xl border border-outline/15 bg-surface-container p-1">
            {periods.map((item) => (
              <button
                key={item}
                type="button"
                className={`min-h-9 rounded-lg px-4 py-2 text-xs font-semibold transition ${
                  period === item ? 'bg-white text-primary shadow-sm' : 'text-on-surface-variant hover:text-primary'
                }`}
                onClick={() => setPeriod(item)}
              >
                {t(`dashboard.periods.${item}`)}
              </button>
            ))}
          </div>
        </div>

        <div className="mt-6 grid gap-4 xl:grid-cols-[1fr_1fr_0.95fr]">
          <article className="rounded-2xl border border-[#67d79c]/70 bg-white p-5">
            <div className="flex items-center justify-between text-xs font-bold uppercase tracking-[0.14em] text-on-surface-variant">
              <span>{t(`dashboard.periods.${period}`)} {t('dashboard.income')}</span>
              <span className="rounded-full bg-emerald-100 px-2 py-1 text-[10px] text-emerald-700">
                {derived.incomeDelta}
              </span>
            </div>
            <p className="mt-4 font-headline text-4xl font-extrabold text-on-surface">{formatCurrency(derived.income)}</p>
            <div className="mt-5 h-[4px] w-[62%] rounded-full bg-emerald-500/90" />
          </article>

          <article className="rounded-2xl border border-[#f3a0a8]/70 bg-white p-5">
            <div className="flex items-center justify-between text-xs font-bold uppercase tracking-[0.14em] text-on-surface-variant">
              <span>{t(`dashboard.periods.${period}`)} {t('dashboard.expense')}</span>
              <span className="rounded-full bg-rose-100 px-2 py-1 text-[10px] text-rose-700">
                {derived.expenseDelta}
              </span>
            </div>
            <p className="mt-4 font-headline text-4xl font-extrabold text-on-surface">
              {formatCurrency(derived.expense)}
            </p>
            <div className="mt-5 h-[4px] w-[46%] rounded-full bg-rose-500/90" />
          </article>

          <article className="overflow-hidden rounded-2xl bg-primary p-5 text-white">
            <div className="text-xs font-bold uppercase tracking-[0.14em] text-primary-fixed">{t('dashboard.totalAssets')}</div>
            <p className="mt-4 font-headline text-4xl font-extrabold">
              {formatCurrency(totalOverviewQuery.data?.total_assets ?? 0)}
            </p>
            <p className="mt-3 text-sm text-primary-fixed">
              {t('dashboard.net')} <span>{formatCurrency(derived.net)}</span>
            </p>
            <p className="mt-4 text-xs text-primary-fixed">{t('dashboard.lastSynced')} {syncedLabel}</p>
          </article>
        </div>

        <article className="mt-6 rounded-2xl border border-outline/15 bg-white p-5">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <h3 className="font-headline text-2xl font-bold leading-tight text-on-surface">{t('dashboard.actionCenter.title')}</h3>
              <p className="mt-1 text-sm text-on-surface-variant">{t('dashboard.actionCenter.description')}</p>
            </div>
            <button
              type="button"
              className="min-h-10 rounded-xl border border-primary/20 bg-primary-fixed px-4 py-2 text-sm font-bold text-primary transition hover:bg-primary-fixed/80"
              onClick={() => navigate('/transactions?smart=review')}
            >
              {t('dashboard.actionCenter.reviewButton')}
            </button>
          </div>

          <div className="mt-4 grid gap-3 md:grid-cols-3">
            <div className="rounded-xl bg-surface-container-low p-4">
              <p className="text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">
                {t('dashboard.actionCenter.reviewLabel')}
              </p>
              <p className="mt-2 text-lg font-extrabold text-on-surface">
                {t('dashboard.actionCenter.reviewCount', { count: actionCenter.reviewCount })}
              </p>
              <p className="mt-1 text-xs text-on-surface-variant">
                {t('dashboard.actionCenter.reviewHint', {
                  uncategorized: actionCenter.uncategorizedCount,
                  duplicates: actionCenter.duplicateCount,
                })}
              </p>
            </div>
            <div className="rounded-xl bg-surface-container-low p-4">
              <p className="text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">
                {t('dashboard.actionCenter.topExpense')}
              </p>
              <p className="mt-2 truncate text-lg font-extrabold text-on-surface">
                {actionCenter.topExpense?.category_name ?? actionCenter.topExpense?.memo ?? t('transactionsPage.quickFilters.uncategorized')}
              </p>
              <p className="mt-1 text-xs font-bold text-rose-600">
                {formatCurrency(Math.abs(actionCenter.topExpense?.amount ?? 0))}
              </p>
            </div>
            <div className="rounded-xl bg-primary p-4 text-white">
              <p className="text-[10px] font-bold uppercase tracking-[0.14em] text-primary-fixed">
                {t('dashboard.actionCenter.nextStep')}
              </p>
              <p className="mt-2 text-sm font-semibold">
                {actionCenter.reviewCount > 0
                  ? t('dashboard.actionCenter.nextStepReview')
                  : t('dashboard.actionCenter.nextStepClear')}
              </p>
            </div>
          </div>
        </article>

        <article className="mt-6 rounded-2xl border border-outline/15 bg-white p-5 md:p-6">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div>
              <h3 className="font-headline text-3xl font-bold leading-tight text-on-surface">{t('analytics.trend')}</h3>
              <p className="mt-2 text-sm text-on-surface-variant">{t('dashboard.trendHint')}</p>
            </div>
            <div className="mt-1 flex items-center gap-4 text-xs font-semibold">
              <span className="flex items-center gap-2 text-on-surface-variant">
                <span className="h-2 w-2 rounded-full bg-primary" />
                {t('dashboard.income')}
              </span>
              <span className="flex items-center gap-2 text-on-surface-variant">
                <span className="h-2 w-2 rounded-full bg-rose-500" />
                {t('dashboard.expense')}
              </span>
            </div>
          </div>

          <div className="mt-8 overflow-x-auto pb-2">
            <div className="grid h-[360px] min-w-[720px] grid-cols-12 items-end gap-3">
              {bars.map((bar) => {
              const incomeHeight = bar.total > 0 ? Math.round((bar.income / maxTotal) * 100) : 0
              const expenseHeight = bar.total > 0 ? Math.round((bar.expense / maxTotal) * 100) : 0
              return (
                <div key={bar.key} className="flex h-full flex-col justify-end gap-2">
                  <button
                    type="button"
                    className={`relative h-full rounded-xl border p-1 text-left transition ${
                      activeBar?.key === bar.key
                        ? 'border-primary/50 bg-surface-container'
                        : 'border-outline/10 bg-surface-container-low'
                    }`}
                    onMouseEnter={() => setActiveBarKey(bar.key)}
                    onFocus={() => setActiveBarKey(bar.key)}
                    onClick={() => setActiveBarKey(bar.key)}
                    title={`${bar.label} • ${t('dashboard.income')} ${formatCurrency(bar.income)} • ${t('dashboard.expense')} ${formatCurrency(bar.expense)}`}
                  >
                    <div
                      className="absolute inset-x-1 bottom-1 rounded-b-md bg-rose-500/90"
                      style={{ height: `${Math.max(0, expenseHeight)}%` }}
                    />
                    <div
                      className="absolute inset-x-1 rounded-t-md bg-primary"
                      style={{
                        height: `${Math.max(0, incomeHeight)}%`,
                        bottom: `${Math.max(0, expenseHeight)}%`,
                      }}
                    />
                  </button>
                  <p className="text-center text-[10px] font-bold uppercase tracking-[0.12em] text-on-surface-variant">
                    {bar.label}
                  </p>
                </div>
              )
              })}
            </div>
          </div>

          {activeBar ? (
            <div className="mt-4 rounded-xl border border-outline/15 bg-surface-container-low px-4 py-3 text-sm text-on-surface">
              <span className="font-semibold">{activeBar.label}</span>
              <span className="ml-4 text-primary">{t('dashboard.income')}: {formatCurrency(activeBar.income)}</span>
              <span className="ml-4 text-rose-600">{t('dashboard.expense')}: {formatCurrency(activeBar.expense)}</span>
            </div>
          ) : null}

          {trend12MonthsQuery.isError || currentOverviewQuery.isError || previousOverviewQuery.isError ? (
            <div className="mt-4 rounded-xl bg-rose-50 px-4 py-3 text-sm text-rose-700">
              {t('errors.serverError')}
            </div>
          ) : null}
        </article>
      </section>
    </div>
  )
}
