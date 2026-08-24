import { SimpleBarChart, SimpleLineChart, SimplePieChart } from '../../../components/ui/charts';

export function ReportCharts() {
  return (
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
  );
}
