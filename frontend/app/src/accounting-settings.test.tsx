import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import App from './App'
import { AuthProvider } from './features/auth/auth-context'

const originalFetch = global.fetch
const originalCreateObjectURL = URL.createObjectURL
const originalRevokeObjectURL = URL.revokeObjectURL

function renderManagementApp(initialEntries: string[]) {
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

function createManagementFetchMock() {
  return vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input)

    if (url.endsWith('/api/auth/me')) {
      return new Response(
        JSON.stringify({ code: 'OK', message: '成功', data: { email: 'demo@example.com' } }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    }

    if (url.endsWith('/api/accounts') && (!init?.method || init.method === 'GET')) {
      return new Response(
        JSON.stringify({
          code: 'OK',
          message: '成功',
          data: {
            items: [{ id: 'acc-1', name: 'Cash Wallet', type: 'cash', initial_balance: 1000 }],
            pagination: { page: 1, page_size: 1, total: 1, total_pages: 1 },
          },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    }

    if (url.endsWith('/api/accounts') && init?.method === 'POST') {
      return new Response(
        JSON.stringify({
          code: 'OK',
          message: '成功',
          data: { id: 'acc-2', name: 'Savings Vault', type: 'bank', initial_balance: 2000 },
        }),
        { status: 201, headers: { 'Content-Type': 'application/json' } },
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
            items: [{ id: 'cat-1', name: 'Food' }],
            pagination: { page: 1, page_size: 1, total: 1, total_pages: 1 },
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
            items: [{ id: 'tag-1', name: 'Dining' }],
            pagination: { page: 1, page_size: 1, total: 1, total_pages: 1 },
          },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    }

    if (url.endsWith('/api/personal-access-tokens') && init?.method === 'POST') {
      return new Response(
        JSON.stringify({ code: 'OK', message: '成功', data: { token: 'pat.secret.value', id: 'pat-2', expires_at: '2027-01-01T00:00:00Z' } }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    }

    if (url.endsWith('/api/personal-access-tokens')) {
      return new Response(
        JSON.stringify({
          code: 'OK',
          message: '成功',
          data: {
            items: [{ id: 'pat-1', name: 'default', expires_at: '2026-12-31T00:00:00Z' }],
            pagination: { page: 1, page_size: 1, total: 1, total_pages: 1 },
          },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    }

    if (url.endsWith('/api/personal-access-tokens/pat-1') && init?.method === 'DELETE') {
      return new Response(
        JSON.stringify({ code: 'OK', message: '成功', data: { revoked: true } }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    }

    if (url.includes('/api/export?')) {
      return new Response('date,amount\n2026-03-01,25.00', {
        status: 200,
        headers: { 'Content-Type': 'text/csv' },
      })
    }

    throw new Error(`Unexpected URL: ${url}`)
  })
}

describe('accounting settings domain', () => {
  beforeEach(() => {
    URL.createObjectURL = vi.fn(() => 'blob:mock-url')
    URL.revokeObjectURL = vi.fn()
  })

  afterEach(() => {
    global.fetch = originalFetch
    URL.createObjectURL = originalCreateObjectURL
    URL.revokeObjectURL = originalRevokeObjectURL
    window.localStorage.clear()
  })

  it('renders accounts and creates a new account from the management page', async () => {
    const fetchMock = createManagementFetchMock()
    global.fetch = fetchMock as typeof fetch

    renderManagementApp(['/accounts'])
    const user = userEvent.setup()

    await waitFor(() => {
      expect(screen.getByText('Cash Wallet')).toBeInTheDocument()
      expect(screen.getByText('Default Ledger')).toBeInTheDocument()
      expect(screen.getByText('Food')).toBeInTheDocument()
      expect(screen.getByText('Dining')).toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: /new account/i }))
    await user.type(screen.getByLabelText(/account name/i), 'Savings Vault')
    await user.selectOptions(screen.getByLabelText(/account type/i), 'bank')
    await user.type(screen.getByLabelText(/initial balance/i), '2000')
    await user.click(screen.getByRole('button', { name: /create account/i }))

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        '/api/accounts',
        expect.objectContaining({ method: 'POST' }),
      )
    })
  })

  it('creates and revokes PATs and triggers CSV export from settings', async () => {
    const fetchMock = createManagementFetchMock()
    global.fetch = fetchMock as typeof fetch

    renderManagementApp(['/settings'])
    const user = userEvent.setup()

    await waitFor(() => {
      expect(screen.getByText('pat-1')).toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: /create pat/i }))

    await waitFor(() => {
      expect(screen.getByText(/pat.secret.value/i)).toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: /revoke pat-1/i }))

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        '/api/personal-access-tokens/pat-1',
        expect.objectContaining({ method: 'DELETE' }),
      )
    })

    await user.click(screen.getByRole('button', { name: /export csv/i }))

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringMatching(/\/api\/export\?/),
        expect.any(Object),
      )
      expect(URL.createObjectURL).toHaveBeenCalled()
    })
  })
})
