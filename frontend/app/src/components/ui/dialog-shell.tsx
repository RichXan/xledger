import { useEffect, type PropsWithChildren, type ReactNode } from 'react'
import { X } from 'lucide-react'
import { cn } from '@/lib/utils'

interface DialogShellProps extends PropsWithChildren {
  title: string
  description?: string
  footer?: ReactNode
  className?: string
  onClose?: () => void
}

export function DialogShell({ title, description, footer, className, children, onClose }: DialogShellProps) {
  useEffect(() => {
    if (!onClose) {
      return
    }
    const handler = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose()
      }
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [onClose])

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-[#0f172a]/24 p-4 backdrop-blur-sm"
      onClick={(event) => {
        if (event.target === event.currentTarget) {
          onClose?.()
        }
      }}
    >
      <div className={cn('relative max-h-[calc(100vh-32px)] w-full max-w-3xl overflow-y-auto rounded-2xl border border-outline/15 bg-surface-container-lowest p-5 shadow-ambient md:p-6', className)}>
        {onClose ? (
          <button
            type="button"
            className="absolute right-5 top-5 grid h-8 w-8 place-items-center rounded-full border border-transparent text-on-surface-variant transition hover:border-outline/15 hover:bg-surface-container-low"
            onClick={onClose}
            aria-label="Close dialog"
          >
            <X className="h-5 w-5" />
          </button>
        ) : null}
        <div>
          <h2 className="pr-10 font-headline text-3xl font-extrabold leading-tight text-primary md:text-4xl">{title}</h2>
          {description ? <p className="mt-2 text-sm text-on-surface-variant">{description}</p> : null}
        </div>
        <div className="mt-6">{children}</div>
        {footer ? <div className="mt-6 flex flex-wrap items-center justify-end gap-3">{footer}</div> : null}
      </div>
    </div>
  )
}
