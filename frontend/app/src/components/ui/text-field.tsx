import { forwardRef, type InputHTMLAttributes } from 'react'
import { cn } from '@/lib/utils'

interface TextFieldProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string
  helperText?: string
  error?: string
}

export const TextField = forwardRef<HTMLInputElement, TextFieldProps>(function TextField(
  { className, error, helperText, id, label, ...props },
  ref,
) {
  const inputId = id ?? label.toLowerCase().replace(/\s+/g, '-')

  return (
    <label className="block space-y-2" htmlFor={inputId}>
      <span className="font-label text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">
        {label}
      </span>
      <input
        ref={ref}
        id={inputId}
        className={cn(
          'w-full border-0 border-b bg-transparent px-0 py-3 text-sm text-on-surface outline-none transition placeholder:text-outline',
          error ? 'border-error focus:border-error' : 'border-outline-variant focus:border-primary',
          className,
        )}
        {...props}
      />
      {helperText ? <p className="text-xs text-on-surface-variant">{helperText}</p> : null}
      {error ? <p className="text-xs font-medium text-error">{error}</p> : null}
    </label>
  )
})
