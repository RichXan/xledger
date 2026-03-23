import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useAuth } from '@/features/auth/auth-context'
import {
  createAccount,
  createPAT,
  exportCsv,
  getAccounts,
  getCategories,
  getLedgers,
  getPATs,
  getTags,
  revokePAT,
} from './management-api'

export function useManagementOverview() {
  const { session } = useAuth()

  const accountsQuery = useQuery({
    queryKey: ['management', 'accounts'],
    queryFn: () => getAccounts(session!.accessToken),
    enabled: Boolean(session?.accessToken),
  })

  const ledgersQuery = useQuery({
    queryKey: ['management', 'ledgers'],
    queryFn: () => getLedgers(session!.accessToken),
    enabled: Boolean(session?.accessToken),
  })

  const categoriesQuery = useQuery({
    queryKey: ['management', 'categories'],
    queryFn: () => getCategories(session!.accessToken),
    enabled: Boolean(session?.accessToken),
  })

  const tagsQuery = useQuery({
    queryKey: ['management', 'tags'],
    queryFn: () => getTags(session!.accessToken),
    enabled: Boolean(session?.accessToken),
  })

  return { accountsQuery, ledgersQuery, categoriesQuery, tagsQuery }
}

export function useCreateAccount() {
  const { session } = useAuth()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: { name: string; type: string; initial_balance: number }) => createAccount(session!.accessToken, input),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['management', 'accounts'] })
    },
  })
}

export function usePATs() {
  const { session } = useAuth()

  return useQuery({
    queryKey: ['management', 'pats'],
    queryFn: () => getPATs(session!.accessToken),
    enabled: Boolean(session?.accessToken),
  })
}

export function useCreatePAT() {
  const { session } = useAuth()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => createPAT(session!.accessToken),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['management', 'pats'] })
    },
  })
}

export function useRevokePAT() {
  const { session } = useAuth()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => revokePAT(session!.accessToken, id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['management', 'pats'] })
    },
  })
}

export function useExportCsv() {
  const { session } = useAuth()

  return useMutation({
    mutationFn: () => exportCsv(session!.accessToken),
  })
}
