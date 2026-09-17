import { useNavigate } from 'react-router-dom';
import { useTranslation } from '../../../hooks/useTranslation';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { AlertCircle, Package, AlertTriangle, CheckCircle, ArrowRight } from 'lucide-react';
import { formatStoreDate } from '../../../utils/store-time';

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
  const outOfStockCount = lowStockItems.filter((item) => item.quantity === 0).length;
  const lowStockOnlyCount = Math.max(lowStockCount - outOfStockCount, 0);

  return (
    <Card variant="ai" style={{
      background: 'transparent',
      border: 'none',
      boxShadow: 'none'
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
              padding: '8px 0 8px 14px',
              borderLeft: '3px solid var(--color-warning)',
              background: 'transparent'
            }}>
              <div className="flex items-center gap-3 mb-3">
                <div className="w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0" style={{ background: 'var(--color-warning-15)' }}>
                  <Package className="w-5 h-5" style={{ color: 'var(--color-warning)' }} />
                </div>
                <div>
                  <p style={{ fontSize: '14px', fontWeight: '600', color: 'var(--text-primary)' }}>
                    <span className="numeric-quantity">{lowStockCount}</span> منتجات تحتاج متابعة
                  </p>
                  <p style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
                    المخزون العام عند الحد الأدنى أو أقل
                  </p>
                </div>
              </div>

              <div className="flex flex-wrap gap-2 mb-3" aria-label="ملخص حالة المخزون">
                <div style={{ padding: '6px 10px', borderRadius: '8px', background: 'var(--color-danger-10)', color: 'var(--color-danger)', fontSize: '12px', fontWeight: 600 }}>
                  نفذ: <span className="numeric-quantity">{outOfStockCount}</span>
                </div>
                <div style={{ padding: '6px 10px', borderRadius: '8px', background: 'var(--color-warning-10)', color: 'var(--color-warning)', fontSize: '12px', fontWeight: 600 }}>
                  عند الحد الأدنى: <span className="numeric-quantity">{lowStockOnlyCount}</span>
                </div>
              </div>
              
              {/* Detailed List */}
              {lowStockItems.length > 0 && (
                <div className="flex flex-col gap-2 mb-3">
                  {lowStockItems.slice(0, 3).map((item) => (
                    <div key={item.id} className="flex items-center justify-between" style={{
                      padding: '8px 12px',
                      borderBottom: `1px solid ${item.quantity === 0 ? 'var(--color-danger-15)' : 'var(--color-warning-15)'}`,
                      background: 'transparent'
                    }}>
                      <div className="flex-1">
                        <p style={{ fontSize: '13px', fontWeight: '500', color: 'var(--text-primary)' }}>
                          {item.product_name}
                        </p>
                        <p style={{ fontSize: '11px', color: 'var(--text-secondary)' }}>
                          <span style={{ color: item.quantity === 0 ? 'var(--color-danger)' : 'inherit', fontWeight: item.quantity === 0 ? 600 : 'normal' }}>
                            {item.quantity === 0 ? 'نفد المخزون' : `المتاح: ${item.quantity}`}
                          </span>
                          {` / الحد الأدنى: ${item.min_stock_level}`}
                        </p>
                      </div>
                      <span style={{ marginInlineEnd: '8px', padding: '4px 8px', borderRadius: '6px', background: item.quantity === 0 ? 'var(--color-danger-10)' : 'var(--color-warning-10)', color: item.quantity === 0 ? 'var(--color-danger)' : 'var(--color-warning)', fontSize: '11px', fontWeight: 600, whiteSpace: 'nowrap' }}>
                        {item.quantity === 0 ? 'شراء عاجل' : 'إعادة طلب'}
                      </span>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => navigate(`/app/inventory?search=${item.product_name}`)}
                        style={{ color: item.quantity === 0 ? 'var(--color-danger)' : 'var(--color-warning)' }}
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
                style={{ background: 'var(--color-warning-10)', border: '1px solid var(--color-warning-20)', borderRadius: '999px' }}
              >
                عرض كل المنتجات منخفضة المخزون ←
              </Button>
            </div>
          )}

          {/* Unpaid Debts Alert with Details */}
          {unpaidDebtsCount > 0 && (
            <div style={{
              padding: '8px 0 8px 14px',
              borderLeft: '3px solid var(--color-danger)',
              background: 'transparent'
            }}>
              <div className="flex items-center gap-3 mb-3">
                <div className="w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0" style={{ background: 'var(--color-danger-15)' }}>
                  <AlertTriangle className="w-5 h-5" style={{ color: 'var(--color-danger)' }} />
                </div>
                <div>
                  <p style={{ fontSize: '14px', fontWeight: '600', color: 'var(--text-primary)' }}>
                    <span className="numeric-quantity">{unpaidDebtsCount}</span> عملاء تجاوزوا موعد السداد
                  </p>
                  <p style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
                    تحتاج متابعة التحصيل الآن
                  </p>
                </div>
              </div>

              {unpaidDebtItems.length > 0 && (
                <div className="flex flex-col gap-2 mb-3">
                  {unpaidDebtItems.slice(0, 3).map((debt) => (
                    <div key={debt.id} className="flex items-center justify-between" style={{
                      padding: '8px 12px',
                      borderBottom: '1px solid var(--color-danger-15)',
                      background: 'transparent'
                    }}>
                      <div className="flex-1">
                        <p style={{ fontSize: '13px', fontWeight: '500', color: 'var(--text-primary)' }}>
                          {debt.customer_name}
                        </p>
                        <p style={{ fontSize: '11px', color: 'var(--text-secondary)' }}>
                          غير مسدد: ₪{debt.remaining_amount.toLocaleString()}
                          {` | استحق في ${formatStoreDate(debt.due_date, 'ar')}`}
                          {debt.days_overdue > 0 ? ` | متأخر ${debt.days_overdue} يوم` : ''}
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
                style={{ background: 'var(--color-danger-10)', border: '1px solid var(--color-danger-20)', borderRadius: '999px' }}
              >
                عرض الديون وتسجيل التحصيل ←
              </Button>
            </div>
          )}

          {/* Legacy overdue source remains as a fallback while debt rows load. */}
          {overdueDebtsCount > 0 && unpaidDebtsCount === 0 && (
            <div style={{
              padding: '8px 0 8px 14px',
              borderLeft: '3px solid var(--color-danger)',
              background: 'transparent'
            }}>
              <div className="flex items-center gap-3 mb-3">
                <div className="w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0" style={{ background: 'var(--color-danger-15)' }}>
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
                      borderBottom: '1px solid var(--color-danger-15)',
                      background: 'transparent'
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
                style={{ background: 'var(--color-danger-10)', border: '1px solid var(--color-danger-20)', borderRadius: '999px' }}
              >
                عرض كل الديون المتأخرة ←
              </Button>
            </div>
          )}

          {/* No Alerts */}
          {lowStockCount === 0 && overdueDebtsCount === 0 && unpaidDebtsCount === 0 && (
            <div style={{
              padding: '20px 12px',
              background: 'transparent',
              borderTop: '2px solid var(--color-success)',
              textAlign: 'center'
            }}>
              <div className="w-10 h-10 rounded-full flex items-center justify-center mx-auto mb-2" style={{ background: 'var(--color-success-10)' }}>
                <CheckCircle className="w-5 h-5" style={{ color: 'var(--color-success)' }} />
              </div>
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