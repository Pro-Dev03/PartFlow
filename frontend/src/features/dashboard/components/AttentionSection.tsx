import { useNavigate } from 'react-router-dom';
import { useTranslation } from '../../../hooks/useTranslation';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { AlertCircle, Package, AlertTriangle, CheckCircle } from 'lucide-react';

interface AttentionSectionProps {
  lowStockCount?: number;
  overdueDebtsCount?: number;
}

export function AttentionSection({
  lowStockCount = 0,
  overdueDebtsCount = 0
}: AttentionSectionProps) {
  const navigate = useNavigate();
  const { t } = useTranslation();

  return (
    <Card variant="ai" style={{
      background: 'linear-gradient(135deg, rgba(245, 158, 11, 0.1) 0%, rgba(239, 68, 68, 0.1) 100%)',
      border: '1px solid rgba(245, 158, 11, 0.3)'
    }}>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <AlertCircle className="w-5 h-5" style={{ color: '#f59e0b' }} />
          {t('dashboard.attentionSection')}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col gap-4">
          {/* Low Stock Alert */}
          {lowStockCount > 0 && (
            <div className="flex gap-4 items-start" style={{
              padding: '16px',
              borderRadius: '12px',
              background: 'rgba(245, 158, 11, 0.1)',
              border: '1px solid rgba(245, 158, 11, 0.3)'
            }}>
              <div className="w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0" style={{ background: 'rgba(245, 158, 11, 0.2)' }}>
                <Package className="w-5 h-5" style={{ color: '#f59e0b' }} />
              </div>
              <div className="flex-1">
                <p style={{ fontSize: '14px', fontWeight: '600', color: 'var(--text-primary)', marginBottom: '4px' }}>
                  {lowStockCount} {t('dashboard.lowStock')}
                </p>
                <p style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
                  {t('dashboard.lowStock')} - {t('dashboard.urgent')}
                </p>
                <Button
                  variant="secondary"
                  onClick={() => navigate('/app/inventory')}
                  className="mt-2 text-xs"
                  style={{ background: 'rgba(245, 158, 11, 0.2)', border: '1px solid rgba(245, 158, 11, 0.3)' }}
                >
                  {t('dashboard.takeAction')} ←
                </Button>
              </div>
            </div>
          )}

          {/* Overdue Debts Alert */}
          {overdueDebtsCount > 0 && (
            <div className="flex gap-4 items-start" style={{
              padding: '16px',
              borderRadius: '12px',
              background: 'rgba(239, 68, 68, 0.1)',
              border: '1px solid rgba(239, 68, 68, 0.3)'
            }}>
              <div className="w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0" style={{ background: 'rgba(239, 68, 68, 0.2)' }}>
                <AlertTriangle className="w-5 h-5" style={{ color: '#ef4444' }} />
              </div>
              <div className="flex-1">
                <p style={{ fontSize: '14px', fontWeight: '600', color: 'var(--text-primary)', marginBottom: '4px' }}>
                  {overdueDebtsCount} {t('dashboard.overdueDebts')}
                </p>
                <p style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
                  {t('dashboard.overdueDebts')} - {t('dashboard.critical')}
                </p>
                <Button
                  variant="secondary"
                  onClick={() => navigate('/app/debts')}
                  className="mt-2 text-xs"
                  style={{ background: 'rgba(239, 68, 68, 0.2)', border: '1px solid rgba(239, 68, 68, 0.3)' }}
                >
                  {t('dashboard.takeAction')} ←
                </Button>
              </div>
            </div>
          )}

          {/* No Alerts */}
          {lowStockCount === 0 && overdueDebtsCount === 0 && (
            <div style={{
              padding: '20px',
              borderRadius: '12px',
              background: 'rgba(52, 211, 153, 0.1)',
              border: '1px solid rgba(52, 211, 153, 0.3)',
              textAlign: 'center'
            }}>
              <CheckCircle className="w-5 h-5 mx-auto mb-2" style={{ color: '#34D399' }} />
              <p style={{ fontSize: '14px', fontWeight: '600', color: '#34D399' }}>
                لا توجد تنبيهات حالية
              </p>
              <p style={{ fontSize: '12px', color: 'var(--text-secondary)', marginTop: '4px' }}>
                كل شيء يعمل بشكل طبيعي
              </p>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}