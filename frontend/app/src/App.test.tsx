import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
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
})
