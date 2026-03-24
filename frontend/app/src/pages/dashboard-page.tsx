import { Button } from '@/components/ui/button'
import { MetricCard } from '@/components/ui/metric-card'
import { useTrendStats, useOverviewStats } from '@/features/reporting/reporting-hooks'
import { formatCurrency, formatShortDate } from '@/lib/format'

export function DashboardPage() {
  const overviewQuery = useOverviewStats()
  const trendQuery = useTrendStats()

  const overview = overviewQuery.data
  const trendPoints = trendQuery.data?.points ?? []

  return (
    <div className="space-y-8">
      <section className="grid gap-4 xl:grid-cols-[1.2fr_0.8fr]">
        <div className="grid gap-4 md:grid-cols-3">
          <MetricCard
            label="Monthly Income"
            value={formatCurrency(overview?.income ?? 0)}
            tone="positive"
            delta="+12.4%"
          />
          <MetricCard
            label="Total Expenses"
            value={formatCurrency(overview?.expense ?? 0)}
            tone="negative"
            delta="-4.2%"
          />
          <MetricCard
            label="Net Position"
            value={formatCurrency(overview?.net ?? 0)}
            tone="primary"
            delta="Balanced"
          />
        </div>

        <article className="rounded-[32px] bg-primary-gradient p-8 text-white shadow-ambient">
          <p className="font-label text-[10px] uppercase tracking-[0.2em] text-primary-fixed">Total Assets</p>
          <p className="mt-6 font-headline text-5xl font-extrabold tracking-tight">
            {formatCurrency(overview?.total_assets ?? 0)}
          </p>
          <p className="mt-3 max-w-sm text-sm text-primary-fixed">
            Unified balance across all connected accounts. Unlinked transactions still contribute to ledger performance.
          </p>
        </article>
      </section>

      <section className="grid gap-6 xl:grid-cols-[1.4fr_0.8fr]">
        <article className="rounded-[32px] bg-surface-container-lowest p-8 shadow-ambient">
          <div className="flex items-center justify-between gap-4">
            <div>
              <h2 className="font-headline text-2xl font-bold text-on-surface">Financial Overview</h2>
              <p className="mt-2 text-sm text-on-surface-variant">A 12-month trend surface for your ledgers and account-backed balances.</p>
            </div>
            <Button>Add Transaction</Button>
          </div>
          <div className="mt-10 grid min-h-72 grid-cols-2 items-end gap-3 md:grid-cols-7">
            {trendPoints.map((point) => {
              const total = point.income + point.expense
              const height = Math.max(36, Math.min(180, total / 10))
              return (
                <div key={point.bucket_start} className="flex flex-col items-center gap-3">
                  <div className="w-full rounded-3xl bg-surface-container-low p-3">
                    <div className="rounded-t-2xl bg-primary" style={{ height }} />
                  </div>
                  <span className="font-label text-[10px] uppercase tracking-[0.18em] text-on-surface-variant">
                    {formatShortDate(point.bucket_start)}
                  </span>
                </div>
              )
            })}
          </div>
        </article>

        <article className="rounded-[32px] bg-surface-container-low p-8">
          <p className="font-label text-[10px] uppercase tracking-[0.2em] text-on-surface-variant">Recent activity</p>
          <div className="mt-6 space-y-4">
            {trendPoints.slice(-3).map((point) => (
              <div key={point.bucket_start} className="rounded-2xl bg-surface-container-lowest p-4">
                <div className="flex items-center justify-between gap-3">
                  <div>
                    <p className="font-medium text-on-surface">{formatShortDate(point.bucket_start)}</p>
                    <p className="mt-1 text-xs text-on-surface-variant">Income {formatCurrency(point.income)} • Expense {formatCurrency(point.expense)}</p>
                  </div>
                  <p className="font-headline text-lg font-bold text-on-surface">{formatCurrency(point.income - point.expense)}</p>
                </div>
              </div>
            ))}
          </div>
        </article>
      </section>
    </div>
  )
}
