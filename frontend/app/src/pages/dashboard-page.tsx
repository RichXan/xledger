import { useMemo, useState } from 'react'
import { useOverviewStats, useTrendStatsRange } from '@/features/reporting/reporting-hooks'
import { formatCurrency } from '@/lib/format'

const periods = ['Today', 'Week', 'Month', 'Year'] as const
type Period = (typeof periods)[number]

function sumWindow(
  points: Array<{ bucket_start: string; income: number; expense: number }>,
  anchor: Date,
  days: number,
) {
  const end = new Date(anchor)
  end.setHours(23, 59, 59, 999)
  const start = new Date(anchor)
  start.setDate(anchor.getDate() - (days - 1))
  start.setHours(0, 0, 0, 0)

  return points.reduce(
    (acc, p) => {
      const t = new Date(p.bucket_start)
      if (t >= start && t <= end) {
        acc.income += p.income
        acc.expense += p.expense
      }
      return acc
    },
    { income: 0, expense: 0 },
  )
}

function pctLabel(current: number, previous: number) {
  if (!Number.isFinite(current) || !Number.isFinite(previous) || previous === 0) return '0.0%'
  const percent = ((current - previous) / Math.abs(previous)) * 100
  const sign = percent >= 0 ? '+' : ''
  return `${sign}${percent.toFixed(1)}%`
}

export function DashboardPage() {
  const [period, setPeriod] = useState<Period>('Month')
  const overviewQuery = useOverviewStats()
  const trendQuery = useTrendStatsRange({ days: 370, granularity: 'day' })

  const overview = overviewQuery.data
  const points = trendQuery.data?.points ?? []

  const derived = useMemo(() => {
    if (points.length === 0) {
      return {
        income: 0,
        expense: 0,
        incomeDelta: '0.0%',
        expenseDelta: '0.0%',
        bars: [] as Array<{ label: string; revenue: number; profit: number }>,
      }
    }
    const sorted = [...points].sort(
      (a, b) => new Date(a.bucket_start).getTime() - new Date(b.bucket_start).getTime(),
    )
    const anchor = new Date(sorted[sorted.length - 1].bucket_start)
    const days = period === 'Today' ? 1 : period === 'Week' ? 7 : period === 'Month' ? 30 : 365

    const current = sumWindow(sorted, anchor, days)
    const prevAnchor = new Date(anchor)
    prevAnchor.setDate(anchor.getDate() - days)
    const previous = sumWindow(sorted, prevAnchor, days)

    const byMonth = new Map<string, { label: string; revenue: number; profit: number }>()
    sorted.forEach((point) => {
      const dt = new Date(point.bucket_start)
      const key = `${dt.getFullYear()}-${dt.getMonth()}`
      const existing = byMonth.get(key) ?? {
        label: dt.toLocaleString('en-US', { month: 'short' }).toUpperCase(),
        revenue: 0,
        profit: 0,
      }
      existing.revenue += point.income
      existing.profit += point.income - point.expense
      byMonth.set(key, existing)
    })
    const bars = Array.from(byMonth.values()).slice(-12)

    return {
      income: current.income,
      expense: current.expense,
      incomeDelta: pctLabel(current.income, previous.income),
      expenseDelta: pctLabel(current.expense, previous.expense),
      bars,
    }
  }, [period, points])

  const maxBar = Math.max(
    1,
    ...derived.bars.map((bar) => Math.max(Math.abs(bar.revenue), Math.abs(bar.profit))),
  )

  return (
    <div className="space-y-6">
      <section className="rounded-[28px] bg-surface-container-lowest p-6 shadow-ambient md:p-8">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <h2 className="font-headline text-[44px] font-extrabold leading-none tracking-tight text-on-surface">
              Financial Overview
            </h2>
            <p className="mt-2 text-sm text-on-surface-variant">
              Real-time precision analytics for your enterprise accounts.
            </p>
          </div>
          <div className="inline-flex rounded-xl bg-surface-container p-1">
            {periods.map((item) => (
              <button
                key={item}
                type="button"
                className={`rounded-lg px-4 py-2 text-xs font-semibold transition ${
                  period === item ? 'bg-surface-container-lowest text-primary shadow-sm' : 'text-on-surface-variant'
                }`}
                onClick={() => setPeriod(item)}
              >
                {item}
              </button>
            ))}
          </div>
        </div>

        <div className="mt-6 grid gap-4 lg:grid-cols-[1fr_1fr_1fr]">
          <article className="rounded-2xl border border-emerald-300/70 bg-white p-5">
            <div className="flex items-center justify-between text-xs font-bold uppercase tracking-[0.14em] text-on-surface-variant">
              <span>Monthly Income</span>
              <span className="rounded-full bg-emerald-100 px-2 py-1 text-[10px] text-emerald-700">{derived.incomeDelta}</span>
            </div>
            <p className="mt-4 font-headline text-4xl font-extrabold text-on-surface">
              {formatCurrency(derived.income)}
            </p>
            <div className="mt-4 h-[3px] w-2/3 rounded-full bg-emerald-500/80" />
          </article>

          <article className="rounded-2xl border border-rose-300/70 bg-white p-5">
            <div className="flex items-center justify-between text-xs font-bold uppercase tracking-[0.14em] text-on-surface-variant">
              <span>Total Expenses</span>
              <span className="rounded-full bg-rose-100 px-2 py-1 text-[10px] text-rose-700">{derived.expenseDelta}</span>
            </div>
            <p className="mt-4 font-headline text-4xl font-extrabold text-on-surface">
              {formatCurrency(derived.expense)}
            </p>
            <div className="mt-4 h-[3px] w-2/5 rounded-full bg-rose-500/80" />
          </article>

          <article className="rounded-2xl bg-primary p-5 text-white">
            <div className="text-xs font-bold uppercase tracking-[0.14em] text-primary-fixed">Total Assets</div>
            <p className="mt-4 font-headline text-4xl font-extrabold">
              {formatCurrency(overview?.total_assets ?? 0)}
            </p>
            <p className="mt-4 text-xs text-primary-fixed">Last synced 2 minutes ago</p>
          </article>
        </div>

        <article className="mt-6 rounded-2xl border border-outline/15 bg-white p-5 md:p-6">
          <div className="flex items-start justify-between gap-4">
            <div>
              <h3 className="font-headline text-2xl font-bold text-on-surface">12-Month Spending Trend</h3>
              <p className="text-sm text-on-surface-variant">Aggregated cash flow analysis across all departments.</p>
            </div>
            <div className="flex items-center gap-4 text-xs font-semibold">
              <span className="flex items-center gap-2 text-on-surface-variant">
                <span className="h-2 w-2 rounded-full bg-primary" />
                Revenue
              </span>
              <span className="flex items-center gap-2 text-on-surface-variant">
                <span className="h-2 w-2 rounded-full bg-emerald-400" />
                Profit
              </span>
            </div>
          </div>

          {derived.bars.length > 0 ? (
            <div className="mt-6 grid h-80 grid-cols-6 items-end gap-3 md:grid-cols-12">
              {derived.bars.map((bar, index) => (
                <div key={`${bar.label}-${index}`} className="flex h-full flex-col justify-end gap-2">
                  <div className="flex flex-col rounded-t-md">
                    <div
                      className="bg-surface-container-high"
                      style={{ height: `${Math.max(12, Math.round((Math.abs(bar.revenue) / maxBar) * 100))}%` }}
                    />
                    <div
                      className={`${index === derived.bars.length - 8 ? 'bg-primary' : 'bg-surface-container'}`}
                      style={{ height: `${Math.max(12, Math.round((Math.abs(bar.profit) / maxBar) * 100))}%` }}
                    />
                  </div>
                  <p className="text-center text-[10px] font-bold uppercase tracking-[0.12em] text-on-surface-variant">
                    {bar.label}
                  </p>
                </div>
              ))}
            </div>
          ) : (
            <div className="mt-6 rounded-xl bg-surface-container-low p-4 text-sm text-on-surface-variant">
              No trend data available.
            </div>
          )}
        </article>
      </section>
    </div>
  )
}
