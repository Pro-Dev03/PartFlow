import { SimpleBarChart, SimpleLineChart, SimplePieChart } from '../../../design-system/components/charts';
import { addStoreDays, formatStoreDate, getStoreDateKey } from '../../../utils/store-time';

interface ReportChartsProps {
  data?: any;
  loading?: boolean;
  reportType?: string;
}

export function ReportCharts({ data, loading, reportType }: ReportChartsProps) {
  const normalizeReportLabel = (value: unknown): string => {
    const label = String(value ?? '').trim();
    return label.includes('Ø') || label.includes('Ù') ? 'غير مصنف' : label;
  };

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
          label: item.date ? formatStoreDate(item.date, 'ar-SA') : 'غير محدد',
            value: Number(reportType === 'profit' ? item.net_profit ?? 0 : item.revenue ?? item.net_revenue ?? 0),
            secondaryValue: reportType === 'sales-profit' ? Number(item.net_profit ?? 0) : undefined,
        })).filter((item: { value: number }) => Number.isFinite(item.value))
      : [];
      let dailyItems = rawDailyItems;
    const rangeStart = report.start_date ? getStoreDateKey(String(report.start_date)) : null;
    const rangeEnd = report.end_date ? getStoreDateKey(String(report.end_date)) : null;
    if (rawDailyItems.length > 0 && rangeStart && rangeEnd) {
      const startParts = rangeStart.split('-').map(Number);
      const endParts = rangeEnd.split('-').map(Number);
      const dayCount = Math.ceil((Date.UTC(endParts[0], endParts[1] - 1, endParts[2]) - Date.UTC(startParts[0], startParts[1] - 1, startParts[2])) / (24 * 60 * 60 * 1000));
      if (dayCount > 0 && dayCount <= 366) {
        const valuesByDate = new Map(rawDailyItems.map((item: any) => [item.dateKey, item.value]));
        const secondaryValuesByDate = new Map(rawDailyItems.map((item: any) => [item.dateKey, item.secondaryValue]));
        dailyItems = Array.from({ length: dayCount }, (_, index) => {
          const dateKey = addStoreDays(rangeStart, index);
          return {
            label: formatStoreDate(dateKey, 'ar-SA'),
            value: Number(valuesByDate.get(dateKey) ?? 0),
            secondaryValue: secondaryValuesByDate.has(dateKey) ? Number(secondaryValuesByDate.get(dateKey) ?? 0) : 0,
          };
        });
      }
    }
    const paymentItems = Object.entries(report.by_payment_method || {})
      .map(([label, value]) => ({ label, value: Number(value) }))
      .filter(item => ['cash', 'card', 'checks', 'check', 'cheque', 'transfer', 'bank_transfer'].includes(item.label.toLowerCase()))
      .filter(item => Number.isFinite(item.value) && item.value > 0);
    const monthlyItems = Array.isArray(report.by_month)
      ? report.by_month.map((item: any) => ({
          label: item.month ? formatStoreDate(item.month, 'ar-SA') : 'غير محدد',
          value: Number(reportType === 'profit' ? item.net_profit ?? 0 : item.amount ?? item.cost ?? item.revenue ?? item.net_profit ?? 0),
        })).filter((item: { value: number }) => Number.isFinite(item.value))
      : [];
    const objectData = (source: Record<string, unknown> | undefined) =>
      Object.entries(source || {}).map(([label, value]) => ({ label: normalizeReportLabel(label), value: Number(value) }))
        .filter(item => Number.isFinite(item.value) && item.value > 0);
    const categoryData = objectData(report.by_category);
    const inventoryProductData = Array.isArray(report.low_stock_items)
      ? report.low_stock_items.map((item: any) => ({
          label: String(item.product_name || item.name || 'منتج غير معروف'),
          value: Math.max(1, Number(item.min_stock ?? item.current_stock ?? item.stock ?? 1)),
        }))
      : [];
    const supplierData = Array.isArray(report.by_supplier)
      ? report.by_supplier.map((item: any) => ({ label: item.supplier_name || 'تاجر', value: Number(item.total_purchases ?? item.total_cost ?? 0) }))
      : [];
    const supplierBalanceData = Array.isArray(report.by_supplier)
      ? report.by_supplier.map((item: any) => ({ label: item.supplier_name || 'تاجر', value: Number(item.outstanding || 0) }))
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
    const netSalesProductData = Array.isArray(report.top_returned_products)
      ? report.top_returned_products.map((item: any) => ({
          label: String(item.product_name || 'غير معروف'),
          value: Number(item.net_revenue ?? item.gross_revenue ?? 0),
        }))
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
    const secondaryDateMap = new Map<string, number>();
    items.forEach((item: any) => {
      const date = item.date ? formatStoreDate(item.date, 'ar-SA') : item.label || 'غير محدد';
      const value = Number(item.value ?? item.amount ?? item.revenue ?? 0);
      dateMap.set(date, (dateMap.get(date) || 0) + value);
      if (item.secondaryValue !== undefined) {
        secondaryDateMap.set(date, (secondaryDateMap.get(date) || 0) + Number(item.secondaryValue || 0));
      }
    });

    const trendData = Array.from(dateMap.entries()).map(([label, value]) => ({
      label,
      value,
      ...(secondaryDateMap.has(label) ? { secondaryValue: secondaryDateMap.get(label) } : {}),
    }));
    const reportedNetProfit = Number(report.net_profit);
    const trendTotal = trendData.reduce((sum, item) => sum + item.value, 0);
    const hasReliableNetProfit = Number.isFinite(reportedNetProfit)
      && (reportedNetProfit !== 0 || trendData.length > 0);
    const profitTrendData = reportType === 'profit' && hasReliableNetProfit
      && (trendData.length === 0 || Math.abs(trendTotal - reportedNetProfit) > 0.01)
      ? [{ label: 'إجمالي الفترة', value: reportedNetProfit }]
      : trendData;

    // Source distribution (if items have payment_method or status field)
    const paymentLabels: Record<string, string> = {
      cash: 'نقدي',
      card: 'بطاقة',
      credit: 'دين',
      debt: 'دين',
      checks: 'شيكات',
      cheque: 'شيكات',
      check: 'شيكات',
      bitcoin: 'بيتكوين',
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
      trendData: reportType === 'suppliers' || reportType === 'purchases-suppliers'
        ? supplierBalanceData
        : profitTrendData,
      sourceData: reportType === 'suppliers' || reportType === 'purchases-suppliers' ? supplierSourceData : sourceData,
      productData: reportType === 'products'
        ? categoryData
        : reportType === 'net-sales'
          ? netSalesProductData.length > 0 ? netSalesProductData : categoryData
        : reportType === 'suppliers' || reportType === 'purchases-suppliers'
          ? supplierData
          : reportType === 'inventory'
            ? inventoryProductData.length > 0 ? inventoryProductData : distributionData
            : distributionData,
    };
  };

  const { trendData, sourceData, productData } = processChartData();
  const report = data?.data ?? data;
  const hasReturnedProductData = Array.isArray(report?.by_product) && report.by_product.length > 0;

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
      {(reportType === 'profit' ? false : (reportType !== 'returns' && reportType !== 'debts') || (reportType === 'returns' && hasReturnedProductData)) ? (
        <SimpleBarChart
          title={reportType === 'products'
            ? 'يحتاج انتباهك'
            : reportType === 'net-sales'
              ? 'صافي المبيعات حسب المنتج'
            : reportType === 'expenses'
              ? 'المصروفات حسب الفئة'
              : reportType === 'suppliers' || reportType === 'purchases-suppliers'
                ? 'المشتريات حسب التاجر'
                : productData.length > 0 ? 'أفضل المنتجات والمصادر' : 'التوزيع حسب الفئة'}
          data={productData}
          color="#14b8a6"
          loading={loading}
        />
      ) : null}
      {sourceData.length > 0 && (
        <SimplePieChart
          title={reportType === 'suppliers' || reportType === 'purchases-suppliers'
            ? 'المدفوع مقابل المستحق'
            : reportType === 'expenses' ? 'المصروفات حسب طريقة الدفع' : 'توزيع المدفوعات حسب الطريقة'}
          data={sourceData}
          loading={loading}
        />
      )}
      {reportType !== 'inventory' && reportType !== 'debts' && reportType !== 'returns' && !(reportType === 'purchases' && trendData.length <= 1) && (
        <div style={{ gridColumn: '1 / -1', width: '100%', maxWidth: '1200px', marginInline: 'auto' }}>
          <SimpleLineChart
            title={reportType === 'suppliers' || reportType === 'purchases-suppliers'
              ? 'المستحق حسب التاجر'
              : reportType === 'sales-profit'
                ? 'اتجاه المبيعات والأرباح'
              : reportType === 'profit'
                ? 'اتجاه صافي الربح'
                : reportType === 'expenses'
                  ? 'اتجاه المصروفات'
                  : reportType === 'purchases' ? 'اتجاه المشتريات الشهري' : 'اتجاه المبيعات قبل الضريبة'}
            valueLabel={reportType === 'suppliers' || reportType === 'purchases-suppliers'
              ? 'إجمالي المستحق'
              : reportType === 'sales-profit'
                ? 'إجمالي المبيعات قبل الضريبة'
              : reportType === 'profit'
                ? 'إجمالي صافي الربح'
                : reportType === 'expenses'
                  ? 'إجمالي المصروفات'
                  : reportType === 'purchases' ? 'إجمالي المشتريات' : 'إجمالي المبيعات قبل الضريبة'}
            peakLabel={reportType === 'profit' ? 'أعلى فترة' : reportType === 'purchases' ? 'أعلى شهر' : 'أعلى يوم'}
            summaryValue={reportType === 'sales-profit' ? Number(report?.total_revenue ?? 0) : reportType === 'profit' ? Number(report?.net_profit ?? 0) : undefined}
            secondaryLabel={reportType === 'sales-profit' ? 'صافي الربح' : undefined}
            secondarySummaryValue={reportType === 'sales-profit' ? Number(report?.net_profit ?? 0) : undefined}
            secondaryPeakLabel={reportType === 'sales-profit' ? 'أعلى صافي ربح' : undefined}
            secondaryColor="#2563eb"
            data={trendData}
            color="#10b981"
            loading={loading}
          />
        </div>
      )}
    </div>
  );
}
