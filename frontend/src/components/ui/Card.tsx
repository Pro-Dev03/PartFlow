import { HTMLAttributes, forwardRef } from 'react'
import { clsx } from 'clsx'

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  hover?: boolean
  padding?: 'none' | 'sm' | 'md' | 'lg'
}

export const Card = forwardRef<HTMLDivElement, CardProps>(
  ({ 
    hover = false, 
    padding = 'md',
    className,
    children,
    ...props 
  }, ref) => {
    const baseStyles = 'bg-surface border border-border rounded-lg shadow-sm'
    
    const hoverStyles = hover ? 'hover:shadow-md transition-shadow cursor-pointer' : ''
    
    const paddingStyles = {
      none: '',
      sm: 'p-3',
      md: 'p-4',
      lg: 'p-6',
    }
    
    return (
      <div
        ref={ref}
        className={clsx(
          baseStyles,
          hoverStyles,
          paddingStyles[padding],
          className
        )}
        {...props}
      >
        {children}
      </div>
    )
  }
)

Card.displayName = 'Card'

interface CardHeaderProps extends HTMLAttributes<HTMLDivElement> {
  title?: string
  description?: string
}

export const CardHeader = forwardRef<HTMLDivElement, CardHeaderProps>(
  ({ title, description, className, children, ...props }, ref) => (
    <div ref={ref} className={clsx('mb-4', className)} {...props}>
      {title && (
        <h3 className="text-lg font-semibold text-text">{title}</h3>
      )}
      {description && (
        <p className="text-sm text-muted mt-1">{description}</p>
      )}
      {children}
    </div>
  )
)

CardHeader.displayName = 'CardHeader'

export const CardContent = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div ref={ref} className={clsx('', className)} {...props} />
  )
)

CardContent.displayName = 'CardContent'

export const CardFooter = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div 
      ref={ref} 
      className={clsx('mt-4 pt-4 border-t border-border flex items-center justify-end gap-2', className)} 
      {...props} 
    />
  )
)

CardFooter.displayName = 'CardFooter'