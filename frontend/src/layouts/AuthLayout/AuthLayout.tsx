import type { ReactNode } from 'react';
import { useTranslation } from '../../hooks/useTranslation';

interface AuthLayoutProps {
  children: ReactNode;
}

/**
 * AuthLayout - Layout component فقط
 * لا يفرض تصميماً محدداً، يوفر البنية الأساسية فقط
 * تصميم الصفحة مسؤولية الصفحة نفسها (مثل LoginPage)
 */
export function AuthLayout({ children }: AuthLayoutProps) {
  const { direction } = useTranslation();

  return (
    <div 
      className={`min-h-screen ${direction}`}
      dir={direction === 'rtl' ? 'rtl' : 'ltr'}
    >
      {children}
    </div>
  );
}