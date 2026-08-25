import { SimpleBarChart, SimpleLineChart, SimplePieChart } from '../../../components/ui/charts';

interface ReportChartsProps {
  data?: any;
  loading?: boolean;
}

export function ReportCharts({ data, loading }: ReportChartsProps) {
  // Process data for charts
  const processChartData = () => {
    if (!data || !data.data || data.data.length === 0) {
      return {
        categoryData: [],
        trendData: [],
        sourceData: [],
      };
    }

    const items = data.data;

    // Category distribution (if items have category field)
    const categoryMap = new Map<string, number>();
    items.forEach((item: any) => {
      const category = item.category || item.type || 'غير مصنف';
      const value = item.value || item.amount || 0;
      categoryMap.set(category, (categoryMap.get(category) || 0) + value);
    });

    const categoryData = Array.from(categoryMap.entries()).map(([label, value]) => ({
      label,
      value,
    }));

    // Trend data by date (if items have date field)
    const dateMap = new Map<string, number>();
    items.forEach((item: any) => {
      const date = item.date ? new Date(item.date).toLocaleDateString('ar-SA', { month: 'short' }) : 'غير محدد';
      const value = item.value || item.amount || 0;
      dateMap.set(date, (dateMap.get(date) || 0) + value);
    });

    const trendData = Array.from(dateMap.entries()).map(([label, value]) => ({
      label,
      value,
    }));

    // Source distribution (if items have payment_method or status field)
    const sourceMap = new Map<string, number>();
    items.forEach((item: any) => {
      const source = item.payment_method || item.status || 'غير محدد';
      const value = item.value || item.amount || 0;
      sourceMap.set(source, (sourceMap.get(source) || 0) + value);
    });

    const sourceData = Array.from(sourceMap.entries()).map(([label, value], index) => ({
      label,
      value,
      color: ['#14b8a6', '#10b981', '#f59e0b', '#ef4444'][index % 4],
    }));

    return {
      categoryData,
      trendData,
      sourceData,
    };
  };

  const { categoryData, trendData, sourceData } = processChartData();

  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: '14px' }}
         className="grid-cols-1 lg:grid-cols-2">
      <SimpleBarChart
        title="توزيع المبيعات حسب الفئة"
        data={categoryData.length > 0 ? categoryData : [{ label: 'لا توجد بيانات', value: 0 }]}
        color="#14b8a6"
        loading={loading}
      />
      <SimpleLineChart
        title="اتجاه المبيعات"
        data={trendData.length > 0 ? trendData : [{ label: 'لا توجد بيانات', value: 0 }]}
        color="#10b981"
        loading={loading}
      />
      <SimplePieChart
        title="توزيع مصادر الدخل"
        data={sourceData.length > 0 ? sourceData : [{ label: 'لا توجد بيانات', value: 0, color: '#14b8a6' }]}
        loading={loading}
      />
    </div>
  );
}
