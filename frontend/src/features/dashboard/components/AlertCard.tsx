import { Card } from '@components/ui/card'
import { Badge } from '@components/ui/badge'
import { Button } from '@components/ui/button'
import { Alert } from '../types/dashboard.types'
import { useTranslation } from 'react-i18next'

interface AlertCardProps {
  alert: Alert
}

export function AlertCard({ alert }: AlertCardProps) {
  const { t } = useTranslation()

  const severityStyles = {
    low: {
      bg: 'bg-success-50',
      border: 'border-success-200',
      icon: '✓',
      badge: 'success',
      text: 'text-success-700'
    },
    medium: {
      bg: 'bg-warning-50',
      border: 'border-warning-200',
      icon: '⚠️',
      badge: 'warning',
      text: 'text-warning-700'
    },
    high: {
      bg: 'bg-warning-50',
      border: 'border-warning-300',
      icon: '⚠️',
      badge: 'warning',
      text: 'text-warning-700'
    },
    critical: {
      bg: 'bg-danger-50',
      border: 'border-danger-300',
      icon: '🔴',
      badge: 'danger',
      text: 'text-danger-700'
    },
  }

  const style = severityStyles[alert.severity]

  return (
    <Card className={`border-2 ${style.bg} ${style.border} transition-all hover:shadow-md`}>
      <div className="flex items-start gap-4">
        <div className={`flex-shrink-0 w-10 h-10 rounded-full ${style.bg} flex items-center justify-center text-xl`}>
          {style.icon}
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-2">
            <Badge variant={style.badge} size="sm">
              {alert.severity}
            </Badge>
            <h4 className="font-semibold text-text-primary">{alert.title}</h4>
          </div>
          <p className="text-sm text-text-secondary mb-3">{alert.description}</p>
          
          {alert.actionLabel && alert.actionLink && (
            <Button 
              variant="primary" 
              size="sm"
              onClick={() => window.location.href = alert.actionLink!}
              className="w-full sm:w-auto"
            >
              {alert.actionLabel}
            </Button>
          )}
        </div>
      </div>
    </Card>
  )
}