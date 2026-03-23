import type { SelectHTMLAttributes } from 'react'
import { cn } from '@/lib/utils'

interface SelectFieldProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label: string
  helperText?: string
  error?: string
}

export function SelectField({
  children,
  className,
  error,
  helperText,
  id,
  label,
  ...props
}: SelectFieldProps) {
  const selectId = id ?? label.toLowerCase().replace(/\s+/g, '-')

  return (
    <label className="block space-y-2" htmlFor={selectId}>
      <span className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">
        {label}
      </span>
      <select
        id={selectId}
        className={cn(
          'w-full border-0 border-b bg-transparent px-0 py-3 text-sm text-on-surface outline-none transition',
          error ? 'border-error focus:border-error' : 'border-outline-variant focus:border-primary',
          className,
        )}
        {...props}
      >
        {children}
      </select>
      {helperText ? <p className="text-xs text-on-surface-variant">{helperText}</p> : null}
      {error ? <p className="text-xs font-medium text-error">{error}</p> : null}
    </label>
  )
}
