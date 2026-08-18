import {
  LineChart,
  Line,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'

interface SalesChartData {
  date: string
  sales: number
  revenue: number
}

interface SalesChartProps {
  data: SalesChartData[]
  type?: 'line' | 'bar'
  className?: string
}

export function SalesChart({ data, type = 'line', className }: SalesChartProps) {
  const formattedData = data.map(item => ({
    ...item,
    date: new Date(item.date).toLocaleDateString('ar-SA', { month: 'short', day: 'numeric' }),
  }))

  if (type === 'line') {
    return (
      <ResponsiveContainer width="100%" height={300} className={className}>
        <LineChart data={formattedData}>
          <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
          <XAxis dataKey="date" className="text-xs text-muted" />
          <YAxis className="text-xs text-muted" />
          <Tooltip
            contentStyle={{
              backgroundColor: 'var(--color-surface)',
              border: '1px solid var(--color-border)',
              borderRadius: '8px',
            }}
          />
          <Legend />
          <Line
            type="monotone"
            dataKey="sales"
            stroke="var(--color-primary)"
            strokeWidth={2}
            name="المبيعات"
          />
          <Line
            type="monotone"
            dataKey="revenue"
            stroke="var(--color-success)"
            strokeWidth={2}
            name="الإيرادات"
          />
        </LineChart>
      </ResponsiveContainer>
    )
  }

  return (
    <ResponsiveContainer width="100%" height={300} className={className}>
      <BarChart data={formattedData}>
        <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
        <XAxis dataKey="date" className="text-xs text-muted" />
        <YAxis className="text-xs text-muted" />
        <Tooltip
          contentStyle={{
            backgroundColor: 'var(--color-surface)',
            border: '1px solid var(--color-border)',
            borderRadius: '8px',
          }}
        />
        <Legend />
        <Bar dataKey="sales" fill="var(--color-primary)" name="المبيعات" />
        <Bar dataKey="revenue" fill="var(--color-success)" name="الإيرادات" />
      </BarChart>
    </ResponsiveContainer>
  )
}
