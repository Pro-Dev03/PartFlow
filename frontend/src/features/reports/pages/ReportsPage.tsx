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

export function ReportsPage() {
  const { t } = useTranslation();
  const [selectedReport, setSelectedReport] = useState('sales');
  const [dateRange, setDateRange] = useState('thisMonth');

  // Custom hook
  const { reportData, reportLoading, refetch } = useReports(selectedReport, dateRange);

  const handleExport = () => {
    if (!reportData?.data) return;
    
    const dataToExport = reportData.data.map((item: any) => ({
      'التاريخ': item.date || new Date().toLocaleDateString('ar-SA'),
      'القيمة': item.value || item.amount || 0,
      'الوصف': item.description || item.name || '-',
      'الحالة': item.status || 'مكتمل',
    }));
    
    exportToCSV(dataToExport, `${selectedReport}-report-${new Date().toISOString().split('T')[0]}`);
  };

  const handlePrint = () => {
    if (!reportData?.data) return;
    
    const dataToPrint = reportData.data.map((item: any) => ({
      'التاريخ': item.date || new Date().toLocaleDateString('ar-SA'),
      'القيمة': item.value || item.amount || 0,
      'الوصف': item.description || item.name || '-',
      'الحالة': item.status || 'مكتمل',
    }));
    
    printTable(dataToPrint, ['التاريخ', 'القيمة', 'الوصف', 'الحالة'], selectedReportType?.label || 'تقرير');
  };

  const reportTypes: ReportType[] = [
    { id: 'sales', label: t('reports.salesReport'), icon: BarChart3 },
    { id: 'net-sales', label: 'المبيعات الصافية', icon: Target },
    { id: 'profit', label: t('reports.profitReport'), icon: Target },
    { id: 'inventory', label: t('reports.inventoryReport'), icon: BarChart3 },
    { id: 'used-items', label: 'تقرير القطع المستعملة', icon: Zap },
    { id: 'debts', label: t('reports.debtsReport'), icon: Target },
    { id: 'products', label: t('reports.productsReport'), icon: BarChart3 },
    { id: 'suppliers', label: t('reports.suppliersReport'), icon: BarChart3 },
    { id: 'expenses', label: t('reports.expensesReport'), icon: BarChart3 },
    { id: 'returns', label: 'تقرير المرتجعات المحسّن', icon: RotateCcw },
    { id: 'returns-analysis', label: 'تحليل المرتجعات', icon: BarChart3 },
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

  return (
    <div>
      {/* Page Header */}
      <PageHeader
        eyebrow="Analytics Hub"
        title={t('reports.title')}
        description="تحليلات وتقارير شاملة عن أداء المحل مع رؤى ذكية"
        actions={
          <div style={{ display: 'flex', gap: '10px' }}>
            <Button variant="secondary" size={getButtonSize('reports', 'headerActions')} onClick={handleExport} disabled={!reportData?.data}>
              <Download className="w-4 h-4 mr-2" />
              {t('reports.export')}
            </Button>
            <Button variant="secondary" size={getButtonSize('reports', 'headerActions')} onClick={handlePrint} disabled={!reportData?.data}>
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
              <p style={{ fontSize: '13px', fontWeight: '600', color: 'var(--text-primary)' }}>فرصة نمو محتملة</p>
              <p style={{ fontSize: '11px', color: 'var(--text-secondary)', marginTop: '4px' }}>
                المبيعات من كروت الشاشة زادت 23% هذا الشهر. يُنصح بزيادة المخزون من هذه الفئة.
              </p>
              <Button variant="secondary" size={getButtonSize('reports', 'recommendation')}>
                عرض التوصية ←
              </Button>
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

      {/* Date Range Selection */}
      <DateRangeSelector 
        dateRanges={dateRanges}
        selectedRange={dateRange}
        onSelectRange={setDateRange}
      />

      {/* Stats Cards */}
      <ReportStats data={reportData} loading={reportLoading} />

      {/* Charts Section */}
      <ReportCharts data={reportData} loading={reportLoading} />

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
                  {reportData?.data && reportData.data.length > 0 ? (
                    reportData.data.map((item: any, index: number) => {
                      const profit = ((item.selling_price || 0) - (item.purchase_cost || 0)) / 100;
                      return (
                        <tr key={index} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                          <td style={{ padding: '12px', color: 'var(--text-primary)', fontSize: '13px' }}>
                            {item.product_name || item.name || '-'}
                          </td>
                          <td style={{ padding: '12px', color: 'var(--text-secondary)', fontSize: '13px' }}>
                            ₪{((item.purchase_cost || 0) / 100).toFixed(2)}
                          </td>
                          <td style={{ padding: '12px', color: 'var(--color-success)', fontSize: '13px', fontWeight: '600' }}>
                            ₪{((item.selling_price || 0) / 100).toFixed(2)}
                          </td>
                          <td style={{ padding: '12px', color: profit > 0 ? 'var(--color-success)' : 'var(--color-error)', fontSize: '13px', fontWeight: '600' }}>
                            ₪{(profit / 100).toFixed(2)}
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
          ) : reportData?.data && reportData.data.length > 0 ? (
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
                  {reportData.data.map((item: any, index: number) => (
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
