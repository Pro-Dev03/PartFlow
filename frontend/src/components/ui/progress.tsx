import { cn } from '../../lib/utils';

interface ProgressProps extends React.HTMLAttributes<HTMLDivElement> {
  value?: number;
  max?: number;
  size?: 'sm' | 'md' | 'lg';
  color?: 'primary' | 'success' | 'warning' | 'error';
}

export function Progress({ 
  value = 0, 
  max = 100, 
  size = 'md',
  color = 'primary',
  className,
  ...props 
}: ProgressProps) {
  const percentage = Math.min(Math.max((value / max) * 100, 0), 100);
  
  const sizeClasses = {
    sm: 'h-1',
    md: 'h-2',
    lg: 'h-4',
  };

  const colorClasses = {
    primary: 'bg-primary-600',
    success: 'bg-green-600',
    warning: 'bg-yellow-600',
    error: 'bg-red-600',
  };

  return (
    <div
      className={cn(
        'w-full bg-gray-200 dark:bg-gray-800 rounded-full overflow-hidden',
        sizeClasses[size],
        className
      )}
      {...props}
    >
      <div
        className={cn(
          'h-full transition-all duration-300 ease-out',
          colorClasses[color]
        )}
        style={{ width: `${percentage}%` }}
      />
    </div>
  );
}

export function CircularProgress({ 
  value = 0, 
  size = 40,
  strokeWidth = 4,
  className 
}: { 
  value?: number; 
  size?: number; 
  strokeWidth?: number;
  className?: string;
}) {
  const radius = (size - strokeWidth) / 2;
  const circumference = radius * 2 * Math.PI;
  const percentage = Math.min(Math.max(value, 0), 100);
  const offset = circumference - (percentage / 100) * circumference;

  return (
    <div 
      className={cn('relative inline-flex items-center justify-center', className)}
      style={{ width: size, height: size }}
    >
      <svg
        className="transform -rotate-90"
        width={size}
        height={size}
      >
        <circle
          className="text-gray-200 dark:text-gray-800"
          stroke="currentColor"
          strokeWidth={strokeWidth}
          fill="transparent"
          r={radius}
          cx={size / 2}
          cy={size / 2}
        />
        <circle
          className="text-primary-600"
          stroke="currentColor"
          strokeWidth={strokeWidth}
          fill="transparent"
          r={radius}
          cx={size / 2}
          cy={size / 2}
          strokeDasharray={circumference}
          strokeDashoffset={offset}
          style={{
            transition: 'stroke-dashoffset 0.3s ease-out',
          }}
        />
      </svg>
      <span className="absolute text-xs font-medium text-gray-700 dark:text-gray-300">
        {Math.round(percentage)}%
      </span>
    </div>
  );
}

export function LinearProgress({ 
  isIndeterminate = false,
  ...props 
}: ProgressProps & { isIndeterminate?: boolean }) {
  if (isIndeterminate) {
    return (
      <div
        className={cn(
          'w-full bg-gray-200 dark:bg-gray-800 rounded-full overflow-hidden h-2',
          props.className
        )}
      >
        <div className="h-full bg-primary-600 animate-pulse w-1/3" />
      </div>
    );
  }

  return <Progress {...props} />;
}
