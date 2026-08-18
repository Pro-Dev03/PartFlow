import { Button } from '../../../components/ui/Button'
import { QuickAction } from '../types/dashboard.types'

interface QuickActionCardProps {
  action: QuickAction
}

export function QuickActionCard({ action }: QuickActionCardProps) {
  const colorVariants = {
    primary: 'btn-primary',
    success: 'btn-success',
    warning: 'btn-warning',
    danger: 'btn-danger',
  }

  return (
    <Button
      variant={action.color === 'primary' ? 'primary' : action.color === 'success' ? 'success' : action.color === 'warning' ? 'warning' : 'danger'}
      className="w-full h-full flex flex-col items-center justify-center gap-2 p-4 text-center"
      onClick={() => window.location.href = action.path}
    >
      <span className="text-2xl">{action.icon}</span>
      <span className="text-sm font-medium">{action.label}</span>
    </Button>
  )
}