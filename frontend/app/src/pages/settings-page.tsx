import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { PageSection } from '@/components/ui/page-section'
import { useCreatePAT, useExportCsv, usePATs, useRevokePAT } from '@/features/management/management-hooks'

export function SettingsPage() {
  const { t } = useTranslation()
  const [latestSecret, setLatestSecret] = useState<string | null>(null)
  const patsQuery = usePATs()
  const createPATMutation = useCreatePAT()
  const revokePATMutation = useRevokePAT()
  const exportCsvMutation = useExportCsv()

  const pats = patsQuery.data?.items ?? []

  async function handleCreatePAT() {
    const result = await createPATMutation.mutateAsync()
    setLatestSecret(result.token)
  }

  async function handleExportCsv() {
    const content = await exportCsvMutation.mutateAsync()
    const blob = new Blob([content], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = 'xledger-export.csv'
    anchor.rel = 'noopener'
    document.body.appendChild(anchor)
    try {
      anchor.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true }))
    } finally {
      anchor.remove()
      URL.revokeObjectURL(url)
    }
  }

  return (
    <div className="space-y-8">
      <PageSection
        eyebrow={t('settingsPage.eyebrow')}
        title={t('settingsPage.title')}
        description={t('settingsPage.description')}
        actions={<Button variant="secondary" onClick={() => void handleExportCsv()}>{t('settingsPage.exportCsv')}</Button>}
      >
        <div className="grid gap-6 xl:grid-cols-[1.2fr_0.8fr]">
          <article className="rounded-[28px] bg-surface-container-low p-6">
            <div className="flex items-center justify-between gap-4">
              <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">
                {t('settingsPage.personalAccessTokens')}
              </p>
              <Button onClick={() => void handleCreatePAT()}>{t('settingsPage.createPat')}</Button>
            </div>
            <div className="mt-4 rounded-2xl border border-outline/15 bg-white p-4 text-sm text-on-surface-variant">
              {t('settingsPage.patWarning')}
            </div>
            <div className="mt-6 space-y-4">
              {pats.map((pat) => (
                <div key={pat.id} className="rounded-2xl bg-surface-container-lowest p-4">
                  <div className="flex items-center justify-between gap-3">
                    <div>
                      <p className="font-medium text-on-surface">{pat.id}</p>
                      <p className="mt-1 text-xs text-on-surface-variant">{pat.name}</p>
                    </div>
                    <Button variant="ghost" onClick={() => void revokePATMutation.mutateAsync(pat.id)}>
                      {t('settingsPage.revoke', { id: pat.id })}
                    </Button>
                  </div>
                </div>
              ))}
              {pats.length === 0 ? (
                <div className="rounded-2xl border border-outline/10 bg-surface-container-lowest p-4 text-sm text-on-surface-variant">
                  <p>{t('settingsPage.noPat')}</p>
                  <p className="mt-2">{t('settingsPage.noPatHint')}</p>
                  <div className="mt-3">
                    <Button className="px-3 py-1.5 text-xs" onClick={() => void handleCreatePAT()}>
                      {t('settingsPage.createFirstPat')}
                    </Button>
                  </div>
                </div>
              ) : null}
            </div>
          </article>

          <article className="rounded-[28px] bg-surface-container-low p-6">
            <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">
              {t('settingsPage.latestSecret')}
            </p>
            <div className="mt-6 rounded-2xl bg-surface-container-lowest p-4">
              <p className="text-sm text-on-surface-variant">{t('settingsPage.secretHint')}</p>
              <p className="mt-3 font-mono text-sm text-on-surface">{latestSecret ?? t('settingsPage.noSecret')}</p>
            </div>
          </article>
        </div>
      </PageSection>
    </div>
  )
}
