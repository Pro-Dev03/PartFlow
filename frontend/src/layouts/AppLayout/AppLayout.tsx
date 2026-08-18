import { useState } from 'react';
import { Header } from '../../components/navigation/header';
import { Sidebar } from '../../components/navigation/sidebar';
import { useTranslation } from '../../hooks/useTranslation';

interface AppLayoutProps {
  children: React.ReactNode;
}

export function AppLayout({ children }: AppLayoutProps) {
  const { direction } = useTranslation();
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);

  return (
    <div className={`min-h-screen bg-gray-50 dark:bg-gray-900 ${direction}`}>
      <div className="flex flex-col h-screen">
        <Header onToggleSidebar={() => setIsSidebarCollapsed(!isSidebarCollapsed)} />
        <div className="flex flex-1 overflow-hidden">
          <Sidebar 
            isCollapsed={isSidebarCollapsed} 
            onToggle={() => setIsSidebarCollapsed(!isSidebarCollapsed)} 
          />
          <main className="flex-1 overflow-y-auto p-6">
            {children}
          </main>
        </div>
      </div>
    </div>
  );
}