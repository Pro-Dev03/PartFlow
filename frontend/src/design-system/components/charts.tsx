import { Card, CardContent, CardHeader, CardTitle } from './card';
import { Area, AreaChart, CartesianGrid, Legend, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';

interface SimpleBarChartProps {
  data: { label: string; value: number }[];
  title?: string;
  color?: string;
  loading?: boolean;
}

export function SimpleBarChart({ data, title, color = '#14b8a6', loading }: SimpleBarChartProps) {
  const maxValue = data.length > 0 ? Math.max(...data.map(d => d.value), 1) : 1;
  
  return (
    <Card className="pf-report-bar-chart">
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
                  <div className="pf-report-chart-track" style={{ flex: 1, height: '8px', background: 'rgba(148, 163, 184, 0.13)', borderRadius: '4px', overflow: 'hidden' }}>
                    <div
                      className="pf-report-chart-bar"
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
  data: { label: string; value: number; secondaryValue?: number }[];
  title?: string;
  valueLabel?: string;
  peakLabel?: string;
  summaryValue?: number;
  secondaryLabel?: string;
  secondarySummaryValue?: number;
  secondaryPeakLabel?: string;
  secondaryColor?: string;
  color?: string;
  loading?: boolean;
  className?: string;
}

export function SimpleLineChart({ data, title, valueLabel = 'إجمالي الفترة', peakLabel = 'أعلى يوم', summaryValue, secondaryLabel, secondarySummaryValue, secondaryPeakLabel, secondaryColor = '#2563eb', color = '#0f8f6d', loading, className }: SimpleLineChartProps) {
  const safeData = data.map(item => ({
    ...item,
    value: Number.isFinite(Number(item.value)) ? Number(item.value) : 0,
    secondaryValue: item.secondaryValue === undefined ? undefined : Number.isFinite(Number(item.secondaryValue)) ? Number(item.secondaryValue) : 0,
  }));
  const maxValue = safeData.length > 0 ? Math.max(...safeData.map(item => item.value)) : 0;
  const maxItem = safeData.reduce((highest, item) => item.value > highest.value ? item : highest, safeData[0] || { label: '', value: 0 });
  const maxSecondaryItem = safeData.reduce((highest, item) => (item.secondaryValue ?? 0) > (highest.secondaryValue ?? 0) ? item : highest, safeData[0] || { label: '', value: 0, secondaryValue: 0 });
  const totalValue = safeData.reduce((sum, item) => sum + item.value, 0);
  const totalSecondaryValue = safeData.reduce((sum, item) => sum + (item.secondaryValue ?? 0), 0);
  const currency = '\u20AA';
  const formatValue = (value: number) => Number(value || 0).toLocaleString('ar-SA', { maximumFractionDigits: 0 });
  const formatAxisValue = (value: number) => `${currency}${new Intl.NumberFormat('ar-SA', { notation: 'compact', maximumFractionDigits: 1 }).format(value)}`;
  const formatAxisLabel = (label: string) => {
    const parts = String(label || '').replace(/[\u061c\u200e\u200f]/g, '').match(/[\p{N}]+/gu);
    if (parts?.length !== 3) return label;
    return parts[0].length === 4 ? `${parts[2]}/${parts[1]}` : `${parts[0]}/${parts[1]}`;
  };
  const xAxisInterval = Math.max(0, Math.ceil(safeData.length / 6) - 1);
  const summary = summaryValue ?? totalValue;
  const secondarySummary = secondarySummaryValue ?? totalSecondaryValue;

  return (
    <Card className={`pf-report-line-chart ${className || ''}`}>
      {title && (
        <CardHeader className="pf-report-line-chart-header">
          <CardTitle>{title}</CardTitle>
          <span className="pf-report-line-chart-unit">{currency}</span>
        </CardHeader>
      )}
      <CardContent className="pf-report-line-chart-content">
        {loading ? (
          <div className="pf-report-line-chart-state" role="status">
            <div className="pf-report-chart-spinner" />
          </div>
        ) : safeData.length === 0 ? (
          <div className="pf-report-line-chart-state">لا توجد بيانات فعلية للفترة المحددة</div>
        ) : (
          <>
            <div className={`pf-report-line-chart-summary ${secondaryLabel ? 'has-secondary' : ''}`}>
              <div className="pf-report-line-chart-metric primary">
                <span className="pf-report-line-chart-marker" style={{ background: color }} />
                <div>
                  <p>{valueLabel}</p>
                  <strong>{currency}{formatValue(summary)}</strong>
                </div>
              </div>
              <div className="pf-report-line-chart-metric">
                <div>
                  <p>{peakLabel}</p>
                  <strong>{maxValue !== 0 ? maxItem.label : '—'} <span>{currency}{formatValue(maxValue)}</span></strong>
                </div>
              </div>
              {secondaryLabel && (
                <div className="pf-report-line-chart-metric secondary">
                  <span className="pf-report-line-chart-marker" style={{ background: secondaryColor }} />
                  <div>
                    <p>{secondaryLabel}</p>
                    <strong>{currency}{formatValue(secondarySummary)}</strong>
                    <small>{secondaryPeakLabel || secondaryLabel}: {(maxSecondaryItem.secondaryValue ?? 0) !== 0 ? maxSecondaryItem.label : '—'} · {currency}{formatValue(maxSecondaryItem.secondaryValue ?? 0)}</small>
                  </div>
                </div>
              )}
            </div>
            <div className="pf-report-line-chart-plot" role="img" aria-label={title || 'Sales trend'}>
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={safeData} margin={{ top: 16, right: 18, left: 8, bottom: 4 }}>
                  <defs>
                    <linearGradient id="pf-report-sales-area" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="0%" stopColor={color} stopOpacity={0.2} />
                      <stop offset="95%" stopColor={color} stopOpacity={0.015} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid vertical={false} stroke="var(--border-default)" strokeDasharray="3 5" />
                  <XAxis
                    dataKey="label"
                    axisLine={false}
                    tickLine={false}
                    tickMargin={12}
                    interval={xAxisInterval}
                    minTickGap={22}
                    tick={{ fill: 'var(--text-muted)', fontSize: 11 }}
                    tickFormatter={formatAxisLabel}
                  />
                  <YAxis
                    width={66}
                    axisLine={false}
                    tickLine={false}
                    tickMargin={8}
                    tickCount={5}
                    domain={['auto', 'auto']}
                    tick={{ fill: 'var(--text-muted)', fontSize: 11 }}
                    tickFormatter={formatAxisValue}
                  />
                  <Tooltip
                    contentStyle={{
                      background: 'var(--bg-surface-elevated)',
                      border: '1px solid var(--border-default)',
                      borderRadius: '12px',
                      boxShadow: 'var(--shadow-md)',
                      color: 'var(--text-primary)',
                    }}
                    labelStyle={{ color: 'var(--text-secondary)', marginBottom: 5 }}
                    formatter={(value) => `${currency}${formatValue(Number(value ?? 0))}`}
                  />
                  <Area
                    type="monotone"
                    dataKey="value"
                    name={valueLabel}
                    stroke={color}
                    strokeWidth={3}
                    fill="url(#pf-report-sales-area)"
                    dot={false}
                    activeDot={{ r: 5, fill: color, stroke: 'var(--bg-surface)', strokeWidth: 2 }}
                    isAnimationActive={false}
                  />
                  {secondaryLabel && (
                    <Line
                      type="monotone"
                      dataKey="secondaryValue"
                      name={secondaryLabel}
                      stroke={secondaryColor}
                      strokeWidth={2.5}
                      dot={false}
                      activeDot={{ r: 5, fill: secondaryColor, stroke: 'var(--bg-surface)', strokeWidth: 2 }}
                      isAnimationActive={false}
                    />
                  )}
                  {secondaryLabel && (
                    <Legend
                      verticalAlign="bottom"
                      align="center"
                      iconType="circle"
                      wrapperStyle={{ paddingTop: 8, fontSize: 11, color: 'var(--text-secondary)' }}
                    />
                  )}
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </>
        )}
      </CardContent>
    </Card>
  );
}

interface ProfitTrendChartProps {
  data: { label: string; value: number }[];
  title?: string;
  valueLabel?: string;
  peakLabel?: string;
  loading?: boolean;
}

export function ProfitTrendChart({ data, title = 'اتجاه صافي الربح', valueLabel = 'إجمالي صافي الربح', peakLabel = 'أعلى شهر', loading }: ProfitTrendChartProps) {
  const safeData = data.map(item => ({ label: item.label, value: Number.isFinite(Number(item.value)) ? Number(item.value) : 0 }));
  const totalValue = safeData.reduce((sum, item) => sum + item.value, 0);
  const maxItem = safeData.reduce((highest, item) => item.value > highest.value ? item : highest, safeData[0] || { label: '', value: 0 });
  const formatValue = (value: number) => `₪${Number(value).toLocaleString('ar-SA', { maximumFractionDigits: 0 })}`;

  return (
    <Card className="profit-trend-chart">
      {title && (
        <CardHeader>
          <CardTitle>{title}</CardTitle>
        </CardHeader>
      )}
      <CardContent>
        {loading ? (
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '270px' }}>
            <div style={{ animation: 'spin 1s linear infinite', borderRadius: '50%', height: '32px', width: '32px', borderBottom: '2px solid var(--color-success)' }} />
          </div>
        ) : safeData.length === 0 ? (
          <div style={{ padding: '24px', textAlign: 'center', color: 'var(--text-muted)' }}>
            لا توجد بيانات فعلية للفترة المحددة
          </div>
        ) : (
          <>
            <div className="mb-3 flex items-center justify-between gap-3">
              <div>
                <p className="text-xs text-text-muted">{valueLabel}</p>
                <p className="text-lg font-bold text-text-primary">{formatValue(totalValue)}</p>
              </div>
              <div className="rounded-lg border border-border bg-surface-elevated px-3 py-2 text-left">
                <p className="text-xs text-text-muted">{peakLabel}: {maxItem.label}</p>
                <p className="text-sm font-semibold text-success">{formatValue(maxItem.value)}</p>
              </div>
            </div>
            <div style={{ width: '100%', height: '270px', direction: 'ltr' }}>
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={safeData} margin={{ top: 8, right: 8, left: 4, bottom: 4 }}>
                  <CartesianGrid vertical={false} stroke="var(--border-default)" strokeDasharray="4 5" opacity={0.65} />
                  <XAxis dataKey="label" axisLine={false} tickLine={false} tick={{ fill: 'var(--text-secondary)', fontSize: 11 }} />
                  <YAxis
                    width={58}
                    axisLine={false}
                    tickLine={false}
                    tick={{ fill: 'var(--text-secondary)', fontSize: 11 }}
                    tickFormatter={(value: number) => formatValue(value)}
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
                    formatter={(value: number | undefined) => [formatValue(value ?? 0), 'صافي الربح']}
                  />
                  <Legend
                    verticalAlign="bottom"
                    align="right"
                    iconType="circle"
                    wrapperStyle={{ paddingTop: '12px', fontSize: '11px', color: 'var(--text-secondary)' }}
                    formatter={() => 'صافي الربح'}
                  />
                  <Line
                    type="monotone"
                    dataKey="value"
                    name="netProfit"
                    stroke="var(--color-success)"
                    strokeWidth={3}
                    dot={{ r: 4, fill: 'var(--color-success)', strokeWidth: 0 }}
                    activeDot={{ r: 6, strokeWidth: 2, stroke: 'var(--bg-surface)' }}
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </>
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

  const exactPercentages = data.map((item) => (item.value / safeTotal) * 100);
  const roundedPercentages = exactPercentages.map((percentage) => Math.floor(percentage * 10) / 10);
  const percentageRemainder = Math.round((100 - roundedPercentages.reduce((sum, percentage) => sum + percentage, 0)) * 10);
  if (percentageRemainder > 0 && roundedPercentages.length > 0) {
    const largestRemainderIndex = exactPercentages.reduce((bestIndex, percentage, index) => (
      percentage - roundedPercentages[index] > exactPercentages[bestIndex] - roundedPercentages[bestIndex]
        ? index
        : bestIndex
    ), 0);
    roundedPercentages[largestRemainderIndex] += percentageRemainder / 10;
  }

  const segments = data.map((item, index) => {
    const angle = (item.value / safeTotal) * 360;
    const segment = {
      ...item,
      displayPercentage: roundedPercentages[index],
      startAngle: currentAngle,
      endAngle: currentAngle + angle
    };
    currentAngle += angle;
    return segment;
  });

  return (
    <Card className="pf-report-pie-chart">
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
                      className="pf-report-pie-segment"
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
                <div className="pf-report-pie-legend-row" key={index} style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '8px' }}>
                  <div className="pf-report-pie-swatch" style={{ width: '12px', height: '12px', borderRadius: '2px', background: segment.color }} />
                  <div style={{ flex: 1, display: 'flex', justifyContent: 'space-between', fontSize: '12px' }}>
                    <span style={{ color: 'var(--text-secondary)' }}>{segment.label}</span>
                    <span style={{ color: 'var(--text-primary)', fontWeight: '600' }}>
                      ₪{segment.value.toLocaleString('ar-SA', { maximumFractionDigits: 2 })} ({segment.displayPercentage.toFixed(1)}%)
                    </span>
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
