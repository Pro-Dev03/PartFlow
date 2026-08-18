import * as React from "react"
import { Cpu, HardDrive, Monitor, Zap, Package, Scan } from 'lucide-react'

interface PartCardProps {
  name: string
  badges: Array<{ type: 'new' | 'used' | 'stock-ok' | 'stock-low'; label: string }>
  barcode: string
  price: number
  currency?: string
  lowStock?: boolean
  onSell?: () => void
  icon?: 'cpu' | 'ram' | 'gpu' | 'storage' | 'charger' | 'ssd'
}

export function PartCard({ 
  name, 
  badges, 
  barcode, 
  price, 
  currency = 'د.أ', 
  lowStock = false, 
  onSell,
  icon = 'cpu'
}: PartCardProps) {
  const iconMap = {
    cpu: <Cpu className="w-[22px] h-[22px]" />,
    ram: <Package className="w-[22px] h-[22px]" />,
    gpu: <Monitor className="w-[22px] h-[22px]" />,
    storage: <HardDrive className="w-[22px] h-[22px]" />,
    charger: <Zap className="w-[22px] h-[22px]" />,
    ssd: <HardDrive className="w-[22px] h-[22px]" />,
  }

  const badgeStyles = {
    new: 'bg-blue-dim text-blue',
    used: 'bg-warn-dim text-warn',
    'stock-ok': 'bg-accent-dim text-accent',
    'stock-low': 'bg-danger-dim text-danger',
  }

  return (
    <div 
      className={`bg-surface border border-border rounded-[8px] p-4 relative transition-all duration-150 hover:border-[#33404f] hover:-translate-y-0.5 ${
        lowStock ? 'border-[rgba(255,92,92,0.4)]' : ''
      }`}
    >
      <div className="w-[44px] h-[44px] rounded-[5px] bg-surface-higher flex items-center justify-center mb-3 text-text-dim">
        {iconMap[icon]}
      </div>
      
      <h3 className="text-[14px] font-semibold mb-2 leading-relaxed text-text">{name}</h3>
      
      <div className="flex gap-1.5 mb-3 flex-wrap">
        {badges.map((badge, index) => (
          <span 
            key={index}
            className={`text-[10.5px] px-2 py-0.5 rounded-full font-semibold flex items-center gap-1 ${badgeStyles[badge.type]}`}
          >
            {badge.label}
          </span>
        ))}
      </div>
      
      <div className="flex items-center gap-1.5 mb-3.5 text-text-faint text-[11px] font-mono">
        <Scan className="w-[14px] h-[14px] flex-shrink-0" />
        {barcode}
      </div>
      
      <div className="flex items-center justify-between">
        <div className="font-mono text-[17px] font-bold text-text">
          {price}
          <span className="text-[11px] text-text-faint mr-1">{currency}</span>
        </div>
        <button 
          onClick={onSell}
          className="bg-accent text-[#04140F] border-none px-4 py-2 rounded-[5px] text-[12.5px] font-semibold cursor-pointer font-inherit transition-all hover:bg-accent-hover"
        >
          بيع
        </button>
      </div>
    </div>
  )
}