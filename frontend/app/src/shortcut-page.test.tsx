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

      if (url.endsWith('/api/shortcuts/quick-add')) {
        expect(init?.method).toBe('POST')
        expect(init?.headers).toEqual(expect.objectContaining({ Authorization: 'Bearer pat.demo.token' }))
        expect(JSON.parse(String(init?.body))).toEqual(expect.objectContaining({
          amount: 35,
          type: 'expense',
          category: 'Lunch',
        }))
        return new Response(
          JSON.stringify({ code: 'OK', message: 'Success', data: { id: 'txn-1' } }),
          { status: 201, headers: { 'Content-Type': 'application/json' } },
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
      expect(screen.getAllByText(/\"amount\": 35/i).length).toBeGreaterThan(0)
      expect(screen.getAllByText(/\"category\": \"Lunch\"/i).length).toBeGreaterThan(0)
    })
    expect(screen.getByText(/copy shortcut setup/i)).toBeInTheDocument()
    expect(screen.getByText(/Authorization: Bearer pat.demo.token/i)).toBeInTheDocument()
    expect(screen.getByText(/Method: POST/i)).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /send test entry/i }))

    expect(await screen.findByRole('status')).toHaveTextContent(/test entry sent/i)
  })

  it('guides users to bind a ledger/account, preview OCR text, and confirm into that ledger', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input)

      if (url.endsWith('/api/auth/me')) {
        return new Response(
          JSON.stringify({ code: 'OK', message: 'Success', data: { email: 'shortcut@example.com' } }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.endsWith('/api/ledgers')) {
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: 'Success',
            data: { items: [{ id: 'ledger-daily', name: 'Daily ledger', is_default: true }], pagination: { page: 1, page_size: 20, total: 1, total_pages: 1 } },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.endsWith('/api/accounts')) {
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: 'Success',
            data: { items: [{ id: 'acct-wechat', name: 'WeChat Wallet', type: 'cash', initial_balance: 0 }], pagination: { page: 1, page_size: 20, total: 1, total_pages: 1 } },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.endsWith('/api/categories')) {
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: 'Success',
            data: { items: [{ id: 'cat-food', name: 'Food' }], pagination: { page: 1, page_size: 20, total: 1, total_pages: 1 } },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.endsWith('/api/shortcuts/generate')) {
        expect(JSON.parse(String(init?.body))).toEqual(expect.objectContaining({
          default_ledger_id: 'ledger-daily',
          default_account_id: 'acct-wechat',
          mode: 'ocr_confirm',
        }))
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: 'Success',
            data: {
              shortcut_id: 'sc-ocr',
              pat_token: 'pat.ocr.token',
              api_endpoint: 'http://127.0.0.1',
              default_ledger_id: 'ledger-daily',
              default_account_id: 'acct-wechat',
              install_url: 'shortcuts://import-shortcut?name=Xledger',
            },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.endsWith('/api/quick-add/preview')) {
        expect(init?.headers).toEqual(expect.objectContaining({ Authorization: 'Bearer pat.ocr.token' }))
        expect(JSON.parse(String(init?.body))).toEqual(expect.objectContaining({
          shortcut_id: 'sc-ocr',
          default_ledger_id: 'ledger-daily',
          default_account_id: 'acct-wechat',
        }))
        return new Response(
          JSON.stringify({
            code: 'OK',
            message: 'Success',
            data: {
              shortcut_id: 'sc-ocr',
              amount: 35,
              type: 'expense',
              memo: 'Luckin Coffee · WeChat Pay',
              occurred_at: '2026-06-01T04:30:00Z',
              category_suggestion: { id: 'cat-food', name: 'Food', reason: 'OCR', confidence: 0.82 },
              ledger_suggestions: [{ id: 'ledger-daily', name: 'Daily ledger', reason: 'Shortcut setup' }],
              account_suggestions: [{ id: 'acct-wechat', name: 'WeChat Wallet', reason: 'Shortcut setup' }],
              needs_review: false,
            },
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url.endsWith('/api/quick-add/confirm')) {
        expect(init?.headers).toEqual(expect.objectContaining({ Authorization: 'Bearer pat.ocr.token' }))
        expect(JSON.parse(String(init?.body))).toEqual(expect.objectContaining({
          ledger_id: 'ledger-daily',
          account_id: 'acct-wechat',
          category_id: 'cat-food',
          amount: 35,
        }))
        return new Response(
          JSON.stringify({ code: 'OK', message: 'Success', data: { id: 'txn-ocr', success: true, amount: 35, type: 'expense' } }),
          { status: 201, headers: { 'Content-Type': 'application/json' } },
        )
      }

      throw new Error(`Unexpected URL: ${url}`)
    })
    global.fetch = fetchMock as typeof fetch

    renderShortcutApp()
    const user = userEvent.setup()

    const ledgerSelect = await screen.findByRole('combobox', { name: /default ledger/i })
    const accountSelect = await screen.findByRole('combobox', { name: /default account/i })
    expect(ledgerSelect).toHaveValue('ledger-daily')
    expect(accountSelect).toHaveValue('acct-wechat')

    await user.click(screen.getByRole('button', { name: /generate ocr shortcut/i }))
    expect(await screen.findByText(/shortcuts:\/\/import-shortcut/i)).toBeInTheDocument()

    const ocrTextInput = screen.getByRole('textbox', { name: /ocr text/i })
    await user.clear(ocrTextInput)
    await user.type(ocrTextInput, 'WeChat Pay\nMerchant: Luckin Coffee\nAmount ¥35.00')
    await user.click(screen.getByRole('button', { name: /preview ocr entry/i }))

    expect(await screen.findByText(/Luckin Coffee · WeChat Pay/i)).toBeInTheDocument()
    expect(screen.getAllByText(/Daily ledger/i).length).toBeGreaterThan(0)
    expect(screen.getAllByText(/WeChat Wallet/i).length).toBeGreaterThan(0)

    await user.click(screen.getByRole('button', { name: /confirm entry/i }))
    expect(await screen.findByRole('status')).toHaveTextContent(/entry saved/i)
  })
})
