import { Search } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
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
  const { t } = useTranslation()
  const navigate = useNavigate()
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
  const [categoryQuery, setCategoryQuery] = useState('')
  const [tagQuery, setTagQuery] = useState('')
  const [showAllCategories, setShowAllCategories] = useState(false)
  const [showAllTags, setShowAllTags] = useState(false)

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
  const normalizedCategoryQuery = categoryQuery.trim().toLowerCase()
  const normalizedTagQuery = tagQuery.trim().toLowerCase()

  const filteredCategories = useMemo(() => {
    if (!normalizedCategoryQuery) return categories
    return categories.filter((category) => category.name.toLowerCase().includes(normalizedCategoryQuery))
  }, [categories, normalizedCategoryQuery])

  const filteredTags = useMemo(() => {
    if (!normalizedTagQuery) return tags
    return tags.filter((tag) => tag.name.toLowerCase().includes(normalizedTagQuery))
  }, [normalizedTagQuery, tags])

  const categoryPreviewCount = 16
  const tagPreviewCount = 14
  const visibleCategories = showAllCategories ? filteredCategories : filteredCategories.slice(0, categoryPreviewCount)
  const visibleTags = showAllTags ? filteredTags : filteredTags.slice(0, tagPreviewCount)

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
        eyebrow={t('accountsPage.eyebrow')}
        title={t('accountsPage.title')}
        description={t('accountsPage.description')}
        actions={<Button onClick={() => setShowDialog(true)}>{t('accountsPage.newAccount')}</Button>}
      >
        <div className="grid gap-4 xl:grid-cols-[1.2fr_0.8fr]">
          <article className="rounded-2xl border border-outline/10 bg-surface-container-low p-5">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">
                {t('accountsPage.accounts')}
              </p>
              <div className="rounded-xl border border-outline/10 bg-white px-3 py-2 text-right">
                <p className="text-[10px] font-bold uppercase tracking-[0.12em] text-on-surface-variant">
                  {t('accountsPage.totalBalance')}
                </p>
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
                        {t('accountsPage.editAccount')}
                      </Button>
                    </div>
                  </div>
                </div>
              ))}
              {accounts.length === 0 ? (
                <div className="rounded-xl border border-outline/10 bg-white p-4 text-sm text-on-surface-variant">
                  <p>{t('accountsPage.noAccounts')}</p>
                  <div className="mt-3 flex flex-wrap gap-2">
                    <Button className="px-3 py-1.5 text-xs" onClick={() => setShowDialog(true)}>{t('accountsPage.createFirstAccount')}</Button>
                    <Button className="px-3 py-1.5 text-xs" variant="secondary" onClick={() => navigate('/transactions')}>
                      {t('accountsPage.goToTransactions')}
                    </Button>
                  </div>
                </div>
              ) : null}
            </div>
          </article>

          <div className="grid gap-4">
            <article className="rounded-2xl border border-outline/10 bg-surface-container-low p-5">
              <div className="flex items-center justify-between gap-3">
                <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">
                  {t('accountsPage.ledgers')}
                </p>
                <Button variant="secondary" onClick={() => setShowLedgerDialog(true)}>{t('accountsPage.newLedger')}</Button>
              </div>
              <div className="mt-4 space-y-2.5">
                {ledgers.map((ledger) => (
                  <div key={ledger.id} className="flex items-center justify-between rounded-xl border border-outline/10 bg-white px-4 py-3">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium text-on-surface">{ledger.name}</span>
                      {ledger.is_default ? <span className="rounded-full bg-surface-container-low px-2 py-0.5 text-xs text-on-surface-variant">{t('common.default')}</span> : null}
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
                        {t('common.edit')}
                      </Button>
                      <Button variant="ghost" className="px-2 py-1 text-xs" disabled={ledger.is_default} onClick={() => void deleteLedgerMutation.mutateAsync(ledger.id)}>
                        {t('common.delete')}
                      </Button>
                    </div>
                  </div>
                ))}
                {ledgers.length === 0 ? (
                  <div className="rounded-xl border border-outline/10 bg-white p-4 text-sm text-on-surface-variant">
                    {t('accountsPage.noLedgers')}
                  </div>
                ) : null}
              </div>
            </article>

            <article className="rounded-2xl border border-outline/10 bg-surface-container-low p-5">
              <div className="flex items-center justify-between gap-3">
                <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">
                  {t('accountsPage.categories', { count: filteredCategories.length })}
                </p>
                {filteredCategories.length > categoryPreviewCount ? (
                  <Button className="px-2.5 py-1 text-xs" variant="ghost" onClick={() => setShowAllCategories((current) => !current)}>
                    {showAllCategories ? t('accountsPage.showLess') : t('accountsPage.showAll')}
                  </Button>
                ) : null}
              </div>
              <div className="mt-3">
                <label className="relative block">
                  <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-on-surface-variant" />
                  <input
                    value={categoryQuery}
                    onChange={(event) => setCategoryQuery(event.target.value)}
                    placeholder={t('accountsPage.searchCategories')}
                    className="h-10 w-full rounded-xl border border-outline/20 bg-white pl-9 pr-3 text-sm text-on-surface placeholder:text-on-surface-variant/70"
                  />
                </label>
              </div>
              <div className="mt-4 flex max-h-44 flex-wrap gap-2 overflow-auto pr-1">
                {visibleCategories.map((category) => (
                  <span key={category.id} className="rounded-full border border-outline/10 bg-white px-3 py-1.5 text-sm text-on-surface">
                    {category.name}
                  </span>
                ))}
                {visibleCategories.length === 0 ? (
                  <p className="text-sm text-on-surface-variant">{t('accountsPage.noCategories')}</p>
                ) : null}
              </div>
            </article>

            <article className="rounded-2xl border border-outline/10 bg-surface-container-low p-5">
              <div className="flex items-center justify-between gap-3">
                <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">
                  {t('accountsPage.tags', { count: filteredTags.length })}
                </p>
                {filteredTags.length > tagPreviewCount ? (
                  <Button className="px-2.5 py-1 text-xs" variant="ghost" onClick={() => setShowAllTags((current) => !current)}>
                    {showAllTags ? t('accountsPage.showLess') : t('accountsPage.showAll')}
                  </Button>
                ) : null}
              </div>
              <div className="mt-3">
                <label className="relative block">
                  <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-on-surface-variant" />
                  <input
                    value={tagQuery}
                    onChange={(event) => setTagQuery(event.target.value)}
                    placeholder={t('accountsPage.searchTags')}
                    className="h-10 w-full rounded-xl border border-outline/20 bg-white pl-9 pr-3 text-sm text-on-surface placeholder:text-on-surface-variant/70"
                  />
                </label>
              </div>
              <div className="mt-4 flex max-h-40 flex-wrap gap-2 overflow-auto pr-1">
                {visibleTags.map((tag) => (
                  <span key={tag.id} className="rounded-full border border-outline/10 bg-white px-3 py-1.5 text-sm text-on-surface">
                    {tag.name}
                  </span>
                ))}
                {visibleTags.length === 0 ? (
                  <p className="text-sm text-on-surface-variant">{t('accountsPage.noTags')}</p>
                ) : null}
              </div>
            </article>
          </div>
        </div>
      </PageSection>

      {showDialog ? (
        <DialogShell
          title={t('accountsPage.dialogs.createAccountTitle')}
          description={t('accountsPage.dialogs.createAccountDescription')}
          onClose={() => setShowDialog(false)}
          footer={
            <>
              <Button variant="ghost" onClick={() => setShowDialog(false)}>{t('common.cancel')}</Button>
              <Button type="submit" form="create-account-form">{t('account.create')}</Button>
            </>
          }
        >
          <form id="create-account-form" className="grid gap-5 md:grid-cols-2" onSubmit={(event) => void handleCreateAccount(event)}>
            <TextField label={t('accountsPage.dialogs.accountName')} value={name} onChange={(event) => setName(event.target.value)} />
            <SelectField label={t('accountsPage.dialogs.accountType')} value={type} onChange={(event) => setType(event.target.value)}>
              <option value="cash">{t('accountTypes.cash')}</option>
              <option value="bank">{t('accountTypes.bank')}</option>
              <option value="credit">{t('accountTypes.credit')}</option>
            </SelectField>
            <TextField label={t('accountsPage.dialogs.initialBalance')} type="number" value={initialBalance} onChange={(event) => setInitialBalance(event.target.value)} />
          </form>
        </DialogShell>
      ) : null}

      {editingAccountID ? (
        <DialogShell
          title={t('accountsPage.dialogs.editAccountTitle')}
          description={t('accountsPage.dialogs.editAccountDescription')}
          onClose={() => setEditingAccountID(null)}
          footer={
            <>
              <Button variant="ghost" onClick={() => setEditingAccountID(null)}>{t('common.cancel')}</Button>
              <Button type="submit" form="edit-account-form">{t('common.saveChanges')}</Button>
            </>
          }
        >
          <form id="edit-account-form" className="grid gap-5 md:grid-cols-2" onSubmit={(event) => void handleUpdateAccount(event)}>
            <TextField label={t('accountsPage.dialogs.accountName')} value={editName} onChange={(event) => setEditName(event.target.value)} />
            <SelectField label={t('accountsPage.dialogs.accountType')} value={editType} onChange={(event) => setEditType(event.target.value)}>
              <option value="cash">{t('accountTypes.cash')}</option>
              <option value="bank">{t('accountTypes.bank')}</option>
              <option value="credit">{t('accountTypes.credit')}</option>
            </SelectField>
          </form>
        </DialogShell>
      ) : null}

      {showLedgerDialog ? (
        <DialogShell
          title={t('accountsPage.dialogs.createLedgerTitle')}
          description={t('accountsPage.dialogs.createLedgerDescription')}
          onClose={() => setShowLedgerDialog(false)}
          footer={
            <>
              <Button variant="ghost" onClick={() => setShowLedgerDialog(false)}>{t('common.cancel')}</Button>
              <Button type="submit" form="create-ledger-form">{t('accountsPage.dialogs.createLedgerTitle')}</Button>
            </>
          }
        >
          <form id="create-ledger-form" className="grid gap-5" onSubmit={(event) => void handleCreateLedger(event)}>
            <TextField label={t('accountsPage.dialogs.ledgerName')} value={ledgerName} onChange={(event) => setLedgerName(event.target.value)} />
          </form>
        </DialogShell>
      ) : null}

      {editingLedgerID ? (
        <DialogShell
          title={t('accountsPage.dialogs.editLedgerTitle')}
          description={t('accountsPage.dialogs.editLedgerDescription')}
          onClose={() => setEditingLedgerID(null)}
          footer={
            <>
              <Button variant="ghost" onClick={() => setEditingLedgerID(null)}>{t('common.cancel')}</Button>
              <Button type="submit" form="edit-ledger-form">{t('common.saveChanges')}</Button>
            </>
          }
        >
          <form id="edit-ledger-form" className="grid gap-5" onSubmit={(event) => void handleUpdateLedger(event)}>
            <TextField label={t('accountsPage.dialogs.ledgerName')} value={editingLedgerName} onChange={(event) => setEditingLedgerName(event.target.value)} />
          </form>
        </DialogShell>
      ) : null}
    </div>
  )
}
