import { Card, CardHeader } from '../../../components/ui/Card'
import { Badge } from '../../../components/ui/Badge'
import { Button } from '../../../components/ui/Button'
import { Alert } from '../types/dashboard.types'
import { useTranslation } from 'react-i18next'

interface AlertCardProps {
  alert: Alert
}

export function AlertCard({ alert }: AlertCardProps) {
  const { t } = useTranslation()

  const severityColors = {
    low: 'neutral',
    medium: 'warning',
    high: 'warning',
    critical: 'danger',
  } as const

  const severityIcons = {
    low: '🟢',
    medium: '🟠',
    high: '🟠',
    critical: '🔴',
  }

  return (
    <Card padding="sm" className="border-l-4 border-l-danger">
      <div className="flex items-start gap-3">
        <span className="text-xl">{severityIcons[alert.severity]}</span>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <Badge variant={severityColors[alert.severity]} size="sm">
              {alert.severity}
            </Badge>
            <h4 className="font-medium text-text truncate">{alert.title}</h4>
          </div>
          <p className="text-sm text-muted">{alert.description}</p>
          
          {alert.actionLabel && alert.actionLink && (
            <div className="mt-2">
              <Button 
                variant="ghost" 
                size="sm"
                onClick={() => window.location.href = alert.actionLink!}
              >
                {alert.actionLabel}
              </Button>
            </div>
          )}
        </div>
      </div>
    </Card>
  )
}