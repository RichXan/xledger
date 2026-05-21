import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import App from './App'
import { AuthProvider } from './features/auth/auth-context'

const originalFetch = global.fetch

function renderApp(initialEntries: string[]) {
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

describe('App shell', () => {
  afterEach(() => {
    global.fetch = originalFetch
    window.localStorage.clear()
  })

  it('renders the login route by default', async () => {
    renderApp(['/login'])

    expect(await screen.findByRole('heading', { name: /welcome back/i })).toBeInTheDocument()
    expect(await screen.findByRole('button', { name: /send verification code/i })).toBeInTheDocument()
  })

  it('renders main navigation on dashboard routes', async () => {
    global.fetch = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input)

      if (url.endsWith('/api/auth/me')) {
        return new Response(
          JSON.stringify({ code: 'OK', message: '成功', data: { email: 'demo@example.com' } }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.endsWith('/api/stats/overview')) {
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: '成功',
            data: { total_assets: 1248300, income: 42850, expense: 18240.5, net: 24609.5 },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.includes('/api/stats/trend?')) {
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: '成功',
            data: { points: [{ bucket_start: '2026-03-01T00:00:00Z', income: 1200, expense: 300 }] },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      throw new Error(`Unexpected URL: ${url}`)
    }) as typeof fetch

    window.localStorage.setItem(
      'xledger.auth',
      JSON.stringify({
        accessToken: 'access.demo.token',
        refreshToken: 'refresh.demo.token',
        email: 'demo@example.com',
      }),
    )

    renderApp(['/dashboard'])

    await waitFor(() => {
      expect(screen.getByRole('navigation', { name: /primary/i })).toBeInTheDocument()
      expect(screen.getByText(/demo@example.com/i)).toBeInTheDocument()
    })
  })

  it('opens a QR code entry for mobile devices from the app shell', async () => {
    const user = userEvent.setup()
    global.fetch = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input)

      if (url.endsWith('/api/auth/me')) {
        return new Response(
          JSON.stringify({ code: 'OK', message: 'Success', data: { email: 'demo@example.com' } }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.endsWith('/api/stats/overview')) {
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: 'Success',
            data: { total_assets: 1248300, income: 42850, expense: 18240.5, net: 24609.5 },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.includes('/api/stats/trend?')) {
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: 'Success',
            data: { points: [{ bucket_start: '2026-03-01T00:00:00Z', income: 1200, expense: 300 }] },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      throw new Error(`Unexpected URL: ${url}`)
    }) as typeof fetch

    window.localStorage.setItem(
      'xledger.auth',
      JSON.stringify({
        accessToken: 'access.demo.token',
        refreshToken: 'refresh.demo.token',
        email: 'demo@example.com',
      }),
    )

    renderApp(['/dashboard'])

    const mobileEntryButtons = await screen.findAllByRole('button', { name: /use on mobile device/i })
    await user.click(mobileEntryButtons[0])

    expect(screen.getByRole('heading', { name: /use on mobile device/i })).toBeInTheDocument()
    expect(screen.getByLabelText(/mobile entry qr code/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /switch to mobile version/i })).toBeInTheDocument()
  })

  it('renders the mobile install guide as a protected app page', async () => {
    global.fetch = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input)

      if (url.endsWith('/api/auth/me')) {
        return new Response(
          JSON.stringify({ code: 'OK', message: 'Success', data: { email: 'mobile@example.com' } }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.endsWith('/api/accounts') || url.endsWith('/api/ledgers')) {
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: 'Success',
            data: { items: [{ id: 'existing', name: 'Cash' }], pagination: { page: 1, page_size: 20, total: 1, total_pages: 1 } },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.includes('/api/transactions?')) {
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: 'Success',
            data: { items: [{ id: 'tx-1' }], pagination: { page: 1, page_size: 1, total: 1, total_pages: 1 } },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      throw new Error(`Unexpected URL: ${url}`)
    }) as typeof fetch

    window.localStorage.setItem(
      'xledger.auth',
      JSON.stringify({
        accessToken: 'access.mobile.token',
        refreshToken: 'refresh.mobile.token',
        email: 'mobile@example.com',
      }),
    )

    renderApp(['/install'])

    expect(await screen.findByRole('heading', { name: /add to home screen/i })).toBeInTheDocument()
    expect(screen.getByText(/home-screen app/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /back to app/i })).toBeInTheDocument()
  })

  it('guides scanned mobile users to add Xledger to the home screen before opening the app', async () => {
    global.fetch = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input)

      if (url.endsWith('/api/auth/me')) {
        return new Response(
          JSON.stringify({ code: 'OK', message: 'Success', data: { email: 'mobile@example.com' } }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      throw new Error(`Unexpected URL: ${url}`)
    }) as typeof fetch

    window.localStorage.setItem(
      'xledger.auth',
      JSON.stringify({
        accessToken: 'access.mobile.token',
        refreshToken: 'refresh.mobile.token',
        email: 'mobile@example.com',
      }),
    )

    renderApp(['/mobile'])

    expect(await screen.findByRole('heading', { name: /add to home screen/i })).toBeInTheDocument()
    expect(screen.getByText(/after scanning the QR code/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /open xledger/i })).toBeInTheDocument()
  })

  it('shows first-time onboarding when ledger setup is still empty', async () => {
    global.fetch = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input)

      if (url.endsWith('/api/auth/me')) {
        return new Response(
          JSON.stringify({ code: 'OK', message: '鎴愬姛', data: { email: 'first@example.com' } }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.endsWith('/api/stats/overview')) {
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: '鎴愬姛',
            data: { total_assets: 0, income: 0, expense: 0, net: 0 },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.includes('/api/stats/trend?')) {
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: '鎴愬姛',
            data: { points: [] },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.endsWith('/api/accounts')) {
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: '鎴愬姛',
            data: { items: [], pagination: { page: 1, page_size: 20, total: 0, total_pages: 0 } },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.endsWith('/api/ledgers')) {
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: '鎴愬姛',
            data: { items: [], pagination: { page: 1, page_size: 20, total: 0, total_pages: 0 } },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.includes('/api/transactions?')) {
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: '鎴愬姛',
            data: { items: null, pagination: { page: 1, page_size: 1, total: 0, total_pages: 0 } },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      throw new Error(`Unexpected URL: ${url}`)
    }) as typeof fetch

    window.localStorage.setItem(
      'xledger.auth',
      JSON.stringify({
        accessToken: 'access.first.token',
        refreshToken: 'refresh.first.token',
        email: 'first@example.com',
      }),
    )

    renderApp(['/dashboard'])

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: /getting started/i })).toBeInTheDocument()
      expect(screen.getByText(/setup checklist/i)).toBeInTheDocument()
      expect(screen.getByText(/create a balance source/i)).toBeInTheDocument()
      expect(screen.getByText(/record the first transaction/i)).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /set up accounts/i })).toBeInTheDocument()
    })
  })

  it('keeps onboarding visible as progress after the first account exists', async () => {
    global.fetch = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input)

      if (url.endsWith('/api/auth/me')) {
        return new Response(
          JSON.stringify({ code: 'OK', message: 'Success', data: { email: 'progress@example.com' } }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.endsWith('/api/stats/overview')) {
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: 'Success',
            data: { total_assets: 1000, income: 0, expense: 0, net: 0 },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.includes('/api/stats/trend?')) {
        return new Response(
          JSON.stringify({ code: 'OK', message: 'Success', data: { points: [] } }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.endsWith('/api/accounts')) {
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: 'Success',
            data: {
              items: [{ id: 'acc-1', name: 'Cash Wallet', type: 'cash', initial_balance: 1000 }],
              pagination: { page: 1, page_size: 20, total: 1, total_pages: 1 },
            },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.endsWith('/api/ledgers')) {
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: 'Success',
            data: {
              items: [{ id: 'ledger-1', name: 'Default Ledger', is_default: true }],
              pagination: { page: 1, page_size: 20, total: 1, total_pages: 1 },
            },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.includes('/api/transactions?')) {
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: 'Success',
            data: { items: [], pagination: { page: 1, page_size: 1, total: 0, total_pages: 0 } },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      throw new Error(`Unexpected URL: ${url}`)
    }) as typeof fetch

    window.localStorage.setItem(
      'xledger.auth',
      JSON.stringify({
        accessToken: 'access.progress.token',
        refreshToken: 'refresh.progress.token',
        email: 'progress@example.com',
      }),
    )

    renderApp(['/dashboard'])

    await waitFor(() => {
      expect(screen.getByText(/1\/2 completed/i)).toBeInTheDocument()
      expect(screen.getByText(/cash wallet/i)).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /add transaction/i })).toBeInTheDocument()
    })
  })
})
