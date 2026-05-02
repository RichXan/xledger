import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import App from './App'
import { AuthProvider } from './features/auth/auth-context'

const originalFetch = global.fetch

function renderProtectedApp(initialEntries: string[]) {
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

describe('dashboard and analytics pages', () => {
  afterEach(() => {
    global.fetch = originalFetch
    window.localStorage.clear()
  })

  it('renders overview metrics and trend data on the dashboard', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input)

      if (url.endsWith('/api/auth/me')) {
        return new Response(
          JSON.stringify({ code: 'OK', message: '成功', data: { email: 'demo@example.com' } }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.includes('/api/stats/overview')) {
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
            data: {
              points: [
                { bucket_start: '2026-03-01T00:00:00Z', income: 1200, expense: 300 },
                { bucket_start: '2026-03-02T00:00:00Z', income: 800, expense: 450 },
              ],
            },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      throw new Error(`Unexpected URL: ${url}`)
    })
    global.fetch = fetchMock as typeof fetch

    renderProtectedApp(['/dashboard'])

    await waitFor(() => {
      expect(screen.getByText('¥1,248,300.00')).toBeInTheDocument()
      expect(screen.getByText('¥42,850.00')).toBeInTheDocument()
      expect(screen.getByText('¥18,240.50')).toBeInTheDocument()
      expect(screen.getByText('¥24,609.50')).toBeInTheDocument()
      expect(screen.getByText(/tap or hover bars/i)).toBeInTheDocument()
      expect(screen.getByText(/income:/i)).toBeInTheDocument()
      expect(screen.getByText(/expense:/i)).toBeInTheDocument()
    })
  })

  it('renders analytics category and trend insights from reporting endpoints', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input)

      if (url.endsWith('/api/auth/me')) {
        return new Response(
          JSON.stringify({ code: 'OK', message: '成功', data: { email: 'demo@example.com' } }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.includes('/api/stats/trend?')) {
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: '成功',
            data: {
              points: [
                { bucket_start: '2026-03-01T00:00:00Z', income: 1000, expense: 500 },
                { bucket_start: '2026-03-02T00:00:00Z', income: 1300, expense: 650 },
              ],
            },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.includes('/api/stats/category')) {
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: '成功',
            data: {
              items: [
                { category_id: 'food', category_name: 'Food', amount: 3200 },
                { category_id: 'travel', category_name: 'Travel', amount: 1800 },
              ],
            },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      throw new Error(`Unexpected URL: ${url}`)
    })
    global.fetch = fetchMock as typeof fetch

    renderProtectedApp(['/analytics'])

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: /analytics/i })).toBeInTheDocument()
      expect(screen.getByRole('heading', { name: /spending cloud/i })).toBeInTheDocument()
      expect(screen.getByRole('heading', { name: /cashflow rhythm/i })).toBeInTheDocument()
      expect(screen.getAllByText('Food').length).toBeGreaterThan(0)
      expect(screen.getAllByText('¥3,200.00').length).toBeGreaterThan(0)
      expect(screen.getAllByText('Travel').length).toBeGreaterThan(0)
      expect(screen.getAllByText('¥1,800.00').length).toBeGreaterThan(0)
    })
  })
})
