import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { AuthProvider } from '@/features/auth/auth-context'
import { AppShell } from '@/components/layout/app-shell'
import { MetricCard } from '@/components/ui/metric-card'
import { PageSection } from '@/components/ui/page-section'
import { TextField } from '@/components/ui/text-field'

function renderShell(content: React.ReactNode) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })

  return render(
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <MemoryRouter>
          <AppShell>{content}</AppShell>
        </MemoryRouter>
      </AuthProvider>
    </QueryClientProvider>,
  )
}

describe('design system primitives', () => {
  it('renders page sections and metric cards with stitch-inspired content', () => {
    renderShell(
      <PageSection
        eyebrow="Architectural Ledger"
        title="Precision Modules"
        description="A composable surface for dashboard, transactions, and settings pages."
      >
        <MetricCard label="Net Cash" value="¥24,609.50" tone="positive" delta="+12.4%" />
      </PageSection>,
    )

    expect(screen.getByRole('heading', { name: /precision modules/i })).toBeInTheDocument()
    expect(screen.getByText(/a composable surface for dashboard, transactions, and settings pages/i)).toBeInTheDocument()
    expect(screen.getByText(/net cash/i)).toBeInTheDocument()
    expect(screen.getAllByText(/\+12.4%/i).length).toBeGreaterThan(0)
  })

  it('renders underline text fields with helper and error states', () => {
    render(
      <TextField
        label="Ledger search"
        helperText="Search vendors, notes, or transaction IDs."
        error="Search term is required"
      />,
    )

    expect(screen.getByLabelText(/ledger search/i)).toBeInTheDocument()
    expect(screen.getByText(/search vendors, notes, or transaction ids/i)).toBeInTheDocument()
    expect(screen.getByText(/search term is required/i)).toBeInTheDocument()
  })
})
