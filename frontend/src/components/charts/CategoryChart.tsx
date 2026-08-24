import {
  PieChart,
  Pie,
  Cell,
  ResponsiveContainer,
  Legend,
  Tooltip,
  DefaultLegendContent,
} from 'recharts';

interface CategoryChartProps {
  data: Array<{
    name: string;
    value: number;
    color: string;
  }>;
}

export function CategoryChart({ data }: CategoryChartProps) {
  const total = data.reduce((sum, item) => sum + item.value, 0);

  return (
    <ResponsiveContainer width="100%" height={300}>
      <PieChart>
        <Pie
          data={data}
          cx="50%"
          cy="50%"
          labelLine={false}
          label={false}
          outerRadius={90}
          innerRadius={50}
          fill="#8884d8"
          dataKey="value"
          strokeWidth={3}
          stroke="var(--bg-surface)"
          animationBegin={0}
          animationDuration={800}
          animationEasing="ease-out"
        >
          {data.map((entry, index) => (
            <Cell
              key={`cell-${index}`}
              fill={entry.color}
              style={{
                filter: 'drop-shadow(0 4px 6px rgba(0,0,0,0.2))',
                transition: 'all 0.3s ease',
              }}
            />
          ))}
        </Pie>
        <Tooltip
          contentStyle={{
            backgroundColor: 'var(--bg-surface)',
            border: '1px solid var(--border-default)',
            borderRadius: '12px',
            padding: '12px',
            boxShadow: '0 8px 32px rgba(0,0,0,0.3)',
            backdropFilter: 'blur(10px)',
          }}
          itemStyle={{
            color: 'var(--text-primary)',
            fontSize: '13px',
            fontWeight: '500',
          }}
          labelStyle={{
            color: 'var(--text-secondary)',
            fontSize: '12px',
            marginBottom: '4px',
          }}
          formatter={(value: any, name: any) => [`${value}%`, name]}
        />
        <Legend
          verticalAlign="bottom"
          height={50}
          iconType="circle"
          wrapperStyle={{
            paddingTop: '12px',
            fontSize: '13px',
            color: 'var(--text-primary)',
            fontWeight: '500',
          }}
          content={
            <DefaultLegendContent
              labelStyle={{
                padding: '4px 12px',
                cursor: 'pointer',
                transition: 'all 0.2s ease',
                display: 'flex',
                alignItems: 'center',
              }}
              payload={data.map((item) => ({
                value: item.name,
                type: 'circle',
                color: item.color,
                payload: {
                  ...item,
                  percentage:
                    total > 0 ? `${Math.round((item.value / total) * 100)}%` : '0%',
                },
              }))}
              formatter={(value, entry) => (
                <span
                  style={{
                    color: 'var(--text-primary)',
                    fontSize: '13px',
                    fontWeight: '500',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '6px',
                  }}
                >
                  <span style={{ color: entry.color, fontWeight: '600' }}>
                    {(entry.payload as any)?.percentage || '0%'}
                  </span>
                  <span style={{ color: 'var(--text-secondary)' }}>{value}</span>
                </span>
              )}
            />
          }
        />
      </PieChart>
    </ResponsiveContainer>
  );
}
