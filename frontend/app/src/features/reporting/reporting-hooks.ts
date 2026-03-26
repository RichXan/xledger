import { useQuery } from '@tanstack/react-query'
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

export function useOverviewStats() {
  const { session } = useAuth()

  return useQuery({
    queryKey: ['reporting', 'overview'],
    queryFn: () => getOverviewStats(session!.accessToken),
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
  const fallbackRange = getDefaultRange(options?.days ?? 365)
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

export function useCategoryStats() {
  const { session } = useAuth()

  return useQuery({
    queryKey: ['reporting', 'category'],
    queryFn: () => getCategoryStats(session!.accessToken),
    enabled: Boolean(session?.accessToken),
  })
}
