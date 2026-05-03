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
              },
              {
                id: 'txn-2',
                type: 'income',
                amount: 15000,
                category_name: 'Salary',
                occurred_at: '2026-03-02T09:00:00Z',
              },
            ],
            pagination: { page: 1, page_size: 20, total: 2, total_pages: 1 },
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

    if (url.endsWith('/api/import/csv') && init?.method === 'POST') {
      return new Response(
        JSON.stringify({
          code: 'OK',
          message: '成功',
          data: {
            columns: ['date', 'amount', 'description'],
            sample_rows: [['2026-03-01', '25.00', 'Lunch']],
            mappingSlots: ['amount', 'date', 'description'],
            mappingCandidates: ['date', 'amount', 'description'],
          },
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

    expect(screen.getByText(/25\.00/)).toBeInTheDocument()
    expect(screen.getByText(/15,000\.00/)).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /calendar view/i }))

    await waitFor(() => {
      expect(screen.getByText(/daily summary/i)).toBeInTheDocument()
      expect(screen.getByText(/month summary/i)).toBeInTheDocument()
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
      expect(screen.getByText(/^date$/i)).toBeInTheDocument()
      expect(screen.getByText(/^description$/i)).toBeInTheDocument()
    })
  })
})
