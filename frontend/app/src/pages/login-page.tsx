import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { TextField } from '@/components/ui/text-field'
import { useAuth } from '@/features/auth/auth-context'
import { ApiError } from '@/lib/api'

export function LoginPage() {
  const navigate = useNavigate()
  const { sendVerificationCode, verifyVerificationCode } = useAuth()
  const [email, setEmail] = useState('')
  const [code, setCode] = useState('')
  const [codeSent, setCodeSent] = useState(false)
  const [pending, setPending] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function handleGoogleSignIn() {
    const configuredBackendOrigin = (
      (import.meta as ImportMeta & { env?: Record<string, string | undefined> }).env?.VITE_BACKEND_ORIGIN
    )?.trim()
    const fallbackOrigin =
      window.location.port === '4173'
        ? `${window.location.protocol}//${window.location.hostname}:8080`
        : window.location.origin
    const backendOrigin = configuredBackendOrigin || fallbackOrigin
    window.location.href = `${backendOrigin}/api/auth/google`
  }

  async function handleSendCode(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setPending(true)
    setError(null)

    try {
      await sendVerificationCode(email)
      setCodeSent(true)
    } catch (caughtError) {
      if (caughtError instanceof ApiError) {
        setError(caughtError.message)
      } else {
        setError('Unable to send a verification code right now.')
      }
    } finally {
      setPending(false)
    }
  }

  async function handleVerifyCode(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setPending(true)
    setError(null)

    try {
      await verifyVerificationCode(email, code)
      navigate('/dashboard', { replace: true })
    } catch (caughtError) {
      if (caughtError instanceof ApiError) {
        setError(caughtError.message)
      } else {
        setError('Unable to verify your code right now.')
      }
    } finally {
      setPending(false)
    }
  }

  return (
    <div className="min-h-screen bg-background text-on-surface lg:grid lg:grid-cols-[1.2fr_0.8fr]">
      <section className="hidden bg-primary px-12 py-16 text-white lg:flex lg:flex-col lg:justify-between">
        <div>
          <p className="font-headline text-4xl font-black tracking-tight">xledger</p>
          <p className="mt-6 max-w-lg font-headline text-5xl font-extrabold leading-tight">
            Precision in every <span className="text-tertiary-fixed">financial perspective.</span>
          </p>
        </div>
        <div className="grid max-w-xl grid-cols-2 gap-4">
          <div className="rounded-[28px] border border-white/10 bg-white/10 p-6 backdrop-blur-md">
            <p className="font-label text-[10px] uppercase tracking-[0.2em] text-primary-fixed">Live cash flow</p>
            <p className="mt-6 font-headline text-4xl font-bold">+24.8%</p>
          </div>
          <div className="rounded-[28px] border border-white/10 bg-white/10 p-6 backdrop-blur-md">
            <p className="font-label text-[10px] uppercase tracking-[0.2em] text-primary-fixed">Synced ledgers</p>
            <p className="mt-6 font-headline text-4xl font-bold">12</p>
          </div>
        </div>
      </section>

      <main className="flex items-center justify-center px-6 py-12 sm:px-10 lg:px-16">
        <div className="w-full max-w-md space-y-8">
          <div>
            <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-primary">Secure access</p>
            <h1 className="mt-3 font-headline text-4xl font-extrabold tracking-tight text-on-surface">
              Welcome back
            </h1>
            <p className="mt-3 text-sm text-on-surface-variant">
              Sign in with your email and verification code to access your ledgers.
            </p>
          </div>

          <form className="space-y-6" onSubmit={codeSent ? handleVerifyCode : handleSendCode}>
            <TextField
              label="Email Address"
              type="email"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              placeholder="name@company.com"
            />

            {codeSent ? (
              <>
                <TextField
                  label="Verification Code"
                  value={code}
                  onChange={(event) => setCode(event.target.value)}
                  placeholder="6-digit code"
                />
                <p className="text-sm text-on-surface-variant">We sent a 6-digit code to {email}.</p>
              </>
            ) : null}

            {error ? <p className="text-sm font-medium text-error">{error}</p> : null}

            <div className="grid gap-3 sm:grid-cols-2">
              <Button className="w-full" type="submit" disabled={pending || !email || (codeSent && !code)}>
                {codeSent ? 'Verify and Continue' : 'Send Verification Code'}
              </Button>
              <Button className="w-full" type="button" variant="secondary" onClick={handleGoogleSignIn}>
                Continue with Google
              </Button>
            </div>
          </form>
        </div>
      </main>
    </div>
  )
}
