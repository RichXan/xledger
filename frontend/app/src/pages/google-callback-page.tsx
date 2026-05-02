import { useEffect, useState, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useAuth } from '@/features/auth/auth-context'
import { ApiError, requestEnvelope } from '@/lib/api'

interface ExchangeCodeResponse {
  access_token: string
  refresh_token: string
}

function messageKeyForOAuthErrorReason(errorReason: string) {
  switch (errorReason) {
    case 'oauth_code_invalid_or_expired':
      return 'oauthCodeInvalidOrExpired'
    case 'google_oauth_not_configured':
      return 'googleOauthNotConfigured'
    case 'google_token_exchange_failed':
      return 'googleTokenExchangeFailed'
    case 'oauth_state_invalid_or_expired':
      return 'oauthStateInvalidOrExpired'
    case 'oauth_callback_invalid':
      return 'oauthCallbackInvalid'
    case 'google_profile_fetch_failed':
      return 'googleProfileFetchFailed'
    case 'google_email_missing':
      return 'googleEmailMissing'
    default:
      return 'defaultFailed'
  }
}

export function GoogleCallbackPage() {
  const { t } = useTranslation()
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
        setError(t(`auth.googleCallback.${messageKeyForOAuthErrorReason(errorReason)}`))
        return
      }
      if (!exchangeCode) {
        setError(t('auth.googleCallback.missingExchangeCode'))
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
          setError(t('auth.googleCallback.failedRetry'))
        }
      }
    }

    void bootstrapFromOAuth()
  }, [applyOAuthTokens, navigate, searchParams, t])

  return (
    <main className="flex min-h-screen items-center justify-center bg-background px-6">
      <div className="w-full max-w-md rounded-2xl border border-outline-variant bg-surface p-6">
        <h1 className="font-headline text-2xl font-bold text-on-surface">{t('auth.googleCallback.completing')}</h1>
        {error ? <p className="mt-3 text-sm text-error">{error}</p> : <p className="mt-3 text-sm text-on-surface-variant">{t('auth.googleCallback.waiting')}</p>}
      </div>
    </main>
  )
}
