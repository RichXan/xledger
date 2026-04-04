import i18n from '@/i18n'

export const API_BASE_URL = '/api'

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
    // Surface JSON parse errors for better debugging
    console.error('api: failed to parse JSON response:', e, 'body:', text.slice(0, 200))
    throw new ApiError('服务器响应格式错误', response.status, 'PARSE_ERROR')
  }
}

export async function requestEnvelope<T>(input: string, init?: RequestInit): Promise<T> {
  const isFormData = typeof FormData !== 'undefined' && init?.body instanceof FormData
  // i18n language detection - safely handle case where i18n may not be initialized yet
  const lang = typeof i18n !== 'undefined' ? i18n.language : 'en'
  const response = await fetch(`${API_BASE_URL}${input}`, {
    headers: {
      ...(isFormData ? {} : { 'Content-Type': 'application/json' }),
      'Accept-Language': lang,
      ...(init?.headers ?? {}),
    },
    ...init,
  })
  const payload = await parseJson<Envelope<T>>(response)

  if (!response.ok) {
    throw new ApiError(payload.message ?? 'Request failed', response.status, payload.code)
  }

  return payload.data
}

export async function requestRaw<T>(input: string, init?: RequestInit): Promise<T> {
  const isFormData = typeof FormData !== 'undefined' && init?.body instanceof FormData
  const lang = typeof i18n !== 'undefined' ? i18n.language : 'en'
  const response = await fetch(`${API_BASE_URL}${input}`, {
    headers: {
      ...(isFormData ? {} : { 'Content-Type': 'application/json' }),
      'Accept-Language': lang,
      ...(init?.headers ?? {}),
    },
    ...init,
  })

  if (!response.ok) {
    const text = await response.text()
    let message = 'Request failed'
    let code = 'UNKNOWN_ERROR'
    if (text) {
      try {
        const parsed = JSON.parse(text) as { code?: string; message?: string; error_code?: string }
        message = parsed.message ?? parsed.error_code ?? message
        code = parsed.code ?? parsed.error_code ?? code
      } catch {
        // use defaults
      }
    }
    throw new ApiError(message, response.status, code)
  }

  const payload = await parseJson<T>(response)
  return payload
}
