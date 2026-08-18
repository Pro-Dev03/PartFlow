import {
  PieChart,
  Pie,
  Cell,
  ResponsiveContainer,
  Tooltip,
  Legend,
} from 'recharts'

interface InventoryChartData {
  category: string
  count: number
  value: number
}

interface InventoryChartProps {
  data: InventoryChartData[]
  metric?: 'count' | 'value'
  className?: string
}

const COLORS = [
  'var(--color-primary)',
  'var(--color-success)',
  'var(--color-warning)',
  'var(--color-danger)',
  'var(--color-info)',
  '#8b5cf6',
  '#ec4899',
  '#14b8a6',
]

export function InventoryChart({ data, metric = 'count', className }: InventoryChartProps) {
  const chartData = data.map(item => ({
    name: item.category,
    value: metric === 'count' ? item.count : item.value,
  }))

  return (
    <ResponsiveContainer width="100%" height={300} className={className}>
      <PieChart>
        <Pie
          data={chartData}
          cx="50%"
          cy="50%"
          labelLine={false}
          label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
          outerRadius={80}
          fill="#8884d8"
          dataKey="value"
        >
          {chartData.map((entry, index) => (
            <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
          ))}
        </Pie>
        <Tooltip
          contentStyle={{
            backgroundColor: 'var(--color-surface)',
            border: '1px solid var(--color-border)',
            borderRadius: '8px',
          }}
        />
        <Legend />
      </PieChart>
    </ResponsiveContainer>
  )
}
