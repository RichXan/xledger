import { useQuery } from '@tanstack/react-query'
import { useAuth } from '@/features/auth/auth-context'
import { getCategoryStats, getOverviewStats, getTrendStats } from './reporting-api'

function getDefaultRange() {
  const end = new Date()
  const start = new Date(end)
  start.setDate(end.getDate() - 6)

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
  const { session } = useAuth()
  const range = getDefaultRange()

  return useQuery({
    queryKey: ['reporting', 'trend', range],
    queryFn: () => getTrendStats(session!.accessToken, range.from, range.to, range.timezone),
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
