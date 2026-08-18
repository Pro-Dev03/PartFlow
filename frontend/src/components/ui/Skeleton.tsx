import { HTMLAttributes, forwardRef } from 'react'
import { clsx } from 'clsx'

interface SkeletonProps extends HTMLAttributes<HTMLDivElement> {
  variant?: 'text' | 'circular' | 'rectangular'
  width?: string | number
  height?: string | number
  animation?: 'pulse' | 'wave' | 'none'
}

export const Skeleton = forwardRef<HTMLDivElement, SkeletonProps>(
  ({ 
    variant = 'rectangular', 
    width, 
    height,
    animation = 'pulse',
    className,
    ...props 
  }, ref) => {
    const baseStyles = 'bg-neutral-200 dark:bg-neutral-700'
    
    const variants = {
      text: 'rounded',
      circular: 'rounded-full',
      rectangular: 'rounded-md',
    }
    
    const animations = {
      pulse: 'animate-pulse',
      wave: 'animate-pulse',
      none: '',
    }
    
    const style = {
      width: width || (variant === 'text' ? '100%' : 'auto'),
      height: height || (variant === 'text' ? '1rem' : 'auto'),
    }
    
    return (
      <div
        ref={ref}
        className={clsx(
          baseStyles,
          variants[variant],
          animations[animation],
          className
        )}
        style={style}
        {...props}
      />
    )
  }
)

Skeleton.displayName = 'Skeleton'

// Text Skeleton - For simulating text lines
export const TextSkeleton = ({ lines = 3, className }: { lines?: number; className?: string }) => (
  <div className={clsx('space-y-2', className)}>
    {Array.from({ length: lines }).map((_, i) => (
      <Skeleton 
        key={i} 
        variant="text" 
        width={i === lines - 1 ? '60%' : '100%'} 
      />
    ))}
  </div>
)

// Card Skeleton - For simulating card content
export const CardSkeleton = () => (
  <div className="space-y-4">
    <Skeleton variant="rectangular" height={200} />
    <div className="space-y-2">
      <Skeleton variant="text" width="60%" />
      <Skeleton variant="text" />
      <Skeleton variant="text" width="80%" />
    </div>
  </div>
)

// Table Skeleton - For simulating table rows
export const TableSkeleton = ({ rows = 5, columns = 4 }: { rows?: number; columns?: number }) => (
  <div className="space-y-2">
    {Array.from({ length: rows }).map((_, i) => (
      <div key={i} className="flex gap-4">
        {Array.from({ length: columns }).map((_, j) => (
          <Skeleton 
            key={j} 
            variant="text" 
            width={j === 0 ? '30%' : '20%'} 
          />
        ))}
      </div>
    ))}
  </div>
)