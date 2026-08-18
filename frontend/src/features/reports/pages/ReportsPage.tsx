import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { reportsApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Select } from '../../../components/ui/select';
import { Badge } from '../../../components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
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
  LineChart
} from 'lucide-react';

export function ReportsPage() {
  const { t } = useTranslation();
  const [selectedReport, setSelectedReport] = useState('sales');
  const [dateRange, setDateRange] = useState('thisMonth');

  const reportTypes = [
    { id: 'sales', label: t('reports.salesReport'), icon: BarChart3 },
    { id: 'profit', label: t('reports.profitReport'), icon: TrendingUp },
    { id: 'inventory', label: t('reports.inventoryReport'), icon: Package },
    { id: 'debts', label: t('reports.debtsReport'), icon: DollarSign },
    { id: 'products', label: t('reports.productsReport'), icon: Package },
    { id: 'suppliers', label: t('reports.suppliersReport'), icon: Truck },
    { id: 'expenses', label: t('reports.expensesReport'), icon: CreditCard },
    { id: 'returns', label: t('reports.returnReport'), icon: Package },
    { id: 'warranty', label: t('reports.warrantyReport'), icon: Clock },
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
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
          {t('reports.title')}
        </h1>
        <p className="text-gray-500 dark:text-gray-400 mt-1">
          تحليلات وتقارير شاملة عن أداء المحل
        </p>
      </div>

      {/* Report Type Selection */}
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-3">
        {reportTypes.map((report) => {
          const Icon = report.icon;
          return (
            <Button
              key={report.id}
              variant={selectedReport === report.id ? 'primary' : 'outline'}
              onClick={() => setSelectedReport(report.id)}
              className="flex flex-col items-center gap-2 h-auto py-4"
            >
              <Icon className="w-5 h-5" />
              <span className="text-xs">{report.label}</span>
            </Button>
          );
        })}
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex flex-col md:flex-row gap-4 items-end">
            <div className="flex-1">
              <label className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1 block">
                {t('reports.dateRange')}
              </label>
              <Select
                value={dateRange}
                onChange={(e) => setDateRange(e.target.value)}
                options={dateRanges}
              />
            </div>
            {dateRange === 'custom' && (
              <>
                <div className="flex-1">
                  <label className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1 block">
                    {t('reports.startDate')}
                  </label>
                  <input
                    type="date"
                    className="w-full h-10 rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-100"
                  />
                </div>
                <div className="flex-1">
                  <label className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1 block">
                    {t('reports.endDate')}
                  </label>
                  <input
                    type="date"
                    className="w-full h-10 rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-100"
                  />
                </div>
              </>
            )}
            <div className="flex gap-2">
              <Button variant="outline" className="gap-2">
                <Download className="w-4 h-4" />
                {t('reports.export')}
              </Button>
              <Button variant="outline" className="gap-2">
                <Printer className="w-4 h-4" />
                {t('reports.print')}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Report Content */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <ReportIcon className="w-5 h-5" />
            {selectedReportType?.label}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {selectedReport === 'sales' && <SalesReportContent />}
          {selectedReport === 'profit' && <ProfitReportContent />}
          {selectedReport === 'inventory' && <InventoryReportContent />}
          {selectedReport === 'debts' && <DebtsReportContent />}
          {selectedReport === 'products' && <ProductsReportContent />}
          {selectedReport === 'suppliers' && <SuppliersReportContent />}
          {selectedReport === 'expenses' && <ExpensesReportContent />}
          {selectedReport === 'returns' && <ReturnsReportContent />}
          {selectedReport === 'warranty' && <WarrantyReportContent />}
        </CardContent>
      </Card>
    </div>
  );
}

function SalesReportContent() {
  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <StatCard title="الإيرادات" value="₪100,000" icon={DollarSign} />
        <StatCard title="عدد الطلبات" value="150" icon={BarChart3} />
        <StatCard title="القطع المباعة" value="320" icon={Package} />
        <StatCard title="متوسط البيع" value="₪667" icon={TrendingUp} />
      </div>
      
      <div className="h-64 bg-gray-50 dark:bg-gray-800 rounded-lg flex items-center justify-center">
        <p className="text-gray-500 dark:text-gray-400">رسم بياني للمبيعات</p>
      </div>
    </div>
  );
}

function ProfitReportContent() {
  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <StatCard title="الإيرادات" value="₪100,000" icon={DollarSign} />
        <StatCard title="التكلفة" value="₪72,000" icon={Package} />
        <StatCard title="الربح الإجمالي" value="₪28,000" icon={TrendingUp} />
        <StatCard title="الربح الصافي" value="₪20,000" icon={TrendingUp} />
      </div>
      
      <div className="h-64 bg-gray-50 dark:bg-gray-800 rounded-lg flex items-center justify-center">
        <p className="text-gray-500 dark:text-gray-400">رسم بياني للأرباح</p>
      </div>
    </div>
  );
}

function InventoryReportContent() {
  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <StatCard title="قيمة المخزون" value="₪185,400" icon={Package} />
        <StatCard title="منخفض المخزون" value="12" icon={AlertTriangle} />
        <StatCard title="المخزون الراكد" value="8" icon={Clock} />
        <StatCard title="سريع الحركة" value="25" icon={TrendingUp} />
      </div>
      
      <div className="h-64 bg-gray-50 dark:bg-gray-800 rounded-lg flex items-center justify-center">
        <p className="text-gray-500 dark:text-gray-400">رسم بياني للمخزون</p>
      </div>
    </div>
  );
}

function DebtsReportContent() {
  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <StatCard title="إجمالي الديون" value="₪24,800" icon={DollarSign} />
        <StatCard title="متأخرة" value="₪8,200" icon={AlertTriangle} />
        <StatCard title="مستحقة قريباً" value="₪16,600" icon={Calendar} />
        <StatCard title="مدفوعة" value="₪12,400" icon={CheckCircle} />
      </div>
      
      <div className="h-64 bg-gray-50 dark:bg-gray-800 rounded-lg flex items-center justify-center">
        <p className="text-gray-500 dark:text-gray-400">رسم بياني للديون</p>
      </div>
    </div>
  );
}

function ProductsReportContent() {
  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <StatCard title="إجمالي المنتجات" value="150" icon={Package} />
        <StatCard title="منتجات نشطة" value="120" icon={CheckCircle} />
        <StatCard title="منتجات مستعملة" value="45" icon={Package} />
        <StatCard title="منتجات جديدة" value="105" icon={Package} />
      </div>
      
      <div className="h-64 bg-gray-50 dark:bg-gray-800 rounded-lg flex items-center justify-center">
        <p className="text-gray-500 dark:text-gray-400">رسم بياني للمنتجات</p>
      </div>
    </div>
  );
}

function SuppliersReportContent() {
  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <StatCard title="إجمالي الموردين" value="25" icon={Truck} />
        <StatCard title="إجمالي المشتريات" value="₪82,000" icon={DollarSign} />
        <StatCard title="المدفوع" value="₪70,000" icon={CheckCircle} />
        <StatCard title="المستحق" value="₪12,000" icon={AlertTriangle} />
      </div>
      
      <div className="h-64 bg-gray-50 dark:bg-gray-800 rounded-lg flex items-center justify-center">
        <p className="text-gray-500 dark:text-gray-400">رسم بياني للموردين</p>
      </div>
    </div>
  );
}

function ExpensesReportContent() {
  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <StatCard title="إجمالي المصروفات" value="₪12,400" icon={CreditCard} />
        <StatCard title="الإيجار" value="₪4,000" icon={DollarSign} />
        <StatCard title="الرواتب" value="₪5,000" icon={DollarSign} />
        <StatCard title="المرافق" value="₪1,200" icon={DollarSign} />
      </div>

      <div className="h-64 bg-gray-50 dark:bg-gray-800 rounded-lg flex items-center justify-center">
        <p className="text-gray-500 dark:text-gray-400">رسم بياني للمصروفات</p>
      </div>
    </div>
  );
}

function ReturnsReportContent() {
  const { t } = useTranslation();
  const { data: returnsData, isLoading } = useQuery({
    queryKey: ['reports', 'returns'],
    queryFn: () => reportsApi.returns(),
  });

  const data = returnsData as any;

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg animate-pulse">
            <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded mb-2"></div>
            <div className="h-8 bg-gray-200 dark:bg-gray-700 rounded"></div>
          </div>
          <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg animate-pulse">
            <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded mb-2"></div>
            <div className="h-8 bg-gray-200 dark:bg-gray-700 rounded"></div>
          </div>
          <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg animate-pulse">
            <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded mb-2"></div>
            <div className="h-8 bg-gray-200 dark:bg-gray-700 rounded"></div>
          </div>
          <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg animate-pulse">
            <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded mb-2"></div>
            <div className="h-8 bg-gray-200 dark:bg-gray-700 rounded"></div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <StatCard 
          title={t('reports.totalReturns')} 
          value={data?.totalReturns || 0} 
          icon={Package} 
        />
        <StatCard 
          title={t('reports.refundAmount')} 
          value={`₪${(data?.refundAmount || 0).toLocaleString()}`} 
          icon={DollarSign} 
        />
        <StatCard 
          title={t('reports.pendingReturns')} 
          value={data?.pendingReturns || 0} 
          icon={Clock} 
        />
        <StatCard 
          title={t('reports.returnRate')} 
          value={`${(data?.returnRate || 0).toFixed(1)}%`} 
          icon={TrendingUp} 
        />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
          <h3 className="font-semibold text-gray-900 dark:text-gray-100 mb-3">حالة المرتجعات</h3>
          <div className="space-y-2">
            <div className="flex justify-between items-center">
              <span className="text-sm text-gray-600 dark:text-gray-400">{t('reports.approvedReturns')}</span>
              <Badge variant="success">{data?.approvedReturns || 0}</Badge>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm text-gray-600 dark:text-gray-400">{t('reports.rejectedReturns')}</span>
              <Badge variant="danger">{data?.rejectedReturns || 0}</Badge>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm text-gray-600 dark:text-gray-400">{t('reports.pendingReturns')}</span>
              <Badge variant="warning">{data?.pendingReturns || 0}</Badge>
            </div>
          </div>
        </div>

        <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
          <h3 className="font-semibold text-gray-900 dark:text-gray-100 mb-3">{t('reports.commonReasons')}</h3>
          <div className="space-y-2">
            {data?.commonReasons?.map((reason: any, index: number) => (
              <div key={index} className="flex justify-between items-center">
                <span className="text-sm text-gray-600 dark:text-gray-400">{reason.reason}</span>
                <span className="text-sm font-medium">{reason.percentage}%</span>
              </div>
            )) || (
              <p className="text-sm text-gray-500 dark:text-gray-400">{t('common.noData')}</p>
            )}
          </div>
        </div>
      </div>

      <div className="h-64 bg-gray-50 dark:bg-gray-800 rounded-lg flex items-center justify-center">
        <p className="text-gray-500 dark:text-gray-400">رسم بياني للمرتجعات</p>
      </div>
    </div>
  );
}

function WarrantyReportContent() {
  const { t } = useTranslation();
  const { data: warrantyData, isLoading } = useQuery({
    queryKey: ['reports', 'warranty'],
    queryFn: () => reportsApi.warranty(),
  });

  const data = warrantyData as any;

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg animate-pulse">
            <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded mb-2"></div>
            <div className="h-8 bg-gray-200 dark:bg-gray-700 rounded"></div>
          </div>
          <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg animate-pulse">
            <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded mb-2"></div>
            <div className="h-8 bg-gray-200 dark:bg-gray-700 rounded"></div>
          </div>
          <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg animate-pulse">
            <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded mb-2"></div>
            <div className="h-8 bg-gray-200 dark:bg-gray-700 rounded"></div>
          </div>
          <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg animate-pulse">
            <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded mb-2"></div>
            <div className="h-8 bg-gray-200 dark:bg-gray-700 rounded"></div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <StatCard 
          title={t('reports.totalWarranties')} 
          value={data?.totalWarranties || 0} 
          icon={Clock} 
        />
        <StatCard 
          title={t('reports.activeWarranties')} 
          value={data?.activeWarranties || 0} 
          icon={CheckCircle} 
        />
        <StatCard 
          title={t('reports.expiredWarranties')} 
          value={data?.expiredWarranties || 0} 
          icon={AlertTriangle} 
        />
        <StatCard 
          title={t('reports.warrantyClaims')} 
          value={data?.warrantyClaims || 0} 
          icon={FileText} 
        />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
          <h3 className="font-semibold text-gray-900 dark:text-gray-100 mb-3">حالة الضمانات</h3>
          <div className="space-y-2">
            <div className="flex justify-between items-center">
              <span className="text-sm text-gray-600 dark:text-gray-400">{t('reports.activeWarranties')}</span>
              <Badge variant="success">{data?.activeWarranties || 0}</Badge>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm text-gray-600 dark:text-gray-400">{t('reports.expiredWarranties')}</span>
              <Badge variant="danger">{data?.expiredWarranties || 0}</Badge>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm text-gray-600 dark:text-gray-400">{t('reports.expiringSoon')}</span>
              <Badge variant="warning">{data?.expiringSoon || 0}</Badge>
            </div>
          </div>
        </div>

        <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
          <h3 className="font-semibold text-gray-900 dark:text-gray-100 mb-3">إحصائيات المطالبات</h3>
          <div className="space-y-2">
            <div className="flex justify-between items-center">
              <span className="text-sm text-gray-600 dark:text-gray-400">{t('reports.claimRate')}</span>
              <span className="text-sm font-medium">{(data?.claimRate || 0).toFixed(1)}%</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm text-gray-600 dark:text-gray-400">{t('reports.averageWarrantyDays')}</span>
              <span className="text-sm font-medium">{data?.averageWarrantyDays || 0} يوم</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm text-gray-600 dark:text-gray-400">مطالبات معلقة</span>
              <span className="text-sm font-medium">{data?.pendingClaims || 0}</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm text-gray-600 dark:text-gray-400">مطالبات مكتملة</span>
              <span className="text-sm font-medium">{data?.completedClaims || 0}</span>
            </div>
          </div>
        </div>
      </div>

      <div className="h-64 bg-gray-50 dark:bg-gray-800 rounded-lg flex items-center justify-center">
        <p className="text-gray-500 dark:text-gray-400">رسم بياني للضمانات</p>
      </div>
    </div>
  );
}

interface StatCardProps {
  title: string;
  value: string | number;
  icon: any;
}

function StatCard({ title, value, icon: Icon }: StatCardProps) {
  return (
    <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm text-gray-600 dark:text-gray-400">{title}</p>
          <p className="text-xl font-bold text-gray-900 dark:text-gray-100 mt-1">
            {value}
          </p>
        </div>
        <Icon className="w-5 h-5 text-gray-400" />
      </div>
    </div>
  );
}