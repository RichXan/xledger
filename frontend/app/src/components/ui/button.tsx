import type { ButtonHTMLAttributes, PropsWithChildren } from 'react'
import { cn } from '@/lib/utils'

type ButtonVariant = 'primary' | 'secondary' | 'ghost'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant
}

const variants: Record<ButtonVariant, string> = {
  primary:
    'bg-primary-gradient text-white shadow-ambient hover:brightness-110 active:scale-[0.98]',
  secondary:
    'bg-surface-container-high text-primary hover:bg-surface-container-highest',
  ghost:
    'bg-transparent text-primary hover:bg-surface-container-low',
}

export function Button({
  children,
  className,
  variant = 'primary',
  type = 'button',
  ...props
}: PropsWithChildren<ButtonProps>) {
  return (
    <button
      type={type}
      className={cn(
        'inline-flex items-center justify-center gap-2 rounded-2xl px-5 py-3 text-sm font-semibold transition duration-200',
        variants[variant],
        className,
      )}
      {...props}
    >
      {children}
    </button>
  )
}
