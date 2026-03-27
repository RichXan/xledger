import { requestEnvelope } from '@/lib/api'

export interface OverviewStats {
  total_assets: number
  income: number
  expense: number
  net: number
}

export interface TrendPoint {
  bucket_start: string
  income: number
  expense: number
}

export interface TrendStats {
  points: TrendPoint[]
}

export interface CategoryPoint {
  category_id: string
  category_name: string
  amount: number
}

export interface CategoryStats {
  items: CategoryPoint[]
}

type RangeOptions = {
  from?: string
  to?: string
  ledgerId?: string
}

export function getOverviewStats(accessToken: string, options?: RangeOptions) {
  const params = new URLSearchParams()
  if (options?.from) params.set('from', options.from)
  if (options?.to) params.set('to', options.to)
  if (options?.ledgerId) params.set('ledger_id', options.ledgerId)
  const suffix = params.toString()
  const path = suffix ? `/stats/overview?${suffix}` : '/stats/overview'

  return requestEnvelope<OverviewStats>(path, {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
}

export function getTrendStats(
  accessToken: string,
  from: string,
  to: string,
  timezone: string,
  granularity: 'day' | 'month' = 'day',
) {
  const params = new URLSearchParams({
    from,
    to,
    granularity,
    timezone,
  })

  return requestEnvelope<TrendStats>(`/stats/trend?${params.toString()}`, {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
}

export function getCategoryStats(accessToken: string, options?: Pick<RangeOptions, 'from' | 'to'>) {
  const params = new URLSearchParams()
  if (options?.from) params.set('from', options.from)
  if (options?.to) params.set('to', options.to)
  const suffix = params.toString()
  const path = suffix ? `/stats/category?${suffix}` : '/stats/category'

  return requestEnvelope<CategoryStats>(path, {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
}
