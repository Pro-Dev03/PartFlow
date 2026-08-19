import { useState } from 'react';
import { Header } from '../../components/navigation/header';
import { Sidebar } from '../../components/navigation/sidebar';
import { useTranslation } from '../../hooks/useTranslation';
import { cn } from '../../lib/utils';

interface AppLayoutProps {
  children: React.ReactNode;
}

export function AppLayout({ children }: AppLayoutProps) {
  const { direction } = useTranslation();
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);

  return (
    <div 
      dir={direction} 
      className="min-h-screen bg-bg-gradient"
      style={{
        background: 'radial-gradient(circle at 80% 0%, rgba(34, 211, 238, 0.10), transparent 30%), radial-gradient(circle at 20% 80%, rgba(59, 130, 246, 0.08), transparent 30%), #070a12'
      }}
    >
      <div className="flex flex-col h-screen">
        <Header onToggleSidebar={() => setIsSidebarCollapsed(!isSidebarCollapsed)} />
        <div className="flex flex-1 overflow-hidden">
          <Sidebar 
            isCollapsed={isSidebarCollapsed} 
            onToggle={() => setIsSidebarCollapsed(!isSidebarCollapsed)} 
          />
          <main className={cn(
            'flex-1 overflow-y-auto',
            'max-w-[1500px] mx-auto',
            'w-full',
            'px-2xl py-2xl pb-4xl'
          )}>
            {children}
          </main>
        </div>
      </div>
    </div>
  );
}