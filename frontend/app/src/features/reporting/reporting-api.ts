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

export function getOverviewStats(accessToken: string) {
  return requestEnvelope<OverviewStats>('/stats/overview', {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
}

export function getTrendStats(accessToken: string, from: string, to: string, timezone: string) {
  const params = new URLSearchParams({
    from,
    to,
    granularity: 'day',
    timezone,
  })

  return requestEnvelope<TrendStats>(`/stats/trend?${params.toString()}`, {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
}

export function getCategoryStats(accessToken: string) {
  return requestEnvelope<CategoryStats>('/stats/category', {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
}
