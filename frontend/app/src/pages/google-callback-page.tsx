import { useEffect, useState, useRef } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useAuth } from '@/features/auth/auth-context'
import { ApiError, requestEnvelope } from '@/lib/api'

interface ExchangeCodeResponse {
  access_token: string
  refresh_token: string
}

function messageForOAuthErrorReason(errorReason: string) {
  switch (errorReason) {
    case 'oauth_code_invalid_or_expired':
      return 'Google 授权码已失效或已被使用，请回到登录页重新点击 Google 登录。'
    case 'google_oauth_not_configured':
      return 'Google OAuth 尚未正确配置，请检查本地客户端 ID、密钥和回调地址。'
    case 'google_token_exchange_failed':
      return 'Google 登录回调地址不匹配或授权换取令牌失败，请检查 Google Console 的 redirect URI 配置。'
    case 'oauth_state_invalid_or_expired':
      return '本次 Google 登录会话已失效，请返回登录页重新发起授权。'
    case 'oauth_callback_invalid':
      return 'Google 登录回调参数不完整，请重新发起授权。'
    case 'google_profile_fetch_failed':
      return '已获取 Google 授权，但读取账户信息失败，请确认该 Google 账号可正常返回邮箱资料。'
    case 'google_email_missing':
      return 'Google 账号未返回可用邮箱，无法完成登录。'
    default:
      return 'Google 登录失败，请重试。'
  }
}

export function GoogleCallbackPage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const { applyOAuthTokens } = useAuth()
  const [error, setError] = useState<string | null>(null)
  const exchangedRef = useRef(false)

  useEffect(() => {
    if (exchangedRef.current) return
    exchangedRef.current = true

    const exchangeCode = (searchParams.get('exchange_code') ?? '').trim()
    const errorCode = (searchParams.get('error_code') ?? '').trim()
    const errorReason = (searchParams.get('error_reason') ?? '').trim()

    async function bootstrapFromOAuth() {
      if (errorCode) {
        setError(messageForOAuthErrorReason(errorReason))
        return
      }
      if (!exchangeCode) {
        setError('Google login failed: missing exchange code.')
        return
      }
      try {
        const tokens = await requestEnvelope<ExchangeCodeResponse>('/auth/google/exchange-code', {
          method: 'POST',
          body: JSON.stringify({ code: exchangeCode }),
        })
        await applyOAuthTokens(tokens.access_token, tokens.refresh_token)
        navigate('/dashboard', { replace: true })
      } catch (caughtError) {
        if (caughtError instanceof ApiError) {
          setError(caughtError.message)
        } else {
          setError('Google login failed. Please try again.')
        }
      }
    }

    void bootstrapFromOAuth()
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
