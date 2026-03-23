import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { DialogShell } from '@/components/ui/dialog-shell'
import { PageSection } from '@/components/ui/page-section'
import { SelectField } from '@/components/ui/select-field'
import { TextField } from '@/components/ui/text-field'
import { useCreateAccount, useManagementOverview } from '@/features/management/management-hooks'
import { formatCurrency } from '@/lib/format'

export function AccountsPage() {
  const [showDialog, setShowDialog] = useState(false)
  const [name, setName] = useState('')
  const [type, setType] = useState('cash')
  const [initialBalance, setInitialBalance] = useState('0')

  const { accountsQuery, ledgersQuery, categoriesQuery, tagsQuery } = useManagementOverview()
  const createAccountMutation = useCreateAccount()

  const accounts = accountsQuery.data?.items ?? []
  const ledgers = ledgersQuery.data?.items ?? []
  const categories = categoriesQuery.data?.items ?? []
  const tags = tagsQuery.data?.items ?? []

  async function handleCreateAccount(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    await createAccountMutation.mutateAsync({
      name,
      type,
      initial_balance: Number(initialBalance),
    })
    setShowDialog(false)
    setName('')
    setType('cash')
    setInitialBalance('0')
  }

  return (
    <div className="space-y-8">
      <PageSection
        eyebrow="Financial architecture"
        title="Accounts"
        description="Manage connected balance sources, ledgers, categories, and operational tags from one place."
        actions={<Button onClick={() => setShowDialog(true)}>New Account</Button>}
      >
        <div className="grid gap-6 xl:grid-cols-2">
          <article className="rounded-[28px] bg-surface-container-low p-6">
            <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">Accounts</p>
            <div className="mt-6 space-y-4">
              {accounts.map((account) => (
                <div key={account.id} className="rounded-2xl bg-surface-container-lowest p-4">
                  <div className="flex items-center justify-between gap-3">
                    <div>
                      <p className="font-medium text-on-surface">{account.name}</p>
                      <p className="mt-1 text-xs text-on-surface-variant">{account.type}</p>
                    </div>
                    <p className="font-headline text-lg font-bold text-on-surface">{formatCurrency(account.initial_balance)}</p>
                  </div>
                </div>
              ))}
            </div>
          </article>

          <div className="grid gap-6">
            <article className="rounded-[28px] bg-surface-container-low p-6">
              <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">Ledgers</p>
              <div className="mt-6 flex flex-wrap gap-3">
                {ledgers.map((ledger) => (
                  <span key={ledger.id} className="rounded-full bg-surface-container-lowest px-4 py-2 text-sm text-on-surface">
                    {ledger.name}
                  </span>
                ))}
              </div>
            </article>

            <article className="rounded-[28px] bg-surface-container-low p-6">
              <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">Categories</p>
              <div className="mt-6 flex flex-wrap gap-3">
                {categories.map((category) => (
                  <span key={category.id} className="rounded-full bg-surface-container-lowest px-4 py-2 text-sm text-on-surface">
                    {category.name}
                  </span>
                ))}
              </div>
            </article>

            <article className="rounded-[28px] bg-surface-container-low p-6">
              <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">Tags</p>
              <div className="mt-6 flex flex-wrap gap-3">
                {tags.map((tag) => (
                  <span key={tag.id} className="rounded-full bg-surface-container-lowest px-4 py-2 text-sm text-on-surface">
                    {tag.name}
                  </span>
                ))}
              </div>
            </article>
          </div>
        </div>
      </PageSection>

      {showDialog ? (
        <DialogShell title="Create Account" description="Add a new balance source for transactions and reporting." footer={<Button type="submit" form="create-account-form">Create Account</Button>}>
          <form id="create-account-form" className="grid gap-6 md:grid-cols-2" onSubmit={(event) => void handleCreateAccount(event)}>
            <TextField label="Account Name" value={name} onChange={(event) => setName(event.target.value)} />
            <SelectField label="Account Type" value={type} onChange={(event) => setType(event.target.value)}>
              <option value="cash">cash</option>
              <option value="bank">bank</option>
              <option value="credit">credit</option>
            </SelectField>
            <TextField label="Initial Balance" type="number" value={initialBalance} onChange={(event) => setInitialBalance(event.target.value)} />
          </form>
        </DialogShell>
      ) : null}
    </div>
  )
}
