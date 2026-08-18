import { Card } from '../../../components/ui/Card'
import { useTranslation } from 'react-i18next'

interface StatCardProps {
  title: string
  value: string | number
  icon?: string
  trend?: {
    value: number
    isPositive: boolean
  }
  loading?: boolean
}

export function StatCard({ title, value, icon, trend, loading }: StatCardProps) {
  const { t } = useTranslation()

  if (loading) {
    return (
      <Card>
        <div className="space-y-3">
          <div className="h-4 bg-neutral-200 rounded w-1/2 animate-pulse" />
          <div className="h-8 bg-neutral-200 rounded w-3/4 animate-pulse" />
        </div>
      </Card>
    )
  }

  return (
    <Card>
      <div className="flex items-start justify-between">
        <div className="space-y-1">
          <p className="text-sm text-muted">{title}</p>
          <p className="text-2xl font-bold text-text">{value}</p>
        </div>
        {icon && (
          <span className="text-2xl opacity-70">{icon}</span>
        )}
      </div>
      
      {trend && (
        <div className="mt-3 flex items-center gap-1 text-sm">
          <span className={trend.isPositive ? 'text-success' : 'text-danger'}>
            {trend.isPositive ? '↑' : '↓'} {Math.abs(trend.value)}%
          </span>
          <span className="text-muted">
            عن الأمس
          </span>
        </div>
      )}
    </Card>
  )
}