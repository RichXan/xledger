import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useAuth } from '@/features/auth/auth-context'
import { createBudget, getBudgets, type CreateBudgetInput } from './budget-api'

export function useBudgets() {
  const { session } = useAuth()

  return useQuery({
    queryKey: ['budget', 'list'],
    queryFn: () => getBudgets(session!.accessToken),
    enabled: Boolean(session?.accessToken),
  })
}

export function useCreateBudget() {
  const { session } = useAuth()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: CreateBudgetInput) => createBudget(session!.accessToken, input),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['budget', 'list'] })
    },
  })
}
