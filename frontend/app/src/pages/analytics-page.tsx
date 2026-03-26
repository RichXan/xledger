import { useMemo } from 'react'
import { useCategoryStats, useOverviewStats, useTrendStatsRange } from '@/features/reporting/reporting-hooks'
import { formatCurrency } from '@/lib/format'

export function AnalyticsPage() {
  const overviewQuery = useOverviewStats()
  const categoryQuery = useCategoryStats()
  const trendQuery = useTrendStatsRange({ days: 220, granularity: 'day' })

  const overview = overviewQuery.data
  const categories = [...(categoryQuery.data?.items ?? [])].sort((a, b) => b.amount - a.amount).slice(0, 4)
  const totalExpense = categories.reduce((sum, item) => sum + item.amount, 0)
  const trendPoints = trendQuery.data?.points ?? []

  const donutStops = useMemo(() => {
    const palette = ['#00327d', '#7dde85', '#f38a80', '#4d5d85']
    if (totalExpense <= 0) return 'conic-gradient(#dde1ea 0deg 360deg)'
    let acc = 0
    const segments = categories.map((item, idx) => {
      const start = acc
      const end = start + (item.amount / totalExpense) * 360
      acc = end
      return `${palette[idx % palette.length]} ${start.toFixed(2)}deg ${end.toFixed(2)}deg`
    })
    return `conic-gradient(${segments.join(', ')})`
  }, [categories, totalExpense])

  const barsByMonth = useMemo(() => {
    const monthMap = new Map<string, { label: string; revenue: number; investment: number }>()
    trendPoints.forEach((point) => {
      const dt = new Date(point.bucket_start)
      const key = `${dt.getFullYear()}-${dt.getMonth()}`
      const bucket = monthMap.get(key) ?? {
        label: dt.toLocaleString('en-US', { month: 'short' }).toUpperCase(),
        revenue: 0,
        investment: 0,
      }
      bucket.revenue += point.income
      bucket.investment += point.expense
      monthMap.set(key, bucket)
    })
    return Array.from(monthMap.values()).slice(-6)
  }, [trendPoints])

  const chartBars =
    barsByMonth.length > 0
      ? barsByMonth
      : [
          { label: 'JAN', revenue: 10, investment: 8 },
          { label: 'FEB', revenue: 14, investment: 11 },
          { label: 'MAR', revenue: 18, investment: 15 },
          { label: 'APR', revenue: 16, investment: 12 },
          { label: 'MAY', revenue: 22, investment: 18 },
          { label: 'JUN', revenue: 15, investment: 11 },
        ]

  return (
    <div className="space-y-6">
      <section className="rounded-[28px] bg-surface-container-lowest p-6 shadow-ambient md:p-8">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h2 className="font-headline text-[40px] font-extrabold tracking-tight text-on-surface">Analytics</h2>
            <p className="text-sm text-on-surface-variant">Comparative insight across category concentration and cashflow rhythm.</p>
          </div>
          <div className="hidden w-full max-w-sm items-center rounded-xl bg-surface-container px-3 py-2 lg:flex">
            <span className="text-sm text-on-surface-variant">Search analytics...</span>
          </div>
        </div>

        <div className="mt-6 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <article className="rounded-2xl border border-emerald-300/70 bg-white p-5">
            <p className="text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">Total Net Worth</p>
            <p className="mt-3 font-headline text-4xl font-extrabold text-on-surface">{formatCurrency(overview?.total_assets ?? 0)}</p>
            <p className="mt-2 text-xs font-semibold text-emerald-700">+12.4%</p>
          </article>
          <article className="rounded-2xl bg-white p-5">
            <p className="text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">Monthly Spend</p>
            <p className="mt-3 font-headline text-4xl font-extrabold text-on-surface">{formatCurrency(overview?.expense ?? 0)}</p>
            <p className="mt-2 text-xs font-semibold text-rose-700">+2.1%</p>
          </article>
          <article className="rounded-2xl bg-white p-5">
            <p className="text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">Active Accounts</p>
            <p className="mt-3 font-headline text-4xl font-extrabold text-on-surface">14</p>
            <p className="mt-2 text-xs text-on-surface-variant">8 Investment • 6 Operational</p>
          </article>
          <article className="rounded-2xl bg-primary p-5 text-white">
            <p className="text-[10px] font-bold uppercase tracking-[0.14em] text-primary-fixed">Liquidity Ratio</p>
            <p className="mt-3 font-headline text-4xl font-extrabold">1.84</p>
            <p className="mt-2 text-xs text-primary-fixed">Optimal range reached</p>
          </article>
        </div>

        <div className="mt-6 grid gap-4 xl:grid-cols-[1fr_1.25fr]">
          <article className="rounded-2xl bg-white p-5">
            <div className="flex items-center justify-between">
              <h3 className="font-headline text-2xl font-bold text-on-surface">Expense Categories</h3>
              <button type="button" className="text-xs font-semibold uppercase tracking-[0.12em] text-primary">Export</button>
            </div>
            <div className="mt-6 flex flex-col items-center">
              <div className="relative grid h-52 w-52 place-items-center rounded-full" style={{ background: donutStops }}>
                <div className="grid h-36 w-36 place-items-center rounded-full bg-white">
                  <div className="text-center">
                    <p className="text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">Total Expense</p>
                    <p className="mt-1 font-headline text-3xl font-extrabold text-on-surface">{formatCurrency(totalExpense)}</p>
                  </div>
                </div>
              </div>
              <div className="mt-6 grid w-full grid-cols-2 gap-3">
                {categories.map((item) => (
                  <div key={item.category_id} className="rounded-xl bg-surface-container-low p-3">
                    <p className="text-xs font-semibold text-on-surface">{item.category_name}</p>
                    <p className="mt-1 text-sm font-bold text-on-surface">{formatCurrency(item.amount)}</p>
                  </div>
                ))}
              </div>
            </div>
          </article>

          <article className="rounded-2xl bg-white p-5">
            <div className="flex items-center justify-between gap-4">
              <div>
                <h3 className="font-headline text-2xl font-bold text-on-surface">Revenue vs Burn Rate</h3>
                <p className="text-sm text-on-surface-variant">Comparative analysis for the last 6 months</p>
              </div>
              <div className="flex items-center gap-3 text-xs">
                <span className="rounded-full bg-surface-container-low px-2 py-1 font-semibold text-on-surface">Revenue</span>
                <span className="rounded-full bg-surface-container-low px-2 py-1 font-semibold text-on-surface">Investment</span>
              </div>
            </div>
            <div className="mt-6 grid h-72 grid-cols-6 items-end gap-3">
              {chartBars.map((point, index, all) => {
                const total = point.revenue + point.investment
                const max = Math.max(...all.map((bar) => bar.revenue + bar.investment), 1)
                const h = Math.max(30, Math.round((total / max) * 100))
                return (
                  <div key={`${point.label}-${index}`} className="flex h-full flex-col justify-end gap-2">
                    <div className="flex flex-col rounded-t-sm">
                      <div className="bg-primary" style={{ height: `${Math.max(20, h * 0.45)}%` }} />
                      <div className="bg-[#7dde85]" style={{ height: `${Math.max(18, h * 0.35)}%` }} />
                      <div className="bg-[#c8d2e2]" style={{ height: `${Math.max(14, h * 0.2)}%` }} />
                    </div>
                    <p className="text-center text-[10px] font-bold uppercase tracking-[0.12em] text-on-surface-variant">
                      {point.label}
                    </p>
                  </div>
                )
              })}
            </div>
          </article>
        </div>
      </section>
    </div>
  )
}
