import * as React from "react"
import { Home, Package, ShoppingCart, Users, DollarSign, Settings, AlertCircle } from 'lucide-react'

interface NavItem {
  icon: React.ReactNode
  label: string
  active?: boolean
  dangerDot?: boolean
  onClick?: () => void
}

export function Sidebar() {
  const navItems: NavItem[] = [
    { icon: <Home className="w-[21px] h-[21px]" />, label: 'الرئيسية', active: true },
    { icon: <Package className="w-[21px] h-[21px]" />, label: 'المخزون' },
    { icon: <ShoppingCart className="w-[21px] h-[21px]" />, label: 'المبيعات' },
    { icon: <Users className="w-[21px] h-[21px]" />, label: 'العملاء' },
    { icon: <DollarSign className="w-[21px] h-[21px]" />, label: 'الديون', dangerDot: true },
    { icon: <Settings className="w-[21px] h-[21px]" />, label: 'الإعدادات' },
  ]

  return (
    <aside className="bg-surface border-l border-border flex flex-col items-center py-[22px] gap-1.5 hidden md:flex">
      {/* Logo */}
      <div className="w-[40px] h-[40px] bg-accent rounded-[5px] flex items-center justify-center mb-7 text-[#04140F] font-bold font-mono text-[15px]">
        PF
      </div>

      {/* Navigation */}
      {navItems.map((item, index) => (
        <button
          key={index}
          onClick={item.onClick}
          className={`w-[52px] h-[52px] flex items-center justify-center rounded-[8px] text-text-faint cursor-pointer relative transition-all duration-150 mb-0.5 ${
            item.active 
              ? 'bg-accent-dim text-accent' 
              : 'hover:bg-surface-elevated hover:text-text-dim'
          } ${item.dangerDot ? 'danger-dot' : ''}`}
          title={item.label}
        >
          {item.icon}
          {item.active && (
            <div className="absolute -right-[11px] top-1/2 -translate-y-1/2 w-[3px] h-[20px] bg-accent rounded-[2px]" />
          )}
        </button>
      ))}

      <div className="flex-1" />

      {/* Bottom action */}
      <button className="w-[52px] h-[52px] flex items-center justify-center rounded-[8px] text-text-faint cursor-pointer hover:bg-surface-elevated hover:text-text-dim transition-all duration-150">
        <AlertCircle className="w-[21px] h-[21px]" />
      </button>
    </aside>
  )
}