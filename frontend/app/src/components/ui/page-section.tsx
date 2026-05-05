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
    <section className="rounded-2xl border border-outline/15 bg-surface-container-lowest p-5 shadow-ambient md:p-6">
      <div className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
        <div className="max-w-3xl">
          {eyebrow ? (
            <p className="font-label text-xs font-bold uppercase tracking-[0.08em] text-primary">
              {eyebrow}
            </p>
          ) : null}
          <h2 className="mt-2 font-headline text-3xl font-extrabold leading-tight text-on-surface md:text-4xl">
            {title}
          </h2>
          {description ? <p className="mt-3 max-w-3xl text-sm leading-6 text-on-surface-variant">{description}</p> : null}
        </div>
        {actions ? <div className="flex items-center gap-3">{actions}</div> : null}
      </div>
      {children ? <div className="mt-6">{children}</div> : null}
    </section>
  )
}
