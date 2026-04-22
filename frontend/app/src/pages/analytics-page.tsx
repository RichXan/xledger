import { useEffect, useMemo, useState } from 'react'
import { useCategoryStats, useTrendStatsRange } from '@/features/reporting/reporting-hooks'
import { formatCurrency } from '@/lib/format'

type FilterMode = 'month' | 'year'

type DonutSlice = {
  categoryName: string
  amount: number
  color: string
  dashArray: string
  dashOffset: number
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

export function AnalyticsPage() {
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
  const savingsRate = income > 0 ? ((income - expense) / income) * 100 : 0
  const savingsRateLabel = `${savingsRate >= 0 ? '+' : ''}${savingsRate.toFixed(1)}%`

  const categoryItems = useMemo(() => {
    return categoryPoints
      .map((point, index) => ({
        category_id: point.category_id || `${point.category_name}-${index}`,
        category_name: point.category_name || 'Uncategorized',
        amount: point.amount,
      }))
      .sort((a, b) => b.amount - a.amount)
      .slice(0, 6)
  }, [categoryPoints])

  const totalExpense = categoryItems.reduce((sum, item) => sum + item.amount, 0)

  const donutSlices = useMemo<DonutSlice[]>(() => {
    const palette = ['#00327d', '#7dde85', '#f38a80', '#4d5d85', '#8ea4d2', '#9ad6b4']
    if (totalExpense <= 0 || categoryItems.length === 0) {
      return []
    }
    const radius = 42
    const circumference = 2 * Math.PI * radius
    let accumulated = 0
    return categoryItems.map((category, index) => {
      const ratio = category.amount / totalExpense
      const length = circumference * ratio
      const slice = {
        categoryName: category.category_name,
        amount: category.amount,
        color: palette[index % palette.length],
        dashArray: `${length} ${Math.max(0, circumference - length)}`,
        dashOffset: -accumulated,
      }
      accumulated += length
      return slice
    })
  }, [categoryItems, totalExpense])

  const activeCategory =
    categoryItems.find((item) => item.category_name === activeCategoryName) ?? categoryItems[0] ?? null

  const barsByBucket = useMemo(() => {
    const map = new Map<string, { label: string; revenue: number; investment: number }>()
    trendPoints.forEach((point) => {
      const dt = new Date(point.bucket_start)
      const key = mode === 'year' ? `${dt.getMonth()}` : `${dt.getDate()}`
      const label = mode === 'year' ? dt.toLocaleString('en-US', { month: 'short' }).toUpperCase() : String(dt.getDate())
      const bucket = map.get(key) ?? { label, revenue: 0, investment: 0 }
      bucket.revenue += point.income
      bucket.investment += point.expense
      map.set(key, bucket)
    })

    const values = Array.from(map.values())
    if (mode === 'year') return values.slice(0, 12)
    return values.slice(0, 10)
  }, [mode, trendPoints])

  const maxCombined = Math.max(...barsByBucket.map((bar) => bar.revenue + bar.investment), 1)
  const activeBar = barsByBucket.find((bar) => bar.label === activeBarLabel) ?? null

  const yearOptions = Array.from({ length: 8 }, (_, index) => String(now.getFullYear() - index))

  useEffect(() => {
    if (categoryItems.length === 0) {
      setActiveCategoryName(null)
      return
    }
    const hasSelected = activeCategoryName
      ? categoryItems.some((item) => item.category_name === activeCategoryName)
      : false
    if (!hasSelected) {
      setActiveCategoryName(categoryItems[0].category_name)
    }
  }, [activeCategoryName, categoryItems])

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
            <h2 className="font-headline text-[56px] font-extrabold leading-none tracking-tight text-on-surface">Analytics</h2>
            <p className="mt-2 text-sm text-on-surface-variant">Comparative insight across category concentration and cashflow rhythm.</p>
          </div>

          <div className="flex items-center gap-2 rounded-xl border border-outline/15 bg-surface-container p-2">
            <select
              value={mode}
              onChange={(event) => setMode(event.target.value as FilterMode)}
              className="h-9 rounded-lg border border-outline/20 bg-white px-3 text-sm"
            >
              <option value="month">By Month</option>
              <option value="year">By Year</option>
            </select>
            <select
              value={year}
              onChange={(event) => setYear(event.target.value)}
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
            <p className="text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">Total Net Worth</p>
            <p className="mt-3 font-headline text-5xl font-extrabold text-on-surface">{formatCurrency(netWorth)}</p>
          </article>
          <article className="rounded-2xl bg-white p-5">
            <p className="text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">Income</p>
            <p className="mt-3 font-headline text-5xl font-extrabold text-on-surface">{formatCurrency(income)}</p>
          </article>
          <article className="rounded-2xl bg-white p-5">
            <p className="text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">Expense</p>
            <p className="mt-3 font-headline text-5xl font-extrabold text-on-surface">{formatCurrency(expense)}</p>
          </article>
          <article className="rounded-2xl bg-primary p-5 text-white">
            <p className="text-[10px] font-bold uppercase tracking-[0.14em] text-primary-fixed">Savings Rate</p>
            <p className="mt-3 font-headline text-5xl font-extrabold">{savingsRateLabel}</p>
            <p className="mt-2 text-xs text-primary-fixed">Net Income / Income</p>
          </article>
        </div>

        <div className="mt-6 grid gap-4 xl:grid-cols-[1fr_1.25fr]">
          <article className="rounded-2xl border border-outline/10 bg-white p-5">
            <h3 className="font-headline text-4xl font-bold leading-none text-on-surface">Expense Categories</h3>

            {categoryItems.length > 0 ? (
              <div className="mt-6 flex flex-col items-center">
                <svg viewBox="0 0 120 120" className="h-56 w-56" aria-label="Expense categories donut chart">
                  <circle cx="60" cy="60" r="42" fill="none" stroke="#e6e8eb" strokeWidth="16" />
                  {donutSlices.map((slice) => (
                    <circle
                      key={slice.categoryName}
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
                      className="cursor-pointer transition-opacity hover:opacity-85"
                    >
                      <title>{`${slice.categoryName}: ${formatCurrency(slice.amount)}`}</title>
                    </circle>
                  ))}
                  <circle cx="60" cy="60" r="30" fill="white" />
                  <text x="60" y="52" textAnchor="middle" className="fill-[#434653] text-[4px] font-bold tracking-[0.14em] uppercase">
                    Total Expense
                  </text>
                  <text x="60" y="66" textAnchor="middle" className="fill-[#191c1e] text-[7px] font-bold">
                    {formatCurrency(activeCategory?.amount ?? totalExpense)}
                  </text>
                  <text x="60" y="76" textAnchor="middle" className="fill-[#434653] text-[3.8px] font-semibold">
                    {activeCategory?.category_name ?? 'All categories'}
                  </text>
                </svg>

                <div className="mt-6 grid w-full grid-cols-2 gap-3">
                  {categoryItems.map((item) => (
                    <button
                      key={item.category_id}
                      type="button"
                      onMouseEnter={() => setActiveCategoryName(item.category_name)}
                      onFocus={() => setActiveCategoryName(item.category_name)}
                      onClick={() => setActiveCategoryName(item.category_name)}
                      className="rounded-xl bg-surface-container-low p-3 text-left transition hover:bg-surface-container"
                    >
                      <p className="text-xs font-semibold text-on-surface">{item.category_name}</p>
                      <p className="mt-1 text-sm font-bold text-on-surface">{formatCurrency(item.amount)}</p>
                    </button>
                  ))}
                </div>
              </div>
            ) : (
              <div className="mt-6 rounded-xl bg-surface-container-low p-4 text-sm text-on-surface-variant">
                No category data for selected range.
              </div>
            )}
          </article>

          <article className="rounded-2xl border border-outline/10 bg-white p-5">
            <h3 className="font-headline text-4xl font-bold leading-none text-on-surface">Revenue vs Burn Rate</h3>
            <p className="mt-2 text-sm text-on-surface-variant">
              {mode === 'year' ? `Monthly comparison for ${year}` : `Daily comparison for ${year}-${month}`}
            </p>

            {barsByBucket.length > 0 ? (
              <>
                <div className="mt-6 grid h-[340px] grid-cols-6 items-end gap-3">
                  {barsByBucket.map((point, index) => {
                    const total = point.revenue + point.investment
                    const h = Math.max(28, Math.round((total / maxCombined) * 100))
                    let revenueHeight = 0
                    let burnHeight = 0
                    if (total > 0) {
                      revenueHeight = h * (point.revenue / total)
                      burnHeight = h * (point.investment / total)
                      const minSegment = 8
                      if (point.revenue > 0 && revenueHeight < minSegment) {
                        const delta = minSegment - revenueHeight
                        revenueHeight = minSegment
                        burnHeight = Math.max(0, burnHeight - delta)
                      }
                      if (point.investment > 0 && burnHeight < minSegment) {
                        const delta = minSegment - burnHeight
                        burnHeight = minSegment
                        revenueHeight = Math.max(0, revenueHeight - delta)
                      }
                    }
                    return (
                      <div key={`${point.label}-${index}`} className="relative flex h-full flex-col justify-end gap-2">
                        <button
                          type="button"
                          className="flex h-full flex-col justify-end rounded-xl bg-surface-container-low p-1 text-left"
                          onMouseEnter={() => setActiveBarLabel(point.label)}
                          onFocus={() => setActiveBarLabel(point.label)}
                          onClick={() => setActiveBarLabel(point.label)}
                        >
                          <div className="rounded-t-sm bg-primary" style={{ height: `${revenueHeight}%` }} />
                          <div className="rounded-b-sm bg-[#7dde85]" style={{ height: `${burnHeight}%` }} />
                        </button>
                        <p className="text-center text-[10px] font-bold uppercase tracking-[0.12em] text-on-surface-variant">
                          {point.label}
                        </p>
                      </div>
                    )
                  })}
                </div>
                {activeBar ? (
                  <div className="mt-3 rounded-xl border border-outline/10 bg-surface-container-low px-3 py-2 text-sm text-on-surface">
                    <span className="font-semibold">{activeBar.label}</span>
                    <span className="ml-4">Revenue: {formatCurrency(activeBar.revenue)}</span>
                    <span className="ml-4">Burn: {formatCurrency(activeBar.investment)}</span>
                  </div>
                ) : null}
              </>
            ) : (
              <div className="mt-6 rounded-xl bg-surface-container-low p-4 text-sm text-on-surface-variant">
                No trend data for selected range.
              </div>
            )}
          </article>
        </div>
      </section>
    </div>
  )
}
