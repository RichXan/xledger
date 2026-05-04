import { requestEnvelope } from './lib/api'

const originalFetch = global.fetch

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

function getHeader(init: RequestInit | undefined, name: string) {
  return new Headers(init?.headers).get(name)
}

describe('api token refresh', () => {
  afterEach(() => {
    global.fetch = originalFetch
    window.localStorage.clear()
  })

  it('refreshes an expired access token and retries the original request once', async () => {
    window.localStorage.setItem(
      'xledger.auth',
      JSON.stringify({
        accessToken: 'old.access',
        refreshToken: 'refresh.token',
        email: 'demo@example.com',
        name: 'Demo User',
      }),
    )

    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input)

      if (url.endsWith('/api/accounts')) {
        const authorization = getHeader(init, 'Authorization')

        if (authorization === 'Bearer old.access') {
          return jsonResponse(
            {
              code: 'AUTH_UNAUTHORIZED',
              message: '未认证或凭证无效',
              data: null,
            },
            401,
          )
        }

        if (authorization === 'Bearer new.access') {
          return jsonResponse({
            code: 'OK',
            message: '成功',
            data: { items: [] },
          })
        }
      }

      if (url.endsWith('/api/auth/refresh')) {
        expect(init?.method).toBe('POST')
        expect(JSON.parse(String(init?.body))).toEqual({ refresh_token: 'refresh.token' })
        return jsonResponse({
          code: 'OK',
          message: '成功',
          data: {
            access_token: 'new.access',
            refresh_token: 'new.refresh',
          },
        })
      }

      throw new Error(`Unexpected request: ${url}`)
    })
    global.fetch = fetchMock as typeof fetch

    const result = await requestEnvelope<{ items: unknown[] }>('/accounts', {
      headers: { Authorization: 'Bearer old.access' },
    })

    expect(result.items).toEqual([])
    expect(fetchMock).toHaveBeenCalledTimes(3)
    expect(String(fetchMock.mock.calls[0]?.[0])).toMatch('/api/accounts')
    expect(String(fetchMock.mock.calls[1]?.[0])).toMatch('/api/auth/refresh')
    expect(String(fetchMock.mock.calls[2]?.[0])).toMatch('/api/accounts')
    expect(getHeader(fetchMock.mock.calls[2]?.[1], 'Authorization')).toBe('Bearer new.access')
    expect(window.localStorage.getItem('xledger.auth')).toContain('new.access')
    expect(window.localStorage.getItem('xledger.auth')).toContain('new.refresh')
  })
})
