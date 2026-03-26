import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useAuth } from '@/features/auth/auth-context'
import { ApiError } from '@/lib/api'

export function GoogleCallbackPage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const { applyOAuthTokens } = useAuth()
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let canceled = false
    const accessToken = (searchParams.get('access_token') ?? '').trim()
    const refreshToken = (searchParams.get('refresh_token') ?? '').trim()
    const errorCode = (searchParams.get('error_code') ?? '').trim()
    const errorReason = (searchParams.get('error_reason') ?? '').trim()

    async function bootstrapFromOAuth() {
      if (errorCode) {
        if (errorReason === 'oauth_code_invalid_or_expired') {
          setError('Google 授权码已失效或已被使用，请回到登录页重新点击 Google 登录。')
          return
        }
        setError('Google 登录失败，请重试。')
        return
      }
      if (!accessToken || !refreshToken) {
        setError('Google login failed: missing token payload.')
        return
      }
      try {
        await applyOAuthTokens(accessToken, refreshToken)
        if (!canceled) {
          navigate('/dashboard', { replace: true })
        }
      } catch (caughtError) {
        if (caughtError instanceof ApiError) {
          setError(caughtError.message)
        } else {
          setError('Google login failed. Please try again.')
        }
      }
    }

    void bootstrapFromOAuth()
    return () => {
      canceled = true
    }
  }, [applyOAuthTokens, navigate, searchParams])

  return (
    <main className="flex min-h-screen items-center justify-center bg-background px-6">
      <div className="w-full max-w-md rounded-2xl border border-outline-variant bg-surface p-6">
        <h1 className="font-headline text-2xl font-bold text-on-surface">Completing Google sign-in...</h1>
        {error ? <p className="mt-3 text-sm text-error">{error}</p> : <p className="mt-3 text-sm text-on-surface-variant">Please wait while we finish authentication.</p>}
      </div>
    </main>
  )
}
