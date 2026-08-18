import { clsx } from 'clsx'

interface MobileSafeAreaProps {
  children: React.ReactNode
  className?: string
}

export function MobileSafeArea({ children, className }: MobileSafeAreaProps) {
  return (
    <div className={clsx('pb-20', className)}>
      {children}
    </div>
  )
}
