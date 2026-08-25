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
      background: 'var(--color-primary-08)',
      border: '1px solid var(--color-primary-20)'
    }}>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <AlertCircle className="w-5 h-5" style={{ color: 'var(--warning)' }} />
          {t('dashboard.attentionSection')}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col gap-4">
          {/* Low Stock Alert */}
          {lowStockCount > 0 && (
            <div className="flex gap-4 items-start" style={{
              padding: '16px',
              borderRadius: 'var(--radius-md)',
              background: 'var(--color-warning-08)',
              border: '1px solid var(--color-warning-20)'
            }}>
              <div className="w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0" style={{ background: 'var(--color-warning-15)' }}>
                <Package className="w-5 h-5" style={{ color: 'var(--color-warning)' }} />
              </div>
              <div className="flex-1">
                <p style={{ fontSize: '14px', fontWeight: '600', color: 'var(--text-primary)', marginBottom: '4px' }}>
                  <span className="numeric-quantity">{lowStockCount}</span> {t('dashboard.lowStock')}
                </p>
                <p style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
                  {t('dashboard.lowStock')} - {t('dashboard.urgent')}
                </p>
                <Button
                  variant="secondary"
                  onClick={() => navigate('/app/inventory')}
                  className="mt-2 text-xs"
                  style={{ background: 'var(--color-warning-10)', border: '1px solid var(--color-warning-20)' }}
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
              borderRadius: 'var(--radius-md)',
              background: 'var(--color-danger-08)',
              border: '1px solid var(--color-danger-20)'
            }}>
              <div className="w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0" style={{ background: 'var(--color-danger-15)' }}>
                <AlertTriangle className="w-5 h-5" style={{ color: 'var(--color-danger)' }} />
              </div>
              <div className="flex-1">
                <p style={{ fontSize: '14px', fontWeight: '600', color: 'var(--text-primary)', marginBottom: '4px' }}>
                  <span className="numeric-quantity">{overdueDebtsCount}</span> {t('dashboard.overdueDebts')}
                </p>
                <p style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
                  {t('dashboard.overdueDebts')} - {t('dashboard.critical')}
                </p>
                <Button
                  variant="secondary"
                  onClick={() => navigate('/app/debts')}
                  className="mt-2 text-xs"
                  style={{ background: 'var(--color-danger-10)', border: '1px solid var(--color-danger-20)' }}
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
              borderRadius: 'var(--radius-md)',
              background: 'var(--color-success-10)',
              border: '1px solid var(--color-success-20)',
              textAlign: 'center'
            }}>
              <CheckCircle className="w-5 h-5 mx-auto mb-2" style={{ color: 'var(--color-success)' }} />
              <p style={{ fontSize: '14px', fontWeight: '600', color: 'var(--color-success)' }}>
                لا توجد تنبيهات حالية
              </p>
              <p style={{ fontSize: '12px', color: 'var(--text-secondary)', marginTop: '4px' }}>
                كل شيء يعمل بشكل طبيعي
              </p>
              <p style={{ fontSize: '11px', color: 'var(--text-muted)', marginTop: '8px' }}>
                * ميزات التنبيهات المتقدمة قيد التطوير
              </p>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}