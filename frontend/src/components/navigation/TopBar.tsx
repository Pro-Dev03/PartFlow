import { useState } from 'react'
import { ThemeToggle } from '../theme/ThemeToggle'
import { Search, Bell, User } from 'lucide-react'

export function TopBar() {
  const [searchQuery, setSearchQuery] = useState('')

  return (
    <header
      className="bg-surface dark:bg-surface-dark border-b border-border dark:border-border-dark"
      style={{
        height: '70px',
        padding: '0 32px',
        direction: 'rtl',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between'
      }}
    >
      {/* Left Side - Title */}
      <div style={{ display: 'flex', alignItems: 'center' }}>
        <h1 className="text-text dark:text-text-dark" style={{ fontSize: '22px', fontWeight: 700, margin: 0 }}>
          مرحباً، أبو خالد 👋
        </h1>
        <div className="text-text-secondary dark:text-text-dim" style={{ fontSize: '13px', marginTop: '4px' }}>
          الثلاثاء، ١٨ أغسطس ٢٠٢٦
        </div>
      </div>

      {/* Right Side - Search & Actions */}
      <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
        {/* Search Bar */}
        <div className="flex items-center gap-2 bg-surface dark:bg-surface-dark border border-border dark:border-border-dark rounded-[8px] px-4 py-2" style={{ width: '320px' }}>
          <Search className="w-4 h-4 text-text-tertiary dark:text-text-faint" />
          <input
            type="text"
            placeholder="ابحث بالاسم أو الباركود..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="bg-transparent border-none outline-none text-text dark:text-text-dark text-sm w-full placeholder:text-text-tertiary dark:placeholder:text-text-faint"
          />
        </div>

        {/* Notification */}
        <button
          className="hover:bg-surface-elevated dark:hover:bg-surface-elevated rounded transition-colors p-2"
          style={{
            background: 'transparent',
            border: 'none',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center'
          }}
        >
          <Bell className="w-5 h-5 text-text-secondary dark:text-text-dim" />
        </button>

        {/* Theme Toggle */}
        <ThemeToggle />

        {/* Avatar */}
        <div 
          style={{
            position: 'relative',
            cursor: 'pointer'
          }}
        >
          <div 
            style={{
              width: '34px',
              height: '34px',
              borderRadius: '50%',
              backgroundColor: '#232A34',
              border: '2px solid #262E39',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              transition: 'all 0.2s'
            }}
          >
            <User className="w-4 h-4 text-text-secondary dark:text-text-dim" />
          </div>
          <div 
            style={{
              position: 'absolute',
              bottom: '0',
              right: '0',
              width: '12px',
              height: '12px',
              borderRadius: '50%',
              backgroundColor: '#00D9A3',
              border: '2px solid #151A21'
            }}
          />
        </div>
      </div>
    </header>
  )
}
