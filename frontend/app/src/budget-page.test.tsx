import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import App from './App'
import { AuthProvider } from './features/auth/auth-context'

const originalFetch = global.fetch

function renderBudgetApp() {
  window.localStorage.setItem(
    'xledger.auth',
    JSON.stringify({
      accessToken: 'access.demo.token',
      refreshToken: 'refresh.demo.token',
      email: 'budget@example.com',
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
        <MemoryRouter initialEntries={['/budgets']}>
          <App />
        </MemoryRouter>
      </AuthProvider>
    </QueryClientProvider>,
  )
}

describe('budget page', () => {
  afterEach(() => {
    global.fetch = originalFetch
    window.localStorage.clear()
  })

  it('shows monthly budget usage and lets users create a category budget', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input)

      if (url.endsWith('/api/auth/me')) {
        return new Response(
          JSON.stringify({ code: 'OK', message: 'Success', data: { email: 'budget@example.com' } }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.endsWith('/api/categories')) {
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: 'Success',
            data: {
              items: [
                { id: 'cat-food', name: 'Food' },
                { id: 'cat-old', name: 'Archived Old', archived_at: '2026-05-01T00:00:00Z' },
              ],
              pagination: { page: 1, page_size: 2, total: 2, total_pages: 1 },
            },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.endsWith('/api/budgets') && init?.method === 'POST') {
        expect(init.body).toBe(JSON.stringify({ category_id: 'cat-food', amount: 1200, alert_at: 75 }))
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: 'Success',
            data: { id: 'budget-new', category_id: 'cat-food', amount: 1200, period: 'monthly', alert_at: 75 },
          }),
          { status: 201, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.endsWith('/api/budgets')) {
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: 'Success',
            data: {
              budgets: [
                {
                  id: 'budget-food',
                  category_id: 'cat-food',
                  amount: 1000,
                  period: 'monthly',
                  alert_at: 80,
                  spent: 620,
                  remaining: 380,
                  percent: 62,
                },
              ],
            },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      throw new Error(`Unexpected URL: ${url}`)
    })
    global.fetch = fetchMock as typeof fetch

    renderBudgetApp()
    const user = userEvent.setup()

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: /budgets/i })).toBeInTheDocument()
      expect(screen.getAllByText('Food').length).toBeGreaterThan(0)
      expect(screen.getAllByText('62%').length).toBeGreaterThan(0)
      expect(screen.getAllByText('¥380.00').length).toBeGreaterThan(0)
      expect(screen.queryByRole('option', { name: /archived old/i })).not.toBeInTheDocument()
    })

    await user.selectOptions(screen.getByLabelText(/category/i), 'cat-food')
    await user.clear(screen.getByLabelText(/monthly limit/i))
    await user.type(screen.getByLabelText(/monthly limit/i), '1200')
    await user.clear(screen.getByLabelText(/alert at/i))
    await user.type(screen.getByLabelText(/alert at/i), '75')
    await user.click(screen.getByRole('button', { name: /create budget/i }))

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        '/api/budgets',
        expect.objectContaining({ method: 'POST' }),
      )
    })
  })
})
