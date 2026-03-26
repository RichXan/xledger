import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type PropsWithChildren,
} from 'react'
import { getCurrentUser, logout as logoutRequest, refreshSession, sendCode, verifyCode } from './auth-api'
import { clearAuthSession, readAuthSession, writeAuthSession, type AuthSession } from './auth-storage'

interface AuthContextValue {
  session: AuthSession | null
  isAuthenticated: boolean
  isBootstrapping: boolean
  sendVerificationCode: (email: string) => Promise<void>
  verifyVerificationCode: (email: string, code: string) => Promise<void>
  applyOAuthTokens: (accessToken: string, refreshToken: string) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: PropsWithChildren) {
  const [session, setSession] = useState<AuthSession | null>(() => readAuthSession())
  const [isBootstrapping, setIsBootstrapping] = useState(true)

  const persistSession = useCallback((nextSession: AuthSession | null) => {
    setSession(nextSession)
    if (nextSession) {
      writeAuthSession(nextSession)
    } else {
      clearAuthSession()
    }
  }, [])

  useEffect(() => {
    let isMounted = true

    async function bootstrap() {
      if (!session?.accessToken) {
        setIsBootstrapping(false)
        return
      }

      try {
        const user = await getCurrentUser(session.accessToken)
        if (isMounted && (session.email !== user.email || session.name !== (user.name ?? null))) {
          persistSession({ ...session, email: user.email, name: user.name ?? null })
        }
      } catch {
        if (!session.refreshToken) {
          if (isMounted) {
            persistSession(null)
          }
          return
        }

        try {
          const tokens = await refreshSession(session.refreshToken)
          const user = await getCurrentUser(tokens.access_token)
          if (isMounted) {
            persistSession({
              accessToken: tokens.access_token,
              refreshToken: tokens.refresh_token,
              email: user.email,
              name: user.name ?? null,
            })
          }
        } catch {
          if (isMounted) {
            persistSession(null)
          }
        }
      } finally {
        if (isMounted) {
          setIsBootstrapping(false)
        }
      }
    }

    void bootstrap()

    return () => {
      isMounted = false
    }
  }, [persistSession, session?.accessToken, session?.email, session?.refreshToken])

  const sendVerificationCode = useCallback(async (email: string) => {
    await sendCode(email)
  }, [])

  const verifyVerificationCode = useCallback(
    async (email: string, code: string) => {
      const tokens = await verifyCode(email, code)
      const user = await getCurrentUser(tokens.access_token)
      persistSession({
        accessToken: tokens.access_token,
        refreshToken: tokens.refresh_token,
        email: user.email,
        name: user.name ?? null,
      })
    },
    [persistSession],
  )

  const applyOAuthTokens = useCallback(
    async (accessToken: string, refreshToken: string) => {
      const user = await getCurrentUser(accessToken)
      persistSession({
        accessToken,
        refreshToken,
        email: user.email,
        name: user.name ?? null,
      })
    },
    [persistSession],
  )

  const logout = useCallback(async () => {
    if (session?.refreshToken) {
      try {
        await logoutRequest(session.refreshToken)
      } catch {
        // swallow network/logout contract issues; local logout still proceeds
      }
    }
    persistSession(null)
  }, [persistSession, session?.refreshToken])

  const value = useMemo<AuthContextValue>(
    () => ({
      session,
      isAuthenticated: Boolean(session?.accessToken),
      isBootstrapping,
      sendVerificationCode,
      verifyVerificationCode,
      applyOAuthTokens,
      logout,
    }),
    [applyOAuthTokens, isBootstrapping, logout, sendVerificationCode, session, verifyVerificationCode],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return context
}
