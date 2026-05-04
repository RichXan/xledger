import { clearAuthSession, readAuthSession, writeAuthSession, type AuthSession } from '@/features/auth/auth-storage'
import { getCurrentLanguage } from '@/i18n'

export const API_BASE_URL = '/api'

const AUTH_UNAUTHORIZED_CODES = new Set(['AUTH_UNAUTHORIZED', 'AUTH_REQUIRED'])
const REFRESH_EXCLUDED_PATHS = new Set([
  '/auth/login',
  '/auth/logout',
  '/auth/refresh',
  '/auth/register',
  '/auth/send-code',
  '/auth/verify-code',
])
let refreshInFlight: Promise<AuthSession | null> | null = null

export interface Envelope<T> {
  code: string
  message: string
  data: T
}

export class ApiError extends Error {
  status: number
  code: string

  constructor(message: string, status: number, code = 'UNKNOWN_ERROR') {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

async function parseJson<T>(response: Response): Promise<T> {
  const text = await response.text()
  if (!text) {
    return {} as T
  }

  try {
    return JSON.parse(text) as T
  } catch (e) {
    // Surface JSON parse errors for better debugging.
    console.error('api: failed to parse JSON response:', e, 'body:', text.slice(0, 200))
    throw new ApiError('Invalid server response', response.status, 'PARSE_ERROR')
  }
}

function headersToRecord(headers?: HeadersInit): Record<string, string> {
  if (!headers) {
    return {}
  }

  if (headers instanceof Headers) {
    const record: Record<string, string> = {}
    headers.forEach((value, key) => {
      record[key] = value
    })
    return record
  }

  if (Array.isArray(headers)) {
    return Object.fromEntries(headers)
  }

  return { ...headers }
}

function getHeader(headers: Record<string, string>, name: string) {
  const key = Object.keys(headers).find((headerName) => headerName.toLowerCase() === name.toLowerCase())
  return key ? headers[key] : null
}

function setHeader(headers: Record<string, string>, name: string, value: string) {
  const key = Object.keys(headers).find((headerName) => headerName.toLowerCase() === name.toLowerCase()) ?? name
  headers[key] = value
}

function buildRequestInit(init?: RequestInit, accessToken?: string): RequestInit {
  const isFormData = typeof FormData !== 'undefined' && init?.body instanceof FormData
  const headers = headersToRecord(init?.headers)

  if (!isFormData && !getHeader(headers, 'Content-Type')) {
    setHeader(headers, 'Content-Type', 'application/json')
  }
  setHeader(headers, 'Accept-Language', getCurrentLanguage())

  if (accessToken) {
    setHeader(headers, 'Authorization', `Bearer ${accessToken}`)
  }

  return {
    ...init,
    headers,
  }
}

function fetchApi(input: string, init?: RequestInit, accessToken?: string) {
  return fetch(`${API_BASE_URL}${input}`, buildRequestInit(init, accessToken))
}

function hasBearerAuthorization(init?: RequestInit) {
  return getHeader(headersToRecord(init?.headers), 'Authorization')?.startsWith('Bearer ') ?? false
}

function shouldRefreshToken(input: string, init: RequestInit | undefined, status: number, code?: string) {
  return (
    status === 401 &&
    Boolean(code && AUTH_UNAUTHORIZED_CODES.has(code)) &&
    !REFRESH_EXCLUDED_PATHS.has(input) &&
    hasBearerAuthorization(init) &&
    Boolean(readAuthSession()?.refreshToken)
  )
}

interface RefreshResponse {
  access_token: string
  refresh_token: string
}

async function refreshStoredSession() {
  if (refreshInFlight) {
    return refreshInFlight
  }

  const currentSession = readAuthSession()
  if (!currentSession?.refreshToken) {
    return null
  }

  refreshInFlight = (async () => {
    try {
      const response = await fetchApi('/auth/refresh', {
        method: 'POST',
        body: JSON.stringify({ refresh_token: currentSession.refreshToken }),
      })
      const payload = await parseJson<Envelope<RefreshResponse>>(response)

      if (!response.ok || !payload.data?.access_token || !payload.data?.refresh_token) {
        clearAuthSession()
        return null
      }

      const nextSession: AuthSession = {
        ...currentSession,
        accessToken: payload.data.access_token,
        refreshToken: payload.data.refresh_token,
      }
      writeAuthSession(nextSession)
      return nextSession
    } catch {
      clearAuthSession()
      return null
    } finally {
      refreshInFlight = null
    }
  })()

  return refreshInFlight
}

interface ParsedError {
  message: string
  code: string
}

async function parseError(response: Response): Promise<ParsedError> {
  const text = await response.text()
  let message = 'Request failed'
  let code = 'UNKNOWN_ERROR'

  if (text) {
    try {
      const parsed = JSON.parse(text) as { code?: string; message?: string; error_code?: string }
      message = parsed.message ?? parsed.error_code ?? message
      code = parsed.code ?? parsed.error_code ?? code
    } catch {
      // Use defaults for non-JSON error bodies.
    }
  }

  return { message, code }
}

export async function requestEnvelope<T>(input: string, init?: RequestInit): Promise<T> {
  let response = await fetchApi(input, init)
  const payload = await parseJson<Envelope<T>>(response)

  if (!response.ok && shouldRefreshToken(input, init, response.status, payload.code)) {
    const refreshedSession = await refreshStoredSession()
    if (refreshedSession) {
      response = await fetchApi(input, init, refreshedSession.accessToken)
      const retryPayload = await parseJson<Envelope<T>>(response)
      if (!response.ok) {
        throw new ApiError(retryPayload.message ?? 'Request failed', response.status, retryPayload.code)
      }
      return retryPayload.data
    }
  }

  if (!response.ok) {
    throw new ApiError(payload.message ?? 'Request failed', response.status, payload.code)
  }

  return payload.data
}

export async function requestRaw<T>(input: string, init?: RequestInit): Promise<T> {
  let response = await fetchApi(input, init)

  if (!response.ok) {
    const error = await parseError(response)
    if (shouldRefreshToken(input, init, response.status, error.code)) {
      const refreshedSession = await refreshStoredSession()
      if (refreshedSession) {
        response = await fetchApi(input, init, refreshedSession.accessToken)
        if (!response.ok) {
          const retryError = await parseError(response)
          throw new ApiError(retryError.message, response.status, retryError.code)
        }
        return parseJson<T>(response)
      }
    }

    throw new ApiError(error.message, response.status, error.code)
  }

  return parseJson<T>(response)
}
