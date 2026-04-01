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
    const result = await generateMutation.mutateAsync({ name: '快捷记账' })
    setGeneratedToken(result.pat_token)
    setApiEndpoint(result.api_endpoint)
    setCopied(false)
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
        title="快捷记账"
        description="通过 Apple 快捷指令实现快速记账，支持截图识别、语音输入等多种方式。"
      >
        <div className="grid gap-6 xl:grid-cols-[1.2fr_0.8fr]">
          <article className="rounded-[28px] bg-surface-container-low p-6">
            <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">
              快捷指令配置
            </p>

            <div className="mt-6 space-y-4">
              <div className="rounded-2xl bg-surface-container-lowest p-4">
                <p className="text-sm font-medium text-on-surface">第 1 步：生成 Token</p>
                <p className="mt-1 text-xs text-on-surface-variant">
                  点击下方按钮生成专属的访问令牌，用于快捷指令认证。
                </p>
                <Button
                  className="mt-3"
                  onClick={() => void handleGenerateShortcut()}
                  disabled={generateMutation.isPending}
                >
                  {generateMutation.isPending ? '生成中...' : '生成快捷记账 Token'}
                </Button>
              </div>

              {generatedToken && (
                <div className="rounded-2xl bg-surface-container-lowest p-4">
                  <p className="text-sm font-medium text-on-surface">第 2 步：复制配置信息</p>
                  <p className="mt-1 text-xs text-on-surface-variant">
                    将以下信息填入快捷指令中。
                  </p>

                  <div className="mt-3 space-y-3">
                    <div>
                      <p className="text-xs font-medium text-on-surface-variant">API 地址</p>
                      <div className="mt-1 flex items-center gap-2">
                        <code className="flex-1 rounded-lg bg-surface-container p-2 text-xs text-on-surface">
                          {apiEndpoint}/api/shortcuts/quick-add
                        </code>
                        <Button variant="ghost" onClick={handleCopyApiEndpoint}>
                          复制
                        </Button>
                      </div>
                    </div>

                    <div>
                      <p className="text-xs font-medium text-on-surface-variant">Token</p>
                      <div className="mt-1 flex items-center gap-2">
                        <code className="flex-1 rounded-lg bg-surface-container p-2 text-xs text-on-surface truncate">
                          {generatedToken}
                        </code>
                        <Button variant="ghost" onClick={handleCopyToken}>
                          {copied ? '已复制' : '复制'}
                        </Button>
                      </div>
                    </div>
                  </div>
                </div>
              )}

              <div className="rounded-2xl bg-surface-container-lowest p-4">
                <p className="text-sm font-medium text-on-surface">第 3 步：创建快捷指令</p>
                <p className="mt-1 text-xs text-on-surface-variant">
                  在 iPhone 上打开"快捷指令" App，创建新的快捷指令。
                </p>
                <div className="mt-3 space-y-2 text-xs text-on-surface-variant">
                  <p>1. 添加"询问输入"动作，提示输入金额</p>
                  <p>2. 添加"从菜单中选择"动作，选项为：支出、收入</p>
                  <p>3. 添加"从菜单中选择"动作，选项为：餐饮、交通、购物等</p>
                  <p>4. 添加"URL"动作，填入上方的 API 地址</p>
                  <p>5. 添加"获取 URL 内容"动作，方法选择 POST</p>
                  <p>6. 设置请求头：Authorization: Bearer [Token]</p>
                  <p>7. 设置请求体为 JSON 格式</p>
                </div>
              </div>
            </div>
          </article>

          <article className="rounded-[28px] bg-surface-container-low p-6">
            <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">
              使用说明
            </p>

            <div className="mt-6 space-y-4">
              <div className="rounded-2xl bg-surface-container-lowest p-4">
                <p className="text-sm font-medium text-on-surface">触发方式</p>
                <ul className="mt-2 space-y-1 text-xs text-on-surface-variant">
                  <li>• 轻点手机背面两下</li>
                  <li>• Siri 语音唤醒</li>
                  <li>• 主屏幕小组件</li>
                  <li>• 控制中心快捷方式</li>
                </ul>
              </div>

              <div className="rounded-2xl bg-surface-container-lowest p-4">
                <p className="text-sm font-medium text-on-surface">支持功能</p>
                <ul className="mt-2 space-y-1 text-xs text-on-surface-variant">
                  <li>• 手动输入金额记账</li>
                  <li>• 截图识别金额（即将支持）</li>
                  <li>• 语音输入记账（即将支持）</li>
                  <li>• 自动分类推荐（即将支持）</li>
                </ul>
              </div>

              <div className="rounded-2xl bg-primary-container p-4">
                <p className="text-sm font-medium text-on-primary-container">提示</p>
                <p className="mt-2 text-xs text-on-primary-container">
                  Token 默认永久有效。如需更换 Token，只需重新生成即可，旧 Token 将自动失效。
                </p>
              </div>
            </div>
          </article>
        </div>
      </PageSection>
    </div>
  )
}
