import { Card, CardContent, CardHeader, CardTitle } from './card';

interface SimpleBarChartProps {
  data: { label: string; value: number }[];
  title?: string;
  color?: string;
  loading?: boolean;
}

export function SimpleBarChart({ data, title, color = '#14b8a6', loading }: SimpleBarChartProps) {
  const maxValue = data.length > 0 ? Math.max(...data.map(d => d.value), 1) : 1;
  
  return (
    <Card>
      {title && (
        <CardHeader>
          <CardTitle>{title}</CardTitle>
        </CardHeader>
      )}
      <CardContent>
        {loading ? (
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '120px' }}>
            <div style={{ animation: 'spin 1s linear infinite', borderRadius: '50%', height: '32px', width: '32px', borderBottom: '2px solid #14b8a6' }} />
          </div>
        ) : data.length === 0 ? (
          <div style={{ padding: '24px', textAlign: 'center', color: 'var(--text-muted)' }}>
            لا توجد بيانات فعلية للفترة المحددة
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
            {data.map((item, index) => {
              const percentage = (item.value / maxValue) * 100;
              return (
                <div key={index} style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                  <div style={{ minWidth: '80px', fontSize: '12px', color: 'var(--text-secondary)' }}>
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
                  <div style={{ minWidth: '50px', fontSize: '12px', fontWeight: '600', color: 'var(--text-primary)', textAlign: 'left' }}>
                    {item.value.toLocaleString()}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

interface SimpleLineChartProps {
  data: { label: string; value: number }[];
  title?: string;
  color?: string;
  loading?: boolean;
  className?: string;
}

export function SimpleLineChart({ data, title, color = '#14b8a6', loading, className }: SimpleLineChartProps) {
  const safeData = data.map(item => ({ ...item, value: Number.isFinite(Number(item.value)) ? Number(item.value) : 0 }));
  const maxValue = safeData.length > 0 ? Math.max(...safeData.map(d => d.value), 1) : 1;
  const totalValue = safeData.reduce((sum, item) => sum + item.value, 0);
  const chartWidth = 760;
  const chartHeight = 250;
  const leftPadding = 58;
  const rightPadding = 18;
  const topPadding = 20;
  const bottomPadding = 38;
  const plotWidth = chartWidth - leftPadding - rightPadding;
  const plotHeight = chartHeight - topPadding - bottomPadding;
  const coordinates = safeData.map((item, index) => ({
    ...item,
    x: leftPadding + (safeData.length > 1 ? (index / (safeData.length - 1)) * plotWidth : plotWidth / 2),
    y: topPadding + plotHeight - (item.value / maxValue) * plotHeight,
  }));
  const points = coordinates.map(({ x, y }) => `${x},${y}`).join(' ');
  const areaPoints = `${leftPadding},${topPadding + plotHeight} ${points} ${leftPadding + plotWidth},${topPadding + plotHeight}`;
  const formatValue = (value: number) => value.toLocaleString('ar-SA', { maximumFractionDigits: 0 });

  return (
    <Card className={className}>
      {title && (
        <CardHeader>
          <CardTitle>{title}</CardTitle>
        </CardHeader>
      )}
      <CardContent>
        {loading ? (
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '200px' }}>
            <div style={{ animation: 'spin 1s linear infinite', borderRadius: '50%', height: '32px', width: '32px', borderBottom: '2px solid #14b8a6' }} />
          </div>
        ) : safeData.length === 0 ? (
          <div style={{ padding: '24px', textAlign: 'center', color: 'var(--text-muted)' }}>
            لا توجد بيانات فعلية للفترة المحددة
          </div>
        ) : (
          <div className="rounded-lg border border-border bg-surface-elevated px-3 py-2">
            <div className="mb-3 flex items-center justify-between gap-3">
              <div>
                <p className="text-xs text-text-muted">إجمالي الإيراد في الفترة</p>
                <p className="text-lg font-bold text-text-primary">₪{formatValue(totalValue)}</p>
              </div>
              <div className="rounded-lg border border-border bg-surface-elevated px-3 py-2 text-left">
                <p className="text-xs text-text-muted">أعلى يوم</p>
                <p className="text-sm font-semibold" style={{ color }}>₪{formatValue(maxValue)}</p>
              </div>
            </div>
            <div style={{ position: 'relative', height: '250px', width: '100%' }}>
            <svg
              viewBox={`0 0 ${chartWidth} ${chartHeight}`}
              role="img"
              aria-label={title || 'اتجاه المبيعات'}
              style={{ width: '100%', height: '100%', overflow: 'visible' }}
            >
              <defs>
                <linearGradient id="sales-area-gradient" x1="0" x2="0" y1="0" y2="1">
                  <stop offset="0%" stopColor={color} stopOpacity="0.25" />
                  <stop offset="100%" stopColor={color} stopOpacity="0.02" />
                </linearGradient>
              </defs>
              {[0, 1, 2, 3, 4].map((step) => {
                const y = topPadding + (step / 4) * plotHeight;
                const value = maxValue - (step / 4) * maxValue;
                return (
                  <g key={step}>
                    <line x1={leftPadding} x2={leftPadding + plotWidth} y1={y} y2={y} stroke="var(--border-subtle)" strokeWidth="1" />
                    <text x={leftPadding - 10} y={y + 4} textAnchor="end" fontSize="14" fill="var(--text-muted)">
                      {formatValue(value)}
                    </text>
                  </g>
                );
              })}
              <polygon points={areaPoints} fill="url(#sales-area-gradient)" />
              <polyline
                points={points}
                fill="none"
                stroke={color}
                strokeWidth="3"
                strokeLinecap="round"
                strokeLinejoin="round"
              />
              {coordinates.map((item, index) => {
                return (
                  <g key={index}>
                    <circle cx={item.x} cy={item.y} r="6" fill="var(--bg-surface)" stroke={color} strokeWidth="3" />
                    <text x={item.x} y={item.y - 14} textAnchor="middle" fontSize="14" fontWeight="600" fill="var(--text-primary)">
                      {formatValue(item.value)}
                    </text>
                  </g>
                );
              })}
            </svg>
            <div className="mt-1 flex justify-between pl-14 text-[11px] text-text-muted">
              {coordinates.map((item, index) => (
                <div key={index} className="max-w-20 truncate text-center">{item.label}</div>
              ))}
            </div>
          </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

interface SimplePieChartProps {
  data: { label: string; value: number; color: string }[];
  title?: string;
  loading?: boolean;
}

export function SimplePieChart({ data, title, loading }: SimplePieChartProps) {
  const total = data.reduce((sum, item) => sum + item.value, 0);
  const safeTotal = total > 0 ? total : 1;
  let currentAngle = 0;

  const segments = data.map((item) => {
    const percentage = (item.value / safeTotal) * 100;
    const angle = (item.value / safeTotal) * 360;
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
        {loading ? (
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '120px' }}>
            <div style={{ animation: 'spin 1s linear infinite', borderRadius: '50%', height: '32px', width: '32px', borderBottom: '2px solid #14b8a6' }} />
          </div>
        ) : segments.length === 0 ? (
          <div style={{ padding: '24px', textAlign: 'center', color: 'var(--text-muted)' }}>
            لا توجد بيانات فعلية للفترة المحددة
          </div>
        ) : (
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
                      stroke="var(--bg-border)"
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
                    <span style={{ color: 'var(--text-secondary)' }}>{segment.label}</span>
                    <span style={{ color: 'var(--text-primary)', fontWeight: '600' }}>{segment.percentage.toFixed(1)}%</span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}