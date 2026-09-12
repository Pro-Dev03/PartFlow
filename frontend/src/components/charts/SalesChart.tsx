import { Bar, CartesianGrid, ComposedChart, Legend, Line, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';

interface SalesChartProps {
  data: Array<{
    name: string;
    sales: number;
    profit: number;
  }>;
}

function formatChartDate(value: string) {
  const date = new Date(`${value}T00:00:00`);
  if (Number.isNaN(date.getTime())) return value;

  return new Intl.DateTimeFormat('ar', {
    day: 'numeric',
    month: 'short',
  }).format(date);
}

export function SalesChart({ data }: SalesChartProps) {
  return (
    <ResponsiveContainer width="100%" height={220}>
      <ComposedChart data={data} margin={{ top: 8, right: 8, left: 4, bottom: 0 }}>
        <CartesianGrid vertical={false} stroke="var(--border-default)" strokeDasharray="4 5" opacity={0.65} />
        <XAxis
          dataKey="name"
          axisLine={false}
          tickLine={false}
          tick={{ fill: 'var(--text-secondary)', fontSize: 11 }}
          tickFormatter={formatChartDate}
        />
        <YAxis
          width={48}
          axisLine={false}
          tickLine={false}
          tick={{ fill: 'var(--text-secondary)', fontSize: 11 }}
          tickFormatter={(value: number) => `₪${Number(value).toLocaleString('en-US')}`}
        />
        <Tooltip
          contentStyle={{
            backgroundColor: 'var(--bg-surface-elevated)',
            border: '1px solid var(--border-default)',
            borderRadius: '12px',
            boxShadow: 'var(--shadow-md)',
            color: 'var(--text-primary)',
          }}
          labelStyle={{ color: 'var(--text-secondary)', marginBottom: '4px' }}
          formatter={(value: number | undefined, name: string | undefined) => [
            `₪${Number(value ?? 0).toLocaleString()}`,
            name === 'sales' ? 'المبيعات' : 'الربح الإجمالي',
          ]}
        />
        <Legend
          verticalAlign="top"
          align="right"
          iconType="circle"
          wrapperStyle={{ paddingBottom: '12px', fontSize: '11px', color: 'var(--text-secondary)' }}
          formatter={(value) => value === 'sales' ? 'المبيعات' : 'الربح الإجمالي'}
        />
        <Bar
          dataKey="sales"
          name="sales"
          fill="var(--color-primary)"
          radius={[5, 5, 0, 0]}
          maxBarSize={42}
        />
        <Line
          type="monotone"
          dataKey="profit"
          stroke="var(--color-success)"
          strokeWidth={3}
          dot={{ r: 3, fill: 'var(--color-success)', strokeWidth: 0 }}
          activeDot={{ r: 5, strokeWidth: 2, stroke: 'var(--bg-surface)' }}
          name="profit"
        />
      </ComposedChart>
    </ResponsiveContainer>
  );
}