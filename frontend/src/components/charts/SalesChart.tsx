import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

interface SalesChartProps {
  data: Array<{
    name: string;
    sales: number;
    profit: number;
  }>;
}

export function SalesChart({ data }: SalesChartProps) {
  return (
    <ResponsiveContainer width="100%" height={300}>
      <LineChart data={data}>
        <CartesianGrid strokeDasharray="3 3" className="stroke-gray-200 dark:stroke-gray-700" />
        <XAxis 
          dataKey="name" 
          className="text-gray-600 dark:text-gray-400"
          stroke="currentColor"
        />
        <YAxis 
          className="text-gray-600 dark:text-gray-400"
          stroke="currentColor"
        />
        <Tooltip 
          contentStyle={{
            backgroundColor: 'hsl(var(--card))',
            border: '1px solid hsl(var(--border))',
            borderRadius: '0.5rem',
          }}
          itemStyle={{ color: 'hsl(var(--foreground))' }}
        />
        <Line 
          type="monotone" 
          dataKey="sales" 
          stroke="hsl(var(--primary))" 
          strokeWidth={2}
          name="المبيعات"
        />
        <Line 
          type="monotone" 
          dataKey="profit" 
          stroke="hsl(var(--secondary))" 
          strokeWidth={2}
          name="الأرباح"
        />
      </LineChart>
    </ResponsiveContainer>
  );
}