import { Target } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { PageSection } from '@/components/ui/page-section'
import { SelectField } from '@/components/ui/select-field'
import { TextField } from '@/components/ui/text-field'
import { useBudgets, useCreateBudget } from '@/features/budget/budget-hooks'
import { useManagementOverview } from '@/features/management/management-hooks'
import { formatCurrency } from '@/lib/format'

function clampPercent(value: number) {
  if (!Number.isFinite(value)) return 0
  return Math.max(0, Math.min(100, Math.round(value)))
}

export function BudgetPage() {
  const { t } = useTranslation()
  const budgetsQuery = useBudgets()
  const createBudgetMutation = useCreateBudget()
  const { categoriesQuery } = useManagementOverview()
  const [categoryID, setCategoryID] = useState('')
  const [amount, setAmount] = useState('1000')
  const [alertAt, setAlertAt] = useState('80')

  const budgets = budgetsQuery.data?.budgets ?? []
  const categories = categoriesQuery.data?.items ?? []
  const activeCategories = useMemo(() => categories.filter((category) => !category.archived_at), [categories])
  const categoryNameByID = useMemo(() => new Map(categories.map((category) => [category.id, category.name])), [categories])
  const totalBudgeted = budgets.reduce((sum, budget) => sum + budget.amount, 0)
  const totalSpent = budgets.reduce((sum, budget) => sum + budget.spent, 0)
  const totalRemaining = budgets.reduce((sum, budget) => sum + Math.max(0, budget.remaining), 0)
  const portfolioPercent = totalBudgeted > 0 ? clampPercent((totalSpent / totalBudgeted) * 100) : 0

  async function handleCreateBudget(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!categoryID) return
    await createBudgetMutation.mutateAsync({
      category_id: categoryID,
      amount: Number(amount),
      alert_at: Number(alertAt),
    })
    setCategoryID('')
    setAmount('1000')
    setAlertAt('80')
  }

  return (
    <div className="space-y-6">
      <PageSection
        eyebrow={t('budgetPage.eyebrow')}
        title={t('budgetPage.title')}
        description={t('budgetPage.description')}
      >
        <div className="grid gap-4 xl:grid-cols-[1fr_0.8fr]">
          <section className="rounded-2xl border border-outline/10 bg-surface-container-low p-5">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">
                  {t('budgetPage.monthlyPlan')}
                </p>
                <p className="mt-2 text-sm text-on-surface-variant">{t('budgetPage.monthlyPlanDescription')}</p>
              </div>
              <div className="rounded-xl border border-outline/10 bg-white px-4 py-3 text-right">
                <p className="text-[10px] font-bold uppercase tracking-[0.12em] text-on-surface-variant">
                  {t('budgetPage.remaining')}
                </p>
                <p className="font-headline text-3xl font-extrabold text-on-surface">{formatCurrency(totalRemaining)}</p>
              </div>
            </div>

            <div className="mt-5 grid gap-3 sm:grid-cols-3">
              <div className="rounded-xl bg-white p-4">
                <p className="text-xs font-semibold text-on-surface-variant">{t('budgetPage.totalBudgeted')}</p>
                <p className="mt-2 font-headline text-2xl font-extrabold text-on-surface">{formatCurrency(totalBudgeted)}</p>
              </div>
              <div className="rounded-xl bg-white p-4">
                <p className="text-xs font-semibold text-on-surface-variant">{t('budgetPage.spent')}</p>
                <p className="mt-2 font-headline text-2xl font-extrabold text-rose-600">{formatCurrency(totalSpent)}</p>
              </div>
              <div className="rounded-xl bg-white p-4">
                <p className="text-xs font-semibold text-on-surface-variant">{t('budgetPage.used')}</p>
                <p className="mt-2 font-headline text-2xl font-extrabold text-primary">{portfolioPercent}%</p>
              </div>
            </div>

            <div className="mt-5 space-y-3">
              {budgets.map((budget) => {
                const percent = clampPercent(budget.percent)
                const isOver = budget.spent > budget.amount
                const isNearAlert = !isOver && budget.alert_at > 0 && percent >= budget.alert_at
                return (
                  <article key={budget.id} className="rounded-xl border border-outline/10 bg-white p-4">
                    <div className="flex flex-wrap items-start justify-between gap-3">
                      <div>
                        <p className="font-semibold text-on-surface">{categoryNameByID.get(budget.category_id) ?? t('budgetPage.unknownCategory')}</p>
                        <p className="mt-1 text-xs uppercase tracking-[0.12em] text-on-surface-variant">
                          {isOver ? t('budgetPage.overBudget') : isNearAlert ? t('budgetPage.nearAlert') : t('budgetPage.onTrack')}
                        </p>
                      </div>
                      <div className="text-right">
                        <p className="text-sm font-bold text-on-surface">{percent}%</p>
                        <p className="text-xs text-on-surface-variant">
                          {formatCurrency(budget.spent)} / {formatCurrency(budget.amount)}
                        </p>
                      </div>
                    </div>
                    <div className="mt-3 h-2 overflow-hidden rounded-full bg-surface-container">
                      <div
                        className={`h-full rounded-full ${isOver ? 'bg-rose-600' : isNearAlert ? 'bg-amber-500' : 'bg-primary'}`}
                        style={{ width: `${Math.min(100, Math.max(4, percent))}%` }}
                      />
                    </div>
                    <div className="mt-3 flex flex-wrap justify-between gap-2 text-xs text-on-surface-variant">
                      <span>{t('budgetPage.remaining')}: {formatCurrency(budget.remaining)}</span>
                      <span>{t('budgetPage.alertAt')}: {budget.alert_at}%</span>
                    </div>
                  </article>
                )
              })}
              {budgets.length === 0 ? (
                <div className="rounded-xl border border-dashed border-outline/20 bg-white p-6 text-sm text-on-surface-variant">
                  {t('budgetPage.empty')}
                </div>
              ) : null}
            </div>
          </section>

          <section className="rounded-2xl border border-outline/10 bg-white p-5">
            <div className="flex items-center gap-3">
              <span className="grid h-10 w-10 place-items-center rounded-xl bg-primary-fixed text-primary">
                <Target className="h-5 w-5" />
              </span>
              <div>
                <h2 className="font-headline text-xl font-extrabold text-on-surface">{t('budgetPage.createTitle')}</h2>
                <p className="mt-1 text-sm text-on-surface-variant">{t('budgetPage.createDescription')}</p>
              </div>
            </div>

            <form className="mt-5 grid gap-4" onSubmit={(event) => void handleCreateBudget(event)}>
              <SelectField label={t('budgetPage.category')} value={categoryID} onChange={(event) => setCategoryID(event.target.value)}>
                <option value="">{t('budgetPage.selectCategory')}</option>
                {activeCategories.map((category) => (
                  <option key={category.id} value={category.id}>
                    {category.name}
                  </option>
                ))}
              </SelectField>
              <TextField
                label={t('budgetPage.monthlyLimit')}
                type="number"
                min="1"
                step="0.01"
                value={amount}
                onChange={(event) => setAmount(event.target.value)}
              />
              <TextField
                label={t('budgetPage.alertAt')}
                type="number"
                min="1"
                max="100"
                step="1"
                value={alertAt}
                onChange={(event) => setAlertAt(event.target.value)}
              />
              <Button type="submit" disabled={!categoryID || Number(amount) <= 0 || createBudgetMutation.isPending}>
                {t('budgetPage.createBudget')}
              </Button>
            </form>
          </section>
        </div>
      </PageSection>
    </div>
  )
}
