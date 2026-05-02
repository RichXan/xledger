import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { PageSection } from '@/components/ui/page-section'
import { useGenerateShortcut } from '@/features/management/management-hooks'

export function ShortcutPage() {
  const { t } = useTranslation()
  const [generatedToken, setGeneratedToken] = useState<string | null>(null)
  const [apiEndpoint, setApiEndpoint] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const generateMutation = useGenerateShortcut()

  async function handleGenerateShortcut() {
    try {
      const result = await generateMutation.mutateAsync({ name: t('shortcutPage.eyebrow') })
      setGeneratedToken(result.pat_token)
      setApiEndpoint(result.api_endpoint)
      setCopied(false)
    } catch (error: unknown) {
      console.error('API Error:', error)
      const message = error instanceof Error ? error.message : String(error)
      alert(t('shortcutPage.generateFailed', { message }))
    }
  }

  function handleCopyToken() {
    if (generatedToken) {
      navigator.clipboard.writeText(generatedToken)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  function handleCopyApiEndpoint() {
    if (apiEndpoint) {
      navigator.clipboard.writeText(apiEndpoint)
    }
  }

  return (
    <div className="space-y-8">
      <PageSection
        eyebrow={t('shortcutPage.eyebrow')}
        title={t('shortcutPage.title')}
        description={t('shortcutPage.description')}
      >
        <div className="grid gap-6 xl:grid-cols-[1.2fr_0.8fr]">
          <article className="rounded-[28px] bg-surface-container-low p-6">
            <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">
              {t('shortcutPage.installationFlow')}
            </p>

            <div className="mt-6 space-y-4">
              <div className="rounded-2xl bg-surface-container-lowest p-4">
                <p className="text-sm font-medium text-on-surface">{t('shortcutPage.step1Title')}</p>
                <p className="mt-1 text-xs text-on-surface-variant">
                  {t('shortcutPage.step1Description')}
                </p>
                <Button
                  className="mt-3"
                  onClick={() => void handleGenerateShortcut()}
                  disabled={generateMutation.isPending}
                >
                  {generateMutation.isPending ? t('shortcutPage.generating') : t('shortcutPage.generateCredentials')}
                </Button>
              </div>

              {generatedToken && (
                <div className="rounded-2xl border-2 border-primary/20 bg-surface-container-lowest p-4">
                  <p className="text-sm font-medium text-on-surface">{t('shortcutPage.step2Title')}</p>
                  <p className="mt-1 text-xs text-on-surface-variant">
                    {t('shortcutPage.step2Description')}
                  </p>

                  <div className="mt-4 space-y-3 rounded-xl bg-surface-container p-3">
                    <div>
                      <p className="text-xs font-medium text-on-surface-variant">{t('shortcutPage.apiEndpoint')}</p>
                      <div className="mt-1 flex items-center justify-between gap-2">
                        <code className="text-xs text-primary font-mono truncate">
                          {apiEndpoint}/api/shortcuts/quick-add
                        </code>
                        <Button variant="ghost" onClick={handleCopyApiEndpoint}>
                          {t('common.copy')}
                        </Button>
                      </div>
                    </div>
                    <div className="border-t border-outline/10" />
                    <div>
                      <p className="text-xs font-medium text-on-surface-variant">{t('shortcutPage.accessToken')}</p>
                      <div className="mt-1 flex items-center justify-between gap-2">
                        <code className="text-xs text-primary font-mono truncate">
                          {generatedToken}
                        </code>
                        <Button variant="ghost" onClick={handleCopyToken}>
                          {copied ? t('common.copied') : t('common.copy')}
                        </Button>
                      </div>
                    </div>
                  </div>

                  <div className="mt-4 rounded-xl bg-surface-container-high p-4 text-xs text-on-surface">
                    <p className="font-bold mb-2">{t('shortcutPage.manualSetupTitle')}</p>
                    <ol className="list-decimal pl-4 space-y-1.5 opacity-90">
                      <li>{t('shortcutPage.manualSteps.openShortcuts')}</li>
                      <li>{t('shortcutPage.manualSteps.createShortcut')}</li>
                      <li>{t('shortcutPage.manualSteps.addAction')}</li>
                      <li>{t('shortcutPage.manualSteps.pasteEndpoint')}</li>
                      <li>{t('shortcutPage.manualSteps.setPost')}</li>
                      <li>{t('shortcutPage.manualSteps.addAuthorization')}</li>
                      <li>{t('shortcutPage.manualSteps.setJson')}</li>
                    </ol>
                  </div>
                </div>
              )}
            </div>
          </article>

          <article className="rounded-[28px] bg-surface-container-low p-6">
            <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">
              {t('shortcutPage.featuresUsage')}
            </p>

            <div className="mt-6 space-y-4">
              <div className="rounded-2xl bg-surface-container-lowest p-4">
                <p className="text-sm font-medium text-on-surface">{t('shortcutPage.waysToTrigger')}</p>
                <ul className="mt-2 space-y-1 text-xs text-on-surface-variant">
                  <li>• {t('shortcutPage.triggerBackTap')}</li>
                  <li>• {t('shortcutPage.triggerSiri')}</li>
                  <li>• {t('shortcutPage.triggerWidgets')}</li>
                  <li>• {t('shortcutPage.triggerControlCenter')}</li>
                </ul>
              </div>

              <div className="rounded-2xl bg-surface-container-lowest p-4">
                <p className="text-sm font-medium text-on-surface">{t('shortcutPage.supportedActions')}</p>
                <ul className="mt-2 space-y-1 text-xs text-on-surface-variant">
                  <li>• {t('shortcutPage.manualAmount')}</li>
                  <li>• {t('shortcutPage.ocr')}</li>
                  <li>• {t('shortcutPage.nlp')}</li>
                  <li>• {t('shortcutPage.autoCategory')}</li>
                </ul>
              </div>

              <div className="rounded-2xl bg-primary-container p-4">
                <p className="text-sm font-medium text-on-primary-container">{t('shortcutPage.securityNote')}</p>
                <p className="mt-2 text-xs text-on-primary-container">
                  {t('shortcutPage.securityDescription')}
                </p>
              </div>
            </div>
          </article>
        </div>
      </PageSection>
    </div>
  )
}
