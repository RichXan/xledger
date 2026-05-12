import { requestEnvelope } from '@/lib/api'

export interface BudgetWithUsage {
  id: string
  user_id?: string
  category_id: string
  amount: number
  period: 'monthly'
  alert_at: number
  spent: number
  remaining: number
  percent: number
  created_at?: string
  updated_at?: string
}

export interface BudgetListResponse {
  budgets: BudgetWithUsage[]
}

export interface CreateBudgetInput {
  category_id: string
  amount: number
  alert_at: number
}

export function getBudgets(accessToken: string) {
  return requestEnvelope<BudgetListResponse>('/budgets', {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

export function createBudget(accessToken: string, input: CreateBudgetInput) {
  return requestEnvelope<BudgetWithUsage>('/budgets', {
    method: 'POST',
    headers: { Authorization: `Bearer ${accessToken}` },
    body: JSON.stringify(input),
  })
}
