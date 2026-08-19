import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { reportsApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { PageHeader } from '../../../components/ui/page-header';
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
  LineChart,
  Sparkles,
  TrendingDown,
  Database,
  Zap,
  Target
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
    <div className="space-y-md">
      {/* Page Header - Futuristic + Analytical */}
      <PageHeader
        eyebrow="Analytics Hub"
        title={t('reports.title')}
        description="تحليلات وتقارير شاملة عن أداء المحل مع رؤى ذكية"
        actions={
          <div className="flex items-center gap-sm">
            <Button variant="secondary" className="gap-2">
              <Download className="w-4 h-4" />
              {t('reports.export')}
            </Button>
            <Button variant="secondary" className="gap-2">
              <Printer className="w-4 h-4" />
              {t('reports.print')}
            </Button>
            <Button variant="secondary" className="gap-2">
              <Zap className="w-4 h-4" />
              تحديث
            </Button>
          </div>
        }
      />

      {/* AI Analytics Insight - Futuristic + Analytical */}
      <Card variant="ai">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Sparkles className="w-5 h-5 text-cyan" />
            AI Analytics Insight
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-md">
            <div className="w-8 h-8 rounded-sm bg-cyan/10 flex items-center justify-center flex-shrink-0">
              <Target className="w-4 h-4 text-cyan" />
            </div>
            <div>
              <p className="text-small font-semibold text-text">أداء قوي في المبيعات</p>
              <p className="text-tiny text-text-muted mt-1">
                نمو المبيعات بنسبة 15% مقارنة بالشهر الماضي. الفئة الأكثر أداءً: كروت الشاشة.
              </p>
              <Button variant="ghost" size="sm" className="mt-2 text-cyan">
                عرض التفاصيل ←
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Report Type Selection - Futuristic + Analytical */}
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-sm">
        {reportTypes.map((report) => {
          const Icon = report.icon;
          return (
            <Button
              key={report.id}
              variant={selectedReport === report.id ? 'primary' : 'secondary'}
              onClick={() => setSelectedReport(report.id)}
              className="flex flex-col items-center gap-sm h-auto py-md"
            >
              <Icon className="w-5 h-5" />
              <span className="text-tiny">{report.label}</span>
            </Button>
          );
        })}
      </div>

      {/* Filters - Futuristic + Analytical */}
      <Card>
        <CardContent className="p-lg">
          <div className="flex flex-col md:flex-row gap-md items-end">
            <div className="flex-1">
              <label className="text-small font-medium text-text mb-sm block">
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
                  <label className="text-small font-medium text-text mb-sm block">
                    {t('reports.startDate')}
                  </label>
                  <input
                    type="date"
                    className="w-full h-10 rounded-sm border border-border bg-surface px-3 py-2 text-small focus:outline-none focus:ring-2 focus:ring-cyan dark:bg-surface dark:border-border dark:text-text"
                  />
                </div>
                <div className="flex-1">
                  <label className="text-small font-medium text-text mb-sm block">
                    {t('reports.endDate')}
                  </label>
                  <input
                    type="date"
                    className="w-full h-10 rounded-sm border border-border bg-surface px-3 py-2 text-small focus:outline-none focus:ring-2 focus:ring-cyan dark:bg-surface dark:border-border dark:text-text"
                  />
                </div>
              </>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Report Content - Futuristic + Analytical */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <ReportIcon className="w-5 h-5 text-cyan" />
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
    <div className="space-y-md">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-md">
        <StatCard title="الإيرادات" value="₪100,000" icon={DollarSign} variant="featured" trend="+15%" trendUp={true} />
        <StatCard title="عدد الطلبات" value="150" icon={BarChart3} variant="default" trend="+8%" trendUp={true} />
        <StatCard title="القطع المباعة" value="320" icon={Package} variant="default" trend="+12%" trendUp={true} />
        <StatCard title="متوسط البيع" value="₪667" icon={TrendingUp} variant="info" trend="+5%" trendUp={true} />
      </div>
      
      <div className="h-64 bg-surface-2 rounded-sm flex items-center justify-center border border-border">
        <p className="text-small text-text-muted">رسم بياني للمبيعات</p>
      </div>
    </div>
  );
}

function ProfitReportContent() {
  return (
    <div className="space-y-md">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-md">
        <StatCard title="الإيرادات" value="₪100,000" icon={DollarSign} variant="featured" trend="+15%" trendUp={true} />
        <StatCard title="التكلفة" value="₪72,000" icon={Package} variant="warning" trend="+5%" trendUp={false} />
        <StatCard title="الربح الإجمالي" value="₪28,000" icon={TrendingUp} variant="success" trend="+20%" trendUp={true} />
        <StatCard title="الربح الصافي" value="₪20,000" icon={TrendingUp} variant="success" trend="+18%" trendUp={true} />
      </div>
      
      <div className="h-64 bg-surface-2 rounded-sm flex items-center justify-center border border-border">
        <p className="text-small text-text-muted">رسم بياني للأرباح</p>
      </div>
    </div>
  );
}

function InventoryReportContent() {
  return (
    <div className="space-y-md">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-md">
        <StatCard title="قيمة المخزون" value="₪185,400" icon={Package} variant="featured" trend="+10%" trendUp={true} />
        <StatCard title="منخفض المخزون" value="12" icon={AlertTriangle} variant="danger" trend="+2" trendUp={false} />
        <StatCard title="المخزون الراكد" value="8" icon={Clock} variant="warning" trend="-1" trendUp={true} />
        <StatCard title="سريع الحركة" value="25" icon={TrendingUp} variant="success" trend="+5%" trendUp={true} />
      </div>
      
      <div className="h-64 bg-surface-2 rounded-sm flex items-center justify-center border border-border">
        <p className="text-small text-text-muted">رسم بياني للمخزون</p>
      </div>
    </div>
  );
}

function DebtsReportContent() {
  return (
    <div className="space-y-md">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-md">
        <StatCard title="إجمالي الديون" value="₪24,800" icon={DollarSign} variant="default" trend="+3%" trendUp={false} />
        <StatCard title="متأخرة" value="₪8,200" icon={AlertTriangle} variant="danger" trend="+2" trendUp={false} />
        <StatCard title="مستحقة قريباً" value="₪16,600" icon={Calendar} variant="warning" trend="+1" trendUp={false} />
        <StatCard title="مدفوعة" value="₪12,400" icon={CheckCircle} variant="success" trend="+15%" trendUp={true} />
      </div>
      
      <div className="h-64 bg-surface-2 rounded-sm flex items-center justify-center border border-border">
        <p className="text-small text-text-muted">رسم بياني للديون</p>
      </div>
    </div>
  );
}

function ProductsReportContent() {
  return (
    <div className="space-y-md">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-md">
        <StatCard title="إجمالي المنتجات" value="150" icon={Package} variant="featured" trend="+8%" trendUp={true} />
        <StatCard title="منتجات نشطة" value="120" icon={CheckCircle} variant="success" trend="+10%" trendUp={true} />
        <StatCard title="منتجات مستعملة" value="45" icon={Package} variant="info" trend="+5%" trendUp={true} />
        <StatCard title="منتجات جديدة" value="105" icon={Package} variant="default" trend="+12%" trendUp={true} />
      </div>
      
      <div className="h-64 bg-surface-2 rounded-sm flex items-center justify-center border border-border">
        <p className="text-small text-text-muted">رسم بياني للمنتجات</p>
      </div>
    </div>
  );
}

function SuppliersReportContent() {
  return (
    <div className="space-y-md">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-md">
        <StatCard title="إجمالي الموردين" value="25" icon={Truck} variant="featured" trend="+2" trendUp={true} />
        <StatCard title="إجمالي المشتريات" value="₪82,000" icon={DollarSign} variant="default" trend="+8%" trendUp={true} />
        <StatCard title="المدفوع" value="₪70,000" icon={CheckCircle} variant="success" trend="+12%" trendUp={true} />
        <StatCard title="المستحق" value="₪12,000" icon={AlertTriangle} variant="warning" trend="+3" trendUp={false} />
      </div>
      
      <div className="h-64 bg-surface-2 rounded-sm flex items-center justify-center border border-border">
        <p className="text-small text-text-muted">رسم بياني للموردين</p>
      </div>
    </div>
  );
}

function ExpensesReportContent() {
  return (
    <div className="space-y-md">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-md">
        <StatCard title="إجمالي المصروفات" value="₪12,400" icon={CreditCard} variant="default" trend="+5%" trendUp={false} />
        <StatCard title="الإيجار" value="₪4,000" icon={DollarSign} variant="info" trend="0%" trendUp={null} />
        <StatCard title="الرواتب" value="₪5,000" icon={DollarSign} variant="info" trend="+3%" trendUp={false} />
        <StatCard title="المرافق" value="₪1,200" icon={DollarSign} variant="info" trend="+2%" trendUp={false} />
      </div>

      <div className="h-64 bg-surface-2 rounded-sm flex items-center justify-center border border-border">
        <p className="text-small text-text-muted">رسم بياني للمصروفات</p>
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
      <div className="space-y-md">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-md">
          <div className="p-lg bg-surface-2 rounded-sm animate-pulse">
            <div className="h-4 bg-surface-2 rounded-sm mb-2"></div>
            <div className="h-8 bg-surface-2 rounded-sm"></div>
          </div>
          <div className="p-lg bg-surface-2 rounded-sm animate-pulse">
            <div className="h-4 bg-surface-2 rounded-sm mb-2"></div>
            <div className="h-8 bg-surface-2 rounded-sm"></div>
          </div>
          <div className="p-lg bg-surface-2 rounded-sm animate-pulse">
            <div className="h-4 bg-surface-2 rounded-sm mb-2"></div>
            <div className="h-8 bg-surface-2 rounded-sm"></div>
          </div>
          <div className="p-lg bg-surface-2 rounded-sm animate-pulse">
            <div className="h-4 bg-surface-2 rounded-sm mb-2"></div>
            <div className="h-8 bg-surface-2 rounded-sm"></div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-md">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-md">
        <StatCard 
          title={t('reports.totalReturns')} 
          value={data?.totalReturns || 0} 
          icon={Package} 
          variant="default"
          trend="+5%"
          trendUp={false}
        />
        <StatCard 
          title={t('reports.refundAmount')} 
          value={`₪${(data?.refundAmount || 0).toLocaleString()}`} 
          icon={DollarSign} 
          variant="warning"
          trend="+3%"
          trendUp={false}
        />
        <StatCard 
          title={t('reports.pendingReturns')} 
          value={data?.pendingReturns || 0} 
          icon={Clock} 
          variant="warning"
          trend="+2"
          trendUp={false}
        />
        <StatCard 
          title={t('reports.returnRate')} 
          value={`${(data?.returnRate || 0).toFixed(1)}%`} 
          icon={TrendingUp} 
          variant="danger"
          trend="+1%"
          trendUp={false}
        />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-md">
        <div className="p-lg bg-surface-2 rounded-sm border border-border">
          <h3 className="text-small font-semibold text-text mb-md">حالة المرتجعات</h3>
          <div className="space-y-sm">
            <div className="flex justify-between items-center">
              <span className="text-small text-text-muted">{t('reports.approvedReturns')}</span>
              <Badge variant="success" size="sm">{data?.approvedReturns || 0}</Badge>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-small text-text-muted">{t('reports.rejectedReturns')}</span>
              <Badge variant="danger" size="sm">{data?.rejectedReturns || 0}</Badge>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-small text-text-muted">{t('reports.pendingReturns')}</span>
              <Badge variant="warning" size="sm">{data?.pendingReturns || 0}</Badge>
            </div>
          </div>
        </div>

        <div className="p-lg bg-surface-2 rounded-sm border border-border">
          <h3 className="text-small font-semibold text-text mb-md">{t('reports.commonReasons')}</h3>
          <div className="space-y-sm">
            {data?.commonReasons?.map((reason: any, index: number) => (
              <div key={index} className="flex justify-between items-center">
                <span className="text-small text-text-muted">{reason.reason}</span>
                <span className="text-small font-medium text-text">{reason.percentage}%</span>
              </div>
            )) || (
              <p className="text-small text-text-muted">{t('common.noData')}</p>
            )}
          </div>
        </div>
      </div>

      <div className="h-64 bg-surface-2 rounded-sm flex items-center justify-center border border-border">
        <p className="text-small text-text-muted">رسم بياني للمرتجعات</p>
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
      <div className="space-y-md">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-md">
          <div className="p-lg bg-surface-2 rounded-sm animate-pulse">
            <div className="h-4 bg-surface-2 rounded-sm mb-2"></div>
            <div className="h-8 bg-surface-2 rounded-sm"></div>
          </div>
          <div className="p-lg bg-surface-2 rounded-sm animate-pulse">
            <div className="h-4 bg-surface-2 rounded-sm mb-2"></div>
            <div className="h-8 bg-surface-2 rounded-sm"></div>
          </div>
          <div className="p-lg bg-surface-2 rounded-sm animate-pulse">
            <div className="h-4 bg-surface-2 rounded-sm mb-2"></div>
            <div className="h-8 bg-surface-2 rounded-sm"></div>
          </div>
          <div className="p-lg bg-surface-2 rounded-sm animate-pulse">
            <div className="h-4 bg-surface-2 rounded-sm mb-2"></div>
            <div className="h-8 bg-surface-2 rounded-sm"></div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-md">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-md">
        <StatCard 
          title={t('reports.totalWarranties')} 
          value={data?.totalWarranties || 0} 
          icon={Clock} 
          variant="featured"
          trend="+8%"
          trendUp={true}
        />
        <StatCard 
          title={t('reports.activeWarranties')} 
          value={data?.activeWarranties || 0} 
          icon={CheckCircle} 
          variant="success"
          trend="+10%"
          trendUp={true}
        />
        <StatCard 
          title={t('reports.expiredWarranties')} 
          value={data?.expiredWarranties || 0} 
          icon={AlertTriangle} 
          variant="danger"
          trend="+2"
          trendUp={false}
        />
        <StatCard 
          title={t('reports.warrantyClaims')} 
          value={data?.warrantyClaims || 0} 
          icon={FileText} 
          variant="warning"
          trend="+5%"
          trendUp={false}
        />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-md">
        <div className="p-lg bg-surface-2 rounded-sm border border-border">
          <h3 className="text-small font-semibold text-text mb-md">حالة الضمانات</h3>
          <div className="space-y-sm">
            <div className="flex justify-between items-center">
              <span className="text-small text-text-muted">{t('reports.activeWarranties')}</span>
              <Badge variant="success" size="sm">{data?.activeWarranties || 0}</Badge>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-small text-text-muted">{t('reports.expiredWarranties')}</span>
              <Badge variant="danger" size="sm">{data?.expiredWarranties || 0}</Badge>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-small text-text-muted">{t('reports.expiringSoon')}</span>
              <Badge variant="warning" size="sm">{data?.expiringSoon || 0}</Badge>
            </div>
          </div>
        </div>

        <div className="p-lg bg-surface-2 rounded-sm border border-border">
          <h3 className="text-small font-semibold text-text mb-md">إحصائيات المطالبات</h3>
          <div className="space-y-sm">
            <div className="flex justify-between items-center">
              <span className="text-small text-text-muted">{t('reports.claimRate')}</span>
              <span className="text-small font-medium text-text">{(data?.claimRate || 0).toFixed(1)}%</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-small text-text-muted">{t('reports.averageWarrantyDays')}</span>
              <span className="text-small font-medium text-text">{data?.averageWarrantyDays || 0} يوم</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-small text-text-muted">مطالبات معلقة</span>
              <span className="text-small font-medium text-text">{data?.pendingClaims || 0}</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-small text-text-muted">مطالبات مكتملة</span>
              <span className="text-small font-medium text-text">{data?.completedClaims || 0}</span>
            </div>
          </div>
        </div>
      </div>

      <div className="h-64 bg-surface-2 rounded-sm flex items-center justify-center border border-border">
        <p className="text-small text-text-muted">رسم بياني للضمانات</p>
      </div>
    </div>
  );
}

interface StatCardProps {
  title: string;
  value: string | number;
  icon: any;
  subtitle?: string;
  variant?: 'default' | 'featured' | 'warning' | 'ai' | 'danger' | 'success' | 'info';
  trend?: string | null;
  trendUp?: boolean | null;
}

function StatCard({ title, value, icon: Icon, subtitle, variant = 'default', trend, trendUp }: StatCardProps) {
  return (
    <Card 
      variant={variant} 
      className="hover:border-border/22 hover:-translate-y-1 cursor-pointer"
      hoverable
    >
      <CardContent className="p-lg">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <p className="text-small text-text-muted">{title}</p>
            <p className="text-metric font-bold text-text mt-1">
              {value}
            </p>
            {subtitle && (
              <p className="text-tiny text-text-muted mt-1">
                {subtitle}
              </p>
            )}
            {trend && (
              <div className="flex items-center gap-1 mt-2">
                {trendUp ? (
                  <TrendingUp className="w-3 h-3 text-green" />
                ) : (
                  <TrendingDown className="w-3 h-3 text-red" />
                )}
                <span className={`text-tiny ${trendUp ? 'text-green' : 'text-red'}`}>
                  {trend}
                </span>
              </div>
            )}
          </div>
          <div className={`w-10 h-10 rounded-sm flex items-center justify-center ${
            variant === 'featured' ? 'bg-cyan/10' :
            variant === 'warning' ? 'bg-yellow/10' :
            variant === 'danger' ? 'bg-red/10' :
            variant === 'success' ? 'bg-green/10' :
            variant === 'info' ? 'bg-cyan/10' :
            variant === 'ai' ? 'bg-cyan/10' :
            'bg-cyan/10'
          }`}>
            <Icon className={`w-5 h-5 ${
              variant === 'featured' ? 'text-cyan' :
              variant === 'warning' ? 'text-yellow' :
              variant === 'danger' ? 'text-red' :
              variant === 'success' ? 'text-green' :
              variant === 'info' ? 'text-cyan' :
              variant === 'ai' ? 'text-cyan' :
              'text-cyan'
            }`} />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}