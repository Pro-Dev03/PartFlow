import { Card, CardContent } from './card';
import { TrendingUp, TrendingDown } from 'lucide-react';
import type { CSSProperties } from 'react';

interface StatCardProps {
  title: string;
  value: number | string;
  icon: any;
  subtitle?: string;
  variant?: 'default' | 'featured' | 'warning' | 'ai' | 'danger' | 'success' | 'info';
  trend?: string | null;
  trendUp?: boolean | null;
  onClick?: () => void;
  compact?: boolean;
  size?: 'sm' | 'md' | 'lg';
}

export function StatCard({
  title,
  value,
  icon: Icon,
  subtitle,
  variant = 'default',
  trend,
  trendUp,
  onClick,
  compact = false,
  size = 'md'
}: StatCardProps) {
  const accentColor = variant === 'featured' ? 'var(--color-primary)' :
    variant === 'warning' ? 'var(--color-warning)' :
    variant === 'danger' ? 'var(--color-danger)' :
    variant === 'success' ? 'var(--color-success)' :
    variant === 'info' ? 'var(--color-info)' : 'var(--color-primary)';

  const sizes = {
    sm: { icon: 26, iconInner: 15, title: 11, value: 18, subtitle: 10, padding: '12px 16px', marginTop: '6px', trendMargin: '4px', trendSize: 10 },
    md: { icon: 30, iconInner: 20, title: 12, value: 22, subtitle: 11, padding: '18px 20px', marginTop: '10px', trendMargin: '6px', trendSize: 11 },
    lg: { icon: 44, iconInner: 26, title: 14, value: 34, subtitle: 13, padding: '26px 30px', marginTop: '14px', trendMargin: '10px', trendSize: 13 }
  };

  const s = compact
    ? { icon: 26, iconInner: 16, title: 11, value: 18, subtitle: 11, padding: '10px 12px', marginTop: '6px', trendMargin: '2px', trendSize: 11 }
    : sizes[size];

  const iconBg = variant === 'featured' ? 'var(--color-primary-10)' :
    variant === 'warning' ? 'var(--color-warning-10)' :
    variant === 'danger' ? 'var(--color-danger-10)' :
    variant === 'success' ? 'var(--color-success-10)' :
    variant === 'info' ? 'var(--color-info-10)' :
    variant === 'ai' ? 'var(--color-primary-10)' :
    'var(--color-primary-10)';

  const iconColor = variant === 'featured' ? 'var(--color-primary)' :
    variant === 'warning' ? 'var(--color-warning)' :
    variant === 'danger' ? 'var(--color-danger)' :
    variant === 'success' ? 'var(--color-success)' :
    variant === 'info' ? 'var(--color-info)' :
    variant === 'ai' ? 'var(--color-primary)' :
    'var(--color-primary)';

  return (
    <Card
      variant={variant}
      hoverable
      onClick={onClick}
      className={`unified-stat-card ${compact ? 'compact-stat-card' : ''} ${size === 'sm' ? 'sm-stat-card' : ''} ${size === 'lg' ? 'lg-stat-card' : ''}`}
      style={{ '--stat-accent': accentColor } as CSSProperties}
    >
      <CardContent className="unified-stat-card-content" style={{ padding: s.padding }}>
        <div style={{
          display: 'flex',
          justifyContent: 'space-between',
          color: 'var(--text-secondary)',
          fontSize: `${s.title}px`
        }}>
          <span>{title}</span>
          <div style={{
            width: `${s.icon}px`,
            height: `${s.icon}px`,
            borderRadius: '10px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            background: iconBg
          }}>
            <Icon style={{
              width: `${s.iconInner}px`,
              height: `${s.iconInner}px`,
              color: iconColor
            }} />
          </div>
        </div>
        <div style={{
          marginTop: s.marginTop,
          fontSize: `${s.value}px`,
          fontWeight: '600',
          color: 'var(--text-primary)',
          lineHeight: '1.2'
        }}>
          {value}
        </div>
        {subtitle && (
          <div style={{
            marginTop: size === 'lg' ? '6px' : '4px',
            color: 'var(--text-secondary)',
            fontSize: `${s.subtitle}px`
          }}>
            {subtitle}
          </div>
        )}
        {trend && (
          <div style={{
            marginTop: s.trendMargin,
            color: trendUp ? 'var(--color-success)' : 'var(--color-danger)',
            fontSize: `${s.trendSize}px`,
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
