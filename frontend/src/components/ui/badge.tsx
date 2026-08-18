import * as React from "react"

interface BadgeProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: 'default' | 'success' | 'warning' | 'info' | 'outline'
}

const Badge = React.forwardRef<HTMLDivElement, BadgeProps>(
  ({ className, variant = 'default', ...props }, ref) => {
    const baseStyles = "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors"
    
    const variantStyles = {
      default: "bg-ink text-white dark:bg-ink-light dark:text-white",
      success: "bg-refund/10 text-refund dark:bg-refund/20 dark:text-green-400",
      warning: "bg-debt/10 text-debt dark:bg-debt/20 dark:text-red-400",
      info: "bg-seal-soft text-seal dark:bg-seal/20 dark:text-yellow-400",
      outline: "text-ink-muted border border-rule dark:text-gray-400 dark:border-rule-dark",
    }

    return (
      <div
        ref={ref}
        className={`${baseStyles} ${variantStyles[variant]} ${className || ''}`}
        {...props}
      />
    )
  }
)
Badge.displayName = "Badge"

export { Badge }