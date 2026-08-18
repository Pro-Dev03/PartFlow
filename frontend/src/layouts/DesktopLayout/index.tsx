import { Outlet } from 'react-router-dom'
import { Sidebar } from '../../components/navigation/Sidebar'
import { TopBar } from '../../components/navigation/TopBar'

export function DesktopLayout() {
  return (
    <div className="flex h-screen bg-background">
      <Sidebar />
      <div className="flex-1 flex flex-col overflow-hidden me-[52px]">
        <TopBar />
        <main className="flex-1 overflow-auto scrollbar-thin">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
