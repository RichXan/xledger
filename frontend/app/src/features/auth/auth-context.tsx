import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  useRef,
  type PropsWithChildren,
} from 'react'
import {
  changePassword as changePasswordRequest,
  getCurrentUser,
  loginWithPassword,
  logout as logoutRequest,
  refreshSession,
  registerWithPassword,
  sendCode,
  updateProfile,
  verifyCode,
} from './auth-api'
import { clearAuthSession, readAuthSession, writeAuthSession, type AuthSession } from './auth-storage'
import { changeLanguage, supportedLanguages } from '@/i18n'

interface AuthContextValue {
  session: AuthSession | null
  isAuthenticated: boolean
  isBootstrapping: boolean
  sendVerificationCode: (email: string) => Promise<void>
  verifyVerificationCode: (email: string, code: string) => Promise<void>
  loginWithPassword: (email: string, password: string) => Promise<void>
  registerWithPassword: (email: string, password: string, displayName?: string) => Promise<void>
  updateDisplayName: (displayName: string) => Promise<void>
  changePassword: (oldPassword: string, newPassword: string) => Promise<void>
  applyOAuthTokens: (accessToken: string, refreshToken: string) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: PropsWithChildren) {
  const [session, setSession] = useState<AuthSession | null>(() => readAuthSession())
  const [isBootstrapping, setIsBootstrapping] = useState(true)
  const isRefreshing = useRef<Promise<void> | null>(null)

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
        // Sync language preference from backend (once backend supports it)
        if (isMounted && user.language && supportedLanguages.includes(user.language as 'en' | 'zh')) {
          changeLanguage(user.language as 'en' | 'zh')
        }
      } catch {
        if (!session.refreshToken) {
          if (isMounted) {
            persistSession(null)
          }
          setIsBootstrapping(false)
          return
        }

        // If a refresh is already in progress, wait for it instead of starting another
        if (isRefreshing.current) {
          try {
            await isRefreshing.current
          } catch {
            // refresh failed, session was cleared by the in-progress refresh
          }
          setIsBootstrapping(false)
          return
        }

        // Start a new refresh and store the promise so concurrent requests can await it
        let refreshCompleted = false
        isRefreshing.current = (async () => {
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
              // Sync language preference
              if (user.language && supportedLanguages.includes(user.language as 'en' | 'zh')) {
                changeLanguage(user.language as 'en' | 'zh')
              }
            }
            refreshCompleted = true
          } catch {
            if (isMounted) {
              persistSession(null)
            }
            refreshCompleted = true
          } finally {
            if (isMounted) {
              isRefreshing.current = null
            }
          }
        })()

        await isRefreshing.current
        if (!refreshCompleted && isMounted) {
          persistSession(null)
        }
      }
      if (isMounted) {
        setIsBootstrapping(false)
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
      // Sync language preference
      if (user.language && supportedLanguages.includes(user.language as 'en' | 'zh')) {
        changeLanguage(user.language as 'en' | 'zh')
      }
    },
    [persistSession],
  )

  const loginWithPasswordFn = useCallback(
    async (email: string, password: string) => {
      const tokens = await loginWithPassword(email, password)
      const user = await getCurrentUser(tokens.access_token)
      persistSession({
        accessToken: tokens.access_token,
        refreshToken: tokens.refresh_token,
        email: user.email,
        name: user.name ?? null,
      })
      // Sync language preference
      if (user.language && supportedLanguages.includes(user.language as 'en' | 'zh')) {
        changeLanguage(user.language as 'en' | 'zh')
      }
    },
    [persistSession],
  )

  const registerWithPasswordFn = useCallback(
    async (email: string, password: string, displayName?: string) => {
      const tokens = await registerWithPassword({ email, password, displayName })
      const user = await getCurrentUser(tokens.access_token)
      persistSession({
        accessToken: tokens.access_token,
        refreshToken: tokens.refresh_token,
        email: user.email,
        name: user.name ?? null,
      })
      // Sync language preference
      if (user.language && supportedLanguages.includes(user.language as 'en' | 'zh')) {
        changeLanguage(user.language as 'en' | 'zh')
      }
    },
    [persistSession],
  )

  const updateDisplayNameFn = useCallback(
    async (displayName: string) => {
      if (!session?.accessToken) throw new Error('Missing access token')
      const user = await updateProfile(session.accessToken, displayName)
      persistSession({
        accessToken: session.accessToken,
        refreshToken: session.refreshToken,
        email: user.email,
        name: user.name ?? null,
      })
    },
    [persistSession, session?.accessToken, session?.refreshToken],
  )

  const changePasswordFn = useCallback(
    async (oldPassword: string, newPassword: string) => {
      if (!session?.accessToken) throw new Error('Missing access token')
      await changePasswordRequest(session.accessToken, oldPassword, newPassword)
    },
    [session?.accessToken],
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
      // Sync language preference
      if (user.language && supportedLanguages.includes(user.language as 'en' | 'zh')) {
        changeLanguage(user.language as 'en' | 'zh')
      }
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
      loginWithPassword: loginWithPasswordFn,
      registerWithPassword: registerWithPasswordFn,
      updateDisplayName: updateDisplayNameFn,
      changePassword: changePasswordFn,
      applyOAuthTokens,
      logout,
    }),
    [
      applyOAuthTokens,
      changePasswordFn,
      isBootstrapping,
      loginWithPasswordFn,
      logout,
      registerWithPasswordFn,
      sendVerificationCode,
      session,
      updateDisplayNameFn,
      verifyVerificationCode,
    ],
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
