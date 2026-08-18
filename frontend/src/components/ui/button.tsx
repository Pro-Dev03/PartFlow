import * as React from "react"
import { Slot } from '@radix-ui/react-slot'

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'ghost' | 'destructive' | 'accent'
  size?: 'sm' | 'md' | 'lg'
  asChild?: boolean
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'primary', size = 'md', asChild = false, ...props }, ref) => {
    const baseStyles = "inline-flex items-center justify-center rounded-lg font-medium transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-accent focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
    
    const variantStyles = {
      primary: "bg-surface border border-border text-text hover:bg-surface-elevated active:scale-98",
      secondary: "bg-surface-elevated border border-border text-text-dim hover:bg-surface-higher active:scale-98",
      ghost: "bg-transparent text-text-dim hover:bg-surface-elevated active:scale-98",
      destructive: "bg-danger-dim border border-danger text-danger hover:bg-danger active:scale-98",
      accent: "bg-accent text-[#04140F] border border-accent hover:bg-accent-hover active:scale-98",
    }
    
    const sizeStyles = {
      sm: "px-3 py-1.5 text-sm rounded-[5px]",
      md: "px-4 py-2 text-base rounded-[8px]",
      lg: "px-6 py-3 text-lg rounded-[8px]",
    }

    const Comp = asChild ? Slot : "button"

    return (
      <Comp
        ref={ref}
        className={`${baseStyles} ${variantStyles[variant]} ${sizeStyles[size]} ${className || ''}`}
        {...props}
      />
    )
  }
)
Button.displayName = "Button"

export { Button }