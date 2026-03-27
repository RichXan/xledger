import type { ButtonHTMLAttributes, PropsWithChildren } from 'react'
import { cn } from '@/lib/utils'

type ButtonVariant = 'primary' | 'secondary' | 'ghost'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant
}

const variants: Record<ButtonVariant, string> = {
  primary:
    'bg-primary-gradient text-white shadow-ambient hover:brightness-110 active:scale-[0.98] disabled:opacity-60 disabled:cursor-not-allowed',
  secondary:
    'border border-outline/20 bg-surface-container-high text-primary hover:bg-surface-container-highest disabled:opacity-60 disabled:cursor-not-allowed',
  ghost:
    'bg-transparent text-primary hover:bg-surface-container-low disabled:opacity-60 disabled:cursor-not-allowed',
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
        'inline-flex items-center justify-center gap-2 rounded-xl px-5 py-2.5 text-sm font-semibold transition duration-200',
        variants[variant],
        className,
      )}
      {...props}
    >
      {children}
    </button>
  )
}
