import { useEffect, useMemo, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { useAuth } from '@/features/auth/auth-context'
import { useManagementOverview } from '@/features/management/management-hooks'
import { useTransactionsWithOptions } from '@/features/transactions/transactions-hooks'

function getDismissKey(email: string) {
  return `xledger.first-login-onboarding.dismissed:${email.toLowerCase()}`
}

function getItemCount(data?: { items?: unknown }) {
  return Array.isArray(data?.items) ? data.items.length : 0
}

export function FirstLoginOnboarding() {
  const location = useLocation()
  const navigate = useNavigate()
  const { session } = useAuth()
  const { accountsQuery, ledgersQuery } = useManagementOverview()
  const transactionsQuery = useTransactionsWithOptions({ page: 1, pageSize: 1 })
  const [dismissed, setDismissed] = useState(true)

  useEffect(() => {
    if (!session?.email) {
      setDismissed(true)
      return
    }
    const key = getDismissKey(session.email)
    setDismissed(window.localStorage.getItem(key) === '1')
  }, [session?.email])

  const shouldEvaluate = !['/login', '/auth/google/callback'].includes(location.pathname)
  const accountsCount = getItemCount(accountsQuery.data)
  const ledgersCount = getItemCount(ledgersQuery.data)
  const transactionsCount = getItemCount(transactionsQuery.data)

  const isReady =
    shouldEvaluate &&
    !accountsQuery.isLoading &&
    !ledgersQuery.isLoading &&
    !transactionsQuery.isLoading &&
    !accountsQuery.isError &&
    !ledgersQuery.isError &&
    !transactionsQuery.isError

  const shouldShow = useMemo(() => {
    if (!session?.email || dismissed || !isReady) {
      return false
    }
    return accountsCount === 0 && transactionsCount === 0 && ledgersCount <= 1
  }, [accountsCount, dismissed, isReady, ledgersCount, session?.email, transactionsCount])

  function dismiss() {
    if (session?.email) {
      window.localStorage.setItem(getDismissKey(session.email), '1')
    }
    setDismissed(true)
  }

  if (!shouldShow) {
    return null
  }

  return (
    <div className="fixed inset-0 z-[90] flex items-center justify-center bg-black/45 p-4">
      <div className="w-full max-w-xl rounded-3xl border border-outline/15 bg-white p-6 shadow-ambient md:p-7">
        <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-primary">First Login Guide</p>
        <h2 className="mt-3 font-headline text-3xl font-extrabold tracking-tight text-on-surface">Getting Started</h2>
        <p className="mt-3 text-sm text-on-surface-variant">
          Build your setup in 2 steps: create an account first, then add your first transaction to unlock dashboard and analytics.
        </p>

        <div className="mt-5 rounded-2xl border border-outline/10 bg-surface-container-low p-4 text-sm text-on-surface-variant">
          Tip: if email verification is unavailable in local dev, use password registration/login or Google sign-in.
        </div>

        <div className="mt-6 grid gap-2 sm:grid-cols-3">
          <Button
            className="px-3 py-2 text-sm"
            onClick={() => {
              dismiss()
              void navigate('/accounts')
            }}
          >
            Set Up Accounts
          </Button>
          <Button
            className="px-3 py-2 text-sm"
            variant="secondary"
            onClick={() => {
              dismiss()
              void navigate('/transactions')
            }}
          >
            Add Transaction
          </Button>
          <Button className="px-3 py-2 text-sm" variant="ghost" onClick={dismiss}>
            Skip for now
          </Button>
        </div>
      </div>
    </div>
  )
}
