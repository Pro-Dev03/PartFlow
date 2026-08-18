import * as React from "react"
import { useLocation, useNavigate } from 'react-router-dom'
import { Home, Package, ShoppingCart, Users, DollarSign, Settings, AlertCircle, FileText, ShoppingCart as PurchasesIcon, Truck, Shield, RefreshCw, Upload } from 'lucide-react'

interface NavItem {
  icon: React.ReactNode
  label: string
  path: string
  dangerDot?: boolean
}

export function Sidebar() {
  const location = useLocation()
  const navigate = useNavigate()
  
  const navItems: NavItem[] = [
    { icon: <Home className="w-[21px] h-[21px]" />, label: 'الرئيسية', path: '/' },
    { icon: <Package className="w-[21px] h-[21px]" />, label: 'المخزون', path: '/inventory' },
    { icon: <ShoppingCart className="w-[21px] h-[21px]" />, label: 'المبيعات', path: '/sales' },
    { icon: <Users className="w-[21px] h-[21px]" />, label: 'العملاء', path: '/customers' },
    { icon: <DollarSign className="w-[21px] h-[21px]" />, label: 'الديون', path: '/debts', dangerDot: true },
    { icon: <FileText className="w-[21px] h-[21px]" />, label: 'التقارير', path: '/reports' },
    { icon: <PurchasesIcon className="w-[21px] h-[21px]" />, label: 'المشتريات', path: '/purchases' },
    { icon: <Truck className="w-[21px] h-[21px]" />, label: 'الموردون', path: '/suppliers' },
    { icon: <Shield className="w-[21px] h-[21px]" />, label: 'الضمانات', path: '/warranties' },
    { icon: <Settings className="w-[21px] h-[21px]" />, label: 'الإعدادات', path: '/settings' },
  ]

  const isActive = (path: string) => {
    if (path === '/') {
      return location.pathname === '/' || location.pathname === '/dashboard'
    }
    return location.pathname.startsWith(path)
  }

  const handleNavClick = (path: string) => {
    navigate(path)
  }

  return (
    <aside className="bg-surface border-e border-border flex flex-col items-center py-[22px] gap-1.5 hidden md:flex w-[52px]">
      {/* Logo */}
      <div className="w-[40px] h-[40px] bg-accent rounded-[5px] flex items-center justify-center mb-7 text-[#04140F] font-bold font-mono text-[15px]">
        PF
      </div>

      {/* Navigation */}
      {navItems.map((item, index) => (
        <button
          key={index}
          onClick={() => handleNavClick(item.path)}
          className={`w-[52px] h-[52px] flex items-center justify-center rounded-[8px] text-text-faint cursor-pointer relative transition-all duration-150 mb-0.5 ${
            isActive(item.path) 
              ? 'bg-accent-dim text-accent' 
              : 'hover:bg-surface-elevated hover:text-text-dim'
          } ${item.dangerDot ? 'danger-dot' : ''}`}
          title={item.label}
        >
          {item.icon}
          {isActive(item.path) && (
            <div className="absolute -start-[11px] top-1/2 -translate-y-1/2 w-[3px] h-[20px] bg-accent rounded-[2px]" />
          )}
        </button>
      ))}

      <div className="flex-1" />

      {/* Bottom actions */}
      <div className="flex flex-col gap-1.5">
        <button 
          onClick={() => handleNavClick('/barcode')}
          className="w-[52px] h-[52px] flex items-center justify-center rounded-[8px] text-text-faint cursor-pointer hover:bg-surface-elevated hover:text-text-dim transition-all duration-150"
          title="مسح باركود"
        >
          <RefreshCw className="w-[21px] h-[21px]" />
        </button>
        <button 
          onClick={() => handleNavClick('/import-export')}
          className="w-[52px] h-[52px] flex items-center justify-center rounded-[8px] text-text-faint cursor-pointer hover:bg-surface-elevated hover:text-text-dim transition-all duration-150"
          title="استيراد/تصدير"
        >
          <Upload className="w-[21px] h-[21px]" />
        </button>
        <button 
          onClick={() => handleNavClick('/audit')}
          className="w-[52px] h-[52px] flex items-center justify-center rounded-[8px] text-text-faint cursor-pointer hover:bg-surface-elevated hover:text-text-dim transition-all duration-150"
          title="سجل التدقيق"
        >
          <AlertCircle className="w-[21px] h-[21px]" />
        </button>
      </div>
    </aside>
  )
}