import { useQuery } from '@tanstack/react-query'
import { useMemo } from 'react'
import { useAuth } from '@/features/auth/auth-context'
import { getCategoryStats, getOverviewStats, getTrendStats } from './reporting-api'

function getDefaultRange(days = 365) {
  const end = new Date()
  const start = new Date(end)
  start.setDate(end.getDate() - days)

  return {
    from: start.toISOString(),
    to: end.toISOString(),
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC',
  }
}

type RangeOptions = {
  from?: string
  to?: string
}

export function useOverviewStats(options?: RangeOptions & { ledgerId?: string }) {
  const { session } = useAuth()
  const from = options?.from ?? ''
  const to = options?.to ?? ''
  const ledgerId = options?.ledgerId ?? ''

  return useQuery({
    queryKey: ['reporting', 'overview', from, to, ledgerId],
    queryFn: () => getOverviewStats(session!.accessToken, { from: options?.from, to: options?.to, ledgerId: options?.ledgerId }),
    enabled: Boolean(session?.accessToken),
  })
}

export function useTrendStats() {
  return useTrendStatsRange()
}

export function useTrendStatsRange(options?: {
  days?: number
  from?: string
  to?: string
  granularity?: 'day' | 'month'
}) {
  const { session } = useAuth()
  const fallbackRange = useMemo(() => getDefaultRange(options?.days ?? 365), [options?.days])
  const from = options?.from ?? fallbackRange.from
  const to = options?.to ?? fallbackRange.to
  const timezone = fallbackRange.timezone
  const granularity = options?.granularity ?? 'day'

  return useQuery({
    queryKey: ['reporting', 'trend', from, to, timezone, granularity],
    queryFn: () => getTrendStats(session!.accessToken, from, to, timezone, granularity),
    enabled: Boolean(session?.accessToken),
  })
}

export function useCategoryStats(options?: RangeOptions) {
  const { session } = useAuth()
  const from = options?.from ?? ''
  const to = options?.to ?? ''

  return useQuery({
    queryKey: ['reporting', 'category', from, to],
    queryFn: () => getCategoryStats(session!.accessToken, { from: options?.from, to: options?.to }),
    enabled: Boolean(session?.accessToken),
  })
}
