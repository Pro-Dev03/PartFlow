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
        <div style={{ padding: '20px' }}> {/* worktrack: 20px padding */}
          <div style={{
            display: 'flex',
            justifyContent: 'space-between',
            color: 'var(--text-secondary)',
            fontSize: '13px' // worktrack: 13px
          }}>
            <span>{title}</span>
            <div style={{
              width: '36px', // worktrack: slightly larger
              height: '36px',
              borderRadius: '8px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              background: variant === 'featured' ? 'rgba(99, 102, 241, 0.1)' :
                         variant === 'warning' ? 'rgba(245, 158, 11, 0.1)' :
                         variant === 'danger' ? 'rgba(239, 68, 68, 0.1)' :
                         variant === 'success' ? 'rgba(34, 197, 94, 0.1)' :
                         variant === 'info' ? 'rgba(6, 182, 212, 0.1)' :
                         variant === 'ai' ? 'rgba(99, 102, 241, 0.1)' :
                         'rgba(99, 102, 241, 0.1)'
            }}>
              <Icon style={{
                width: '28px', // worktrack: 28px
                height: '28px',
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
            marginTop: '13px',
            fontSize: '28px', // worktrack: 28px
            fontWeight: '600', // worktrack: 600
            color: 'var(--text-primary)'
          }}>
            {value}
          </div>
          {subtitle && (
            <div style={{
              marginTop: '5px',
              color: 'var(--text-secondary)',
              fontSize: '12px' // worktrack: 12px
            }}>
              {subtitle}
            </div>
          )}
          {trend && (
            <div style={{
              marginTop: '7px',
              color: trendUp ? 'var(--color-success)' : 'var(--color-danger)',
              fontSize: '12px', // worktrack: 12px
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
        </div>
      </CardContent>
    </Card>
  );
}
