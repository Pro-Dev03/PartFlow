import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../../../stores/authStore';

// Components
import { LoginBackground } from '../components/LoginBackground';
import { BrandPanel } from '../components/BrandPanel';
import { LoginForm } from '../components/LoginForm';

export function LoginPage() {
  const navigate = useNavigate();
  const { login, isLoading } = useAuthStore();
  const [isDark, setIsDark] = useState(true);

  // Check theme on mount and listen for changes
  useEffect(() => {
    const checkTheme = () => {
      setIsDark(!document.documentElement.classList.contains('light'));
    };
    
    checkTheme();
    
    const observer = new MutationObserver(checkTheme);
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class']
    });
    
    return () => observer.disconnect();
  }, []);

  const handleSubmit = async (email: string, password: string) => {
    try {
      await login(email, password);
      navigate('/app');
    } catch (err) {
      throw err;
    }
  };

  return (
    <div dir="rtl" className="min-h-screen grid place-items-center relative overflow-hidden" style={{ background: 'var(--bg-background)' }}>
      <LoginBackground isDark={isDark} />
      
      {/* Main Container - Split Layout */}
      <div style={{ position: 'relative', zIndex: 2, width: 'min(920px, calc(100% - 32px))' }}>
        <div style={{
          display: 'grid',
          gridTemplateColumns: '1fr 1fr',
          border: isDark ? '1px solid rgba(148, 163, 184, 0.13)' : '1px solid rgba(0, 0, 0, 0.08)',
          borderRadius: '22px',
          background: isDark ? 'linear-gradient(145deg, rgba(17, 24, 39, 0.94), rgba(9, 14, 24, 0.96))' : 'linear-gradient(145deg, rgba(255, 255, 255, 0.95), rgba(249, 250, 251, 0.97))',
          boxShadow: isDark ? '0 20px 70px rgba(0, 0, 0, 0.35), 0 0 35px rgba(34, 211, 238, 0.08)' : '0 20px 70px rgba(0, 0, 0, 0.1), 0 0 35px rgba(37, 99, 235, 0.05)',
          overflow: 'hidden',
        }}>
          <BrandPanel isDark={isDark} />
          
          {/* LOGIN PANEL */}
          <div style={{ padding: '46px', display: 'flex', alignItems: 'center' }} className="md:p-[30px]">
            <LoginForm isDark={isDark} isLoading={isLoading} onSubmit={handleSubmit} />
          </div>
        </div>
      </div>
    </div>
  );
}
