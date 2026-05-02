import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { useCategoryStats, useTrendStatsRange } from '@/features/reporting/reporting-hooks'
import { formatCurrency } from '@/lib/format'

type FilterMode = 'month' | 'year'

type DonutSlice = {
  categoryID: string
  categoryName: string
  amount: number
  color: string
  dashArray: string
  dashOffset: number
}

type CategoryItem = {
  category_id: string
  category_name: string
  amount: number
}

type TrendBucket = {
  key: string
  label: string
  revenue: number
  investment: number
}

const CHART_PALETTE = ['#003f8f', '#21a67a', '#e25563', '#c08725', '#5b66c9', '#67727f']

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

function buildRange(mode: FilterMode, year: number, month: number) {
  if (mode === 'year') {
    const from = startOfDay(new Date(year, 0, 1))
    const to = endOfDay(new Date(year, 11, 31))
    return { from: from.toISOString(), to: to.toISOString() }
  }
  const from = startOfDay(new Date(year, month - 1, 1))
  const to = endOfDay(new Date(year, month, 0))
  return { from: from.toISOString(), to: to.toISOString() }
}

function formatPercent(value: number) {
  return `${value.toFixed(1)}%`
}

function buildTrendBuckets(
  points: Array<{ bucket_start: string; income: number; expense: number }>,
  mode: FilterMode,
  year: number,
  month: number,
  language: string,
) {
  const locale = language === 'zh' ? 'zh-CN' : 'en-US'
  const buckets = new Map<string, TrendBucket>()

  if (mode === 'year') {
    Array.from({ length: 12 }, (_, index) => {
      const label = new Date(year, index, 1).toLocaleString(locale, { month: 'short' }).toUpperCase()
      buckets.set(String(index), { key: String(index), label, revenue: 0, investment: 0 })
      return null
    })
  } else {
    const daysInMonth = new Date(year, month, 0).getDate()
    Array.from({ length: daysInMonth }, (_, index) => {
      const day = index + 1
      buckets.set(String(day), { key: String(day), label: String(day), revenue: 0, investment: 0 })
      return null
    })
  }

  points.forEach((point) => {
    const dt = new Date(point.bucket_start)
    const key = mode === 'year' ? String(dt.getMonth()) : String(dt.getDate())
    const bucket = buckets.get(key)
    if (!bucket) return
    bucket.revenue += point.income
    bucket.investment += point.expense
  })

  return Array.from(buckets.values())
}

function buildCategoryDisplayItems(items: CategoryItem[], otherLabel: string) {
  if (items.length <= 6) return items
  const top = items.slice(0, 5)
  const otherAmount = items.slice(5).reduce((sum, item) => sum + item.amount, 0)
  return [...top, { category_id: 'other-categories', category_name: otherLabel, amount: otherAmount }]
}

export function AnalyticsPage() {
  const { t, i18n } = useTranslation()
  const navigate = useNavigate()
  const now = new Date()
  const [mode, setMode] = useState<FilterMode>('month')
  const [year, setYear] = useState(String(now.getFullYear()))
  const [month, setMonth] = useState(String(now.getMonth() + 1).padStart(2, '0'))
  const [activeCategoryName, setActiveCategoryName] = useState<string | null>(null)
  const [activeBarLabel, setActiveBarLabel] = useState<string | null>(null)

  const selectedYear = Number(year)
  const selectedMonth = Number(month)
  const range = useMemo(() => buildRange(mode, selectedYear, selectedMonth), [mode, selectedMonth, selectedYear])

  const categoryQuery = useCategoryStats({ from: range.from, to: range.to })
  const trendQuery = useTrendStatsRange({
    from: range.from,
    to: range.to,
    granularity: 'day',
  })

  const categoryPoints = categoryQuery.data?.items ?? []
  const trendPoints = trendQuery.data?.points ?? []

  const income = useMemo(() => trendPoints.reduce((sum, point) => sum + point.income, 0), [trendPoints])
  const expense = useMemo(() => trendPoints.reduce((sum, point) => sum + point.expense, 0), [trendPoints])
  const netWorth = income - expense
  const savingsRate = income > 0 ? ((income - expense) / income) * 100 : null
  const savingsRateLabel = savingsRate === null ? t('analyticsPage.noIncomeRate') : `${savingsRate >= 0 ? '+' : ''}${savingsRate.toFixed(1)}%`

  const categoryItems = useMemo<CategoryItem[]>(() => {
    return categoryPoints
      .map((point, index) => ({
        category_id: point.category_id || `${point.category_name}-${index}`,
        category_name: point.category_name || t('analyticsPage.uncategorized'),
        amount: point.amount,
      }))
      .sort((a, b) => b.amount - a.amount)
  }, [categoryPoints, t])

  const totalExpense = categoryItems.reduce((sum, item) => sum + item.amount, 0)
  const displayCategoryItems = useMemo(
    () => buildCategoryDisplayItems(categoryItems, t('analyticsPage.otherCategories')),
    [categoryItems, t],
  )

  const donutSlices = useMemo<DonutSlice[]>(() => {
    if (totalExpense <= 0 || displayCategoryItems.length === 0) {
      return []
    }
    const radius = 42
    const circumference = 2 * Math.PI * radius
    let accumulated = 0
    return displayCategoryItems.map((category, index) => {
      const ratio = category.amount / totalExpense
      const length = circumference * ratio
      const slice = {
        categoryID: category.category_id,
        categoryName: category.category_name,
        amount: category.amount,
        color: CHART_PALETTE[index % CHART_PALETTE.length],
        dashArray: `${length} ${Math.max(0, circumference - length)}`,
        dashOffset: -accumulated,
      }
      accumulated += length
      return slice
    })
  }, [displayCategoryItems, totalExpense])

  const activeCategory =
    displayCategoryItems.find((item) => item.category_name === activeCategoryName) ?? displayCategoryItems[0] ?? null

  const wordCloudItems = useMemo(() => categoryItems.slice(0, 14), [categoryItems])

  const barsByBucket = useMemo(
    () => buildTrendBuckets(trendPoints, mode, selectedYear, selectedMonth, i18n.language),
    [i18n.language, mode, selectedMonth, selectedYear, trendPoints],
  )

  const maxCombined = Math.max(...barsByBucket.map((bar) => bar.revenue + bar.investment), 1)
  const activeBar = barsByBucket.find((bar) => bar.label === activeBarLabel) ?? null
  const activeBarNet = activeBar ? activeBar.revenue - activeBar.investment : 0
  const topCategoryShare = totalExpense > 0 && displayCategoryItems[0] ? (displayCategoryItems[0].amount / totalExpense) * 100 : 0

  const yearOptions = Array.from({ length: 8 }, (_, index) => String(now.getFullYear() - index))

  useEffect(() => {
    if (displayCategoryItems.length === 0) {
      setActiveCategoryName(null)
      return
    }
    const hasSelected = activeCategoryName
      ? displayCategoryItems.some((item) => item.category_name === activeCategoryName)
      : false
    if (!hasSelected) {
      setActiveCategoryName(displayCategoryItems[0].category_name)
    }
  }, [activeCategoryName, displayCategoryItems])

  useEffect(() => {
    if (barsByBucket.length === 0) {
      setActiveBarLabel(null)
      return
    }
    const hasSelected = activeBarLabel ? barsByBucket.some((bar) => bar.label === activeBarLabel) : false
    if (!hasSelected) {
      setActiveBarLabel(barsByBucket[0].label)
    }
  }, [activeBarLabel, barsByBucket])

  return (
    <div className="space-y-5">
      <section className="rounded-[28px] border border-outline/15 bg-surface-container-lowest p-6 shadow-ambient md:p-7">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h2 className="font-headline text-[56px] font-extrabold leading-none tracking-tight text-on-surface">{t('analyticsPage.title')}</h2>
            <p className="mt-2 text-sm text-on-surface-variant">{t('analyticsPage.description')}</p>
          </div>

          <div className="flex items-center gap-2 rounded-xl border border-outline/15 bg-surface-container p-2">
            <select
              value={mode}
              onChange={(event) => setMode(event.target.value as FilterMode)}
              aria-label={t('analyticsPage.groupingModeLabel')}
              className="h-9 rounded-lg border border-outline/20 bg-white px-3 text-sm"
            >
              <option value="month">{t('analyticsPage.byMonth')}</option>
              <option value="year">{t('analyticsPage.byYear')}</option>
            </select>
            <select
              value={year}
              onChange={(event) => setYear(event.target.value)}
              aria-label={t('analyticsPage.yearLabel')}
              className="h-9 rounded-lg border border-outline/20 bg-white px-3 text-sm"
            >
              {yearOptions.map((value) => (
                <option key={value} value={value}>
                  {value}
                </option>
              ))}
            </select>
            {mode === 'month' ? (
              <select
                value={month}
                onChange={(event) => setMonth(event.target.value)}
                aria-label={t('analyticsPage.monthLabel')}
                className="h-9 rounded-lg border border-outline/20 bg-white px-3 text-sm"
              >
                {Array.from({ length: 12 }, (_, idx) => String(idx + 1).padStart(2, '0')).map((value) => (
                  <option key={value} value={value}>
                    {value}
                  </option>
                ))}
              </select>
            ) : null}
          </div>
        </div>

        <div className="mt-6 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <article className="rounded-2xl border border-[#67d79c]/70 bg-white p-5">
            <p className="text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">{t('analyticsPage.totalNetWorth')}</p>
            <p className="mt-3 font-headline text-5xl font-extrabold text-on-surface">{formatCurrency(netWorth)}</p>
          </article>
          <article className="rounded-2xl bg-white p-5">
            <p className="text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">{t('analyticsPage.income')}</p>
            <p className="mt-3 font-headline text-5xl font-extrabold text-on-surface">{formatCurrency(income)}</p>
          </article>
          <article className="rounded-2xl bg-white p-5">
            <p className="text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">{t('analyticsPage.expense')}</p>
            <p className="mt-3 font-headline text-5xl font-extrabold text-on-surface">{formatCurrency(expense)}</p>
          </article>
          <article className="rounded-2xl bg-primary p-5 text-white">
            <p className="text-[10px] font-bold uppercase tracking-[0.14em] text-primary-fixed">{t('analyticsPage.savingsRate')}</p>
            <p className="mt-3 font-headline text-5xl font-extrabold">{savingsRateLabel}</p>
            <p className="mt-2 text-xs text-primary-fixed">
              {savingsRate === null ? t('analyticsPage.noIncomeRateHint') : t('analyticsPage.savingsFormula')}
            </p>
          </article>
        </div>

        <div className="mt-6 grid gap-4 xl:grid-cols-[1.1fr_0.9fr]">
          <article className="rounded-2xl border border-outline/10 bg-white p-5">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <h3 className="font-headline text-4xl font-bold leading-none text-on-surface">{t('analyticsPage.expenseStructure')}</h3>
                <p className="mt-2 text-sm text-on-surface-variant">{t('analyticsPage.categoryShare')}</p>
              </div>
              {displayCategoryItems[0] ? (
                <div className="rounded-xl bg-surface-container-low px-4 py-3 text-right">
                  <p className="text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">{t('analyticsPage.topCategory')}</p>
                  <p className="mt-1 text-sm font-bold text-on-surface">
                    {displayCategoryItems[0].category_name} · {formatPercent(topCategoryShare)}
                  </p>
                </div>
              ) : null}
            </div>

            {categoryItems.length > 0 ? (
              <div className="mt-6 grid gap-6 lg:grid-cols-[240px_1fr]">
                <div className="flex items-center justify-center">
                  <svg viewBox="0 0 120 120" className="h-60 w-60" aria-label={t('analyticsPage.donutAria')}>
                    <circle cx="60" cy="60" r="42" fill="none" stroke="#edf0f3" strokeWidth="16" />
                    {donutSlices.map((slice) => (
                      <circle
                        key={slice.categoryID}
                        cx="60"
                        cy="60"
                        r="42"
                        fill="none"
                        stroke={slice.color}
                        strokeWidth="16"
                        strokeDasharray={slice.dashArray}
                        strokeDashoffset={slice.dashOffset}
                        transform="rotate(-90 60 60)"
                        onMouseEnter={() => setActiveCategoryName(slice.categoryName)}
                        onFocus={() => setActiveCategoryName(slice.categoryName)}
                        onClick={() => setActiveCategoryName(slice.categoryName)}
                        className={`cursor-pointer transition-opacity ${activeCategory?.category_name === slice.categoryName ? 'opacity-100' : 'opacity-75 hover:opacity-100'}`}
                      >
                        <title>{`${slice.categoryName}: ${formatCurrency(slice.amount)}`}</title>
                      </circle>
                    ))}
                    <circle cx="60" cy="60" r="30" fill="white" />
                    <text x="60" y="52" textAnchor="middle" className="fill-[#434653] text-[4px] font-bold uppercase">
                      {t('analyticsPage.totalExpense')}
                    </text>
                    <text x="60" y="66" textAnchor="middle" className="fill-[#191c1e] text-[7px] font-bold">
                      {formatCurrency(activeCategory?.amount ?? totalExpense)}
                    </text>
                    <text x="60" y="76" textAnchor="middle" className="fill-[#434653] text-[3.8px] font-semibold">
                      {activeCategory?.category_name ?? t('analyticsPage.allCategories')}
                    </text>
                  </svg>
                </div>

                <div className="space-y-3">
                  {displayCategoryItems.map((item, index) => {
                    const share = totalExpense > 0 ? (item.amount / totalExpense) * 100 : 0
                    const isActive = activeCategory?.category_name === item.category_name
                    return (
                    <button
                      key={item.category_id}
                      type="button"
                      onMouseEnter={() => setActiveCategoryName(item.category_name)}
                      onFocus={() => setActiveCategoryName(item.category_name)}
                      onClick={() => setActiveCategoryName(item.category_name)}
                        className={`w-full rounded-xl border p-3 text-left transition ${isActive ? 'border-primary/40 bg-primary/5' : 'border-outline/10 bg-surface-container-low hover:bg-surface-container'}`}
                    >
                        <div className="flex items-center justify-between gap-3">
                          <div className="flex min-w-0 items-center gap-2">
                            <span
                              className="h-2.5 w-2.5 shrink-0 rounded-full"
                              style={{ backgroundColor: CHART_PALETTE[index % CHART_PALETTE.length] }}
                            />
                            <p className="truncate text-sm font-bold text-on-surface">{item.category_name}</p>
                          </div>
                          <p className="shrink-0 text-sm font-bold text-on-surface">{formatCurrency(item.amount)}</p>
                        </div>
                        <div className="mt-2 flex items-center gap-3">
                          <div className="h-2 flex-1 overflow-hidden rounded-full bg-white">
                            <div
                              className="h-full rounded-full"
                              style={{ width: `${share}%`, backgroundColor: CHART_PALETTE[index % CHART_PALETTE.length] }}
                            />
                          </div>
                          <p className="w-12 text-right text-xs font-bold text-on-surface-variant">{formatPercent(share)}</p>
                        </div>
                    </button>
                    )
                  })}
                </div>
              </div>
            ) : (
              <div className="mt-6 rounded-xl bg-surface-container-low p-4 text-sm text-on-surface-variant">
                <p>{t('analyticsPage.noCategoryData')}</p>
                <p className="mt-2">{t('analyticsPage.categoryEmptyHint')}</p>
                <div className="mt-3">
                  <Button className="px-3 py-1.5 text-xs" onClick={() => navigate('/transactions')}>
                    {t('analyticsPage.addTransactions')}
                  </Button>
                </div>
              </div>
            )}
          </article>

          <article className="rounded-2xl border border-outline/10 bg-white p-5">
            <h3 className="font-headline text-4xl font-bold leading-none text-on-surface">{t('analyticsPage.spendingCloud')}</h3>
            <p className="mt-2 text-sm text-on-surface-variant">{t('analyticsPage.spendingCloudHint')}</p>

            {wordCloudItems.length > 0 ? (
              <div className="mt-6 flex min-h-[310px] flex-wrap content-center items-center justify-center gap-x-4 gap-y-3 rounded-2xl bg-surface-container-low p-6">
                {wordCloudItems.map((item, index) => {
                  const weight = totalExpense > 0 ? item.amount / totalExpense : 0
                  const fontSize = Math.round(14 + Math.min(1, weight * 4) * 22)
                  const isActive = activeCategory?.category_name === item.category_name
                  return (
                    <button
                      key={item.category_id}
                      type="button"
                      aria-label={`${item.category_name}: ${formatCurrency(item.amount)}`}
                      className={`font-headline font-extrabold leading-none transition hover:scale-105 ${isActive ? 'opacity-100' : 'opacity-75 hover:opacity-100'}`}
                      style={{
                        color: CHART_PALETTE[index % CHART_PALETTE.length],
                        fontSize,
                      }}
                      onMouseEnter={() => setActiveCategoryName(item.category_name)}
                      onFocus={() => setActiveCategoryName(item.category_name)}
                      onClick={() => setActiveCategoryName(item.category_name)}
                    >
                      {item.category_name}
                    </button>
                  )
                })}
              </div>
            ) : (
              <div className="mt-6 rounded-xl bg-surface-container-low p-4 text-sm text-on-surface-variant">
                <p>{t('analyticsPage.noCloudData')}</p>
                <div className="mt-3">
                  <Button className="px-3 py-1.5 text-xs" variant="secondary" onClick={() => navigate('/transactions')}>
                    {t('analyticsPage.goToTransactions')}
                  </Button>
                </div>
              </div>
            )}
          </article>
        </div>

        <article className="mt-4 rounded-2xl border border-outline/10 bg-white p-5">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <h3 className="font-headline text-4xl font-bold leading-none text-on-surface">{t('analyticsPage.cashflowRhythm')}</h3>
              <p className="mt-2 text-sm text-on-surface-variant">
                {mode === 'year'
                  ? t('analyticsPage.monthlyComparison', { year })
                  : t('analyticsPage.dailyComparison', { year, month })}
              </p>
            </div>
            {activeBar ? (
              <div className="rounded-xl bg-surface-container-low px-4 py-3 text-sm text-on-surface">
                <span className="font-bold">{activeBar.label}</span>
                <span className="ml-4">{t('analyticsPage.revenue')}: {formatCurrency(activeBar.revenue)}</span>
                <span className="ml-4">{t('analyticsPage.burn')}: {formatCurrency(activeBar.investment)}</span>
                <span className="ml-4">{t('analyticsPage.net')}: {formatCurrency(activeBarNet)}</span>
              </div>
            ) : null}
          </div>

          {barsByBucket.length > 0 ? (
            <div className="mt-6 overflow-x-auto pb-2">
              <div
                className="grid h-[320px] items-end gap-2"
                style={{
                  gridTemplateColumns: `repeat(${barsByBucket.length}, minmax(${mode === 'year' ? '58px' : '34px'}, 1fr))`,
                  minWidth: mode === 'year' ? '720px' : `${Math.max(720, barsByBucket.length * 42)}px`,
                }}
              >
                {barsByBucket.map((point) => {
                  const total = point.revenue + point.investment
                  const barHeight = total > 0 ? Math.max(10, Math.round((total / maxCombined) * 100)) : 0
                  const revenueHeight = total > 0 ? Math.round((point.revenue / total) * 100) : 0
                  const burnHeight = total > 0 ? Math.round((point.investment / total) * 100) : 0
                  const isActive = activeBar?.key === point.key
                  return (
                    <div key={point.key} className="relative flex h-full flex-col justify-end gap-2">
                      <button
                        type="button"
                        className={`flex h-full flex-col justify-end rounded-xl border bg-surface-container-low p-1 text-left transition ${isActive ? 'border-primary/50 ring-2 ring-primary/15' : 'border-transparent hover:border-outline/20'}`}
                        aria-label={t('analyticsPage.barAria', {
                          label: point.label,
                          revenue: formatCurrency(point.revenue),
                          burn: formatCurrency(point.investment),
                        })}
                        onMouseEnter={() => setActiveBarLabel(point.label)}
                        onFocus={() => setActiveBarLabel(point.label)}
                        onClick={() => setActiveBarLabel(point.label)}
                      >
                        <div className="flex w-full flex-col justify-end overflow-hidden rounded-lg" style={{ height: `${barHeight}%` }}>
                          {point.revenue > 0 ? <div className="bg-emerald-500" style={{ height: `${revenueHeight}%` }} /> : null}
                          {point.investment > 0 ? <div className="bg-rose-500" style={{ height: `${burnHeight}%` }} /> : null}
                        </div>
                      </button>
                      <p className="text-center text-[10px] font-bold uppercase text-on-surface-variant">{point.label}</p>
                    </div>
                  )
                })}
              </div>
            </div>
          ) : (
            <div className="mt-6 rounded-xl bg-surface-container-low p-4 text-sm text-on-surface-variant">
              <p>{t('analyticsPage.noTrendData')}</p>
              <p className="mt-2">{t('analyticsPage.trendEmptyHint')}</p>
              <div className="mt-3">
                <Button className="px-3 py-1.5 text-xs" variant="secondary" onClick={() => navigate('/transactions')}>
                  {t('analyticsPage.goToTransactions')}
                </Button>
              </div>
            </div>
          )}
        </article>
      </section>
    </div>
  )
}
