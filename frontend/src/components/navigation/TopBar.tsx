import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { ThemeToggle } from '../theme/ThemeToggle'
import { Search, Bell, User, LogOut, Settings } from 'lucide-react'
import { useUIStore } from '@stores'

export function TopBar() {
  const [searchQuery, setSearchQuery] = useState('')
  const [showUserMenu, setShowUserMenu] = useState(false)
  const [currentDate, setCurrentDate] = useState('')
  const navigate = useNavigate()
  const { user, addNotification } = useUIStore()

  useEffect(() => {
    // Update current date dynamically
    const updateDate = () => {
      const now = new Date()
      const options: Intl.DateTimeFormatOptions = { 
        weekday: 'long', 
        year: 'numeric', 
        month: 'long', 
        day: 'numeric' 
      }
      setCurrentDate(now.toLocaleDateString('ar-SA', options))
    }
    
    updateDate()
    const interval = setInterval(updateDate, 60000) // Update every minute
    
    return () => clearInterval(interval)
  }, [])

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    if (searchQuery.trim()) {
      // Navigate to search results or trigger global search
      navigate(`/search?q=${encodeURIComponent(searchQuery)}`)
      addNotification({
        type: 'info',
        message: `جاري البحث عن: ${searchQuery}`,
        duration: 3000
      })
    }
  }

  const handleNotifications = () => {
    addNotification({
      type: 'info',
      message: 'مركز الإشعارات قيد التطوير',
      duration: 3000
    })
  }

  const handleLogout = () => {
    addNotification({
      type: 'success',
      message: 'تم تسجيل الخروج بنجاح',
      duration: 3000
    })
    navigate('/auth')
    setShowUserMenu(false)
  }

  const handleSettings = () => {
    navigate('/settings')
    setShowUserMenu(false)
  }

  return (
    <header className="h-[70px] px-8 bg-surface dark:bg-surface-dark border-b border-border dark:border-border-dark flex items-center justify-between">
      {/* Left Side - Title */}
      <div className="flex flex-col">
        <h1 className="text-text dark:text-text-dark text-[22px] font-bold m-0">
          مرحباً، {user?.name || 'أبو خالد'} 👋
        </h1>
        <div className="text-text-secondary dark:text-text-dim text-[13px] mt-1">
          {currentDate}
        </div>
      </div>

      {/* Right Side - Search & Actions */}
      <div className="flex items-center gap-4">
        {/* Search Bar */}
        <form onSubmit={handleSearch} className="flex items-center gap-2 bg-surface dark:bg-surface-dark border border-border dark:border-border-dark rounded-lg px-4 py-2 w-[320px]">
          <Search className="w-4 h-4 text-text-tertiary dark:text-text-faint" />
          <input
            type="text"
            placeholder="ابحث بالاسم أو الباركود..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="bg-transparent border-none outline-none text-text dark:text-text-dark text-sm w-full placeholder:text-text-tertiary dark:placeholder:text-text-faint"
          />
        </form>

        {/* Notification */}
        <button
          onClick={handleNotifications}
          className="hover:bg-surface-elevated dark:hover:bg-surface-elevated rounded-lg transition-colors p-2 relative"
        >
          <Bell className="w-5 h-5 text-text-secondary dark:text-text-dim" />
          <span className="absolute top-1 end-1 w-2 h-2 bg-accent rounded-full" />
        </button>

        {/* Theme Toggle */}
        <ThemeToggle />

        {/* Avatar with Dropdown */}
        <div className="relative">
          <button
            onClick={() => setShowUserMenu(!showUserMenu)}
            className="relative cursor-pointer hover:opacity-80 transition-opacity"
          >
            <div className="w-[34px] h-[34px] rounded-full bg-[#232A34] border-2 border-[#262E39] flex items-center justify-center transition-all">
              <User className="w-4 h-4 text-text-secondary dark:text-text-dim" />
            </div>
            <div className="absolute bottom-0 end-0 w-3 h-3 rounded-full bg-[#00D9A3] border-2 border-[#151A21]" />
          </button>

          {/* User Menu Dropdown */}
          {showUserMenu && (
            <div className="absolute top-full start-0 mt-2 w-48 bg-surface dark:bg-surface-dark border border-border dark:border-border-dark rounded-lg shadow-lg z-50">
              <div className="p-3 border-b border-border dark:border-border-dark">
                <p className="text-sm font-medium text-text dark:text-text-dark">
                  {user?.name || 'أبو خالد'}
                </p>
                <p className="text-xs text-text-secondary dark:text-text-dim">
                  {user?.email || 'user@example.com'}
                </p>
              </div>
              <div className="p-1">
                <button
                  onClick={handleSettings}
                  className="w-full flex items-center gap-2 px-3 py-2 text-sm text-text dark:text-text-dark hover:bg-surface-elevated dark:hover:bg-surface-elevated rounded transition-colors"
                >
                  <Settings className="w-4 h-4" />
                  الإعدادات
                </button>
                <button
                  onClick={handleLogout}
                  className="w-full flex items-center gap-2 px-3 py-2 text-sm text-danger hover:bg-surface-elevated dark:hover:bg-surface-elevated rounded transition-colors"
                >
                  <LogOut className="w-4 h-4" />
                  تسجيل الخروج
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </header>
  )
}
