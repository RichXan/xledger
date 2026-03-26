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
  return text ? (JSON.parse(text) as T) : ({} as T)
}

export async function requestEnvelope<T>(input: string, init?: RequestInit): Promise<T> {
  const isFormData = typeof FormData !== 'undefined' && init?.body instanceof FormData
  const response = await fetch(`${API_BASE_URL}${input}`, {
    headers: {
      ...(isFormData ? {} : { 'Content-Type': 'application/json' }),
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
  const response = await fetch(`${API_BASE_URL}${input}`, {
    headers: {
      ...(isFormData ? {} : { 'Content-Type': 'application/json' }),
      ...(init?.headers ?? {}),
    },
    ...init,
  })
  const payload = await parseJson<T & { error_code?: string }>(response)

  if (!response.ok) {
    const code = typeof payload === 'object' && payload && 'error_code' in payload ? payload.error_code ?? 'UNKNOWN_ERROR' : 'UNKNOWN_ERROR'
    throw new ApiError('Request failed', response.status, code)
  }

  return payload
}
