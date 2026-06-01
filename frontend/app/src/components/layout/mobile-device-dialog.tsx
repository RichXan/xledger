import { QRCodeSVG } from 'qrcode.react'
import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { DialogShell } from '@/components/ui/dialog-shell'

interface MobileDeviceDialogProps {
  onClose: () => void
}

function getMobileEntryUrl() {
  if (typeof window === 'undefined') {
    return '/mobile'
  }

  return new URL('/mobile', window.location.origin).toString()
}

export function MobileDeviceDialog({ onClose }: MobileDeviceDialogProps) {
  const { t } = useTranslation()
  const mobileUrl = useMemo(() => getMobileEntryUrl(), [])

  return (
    <DialogShell
      title={t('layout.mobileDevice.title')}
      description={t('layout.mobileDevice.description')}
      onClose={onClose}
      className="max-w-md"
      footer={
        <Button variant="secondary" onClick={onClose}>
          {t('common.close')}
        </Button>
      }
    >
      <div className="flex flex-col items-center gap-4">
        <div className="rounded-2xl border border-outline/15 bg-white p-4 shadow-sm">
          <QRCodeSVG value={mobileUrl} size={280} level="M" includeMargin aria-label={t('layout.mobileDevice.qrAlt')} />
        </div>
        <p className="max-w-[320px] break-all text-center text-xs text-on-surface-variant">{mobileUrl}</p>
        {mobileUrl.includes('127.0.0.1') || mobileUrl.includes('localhost') ? (
          <p className="rounded-xl border border-amber-200 bg-amber-50 px-3 py-2 text-xs font-medium text-amber-900">
            {t('layout.mobileDevice.localhostHint')}
          </p>
        ) : null}
      </div>
    </DialogShell>
  )
}
