import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'

interface ProfitLossChartData {
  month: string
  revenue: number
  expenses: number
  profit: number
}

interface ProfitLossChartProps {
  data: ProfitLossChartData[]
  className?: string
}

export function ProfitLossChart({ data, className }: ProfitLossChartProps) {
  const formattedData = data.map(item => ({
    ...item,
    month: new Date(item.month).toLocaleDateString('ar-SA', { month: 'short' }),
  }))

  return (
    <ResponsiveContainer width="100%" height={300} className={className}>
      <BarChart data={formattedData}>
        <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
        <XAxis dataKey="month" className="text-xs text-muted" />
        <YAxis className="text-xs text-muted" />
        <Tooltip
          contentStyle={{
            backgroundColor: 'var(--color-surface)',
            border: '1px solid var(--color-border)',
            borderRadius: '8px',
          }}
        />
        <Legend />
        <Bar dataKey="revenue" fill="var(--color-success)" name="الإيرادات" />
        <Bar dataKey="expenses" fill="var(--color-danger)" name="المصروفات" />
        <Bar dataKey="profit" fill="var(--color-primary)" name="الربح" />
      </BarChart>
    </ResponsiveContainer>
  )
}
