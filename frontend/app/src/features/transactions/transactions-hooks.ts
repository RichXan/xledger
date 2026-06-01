import { keepPreviousData, useMutation, useQuery, useQueryClient, type QueryClient } from '@tanstack/react-query'
import { useAuth } from '@/features/auth/auth-context'
import {
  confirmImport,
  createTransaction,
  deleteTransaction,
  exportTransactions,
  getAccounts,
  getCategories,
  getImportJobStatus,
  getLedgers,
  getTags,
  getTransactionReviewItems,
  getTransactionReviewSummary,
  getTransactions,
  previewImport,
  updateTransaction,
  type CreateTransactionInput,
  type ImportConfirmRequest,
  type ExportTransactionsOptions,
  type TransactionReviewReason,
} from './transactions-api'

export function useTransactions() {
  return useTransactionsWithOptions()
}

export function useTransactionsWithOptions(options?: {
  page?: number
  pageSize?: number
  q?: string
  dateFrom?: string
  dateTo?: string
  accountId?: string
  ledgerId?: string
  amountMin?: string
  amountMax?: string
  enabled?: boolean
  preservePreviousData?: boolean
}) {
  const { session } = useAuth()

  return useQuery({
    queryKey: ['transactions', 'list', options?.page ?? 1, options?.pageSize ?? 20, options?.q ?? '', options?.dateFrom ?? '', options?.dateTo ?? '', options?.accountId ?? '', options?.ledgerId ?? '', options?.amountMin ?? '', options?.amountMax ?? ''],
    queryFn: () => getTransactions(session!.accessToken, options),
    enabled: Boolean(session?.accessToken) && options?.enabled !== false,
    placeholderData: options?.preservePreviousData ? keepPreviousData : undefined,
  })
}

export function useTransactionReviewSummary(options?: {
  q?: string
  dateFrom?: string
  dateTo?: string
  accountId?: string
  ledgerId?: string
  amountMin?: string
  amountMax?: string
}) {
  const { session } = useAuth()

  return useQuery({
    queryKey: ['transactions', 'review-summary', options?.q ?? '', options?.dateFrom ?? '', options?.dateTo ?? '', options?.accountId ?? '', options?.ledgerId ?? '', options?.amountMin ?? '', options?.amountMax ?? ''],
    queryFn: () => getTransactionReviewSummary(session!.accessToken, options),
    enabled: Boolean(session?.accessToken),
    retry: false,
    refetchOnWindowFocus: false,
  })
}

export function useTransactionReviewItems(options?: {
  page?: number
  pageSize?: number
  reason?: 'all' | TransactionReviewReason
  q?: string
  dateFrom?: string
  dateTo?: string
  accountId?: string
  ledgerId?: string
  amountMin?: string
  amountMax?: string
  enabled?: boolean
}) {
  const { session } = useAuth()

  return useQuery({
    queryKey: [
      'transactions',
      'review-items',
      options?.page ?? 1,
      options?.pageSize ?? 20,
      options?.reason ?? 'all',
      options?.q ?? '',
      options?.dateFrom ?? '',
      options?.dateTo ?? '',
      options?.accountId ?? '',
      options?.ledgerId ?? '',
      options?.amountMin ?? '',
      options?.amountMax ?? '',
    ],
    queryFn: () => getTransactionReviewItems(session!.accessToken, options),
    enabled: Boolean(session?.accessToken) && options?.enabled !== false,
    retry: false,
    refetchOnWindowFocus: false,
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
      await queryClient.invalidateQueries({ queryKey: ['transactions', 'categories'] })
      await queryClient.invalidateQueries({ queryKey: ['management', 'categories'] })
      await queryClient.invalidateQueries({ queryKey: ['transactions', 'review-summary'] })
      await queryClient.invalidateQueries({ queryKey: ['transactions', 'review-items'] })
      await queryClient.invalidateQueries({ queryKey: ['reporting', 'overview'] })
      await queryClient.invalidateQueries({ queryKey: ['reporting', 'trend'] })
      await queryClient.invalidateQueries({ queryKey: ['reporting', 'category'] })
      await queryClient.invalidateQueries({ queryKey: ['reporting', 'keywords'] })
      await queryClient.invalidateQueries({ queryKey: ['budget', 'list'] })
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
      await queryClient.invalidateQueries({ queryKey: ['transactions', 'review-summary'] })
      await queryClient.invalidateQueries({ queryKey: ['transactions', 'review-items'] })
      await queryClient.invalidateQueries({ queryKey: ['reporting', 'overview'] })
      await queryClient.invalidateQueries({ queryKey: ['reporting', 'trend'] })
      await queryClient.invalidateQueries({ queryKey: ['reporting', 'category'] })
      await queryClient.invalidateQueries({ queryKey: ['budget', 'list'] })
    },
  })
}

export function useUpdateTransaction() {
  const { session } = useAuth()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, input }: {
      id: string
      input: {
        amount: number
        category_id?: string | null
        memo?: string | null
        tag_ids?: string[]
        version?: number
      }
    }) => updateTransaction(session!.accessToken, id, input),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['transactions', 'list'] })
      await queryClient.invalidateQueries({ queryKey: ['transactions', 'review-summary'] })
      await queryClient.invalidateQueries({ queryKey: ['transactions', 'review-items'] })
      await queryClient.invalidateQueries({ queryKey: ['reporting'] })
      await queryClient.invalidateQueries({ queryKey: ['budget', 'list'] })
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

  return useMutation({
    mutationFn: ({ payload, idempotencyKey }: { payload: File | ImportConfirmRequest; idempotencyKey: string }) =>
      confirmImport(session!.accessToken, payload, idempotencyKey),
  })
}

export function useImportJobStatus(jobId: string | null) {
  const { session } = useAuth()

  return useQuery({
    queryKey: ['import', 'job', jobId],
    queryFn: () => getImportJobStatus(session!.accessToken, jobId!),
    enabled: Boolean(session?.accessToken && jobId),
    refetchInterval: (query) => (query.state.data?.status === 'running' ? 1000 : false),
  })
}

export async function invalidateImportAffectedQueries(queryClient: QueryClient) {
  await queryClient.invalidateQueries({ queryKey: ['transactions', 'list'] })
  await queryClient.invalidateQueries({ queryKey: ['transactions', 'review-summary'] })
  await queryClient.invalidateQueries({ queryKey: ['transactions', 'review-items'] })
  await queryClient.invalidateQueries({ queryKey: ['reporting', 'overview'] })
  await queryClient.invalidateQueries({ queryKey: ['reporting', 'trend'] })
  await queryClient.invalidateQueries({ queryKey: ['reporting', 'category'] })
  await queryClient.invalidateQueries({ queryKey: ['budget', 'list'] })
}
