import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useAuth } from '@/features/auth/auth-context'
import {
  createTransaction,
  getAccounts,
  getCategories,
  getLedgers,
  getTags,
  getTransactions,
  previewImport,
  type CreateTransactionInput,
} from './transactions-api'

export function useTransactions() {
  const { session } = useAuth()

  return useQuery({
    queryKey: ['transactions', 'list'],
    queryFn: () => getTransactions(session!.accessToken),
    enabled: Boolean(session?.accessToken),
  })
}

export function useTransactionFormOptions() {
  const { session } = useAuth()

  const accountsQuery = useQuery({
    queryKey: ['transactions', 'accounts'],
    queryFn: () => getAccounts(session!.accessToken),
    enabled: Boolean(session?.accessToken),
  })

  const ledgersQuery = useQuery({
    queryKey: ['transactions', 'ledgers'],
    queryFn: () => getLedgers(session!.accessToken),
    enabled: Boolean(session?.accessToken),
  })

  const categoriesQuery = useQuery({
    queryKey: ['transactions', 'categories'],
    queryFn: () => getCategories(session!.accessToken),
    enabled: Boolean(session?.accessToken),
  })

  const tagsQuery = useQuery({
    queryKey: ['transactions', 'tags'],
    queryFn: () => getTags(session!.accessToken),
    enabled: Boolean(session?.accessToken),
  })

  return {
    accountsQuery,
    ledgersQuery,
    categoriesQuery,
    tagsQuery,
  }
}

export function useCreateTransaction() {
  const { session } = useAuth()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: CreateTransactionInput) => createTransaction(session!.accessToken, input),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['transactions', 'list'] })
    },
  })
}

export function useImportPreview() {
  const { session } = useAuth()

  return useMutation({
    mutationFn: (file: File) => previewImport(session!.accessToken, file),
  })
}
