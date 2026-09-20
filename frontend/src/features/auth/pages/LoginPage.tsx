import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../../../stores/authStore';
import { useTranslation } from '../../../hooks/useTranslation';
import { Sun, Moon, Globe, Cloud, HardDrive } from 'lucide-react';
import { getConnectionMode, setConnectionMode, type ConnectionMode } from '../../../lib/config/app';

// Components
import { LoginBackground } from '../components/LoginBackground';
import { BrandPanel } from '../components/BrandPanel';
import { LoginForm } from '../components/LoginForm';
import { SubscriptionVerificationScreen } from '../components/SubscriptionVerificationScreen';
import { isNetworkError } from '../../../lib/error-messages';

export function LoginPage() {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const { login, isLoading, loginError: authLoginError, setPostLoginVerifying } = useAuthStore();
  const [isVerifyingSubscription, setIsVerifyingSubscription] = useState(false);
  const [loginError, setLoginError] = useState('');
  const [isDark, setIsDark] = useState(true);
  const [language, setLanguage] = useState('ar');
  const [connectionMode, setSelectedConnectionMode] = useState<ConnectionMode>(getConnectionMode);

  // Check theme on mount and listen for changes
  useEffect(() => {
    sessionStorage.removeItem('partflow-login-error');

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

  const handleConnectionModeChange = (mode: ConnectionMode) => {
    setSelectedConnectionMode(mode);
    setConnectionMode(mode);
    setLoginError('');
  };

  const handleSubmit = async (email: string, password: string) => {
    setLoginError('');
    sessionStorage.removeItem('partflow-login-error');
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
      const message = isNetworkError(err)
        ? t('auth.connectionError')
        : t('auth.invalidCredentials');
      sessionStorage.setItem('partflow-login-error', message);
      setLoginError(message);
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
              <div style={{ marginBottom: '24px' }}>
                <div style={{
                  marginBottom: '9px',
                  color: isDark ? '#cbd5e1' : '#374151',
                  fontSize: '11px',
                  fontWeight: 700,
                }}>
                  طريقة الاتصال
                </div>
                <div
                  role="group"
                  aria-label="طريقة الاتصال"
                  style={{
                    display: 'grid',
                    gridTemplateColumns: '1fr 1fr',
                    gap: '7px',
                    padding: '5px',
                    borderRadius: '12px',
                    border: isDark ? '1px solid rgba(148, 163, 184, 0.16)' : '1px solid rgba(0, 0, 0, 0.09)',
                    background: isDark ? 'rgba(15, 23, 42, 0.72)' : 'rgba(243, 244, 246, 0.82)',
                  }}
                >
                  {([
                    { mode: 'local' as const, label: 'محلي', description: 'بيانات على الجهاز', Icon: HardDrive },
                    { mode: 'cloud' as const, label: 'سحابي Online', description: 'بيانات سحابية مباشرة', Icon: Cloud },
                  ]).map(({ mode, label, description, Icon }) => {
                    const selected = connectionMode === mode;
                    return (
                      <button
                        key={mode}
                        type="button"
                        onClick={() => handleConnectionModeChange(mode)}
                        aria-pressed={selected}
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: '8px',
                          minWidth: 0,
                          padding: '9px 8px',
                          border: selected ? '1px solid rgba(34, 211, 238, 0.48)' : '1px solid transparent',
                          borderRadius: '9px',
                          color: selected ? (isDark ? '#ecfeff' : '#0f172a') : (isDark ? '#94a3b8' : '#6b7280'),
                          background: selected
                            ? (isDark ? 'rgba(8, 145, 178, 0.18)' : 'rgba(14, 165, 233, 0.1)')
                            : 'transparent',
                          cursor: 'pointer',
                          textAlign: 'right',
                        }}
                      >
                        <Icon style={{ width: '16px', height: '16px', flexShrink: 0 }} />
                        <span style={{ minWidth: 0 }}>
                          <span style={{ display: 'block', fontSize: '11px', fontWeight: 750 }}>{label}</span>
                          <span style={{ display: 'block', marginTop: '2px', fontSize: '9px', opacity: 0.78, whiteSpace: 'nowrap' }}>{description}</span>
                        </span>
                      </button>
                    );
                  })}
                </div>
              </div>
              <LoginForm
                isDark={isDark}
                isLoading={isLoading}
                externalError={authLoginError === 'invalid credentials'
                  ? t('auth.invalidCredentials')
                  : authLoginError === 'connection error'
                    ? t('auth.connectionError')
                    : loginError}
                onSubmit={handleSubmit}
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
