import { Card, CardContent, CardHeader, CardTitle } from './card';

interface SimpleBarChartProps {
  data: { label: string; value: number }[];
  title?: string;
  color?: string;
}

export function SimpleBarChart({ data, title, color = '#22d3ee' }: SimpleBarChartProps) {
  const maxValue = Math.max(...data.map(d => d.value));
  
  return (
    <Card>
      {title && (
        <CardHeader>
          <CardTitle>{title}</CardTitle>
        </CardHeader>
      )}
      <CardContent>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
          {data.map((item, index) => {
            const percentage = (item.value / maxValue) * 100;
            return (
              <div key={index} style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                <div style={{ minWidth: '80px', fontSize: '12px', color: '#94a3b8' }}>
                  {item.label}
                </div>
                <div style={{ flex: 1, height: '8px', background: 'rgba(148, 163, 184, 0.13)', borderRadius: '4px', overflow: 'hidden' }}>
                  <div
                    style={{
                      width: `${percentage}%`,
                      height: '100%',
                      background: color,
                      borderRadius: '4px',
                      transition: 'width 0.3s ease'
                    }}
                  />
                </div>
                <div style={{ minWidth: '50px', fontSize: '12px', fontWeight: '600', color: '#f1f7ff', textAlign: 'left' }}>
                  {item.value.toLocaleString()}
                </div>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}

interface SimpleLineChartProps {
  data: { label: string; value: number }[];
  title?: string;
  color?: string;
}

export function SimpleLineChart({ data, title, color = '#22d3ee' }: SimpleLineChartProps) {
  const maxValue = Math.max(...data.map(d => d.value));
  const points = data.map((item, index) => {
    const x = (index / (data.length - 1)) * 100;
    const y = 100 - ((item.value / maxValue) * 100);
    return `${x},${y}`;
  }).join(' ');

  return (
    <Card>
      {title && (
        <CardHeader>
          <CardTitle>{title}</CardTitle>
        </CardHeader>
      )}
      <CardContent>
        <div style={{ position: 'relative', height: '200px' }}>
          <svg
            viewBox="0 0 100 100"
            preserveAspectRatio="none"
            style={{ width: '100%', height: '100%' }}
          >
            <polyline
              points={points}
              fill="none"
              stroke={color}
              strokeWidth="2"
              vectorEffect="non-scaling-stroke"
            />
            {data.map((item, index) => {
              const x = (index / (data.length - 1)) * 100;
              const y = 100 - ((item.value / maxValue) * 100);
              return (
                <circle
                  key={index}
                  cx={x}
                  cy={y}
                  r="3"
                  fill={color}
                />
              );
            })}
          </svg>
          <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '8px', fontSize: '10px', color: '#8290a7' }}>
            {data.map((item, index) => (
              <div key={index}>{item.label}</div>
            ))}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

interface SimplePieChartProps {
  data: { label: string; value: number; color: string }[];
  title?: string;
}

export function SimplePieChart({ data, title }: SimplePieChartProps) {
  const total = data.reduce((sum, item) => sum + item.value, 0);
  let currentAngle = 0;

  const segments = data.map((item) => {
    const percentage = (item.value / total) * 100;
    const angle = (item.value / total) * 360;
    const segment = {
      ...item,
      percentage,
      startAngle: currentAngle,
      endAngle: currentAngle + angle
    };
    currentAngle += angle;
    return segment;
  });

  return (
    <Card>
      {title && (
        <CardHeader>
          <CardTitle>{title}</CardTitle>
        </CardHeader>
      )}
      <CardContent>
        <div style={{ display: 'flex', alignItems: 'center', gap: '20px' }}>
          <div style={{ width: '120px', height: '120px', borderRadius: '50%', position: 'relative' }}>
            <svg viewBox="0 0 100 100" style={{ width: '100%', height: '100%' }}>
              {segments.map((segment, index) => {
                const startRad = (segment.startAngle - 90) * (Math.PI / 180);
                const endRad = (segment.endAngle - 90) * (Math.PI / 180);
                const x1 = 50 + 40 * Math.cos(startRad);
                const y1 = 50 + 40 * Math.sin(startRad);
                const x2 = 50 + 40 * Math.cos(endRad);
                const y2 = 50 + 40 * Math.sin(endRad);
                const largeArcFlag = segment.endAngle - segment.startAngle > 180 ? 1 : 0;
                
                return (
                  <path
                    key={index}
                    d={`M 50 50 L ${x1} ${y1} A 40 40 0 ${largeArcFlag} 1 ${x2} ${y2} Z`}
                    fill={segment.color}
                    stroke="rgba(7, 10, 18, 0.5)"
                    strokeWidth="1"
                  />
                );
              })}
            </svg>
          </div>
          <div style={{ flex: 1 }}>
            {segments.map((segment, index) => (
              <div key={index} style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '8px' }}>
                <div style={{ width: '12px', height: '12px', borderRadius: '2px', background: segment.color }} />
                <div style={{ flex: 1, display: 'flex', justifyContent: 'space-between', fontSize: '12px' }}>
                  <span style={{ color: '#94a3b8' }}>{segment.label}</span>
                  <span style={{ color: '#f1f7ff', fontWeight: '600' }}>{segment.percentage.toFixed(1)}%</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}