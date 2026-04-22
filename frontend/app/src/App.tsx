import { Suspense, lazy, type JSX } from 'react'
import { Navigate, Route, Routes } from 'react-router-dom'
import { AppShell } from '@/components/layout/app-shell'
import { RequireAuth } from '@/features/auth/require-auth'

const DashboardPage = lazy(async () => {
  const module = await import('@/pages/dashboard-page')
  return { default: module.DashboardPage }
})
const AnalyticsPage = lazy(async () => {
  const module = await import('@/pages/analytics-page')
  return { default: module.AnalyticsPage }
})
const AccountsPage = lazy(async () => {
  const module = await import('@/pages/accounts-page')
  return { default: module.AccountsPage }
})
const LoginPage = lazy(async () => {
  const module = await import('@/pages/login-page')
  return { default: module.LoginPage }
})
const SettingsPage = lazy(async () => {
  const module = await import('@/pages/settings-page')
  return { default: module.SettingsPage }
})
const TransactionsPage = lazy(async () => {
  const module = await import('@/pages/transactions-page')
  return { default: module.TransactionsPage }
})
const GoogleCallbackPage = lazy(async () => {
  const module = await import('@/pages/google-callback-page')
  return { default: module.GoogleCallbackPage }
})
const ShortcutPage = lazy(async () => {
  const module = await import('@/pages/shortcut-page')
  return { default: module.ShortcutPage }
})
const PWAOnboardingPage = lazy(async () => {
  const module = await import('@/pages/onboarding-pwa-page')
  return { default: module.PWAOnboardingPage }
})
const ImportPage = lazy(async () => {
  const module = await import('@/pages/import-page')
  return { default: module.ImportPage }
})

function ProtectedLayout({ children }: { children: JSX.Element }) {
  return <RequireAuth>{children}</RequireAuth>
}

function RouteFallback() {
  return (
    <div className="flex min-h-[40vh] items-center justify-center text-sm text-on-surface-variant">
      Loading...
    </div>
  )
}

export default function App() {
  return (
    <Suspense fallback={<RouteFallback />}>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/auth/google/callback" element={<GoogleCallbackPage />} />
        <Route path="/pwa-onboarding" element={<PWAOnboardingPage />} />
        <Route
          path="/dashboard"
          element={
            <ProtectedLayout>
              <AppShell>
                <DashboardPage />
              </AppShell>
            </ProtectedLayout>
          }
        />
        <Route
          path="/transactions"
          element={
            <ProtectedLayout>
              <AppShell>
                <TransactionsPage />
              </AppShell>
            </ProtectedLayout>
          }
        />
        <Route
          path="/analytics"
          element={
            <ProtectedLayout>
              <AppShell>
                <AnalyticsPage />
              </AppShell>
            </ProtectedLayout>
          }
        />
        <Route
          path="/accounts"
          element={
            <ProtectedLayout>
              <AppShell>
                <AccountsPage />
              </AppShell>
            </ProtectedLayout>
          }
        />
        <Route
          path="/shortcut"
          element={
            <ProtectedLayout>
              <AppShell>
                <ShortcutPage />
              </AppShell>
            </ProtectedLayout>
          }
        />
        <Route
          path="/import"
          element={
            <ProtectedLayout>
              <AppShell>
                <ImportPage />
              </AppShell>
            </ProtectedLayout>
          }
        />
        <Route
          path="/settings"
          element={
            <ProtectedLayout>
              <AppShell>
                <SettingsPage />
              </AppShell>
            </ProtectedLayout>
          }
        />
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>
    </Suspense>
  )
}
