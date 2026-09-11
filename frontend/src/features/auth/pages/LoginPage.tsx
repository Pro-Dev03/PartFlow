import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../../../stores/authStore';
import { Sun, Moon, Globe } from 'lucide-react';

// Components
import { LoginBackground } from '../components/LoginBackground';
import { BrandPanel } from '../components/BrandPanel';
import { LoginForm } from '../components/LoginForm';
import { SubscriptionVerificationScreen } from '../components/SubscriptionVerificationScreen';

export function LoginPage() {
  const navigate = useNavigate();
  const { login, isLoading, setPostLoginVerifying } = useAuthStore();
  const [isVerifyingSubscription, setIsVerifyingSubscription] = useState(false);
  const [isDark, setIsDark] = useState(true);
  const [language, setLanguage] = useState('ar');

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

  const toggleTheme = () => {
    const html = document.documentElement;
    if (isDark) {
      html.classList.add('light');
      html.classList.remove('dark');
    } else {
      html.classList.add('dark');
      html.classList.remove('light');
    }
    setIsDark(!isDark);
  };

  const toggleLanguage = () => {
    const newLang = language === 'ar' ? 'en' : 'ar';
    setLanguage(newLang);
    document.documentElement.setAttribute('dir', newLang === 'ar' ? 'rtl' : 'ltr');
    // Store preference
    localStorage.setItem('language', newLang);
  };

  const handleSubmit = async (email: string, password: string) => {
    setPostLoginVerifying(true);
    setIsVerifyingSubscription(true);
    try {
      await login(email, password);
      await new Promise((resolve) => window.setTimeout(resolve, 6000));
      setPostLoginVerifying(false);
      navigate('/app');
    } catch (err) {
      setPostLoginVerifying(false);
      setIsVerifyingSubscription(false);
      throw err;
    }
  };

  if (isVerifyingSubscription) {
    return <SubscriptionVerificationScreen />;
  }

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
          <div style={{ position: 'relative', padding: '92px 46px 46px', display: 'flex', alignItems: 'center' }} className="md:p-[30px]">
            {/* Theme and Language Toggles */}
            <div
              style={{
                position: 'absolute',
                top: '20px',
                left: '20px',
                display: 'flex',
                gap: '12px',
                direction: 'ltr',
              }}
            >
              <button
                onClick={toggleLanguage}
                style={{
                  width: '40px',
                  height: '40px',
                  borderRadius: '10px',
                  background: isDark ? 'rgba(17, 24, 39, 0.8)' : 'rgba(255, 255, 255, 0.9)',
                  border: isDark ? '1px solid rgba(148, 163, 184, 0.13)' : '1px solid rgba(0, 0, 0, 0.08)',
                  color: isDark ? '#8290a7' : '#6B7280',
                  cursor: 'pointer',
                  transition: 'all 180ms ease',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  boxShadow: isDark ? '0 2px 8px rgba(0, 0, 0, 0.2)' : '0 2px 8px rgba(0, 0, 0, 0.1)',
                }}
                title={language === 'ar' ? 'English' : 'العربية'}
              >
                <Globe style={{ width: '18px', height: '18px' }} />
              </button>
              <button
                onClick={toggleTheme}
                style={{
                  width: '40px',
                  height: '40px',
                  borderRadius: '10px',
                  background: isDark ? 'rgba(17, 24, 39, 0.8)' : 'rgba(255, 255, 255, 0.9)',
                  border: isDark ? '1px solid rgba(148, 163, 184, 0.13)' : '1px solid rgba(0, 0, 0, 0.08)',
                  color: isDark ? '#8290a7' : '#6B7280',
                  cursor: 'pointer',
                  transition: 'all 180ms ease',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  boxShadow: isDark ? '0 2px 8px rgba(0, 0, 0, 0.2)' : '0 2px 8px rgba(0, 0, 0, 0.1)',
                }}
                title={isDark ? 'الوضع الفاتح' : 'الوضع الليلي'}
              >
                {isDark ? <Sun style={{ width: '18px', height: '18px' }} /> : <Moon style={{ width: '18px', height: '18px' }} />}
              </button>
            </div>
            <div style={{ width: '100%', maxWidth: '360px', margin: '0 auto' }}>
              <LoginForm isDark={isDark} isLoading={isLoading} onSubmit={handleSubmit} />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
