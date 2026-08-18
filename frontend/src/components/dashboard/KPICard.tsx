import * as React from "react"

interface KPICardProps {
  label: string
  value: string | number
  sub?: string
  trend?: string
  variant?: 'default' | 'warn' | 'danger'
  currency?: string
}

export function KPICard({ label, value, sub, trend, variant = 'default', currency = '' }: KPICardProps) {
  const variantStyles = {
    default: 'text-text',
    warn: 'text-warn',
    danger: 'text-danger',
  }

  return (
    <div className="bg-surface border border-border rounded-[8px] p-[18px_18px_16px]">
      <div className="flex items-center justify-between mb-2.5">
        <span className="text-[12.5px] text-text-dim">{label}</span>
        {trend && (
          <span className="text-[11px] text-accent font-mono">{trend}</span>
        )}
      </div>
      <div className={`font-mono text-[26px] font-semibold ${variantStyles[variant]}`}>
        {value}
        {currency && (
          <span className="text-[13px] text-text-faint ml-1">{currency}</span>
        )}
      </div>
      {sub && (
        <div className="text-[11.5px] text-text-faint mt-1.5">{sub}</div>
      )}
    </div>
  )
}