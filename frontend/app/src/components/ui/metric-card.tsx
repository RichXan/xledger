import { cn } from '@/lib/utils'

type MetricCardTone = 'neutral' | 'positive' | 'negative' | 'primary'

interface MetricCardProps {
  label: string
  value: string
  delta?: string
  tone?: MetricCardTone
}

const toneStyles: Record<MetricCardTone, string> = {
  neutral: 'border-surface-container-high text-on-surface',
  positive: 'border-tertiary-fixed text-[#0d5b24]',
  negative: 'border-error text-error',
  primary: 'border-primary text-primary',
}

export function MetricCard({ label, value, delta, tone = 'neutral' }: MetricCardProps) {
  return (
    <article className={cn('rounded-[28px] border-l-4 bg-surface-container-low p-6', toneStyles[tone])}>
      <div className="flex items-start justify-between gap-4">
        <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">
          {label}
        </p>
        {delta ? (
          <span className="rounded-full bg-surface-container-high px-2 py-1 font-label text-[10px] font-bold uppercase tracking-[0.14em] text-on-surface-variant">
            {delta}
          </span>
        ) : null}
      </div>
      <p className="mt-5 font-headline text-3xl font-extrabold tracking-tight">{value}</p>
    </article>
  )
}
