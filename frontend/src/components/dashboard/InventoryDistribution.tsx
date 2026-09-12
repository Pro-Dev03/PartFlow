import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import { Package, TrendingUp, AlertTriangle } from 'lucide-react';

interface CategoryItem {
  name: string;
  count: number;
  value: number;
  taxInclusiveValue?: number;
  color: string;
  status?: 'good' | 'low' | 'critical';
}

interface InventoryDistributionProps {
  data: CategoryItem[];
  totalValue?: number;
  totalItems?: number;
}

export function InventoryDistribution({ 
  data, 
  totalValue = 0
}: InventoryDistributionProps) {
  const availableItem = data.find((item) => item.name === 'متاح');
  const availableItems = availableItem?.count ?? 0;
  const availableValue = availableItem?.value ?? 0;

  const getStatusColor = (status?: string) => {
    switch (status) {
      case 'good': return 'var(--color-success)';
      case 'low': return 'var(--color-warning)';
      case 'critical': return 'var(--color-danger)';
      default: return 'var(--text-muted)';
    }
  };

  const getStatusText = (status?: string) => {
    switch (status) {
      case 'good': return 'جيد';
      case 'low': return 'منخفض';
      case 'critical': return 'حرج';
      default: return '—';
    }
  };

  const getStatusIcon = (status?: string) => {
    switch (status) {
      case 'good': return TrendingUp;
      case 'low': return AlertTriangle;
      case 'critical': return AlertTriangle;
      default: return Package;
    }
  };

  return (
    <Card style={{
      background: 'var(--bg-surface)',
      border: '1px solid var(--border-default)'
    }}>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Package className="w-5 h-5" style={{ color: 'var(--color-info)' }} />
          حالة المخزون
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
          {data.map((item, index) => {
            const percentage = totalValue > 0 ? (item.value / totalValue) * 100 : 0;
            const StatusIcon = getStatusIcon(item.status);
            const statusColor = getStatusColor(item.status);
            
            return (
              <div
                key={index}
                className="group"
                style={{
                  padding: '14px',
                  borderRadius: 'var(--radius-lg)',
                  border: '1px solid var(--border-default)',
                  background: 'var(--bg-surface-elevated)',
                  transition: 'all 200ms ease',
                  cursor: 'pointer'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.borderColor = 'var(--color-primary-25)';
                  e.currentTarget.style.transform = 'translateX(4px)';
                  e.currentTarget.style.boxShadow = 'var(--shadow-sm)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.borderColor = 'var(--border-default)';
                  e.currentTarget.style.transform = 'translateX(0)';
                  e.currentTarget.style.boxShadow = 'none';
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '10px' }}>
                  {/* Category Icon */}
                  <div
                    className="flex-shrink-0"
                    style={{
                      width: '36px',
                      height: '36px',
                      borderRadius: 'var(--radius-md)',
                      background: `${item.color}15`,
                      border: `1px solid ${item.color}30`,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center'
                    }}
                  >
                    <Package 
                      className="w-4 h-4" 
                      style={{ color: item.color }} 
                    />
                  </div>

                  {/* Category Info */}
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ 
                      display: 'flex', 
                      alignItems: 'center', 
                      justifyContent: 'space-between',
                      marginBottom: '4px'
                    }}>
                      <span style={{ 
                        fontSize: '13px', 
                        fontWeight: '600', 
                        color: 'var(--text-primary)',
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        whiteSpace: 'nowrap'
                      }}>
                        {item.name}
                      </span>
                      <div style={{ 
                        display: 'flex', 
                        alignItems: 'center', 
                        gap: '6px',
                        flexShrink: 0
                      }}>
                        <StatusIcon 
                          className="w-3.5 h-3.5" 
                          style={{ color: statusColor }} 
                        />
                        <span style={{ 
                          fontSize: '11px', 
                          fontWeight: '500',
                          color: statusColor 
                        }}>
                          {getStatusText(item.status)}
                        </span>
                      </div>
                    </div>

                    {/* Stats Row */}
                    <div style={{ 
                      display: 'flex', 
                      alignItems: 'center', 
                      justifyContent: 'space-between',
                      marginBottom: '8px'
                    }}>
                      <div style={{ display: 'flex', gap: '12px' }}>
                        <span style={{ 
                          fontSize: '11px', 
                          color: 'var(--text-secondary)' 
                        }}>
                          {item.count} قطعة
                        </span>
                        <span style={{ 
                          fontSize: '11px', 
                          fontWeight: '600',
                          color: 'var(--text-primary)',
                          fontFamily: 'var(--font-family-mono)'
                        }}>
                          ₪{item.value.toLocaleString()}
                        </span>
                      </div>
                      <span style={{ 
                        fontSize: '11px', 
                        fontWeight: '600',
                        color: item.color,
                        fontFamily: 'var(--font-family-mono)'
                      }}>
                        {percentage.toFixed(1)}%
                      </span>
                    </div>
                    {item.taxInclusiveValue && item.taxInclusiveValue > item.value && (
                      <div style={{
                        fontSize: '10px',
                        color: 'var(--text-secondary)',
                        marginBottom: '8px'
                      }}>
                        مباع شامل الضريبة: ₪{item.taxInclusiveValue.toLocaleString()}
                      </div>
                    )}

                    {/* Progress Bar */}
                    <div 
                      style={{
                        height: '6px',
                        background: 'var(--bg-surface-3)',
                        borderRadius: 'var(--radius-full)',
                        overflow: 'hidden',
                        boxShadow: 'var(--shadow-inner)'
                      }}
                    >
                      <div
                        style={{
                          height: '100%',
                          width: `${Math.min(percentage, 100)}%`,
                          background: `linear-gradient(90deg, ${item.color} 0%, ${item.color}99 100%)`,
                          borderRadius: 'var(--radius-full)',
                          transition: 'width 600ms cubic-bezier(0.4, 0, 0.2, 1)',
                          boxShadow: `0 0 8px ${item.color}40`
                        }}
                      />
                    </div>
                  </div>
                </div>
              </div>
            );
          })}

          {/* Total Summary */}
          <div 
            style={{
              padding: '12px 14px',
              borderRadius: 'var(--radius-lg)',
              background: 'var(--color-primary-08)',
              border: '1px solid var(--color-primary-15)',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              marginTop: '4px'
            }}
          >
            <span style={{ 
              fontSize: '12px', 
              fontWeight: '600', 
              color: 'var(--text-secondary)' 
            }}>
              المتاح للبيع
            </span>
            <div style={{ display: 'flex', gap: '16px' }}>
              <span style={{ 
                fontSize: '12px', 
                color: 'var(--text-secondary)' 
              }}>
                {availableItems} قطعة
              </span>
              <span style={{ 
                fontSize: '13px', 
                fontWeight: '700',
                color: 'var(--color-primary)',
                fontFamily: 'var(--font-family-mono)'
              }}>
                ₪{availableValue.toLocaleString()}
              </span>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
