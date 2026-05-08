import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import App from './App'
import { AuthProvider } from './features/auth/auth-context'

const originalFetch = global.fetch

function renderShortcutApp() {
  window.localStorage.setItem(
    'xledger.auth',
    JSON.stringify({
      accessToken: 'access.shortcut.token',
      refreshToken: 'refresh.shortcut.token',
      email: 'shortcut@example.com',
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
        <MemoryRouter initialEntries={['/shortcut']}>
          <App />
        </MemoryRouter>
      </AuthProvider>
    </QueryClientProvider>,
  )
}

describe('shortcut page', () => {
  afterEach(() => {
    global.fetch = originalFetch
    window.localStorage.clear()
  })

  it('shows a same-origin endpoint and a concrete JSON example after generating credentials', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input)

      if (url.endsWith('/api/auth/me')) {
        return new Response(
          JSON.stringify({ code: 'OK', message: 'Success', data: { email: 'shortcut@example.com' } }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.endsWith('/api/shortcuts/generate')) {
        expect(init?.method).toBe('POST')
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: 'Success',
            data: {
              pat_token: 'pat.demo.token',
              api_endpoint: 'http://127.0.0.1',
              expires_at: '2026-08-06T00:00:00Z',
            },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      throw new Error(`Unexpected URL: ${url}`)
    })
    global.fetch = fetchMock as typeof fetch

    renderShortcutApp()
    const user = userEvent.setup()

    await user.click(await screen.findByRole('button', { name: /generate new credentials/i }))

    expect(await screen.findByText(`${window.location.origin}/api/shortcuts/quick-add`)).toBeInTheDocument()
    expect(screen.queryByText('http://127.0.0.1/api/shortcuts/quick-add')).not.toBeInTheDocument()
    await waitFor(() => {
      expect(screen.getByText(/\"amount\": 35/i)).toBeInTheDocument()
      expect(screen.getByText(/\"category\": \"Lunch\"/i)).toBeInTheDocument()
    })
  })
})
