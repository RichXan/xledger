import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useAuth } from '@/features/auth/auth-context'
import type { PaginatedResponse } from '@/features/transactions/transactions-api'
import {
  createAccount,
  createCategory,
  createLedger,
  createPAT,
  deleteCategory,
  deleteLedger,
  exportCsv,
  generateShortcut,
  getAccounts,
  getCategories,
  getLedgers,
  getPATs,
  getTags,
  revokePAT,
  updateCategory,
  updateLedger,
  updateAccount,
  type CategoryItem,
} from './management-api'

type CategoryListCache = PaginatedResponse<CategoryItem>

function removeArchivedCategoryFromCache(current: CategoryListCache | undefined, categoryID: string) {
  if (!current?.items) return current
  const nextItems = current.items.filter((category) => category.id !== categoryID && !category.archived_at)
  if (nextItems.length === current.items.length) return current
  return {
    ...current,
    items: nextItems,
    pagination: {
      ...current.pagination,
      total: Math.max(0, current.pagination.total - (current.items.length - nextItems.length)),
      total_pages: Math.max(1, Math.ceil(nextItems.length / Math.max(1, current.pagination.page_size))),
    },
  }
}

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
      await queryClient.invalidateQueries({ queryKey: ['transactions', 'accounts'] })
      await queryClient.invalidateQueries({ queryKey: ['reporting', 'overview'] })
    },
  })
}

export function useCreateLedger() {
  const { session } = useAuth()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: { name: string; is_default?: boolean }) => createLedger(session!.accessToken, input),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['management', 'ledgers'] })
      await queryClient.invalidateQueries({ queryKey: ['transactions', 'ledgers'] })
    },
  })
}

export function useUpdateLedger() {
  const { session } = useAuth()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: { id: string; name: string }) => updateLedger(session!.accessToken, input.id, { name: input.name }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['management', 'ledgers'] })
      await queryClient.invalidateQueries({ queryKey: ['transactions', 'ledgers'] })
    },
  })
}

export function useDeleteLedger() {
  const { session } = useAuth()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => deleteLedger(session!.accessToken, id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['management', 'ledgers'] })
      await queryClient.invalidateQueries({ queryKey: ['transactions', 'ledgers'] })
    },
  })
}

export function useCreateCategory() {
  const { session } = useAuth()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: { name: string; parent_id?: string }) => createCategory(session!.accessToken, input),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['management', 'categories'] })
      await queryClient.invalidateQueries({ queryKey: ['transactions', 'categories'] })
    },
  })
}

export function useUpdateCategory() {
  const { session } = useAuth()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: { id: string; name?: string; archived?: boolean }) =>
      updateCategory(session!.accessToken, input.id, { name: input.name, archived: input.archived }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['management', 'categories'] })
      await queryClient.invalidateQueries({ queryKey: ['transactions', 'categories'] })
    },
  })
}

export function useDeleteCategory() {
  const { session } = useAuth()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => deleteCategory(session!.accessToken, id),
    onSuccess: async (_result, id) => {
      queryClient.setQueryData<CategoryListCache>(['management', 'categories'], (current) => removeArchivedCategoryFromCache(current, id))
      queryClient.setQueryData<CategoryListCache>(['transactions', 'categories'], (current) => removeArchivedCategoryFromCache(current, id))
      await queryClient.invalidateQueries({ queryKey: ['management', 'categories'] })
      await queryClient.invalidateQueries({ queryKey: ['transactions', 'categories'] })
    },
  })
}

export function useUpdateAccount() {
  const { session } = useAuth()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: { id: string; name?: string; type?: string }) => updateAccount(session!.accessToken, input.id, { name: input.name, type: input.type }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['management', 'accounts'] })
      await queryClient.invalidateQueries({ queryKey: ['transactions', 'accounts'] })
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

export function useGenerateShortcut() {
  const { session } = useAuth()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (params: { name?: string; expiresIn?: number }) =>
      generateShortcut(session!.accessToken, params.name, params.expiresIn),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['management', 'pats'] })
    },
  })
}
