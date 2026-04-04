import { Bell, CircleHelp, Download, RefreshCcw, Search } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { Button } from '@/components/ui/button'
import { DialogShell } from '@/components/ui/dialog-shell'
import { TextField } from '@/components/ui/text-field'
import { useAuth } from '@/features/auth/auth-context'
import { usePwaInstall } from '@/features/pwa/use-pwa-install'
import { usePwaUpdate } from '@/features/pwa/use-pwa-update'
import { ApiError } from '@/lib/api'
import { changeLanguage, getCurrentLanguage, supportedLanguages } from '@/i18n'

export function TopBar() {
  const { logout, session, updateDisplayName, changePassword } = useAuth()
  const displayName = useMemo(() => session?.name || session?.email || 'Ledger User', [session?.email, session?.name])
  const [profileOpen, setProfileOpen] = useState(false)
  const [displayNameInput, setDisplayNameInput] = useState(displayName)
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [pending, setPending] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const { canInstall, install } = usePwaInstall()
  const { updateAvailable, updating, updateNow } = usePwaUpdate()
  const [currentLang, setCurrentLang] = useState(getCurrentLanguage())

  const handleLanguageChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const newLang = e.target.value as 'en' | 'zh'
    changeLanguage(newLang)
    setCurrentLang(newLang)
  }

  useEffect(() => {
    setDisplayNameInput(displayName)
  }, [displayName])

  async function handleSaveProfile() {
    setPending(true)
    setError(null)
    setNotice(null)
    try {
      await updateDisplayName(displayNameInput)
      setNotice('Profile updated.')
    } catch (caughtError) {
      if (caughtError instanceof ApiError) {
        setError(caughtError.message)
      } else {
        setError('Failed to update profile.')
      }
    } finally {
      setPending(false)
    }
  }

  async function handleChangePassword() {
    setPending(true)
    setError(null)
    setNotice(null)
    try {
      await changePassword(oldPassword, newPassword)
      setOldPassword('')
      setNewPassword('')
      setNotice('Password updated.')
    } catch (caughtError) {
      if (caughtError instanceof ApiError) {
        setError(caughtError.message)
      } else {
        setError('Failed to change password.')
      }
    } finally {
      setPending(false)
    }
  }

  return (
    <>
      <header className="sticky top-0 z-20 border-b border-outline/15 bg-surface-container-lowest/85 px-4 py-3 backdrop-blur md:px-6">
        <div className="mx-auto flex w-full max-w-[1800px] items-center justify-between gap-4">
          <label className="relative hidden w-full max-w-[460px] lg:block">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-on-surface-variant" />
            <input
              readOnly
              value=""
              placeholder="Search transactions..."
              className="h-11 w-full rounded-xl border border-outline/20 bg-surface-container-low pl-9 pr-3 text-sm text-on-surface placeholder:text-on-surface-variant/70"
            />
          </label>

          <div className="ml-auto flex items-center gap-2.5">
            <select
              value={currentLang}
              onChange={handleLanguageChange}
              className="h-10 rounded-xl border border-outline/20 bg-surface-container-low px-2 text-sm text-on-surface-variant transition hover:bg-surface-container cursor-pointer"
              aria-label="Select language"
            >
              {supportedLanguages.map((lang) => (
                <option key={lang} value={lang}>
                  {lang === 'zh' ? '中文' : 'EN'}
                </option>
              ))}
            </select>
            <button
              type="button"
              className="grid h-10 w-10 place-items-center rounded-xl border border-outline/20 bg-surface-container-low text-on-surface-variant transition hover:bg-surface-container"
              aria-label="Notifications"
            >
              <Bell className="h-4 w-4" />
            </button>
            <button
              type="button"
              className="grid h-10 w-10 place-items-center rounded-xl border border-outline/20 bg-surface-container-low text-on-surface-variant transition hover:bg-surface-container"
              aria-label="Help"
            >
              <CircleHelp className="h-4 w-4" />
            </button>
            {canInstall ? (
              <Button
                variant="secondary"
                className="h-10 px-3"
                onClick={() => {
                  void install()
                }}
              >
                <Download className="mr-1 h-4 w-4" />
                Install App
              </Button>
            ) : null}
            {updateAvailable ? (
              <Button variant="secondary" className="h-10 px-3" onClick={() => void updateNow()} disabled={updating}>
                <RefreshCcw className="mr-1 h-4 w-4" />
                {updating ? 'Updating...' : 'Update App'}
              </Button>
            ) : null}
            <button
              type="button"
              className="hidden items-center gap-3 rounded-xl bg-surface-container-low px-3 py-2 sm:flex"
              onClick={() => {
                setProfileOpen(true)
                setDisplayNameInput(displayName)
                setError(null)
                setNotice(null)
              }}
            >
              <div className="text-right leading-tight">
                <p className="text-sm font-semibold text-on-surface">{displayName}</p>
              </div>
              <div className="grid h-8 w-8 place-items-center rounded-lg bg-primary text-xs font-bold text-white">
                {(displayName[0] ?? 'U').toUpperCase()}
              </div>
            </button>
            <Button variant="ghost" className="px-3 py-2" onClick={() => void logout()}>
              Logout
            </Button>
          </div>
        </div>
      </header>

      {profileOpen ? (
        <DialogShell
          title="Profile"
          description="Review your account details, update display name, and change password."
          onClose={() => setProfileOpen(false)}
          className="max-w-2xl"
          footer={
            <>
              <Button variant="secondary" onClick={() => setProfileOpen(false)}>
                Cancel
              </Button>
              <Button onClick={() => void handleSaveProfile()} disabled={pending || !displayNameInput.trim()}>
                Save Profile
              </Button>
            </>
          }
        >
          <div className="space-y-6">
            <div className="grid gap-4 md:grid-cols-2">
              <TextField label="Email" value={session?.email ?? ''} disabled />
              <TextField
                label="Display Name"
                value={displayNameInput}
                onChange={(event) => setDisplayNameInput(event.target.value)}
                placeholder="Your name"
              />
            </div>
            <div className="rounded-2xl border border-outline/15 bg-surface-container-low p-4">
              <p className="text-xs font-bold uppercase tracking-[0.14em] text-on-surface-variant">Change Password</p>
              <div className="mt-3 grid gap-3 md:grid-cols-2">
                <TextField
                  label="Current Password"
                  type="password"
                  value={oldPassword}
                  onChange={(event) => setOldPassword(event.target.value)}
                />
                <TextField
                  label="New Password"
                  type="password"
                  value={newPassword}
                  onChange={(event) => setNewPassword(event.target.value)}
                />
              </div>
              <div className="mt-3">
                <Button
                  variant="secondary"
                  onClick={() => void handleChangePassword()}
                  disabled={pending || !oldPassword || newPassword.length < 8}
                >
                  Update Password
                </Button>
              </div>
            </div>
            {error ? <p className="text-sm font-medium text-error">{error}</p> : null}
            {notice ? <p className="text-sm font-medium text-emerald-700">{notice}</p> : null}
          </div>
        </DialogShell>
      ) : null}
    </>
  )
}
