import type { PropsWithChildren, ReactNode } from 'react'
import { cn } from '@/lib/utils'

interface DialogShellProps extends PropsWithChildren {
  title: string
  description?: string
  footer?: ReactNode
  className?: string
}

export function DialogShell({ title, description, footer, className, children }: DialogShellProps) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-on-background/10 p-4 backdrop-blur-sm">
      <div className={cn('w-full max-w-3xl rounded-[32px] bg-surface-container-lowest p-8 shadow-ambient', className)}>
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
