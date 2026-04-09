import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { PageSection } from '@/components/ui/page-section'
import { useGenerateShortcut } from '@/features/management/management-hooks'

export function ShortcutPage() {
  const [generatedToken, setGeneratedToken] = useState<string | null>(null)
  const [apiEndpoint, setApiEndpoint] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const generateMutation = useGenerateShortcut()

  async function handleGenerateShortcut() {
    try {
      const result = await generateMutation.mutateAsync({ name: 'Quick Entry' })
      setGeneratedToken(result.pat_token)
      setApiEndpoint(result.api_endpoint)
      setCopied(false)
    } catch (e: any) {
      console.error('API Error:', e)
      alert('Failed to generate credentials: ' + (e.message || String(e)))
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
        eyebrow="Quick Entry"
        title="Apple Shortcuts"
        description="Enable lightning-fast expense tracking using Apple Shortcuts. Supports OCR receipts, natural language voice input, and more."
      >
        <div className="grid gap-6 xl:grid-cols-[1.2fr_0.8fr]">
          <article className="rounded-[28px] bg-surface-container-low p-6">
            <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">
              Installation Flow
            </p>

            <div className="mt-6 space-y-4">
              <div className="rounded-2xl bg-surface-container-lowest p-4">
                <p className="text-sm font-medium text-on-surface">Step 1: Generate Credentials</p>
                <p className="mt-1 text-xs text-on-surface-variant">
                  Generate your unique API token and endpoint needed to authenticate the shortcut.
                </p>
                <Button
                  className="mt-3"
                  onClick={() => void handleGenerateShortcut()}
                  disabled={generateMutation.isPending}
                >
                  {generateMutation.isPending ? 'Generating...' : 'Generate New Credentials'}
                </Button>
              </div>

              {generatedToken && (
                <div className="rounded-2xl border-2 border-primary/20 bg-surface-container-lowest p-4">
                  <p className="text-sm font-medium text-on-surface">Step 2: Install Shortcut</p>
                  <p className="mt-1 text-xs text-on-surface-variant">
                    Tap the button below to download the shortcut. Once it opens, paste your URL and Token when prompted.
                  </p>

                  <div className="mt-4 space-y-3 rounded-xl bg-surface-container p-3">
                    <div>
                      <p className="text-xs font-medium text-on-surface-variant">API Endpoint</p>
                      <div className="mt-1 flex items-center justify-between gap-2">
                        <code className="text-xs text-primary font-mono truncate">
                          {apiEndpoint}/api/shortcuts/quick-add
                        </code>
                        <Button variant="ghost" onClick={handleCopyApiEndpoint}>
                          Copy
                        </Button>
                      </div>
                    </div>
                    <div className="border-t border-outline/10" />
                    <div>
                      <p className="text-xs font-medium text-on-surface-variant">Access Token</p>
                      <div className="mt-1 flex items-center justify-between gap-2">
                        <code className="text-xs text-primary font-mono truncate">
                          {generatedToken}
                        </code>
                        <Button variant="ghost" onClick={handleCopyToken}>
                          {copied ? 'Copied' : 'Copy'}
                        </Button>
                      </div>
                    </div>
                  </div>

                  <div className="mt-4 rounded-xl bg-surface-container-high p-4 text-xs text-on-surface">
                    <p className="font-bold mb-2">手动创建快捷指令步骤 (Manual Setup):</p>
                    <ol className="list-decimal pl-4 space-y-1.5 opacity-90">
                      <li>在您的 iPhone 上打开「快捷指令 (Shortcuts)」App</li>
                      <li>点击右上角「+」新建快捷指令</li>
                      <li>添加操作：搜索并选择「获取 URL 内容 (Get Contents of URL)」</li>
                      <li>填入上方复制的 <strong>API Endpoint</strong></li>
                      <li>展开该操作，将方法设为 <strong>POST</strong></li>
                      <li>在头部 (Headers) 中添加：
                        <br/>键(Key): <code>Authorization</code>
                        <br/>值(Value): <code>Bearer 您的Token</code> (记得保留Bearer前缀)
                      </li>
                      <li>请求体 (Request Body) 选择 JSON，并填入您的账单数据（例如 amount, type, category）</li>
                    </ol>
                  </div>
                </div>
              )}
            </div>
          </article>

          <article className="rounded-[28px] bg-surface-container-low p-6">
            <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">
              Features & Usage
            </p>

            <div className="mt-6 space-y-4">
              <div className="rounded-2xl bg-surface-container-lowest p-4">
                <p className="text-sm font-medium text-on-surface">Ways to Trigger</p>
                <ul className="mt-2 space-y-1 text-xs text-on-surface-variant">
                  <li>• iPhone Back Tap (Double/Triple tap)</li>
                  <li>• "Hey Siri, Quick Entry"</li>
                  <li>• Home Screen iOS Widgets</li>
                  <li>• Control Center icon</li>
                </ul>
              </div>

              <div className="rounded-2xl bg-surface-container-lowest p-4">
                <p className="text-sm font-medium text-on-surface">Supported Actions</p>
                <ul className="mt-2 space-y-1 text-xs text-on-surface-variant">
                  <li>• Standard manual amount entry</li>
                  <li>• OCR Receipt scanning (Upcoming)</li>
                  <li>• Natural language processing (Upcoming)</li>
                  <li>• Auto-category inference (Upcoming)</li>
                </ul>
              </div>

              <div className="rounded-2xl bg-primary-container p-4">
                <p className="text-sm font-medium text-on-primary-container">Security Note</p>
                <p className="mt-2 text-xs text-on-primary-container">
                  Your token remains valid permanently. If you suspect a leak, generating a new token will instantly invalidate the old one.
                </p>
              </div>
            </div>
          </article>
        </div>
      </PageSection>
    </div>
  )
}
