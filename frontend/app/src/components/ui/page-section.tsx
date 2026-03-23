import type { PropsWithChildren, ReactNode } from 'react'

interface PageSectionProps extends PropsWithChildren {
  eyebrow?: string
  title: string
  description?: string
  actions?: ReactNode
}

export function PageSection({
  eyebrow,
  title,
  description,
  actions,
  children,
}: PageSectionProps) {
  return (
    <section className="rounded-[32px] bg-surface-container-lowest p-8 shadow-ambient">
      <div className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
        <div className="max-w-3xl">
          {eyebrow ? (
            <p className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-primary">
              {eyebrow}
            </p>
          ) : null}
          <h2 className="mt-3 font-headline text-3xl font-extrabold tracking-tight text-on-surface">
            {title}
          </h2>
          {description ? <p className="mt-3 text-sm leading-6 text-on-surface-variant">{description}</p> : null}
        </div>
        {actions ? <div className="flex items-center gap-3">{actions}</div> : null}
      </div>
      {children ? <div className="mt-8">{children}</div> : null}
    </section>
  )
}
