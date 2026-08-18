import * as React from "react"

interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: 'default' | 'text' | 'circular' | 'rounded'
}

const Skeleton = React.forwardRef<HTMLDivElement, SkeletonProps>(
  ({ className, variant = 'default', ...props }, ref) => {
    const variantStyles = {
      default: "rounded-md",
      text: "h-4 w-full rounded",
      circular: "rounded-full",
      rounded: "rounded-lg",
    }

    return (
      <div
        ref={ref}
        className={`animate-pulse bg-rule ${variantStyles[variant]} ${className || ''}`}
        {...props}
      />
    )
  }
)
Skeleton.displayName = "Skeleton"

const TextSkeleton: React.FC<{ className?: string }> = ({ className }) => (
  <Skeleton variant="text" className={className} />
)

const CardSkeleton: React.FC<{ className?: string }> = ({ className }) => (
  <div className={`p-6 space-y-4 ${className}`}>
    <Skeleton className="h-6 w-1/3" />
    <Skeleton className="h-4 w-full" />
    <Skeleton className="h-4 w-2/3" />
  </div>
)

const TableSkeleton: React.FC<{ rows?: number; className?: string }> = ({ 
  rows = 5, 
  className 
}) => (
  <div className={`space-y-2 ${className}`}>
    {Array.from({ length: rows }).map((_, i) => (
      <div key={i} className="flex space-x-4">
        <Skeleton className="h-12 flex-1" />
        <Skeleton className="h-12 flex-1" />
        <Skeleton className="h-12 flex-1" />
      </div>
    ))}
  </div>
)

export { Skeleton, TextSkeleton, CardSkeleton, TableSkeleton }