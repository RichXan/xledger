import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useAuth } from '@/features/auth/auth-context'
import {
  confirmImport,
  createTransaction,
  deleteTransaction,
  exportTransactions,
  getAccounts,
  getCategories,
  getLedgers,
  getTags,
  getTransactions,
  previewImport,
  type CreateTransactionInput,
  type ExportTransactionsOptions,
} from './transactions-api'

export function useTransactions() {
  return useTransactionsWithOptions()
}

export function useTransactionsWithOptions(options?: {
  page?: number
  pageSize?: number
  dateFrom?: string
  dateTo?: string
  accountId?: string
  ledgerId?: string
}) {
  const { session } = useAuth()

  return useQuery({
    queryKey: ['transactions', 'list', options?.page ?? 1, options?.pageSize ?? 20, options?.dateFrom ?? '', options?.dateTo ?? '', options?.accountId ?? '', options?.ledgerId ?? ''],
    queryFn: () => getTransactions(session!.accessToken, options),
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

export function useDeleteTransaction() {
  const { session } = useAuth()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => deleteTransaction(session!.accessToken, id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['transactions', 'list'] })
      await queryClient.invalidateQueries({ queryKey: ['reporting', 'overview'] })
      await queryClient.invalidateQueries({ queryKey: ['reporting', 'trend'] })
      await queryClient.invalidateQueries({ queryKey: ['reporting', 'category'] })
    },
  })
}

export function useExportTransactions() {
  const { session } = useAuth()

  return useMutation({
    mutationFn: (options?: ExportTransactionsOptions) => exportTransactions(session!.accessToken, options),
  })
}

export function useImportPreview() {
  const { session } = useAuth()

  return useMutation({
    mutationFn: (file: File) => previewImport(session!.accessToken, file),
  })
}

export function useImportConfirm() {
  const { session } = useAuth()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ file, idempotencyKey }: { file: File; idempotencyKey: string }) =>
      confirmImport(session!.accessToken, file, idempotencyKey),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['transactions', 'list'] })
      await queryClient.invalidateQueries({ queryKey: ['reporting', 'overview'] })
      await queryClient.invalidateQueries({ queryKey: ['reporting', 'trend'] })
      await queryClient.invalidateQueries({ queryKey: ['reporting', 'category'] })
    },
  })
}
