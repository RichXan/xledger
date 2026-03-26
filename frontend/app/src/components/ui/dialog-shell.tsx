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
      className="fixed inset-0 z-50 flex items-center justify-center bg-on-background/10 p-4 backdrop-blur-sm"
      onClick={(event) => {
        if (event.target === event.currentTarget) {
          onClose?.()
        }
      }}
    >
      <div className={cn('relative w-full max-w-3xl rounded-[32px] bg-surface-container-lowest p-8 shadow-ambient', className)}>
        {onClose ? (
          <button
            type="button"
            className="absolute right-6 top-6 grid h-8 w-8 place-items-center rounded-full text-on-surface-variant transition hover:bg-surface-container-low"
            onClick={onClose}
            aria-label="Close dialog"
          >
            <X className="h-5 w-5" />
          </button>
        ) : null}
        <div>
          <h2 className="font-headline text-3xl font-extrabold tracking-tight text-primary">{title}</h2>
          {description ? <p className="mt-2 text-sm text-on-surface-variant">{description}</p> : null}
        </div>
        <div className="mt-8">{children}</div>
        {footer ? <div className="mt-8 flex items-center justify-end gap-3">{footer}</div> : null}
      </div>
    </div>
  )
}
