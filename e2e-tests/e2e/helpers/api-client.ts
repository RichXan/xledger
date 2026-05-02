export interface AuthTokenPair {
  access_token: string
  refresh_token: string
}

export interface AuthMeResponse {
  email: string
  name?: string
}

export interface PaginatedData<T> {
  items: T[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
}

export interface LedgerRecord {
  id: string
  name: string
  is_default?: boolean
}

export interface AccountRecord {
  id: string
  name: string
  type: string
  initial_balance: number
}

export interface PersonalAccessTokenRecord {
  id: string
  name: string
  expires_at?: string
}

export interface OverviewStats {
  total_assets: number
  income: number
  expense: number
  net: number
}

export interface CategoryRecord {
  id: string
  name: string
  parent_id?: string | null
  archived_at?: string | null
}

export interface TagRecord {
  id: string
  name: string
}

export interface TransactionRecord {
  id: string
  user_id: string
  ledger_id: string
  account_id?: string | null
  category_id?: string | null
  category_name?: string
  memo?: string
  from_account_id?: string | null
  to_account_id?: string | null
  transfer_pair_id?: string | null
  transfer_side?: 'from' | 'to'
  version: number
  type: 'income' | 'expense' | 'transfer'
  amount: number
  occurred_at: string
}

export interface CategoryStatsResult {
  items: Array<{
    category_name: string
    amount: number
  }>
}

export interface KeywordStatsResult {
  items: Array<{
    text: string
    amount: number
    count: number
  }>
}

interface Envelope<T> {
  code: string
  message: string
  data: T
}

export class ApiHttpError extends Error {
  readonly status: number
  readonly code: string

  constructor(status: number, code: string, message: string) {
    super(message)
    this.name = 'ApiHttpError'
    this.status = status
    this.code = code
  }
}

export class XledgerApiClient {
  private readonly baseUrl: string

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl.replace(/\/$/, '')
  }

  async registerOrLogin(input: { email: string; password: string; displayName: string }): Promise<AuthTokenPair> {
    try {
      return await this.register(input)
    } catch (error) {
      if (error instanceof ApiHttpError && error.status === 409) {
        return this.login(input.email, input.password)
      }
      throw error
    }
  }

  register(input: { email: string; password: string; displayName: string }) {
    return this.requestEnvelope<AuthTokenPair>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({
        email: input.email,
        password: input.password,
        display_name: input.displayName,
      }),
    })
  }

  login(email: string, password: string) {
    return this.requestEnvelope<AuthTokenPair>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    })
  }

  getMe(accessToken: string) {
    return this.requestEnvelope<AuthMeResponse>('/auth/me', {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    })
  }

  getOverview(accessToken: string, query?: { from?: string; to?: string }) {
    const params = new URLSearchParams()
    if (query?.from) {
      params.set('from', query.from)
    }
    if (query?.to) {
      params.set('to', query.to)
    }
    const suffix = params.toString() ? `?${params.toString()}` : ''
    return this.requestEnvelope<OverviewStats>(`/stats/overview${suffix}`, {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    })
  }

  listLedgers(accessToken: string) {
    return this.requestEnvelope<PaginatedData<LedgerRecord>>('/ledgers', {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    })
  }

  listAccounts(accessToken: string) {
    return this.requestEnvelope<PaginatedData<AccountRecord>>('/accounts', {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    })
  }

  createAccount(
    accessToken: string,
    input: { name: string; type: string; initial_balance: number },
  ) {
    return this.requestEnvelope<AccountRecord>('/accounts', {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify(input),
    })
  }

  updateAccount(
    accessToken: string,
    accountID: string,
    input: { name?: string; type?: string; archived?: boolean },
  ) {
    return this.requestEnvelope<AccountRecord>(`/accounts/${accountID}`, {
      method: 'PATCH',
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify(input),
    })
  }

  createLedger(accessToken: string, input: { name: string; is_default?: boolean }) {
    return this.requestEnvelope<LedgerRecord>('/ledgers', {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify(input),
    })
  }

  updateLedger(accessToken: string, ledgerID: string, input: { name: string }) {
    return this.requestEnvelope<LedgerRecord>(`/ledgers/${ledgerID}`, {
      method: 'PATCH',
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify(input),
    })
  }

  deleteLedger(accessToken: string, ledgerID: string) {
    return this.requestEnvelope<{ deleted: boolean }>(`/ledgers/${ledgerID}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    })
  }

  listTransactions(
    accessToken: string,
    query?: {
      page?: number
      page_size?: number
      date_from?: string
      date_to?: string
      account_id?: string
      ledger_id?: string
      category_id?: string
      tag_id?: string
    },
  ) {
    const params = new URLSearchParams()
    params.set('page', String(query?.page ?? 1))
    params.set('page_size', String(query?.page_size ?? 20))
    if (query?.date_from) params.set('date_from', query.date_from)
    if (query?.date_to) params.set('date_to', query.date_to)
    if (query?.account_id) params.set('account_id', query.account_id)
    if (query?.ledger_id) params.set('ledger_id', query.ledger_id)
    if (query?.category_id) params.set('category_id', query.category_id)
    if (query?.tag_id) params.set('tag_id', query.tag_id)
    const suffix = params.toString() ? `?${params.toString()}` : ''

    return this.requestEnvelope<PaginatedData<TransactionRecord>>(`/transactions${suffix}`, {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    })
  }

  createTransaction(
    accessToken: string,
    input: {
      ledger_id: string
      from_ledger_id?: string
      to_ledger_id?: string
      account_id?: string
      category_id?: string
      tag_ids?: string[]
      from_account_id?: string
      to_account_id?: string
      type: 'income' | 'expense' | 'transfer'
      amount: number
      memo?: string
      occurred_at?: string
    },
  ) {
    return this.requestEnvelope<TransactionRecord>('/transactions', {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify(input),
    })
  }

  updateTransaction(
    accessToken: string,
    transactionID: string,
    input: {
      amount: number
      version?: number
      category_id?: string | null
      memo?: string | null
      tag_ids?: string[]
    },
  ) {
    return this.requestEnvelope<TransactionRecord>(`/transactions/${transactionID}`, {
      method: 'PATCH',
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify(input),
    })
  }

  deleteTransaction(
    accessToken: string,
    transactionID: string,
    query?: {
      version?: number
    },
  ) {
    const params = new URLSearchParams()
    if (query?.version !== undefined) {
      params.set('version', String(query.version))
    }
    const suffix = params.toString() ? `?${params.toString()}` : ''
    return this.requestEnvelope<{ deleted: boolean }>(`/transactions/${transactionID}${suffix}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    })
  }

  listCategories(accessToken: string) {
    return this.requestEnvelope<PaginatedData<CategoryRecord>>('/categories', {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    })
  }

  createCategory(accessToken: string, input: { name: string; parent_id?: string }) {
    return this.requestEnvelope<CategoryRecord>('/categories', {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify(input),
    })
  }

  deleteCategory(accessToken: string, categoryID: string) {
    return this.requestEnvelope<{ deleted: boolean; archived?: boolean }>(`/categories/${categoryID}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    })
  }

  listTags(accessToken: string) {
    return this.requestEnvelope<PaginatedData<TagRecord>>('/tags', {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    })
  }

  createTag(accessToken: string, input: { name: string }) {
    return this.requestEnvelope<TagRecord>('/tags', {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify(input),
    })
  }

  getCategoryStats(accessToken: string, query?: { from?: string; to?: string }) {
    const params = new URLSearchParams()
    if (query?.from) params.set('from', query.from)
    if (query?.to) params.set('to', query.to)
    const suffix = params.toString() ? `?${params.toString()}` : ''
    return this.requestEnvelope<CategoryStatsResult>(`/stats/category${suffix}`, {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    })
  }

  getKeywordStats(accessToken: string, query?: { from?: string; to?: string; limit?: number }) {
    const params = new URLSearchParams()
    if (query?.from) params.set('from', query.from)
    if (query?.to) params.set('to', query.to)
    if (query?.limit) params.set('limit', String(query.limit))
    const suffix = params.toString() ? `?${params.toString()}` : ''
    return this.requestEnvelope<KeywordStatsResult>(`/stats/keywords${suffix}`, {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    })
  }

  async exportCsv(accessToken: string, query?: { from?: string; to?: string }): Promise<string> {
    const params = new URLSearchParams()
    params.set('format', 'csv')
    if (query?.from) params.set('from', query.from)
    if (query?.to) params.set('to', query.to)
    const suffix = params.toString() ? `?${params.toString()}` : ''
    return this.requestText(`/export${suffix}`, {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    })
  }

  listPersonalAccessTokens(accessToken: string) {
    return this.requestEnvelope<PaginatedData<PersonalAccessTokenRecord>>('/personal-access-tokens', {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    })
  }

  private async requestEnvelope<T>(path: string, init?: RequestInit): Promise<T> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      headers: {
        'Content-Type': 'application/json',
        ...(init?.headers ?? {}),
      },
      ...init,
    })
    const text = await response.text()
    const parsed = text ? (JSON.parse(text) as Partial<Envelope<T>>) : {}

    if (!response.ok) {
      throw new ApiHttpError(
        response.status,
        typeof parsed.code === 'string' ? parsed.code : 'UNKNOWN_ERROR',
        typeof parsed.message === 'string' ? parsed.message : `HTTP ${response.status}`,
      )
    }

    if (!('data' in parsed)) {
      throw new ApiHttpError(response.status, 'MALFORMED_RESPONSE', 'Missing data field in API response')
    }

    return parsed.data as T
  }

  private async requestText(path: string, init?: RequestInit): Promise<string> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      headers: {
        ...(init?.headers ?? {}),
      },
      ...init,
    })
    const text = await response.text()
    if (!response.ok) {
      throw new ApiHttpError(response.status, 'UNKNOWN_ERROR', `HTTP ${response.status}`)
    }
    return text
  }
}
