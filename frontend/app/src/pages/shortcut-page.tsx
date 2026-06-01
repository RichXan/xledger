import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { PageSection } from '@/components/ui/page-section'
import {
  useConfirmQuickAdd,
  useGenerateOCRShortcut,
  useGenerateShortcut,
  useManagementOverview,
  usePreviewQuickAdd,
} from '@/features/management/management-hooks'
import type { QuickAddPreviewResponse } from '@/features/management/management-api'

function isLoopbackHost(hostname: string) {
  return hostname === 'localhost' || hostname === '127.0.0.1' || hostname === '::1' || hostname === '[::1]'
}

function getShortcutBaseEndpoint(apiEndpoint: string | null) {
  if (!apiEndpoint) return null

  try {
    const apiUrl = new URL(apiEndpoint)
    const appUrl = new URL(window.location.origin)
    if (isLoopbackHost(apiUrl.hostname) && isLoopbackHost(appUrl.hostname)) {
      return window.location.origin
    }
    return apiEndpoint.replace(/\/+$/, '')
  } catch {
    return window.location.origin
  }
}

function newIdempotencyKey() {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    return crypto.randomUUID()
  }
  return `qe-${Date.now()}`
}

function firstDefaultLedger(ledgers: Array<{ id: string; is_default?: boolean }>) {
  return ledgers.find((ledger) => ledger.is_default)?.id ?? ledgers[0]?.id ?? ''
}

export function ShortcutPage() {
  const { t } = useTranslation()
  const [generatedToken, setGeneratedToken] = useState<string | null>(null)
  const [ocrToken, setOCRToken] = useState<string | null>(null)
  const [shortcutID, setShortcutID] = useState<string | null>(null)
  const [apiEndpoint, setApiEndpoint] = useState<string | null>(null)
  const [installURL, setInstallURL] = useState<string | null>(null)
  const [expiresAt, setExpiresAt] = useState<string | null>(null)
  const [selectedLedgerID, setSelectedLedgerID] = useState('')
  const [selectedAccountID, setSelectedAccountID] = useState('')
  const [ocrText, setOCRText] = useState('微信支付\n收款方：瑞幸咖啡\n支付金额 ￥35.00\n支付时间 2026-06-01 12:30')
  const [preview, setPreview] = useState<QuickAddPreviewResponse | null>(null)
  const [idempotencyKey, setIdempotencyKey] = useState(newIdempotencyKey)
  const [copied, setCopied] = useState(false)
  const [testMessage, setTestMessage] = useState<string | null>(null)
  const [testError, setTestError] = useState<string | null>(null)
  const [isTestingShortcut, setIsTestingShortcut] = useState(false)

  const overview = useManagementOverview()
  const generateMutation = useGenerateShortcut()
  const generateOCRMutation = useGenerateOCRShortcut()
  const previewMutation = usePreviewQuickAdd()
  const confirmMutation = useConfirmQuickAdd()

  const ledgers = overview.ledgersQuery.data?.items ?? []
  const accounts = overview.accountsQuery.data?.items ?? []
  const categories = overview.categoriesQuery.data?.items ?? []
  const shortcutBaseEndpoint = getShortcutBaseEndpoint(apiEndpoint)
  const shortcutEndpoint = shortcutBaseEndpoint ? `${shortcutBaseEndpoint}/api/shortcuts/quick-add` : null
  const previewEndpoint = shortcutBaseEndpoint ? `${shortcutBaseEndpoint}/api/quick-add/preview` : `${window.location.origin}/api/quick-add/preview`
  const confirmEndpoint = shortcutBaseEndpoint ? `${shortcutBaseEndpoint}/api/quick-add/confirm` : `${window.location.origin}/api/quick-add/confirm`

  const selectedLedger = ledgers.find((ledger) => ledger.id === selectedLedgerID)
  const selectedAccount = accounts.find((account) => account.id === selectedAccountID)
  const suggestedCategoryID = preview?.category_suggestion?.id ?? categories[0]?.id
  const shortcutSetup = shortcutEndpoint && generatedToken
    ? [
        'Method: POST',
        `URL: ${shortcutEndpoint}`,
        `Authorization: Bearer ${generatedToken}`,
        'Content-Type: application/json',
        '',
        '{',
        '  "amount": 35,',
        '  "type": "expense",',
        '  "category": "Lunch",',
        '  "memo": "weekday lunch"',
        '}',
      ].join('\n')
    : null

  const ocrShortcutSetup = useMemo(() => {
    if (!ocrToken) return null
    return [
      'Xledger OCR shortcut',
      `Preview URL: ${previewEndpoint}`,
      `Confirm URL: ${confirmEndpoint}`,
      `Authorization: Bearer ${ocrToken}`,
      `Shortcut ID: ${shortcutID ?? ''}`,
      `Default ledger: ${selectedLedger?.name ?? selectedLedgerID}`,
      `Default account: ${selectedAccount?.name ?? 'None'}`,
    ].join('\n')
  }, [confirmEndpoint, ocrToken, previewEndpoint, selectedAccount?.name, selectedLedger?.name, selectedLedgerID, shortcutID])

  useEffect(() => {
    if (!selectedLedgerID && ledgers.length > 0) {
      setSelectedLedgerID(firstDefaultLedger(ledgers))
    }
  }, [ledgers, selectedLedgerID])

  useEffect(() => {
    if (!selectedAccountID && accounts.length > 0) {
      setSelectedAccountID(accounts[0].id)
    }
  }, [accounts, selectedAccountID])

  async function handleGenerateShortcut() {
    try {
      const result = await generateMutation.mutateAsync({ name: t('shortcutPage.eyebrow'), expiresIn: 90 })
      setGeneratedToken(result.pat_token)
      setApiEndpoint(result.api_endpoint)
      setExpiresAt(result.expires_at ?? null)
      setCopied(false)
      setTestMessage(null)
      setTestError(null)
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : String(error)
      alert(t('shortcutPage.generateFailed', { message }))
    }
  }

  async function handleGenerateOCRShortcut() {
    if (!selectedLedgerID) return
    const result = await generateOCRMutation.mutateAsync({
      name: 'Xledger OCR 记账',
      expiresIn: 90,
      defaultLedgerId: selectedLedgerID,
      defaultAccountId: selectedAccountID || undefined,
    })
    setOCRToken(result.pat_token)
    setShortcutID(result.shortcut_id ?? null)
    setApiEndpoint(result.api_endpoint)
    setInstallURL(result.install_url ?? result.shortcut_url ?? null)
    setExpiresAt(result.expires_at ?? null)
    setPreview(null)
    setTestMessage(null)
    setTestError(null)
  }

  function handleCopyToken() {
    if (generatedToken) {
      navigator.clipboard.writeText(generatedToken)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  function handleCopyApiEndpoint() {
    if (shortcutEndpoint) navigator.clipboard.writeText(shortcutEndpoint)
  }

  function handleCopyShortcutSetup() {
    if (shortcutSetup) navigator.clipboard.writeText(shortcutSetup)
  }

  async function handlePreviewOCR() {
    if (!ocrToken || !selectedLedgerID) return
    const result = await previewMutation.mutateAsync({
      patToken: ocrToken,
      shortcutId: shortcutID ?? undefined,
      rawText: ocrText,
      source: 'ios_shortcuts_ocr',
      idempotencyKey,
      defaultLedgerId: selectedLedgerID,
      defaultAccountId: selectedAccountID || undefined,
    })
    setPreview(result)
  }

  async function handleConfirmOCR() {
    if (!ocrToken || !preview || !selectedLedgerID) return
    await confirmMutation.mutateAsync({
      patToken: ocrToken,
      shortcutId: shortcutID ?? undefined,
      idempotencyKey,
      ledgerId: selectedLedgerID,
      accountId: selectedAccountID || undefined,
      categoryId: suggestedCategoryID,
      type: preview.type,
      amount: preview.amount,
      memo: preview.memo,
      occurredAt: preview.occurred_at,
    })
    setTestMessage('Entry saved')
    setIdempotencyKey(newIdempotencyKey())
  }

  async function handleTestShortcut() {
    if (!shortcutEndpoint || !generatedToken) return

    setIsTestingShortcut(true)
    setTestMessage(null)
    setTestError(null)
    try {
      const response = await fetch(shortcutEndpoint, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${generatedToken}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          amount: 35,
          type: 'expense',
          category: 'Lunch',
          memo: 'weekday lunch',
        }),
      })
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      setTestMessage(t('shortcutPage.testSuccess'))
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : String(error)
      setTestError(t('shortcutPage.testFailed', { message }))
    } finally {
      setIsTestingShortcut(false)
    }
  }

  return (
    <div className="space-y-8">
      <PageSection eyebrow={t('shortcutPage.eyebrow')} title={t('shortcutPage.title')} description={t('shortcutPage.description')}>
        <div className="grid gap-5 lg:grid-cols-[minmax(0,1fr)_360px]">
          <article className="rounded-2xl bg-surface-container-low p-4 sm:p-6">
            <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">OCR automation setup</p>
            <div className="mt-5 grid gap-4 md:grid-cols-2">
              <label className="block text-sm font-semibold text-on-surface" htmlFor="shortcut-default-ledger">
                Default ledger
                <select
                  id="shortcut-default-ledger"
                  aria-label="Default ledger"
                  className="mt-2 w-full rounded-xl border border-outline/20 bg-white px-3 py-2 text-sm"
                  value={selectedLedgerID}
                  onChange={(event) => setSelectedLedgerID(event.target.value)}
                >
                  {ledgers.map((ledger) => (
                    <option key={ledger.id} value={ledger.id}>
                      {ledger.name}
                    </option>
                  ))}
                </select>
              </label>
              <label className="block text-sm font-semibold text-on-surface" htmlFor="shortcut-default-account">
                Default account
                <select
                  id="shortcut-default-account"
                  aria-label="Default account"
                  className="mt-2 w-full rounded-xl border border-outline/20 bg-white px-3 py-2 text-sm"
                  value={selectedAccountID}
                  onChange={(event) => setSelectedAccountID(event.target.value)}
                >
                  <option value="">No default account</option>
                  {accounts.map((account) => (
                    <option key={account.id} value={account.id}>
                      {account.name}
                    </option>
                  ))}
                </select>
              </label>
            </div>
            <Button className="mt-4 w-full sm:w-auto" onClick={() => void handleGenerateOCRShortcut()} disabled={!selectedLedgerID || generateOCRMutation.isPending}>
              {generateOCRMutation.isPending ? 'Generating...' : 'Generate OCR shortcut'}
            </Button>

            {ocrToken ? (
              <div className="mt-5 rounded-2xl border border-primary/20 bg-surface-container-lowest p-4">
                <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                  <div>
                    <p className="text-sm font-semibold text-on-surface">Install on iPhone</p>
                    <p className="mt-1 text-xs text-on-surface-variant">This shortcut is fixed to your selected ledger and account through a PAT created for your user.</p>
                    {installURL ? <code className="mt-2 block break-all text-xs text-primary">{installURL}</code> : null}
                  </div>
                  <div className="grid h-28 w-28 shrink-0 place-items-center rounded-xl border border-outline/20 bg-white text-center text-[10px] font-bold text-on-surface-variant">
                    Shortcut QR
                  </div>
                </div>
                {ocrShortcutSetup ? (
                  <pre className="mt-4 max-h-44 overflow-auto rounded-xl bg-primary p-3 text-[11px] leading-relaxed text-white">{ocrShortcutSetup}</pre>
                ) : null}
              </div>
            ) : null}
          </article>

          <article className="rounded-2xl bg-surface-container-low p-4 sm:p-6">
            <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">Daily use</p>
            <ol className="mt-4 space-y-3 text-sm text-on-surface-variant">
              <li>1. Add the imported shortcut to Lock Screen, Control Center, Siri, or Back Tap.</li>
              <li>2. Screenshot a payment result page and run the shortcut.</li>
              <li>3. Review amount, ledger, account, category, and memo before saving.</li>
            </ol>
          </article>
        </div>

        <div className="mt-5 grid gap-5 lg:grid-cols-[minmax(0,0.95fr)_minmax(0,1.05fr)]">
          <article className="rounded-2xl bg-surface-container-low p-4 sm:p-6">
            <label className="block text-sm font-semibold text-on-surface" htmlFor="shortcut-ocr-text">
              OCR text
              <textarea
                id="shortcut-ocr-text"
                className="mt-2 min-h-40 w-full rounded-xl border border-outline/20 bg-white px-3 py-2 text-sm"
                value={ocrText}
                onChange={(event) => setOCRText(event.target.value)}
              />
            </label>
            <Button className="mt-4 w-full sm:w-auto" onClick={() => void handlePreviewOCR()} disabled={!ocrToken || previewMutation.isPending}>
              {previewMutation.isPending ? 'Previewing...' : 'Preview OCR entry'}
            </Button>
          </article>

          <article className="rounded-2xl bg-surface-container-low p-4 sm:p-6">
            <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">Mobile confirmation preview</p>
            {preview ? (
              <div className="mt-4 rounded-2xl border border-outline/15 bg-white p-4">
                <p className="text-xs font-semibold text-on-surface-variant">{selectedLedger?.name ?? 'Selected ledger'} · {selectedAccount?.name ?? 'No account'}</p>
                <p className="mt-3 text-4xl font-black text-on-surface">¥{preview.amount.toFixed(2)}</p>
                <p className="mt-2 text-sm font-semibold text-on-surface">{preview.memo}</p>
                <p className="mt-1 text-xs text-on-surface-variant">
                  {preview.category_suggestion?.name ?? 'Uncategorized'} · {new Date(preview.occurred_at).toLocaleString()}
                </p>
                {preview.needs_review ? <p className="mt-3 rounded-xl bg-amber-50 px-3 py-2 text-xs font-semibold text-amber-900">Needs review before saving.</p> : null}
                <Button className="mt-4 w-full" onClick={() => void handleConfirmOCR()} disabled={confirmMutation.isPending}>
                  {confirmMutation.isPending ? 'Saving...' : 'Confirm entry'}
                </Button>
              </div>
            ) : (
              <div className="mt-4 rounded-2xl border border-dashed border-outline/30 bg-white/70 p-6 text-sm text-on-surface-variant">
                Generate the OCR shortcut, paste OCR text, then preview before confirming.
              </div>
            )}
            {testMessage ? <p role="status" className="mt-3 text-sm font-semibold text-primary">{testMessage}</p> : null}
          </article>
        </div>

        <div className="mt-5 grid gap-6 xl:grid-cols-[1.2fr_0.8fr]">
          <article className="rounded-2xl bg-surface-container-low p-4 sm:p-6">
            <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">{t('shortcutPage.installationFlow')}</p>

            <div className="mt-6 space-y-4">
              <div className="rounded-2xl bg-surface-container-lowest p-4">
                <p className="text-sm font-medium text-on-surface">{t('shortcutPage.step1Title')}</p>
                <p className="mt-1 text-xs text-on-surface-variant">{t('shortcutPage.step1Description')}</p>
                <Button className="mt-3" onClick={() => void handleGenerateShortcut()} disabled={generateMutation.isPending}>
                  {generateMutation.isPending ? t('shortcutPage.generating') : t('shortcutPage.generateCredentials')}
                </Button>
              </div>

              {generatedToken && (
                <div className="rounded-2xl border-2 border-primary/20 bg-surface-container-lowest p-4">
                  <p className="text-sm font-medium text-on-surface">{t('shortcutPage.step2Title')}</p>
                  <p className="mt-1 text-xs text-on-surface-variant">{t('shortcutPage.step2Description')}</p>

                  <div className="mt-4 space-y-3 rounded-xl bg-surface-container p-3">
                    <div>
                      <p className="text-xs font-medium text-on-surface-variant">{t('shortcutPage.apiEndpoint')}</p>
                      <div className="mt-1 flex items-center justify-between gap-2">
                        <code className="truncate font-mono text-xs text-primary">{shortcutEndpoint}</code>
                        <Button variant="ghost" onClick={handleCopyApiEndpoint}>{t('common.copy')}</Button>
                      </div>
                    </div>
                    <div className="border-t border-outline/10" />
                    <div>
                      <p className="text-xs font-medium text-on-surface-variant">{t('shortcutPage.accessToken')}</p>
                      <div className="mt-1 flex items-center justify-between gap-2">
                        <code className="truncate font-mono text-xs text-primary">{generatedToken}</code>
                        <Button variant="ghost" onClick={handleCopyToken}>{copied ? t('common.copied') : t('common.copy')}</Button>
                      </div>
                    </div>
                    {expiresAt ? <p className="border-t border-outline/10 pt-3 text-xs font-medium text-on-surface-variant">{t('shortcutPage.expiresAt', { date: new Date(expiresAt).toLocaleDateString() })}</p> : null}
                  </div>

                  <div className="mt-4 rounded-xl bg-surface-container-high p-4 text-xs text-on-surface">
                    <p className="mb-2 font-bold">{t('shortcutPage.samplePayloadTitle')}</p>
                    <pre className="overflow-x-auto rounded-lg bg-white p-3 font-mono text-[11px] leading-relaxed text-on-surface">{`{
  "amount": 35,
  "type": "expense",
  "category": "Lunch",
  "memo": "weekday lunch"
}`}</pre>
                    <div className="mt-3 flex flex-wrap items-center gap-3">
                      <Button className="px-3 py-1.5 text-xs" onClick={() => void handleTestShortcut()} disabled={isTestingShortcut}>
                        {isTestingShortcut ? t('shortcutPage.testingShortcut') : t('shortcutPage.testShortcut')}
                      </Button>
                      {testError ? <p role="alert" className="text-xs font-semibold text-error">{testError}</p> : null}
                    </div>
                  </div>

                  {shortcutSetup ? (
                    <div className="mt-4 rounded-xl bg-primary p-4 text-xs text-white">
                      <div className="flex items-center justify-between gap-3">
                        <p className="font-bold">{t('shortcutPage.copySetupTitle')}</p>
                        <Button className="bg-white px-3 py-1.5 text-xs text-primary hover:bg-primary-fixed" onClick={handleCopyShortcutSetup}>{t('shortcutPage.copySetup')}</Button>
                      </div>
                      <pre className="mt-3 max-h-52 overflow-auto whitespace-pre-wrap rounded-lg bg-white/10 p-3 font-mono text-[11px] leading-relaxed text-white">{shortcutSetup}</pre>
                    </div>
                  ) : null}
                </div>
              )}
            </div>
          </article>

          <article className="rounded-2xl bg-primary-container p-4 sm:p-6">
            <p className="text-sm font-medium text-on-primary-container">{t('shortcutPage.securityNote')}</p>
            <p className="mt-2 text-xs text-on-primary-container">{t('shortcutPage.securityDescription')}</p>
          </article>
        </div>
      </PageSection>
    </div>
  )
}
