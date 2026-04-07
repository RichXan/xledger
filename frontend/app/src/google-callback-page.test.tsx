import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { GoogleCallbackPage } from './pages/google-callback-page'
import { AuthProvider } from './features/auth/auth-context'

const originalFetch = global.fetch

function renderGoogleCallbackPage(initialEntry: string) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })

  return render(
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <MemoryRouter initialEntries={[initialEntry]}>
          <Routes>
            <Route path="/auth/google/callback" element={<GoogleCallbackPage />} />
          </Routes>
        </MemoryRouter>
      </AuthProvider>
    </QueryClientProvider>,
  )
}

describe('GoogleCallbackPage', () => {
  afterEach(() => {
    global.fetch = originalFetch
    window.localStorage.clear()
  })

  it('shows a specific message for token exchange failures', async () => {
    global.fetch = vi.fn(async () => {
      throw new Error('fetch should not be called when error_code is present')
    }) as typeof fetch

    renderGoogleCallbackPage('/auth/google/callback?error_code=AUTH_OAUTH_FAILED&error_reason=google_token_exchange_failed')

    expect(
      await screen.findByText(/Google 登录回调地址不匹配或授权换取令牌失败/i),
    ).toBeInTheDocument()
  })

  it('shows a specific message for unconfigured oauth', async () => {
    global.fetch = vi.fn(async () => {
      throw new Error('fetch should not be called when error_code is present')
    }) as typeof fetch

    renderGoogleCallbackPage('/auth/google/callback?error_code=AUTH_OAUTH_FAILED&error_reason=google_oauth_not_configured')

    expect(await screen.findByText(/Google OAuth 尚未正确配置/i)).toBeInTheDocument()
  })
})
