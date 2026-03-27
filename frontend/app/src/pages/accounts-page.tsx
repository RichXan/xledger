import { useMemo, useState } from 'react'
import { Button } from '@/components/ui/button'
import { DialogShell } from '@/components/ui/dialog-shell'
import { PageSection } from '@/components/ui/page-section'
import { SelectField } from '@/components/ui/select-field'
import { TextField } from '@/components/ui/text-field'
import {
  useCreateAccount,
  useCreateLedger,
  useDeleteLedger,
  useManagementOverview,
  useUpdateAccount,
  useUpdateLedger,
} from '@/features/management/management-hooks'
import { formatCurrency } from '@/lib/format'

export function AccountsPage() {
  const [showDialog, setShowDialog] = useState(false)
  const [editingAccountID, setEditingAccountID] = useState<string | null>(null)
  const [name, setName] = useState('')
  const [type, setType] = useState('cash')
  const [initialBalance, setInitialBalance] = useState('0')
  const [editName, setEditName] = useState('')
  const [editType, setEditType] = useState('cash')
  const [showLedgerDialog, setShowLedgerDialog] = useState(false)
  const [ledgerName, setLedgerName] = useState('')
  const [editingLedgerID, setEditingLedgerID] = useState<string | null>(null)
  const [editingLedgerName, setEditingLedgerName] = useState('')

  const { accountsQuery, ledgersQuery, categoriesQuery, tagsQuery } = useManagementOverview()
  const createAccountMutation = useCreateAccount()
  const updateAccountMutation = useUpdateAccount()
  const createLedgerMutation = useCreateLedger()
  const updateLedgerMutation = useUpdateLedger()
  const deleteLedgerMutation = useDeleteLedger()

  const accounts = accountsQuery.data?.items ?? []
  const ledgers = ledgersQuery.data?.items ?? []
  const categories = categoriesQuery.data?.items ?? []
  const tags = tagsQuery.data?.items ?? []

  const totals = useMemo(
    () => accounts.reduce((acc, account) => acc + (Number.isFinite(account.initial_balance) ? account.initial_balance : 0), 0),
    [accounts],
  )

  async function handleCreateAccount(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    await createAccountMutation.mutateAsync({ name, type, initial_balance: Number(initialBalance) })
    setShowDialog(false)
    setName('')
    setType('cash')
    setInitialBalance('0')
  }

  async function handleUpdateAccount(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!editingAccountID) return
    await updateAccountMutation.mutateAsync({ id: editingAccountID, name: editName, type: editType })
    setEditingAccountID(null)
  }

  async function handleCreateLedger(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    await createLedgerMutation.mutateAsync({ name: ledgerName, is_default: false })
    setShowLedgerDialog(false)
    setLedgerName('')
  }

  async function handleUpdateLedger(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!editingLedgerID) return
    await updateLedgerMutation.mutateAsync({ id: editingLedgerID, name: editingLedgerName })
    setEditingLedgerID(null)
    setEditingLedgerName('')
  }

  return (
    <div className="space-y-6">
      <PageSection
        eyebrow="Financial architecture"
        title="Accounts"
        description="Manage connected balance sources, ledgers, categories, and operational tags from one place."
        actions={<Button onClick={() => setShowDialog(true)}>New Account</Button>}
      >
        <div className="grid gap-4 xl:grid-cols-[1.2fr_0.8fr]">
          <article className="rounded-2xl border border-outline/10 bg-surface-container-low p-5">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">Accounts</p>
              <div className="rounded-xl border border-outline/10 bg-white px-3 py-2 text-right">
                <p className="text-[10px] font-bold uppercase tracking-[0.12em] text-on-surface-variant">Total Balance</p>
                <p className="font-headline text-3xl font-extrabold text-on-surface">{formatCurrency(totals)}</p>
              </div>
            </div>
            <div className="mt-4 grid gap-3">
              {accounts.map((account) => (
                <div key={account.id} className="rounded-xl border border-outline/10 bg-white p-4">
                  <div className="flex items-start justify-between gap-4">
                    <div>
                      <p className="font-semibold text-on-surface">{account.name}</p>
                      <p className="mt-1 text-xs uppercase tracking-[0.12em] text-on-surface-variant">{account.type}</p>
                    </div>
                    <div className="text-right">
                      <p className="font-headline text-2xl font-extrabold text-on-surface">{formatCurrency(account.initial_balance)}</p>
                      <Button
                        variant="ghost"
                        className="mt-2 px-0 py-0 text-xs"
                        onClick={() => {
                          setEditingAccountID(account.id)
                          setEditName(account.name)
                          setEditType(account.type)
                        }}
                      >
                        Edit account
                      </Button>
                    </div>
                  </div>
                </div>
              ))}
              {accounts.length === 0 ? (
                <div className="rounded-xl border border-outline/10 bg-white p-4 text-sm text-on-surface-variant">
                  No accounts yet. Create your first account to start tracking balances.
                </div>
              ) : null}
            </div>
          </article>

          <div className="grid gap-4">
            <article className="rounded-2xl border border-outline/10 bg-surface-container-low p-5">
              <div className="flex items-center justify-between gap-3">
                <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">Ledgers</p>
                <Button variant="secondary" onClick={() => setShowLedgerDialog(true)}>New Ledger</Button>
              </div>
              <div className="mt-4 space-y-2.5">
                {ledgers.map((ledger) => (
                  <div key={ledger.id} className="flex items-center justify-between rounded-xl border border-outline/10 bg-white px-4 py-3">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium text-on-surface">{ledger.name}</span>
                      {ledger.is_default ? <span className="rounded-full bg-surface-container-low px-2 py-0.5 text-xs text-on-surface-variant">default</span> : null}
                    </div>
                    <div className="flex items-center gap-1.5">
                      <Button
                        variant="ghost"
                        className="px-2 py-1 text-xs"
                        onClick={() => {
                          setEditingLedgerID(ledger.id)
                          setEditingLedgerName(ledger.name)
                        }}
                      >
                        Edit
                      </Button>
                      <Button variant="ghost" className="px-2 py-1 text-xs" disabled={ledger.is_default} onClick={() => void deleteLedgerMutation.mutateAsync(ledger.id)}>
                        Delete
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            </article>

            <article className="rounded-2xl border border-outline/10 bg-surface-container-low p-5">
              <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">Categories</p>
              <div className="mt-4 flex flex-wrap gap-2">
                {categories.map((category) => (
                  <span key={category.id} className="rounded-full border border-outline/10 bg-white px-3 py-1.5 text-sm text-on-surface">
                    {category.name}
                  </span>
                ))}
              </div>
            </article>

            <article className="rounded-2xl border border-outline/10 bg-surface-container-low p-5">
              <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">Tags</p>
              <div className="mt-4 flex flex-wrap gap-2">
                {tags.map((tag) => (
                  <span key={tag.id} className="rounded-full border border-outline/10 bg-white px-3 py-1.5 text-sm text-on-surface">
                    {tag.name}
                  </span>
                ))}
              </div>
            </article>
          </div>
        </div>
      </PageSection>

      {showDialog ? (
        <DialogShell
          title="Create Account"
          description="Add a new balance source for transactions and reporting."
          onClose={() => setShowDialog(false)}
          footer={
            <>
              <Button variant="ghost" onClick={() => setShowDialog(false)}>Cancel</Button>
              <Button type="submit" form="create-account-form">Create Account</Button>
            </>
          }
        >
          <form id="create-account-form" className="grid gap-5 md:grid-cols-2" onSubmit={(event) => void handleCreateAccount(event)}>
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

      {editingAccountID ? (
        <DialogShell
          title="Edit Account"
          description="Update account name or type."
          onClose={() => setEditingAccountID(null)}
          footer={
            <>
              <Button variant="ghost" onClick={() => setEditingAccountID(null)}>Cancel</Button>
              <Button type="submit" form="edit-account-form">Save Changes</Button>
            </>
          }
        >
          <form id="edit-account-form" className="grid gap-5 md:grid-cols-2" onSubmit={(event) => void handleUpdateAccount(event)}>
            <TextField label="Account Name" value={editName} onChange={(event) => setEditName(event.target.value)} />
            <SelectField label="Account Type" value={editType} onChange={(event) => setEditType(event.target.value)}>
              <option value="cash">cash</option>
              <option value="bank">bank</option>
              <option value="credit">credit</option>
            </SelectField>
          </form>
        </DialogShell>
      ) : null}

      {showLedgerDialog ? (
        <DialogShell
          title="Create Ledger"
          description="Add a ledger for organizing balances and transactions."
          onClose={() => setShowLedgerDialog(false)}
          footer={
            <>
              <Button variant="ghost" onClick={() => setShowLedgerDialog(false)}>Cancel</Button>
              <Button type="submit" form="create-ledger-form">Create Ledger</Button>
            </>
          }
        >
          <form id="create-ledger-form" className="grid gap-5" onSubmit={(event) => void handleCreateLedger(event)}>
            <TextField label="Ledger Name" value={ledgerName} onChange={(event) => setLedgerName(event.target.value)} />
          </form>
        </DialogShell>
      ) : null}

      {editingLedgerID ? (
        <DialogShell
          title="Edit Ledger"
          description="Rename this ledger."
          onClose={() => setEditingLedgerID(null)}
          footer={
            <>
              <Button variant="ghost" onClick={() => setEditingLedgerID(null)}>Cancel</Button>
              <Button type="submit" form="edit-ledger-form">Save Changes</Button>
            </>
          }
        >
          <form id="edit-ledger-form" className="grid gap-5" onSubmit={(event) => void handleUpdateLedger(event)}>
            <TextField label="Ledger Name" value={editingLedgerName} onChange={(event) => setEditingLedgerName(event.target.value)} />
          </form>
        </DialogShell>
      ) : null}
    </div>
  )
}
