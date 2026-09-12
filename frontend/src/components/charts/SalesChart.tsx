import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

interface SalesChartProps {
  data: Array<{
    name: string;
    sales: number;
    profit: number;
  }>;
}

export function SalesChart({ data }: SalesChartProps) {
  return (
    <ResponsiveContainer width="100%" height={220}>
      <AreaChart data={data} margin={{ top: 8, right: 8, left: -12, bottom: 0 }}>
        <defs>
          <linearGradient id="salesFill" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="var(--color-primary)" stopOpacity={0.28} />
            <stop offset="100%" stopColor="var(--color-primary)" stopOpacity={0.02} />
          </linearGradient>
          <linearGradient id="profitFill" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="var(--color-success)" stopOpacity={0.2} />
            <stop offset="100%" stopColor="var(--color-success)" stopOpacity={0.02} />
          </linearGradient>
        </defs>
        <CartesianGrid vertical={false} stroke="var(--border-default)" strokeDasharray="4 5" opacity={0.65} />
        <XAxis dataKey="name" axisLine={false} tickLine={false} tick={{ fill: 'var(--text-secondary)', fontSize: 11 }} />
        <YAxis axisLine={false} tickLine={false} tick={{ fill: 'var(--text-secondary)', fontSize: 11 }} />
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
            name === 'sales' ? 'المبيعات' : 'الأرباح',
          ]}
        />
        <Legend
          verticalAlign="top"
          align="right"
          iconType="circle"
          wrapperStyle={{ paddingBottom: '12px', fontSize: '11px', color: 'var(--text-secondary)' }}
          formatter={(value) => value === 'sales' ? 'المبيعات' : 'الأرباح'}
        />
        <Area
          type="monotone"
          dataKey="sales"
          stroke="var(--color-primary)"
          strokeWidth={2.5}
          fill="url(#salesFill)"
          dot={{ r: 3, fill: 'var(--color-primary)', strokeWidth: 0 }}
          activeDot={{ r: 5, strokeWidth: 2, stroke: 'var(--bg-surface)' }}
          name="sales"
        />
        <Area
          type="monotone"
          dataKey="profit"
          stroke="var(--color-success)"
          strokeWidth={2.5}
          fill="url(#profitFill)"
          dot={{ r: 3, fill: 'var(--color-success)', strokeWidth: 0 }}
          activeDot={{ r: 5, strokeWidth: 2, stroke: 'var(--bg-surface)' }}
          name="profit"
        />
      </AreaChart>
    </ResponsiveContainer>
  );
}