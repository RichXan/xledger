import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import App from './App'
import { AuthProvider } from './features/auth/auth-context'
import i18n, { getCurrentLanguage } from './i18n'

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

function mockDashboardApi(language?: string) {
  global.fetch = vi.fn(async (input: RequestInfo | URL) => {
    const url = String(input)

    if (url.endsWith('/api/auth/me')) {
      return new Response(
        JSON.stringify({
          code: 'OK',
          message: 'success',
          data: { email: 'demo@example.com', language },
        }),
        {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        },
      )
    }

    if (url.endsWith('/api/stats/overview')) {
      return new Response(
        JSON.stringify({
          code: 'OK',
          message: 'success',
          data: { total_assets: 1248300, income: 42850, expense: 18240.5, net: 24609.5 },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    }

    if (url.includes('/api/stats/trend?')) {
      return new Response(
        JSON.stringify({
          code: 'OK',
          message: 'success',
          data: { points: [{ bucket_start: '2026-03-01T00:00:00Z', income: 1200, expense: 300 }] },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    }

    if (url.endsWith('/api/accounts') || url.endsWith('/api/ledgers')) {
      return new Response(
        JSON.stringify({
          code: 'OK',
          message: 'success',
          data: { items: [{ id: 'existing' }], pagination: { page: 1, page_size: 20, total: 1, total_pages: 1 } },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    }

    if (url.endsWith('/api/categories') || url.endsWith('/api/tags') || url.endsWith('/api/personal-access-tokens')) {
      return new Response(
        JSON.stringify({
          code: 'OK',
          message: 'success',
          data: { items: [], pagination: { page: 1, page_size: 20, total: 0, total_pages: 0 } },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    }

    if (url.includes('/api/stats/category?')) {
      return new Response(
        JSON.stringify({
          code: 'OK',
          message: 'success',
          data: { items: [] },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    }

    if (url.includes('/api/transactions?')) {
      return new Response(
        JSON.stringify({
          code: 'OK',
          message: 'success',
          data: { items: [{ id: 'tx-1' }], pagination: { page: 1, page_size: 1, total: 1, total_pages: 1 } },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    }

    throw new Error(`Unexpected URL: ${url}`)
  }) as typeof fetch
}

describe('language switching', () => {
  beforeEach(async () => {
    await i18n.changeLanguage('en')
    window.localStorage.setItem(
      'xledger.auth',
      JSON.stringify({
        accessToken: 'access.demo.token',
        refreshToken: 'refresh.demo.token',
        email: 'demo@example.com',
      }),
    )
  })

  afterEach(async () => {
    cleanup()
    global.fetch = originalFetch
    window.localStorage.clear()
    await i18n.changeLanguage('en')
  })

  it('switches dashboard text to readable Chinese', async () => {
    mockDashboardApi()
    const user = userEvent.setup()

    renderApp(['/dashboard'])

    expect(await screen.findByRole('heading', { name: /financial overview/i })).toBeInTheDocument()

    await user.selectOptions(screen.getByLabelText(/select language/i), 'zh')

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: '财务概览' })).toBeInTheDocument()
      expect(screen.getByRole('option', { name: '中文' })).toBeInTheDocument()
    })

    const primaryNav = screen.getByRole('navigation', { name: '主导航' })
    expect(within(primaryNav).getByRole('link', { name: '首页' })).toBeInTheDocument()
    expect(within(primaryNav).getByRole('link', { name: '交易' })).toBeInTheDocument()
    expect(within(primaryNav).getByRole('link', { name: '统计' })).toBeInTheDocument()
    expect(within(primaryNav).getByRole('link', { name: '账户' })).toBeInTheDocument()
    expect(within(primaryNav).getByRole('link', { name: '快捷记账' })).toBeInTheDocument()
    expect(within(primaryNav).getByRole('link', { name: '设置' })).toBeInTheDocument()
  })

  it('switches protected page content to readable Chinese', async () => {
    const pageExpectations = [
      { path: '/transactions', english: /transactions/i, chinese: '交易' },
      { path: '/analytics', english: /analytics/i, chinese: '统计分析' },
      { path: '/accounts', english: /accounts/i, chinese: '账户' },
      { path: '/shortcut', english: /apple shortcuts/i, chinese: 'Apple 快捷指令' },
      { path: '/settings', english: /settings/i, chinese: '设置' },
    ]

    for (const { path, english, chinese } of pageExpectations) {
      mockDashboardApi()
      const user = userEvent.setup()
      const view = renderApp([path])

      expect(await screen.findByRole('heading', { name: english })).toBeInTheDocument()

      await user.selectOptions(screen.getByLabelText(/select language/i), 'zh')

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: chinese })).toBeInTheDocument()
      })

      view.unmount()
      await i18n.changeLanguage('en')
    }
  })

  it('normalizes regional browser language codes for app state', async () => {
    await i18n.changeLanguage('zh-CN')

    expect(getCurrentLanguage()).toBe('zh')
  })

  it('only restores language from an explicit saved selection', () => {
    const detectionOptions = i18n.options.detection as { order?: string[] } | undefined

    expect(detectionOptions?.order).toEqual(['localStorage'])
  })

  it('keeps English when the current-user endpoint returns Chinese before the user chooses it', async () => {
    mockDashboardApi('zh-CN')

    renderApp(['/dashboard'])

    expect(await screen.findByRole('heading', { name: /financial overview/i })).toBeInTheDocument()
  })
})
