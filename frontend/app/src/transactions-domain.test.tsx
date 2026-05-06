import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import App from './App'
import { AuthProvider } from './features/auth/auth-context'

const originalFetch = global.fetch

function renderTransactionsApp(initialEntries: string[]) {
  window.localStorage.setItem(
    'xledger.auth',
    JSON.stringify({
      accessToken: 'access.demo.token',
      refreshToken: 'refresh.demo.token',
      email: 'demo@example.com',
    }),
  )

  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })

  return render(
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <MemoryRouter initialEntries={initialEntries}>
          <App />
        </MemoryRouter>
      </AuthProvider>
    </QueryClientProvider>,
  )
}

function createTransactionsFetchMock() {
  return vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input)

    if (url.endsWith('/api/auth/me')) {
      return new Response(
        JSON.stringify({ code: 'OK', message: '成功', data: { email: 'demo@example.com' } }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    }

    if (url.includes('/api/transactions?')) {
      return new Response(
        JSON.stringify({
          code: 'OK',
          message: '成功',
          data: {
            items: [
              {
                id: 'txn-1',
                type: 'expense',
                amount: 25,
                category_name: 'Food',
                occurred_at: '2026-03-01T08:30:00Z',
                memo: 'Lunch',
              },
              {
                id: 'txn-2',
                type: 'income',
                amount: 15000,
                category_name: 'Salary',
                occurred_at: '2026-03-02T09:00:00Z',
              },
              {
                id: 'txn-3',
                type: 'expense',
                amount: 88,
                occurred_at: '2026-03-03T10:30:00Z',
                memo: 'Needs classification',
              },
              {
                id: 'txn-4',
                type: 'expense',
                amount: 1250,
                category_name: 'Travel',
                occurred_at: '2026-03-04T18:20:00Z',
                memo: 'Conference flight',
              },
              {
                id: 'txn-5',
                type: 'expense',
                amount: 25,
                category_name: 'Cafe',
                occurred_at: '2026-03-01T09:00:00Z',
                memo: 'Lunch',
              },
            ],
            pagination: { page: 1, page_size: 20, total: 5, total_pages: 1 },
          },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    }

    if (url.includes('/api/transactions/review-summary')) {
      return new Response(
        JSON.stringify({
          code: 'OK',
          message: 'Success',
          data: { review: 4, uncategorized: 1, duplicates: 1, large: 1 },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    }

    if (url.includes('/api/transactions/review-items')) {
      return new Response(
        JSON.stringify({
          code: 'OK',
          message: 'Success',
          data: {
            items: [
              {
                transaction: {
                  id: 'txn-3',
                  type: 'expense',
                  amount: 88,
                  occurred_at: '2026-03-03T10:30:00Z',
                  memo: 'Needs classification',
                },
                reasons: ['uncategorized'],
              },
              {
                transaction: {
                  id: 'txn-4',
                  type: 'expense',
                  amount: 1250,
                  category_name: 'Travel',
                  occurred_at: '2026-03-04T18:20:00Z',
                  memo: 'Conference flight',
                },
                reasons: ['large'],
              },
              {
                transaction: {
                  id: 'txn-5',
                  type: 'expense',
                  amount: 25,
                  category_name: 'Cafe',
                  occurred_at: '2026-03-01T09:00:00Z',
                  memo: 'Lunch',
                },
                reasons: ['duplicate'],
              },
            ],
            pagination: { page: 1, page_size: 200, total: 3, total_pages: 1 },
          },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    }

    if (url.endsWith('/api/accounts')) {
      return new Response(
        JSON.stringify({
          code: 'OK',
          message: '成功',
          data: {
            items: [
              { id: 'acc-1', name: 'Cash Wallet', type: 'cash', initial_balance: 1000 },
              { id: 'acc-2', name: 'Bank of Tests', type: 'bank', initial_balance: 5000 },
            ],
            pagination: { page: 1, page_size: 2, total: 2, total_pages: 1 },
          },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    }

    if (url.endsWith('/api/ledgers')) {
      return new Response(
        JSON.stringify({
          code: 'OK',
          message: '成功',
          data: {
            items: [{ id: 'ledger-1', name: 'Default Ledger', is_default: true }],
            pagination: { page: 1, page_size: 1, total: 1, total_pages: 1 },
          },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    }

    if (url.endsWith('/api/categories')) {
      return new Response(
        JSON.stringify({
          code: 'OK',
          message: '成功',
          data: {
            items: [{ id: 'cat-1', name: 'Food' }, { id: 'cat-2', name: 'Salary' }],
            pagination: { page: 1, page_size: 2, total: 2, total_pages: 1 },
          },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    }

    if (url.endsWith('/api/tags')) {
      return new Response(
        JSON.stringify({
          code: 'OK',
          message: '成功',
          data: {
            items: [{ id: 'tag-1', name: 'Dining' }, { id: 'tag-2', name: 'Payroll' }],
            pagination: { page: 1, page_size: 2, total: 2, total_pages: 1 },
          },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    }

    if (url.endsWith('/api/transactions') && init?.method === 'POST') {
      return new Response(
        JSON.stringify({
          code: 'OK',
          message: '成功',
          data: { id: 'txn-created', type: 'expense', amount: 88.5, category_name: 'Food', occurred_at: '2026-03-03T12:00:00Z' },
        }),
        { status: 201, headers: { 'Content-Type': 'application/json' } },
      )
    }

    if (url.endsWith('/api/transactions/txn-1') && init?.method === 'DELETE') {
      return new Response(
        JSON.stringify({
          code: 'OK',
          message: '成功',
          data: { deleted: true },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    }

    if (url.endsWith('/api/import/csv') && init?.method === 'POST') {
      return new Response(
        JSON.stringify({
          code: 'OK',
          message: '成功',
          data: {
            columns: ['posted_at', 'value', 'details', 'account', 'ledger'],
            sample_rows: [['2026-03-01 12:30:45', '25.00', 'Lunch', 'Cash Wallet', 'Default Ledger']],
            mappingSlots: ['amount', 'date', 'description'],
            mappingCandidates: ['posted_at', 'value', 'details', 'account', 'ledger'],
          },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    }

    if (url.endsWith('/api/import/csv/confirm') && init?.method === 'POST') {
      return new Response(
        JSON.stringify({
          code: 'OK',
          message: '成功',
          data: {
            success_count: 1,
            skip_count: 1,
            fail_count: 1,
            rows: [
              { row_index: 0, status: 'success' },
              { row_index: 1, status: 'skipped', reason: 'duplicate_transaction' },
              { row_index: 2, status: 'failed', reason: 'invalid_row' },
            ],
          },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    }

    if (url.endsWith('/api/transactions/txn-1') && init?.method === 'PATCH') {
      return new Response(
        JSON.stringify({
          code: 'OK',
          message: '成功',
          data: { id: 'txn-1', type: 'expense', amount: 25, category_name: 'Food', occurred_at: '2026-03-01T08:30:00Z' },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    }

    if (url.endsWith('/api/transactions/txn-5') && init?.method === 'PATCH') {
      return new Response(
        JSON.stringify({
          code: 'OK',
          message: '成功',
          data: { id: 'txn-5', type: 'expense', amount: 25, category_name: 'Food', occurred_at: '2026-03-01T09:00:00Z' },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    }

    throw new Error(`Unexpected URL: ${url}`)
  })
}

describe('transactions domain', () => {
  afterEach(() => {
    global.fetch = originalFetch
    vi.restoreAllMocks()
    window.localStorage.clear()
  })

  it('renders list and calendar transaction views with fetched data', async () => {
    const fetchMock = createTransactionsFetchMock()
    global.fetch = fetchMock as typeof fetch

    renderTransactionsApp(['/transactions'])
    const user = userEvent.setup()

    await waitFor(() => {
      expect(screen.getByText('Food')).toBeInTheDocument()
      expect(screen.getByText('Salary')).toBeInTheDocument()
    })

    expect(screen.getAllByText(/25\.00/).length).toBeGreaterThan(0)
    expect(screen.getByText(/15,000\.00/)).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /calendar view/i }))

    await waitFor(() => {
      expect(screen.getByText(/daily summary/i)).toBeInTheDocument()
      expect(screen.getByText(/month summary/i)).toBeInTheDocument()
    })
  })

  it('filters transactions with quick chips and restores a deleted transaction with undo', async () => {
    const fetchMock = createTransactionsFetchMock()
    global.fetch = fetchMock as typeof fetch

    renderTransactionsApp(['/transactions'])
    const user = userEvent.setup()

    await waitFor(() => {
      expect(screen.getByText('Food')).toBeInTheDocument()
      expect(screen.getByText('Salary')).toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: /^expense$/i }))

    expect(screen.getByText('Food')).toBeInTheDocument()
    expect(screen.queryByText('Salary')).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /^income$/i }))

    expect(screen.queryByText('Food')).not.toBeInTheDocument()
    expect(screen.getByText('Salary')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /^all$/i }))
    await user.click(screen.getAllByRole('button', { name: /delete lunch/i })[0])

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        '/api/transactions/txn-1',
        expect.objectContaining({ method: 'DELETE' }),
      )
    })
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /^undo$/i })).toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: /^undo$/i }))

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        '/api/transactions',
        expect.objectContaining({
          method: 'POST',
          body: expect.stringContaining('"amount":25'),
        }),
      )
    })
  })

  it('surfaces competitor-style smart views for transactions needing review', async () => {
    const fetchMock = createTransactionsFetchMock()
    global.fetch = fetchMock as typeof fetch

    renderTransactionsApp(['/transactions'])
    const user = userEvent.setup()

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: /smart views/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /needs review/i })).toBeInTheDocument()
      expect(screen.getByText(/1 uncategorized/i)).toBeInTheDocument()
      expect(screen.getByText(/1 possible duplicate/i)).toBeInTheDocument()
      expect(screen.getByText(/1 large transaction/i)).toBeInTheDocument()
    })
    expect(fetchMock.mock.calls.some(([url]) => String(url).includes('/api/transactions/review-items'))).toBe(false)

    await user.click(screen.getByRole('button', { name: /needs review/i }))

    expect(screen.getByText('Needs classification')).toBeInTheDocument()
    expect(screen.getByText('Conference flight')).toBeInTheDocument()
    expect(screen.getAllByText(/missing category/i).length).toBeGreaterThan(0)
    expect(screen.getAllByText(/large expense/i).length).toBeGreaterThan(0)
    expect(screen.getAllByText(/possible duplicate/i).length).toBeGreaterThan(0)
    expect(screen.queryByText('Salary')).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /focus uncategorized review/i }))

    expect(screen.getByText('Needs classification')).toBeInTheDocument()
    expect(screen.queryByText('Conference flight')).not.toBeInTheDocument()
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining('/api/transactions/review-items'),
        expect.any(Object),
      )
    })
  })

  it('submits the add transaction modal and previews import files', async () => {
    const fetchMock = createTransactionsFetchMock()
    global.fetch = fetchMock as typeof fetch

    renderTransactionsApp(['/transactions'])
    const user = userEvent.setup()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /add transaction/i })).toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: /add transaction/i }))
    await user.type(screen.getByLabelText(/amount/i), '88.5')
    await user.clear(screen.getByLabelText(/date & time/i))
    await user.type(screen.getByLabelText(/date & time/i), '2026-03-03T12:34:56')
    await user.selectOptions(screen.getByLabelText(/category/i), 'cat-1')
    await user.selectOptions(screen.getByLabelText(/account/i), 'acc-1')
    await user.click(screen.getByRole('button', { name: /save transaction/i }))

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        '/api/transactions',
        expect.objectContaining({ method: 'POST' }),
      )
    })
    const createCall = fetchMock.mock.calls.find(([url, init]) => {
      return String(url).endsWith('/api/transactions') && init?.method === 'POST'
    })
    expect(createCall?.[1]?.body).toEqual(
      expect.stringContaining(`"occurred_at":"${new Date('2026-03-03T12:34:56').toISOString()}"`),
    )

    await user.click(screen.getByRole('button', { name: /import/i }))
    const file = new File(['date,amount,description\n2026-03-01,25,Lunch'], 'transactions.csv', {
      type: 'text/csv',
    })
    await user.upload(screen.getByLabelText(/csv file/i), file)
    await user.click(screen.getByRole('button', { name: /preview import/i }))

    await waitFor(() => {
      expect(screen.getByText(/transactions.csv/i)).toBeInTheDocument()
      expect(screen.getByText(/detected columns/i)).toBeInTheDocument()
      expect(screen.getAllByText(/^posted_at$/i).length).toBeGreaterThan(0)
      expect(screen.getAllByText(/^details$/i).length).toBeGreaterThan(0)
    })
  })

  it('lets users map import fields and remembers default account and ledger for confirmation', async () => {
    const fetchMock = createTransactionsFetchMock()
    global.fetch = fetchMock as typeof fetch

    renderTransactionsApp(['/transactions'])
    const user = userEvent.setup()

    await user.click(await screen.findByRole('button', { name: /import/i }))
    const file = new File(
      ['posted_at,value,details,account,ledger\n2026-03-01 12:30:45,25,Lunch,Cash Wallet,Default Ledger'],
      'custom-ledger.csv',
      { type: 'text/csv' },
    )
    await user.upload(screen.getByLabelText(/csv file/i), file)
    await user.click(screen.getByRole('button', { name: /preview import/i }))

    await screen.findByRole('heading', { name: /field mapping/i })
    await user.selectOptions(screen.getByLabelText(/date column/i), 'posted_at')
    await user.selectOptions(screen.getByLabelText(/amount column/i), 'value')
    await user.selectOptions(screen.getByLabelText(/memo column/i), 'details')
    await user.selectOptions(screen.getByLabelText(/default account/i), 'acc-1')
    await user.selectOptions(screen.getByLabelText(/default ledger/i), 'ledger-1')
    await user.click(screen.getByRole('button', { name: /confirm import/i }))

    await waitFor(() => {
      const confirmCall = fetchMock.mock.calls.find(([url, init]) => (
        String(url).endsWith('/api/import/csv/confirm') && init?.method === 'POST'
      ))
      expect(confirmCall).toBeTruthy()
      expect(confirmCall?.[1]?.body).toEqual(expect.stringContaining('"date":"2026-03-01 12:30:45"'))
      expect(confirmCall?.[1]?.body).toEqual(expect.stringContaining('"description":"Lunch"'))
      expect(confirmCall?.[1]?.body).toEqual(expect.stringContaining('"default_account_id":"acc-1"'))
      expect(confirmCall?.[1]?.body).toEqual(expect.stringContaining('"default_ledger_id":"ledger-1"'))
    })
    expect(window.localStorage.getItem('xledger.import.defaults')).toContain('acc-1')
  })

  it('supports bulk category updates for selected transactions', async () => {
    const fetchMock = createTransactionsFetchMock()
    global.fetch = fetchMock as typeof fetch

    renderTransactionsApp(['/transactions'])
    const user = userEvent.setup()

    await screen.findByText('Food')
    await user.click(screen.getByRole('checkbox', { name: /select food/i }))
    await user.click(screen.getByRole('checkbox', { name: /select cafe/i }))
    await screen.findByText(/2 selected/i)
    await user.selectOptions(screen.getByLabelText(/bulk category/i), 'cat-1')
    await user.click(screen.getByRole('button', { name: /apply bulk category/i }))

    await waitFor(() => {
      const patchCalls = fetchMock.mock.calls.filter(([, init]) => init?.method === 'PATCH')
      expect(patchCalls).toHaveLength(2)
      expect(patchCalls[0][1]?.body).toEqual(expect.stringContaining('"category_id":"cat-1"'))
      expect(patchCalls[0][1]?.body).toEqual(expect.stringContaining('"amount":25'))
    })
  })

  it('keeps import confirmation results visible with row outcomes and problem download', async () => {
    const fetchMock = createTransactionsFetchMock()
    global.fetch = fetchMock as typeof fetch
    const createObjectURL = vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:import-problems')
    const revokeObjectURL = vi.spyOn(URL, 'revokeObjectURL').mockImplementation(() => undefined)
    vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => undefined)

    renderTransactionsApp(['/transactions'])
    const user = userEvent.setup()

    await user.click(await screen.findByRole('button', { name: /import/i }))
    const file = new File(
      ['date,amount,description\n2026-03-01,25,Lunch\n2026-03-01,25,Lunch\n,5,Bad row'],
      'transactions.csv',
      { type: 'text/csv' },
    )
    await user.upload(screen.getByLabelText(/csv file/i), file)
    await user.click(screen.getByRole('button', { name: /preview import/i }))
    await screen.findByText(/detected columns/i)
    await user.click(screen.getByRole('button', { name: /confirm import/i }))

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: /import results/i })).toBeInTheDocument()
      expect(screen.getByText(/1 imported/i)).toBeInTheDocument()
      expect(screen.getByText(/1 skipped/i)).toBeInTheDocument()
      expect(screen.getByText(/1 failed/i)).toBeInTheDocument()
      expect(screen.getAllByText(/duplicate transaction/i).length).toBeGreaterThan(0)
      expect(screen.getByText(/invalid row/i)).toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: /download problem rows/i }))

    expect(createObjectURL).toHaveBeenCalled()
    expect(revokeObjectURL).toHaveBeenCalledWith('blob:import-problems')
  })

  it('exports the current filtered transaction range as a CSV file', async () => {
    const fetchMock = createTransactionsFetchMock()
    fetchMock.mockImplementation(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input)
      if (url.includes('/api/export?')) {
        return new Response('occurred_at,amount,type,category_name\n2026-03-01T08:30:00Z,25,expense,Food\n', {
          status: 200,
          headers: { 'Content-Type': 'text/csv' },
        })
      }
      return createTransactionsFetchMock()(input, init)
    })
    global.fetch = fetchMock as typeof fetch
    const createObjectURL = vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:transactions-export')
    const revokeObjectURL = vi.spyOn(URL, 'revokeObjectURL').mockImplementation(() => undefined)
    vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => undefined)

    renderTransactionsApp(['/transactions'])
    const user = userEvent.setup()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /export/i })).toBeInTheDocument()
      expect(screen.getByRole('option', { name: 'Cash Wallet' })).toBeInTheDocument()
      expect(screen.getByRole('option', { name: 'Default Ledger' })).toBeInTheDocument()
    })

    await user.selectOptions(screen.getByLabelText(/source/i), 'acc-1')
    await user.selectOptions(screen.getByLabelText(/^ledger$/i), 'ledger-1')
    await user.click(screen.getByRole('button', { name: /export/i }))

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringMatching(/\/api\/export\?/),
        expect.objectContaining({
          headers: expect.objectContaining({ Authorization: 'Bearer access.demo.token' }),
        }),
      )
    })
    const exportCall = fetchMock.mock.calls.find(([url]) => String(url).includes('/api/export?'))
    const exportUrl = String(exportCall?.[0] ?? '')
    expect(exportUrl).toContain('format=csv')
    expect(exportUrl).toContain('account_id=acc-1')
    expect(exportUrl).toContain('ledger_id=ledger-1')
    expect(exportUrl).toContain('from=')
    expect(exportUrl).toContain('to=')
    expect(createObjectURL).toHaveBeenCalled()
    expect(revokeObjectURL).toHaveBeenCalledWith('blob:transactions-export')
  })

  it('opens a reviewed transaction day from the review queue', async () => {
    const fetchMock = createTransactionsFetchMock()
    global.fetch = fetchMock as typeof fetch
    const expectedFrom = new Date(2026, 2, 3, 0, 0, 0, 0).toISOString()
    const expectedTo = new Date(2026, 2, 3, 23, 59, 59, 999).toISOString()

    renderTransactionsApp(['/transactions'])
    const user = userEvent.setup()

    await user.click(await screen.findByRole('button', { name: /needs review/i }))
    await user.click(await screen.findByRole('button', { name: /open day for needs classification/i }))

    await waitFor(() => {
      expect(fetchMock.mock.calls.some(([input]) => {
        const url = new URL(String(input), 'http://localhost')
        return (
          url.pathname === '/api/transactions' &&
          url.searchParams.get('date_from') === expectedFrom &&
          url.searchParams.get('date_to') === expectedTo
        )
      })).toBe(true)
    })
  })
})
