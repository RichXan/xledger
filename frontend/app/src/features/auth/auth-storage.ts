export interface AuthSession {
  accessToken: string
  refreshToken: string
  email: string | null
  name?: string | null
}

const STORAGE_KEY = 'xledger.auth'
export const AUTH_SESSION_CHANGED_EVENT = 'xledger.auth.changed'

function notifyAuthSessionChanged(session: AuthSession | null) {
  window.dispatchEvent(new CustomEvent<AuthSession | null>(AUTH_SESSION_CHANGED_EVENT, { detail: session }))
}

export function readAuthSession(): AuthSession | null {
  const raw = window.localStorage.getItem(STORAGE_KEY)
  if (!raw) {
    return null
  }

  try {
    return JSON.parse(raw) as AuthSession
  } catch {
    window.localStorage.removeItem(STORAGE_KEY)
    return null
  }
}

export function writeAuthSession(session: AuthSession) {
  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(session))
  notifyAuthSessionChanged(session)
}

export function clearAuthSession() {
  window.localStorage.removeItem(STORAGE_KEY)
  notifyAuthSessionChanged(null)
}
