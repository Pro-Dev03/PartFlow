import { useState } from 'react';
import { useTranslation } from '../../../hooks/useTranslation';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { PageHeader } from '../../../components/ui/page-header';
import { Badge } from '../../../components/ui/badge';
import { exportToCSV, printTable } from '../../../lib/export-utils';
import { getButtonSize } from '../../../config/button-sizes';
import {
  BarChart3,
  Download,
  Printer,
  Sparkles,
  Target,
  Zap,
  RotateCcw
} from 'lucide-react';

// Custom hooks
import { useReports } from '../hooks/useReports';

// Components
import { ReportTypeSelector } from '../components/ReportTypeSelector';
import { DateRangeSelector } from '../components/DateRangeSelector';
import { ReportStats } from '../components/ReportStats';
import { ReportCharts } from '../components/ReportCharts';

// Types
import { ReportType, DateRange } from '../types/reports.types';

function getReportRows(payload: unknown): Record<string, unknown>[] {
  if (Array.isArray(payload)) {
    return payload.filter((item): item is Record<string, unknown> => Boolean(item) && typeof item === 'object');
  }

  function getDisplayValue(item: Record<string, unknown>): unknown {
    return item.value ?? item.amount ?? item.revenue ?? item.net_revenue ?? item.cost ??
      item.total_cost ?? item.refund_amount ?? item.profit ?? item.net_profit ?? 0;
  }
  if (!payload || typeof payload !== 'object') return [];

  const entries = Object.entries(payload);
  const nestedRows = entries.find(([, value]) => Array.isArray(value) && value.length > 0)?.[1];
  if (Array.isArray(nestedRows)) {
    return nestedRows.filter((item): item is Record<string, unknown> => Boolean(item) && typeof item === 'object');
  }

  return entries
    .filter(([, value]) => value === null || typeof value !== 'object')
    .map(([key, value]) => ({ key, value }));
}

export function ReportsPage() {
  const { t } = useTranslation();
  const [selectedReport, setSelectedReport] = useState('sales');
  const [dateRange, setDateRange] = useState('thisMonth');
  const [customStartDate, setCustomStartDate] = useState('');
  const [customEndDate, setCustomEndDate] = useState('');

  // Custom hook
  const { reportData, reportLoading, reportError, refetch } = useReports(selectedReport, dateRange, customStartDate, customEndDate);

  const handleExport = () => {
    const report = reportData?.data ?? reportData;
    if (!report) return;
    
    const rows = getReportRows(report);
    const dataToExport = rows.map((item: any) => ({
      'التاريخ': item.date || new Date().toLocaleDateString('ar-SA'),
      'القيمة': getDisplayValue(item),
      'الوصف': item.description || item.name || item.product_name || item.customer_name || item.supplier_name || item.category_name || item.key || '-',
      'الحالة': item.status || 'مكتمل',
    }));
    
    exportToCSV(dataToExport, `${selectedReport}-report-${new Date().toISOString().split('T')[0]}`);
  };

  const handlePrint = () => {
    const report = reportData?.data ?? reportData;
    if (!report) return;
    
    const rows = getReportRows(report);
    const dataToPrint = rows.map((item: any) => ({
      'التاريخ': item.date || new Date().toLocaleDateString('ar-SA'),
      'القيمة': getDisplayValue(item),
      'الوصف': item.description || item.name || item.product_name || item.customer_name || item.supplier_name || item.category_name || item.key || '-',
      'الحالة': item.status || 'مكتمل',
    }));
    
    printTable(dataToPrint, ['التاريخ', 'القيمة', 'الوصف', 'الحالة'], selectedReportType?.label || 'تقرير');
  };

  const reportTypes: ReportType[] = [
    { id: 'sales', label: t('reports.salesReport'), icon: BarChart3, group: 'period' },
    { id: 'net-sales', label: 'المبيعات الصافية', icon: Target, group: 'period' },
    { id: 'profit', label: t('reports.profitReport'), icon: Target, group: 'period' },
    { id: 'purchases', label: 'تقرير المشتريات', icon: BarChart3, group: 'period' },
    { id: 'expenses', label: t('reports.expensesReport'), icon: BarChart3, group: 'period' },
    { id: 'returns', label: 'تقرير المرتجعات', icon: RotateCcw, group: 'period' },
    { id: 'inventory', label: t('reports.inventoryReport'), icon: BarChart3, group: 'current' },
    { id: 'used-items', label: 'تقرير القطع المستعملة', icon: Zap, group: 'current' },
    { id: 'debts', label: t('reports.debtsReport'), icon: Target, group: 'current' },
    { id: 'products', label: 'دليل المنتجات', icon: BarChart3, group: 'current' },
    { id: 'suppliers', label: t('reports.suppliersReport'), icon: BarChart3, group: 'current' },
  ];

  const dateRanges: DateRange[] = [
    { value: 'today', label: t('reports.today') },
    { value: 'thisWeek', label: t('reports.thisWeek') },
    { value: 'thisMonth', label: t('reports.thisMonth') },
    { value: 'thisYear', label: t('reports.thisYear') },
    { value: 'custom', label: t('reports.custom') },
  ];

  const selectedReportType = reportTypes.find(r => r.id === selectedReport);
  const ReportIcon = selectedReportType?.icon || BarChart3;
  const isPeriodReport = ['sales', 'net-sales', 'profit', 'purchases', 'expenses', 'returns'].includes(selectedReport);
  const reportPayload = reportData?.data ?? reportData;
  const reportRows = getReportRows(reportPayload);
  const salesReport = selectedReport === 'sales' && reportPayload && typeof reportPayload === 'object'
    ? reportPayload as Record<string, unknown>
    : null;
  const salesRevenue = Number(salesReport?.total_revenue ?? 0);
  const salesProfit = Number(salesReport?.gross_profit ?? 0);
  const salesCount = Number(salesReport?.total_sales ?? 0);
  const topProductName = Array.isArray(salesReport?.top_products)
    ? String((salesReport.top_products[0] as Record<string, unknown> | undefined)?.product_name ?? '')
    : '';

  return (
    <div>
      {/* Page Header */}
      <PageHeader
        eyebrow="Analytics Hub"
        title={t('reports.title')}
        description="تحليلات وتقارير شاملة عن أداء المحل مع رؤى ذكية"
        actions={
          <div style={{ display: 'flex', gap: '10px' }}>
            <Button variant="secondary" size={getButtonSize('reports', 'headerActions')} onClick={handleExport} disabled={!reportData?.data && !reportData}>
              <Download className="w-4 h-4 mr-2" />
              {t('reports.export')}
            </Button>
            <Button variant="secondary" size={getButtonSize('reports', 'headerActions')} onClick={handlePrint} disabled={!reportData?.data && !reportData}>
              <Printer className="w-4 h-4 mr-2" />
              {t('reports.print')}
            </Button>
            <Button variant="secondary" size={getButtonSize('reports', 'headerActions')} onClick={() => refetch()}>
              <Zap className="w-4 h-4 mr-2" />
              تحديث
            </Button>
          </div>
        }
      />

      {/* AI Analytics Insight */}
      <Card variant="ai">
        <CardHeader>
          <CardTitle style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <Sparkles className="w-5 h-5 text-cyan-400" />
            AI Analytics Insight
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div style={{ display: 'flex', gap: '14px' }}>
            <div style={{
              width: '32px',
              height: '32px',
              borderRadius: '8px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              background: 'rgba(34, 211, 238, 0.1)',
              flexShrink: 0
            }}>
              <Target className="w-4 h-4 text-cyan-400" />
            </div>
            <div>
              <p style={{ fontSize: '13px', fontWeight: '600', color: 'var(--text-primary)' }}>
                {selectedReport === 'sales'
                  ? (salesCount > 0 ? 'ملخص أداء المبيعات' : 'لا توجد مبيعات في الفترة')
                  : 'ملخص التقرير'}
              </p>
              <p style={{ fontSize: '11px', color: 'var(--text-secondary)', marginTop: '4px' }}>
                {selectedReport === 'sales' && salesCount > 0
                  ? `تم تنفيذ ${salesCount} عملية بإيراد ₪${salesRevenue.toLocaleString()} وربح إجمالي ₪${salesProfit.toLocaleString()}${topProductName ? `. المنتج الأعلى إيرادًا: ${topProductName}.` : '.'}`
                  : selectedReport === 'sales'
                    ? 'غيّر الفترة أو سجّل عملية بيع لعرض تحليل الأداء والمنتجات الأكثر مبيعًا.'
                    : 'تم تحميل بيانات التقرير للفترة المحددة.'}
              </p>
              {selectedReport === 'sales' && salesCount > 0 && (
                <p style={{ fontSize: '11px', color: 'var(--text-secondary)', marginTop: '6px' }}>
                  هامش الربح المحقق: {salesRevenue > 0 ? ((salesProfit / salesRevenue) * 100).toFixed(1) : '0.0'}%
                </p>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Report Type Selection */}
      <ReportTypeSelector 
        reportTypes={reportTypes}
        selectedReport={selectedReport}
        onSelectReport={setSelectedReport}
      />

      {/* Stats Cards */}
      <ReportStats data={reportData} loading={reportLoading} reportType={selectedReport} />

      {/* Date ranges apply only to reports built from dated transactions. */}
      {isPeriodReport ? (
        <>
          <DateRangeSelector
            dateRanges={dateRanges}
            selectedRange={dateRange}
            onSelectRange={setDateRange}
          />
          {dateRange === 'custom' && (
            <Card>
              <CardContent>
                <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                  <label className="text-sm text-text-secondary">
                    من تاريخ
                    <input
                      type="date"
                      value={customStartDate}
                      onChange={(event) => setCustomStartDate(event.target.value)}
                      className="mt-1 w-full rounded-lg border border-border bg-surface px-3 py-2 text-text-primary"
                    />
                  </label>
                  <label className="text-sm text-text-secondary">
                    إلى تاريخ
                    <input
                      type="date"
                      value={customEndDate}
                      min={customStartDate || undefined}
                      onChange={(event) => setCustomEndDate(event.target.value)}
                      className="mt-1 w-full rounded-lg border border-border bg-surface px-3 py-2 text-text-primary"
                    />
                  </label>
                </div>
              </CardContent>
            </Card>
          )}
        </>
      ) : null}

      {/* Charts Section */}
      <ReportCharts data={reportData} loading={reportLoading} reportType={selectedReport} />

      {/* Report Content */}
      <Card>
        <CardHeader>
          <CardTitle style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <ReportIcon className="w-5 h-5 text-cyan-400" />
            {selectedReportType?.label}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {reportLoading ? (
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '60px 20px' }}>
              <div style={{ animation: 'spin 1s linear infinite', borderRadius: '50%', height: '32px', width: '32px', borderBottom: '2px solid #14b8a6' }} />
            </div>
          ) : reportError ? (
            <div style={{ padding: '40px 20px', textAlign: 'center', color: 'var(--color-error)' }}>
              تعذر تحميل التقرير. تحقق من الاتصال ثم حاول مرة أخرى.
            </div>
          ) : selectedReport === 'sales' && reportPayload && typeof reportPayload === 'object' ? (
            <div>
              <p className="text-sm text-text-secondary" style={{ marginBottom: '12px' }}>
                المنتجات الأكثر مبيعًا وربحًا خلال الفترة المحددة.
              </p>
              <div className="horizontal-scroll">
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                    {['المنتج', 'الوحدات', 'الإيراد', 'الربح', 'هامش الربح'].map((heading) => (
                      <th key={heading} style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>{heading}</th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {Array.isArray((reportPayload as Record<string, unknown>).top_products) && (reportPayload as Record<string, unknown[]>).top_products.length > 0 ? (
                    ((reportPayload as Record<string, unknown[]>).top_products as Record<string, unknown>[]).map((product, index) => {
                      const revenue = Number(product.revenue || 0);
                      const profit = Number(product.profit || 0);
                      return (
                        <tr key={String(product.product_id || index)} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                          <td style={{ padding: '12px', color: 'var(--text-primary)', fontSize: '13px', fontWeight: '600' }}>{String(product.product_name || 'غير معروف')}</td>
                          <td style={{ padding: '12px', color: 'var(--text-secondary)', fontSize: '13px' }}>{String(product.quantity || 0)}</td>
                          <td style={{ padding: '12px', color: 'var(--color-primary)', fontSize: '13px', fontWeight: '600' }}>₪{revenue.toLocaleString()}</td>
                          <td style={{ padding: '12px', color: profit >= 0 ? 'var(--color-success)' : 'var(--color-error)', fontSize: '13px', fontWeight: '600' }}>₪{profit.toLocaleString()}</td>
                          <td style={{ padding: '12px', color: 'var(--text-secondary)', fontSize: '13px' }}>{revenue > 0 ? `${((profit / revenue) * 100).toFixed(1)}%` : '0.0%'}</td>
                        </tr>
                      );
                    })
                  ) : (
                    <tr><td colSpan={5} style={{ padding: '24px', textAlign: 'center', color: 'var(--text-secondary)' }}>لا توجد مبيعات في الفترة المحددة</td></tr>
                  )}
                </tbody>
              </table>
              </div>
            </div>
          ) : selectedReport === 'used-items' ? (
            // تقرير القطع المستعملة - عرض خاص
            <div className="horizontal-scroll">
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>المنتج</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>اشتريت بـ</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>بعت بـ</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>الربح</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>الكمية المباعة</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>الحالة</th>
                  </tr>
                </thead>
                <tbody>
                  {reportRows.length > 0 ? (
                    reportRows.map((item: any, index: number) => {
                      const purchaseCost = Number(item.purchase_cost || 0);
                      const sellingPrice = Number(item.selling_price || 0);
                      const profit = sellingPrice - purchaseCost;
                      return (
                        <tr key={index} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                          <td style={{ padding: '12px', color: 'var(--text-primary)', fontSize: '13px' }}>
                            {item.product_name || item.name || '-'}
                          </td>
                          <td style={{ padding: '12px', color: 'var(--text-secondary)', fontSize: '13px' }}>
                            ₪{purchaseCost.toFixed(2)}
                          </td>
                          <td style={{ padding: '12px', color: 'var(--color-success)', fontSize: '13px', fontWeight: '600' }}>
                            ₪{sellingPrice.toFixed(2)}
                          </td>
                          <td style={{ padding: '12px', color: profit > 0 ? 'var(--color-success)' : 'var(--color-error)', fontSize: '13px', fontWeight: '600' }}>
                            ₪{profit.toFixed(2)}
                          </td>
                          <td style={{ padding: '12px', color: 'var(--text-primary)', fontSize: '13px' }}>
                            {item.status === 'SOLD' ? '1' : '0'}
                          </td>
                          <td style={{ padding: '12px' }}>
                            <Badge variant={item.status === 'AVAILABLE' ? 'success' : item.status === 'SOLD' ? 'default' : 'warning'}>
                              {item.status === 'AVAILABLE' ? 'متاحة' : item.status === 'SOLD' ? 'مباعة' : item.status}
                            </Badge>
                          </td>
                        </tr>
                      );
                    })
                  ) : (
                    <tr>
                      <td colSpan={7} style={{ padding: '24px', textAlign: 'center', color: 'var(--text-secondary)' }}>
                        لا توجد قطع مستعملة
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          ) : selectedReport === 'products' && reportPayload && typeof reportPayload === 'object' ? (
            <div className="horizontal-scroll">
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead><tr style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                  <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)' }}>المؤشر</th>
                  <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)' }}>القيمة</th>
                </tr></thead>
                <tbody>
                  <tr><td style={{ padding: '12px' }}>إجمالي المنتجات النشطة</td><td style={{ padding: '12px' }}>{String((reportPayload as Record<string, unknown>).total_products ?? 0)}</td></tr>
                  <tr><td style={{ padding: '12px' }}>منتجات منخفضة المخزون</td><td style={{ padding: '12px' }}>{String((reportPayload as Record<string, unknown>).low_stock_count ?? 0)}</td></tr>
                  {Object.entries(((reportPayload as Record<string, unknown>).by_category || {}) as Record<string, unknown>).map(([category, count]) => (
                    <tr key={category}><td style={{ padding: '12px' }}>{category}</td><td style={{ padding: '12px' }}>{String(count)}</td></tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : selectedReport === 'returns' && reportPayload && typeof reportPayload === 'object' ? (
            <div className="horizontal-scroll">
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>المنتج</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>عدد المرتجعات</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>إجمالي الاسترداد</th>
                  </tr>
                </thead>
                <tbody>
                  {Array.isArray((reportPayload as Record<string, unknown>).by_product) && ((reportPayload as Record<string, unknown[]>).by_product as Record<string, unknown>[]).length > 0 ? ((reportPayload as Record<string, unknown[]>).by_product as Record<string, unknown>[]).map((item, index) => (
                    <tr key={index} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                      <td style={{ padding: '12px', color: 'var(--text-primary)', fontSize: '13px' }}>{String(item.product_name ?? 'منتج محذوف')}</td>
                      <td style={{ padding: '12px', color: 'var(--text-primary)', fontSize: '13px' }}>{String(item.return_count ?? 0)}</td>
                      <td style={{ padding: '12px', color: 'var(--color-primary)', fontSize: '13px', fontWeight: '600' }}>₪{Number(item.refund_amount ?? 0).toLocaleString()}</td>
                    </tr>
                  )) : (
                    <tr><td colSpan={3} style={{ padding: '24px', textAlign: 'center', color: 'var(--text-secondary)' }}>لا توجد مرتجعات مكتملة في الفترة المحددة</td></tr>
                  )}
                </tbody>
              </table>
            </div>
          ) : reportRows.length > 0 ? (
            <div className="horizontal-scroll">
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>التاريخ</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>القيمة</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>الوصف</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>الحالة</th>
                  </tr>
                </thead>
                <tbody>
                  {reportRows.map((item: any, index: number) => (
                    <tr key={index} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                      <td style={{ padding: '12px', color: 'var(--text-primary)', fontSize: '13px' }}>
                        {item.date || new Date().toLocaleDateString('ar-SA')}
                      </td>
                      <td style={{ padding: '12px', color: 'var(--color-primary)', fontSize: '13px', fontWeight: '600' }}>
                        ₪{(item.value || item.amount || 0).toLocaleString()}
                      </td>
                      <td style={{ padding: '12px', color: 'var(--text-secondary)', fontSize: '13px' }}>
                        {item.description || item.name || '-'}
                      </td>
                      <td style={{ padding: '12px' }}>
                        <Badge variant={item.status === 'completed' ? 'success' : 'warning'}>
                          {item.status || 'مكتمل'}
                        </Badge>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div style={{ textAlign: 'center', padding: '60px 20px' }}>
              <BarChart3 className="w-16 h-16 text-slate-500 mx-auto mb-5" />
              <p style={{ fontSize: '16px', fontWeight: '600', color: 'var(--text-primary)', marginBottom: '8px' }}>
                لا توجد بيانات
              </p>
              <p style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
                لم يتم العثور على بيانات لهذا التقرير في الفترة المحددة
              </p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
