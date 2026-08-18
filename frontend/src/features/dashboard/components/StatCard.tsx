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
  color?: 'primary' | 'success' | 'warning' | 'danger'
}

export function StatCard({ title, value, icon, trend, loading, color = 'primary' }: StatCardProps) {
  const { t } = useTranslation()

  const colorStyles = {
    primary: {
      bg: '#eff6ff',
      border: '#dbeafe',
      iconBg: '#dbeafe',
      iconColor: '#1e40af'
    },
    success: {
      bg: '#ecfdf5',
      border: '#d1fae5',
      iconBg: '#d1fae5',
      iconColor: '#065f46'
    },
    warning: {
      bg: '#fffbeb',
      border: '#fef3c7',
      iconBg: '#fef3c7',
      iconColor: '#92400e'
    },
    danger: {
      bg: '#fef2f2',
      border: '#fee2e2',
      iconBg: '#fee2e2',
      iconColor: '#991b1b'
    },
  }

  const style = colorStyles[color]

  if (loading) {
    return (
      <div style={{
        backgroundColor: style.bg,
        border: `2px solid ${style.border}`,
        borderRadius: '0.5rem',
        padding: '1rem',
        boxShadow: '0 1px 2px 0 rgb(0 0 0 / 0.05)'
      }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
          <div style={{ height: '1rem', backgroundColor: '#e5e7eb', borderRadius: '0.25rem', width: '50%', animation: 'pulse 1.5s cubic-bezier(0.4, 0, 0.6, 1) infinite' }} />
          <div style={{ height: '2rem', backgroundColor: '#e5e7eb', borderRadius: '0.25rem', width: '75%', animation: 'pulse 1.5s cubic-bezier(0.4, 0, 0.6, 1) infinite' }} />
        </div>
      </div>
    )
  }

  return (
    <div style={{
      backgroundColor: style.bg,
      border: `2px solid ${style.border}`,
      borderRadius: '0.75rem',
      padding: '1.5rem',
      boxShadow: '0 1px 2px 0 rgb(0 0 0 / 0.05)',
      transition: 'box-shadow 150ms ease-in-out',
    }}>
      <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between' }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem', flex: 1 }}>
          <p style={{ fontSize: '0.875rem', fontWeight: '500', color: '#475569', margin: 0 }}>{title}</p>
          <p style={{ fontSize: '1.875rem', fontWeight: 'bold', color: '#0f172a', margin: 0 }}>{value}</p>
        </div>
        {icon && (
          <div style={{
            padding: '0.5rem',
            borderRadius: '0.5rem',
            backgroundColor: style.iconBg,
            color: style.iconColor
          }}>
            <span style={{ fontSize: '1.5rem' }}>{icon}</span>
          </div>
        )}
      </div>
      
      {trend && (
        <div style={{ marginTop: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.875rem' }}>
          <span style={{
            fontWeight: '600',
            color: trend.isPositive ? '#059669' : '#dc2626'
          }}>
            {trend.isPositive ? '↑' : '↓'} {Math.abs(trend.value)}%
          </span>
          <span style={{ color: '#64748b' }}>
            عن الأمس
          </span>
        </div>
      )}
    </div>
  )
}