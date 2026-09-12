import { useNavigate } from 'react-router-dom';
import { useTranslation } from '../../../hooks/useTranslation';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { AlertCircle, Package, AlertTriangle, CheckCircle, ArrowRight } from 'lucide-react';

interface LowStockItem {
  id: string;
  product_name: string;
  quantity: number;
  min_stock_level: number;
}

interface OverdueDebtItem {
  id: string;
  customer_name: string;
  remaining_amount: number;
  due_date: string;
  days_overdue: number;
}

interface AttentionSectionProps {
  lowStockCount?: number;
  overdueDebtsCount?: number;
  lowStockItems?: LowStockItem[];
  overdueDebtItems?: OverdueDebtItem[];
  unpaidDebtsCount?: number;
  unpaidDebtItems?: OverdueDebtItem[];
}

export function AttentionSection({
  lowStockCount = 0,
  overdueDebtsCount = 0,
  lowStockItems = [],
  overdueDebtItems = [],
  unpaidDebtsCount = 0,
  unpaidDebtItems = [],
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
          {/* Low Stock Alert with Details */}
          {lowStockCount > 0 && (
            <div style={{
              padding: '16px',
              borderRadius: 'var(--radius-md)',
              background: 'var(--color-warning-08)',
              border: '1px solid var(--color-warning-20)'
            }}>
              <div className="flex items-center gap-3 mb-3">
                <div className="w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0" style={{ background: 'var(--color-warning-15)' }}>
                  <Package className="w-5 h-5" style={{ color: 'var(--color-warning)' }} />
                </div>
                <div>
                  <p style={{ fontSize: '14px', fontWeight: '600', color: 'var(--text-primary)' }}>
                    <span className="numeric-quantity">{lowStockCount}</span> {t('dashboard.lowStock')}
                  </p>
                  <p style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
                    منتجات وصلت للحد الأدنى - {t('dashboard.urgent')}
                  </p>
                </div>
              </div>
              
              {/* Detailed List */}
              {lowStockItems.length > 0 && (
                <div className="flex flex-col gap-2 mb-3">
                  {lowStockItems.slice(0, 3).map((item) => (
                    <div key={item.id} className="flex items-center justify-between" style={{
                      padding: '8px 12px',
                      borderRadius: 'var(--radius-sm)',
                      background: 'var(--color-warning-05)'
                    }}>
                      <div className="flex-1">
                        <p style={{ fontSize: '13px', fontWeight: '500', color: 'var(--text-primary)' }}>
                          {item.product_name}
                        </p>
                        <p style={{ fontSize: '11px', color: 'var(--text-secondary)' }}>
                          الكمية: {item.quantity} / الحد الأدنى: {item.min_stock_level}
                        </p>
                      </div>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => navigate(`/app/inventory?search=${item.product_name}`)}
                        style={{ color: 'var(--color-warning)' }}
                      >
                        <ArrowRight className="w-4 h-4" />
                      </Button>
                    </div>
                  ))}
                </div>
              )}
              
              <Button
                variant="secondary"
                onClick={() => navigate('/app/inventory')}
                className="w-full"
                style={{ background: 'var(--color-warning-10)', border: '1px solid var(--color-warning-20)' }}
              >
                عرض كل المنتجات منخفضة المخزون ←
              </Button>
            </div>
          )}

          {/* Unpaid Debts Alert with Details */}
          {unpaidDebtsCount > 0 && (
            <div style={{
              padding: '16px',
              borderRadius: 'var(--radius-md)',
              background: 'var(--color-danger-08)',
              border: '1px solid var(--color-danger-20)'
            }}>
              <div className="flex items-center gap-3 mb-3">
                <div className="w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0" style={{ background: 'var(--color-danger-15)' }}>
                  <AlertTriangle className="w-5 h-5" style={{ color: 'var(--color-danger)' }} />
                </div>
                <div>
                  <p style={{ fontSize: '14px', fontWeight: '600', color: 'var(--text-primary)' }}>
                    <span className="numeric-quantity">{unpaidDebtsCount}</span> عملاء لديهم ديون غير مسددة
                  </p>
                  <p style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
                    تحتاج متابعة التحصيل
                  </p>
                </div>
              </div>

              {unpaidDebtItems.length > 0 && (
                <div className="flex flex-col gap-2 mb-3">
                  {unpaidDebtItems.slice(0, 3).map((debt) => (
                    <div key={debt.id} className="flex items-center justify-between" style={{
                      padding: '8px 12px',
                      borderRadius: 'var(--radius-sm)',
                      background: 'var(--color-danger-05)'
                    }}>
                      <div className="flex-1">
                        <p style={{ fontSize: '13px', fontWeight: '500', color: 'var(--text-primary)' }}>
                          {debt.customer_name}
                        </p>
                        <p style={{ fontSize: '11px', color: 'var(--text-secondary)' }}>
                          غير مسدد: ₪{debt.remaining_amount.toLocaleString()}
                          {debt.days_overdue > 0 ? ` | متأخر ${debt.days_overdue} يوم` : ' | غير متأخر بعد'}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>
              )}

              <Button
                variant="secondary"
                onClick={() => navigate('/app/debts')}
                className="w-full"
                style={{ background: 'var(--color-danger-10)', border: '1px solid var(--color-danger-20)' }}
              >
                عرض الديون وتسجيل التحصيل ←
              </Button>
            </div>
          )}

          {/* Legacy overdue source remains as a fallback while debt rows load. */}
          {overdueDebtsCount > 0 && unpaidDebtsCount === 0 && (
            <div style={{
              padding: '16px',
              borderRadius: 'var(--radius-md)',
              background: 'var(--color-danger-08)',
              border: '1px solid var(--color-danger-20)'
            }}>
              <div className="flex items-center gap-3 mb-3">
                <div className="w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0" style={{ background: 'var(--color-danger-15)' }}>
                  <AlertTriangle className="w-5 h-5" style={{ color: 'var(--color-danger)' }} />
                </div>
                <div>
                  <p style={{ fontSize: '14px', fontWeight: '600', color: 'var(--text-primary)' }}>
                    <span className="numeric-quantity">{overdueDebtsCount}</span> {t('dashboard.overdueDebts')}
                  </p>
                  <p style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
                    ديون متأخرة - {t('dashboard.critical')}
                  </p>
                </div>
              </div>
              
              {/* Detailed List */}
              {overdueDebtItems.length > 0 && (
                <div className="flex flex-col gap-2 mb-3">
                  {overdueDebtItems.slice(0, 3).map((debt) => (
                    <div key={debt.id} className="flex items-center justify-between" style={{
                      padding: '8px 12px',
                      borderRadius: 'var(--radius-sm)',
                      background: 'var(--color-danger-05)'
                    }}>
                      <div className="flex-1">
                        <p style={{ fontSize: '13px', fontWeight: '500', color: 'var(--text-primary)' }}>
                          {debt.customer_name}
                        </p>
                        <p style={{ fontSize: '11px', color: 'var(--text-secondary)' }}>
                          المبلغ: ₪{debt.remaining_amount.toLocaleString()} | متأخر {debt.days_overdue} يوم
                        </p>
                      </div>
                      <div className="flex gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => navigate(`/app/debts`)}
                          style={{ color: 'var(--color-danger)' }}
                        >
                          <ArrowRight className="w-4 h-4" />
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
              
              <Button
                variant="secondary"
                onClick={() => navigate('/app/debts')}
                className="w-full"
                style={{ background: 'var(--color-danger-10)', border: '1px solid var(--color-danger-20)' }}
              >
                عرض كل الديون المتأخرة ←
              </Button>
            </div>
          )}

          {/* No Alerts */}
          {lowStockCount === 0 && overdueDebtsCount === 0 && unpaidDebtsCount === 0 && (
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
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}