import type { PropsWithChildren } from 'react'
import { cn } from '@/lib/utils'

interface ChipProps extends PropsWithChildren {
  tone?: 'default' | 'positive' | 'negative'
}

const tones = {
  default: 'bg-surface-container-high text-on-surface',
  positive: 'bg-tertiary-fixed text-on-tertiary-fixed',
  negative: 'bg-error-container text-on-error-container',
}

export function Chip({ children, tone = 'default' }: ChipProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full px-3 py-1 font-label text-[10px] font-bold uppercase tracking-[0.14em]',
        tones[tone],
      )}
    >
      {children}
    </span>
  )
}
