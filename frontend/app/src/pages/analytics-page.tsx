import { PageSection } from '@/components/ui/page-section'
import { useCategoryStats, useTrendStats } from '@/features/reporting/reporting-hooks'
import { formatCurrency, formatShortDate } from '@/lib/format'

export function AnalyticsPage() {
  const categoryQuery = useCategoryStats()
  const trendQuery = useTrendStats()

  const categories = categoryQuery.data?.items ?? []
  const trendPoints = trendQuery.data?.points ?? []

  return (
    <div className="space-y-8">
      <PageSection
        eyebrow="Strategic visibility"
        title="Analytics"
        description="Category-level expense concentration paired with the most recent ledger trend buckets."
      >
        <div className="grid gap-6 xl:grid-cols-[1fr_1.1fr]">
          <article className="rounded-[28px] bg-surface-container-low p-6">
            <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">
              Category concentration
            </p>
            <div className="mt-6 space-y-4">
              {categories.map((item) => (
                <div key={item.category_id} className="rounded-2xl bg-surface-container-lowest p-4">
                  <div className="flex items-center justify-between gap-3">
                    <p className="font-medium text-on-surface">{item.category_name}</p>
                    <p className="font-headline text-lg font-bold text-on-surface">{formatCurrency(item.amount)}</p>
                  </div>
                </div>
              ))}
            </div>
          </article>

          <article className="rounded-[28px] bg-surface-container-low p-6">
            <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">
              Trend snapshots
            </p>
            <div className="mt-6 space-y-4">
              {trendPoints.map((point) => (
                <div key={point.bucket_start} className="rounded-2xl bg-surface-container-lowest p-4">
                  <div className="flex items-center justify-between gap-3">
                    <div>
                      <p className="font-medium text-on-surface">{formatShortDate(point.bucket_start)}</p>
                      <p className="mt-1 text-xs text-on-surface-variant">
                        Income {formatCurrency(point.income)} • Expense {formatCurrency(point.expense)}
                      </p>
                    </div>
                    <p className="font-headline text-lg font-bold text-primary">{formatCurrency(point.income - point.expense)}</p>
                  </div>
                </div>
              ))}
            </div>
          </article>
        </div>
      </PageSection>
    </div>
  )
}
