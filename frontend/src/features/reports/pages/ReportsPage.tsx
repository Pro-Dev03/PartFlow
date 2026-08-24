import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { reportsApi, inventoryApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { PageHeader } from '../../../components/ui/page-header';
import { StatCard } from '../../../components/ui/stat-card';
import { Select } from '../../../components/ui/select';
import { Badge } from '../../../components/ui/badge';
import { SimpleBarChart, SimpleLineChart, SimplePieChart } from '../../../components/ui/charts';
import { exportToCSV, printTable } from '../../../lib/export-utils';
import { getButtonSize } from '../../../config/button-sizes';
import {
  BarChart3,
  TrendingUp,
  DollarSign,
  Package,
  Users,
  Truck,
  CreditCard,
  Download,
  Printer,
  Calendar,
  AlertTriangle,
  Clock,
  CheckCircle,
  ArrowUpRight,
  ArrowDownRight,
  Filter,
  FileText,
  PieChart,
  LineChart,
  Sparkles,
  TrendingDown,
  Database,
  Zap,
  Target,
  RefreshCw
} from 'lucide-react';

export function ReportsPage() {
  const { t } = useTranslation();
  const [selectedReport, setSelectedReport] = useState('sales');
  const [dateRange, setDateRange] = useState('thisMonth');

  // Fetch report data based on selected report type
  const { data: reportData, isLoading: reportLoading, refetch } = useQuery({
    queryKey: ['reports', selectedReport, dateRange],
    queryFn: () => {
      switch (selectedReport) {
        case 'sales':
          return reportsApi.sales();
        case 'profit':
          return reportsApi.profit();
        case 'inventory':
          return reportsApi.inventory();
        case 'debts':
          return reportsApi.debts();
        case 'products':
          return reportsApi.products();
        case 'suppliers':
          return reportsApi.suppliers();
        case 'expenses':
          return reportsApi.expenses();
        case 'returns':
          return reportsApi.returns();
        case 'used-items':
          return inventoryApi.list({ condition: 'USED' });
        default:
          return reportsApi.sales();
      }
    },
  });

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

  const reportTypes = [
    { id: 'sales', label: t('reports.salesReport'), icon: BarChart3 },
    { id: 'profit', label: t('reports.profitReport'), icon: TrendingUp },
    { id: 'inventory', label: t('reports.inventoryReport'), icon: Package },
    { id: 'used-items', label: 'تقرير القطع المستعملة', icon: RefreshCw },
    { id: 'debts', label: t('reports.debtsReport'), icon: DollarSign },
    { id: 'products', label: t('reports.productsReport'), icon: Package },
    { id: 'suppliers', label: t('reports.suppliersReport'), icon: Truck },
    { id: 'expenses', label: t('reports.expensesReport'), icon: CreditCard },
    { id: 'returns', label: t('reports.returnReport'), icon: Package },
  ];

  const dateRanges = [
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
              <p style={{ fontSize: '13px', fontWeight: '600', color: '#f1f7ff' }}>فرصة نمو محتملة</p>
              <p style={{ fontSize: '11px', color: '#8290a7', marginTop: '4px' }}>
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
      <Card>
        <CardHeader>
          <CardTitle style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <FileText className="w-5 h-5 text-cyan-400" />
            نوع التقرير
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))', gap: '14px' }}
               className="grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
            {reportTypes.map((report) => {
              const Icon = report.icon;
              return (
                <button
                  key={report.id}
                  onClick={() => setSelectedReport(report.id)}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '10px',
                    padding: '14px',
                    borderRadius: '10px',
                    border: selectedReport === report.id 
                      ? '1px solid rgba(34, 211, 238, 0.35)' 
                      : '1px solid rgba(148, 163, 184, 0.13)',
                    background: selectedReport === report.id
                      ? 'linear-gradient(135deg, rgba(34, 211, 238, 0.17), rgba(59, 130, 246, 0.12))'
                      : 'rgba(17, 24, 39, 0.75)',
                    cursor: 'pointer',
                    transition: '180ms ease',
                    color: selectedReport === report.id ? '#eaffff' : '#94a3b8'
                  }}
                  onMouseEnter={(e) => {
                    if (selectedReport !== report.id) {
                      e.currentTarget.style.borderColor = 'rgba(34, 211, 238, 0.3)';
                      e.currentTarget.style.transform = 'translateY(-2px)';
                    }
                  }}
                  onMouseLeave={(e) => {
                    if (selectedReport !== report.id) {
                      e.currentTarget.style.borderColor = 'rgba(148, 163, 184, 0.13)';
                      e.currentTarget.style.transform = 'translateY(0)';
                    }
                  }}
                >
                  <div style={{
                    width: '32px',
                    height: '32px',
                    borderRadius: '8px',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    background: selectedReport === report.id
                      ? 'rgba(34, 211, 238, 0.2)'
                      : 'rgba(34, 211, 238, 0.1)'
                  }}>
                    <Icon className="w-4.5 h-4.5 text-cyan-400" />
                  </div>
                  <span style={{ fontSize: '13px', fontWeight: '500' }}>{report.label}</span>
                </button>
              );
            })}
          </div>
        </CardContent>
      </Card>

      {/* Date Range Selection */}
      <Card>
        <CardHeader>
          <CardTitle style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <Calendar className="w-5 h-5 text-cyan-400" />
            نطاق التاريخ
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div style={{ display: 'flex', gap: '10px' }}>
            {dateRanges.map((range) => (
              <Button
                key={range.value}
                variant={dateRange === range.value ? 'primary' : 'secondary'}
                onClick={() => setDateRange(range.value)}
              >
                {range.label}
              </Button>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Stats Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '14px' }}
           className="grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard 
          title="إجمالي المبيعات" 
          value="₪45,230" 
          icon={DollarSign}
          subtitle="للفترة المحددة"
          variant="featured"
          trend="+15.3%"
          trendUp={true}
        />
        <StatCard 
          title="إجمالي الأرباح" 
          value="₪12,450" 
          icon={TrendingUp}
          subtitle="هامش الربح"
          variant="success"
          trend="+8.7%"
          trendUp={true}
        />
        <StatCard 
          title="عدد المعاملات" 
          value="234" 
          icon={Database}
          subtitle="عمليات بيع"
          variant="default"
          trend="+12.1%"
          trendUp={true}
        />
        <StatCard 
          title="متوسط الطلب" 
          value="₪193" 
          icon={Target}
          subtitle="لكل معاملة"
          variant="info"
          trend="+5.4%"
          trendUp={true}
        />
      </div>

      {/* Charts Section */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: '14px' }}
           className="grid-cols-1 lg:grid-cols-2">
        <SimpleBarChart
          title="توزيع المبيعات حسب الفئة"
          data={[
            { label: 'إلكترونيات', value: 15230 },
            { label: 'اكسسوارات', value: 8900 },
            { label: 'شاشات', value: 12500 },
            { label: 'بطاريات', value: 8600 },
          ]}
          color="#22d3ee"
        />
        <SimpleLineChart
          title="اتجاه المبيعات الشهرية"
          data={[
            { label: 'يناير', value: 32000 },
            { label: 'فبراير', value: 38000 },
            { label: 'مارس', value: 35000 },
            { label: 'أبريل', value: 42000 },
            { label: 'مايو', value: 45230 },
          ]}
          color="#34d399"
        />
        <SimplePieChart
          title="توزيع مصادر الدخل"
          data={[
            { label: 'نقداً', value: 23450, color: '#22d3ee' },
            { label: 'بطاقات', value: 12800, color: '#34d399' },
            { label: 'ديون', value: 8980, color: '#fbbf24' },
          ]}
        />
      </div>

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
              <div style={{ animation: 'spin 1s linear infinite', borderRadius: '50%', height: '32px', width: '32px', borderBottom: '2px solid #22d3ee' }} />
            </div>
          ) : selectedReport === 'used-items' ? (
            // تقرير القطع المستعملة - عرض خاص
            <div className="horizontal-scroll">
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid rgba(148, 163, 184, 0.13)' }}>
                    <th style={{ padding: '12px', textAlign: 'right', color: '#8290a7', fontSize: '12px', fontWeight: '600' }}>المنتج</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: '#8290a7', fontSize: '12px', fontWeight: '600' }}>اشتريت بـ</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: '#8290a7', fontSize: '12px', fontWeight: '600' }}>بعت بـ</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: '#8290a7', fontSize: '12px', fontWeight: '600' }}>الربح</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: '#8290a7', fontSize: '12px', fontWeight: '600' }}>الكمية المباعة</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: '#8290a7', fontSize: '12px', fontWeight: '600' }}>الحالة</th>
                  </tr>
                </thead>
                <tbody>
                  {reportData?.data && reportData.data.length > 0 ? (
                    reportData.data.map((item: any, index: number) => {
                      const profit = (item.selling_price || 0) - (item.purchase_cost || 0);
                      return (
                        <tr key={index} style={{ borderBottom: '1px solid rgba(148, 163, 184, 0.08)' }}>
                          <td style={{ padding: '12px', color: '#f1f7ff', fontSize: '13px' }}>
                            {item.product_name || item.name || '-'}
                          </td>
                          <td style={{ padding: '12px', color: '#94a3b8', fontSize: '13px' }}>
                            ₪{((item.purchase_cost || 0) / 100).toFixed(2)}
                          </td>
                          <td style={{ padding: '12px', color: '#34d399', fontSize: '13px', fontWeight: '600' }}>
                            ₪{((item.selling_price || 0) / 100).toFixed(2)}
                          </td>
                          <td style={{ padding: '12px', color: profit > 0 ? '#34d399' : '#fb7185', fontSize: '13px', fontWeight: '600' }}>
                            ₪{(profit / 100).toFixed(2)}
                          </td>
                          <td style={{ padding: '12px', color: '#f1f7ff', fontSize: '13px' }}>
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
                      <td colSpan={7} style={{ padding: '24px', textAlign: 'center', color: '#94a3b8' }}>
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
                  <tr style={{ borderBottom: '1px solid rgba(148, 163, 184, 0.13)' }}>
                    <th style={{ padding: '12px', textAlign: 'right', color: '#8290a7', fontSize: '12px', fontWeight: '600' }}>التاريخ</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: '#8290a7', fontSize: '12px', fontWeight: '600' }}>القيمة</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: '#8290a7', fontSize: '12px', fontWeight: '600' }}>الوصف</th>
                    <th style={{ padding: '12px', textAlign: 'right', color: '#8290a7', fontSize: '12px', fontWeight: '600' }}>الحالة</th>
                  </tr>
                </thead>
                <tbody>
                  {reportData.data.map((item: any, index: number) => (
                    <tr key={index} style={{ borderBottom: '1px solid rgba(148, 163, 184, 0.08)' }}>
                      <td style={{ padding: '12px', color: '#f1f7ff', fontSize: '13px' }}>
                        {item.date || new Date().toLocaleDateString('ar-SA')}
                      </td>
                      <td style={{ padding: '12px', color: '#22d3ee', fontSize: '13px', fontWeight: '600' }}>
                        ₪{(item.value || item.amount || 0).toLocaleString()}
                      </td>
                      <td style={{ padding: '12px', color: '#94a3b8', fontSize: '13px' }}>
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
              <p style={{ fontSize: '16px', fontWeight: '600', color: '#f1f7ff', marginBottom: '8px' }}>
                لا توجد بيانات
              </p>
              <p style={{ fontSize: '13px', color: '#8290a7' }}>
                لم يتم العثور على بيانات لهذا التقرير في الفترة المحددة
              </p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
