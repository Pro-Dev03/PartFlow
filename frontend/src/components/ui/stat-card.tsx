import { Card, CardContent } from './card';
import { TrendingUp, TrendingDown } from 'lucide-react';

interface StatCardProps {
  title: string;
  value: number | string;
  icon: any;
  subtitle?: string;
  variant?: 'default' | 'featured' | 'warning' | 'ai' | 'danger' | 'success' | 'info';
  trend?: string | null;
  trendUp?: boolean | null;
  onClick?: () => void;
}

export function StatCard({
  title,
  value,
  icon: Icon,
  subtitle,
  variant = 'default',
  trend,
  trendUp,
  onClick
}: StatCardProps) {
  return (
    <Card
      variant={variant}
      hoverable
      onClick={onClick}
    >
      <CardContent>
        <div style={{
          display: 'flex',
          justifyContent: 'space-between',
          color: 'var(--text-secondary)',
          fontSize: '12px'
        }}>
          <span>{title}</span>
          <div style={{
            width: '30px',
            height: '30px',
            borderRadius: '6px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            background: variant === 'featured' ? 'var(--color-primary-10)' :
                       variant === 'warning' ? 'var(--color-warning-10)' :
                       variant === 'danger' ? 'var(--color-danger-10)' :
                       variant === 'success' ? 'var(--color-success-10)' :
                       variant === 'info' ? 'var(--color-info-10)' :
                       variant === 'ai' ? 'var(--color-primary-10)' :
                       'var(--color-primary-10)'
          }}>
            <Icon style={{
              width: '20px',
              height: '20px',
              color: variant === 'featured' ? 'var(--color-primary)' :
                     variant === 'warning' ? 'var(--color-warning)' :
                     variant === 'danger' ? 'var(--color-danger)' :
                     variant === 'success' ? 'var(--color-success)' :
                     variant === 'info' ? 'var(--color-info)' :
                     variant === 'ai' ? 'var(--color-primary)' :
                     'var(--color-primary)'
            }} />
          </div>
        </div>
        <div style={{
          marginTop: '10px',
          fontSize: '22px',
          fontWeight: '600',
          color: 'var(--text-primary)',
          lineHeight: '1.2'
        }}>
          {value}
        </div>
        {subtitle && (
          <div style={{
            marginTop: '4px',
            color: 'var(--text-secondary)',
            fontSize: '11px'
          }}>
            {subtitle}
          </div>
        )}
        {trend && (
          <div style={{
            marginTop: '6px',
            color: trendUp ? 'var(--color-success)' : 'var(--color-danger)',
            fontSize: '11px',
            display: 'flex',
            alignItems: 'center',
            gap: '4px'
          }}>
            {trendUp ? (
              <TrendingUp style={{ width: '12px', height: '12px' }} />
            ) : (
              <TrendingDown style={{ width: '12px', height: '12px' }} />
            )}
            {trend}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
