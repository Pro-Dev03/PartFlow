import { Outlet } from 'react-router-dom'
import { MobileNav } from '../../components/navigation/MobileNav'

export function MobileLayout() {
  return (
    <div className="flex flex-col h-screen bg-background">
      <main className="flex-1 overflow-auto">
        <Outlet />
      </main>
      <MobileNav />
    </div>
  )
}
