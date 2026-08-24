import { Bell, TrendingUp, Package, AlertTriangle, CheckCircle, Clock, Sparkles, ArrowRight, X } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import { Button } from '../ui/button';
import { Badge } from '../ui/badge';

interface SmartAlert {
  id: string;
  type: 'predictive' | 'recommendation' | 'insight' | 'opportunity' | 'warning' | 'info';
  title: string;
  message: string;
  priority: 'urgent' | 'high' | 'medium' | 'low';
  actionLabel?: string;
  onAction?: () => void;
  dismissible?: boolean;
  onDismiss?: () => void;
  timestamp?: string;
}

interface SmartAlertsProps {
  alerts: SmartAlert[];
  onDismiss?: (id: string) => void;
}

export function SmartAlerts({ alerts, onDismiss }: SmartAlertsProps) {
  const getAlertIcon = (type: SmartAlert['type']) => {
    switch (type) {
      case 'predictive':
        return <TrendingUp className="w-4 h-4" style={{ color: '#6366F1' }} />;
      case 'recommendation':
        return <Sparkles className="w-4 h-4" style={{ color: '#EC4899' }} />;
      case 'insight':
        return <CheckCircle className="w-4 h-4" style={{ color: '#22C55E' }} />;
      case 'opportunity':
        return <Package className="w-4 h-4" style={{ color: '#F59E0B' }} />;
      case 'warning':
        return <AlertTriangle className="w-4 h-4" style={{ color: '#EF4444' }} />;
      case 'info':
        return <Bell className="w-4 h-4" style={{ color: '#06B6D4' }} />;
      default:
        return <Bell className="w-4 h-4" style={{ color: '#06B6D4' }} />;
    }
  };

  const getAlertColor = (type: SmartAlert['type']) => {
    switch (type) {
      case 'predictive':
        return 'rgba(99, 102, 241, 0.1)';
      case 'recommendation':
        return 'rgba(236, 72, 153, 0.1)';
      case 'insight':
        return 'rgba(34, 197, 94, 0.1)';
      case 'opportunity':
        return 'rgba(245, 158, 11, 0.1)';
      case 'warning':
        return 'rgba(239, 68, 68, 0.1)';
      case 'info':
        return 'rgba(6, 182, 212, 0.1)';
      default:
        return 'rgba(6, 182, 212, 0.1)';
    }
  };

  const getAlertBorderColor = (type: SmartAlert['type']) => {
    switch (type) {
      case 'predictive':
        return 'rgba(99, 102, 241, 0.3)';
      case 'recommendation':
        return 'rgba(236, 72, 153, 0.3)';
      case 'insight':
        return 'rgba(34, 197, 94, 0.3)';
      case 'opportunity':
        return 'rgba(245, 158, 11, 0.3)';
      case 'warning':
        return 'rgba(239, 68, 68, 0.3)';
      case 'info':
        return 'rgba(6, 182, 212, 0.3)';
      default:
        return 'rgba(6, 182, 212, 0.3)';
    }
  };

  const getPriorityBadge = (priority: SmartAlert['priority']) => {
    switch (priority) {
      case 'urgent':
        return <Badge variant="danger" className="text-xs">عاجل</Badge>;
      case 'high':
        return <Badge variant="warning" className="text-xs">عالي</Badge>;
      case 'medium':
        return <Badge variant="info" className="text-xs">متوسط</Badge>;
      case 'low':
        return <Badge variant="secondary" className="text-xs">منخفض</Badge>;
      default:
        return null;
    }
  };

  if (alerts.length === 0) {
    return (
      <Card style={{
        background: 'rgba(34, 197, 94, 0.05)',
        border: '1px solid rgba(34, 197, 94, 0.2)'
      }}>
        <CardContent style={{ padding: '24px', textAlign: 'center' }}>
          <CheckCircle className="w-8 h-8 mx-auto mb-3" style={{ color: '#22C55E' }} />
          <p style={{ fontSize: '14px', fontWeight: '600', color: '#22C55E', marginBottom: '4px' }}>
            كل شيء يعمل بشكل ممتاز!
          </p>
          <p style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
            النظام يدير متجرك بدلاً من أنت
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
      {alerts.map((alert) => (
        <div
          key={alert.id}
          style={{
            padding: '16px',
            borderRadius: '12px',
            background: getAlertColor(alert.type),
            border: `1px solid ${getAlertBorderColor(alert.type)}`,
            transition: 'all 200ms ease',
            position: 'relative'
          }}
          className="hover-lift"
        >
          {alert.dismissible && (
            <button
              onClick={() => onDismiss?.(alert.id)}
              style={{
                position: 'absolute',
                top: '12px',
                left: '12px',
                background: 'transparent',
                border: 'none',
                color: 'var(--text-secondary)',
                cursor: 'pointer',
                padding: '4px',
                borderRadius: '4px',
                transition: 'all 150ms ease'
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = 'rgba(0, 0, 0, 0.1)';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = 'transparent';
              }}
            >
              <X className="w-4 h-4" />
            </button>
          )}
          
          <div style={{ display: 'flex', gap: '12px', alignItems: 'flex-start' }}>
            <div style={{
              width: '32px',
              height: '32px',
              borderRadius: '8px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              background: getAlertColor(alert.type),
              flexShrink: 0
            }}>
              {getAlertIcon(alert.type)}
            </div>
            
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '4px' }}>
                <p style={{ fontSize: '14px', fontWeight: '600', color: 'var(--text-primary)' }}>
                  {alert.title}
                </p>
                {getPriorityBadge(alert.priority)}
              </div>
              <p style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '8px' }}>
                {alert.message}
              </p>
              
              <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                {alert.timestamp && (
                  <div style={{ display: 'flex', alignItems: 'center', gap: '4px', fontSize: '11px', color: 'var(--text-tertiary)' }}>
                    <Clock className="w-3 h-3" />
                    {alert.timestamp}
                  </div>
                )}
                
                {alert.actionLabel && alert.onAction && (
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={alert.onAction}
                    style={{
                      padding: '6px 12px',
                      fontSize: '12px',
                      background: getAlertColor(alert.type),
                      border: `1px solid ${getAlertBorderColor(alert.type)}`,
                      color: 'var(--text-primary)'
                    }}
                  >
                    {alert.actionLabel}
                    <ArrowRight className="w-3 h-3" style={{ marginLeft: '4px' }} />
                  </Button>
                )}
              </div>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}