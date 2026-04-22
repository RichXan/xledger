import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
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

async function waitForLoginForm() {
  await screen.findByRole('heading', { name: /welcome back/i })
  await screen.findByLabelText(/email address/i)
  await screen.findByRole('button', { name: /send verification code/i })
}

describe('auth flow', () => {
  afterEach(() => {
    cleanup()
    global.fetch = originalFetch
    window.localStorage.clear()
  })

  it('sends a verification code and reveals the code field', async () => {
    const fetchMock = vi.fn(async () =>
      new Response(
        JSON.stringify({
          code: 'OK',
          message: '成功',
          data: { code_sent: true },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      ),
    )
    global.fetch = fetchMock as typeof fetch

    renderApp(['/login'])
    await waitForLoginForm()
    const user = userEvent.setup()

    await user.type(screen.getByLabelText(/email address/i), 'demo@example.com')
    await user.click(screen.getByRole('button', { name: /send verification code/i }))

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1))
    expect(screen.getByLabelText(/verification code/i)).toBeInTheDocument()
    expect(screen.getByText(/we sent a 6-digit code to demo@example.com/i)).toBeInTheDocument()
  })

  it('persists tokens after successful verification and unlocks protected routes', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input)

      if (url.endsWith('/api/auth/send-code')) {
        return new Response(
          JSON.stringify({ code: 'OK', message: '成功', data: { code_sent: true } }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.endsWith('/api/auth/verify-code')) {
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: '成功',
            data: {
              access_token: 'access.demo.token',
              refresh_token: 'refresh.demo.token',
            },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.endsWith('/api/auth/me')) {
        expect(init?.headers).toMatchObject({ Authorization: 'Bearer access.demo.token' })
        return new Response(
          JSON.stringify({ code: 'OK', message: '成功', data: { email: 'demo@example.com' } }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      throw new Error(`Unexpected URL: ${url}`)
    })
    global.fetch = fetchMock as typeof fetch

    renderApp(['/login'])
    await waitForLoginForm()
    const user = userEvent.setup()

    await user.type(screen.getByLabelText(/email address/i), 'demo@example.com')
    await user.click(screen.getByRole('button', { name: /send verification code/i }))
    await user.type(screen.getByLabelText(/verification code/i), '123456')
    await user.click(screen.getByRole('button', { name: /verify and continue/i }))

    await waitFor(() => {
      expect(window.localStorage.getItem('xledger.auth')).toContain('access.demo.token')
    })

    await waitFor(() => {
      expect(screen.getByText(/demo@example.com/i)).toBeInTheDocument()
    })
  })
})
