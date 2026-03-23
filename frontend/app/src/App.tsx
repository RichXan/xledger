import { Navigate, Route, Routes } from 'react-router-dom'
import { AppShell } from '@/components/layout/app-shell'
import { RequireAuth } from '@/features/auth/require-auth'
import { DashboardPage } from '@/pages/dashboard-page'
import { AnalyticsPage } from '@/pages/analytics-page'
import { AccountsPage } from '@/pages/accounts-page'
import { LoginPage } from '@/pages/login-page'
import { SettingsPage } from '@/pages/settings-page'
import { TransactionsPage } from '@/pages/transactions-page'

function ProtectedLayout({ children }: { children: JSX.Element }) {
  return <RequireAuth>{children}</RequireAuth>
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
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
  )
}
