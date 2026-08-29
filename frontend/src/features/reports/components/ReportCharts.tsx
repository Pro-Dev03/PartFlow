import { SimpleBarChart, SimpleLineChart, SimplePieChart } from '../../../components/ui/charts';

interface ReportChartsProps {
  data?: any;
  loading?: boolean;
  reportType?: string;
}

export function ReportCharts({ data, loading, reportType }: ReportChartsProps) {
  // Process data for charts
  const processChartData = () => {
    const report = data?.data ?? data;
    if (!report || typeof report !== 'object') {
      return {
        categoryData: [],
        trendData: [],
        sourceData: [],
        productData: [],
      };
    }

    const dailyItems = Array.isArray(report.by_day)
      ? report.by_day.map((item: any) => ({
          label: item.date ? new Date(item.date).toLocaleDateString('ar-SA', { day: 'numeric', month: 'short' }) : 'غير محدد',
          value: Number(item.revenue ?? item.net_revenue ?? 0),
        })).filter((item: { value: number }) => Number.isFinite(item.value))
      : [];
    const paymentItems = Object.entries(report.by_payment_method || {})
      .map(([label, value]) => ({ label, value: Number(value) }))
      .filter(item => Number.isFinite(item.value) && item.value > 0);
    const monthlyItems = Array.isArray(report.by_month)
      ? report.by_month.map((item: any) => ({
          label: item.month ? new Date(item.month).toLocaleDateString('ar-SA', { month: 'short', year: 'numeric' }) : 'غير محدد',
          value: Number(item.amount ?? item.cost ?? item.revenue ?? item.net_profit ?? 0),
        })).filter((item: { value: number }) => Number.isFinite(item.value))
      : [];
    const objectData = (source: Record<string, unknown> | undefined) =>
      Object.entries(source || {}).map(([label, value]) => ({ label, value: Number(value) }))
        .filter(item => Number.isFinite(item.value) && item.value > 0);
    const categoryData = objectData(report.by_category);
    const attentionData = [
      { label: 'منخفض المخزون', value: Number(report.low_stock_count || 0) },
      { label: 'غير مصنف', value: Number(report.by_category?.['غير مصنف'] || 0) },
    ].filter(item => Number.isFinite(item.value) && item.value > 0);
    const supplierData = Array.isArray(report.by_supplier)
      ? report.by_supplier.map((item: any) => ({ label: item.supplier_name || 'مورد', value: Number(item.total_purchases ?? item.total_cost ?? 0) }))
      : [];
    const supplierBalanceData = Array.isArray(report.by_supplier)
      ? report.by_supplier.map((item: any) => ({ label: item.supplier_name || 'مورد', value: Number(item.outstanding || 0) }))
          .filter((item: { value: number }) => Number.isFinite(item.value) && item.value > 0)
      : [];
    const productData = Array.isArray(report.top_products)
      ? report.top_products.map((item: any) => ({ label: String(item.product_name || 'غير معروف'), value: Number(item.revenue || 0) }))
          .filter((item: { value: number }) => Number.isFinite(item.value) && item.value > 0)
      : [];
    const returnedProductData = Array.isArray(report.by_product)
      ? report.by_product.map((item: any) => ({ label: String(item.product_name || 'غير معروف'), value: Number(item.refund_amount || 0) }))
          .filter((item: { value: number }) => Number.isFinite(item.value) && item.value > 0)
      : [];
    const distributionData = productData.length > 0 ? productData
      : returnedProductData.length > 0 ? returnedProductData
      : supplierData.length > 0 ? supplierData : categoryData;
    const items = dailyItems.length > 0 ? dailyItems : monthlyItems;

    // Trend data by date (if items have date field)
    const dateMap = new Map<string, number>();
    items.forEach((item: any) => {
      const date = item.date ? new Date(item.date).toLocaleDateString('ar-SA', { day: 'numeric', month: 'short' }) : item.label || 'غير محدد';
      const value = Number(item.value ?? item.amount ?? item.revenue ?? 0);
      dateMap.set(date, (dateMap.get(date) || 0) + value);
    });

    const trendData = Array.from(dateMap.entries()).map(([label, value]) => ({
      label,
      value,
    }));

    // Source distribution (if items have payment_method or status field)
    const paymentLabels: Record<string, string> = {
      cash: 'نقدي',
      card: 'بطاقة',
      credit: 'دين',
      debt: 'دين',
      checks: 'شيكات',
      cheque: 'شيكات',
      check: 'شيكات',
      transfer: 'تحويل بنكي',
      bank_transfer: 'تحويل بنكي',
    };
    const sourceData = paymentItems.map(({ label, value }, index) => ({
      label: paymentLabels[label.toLowerCase()] || label,
      value,
      color: ['#14b8a6', '#10b981', '#f59e0b', '#ef4444'][index % 4],
    }));
    const supplierSourceData = [
      { label: 'المدفوع', value: Number(report.total_paid || 0), color: '#10b981' },
      { label: 'المستحق', value: Number(report.total_outstanding || 0), color: '#f59e0b' },
    ].filter(item => Number.isFinite(item.value) && item.value > 0);
    return {
      categoryData,
      trendData: reportType === 'suppliers' ? supplierBalanceData : trendData,
      sourceData: reportType === 'suppliers' ? supplierSourceData : sourceData,
      productData: reportType === 'products' ? attentionData : reportType === 'suppliers' ? supplierData : distributionData,
    };
  };

  const { categoryData, trendData, sourceData, productData } = processChartData();

  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: '14px' }}
         className="grid-cols-1 lg:grid-cols-2">
      <SimpleBarChart
        title={reportType === 'products' ? 'يحتاج انتباهك' : reportType === 'suppliers' ? 'المشتريات حسب المورد' : productData.length > 0 ? 'أفضل المنتجات والمصادر' : 'التوزيع حسب الفئة'}
        data={productData}
        color="#14b8a6"
        loading={loading}
      />
      <div style={{ gridColumn: '1 / -1' }}>
        <SimpleLineChart
          title={reportType === 'suppliers' ? 'المستحق حسب المورد' : 'اتجاه المبيعات'}
          data={trendData}
          color="#10b981"
          loading={loading}
        />
      </div>
      <SimplePieChart
        title={reportType === 'suppliers' ? 'المدفوع مقابل المستحق' : 'توزيع مصادر الدخل'}
        data={sourceData}
        loading={loading}
      />
    </div>
  );
}
