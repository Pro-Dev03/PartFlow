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

    const rawDailyItems = Array.isArray(report.by_day)
      ? report.by_day.map((item: any) => ({
          dateKey: item.date ? String(item.date).slice(0, 10) : '',
          label: item.date ? new Date(item.date).toLocaleDateString('ar-SA', { day: 'numeric', month: 'short' }) : 'غير محدد',
          value: Number(item.revenue ?? item.net_revenue ?? 0),
        })).filter((item: { value: number }) => Number.isFinite(item.value))
      : [];
    let dailyItems = reportType === 'profit' ? [] : rawDailyItems;
    const rangeStart = report.start_date ? new Date(String(report.start_date)) : null;
    const rangeEnd = report.end_date ? new Date(String(report.end_date)) : null;
    if (reportType !== 'profit' && rawDailyItems.length > 0 && rangeStart && rangeEnd && !Number.isNaN(rangeStart.getTime()) && !Number.isNaN(rangeEnd.getTime())) {
      const dayCount = Math.ceil((rangeEnd.getTime() - rangeStart.getTime()) / (1000 * 60 * 60 * 24));
      if (dayCount > 0 && dayCount <= 366) {
        const valuesByDate = new Map(rawDailyItems.map((item: any) => [item.dateKey, item.value]));
        dailyItems = Array.from({ length: dayCount }, (_, index) => {
          const date = new Date(rangeStart.getTime() + index * 24 * 60 * 60 * 1000);
          const dateKey = date.toISOString().slice(0, 10);
          return {
            label: date.toLocaleDateString('ar-SA', { day: 'numeric', month: 'short', timeZone: 'UTC' }),
            value: Number(valuesByDate.get(dateKey) ?? 0),
          };
        });
      }
    }
    const paymentItems = Object.entries(report.by_payment_method || {})
      .map(([label, value]) => ({ label, value: Number(value) }))
      .filter(item => Number.isFinite(item.value) && item.value > 0);
    const monthlyItems = Array.isArray(report.by_month)
      ? report.by_month.map((item: any) => ({
          label: item.month ? new Date(item.month).toLocaleDateString('ar-SA', { month: 'short', year: 'numeric' }) : 'غير محدد',
          value: Number(reportType === 'profit' ? item.net_profit ?? 0 : item.amount ?? item.cost ?? item.revenue ?? item.net_profit ?? 0),
        })).filter((item: { value: number }) => Number.isFinite(item.value))
      : [];
    const objectData = (source: Record<string, unknown> | undefined) =>
      Object.entries(source || {}).map(([label, value]) => ({ label, value: Number(value) }))
        .filter(item => Number.isFinite(item.value) && item.value > 0);
    const categoryData = objectData(report.by_category);
    const inventoryProductData = Array.isArray(report.low_stock_items)
      ? report.low_stock_items.map((item: any) => ({
          label: String(item.product_name || item.name || 'منتج غير معروف'),
          value: Math.max(1, Number(item.min_stock ?? item.current_stock ?? item.stock ?? 1)),
        }))
      : [];
    const attentionData = [
      { label: 'منخفض المخزون', value: Number(report.low_stock_count || (Array.isArray(report.low_stock_items) ? report.low_stock_items.length : 0)) },
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
    const distributionData = reportType === 'inventory' && inventoryProductData.length > 0
      ? inventoryProductData
      : productData.length > 0 ? productData
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
      productData: reportType === 'products'
        ? categoryData
        : reportType === 'suppliers'
          ? supplierData
          : reportType === 'inventory'
            ? inventoryProductData.length > 0 ? inventoryProductData : distributionData
            : distributionData,
    };
  };

  const { categoryData, trendData, sourceData, productData } = processChartData();

  if (reportType === 'tax') {
    return null;
  }

  if (reportType === 'products') {
    return (
      <div style={{ display: 'grid', gridTemplateColumns: 'minmax(0, 1fr)', gap: '14px' }}>
        <SimpleBarChart
          title="المنتجات حسب التصنيف"
          data={productData}
          color="#14b8a6"
          loading={loading}
        />
      </div>
    );
  }

  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: '14px' }}
         className="grid-cols-1 lg:grid-cols-2">
      <SimpleBarChart
        title={reportType === 'products'
          ? 'يحتاج انتباهك'
          : reportType === 'expenses'
            ? 'المصروفات حسب الفئة'
            : reportType === 'suppliers' || reportType === 'purchases'
              ? 'المشتريات حسب المورد'
              : productData.length > 0 ? 'أفضل المنتجات والمصادر' : 'التوزيع حسب الفئة'}
        data={productData}
        color="#14b8a6"
        loading={loading}
      />
      {!(reportType === 'purchases' && trendData.length <= 1) && (
        <div style={{ gridColumn: '1 / -1', width: '100%', maxWidth: '1200px', marginInline: 'auto' }}>
          <SimpleLineChart
            title={reportType === 'suppliers'
              ? 'المستحق حسب المورد'
              : reportType === 'profit'
                ? 'اتجاه صافي الربح'
                : reportType === 'expenses'
                  ? 'اتجاه المصروفات'
                  : reportType === 'purchases' ? 'اتجاه المشتريات الشهري' : 'اتجاه المبيعات قبل الضريبة'}
            valueLabel={reportType === 'suppliers'
              ? 'إجمالي المستحق'
              : reportType === 'profit'
                ? 'إجمالي صافي الربح'
                : reportType === 'expenses'
                  ? 'إجمالي المصروفات'
                  : reportType === 'purchases' ? 'إجمالي المشتريات' : 'إجمالي المبيعات قبل الضريبة'}
            peakLabel={reportType === 'profit' || reportType === 'purchases' ? 'أعلى شهر' : 'أعلى يوم'}
            data={trendData}
            color="#10b981"
            loading={loading}
          />
        </div>
      )}
      {sourceData.length > 0 && (
        <SimplePieChart
          title={reportType === 'suppliers'
            ? 'المدفوع مقابل المستحق'
            : reportType === 'expenses' ? 'المصروفات حسب طريقة الدفع' : 'توزيع المبيعات حسب طريقة الدفع'}
          data={sourceData}
          loading={loading}
        />
      )}
    </div>
  );
}
