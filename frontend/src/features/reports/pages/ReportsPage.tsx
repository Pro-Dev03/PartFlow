import { useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useTranslation } from '../../../hooks/useTranslation';
import { Card, CardContent, CardHeader, CardTitle } from '../../../design-system/components/card';
import { Button } from '../../../design-system/components/button';
import { PageHeader } from '../../../design-system/components/page-header';
import { Badge } from '../../../design-system/components/badge';
import { exportToCSV, printTable } from '../../../lib/export-utils';
import { ReportActions } from '../../../design-system/components/report-actions';
import { getButtonSize } from '../../../config/button-sizes';
import {
  BarChart3,
  Banknote,
  ReceiptText,
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
import { formatStoreDate } from '../../../utils/store-time';

function getDisplayValue(item: Record<string, unknown>): unknown {
  return item.value ?? item.amount ?? item.total_amount ?? item.revenue ?? item.net_revenue ?? item.cost ??
    item.total_cost ?? item.total_purchases ?? item.refund_amount ?? item.profit ?? item.net_profit ?? item.gross_revenue ??
    item.total_debt ?? item.outstanding ?? item.overdue_amount ?? item.paid_amount ?? item.balance ?? 0;
}

function getReportDisplayValue(item: Record<string, unknown>, reportType: string): unknown {
  if (reportType === 'profit' && item.net_profit !== undefined) return item.net_profit;
  return getDisplayValue(item);
}

function formatReportAmount(value: unknown): string {
  if (typeof value === 'number' && Number.isFinite(value)) return value.toLocaleString();
  if (typeof value === 'string' && value.trim() !== '' && Number.isFinite(Number(value))) {
    return Number(value).toLocaleString();
  }
  return '-';
}

function getReportRowDescription(item: Record<string, unknown>): string {
  if (item.description || item.name || item.product_name || item.customer_name || item.supplier_name || item.category_name || item.key) {
    return String(item.description || item.name || item.product_name || item.customer_name || item.supplier_name || item.category_name || item.key);
  }
  if (item.sales !== undefined) return `${Number(item.sales)} عملية بيع`;
  if (item.net_profit !== undefined) return 'صافي ربح الفترة';
  if (item.revenue !== undefined || item.net_revenue !== undefined) return 'صافي مبيعات اليوم';
  return '-';
}

function getReportStatusLabel(value: unknown): string {
  const labels: Record<string, string> = {
    approved: 'مقبول',
    pending: 'قيد الانتظار',
    completed: 'مكتمل',
    rejected: 'مرفوض',
    cancelled: 'ملغي',
    canceled: 'ملغي',
    paid: 'مدفوع',
    received: 'مستلم',
    ordered: 'تم الطلب',
    draft: 'مسودة',
  };
  const status = String(value || 'completed').toLowerCase();
  return labels[status] || String(value || 'مكتمل');
}

function getReportRows(payload: unknown, reportType?: string): Record<string, unknown>[] {
  if (Array.isArray(payload)) {
    return payload.filter((item): item is Record<string, unknown> => Boolean(item) && typeof item === 'object');
  }

  if (!payload || typeof payload !== 'object') return [];

  if (reportType === 'expenses') {
    const expenseRows = (payload as { top_expenses?: unknown }).top_expenses;
    if (Array.isArray(expenseRows)) {
      return expenseRows.filter((item): item is Record<string, unknown> => Boolean(item) && typeof item === 'object');
    }
  }

  if (reportType === 'suppliers') {
    const supplierRows = (payload as { by_supplier?: unknown }).by_supplier;
    if (Array.isArray(supplierRows)) {
      return supplierRows.filter((item): item is Record<string, unknown> => Boolean(item) && typeof item === 'object');
    }
  }

  const entries = Object.entries(payload);
  const nestedRows = entries.find(([, value]) => Array.isArray(value) && value.length > 0)?.[1];
  if (Array.isArray(nestedRows)) {
    const rows = nestedRows.filter((item): item is Record<string, unknown> => Boolean(item) && typeof item === 'object');
    if (reportType === 'expenses') {
      const categories = Object.keys((payload as { by_category?: Record<string, unknown> }).by_category || {});
      const fallbackDescription = categories.length > 0 ? categories.join('، ') : 'مصروفات الفترة';
      return rows.map((row) => ({
        ...row,
        description: row.description || fallbackDescription,
      }));
    }
    return rows;
  }

  const metadataKeys = new Set([
    'period',
    'start_date',
    'end_date',
    'total_revenue',
    'total_cogs',
    'gross_profit',
    'total_expenses',
    'net_profit',
    'profit_margin',
    'currency',
    'generated_at',
    'created_at',
    'updated_at',
  ]);

  return entries
    .filter(([key, value]) => !metadataKeys.has(key) && value !== null && typeof value !== 'object')
    .map(([key, value]) => ({ key, value }));
}

function formatReportDate(value: unknown): string {
  return value ? formatStoreDate(String(value), 'ar-SA') : 'غير محدد';
}

export function ReportsPage() {
  const { t } = useTranslation();
  const [searchParams, setSearchParams] = useSearchParams();
  const reportFromUrl = searchParams.get('report');
  const rangeFromUrl = searchParams.get('range');
  const normalizedReportFromUrl = reportFromUrl === 'sales' || reportFromUrl === 'profit' || reportFromUrl === 'net-sales' || reportFromUrl === 'used-items' ? 'sales-profit' : reportFromUrl === 'suppliers' || reportFromUrl === 'purchases' ? 'purchases-suppliers' : reportFromUrl;
  const initialReport = ['sales-profit', 'tax', 'purchases-suppliers', 'expenses', 'returns', 'inventory', 'debts'].includes(normalizedReportFromUrl || '')
    ? normalizedReportFromUrl as string
    : 'sales';
  const initialDateRange = ['today', 'thisWeek', 'thisMonth', 'thisYear', 'custom'].includes(rangeFromUrl || '')
    ? rangeFromUrl as DateRange
    : 'thisMonth';
  const [selectedReport, setSelectedReport] = useState(initialReport);
  const [dateRange, setDateRange] = useState<DateRange>(initialDateRange);
  const [customStartDate, setCustomStartDate] = useState('');
  const [customEndDate, setCustomEndDate] = useState('');

  const updateReportQuery = (key: string, value: string) => {
    setSearchParams((currentParams) => {
      const nextParams = new URLSearchParams(currentParams);
      nextParams.set(key, value);
      return nextParams;
    }, { replace: true });
  };

  const handleReportSelect = (report: string) => {
    setSelectedReport(report);
    updateReportQuery('report', report);
  };

  const handleDateRangeSelect = (range: DateRange) => {
    setDateRange(range);
    updateReportQuery('range', range);
  };

  // Custom hook
  const { reportData, reportLoading, reportError, refetch } = useReports(selectedReport, dateRange, customStartDate, customEndDate);

  const handleExport = () => {
    const report = reportData?.data ?? reportData;
    if (!report) return;
    
    const rows = getReportRows(report, selectedReport);
    const dataToExport = rows.map((item: any) => ({
      'التاريخ': item.date || new Date().toLocaleDateString('ar-SA'),
      'القيمة': getReportDisplayValue(item, selectedReport),
      'الوصف': getReportRowDescription(item),
      'الحالة': getReportStatusLabel(item.status),
    }));
    
    exportToCSV(dataToExport, `${selectedReport}-report-${new Date().toISOString().split('T')[0]}`);
  };

  const handlePrint = () => {
    const report = reportData?.data ?? reportData;
    if (!report) return;
    
    const rows = getReportRows(report, selectedReport);
    const dataToPrint = rows.map((item: any) => ({
      'التاريخ': item.date || new Date().toLocaleDateString('ar-SA'),
      'القيمة': getReportDisplayValue(item, selectedReport),
      'الوصف': getReportRowDescription(item),
      'الحالة': getReportStatusLabel(item.status),
    }));
    
    printTable(dataToPrint, ['التاريخ', 'القيمة', 'الوصف', 'الحالة'], selectedReportType?.label || 'تقرير');
  };

  const reportTypes: ReportType[] = [
    { id: 'sales-profit', label: 'تقرير المبيعات والأرباح', icon: BarChart3, group: 'period' },
    { id: 'tax', label: 'تقرير الضرائب', icon: Target, group: 'period' },
    { id: 'purchases-suppliers', label: 'تقرير المشتريات والتجار', icon: BarChart3, group: 'period' },
    { id: 'expenses', label: t('reports.expensesReport'), icon: BarChart3, group: 'period' },
    { id: 'returns', label: 'تقرير مرتجعات العملاء', icon: RotateCcw, group: 'period' },
    { id: 'inventory', label: t('reports.inventoryReport'), icon: BarChart3, group: 'current' },
    { id: 'debts', label: t('reports.debtsReport'), icon: Target, group: 'current' },
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
  const isPeriodReport = ['sales-profit', 'tax', 'purchases-suppliers', 'expenses', 'returns'].includes(selectedReport);
  const reportPayload = reportData?.data ?? reportData;
  const reportRows = getReportRows(reportPayload, selectedReport);
  const summaryMetricKeys = new Set([
    'total_revenue',
    'total_cogs',
    'gross_profit',
    'total_expenses',
    'net_profit',
    'profit_margin',
    'gross_revenue',
    'net_revenue',
    'total_sales',
    'total_items_sold',
  ]);
  const hasSummaryMetrics = !!(reportPayload && typeof reportPayload === 'object' && Object.entries(reportPayload).some(([key, value]) => {
    if (!summaryMetricKeys.has(key)) return false;
    if (typeof value === 'number') return Number.isFinite(value);
    if (typeof value === 'string') return value.trim() !== '' && Number.isFinite(Number(value));
    return false;
  }));
  const purchasesReport = selectedReport === 'purchases-suppliers' && reportPayload && typeof reportPayload === 'object'
    ? reportPayload as Record<string, unknown>
    : null;
  const untaxedPurchaseCount = purchasesReport && Number(purchasesReport.untaxed_purchases ?? 0) > 0 || purchasesReport && Number(purchasesReport.tax_amount ?? 0) > 0
    ? Number(purchasesReport?.untaxed_purchases ?? 0)
    : Number(purchasesReport?.total_purchases ?? 0);
  const untaxedPurchaseCost = purchasesReport && Number(purchasesReport.untaxed_purchase_cost ?? 0) > 0 || purchasesReport && Number(purchasesReport.tax_amount ?? 0) > 0
    ? Number(purchasesReport?.untaxed_purchase_cost ?? 0)
    : Number(purchasesReport?.total_cost ?? 0);
  const taxedPurchaseCount = Number(purchasesReport?.total_purchases ?? 0) - untaxedPurchaseCount;
  const taxedPurchaseCost = Number(purchasesReport?.total_cost ?? 0) - untaxedPurchaseCost;
  const supplierReturnCredits = Number(purchasesReport?.supplier_return_credits ?? 0);
  const netPurchases = Number(purchasesReport?.net_purchases ?? Number(purchasesReport?.total_cost ?? 0) - supplierReturnCredits);
  const salesReport = selectedReport === 'sales-profit' && reportPayload && typeof reportPayload === 'object'
    ? reportPayload as Record<string, unknown>
    : null;
  const salesRevenue = Number(salesReport?.total_revenue ?? 0);
  const salesProfit = Number(salesReport?.gross_profit ?? 0);
  const salesCount = Number(salesReport?.total_sales ?? 0);
  const topProductName = Array.isArray(salesReport?.top_products)
    ? String((salesReport.top_products[0] as Record<string, unknown> | undefined)?.product_name ?? '')
    : '';

  const profitReport = selectedReport === 'sales-profit' && reportPayload && typeof reportPayload === 'object'
    ? reportPayload as Record<string, unknown>
    : null;
  const profitNet = Number(profitReport?.net_profit ?? 0);
  const profitRevenue = Number(profitReport?.total_revenue ?? 0);
  const profitCogs = Number(profitReport?.total_cogs ?? 0);
  const profitMargin = Number(profitReport?.profit_margin ?? 0);
  const cashReceived = Number(salesReport?.cash_received ?? 0);
  const changeAmount = Number(salesReport?.change_amount ?? 0);
  const totalPaid = Number(salesReport?.total_paid ?? 0);
  const netCashReceived = cashReceived - changeAmount;
  const taxReport = selectedReport === 'tax' && reportPayload && typeof reportPayload === 'object'
    ? reportPayload as Record<string, unknown>
    : null;
  const taxGrossSales = Number(taxReport?.gross_sales ?? 0);
  const taxDiscounts = Number(taxReport?.discounts ?? 0);
  const taxTaxableSales = Number(taxReport?.taxable_sales ?? 0);
  const taxCollected = Number(taxReport?.tax_collected ?? 0);
  const taxExemptSales = Number(taxReport?.exempt_sales ?? 0);
  const taxReturns = Number(taxReport?.returns_total ?? 0);
  const taxSalesTotal = Number(taxReport?.sales_total ?? 0);
  const taxNetSales = Number(taxReport?.net_sales_total ?? 0);
  const reportInsight = (() => {
    const money = (value: unknown) => `₪${Number(value || 0).toLocaleString()}`;
    switch (selectedReport) {
      case 'sales-profit':
        return {
          title: 'ملخص المبيعات والأرباح',
          text: salesCount > 0
            ? `تم تنفيذ ${salesCount} عملية بإيراد ${money(salesRevenue)} وربح إجمالي ${money(salesProfit)} وصافي ربح ${money(profitNet)}${topProductName ? `. المنتج الأعلى ربحًا: ${topProductName}.` : '.'}`
            : 'لا توجد مبيعات في الفترة المحددة.',
        };
      case 'tax':
        return {
          title: 'ملخص الضرائب',
          text: `إجمالي المبيعات: ${money(taxSalesTotal)} (خاضعة ${money(taxTaxableSales)} + ضريبة ${money(taxCollected)} + معفاة ${money(taxExemptSales)}). تُخصم جميع المرتجعات (${money(taxReturns)}) للوصول إلى صافي المبيعات: ${money(taxNetSales)}.`,
        };
      case 'purchases-suppliers':
        return {
          title: 'ملخص المشتريات والتجار',
          text: `إجمالي المشتريات: ${money(reportPayload?.total_cost)} • صافي المشتريات: ${money(reportPayload?.net_purchases)} • المدفوع: ${money(reportPayload?.total_paid)} • المستحق: ${money(reportPayload?.total_outstanding)}`,
        };
      case 'expenses':
        return {
          title: 'ملخص المصروفات',
          text: `إجمالي المصروفات: ${money(reportPayload?.total_expenses)} • عدد الفئات: ${Object.keys(reportPayload?.by_category || {}).length} • أعلى المصروفات مسجلة خلال الفترة المحددة.`,
        };
      case 'returns':
        return {
          title: 'ملخص مرتجعات العملاء',
          text: `قيمة المرتجعات: ${money(reportPayload?.total_refunded)} • عدد المرتجعات المكتملة: ${Number(reportPayload?.total_returns || 0).toLocaleString()} • المنتجات المرتجعة: ${Array.isArray(reportPayload?.by_product) ? reportPayload.by_product.length : 0}`,
        };
      case 'inventory':
        return {
          title: 'ملخص المخزون',
          text: `إجمالي الوحدات: ${Number(reportPayload?.total_items || 0).toLocaleString()} • قيمة المخزون بالتكلفة: ${money(reportPayload?.valuation?.total_cost ?? reportPayload?.total_value)} • منخفض المخزون: ${Number(reportPayload?.low_stock_count || 0).toLocaleString()}`,
        };
      case 'debts':
        return {
          title: 'ملخص الديون',
          text: `إجمالي الديون: ${money(reportPayload?.total_debt)} • تم السداد: ${money(reportPayload?.total_paid)} • غير مسدد: ${money(reportPayload?.outstanding)} • المتأخر: ${money(reportPayload?.overdue_debt)}`,
        };
      default:
        return { title: 'ملخص التقرير', text: 'تم تحميل بيانات التقرير للفترة المحددة.' };
    }
  })();

  return (
    <div className="report-page-shell">
      {/* Page Header */}
      <PageHeader
        title={t('reports.title')}
        description="تحليلات وتقارير شاملة عن أداء المحل مع رؤى ذكية"
        actions={
          <div className="report-page-actions">
            <ReportActions
              onExportCurrent={handleExport}
              onPrintCurrent={handlePrint}
              onExportAll={handleExport}
              onPrintAll={handlePrint}
              disabled={!reportData?.data && !reportData}
            />
            <Button variant="secondary" size={getButtonSize('reports', 'headerActions')} onClick={() => refetch()}>
              <Zap className="w-4 h-4 mr-2" />
              تحديث
            </Button>
          </div>
        }
      />
      {/* AI Analytics Insight */}
      <div className="premium-insight">
        <div className="premium-insight-icon"><Target className="h-3.5 w-3.5" /></div>
        <div className="premium-insight-copy">
              <p className="premium-insight-title">
                {reportInsight.title}
              </p>
              <p className="premium-insight-text">
                {reportInsight.text}
              </p>
              {selectedReport === 'sales-profit' && salesCount > 0 && (
                <p className="premium-insight-text">
                  هامش الربح المحقق: {salesRevenue > 0 ? ((salesProfit / salesRevenue) * 100).toFixed(1) : '0.0'}%
                </p>
              )}
            </div>
      </div>

      {/* Report Type Selection */}
      <ReportTypeSelector 
        reportTypes={reportTypes}
        selectedReport={selectedReport}
        onSelectReport={handleReportSelect}
      />

      {/* Choose the period before reading the report totals. */}
      {isPeriodReport ? (
        <>
          <DateRangeSelector
            dateRanges={dateRanges}
            selectedRange={dateRange}
            onSelectRange={handleDateRangeSelect}
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

      {/* Stats Cards */}
      <ReportStats data={reportData} loading={reportLoading} reportType={selectedReport} />

      {selectedReport === 'sales-profit' && (
        <Card className="cashflow-card">
          <CardHeader className="cashflow-card-header">
            <CardTitle className="cashflow-card-title">
              <span className="cashflow-title-icon"><ReceiptText className="h-4 w-4" /></span>
              <span>
                <span className="cashflow-title-kicker">ملخص مالي</span>
                <span>ملخص البيع والتحصيل</span>
              </span>
            </CardTitle>
            <span className="cashflow-period">
              {dateRange === 'custom' && customStartDate && customEndDate
                ? `من ${customStartDate} إلى ${customEndDate}`
                : dateRanges.find((range) => range.value === dateRange)?.label || 'الفترة المحددة'}
            </span>
          </CardHeader>
          <CardContent>
            <div className="cashflow-grid">
              <div className="cashflow-lane cashflow-lane-sales">
                <div className="cashflow-lane-heading">
                  <span className="cashflow-lane-icon"><ReceiptText className="h-4 w-4" /></span>
                  <div>
                    <p className="cashflow-lane-title">المبيعات</p>
                    <p className="cashflow-lane-caption">قيمة ما تم بيعه</p>
                  </div>
                </div>
                <div className="cashflow-amount">₪{salesRevenue.toLocaleString()}</div>
                <div className="cashflow-amount-label">إجمالي المبيعات</div>
              </div>
              <div className="cashflow-lane cashflow-lane-cash">
                <div className="cashflow-lane-heading">
                  <span className="cashflow-lane-icon"><Banknote className="h-4 w-4" /></span>
                  <div>
                    <p className="cashflow-lane-title">النقد</p>
                    <p className="cashflow-lane-caption">ما دخل الصندوق فعليًا</p>
                  </div>
                </div>
                <div className="cashflow-amount">₪{netCashReceived.toLocaleString()}</div>
                <div className="cashflow-amount-label">المتبقي بعد إعادة الباقي</div>
              </div>
            </div>
            <div className="cashflow-details">
              <div className="cashflow-detail-row">
                <span>النقد الذي استلمه الصندوق</span>
                <strong>₪{cashReceived.toLocaleString()}</strong>
              </div>
              <div className="cashflow-detail-row cashflow-detail-row-muted">
                <span>الباقي الذي أُعيد للزبون</span>
                <strong>- ₪{changeAmount.toLocaleString()}</strong>
              </div>
            </div>
            <div className="cashflow-note">
              <span className="cashflow-note-mark">i</span>
              <span>الربح يعتمد على المبيعات وتكلفة المنتجات ومصاريف المحل، وليس على النقد الموجود في الصندوق.</span>
            </div>
            <div className="cashflow-footnote">
              <span>ما دفعه الزبائن خلال الفترة</span>
              <strong>₪{totalPaid.toLocaleString()}</strong>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Charts Section */}
      <ReportCharts
        data={reportData}
        loading={reportLoading}
        reportType={selectedReport}
      />

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
          ) : selectedReport === 'tax' && reportPayload && typeof reportPayload === 'object' ? (
            <div className="tax-report-board">
              <div className="tax-report-note">
                الضريبة تخص المبيعات الخاضعة فقط. جميع المرتجعات تُخصم من إجمالي المبيعات للوصول إلى صافي المبيعات، ولا تغيّر قيمة الضريبة المسجلة على الفواتير.
              </div>
              <div className="tax-report-flow" aria-label="تسلسل حساب تقرير الضرائب">
                {[
                  ['إجمالي المبيعات قبل الضريبة', taxGrossSales, 'base'],
                  ['الخصومات', taxDiscounts, 'deduction'],
                  ['المبيعات الخاضعة', taxTaxableSales, 'taxable'],
                  ['الضريبة المسجلة', taxCollected, 'tax'],
                  ['المبيعات المعفاة', taxExemptSales, 'exempt'],
                  ['إجمالي المرتجعات', taxReturns, 'returns'],
                  ['صافي المبيعات', taxNetSales, 'total'],
                ].map(([label, amount, tone]) => (
                  <div className="tax-report-step-wrap" key={String(label)}>
                    <div className={`tax-report-step tax-report-step-${tone}`}>
                      <p className="tax-report-step-label">{label}</p>
                      <p className="tax-report-step-value">
                        ₪{Number(amount).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                      </p>
                    </div>
                  </div>
                ))}
              </div>
              <div className="tax-report-equation">
                <span>إجمالي المبيعات المسجلة: <strong>₪{taxSalesTotal.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</strong></span>
                <span className="tax-report-equation-sign">−</span>
                <span>إجمالي المرتجعات: <strong>₪{taxReturns.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</strong></span>
                <span className="tax-report-equation-sign">=</span>
                <strong className="tax-report-equation-result">صافي المبيعات ₪{taxNetSales.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</strong>
              </div>
            </div>
          ) : selectedReport === 'sales-profit' && reportPayload && typeof reportPayload === 'object' ? (
            <div className="horizontal-scroll">
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                    {['التاريخ', 'المبيعات قبل الضريبة', 'صافي الربح', 'الهامش'].map((heading) => (
                      <th key={heading} style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>{heading}</th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {Array.isArray((reportPayload as Record<string, unknown>).by_day) && (reportPayload as Record<string, unknown[]>).by_day.length > 0 ? (
                    ((reportPayload as Record<string, unknown[]>).by_day as Record<string, unknown>[]).map((day, index) => {
                      const revenue = Number(day.revenue || 0);
                      const netProfit = Number(day.net_profit || 0);
                      return (
                        <tr key={`${String(day.date || index)}`} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                          <td style={{ padding: '12px', color: 'var(--text-primary)', fontSize: '13px', fontWeight: '600' }}>{formatReportDate(day.date)}</td>
                          <td style={{ padding: '12px', color: 'var(--color-primary)', fontSize: '13px', fontWeight: '600' }}>₪{revenue.toLocaleString()}</td>
                          <td style={{ padding: '12px', color: netProfit >= 0 ? 'var(--color-success)' : 'var(--color-error)', fontSize: '13px', fontWeight: '600' }}>₪{netProfit.toLocaleString()}</td>
                          <td style={{ padding: '12px', color: 'var(--text-secondary)', fontSize: '13px' }}>{revenue > 0 ? `${((netProfit / revenue) * 100).toFixed(1)}%` : '0.0%'}</td>
                        </tr>
                      );
                    })
                  ) : (
                    <tr><td colSpan={4} style={{ padding: '24px', textAlign: 'center', color: 'var(--text-secondary)' }}>لا توجد حركة مالية في الفترة المحددة</td></tr>
                  )}
                </tbody>
              </table>
            </div>
          ) : selectedReport === 'sales' && reportPayload && typeof reportPayload === 'object' ? (
            <div>
              <p className="text-sm text-text-secondary" style={{ marginBottom: '12px' }}>
                المنتجات الأعلى ربحًا خلال الفترة المحددة. الإيراد والربح معروضان قبل الضريبة.
              </p>
              <div className="horizontal-scroll">
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                    {['المنتج', 'الوحدات', 'الإيراد قبل الضريبة', 'الربح قبل الضريبة', 'نسبة الربح من سعر البيع'].map((heading) => (
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
          ) : selectedReport === 'inventory' && reportPayload && typeof reportPayload === 'object' ? (
            <div className="horizontal-scroll">
              {(() => {
                const inventoryItems = Array.isArray((reportPayload as Record<string, unknown>).items)
                  ? (reportPayload as Record<string, unknown[]>).items as Record<string, unknown>[]
                  : [];
                return (
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>المنتج</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>المخزون الحالي</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>الحد الأدنى</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>الحالة</th>
                  </tr>
                </thead>
                <tbody>
                  {inventoryItems.length > 0
                    ? inventoryItems.map((item, index) => (
                        <tr key={String(item.product_id || index)} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                          <td style={{ padding: '12px', color: 'var(--text-primary)', fontSize: '13px', fontWeight: '600' }}>{String(item.product_name || 'منتج غير معروف')}</td>
                          <td style={{ padding: '12px', color: 'var(--text-primary)', fontSize: '13px' }}>{Number(item.current_stock ?? 0)}</td>
                          <td style={{ padding: '12px', color: 'var(--text-secondary)', fontSize: '13px' }}>{Number(item.min_stock ?? 0)}</td>
                          <td style={{ padding: '12px' }}>
                            <Badge variant={Number(item.current_stock ?? 0) <= 0
                              ? 'danger'
                              : Number(item.min_stock ?? 0) > 0 && Number(item.current_stock ?? 0) <= Number(item.min_stock ?? 0)
                                ? 'warning'
                                : 'success'}>
                              {Number(item.current_stock ?? 0) <= 0
                                ? 'نفد'
                                : Number(item.min_stock ?? 0) > 0 && Number(item.current_stock ?? 0) <= Number(item.min_stock ?? 0)
                                  ? 'منخفض'
                                  : 'متوفر'}
                            </Badge>
                          </td>
                        </tr>
                      ))
                    : (
                      <tr><td colSpan={4} style={{ padding: '24px', textAlign: 'center', color: 'var(--text-secondary)' }}>لا توجد منتجات في المخزون</td></tr>
                    )}
                </tbody>
              </table>
                );
              })()}
            </div>
          ) : selectedReport === 'used-items' ? (
            // تقرير القطع المستعملة - عرض خاص
            <div className="horizontal-scroll">
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>المنتج</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>اشتريت من</th>
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
                            {item.seller_name || 'غير محدد'}
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
          ) : selectedReport === 'debts' && reportPayload && typeof reportPayload === 'object' ? (
            <div className="horizontal-scroll">
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>الزبون</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>إجمالي الدين</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>المدفوع</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>المتبقي</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>متأخر</th>
                  </tr>
                </thead>
                <tbody>
                  {(() => {
                    const activeCustomers = Array.isArray((reportPayload as Record<string, unknown>).by_customer)
                      ? ((reportPayload as Record<string, unknown[]>).by_customer as Record<string, unknown>[]).filter(customer => Number(customer.outstanding ?? 0) > 0)
                      : [];
                    return activeCustomers.length > 0 ? (
                    activeCustomers.map((customer, index) => (
                      <tr key={String(customer.customer_id ?? index)} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                        <td style={{ padding: '12px', color: 'var(--text-primary)', fontSize: '13px', fontWeight: '600' }}>{String(customer.customer_name ?? 'زبون غير معروف')}</td>
                        <td style={{ padding: '12px', color: 'var(--color-primary)', fontSize: '13px', fontWeight: '600' }}>₪{Number(customer.total_debt ?? 0).toLocaleString()}</td>
                        <td style={{ padding: '12px', color: 'var(--text-secondary)', fontSize: '13px' }}>₪{Number(customer.paid_amount ?? 0).toLocaleString()}</td>
                        <td style={{ padding: '12px', color: 'var(--text-primary)', fontSize: '13px' }}>₪{Number(customer.outstanding ?? 0).toLocaleString()}</td>
                        <td style={{ padding: '12px', color: Number(customer.overdue_amount ?? 0) > 0 ? 'var(--color-error)' : 'var(--text-secondary)', fontSize: '13px' }}>₪{Number(customer.overdue_amount ?? 0).toLocaleString()}</td>
                      </tr>
                    ))
                    ) : (
                      <tr><td colSpan={5} style={{ padding: '24px', textAlign: 'center', color: 'var(--text-secondary)' }}>لا توجد ديون نشطة في الفترة الحالية</td></tr>
                    );
                  })()}
                </tbody>
              </table>
              {Array.isArray((reportPayload as Record<string, unknown>).payment_history) && ((reportPayload as Record<string, unknown[]>).payment_history as Record<string, unknown>[]).length > 0 && (
                <div style={{ marginTop: '20px' }}>
                  <h4 style={{ marginBottom: '10px', color: 'var(--text-primary)', fontSize: '14px', fontWeight: '700' }}>سجل التحصيلات</h4>
                  <div className="horizontal-scroll">
                    <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                      <thead>
                        <tr style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                          {['العميل', 'المبلغ', 'تاريخ الدفع', 'طريقة الدفع', 'المرجع', 'الملاحظات'].map((heading) => (
                            <th key={heading} style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>{heading}</th>
                          ))}
                        </tr>
                      </thead>
                      <tbody>
                        {((reportPayload as Record<string, unknown[]>).payment_history as Record<string, unknown>[]).map((payment, index) => {
                          const methodLabels: Record<string, string> = { cash: 'نقدي', card: 'بطاقة', transfer: 'تحويل', bank_transfer: 'تحويل بنكي', cheque: 'شيك', check: 'شيك' };
                          const method = String(payment.payment_method || '').toLowerCase();
                          return (
                            <tr key={`${String(payment.customer_id || index)}-${String(payment.date || index)}`} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                              <td style={{ padding: '12px', color: 'var(--text-primary)', fontSize: '13px', fontWeight: '600' }}>{String(payment.customer_name || 'عميل غير معروف')}</td>
                              <td style={{ padding: '12px', color: 'var(--color-success)', fontSize: '13px', fontWeight: '600' }}>₪{Number(payment.amount || 0).toLocaleString()}</td>
                              <td style={{ padding: '12px', color: 'var(--text-secondary)', fontSize: '13px' }}>{formatReportDate(payment.date)}</td>
                              <td style={{ padding: '12px', color: 'var(--text-secondary)', fontSize: '13px' }}>{methodLabels[method] || String(payment.payment_method || 'غير محدد')}</td>
                              <td style={{ padding: '12px', color: 'var(--text-secondary)', fontSize: '13px' }}>{String(payment.reference_number || '-')}</td>
                              <td style={{ padding: '12px', color: 'var(--text-secondary)', fontSize: '13px' }}>{String(payment.notes || '-')}</td>
                            </tr>
                          );
                        })}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}
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
          ) : selectedReport === 'purchases-suppliers' && reportPayload && typeof reportPayload === 'object' ? (
            <div className="horizontal-scroll">
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>التاجر</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>إجمالي المشتريات</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>المدفوع</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>المستحق</th>
                  </tr>
                </thead>
                <tbody>
                  {Array.isArray((reportPayload as Record<string, unknown>).by_supplier) && ((reportPayload as Record<string, unknown[]>).by_supplier as Record<string, unknown>[]).length > 0 ? (
                    ((reportPayload as Record<string, unknown[]>).by_supplier as Record<string, unknown>[]).map((supplier, index) => (
                      <tr key={String(supplier.supplier_id ?? index)} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                        <td style={{ padding: '12px', color: 'var(--text-primary)', fontSize: '13px', fontWeight: '600' }}>{String(supplier.supplier_name ?? 'تاجر غير معروف')}</td>
                        <td style={{ padding: '12px', color: 'var(--color-primary)', fontSize: '13px', fontWeight: '600' }}>₪{Number(supplier.total_purchases ?? 0).toLocaleString()}</td>
                        <td style={{ padding: '12px', color: 'var(--color-success)', fontSize: '13px' }}>₪{Number(supplier.total_paid ?? 0).toLocaleString()}</td>
                        <td style={{ padding: '12px', color: 'var(--text-primary)', fontSize: '13px' }}>₪{Number(supplier.outstanding ?? 0).toLocaleString()}</td>
                      </tr>
                    ))
                  ) : (
                    <tr><td colSpan={4} style={{ padding: '24px', textAlign: 'center', color: 'var(--text-secondary)' }}>لا توجد مشتريات أو أرصدة تجار</td></tr>
                  )}
                </tbody>
              </table>
            </div>
          ) : selectedReport === 'purchases' && reportPayload && typeof reportPayload === 'object' ? (
            <div className="space-y-5">
              <p className="text-sm text-text-secondary">إجمالي المشتريات قبل خصم المرتجعات. صافي المشتريات = إجمالي المشتريات - مرتجعات الموردين.</p>
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
                <div className="rounded-lg border border-border bg-surface-elevated p-4">
                  <p className="text-sm text-text-secondary">المشتريات قبل الضريبة</p>
                  <p className="mt-2 text-xl font-semibold text-text-primary">₪{Number((reportPayload as Record<string, unknown>).subtotal ?? 0).toLocaleString()}</p>
                </div>
                <div className="rounded-lg border border-border bg-surface-elevated p-4">
                  <p className="text-sm text-text-secondary">ضريبة المشتريات</p>
                  <p className="mt-2 text-xl font-semibold text-text-primary">₪{Number((reportPayload as Record<string, unknown>).tax_amount ?? 0).toLocaleString()}</p>
                </div>
                <div className="rounded-lg border border-border bg-surface-elevated p-4">
                  <p className="text-sm text-text-secondary">فواتير بلا ضريبة</p>
                  <p className="mt-2 text-xl font-semibold text-text-primary">₪{untaxedPurchaseCost.toLocaleString()}</p>
                  <p className="mt-1 text-xs text-text-secondary">{untaxedPurchaseCount.toLocaleString()} فواتير معفاة</p>
                </div>
                <div className="rounded-lg border border-border bg-surface-elevated p-4">
                  <p className="text-sm text-text-secondary">فواتير خاضعة للضريبة</p>
                  <p className="mt-2 text-xl font-semibold text-text-primary">₪{taxedPurchaseCost.toLocaleString()}</p>
                  <p className="mt-1 text-xs text-text-secondary">{taxedPurchaseCount.toLocaleString()} فواتير، شامل الضريبة</p>
                </div>
                <div className="rounded-lg border border-border bg-surface-elevated p-4">
                  <p className="text-sm text-text-secondary">مرتجعات الموردين</p>
                  <p className="mt-2 text-xl font-semibold text-text-primary">₪{supplierReturnCredits.toLocaleString()}</p>
                  <p className="mt-1 text-xs text-text-secondary">مرتجعات مكتملة خلال الفترة</p>
                </div>
                <div className="rounded-lg border border-border bg-surface-elevated p-4">
                  <p className="text-sm text-text-secondary">صافي المشتريات</p>
                  <p className="mt-2 text-xl font-semibold text-text-primary">₪{netPurchases.toLocaleString()}</p>
                  <p className="mt-1 text-xs text-text-secondary">الإجمالي - مرتجعات الموردين</p>
                </div>
              </div>
              <div className="horizontal-scroll">
                <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                  <thead>
                    <tr style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                      <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>التاجر</th>
                      <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>التكلفة</th>
                      <th style={{ padding: '12px', textAlign: 'right', color: 'var(--text-secondary)', fontSize: '12px', fontWeight: '600' }}>الأصناف</th>
                    </tr>
                  </thead>
                  <tbody>
                    {Array.isArray((reportPayload as Record<string, unknown>).by_supplier) && ((reportPayload as Record<string, unknown[]>).by_supplier as Record<string, unknown>[]).length > 0 ? (
                      ((reportPayload as Record<string, unknown[]>).by_supplier as Record<string, unknown>[]).map((supplier, index) => (
                        <tr key={String(supplier.supplier_id ?? index)} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                          <td style={{ padding: '12px', color: 'var(--text-primary)', fontSize: '13px', fontWeight: '600' }}>{String(supplier.supplier_name ?? 'تاجر غير معروف')}</td>
                          <td style={{ padding: '12px', color: 'var(--color-primary)', fontSize: '13px', fontWeight: '600' }}>₪{Number(supplier.total_cost ?? 0).toLocaleString()}</td>
                          <td style={{ padding: '12px', color: 'var(--text-secondary)', fontSize: '13px' }}>{Number(supplier.item_count ?? 0).toLocaleString()}</td>
                        </tr>
                      ))
                    ) : (
                      <tr><td colSpan={3} style={{ padding: '24px', textAlign: 'center', color: 'var(--text-secondary)' }}>لا توجد مشتريات في الفترة المحددة</td></tr>
                    )}
                  </tbody>
                </table>
              </div>
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
                        {formatReportDate(item.date || item.month)}
                      </td>
                      <td style={{ padding: '12px', color: 'var(--color-primary)', fontSize: '13px', fontWeight: '600' }}>
                        ₪{formatReportAmount(getReportDisplayValue(item, selectedReport))}
                      </td>
                      <td style={{ padding: '12px', color: 'var(--text-secondary)', fontSize: '13px' }}>
                        {getReportRowDescription(item)}
                      </td>
                      <td style={{ padding: '12px' }}>
                        <Badge variant={String(item.status || 'completed').toLowerCase() === 'completed' || String(item.status || '').toLowerCase() === 'approved' ? 'success' : 'warning'}>
                          {getReportStatusLabel(item.status)}
                        </Badge>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : hasSummaryMetrics ? (
            <div style={{ textAlign: 'center', padding: '40px 20px' }}>
              <BarChart3 className="w-14 h-14 text-slate-500 mx-auto mb-4" />
              <p style={{ fontSize: '16px', fontWeight: '600', color: 'var(--text-primary)', marginBottom: '8px' }}>
                ملخص التقرير
              </p>
              <p style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
                تم تحميل القيم الإجمالية بنجاح في بطاقات التقرير أعلاه.
              </p>
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
