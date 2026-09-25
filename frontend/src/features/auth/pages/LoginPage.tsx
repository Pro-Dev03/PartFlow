import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../../../stores/authStore';
import { useTranslation } from '../../../hooks/useTranslation';
import { Cloud, HardDrive } from 'lucide-react';
import { getConnectionMode, isElectronRuntime, setConnectionMode, type ConnectionMode } from '../../../lib/config/app';

// Components
import { LoginBackground } from '../components/LoginBackground';
import { BrandPanel } from '../components/BrandPanel';
import { LoginForm } from '../components/LoginForm';
import { SubscriptionVerificationScreen } from '../components/SubscriptionVerificationScreen';
import { isNetworkError } from '../../../lib/error-messages';
import { useUIStore } from '../../../stores/uiStore';

export function LoginPage() {
  const navigate = useNavigate();
  const { t, currentLanguage, changeLanguage } = useTranslation();
  const { login, isLoading, loginError: authLoginError, setPostLoginVerifying } = useAuthStore();
  const [isVerifyingSubscription, setIsVerifyingSubscription] = useState(false);
  const [loginError, setLoginError] = useState('');
  const theme = useUIStore((state) => state.theme);
  const setTheme = useUIStore((state) => state.setTheme);
  const isDark = theme === 'dark';
  const [connectionMode, setSelectedConnectionMode] = useState<ConnectionMode>(getConnectionMode);
  const isElectron = isElectronRuntime();
  const connectionOptions = [
    { mode: 'local' as const, label: 'محلي', description: 'بيانات على الجهاز', Icon: HardDrive },
    ...(!isElectron ? [{ mode: 'cloud' as const, label: 'سحابي Online', description: 'بيانات سحابية مباشرة', Icon: Cloud }] : []),
  ];

  useEffect(() => {
    sessionStorage.removeItem('partflow-login-error');
  }, []);

  const toggleTheme = () => {
    setTheme(isDark ? 'light' : 'dark');
  };

  const toggleLanguage = () => {
    changeLanguage(currentLanguage === 'ar' ? 'en' : 'ar');
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
      setIsVerifyingSubscription(false);
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
    <>
      <LoginBackground isDark={isDark} />
      <div className="pf-login-quick-actions">
        <button
          type="button"
          onClick={toggleLanguage}
          className="pf-login-icon-button"
          title="اللغة"
          aria-label={currentLanguage === 'ar' ? 'English' : 'العربية'}
        >
          🌐
        </button>
        <button
          type="button"
          onClick={toggleTheme}
          className="pf-login-icon-button"
          title="الوضع"
          aria-label={isDark ? 'الوضع الفاتح' : 'الوضع الليلي'}
        >
          {isDark ? '🌙' : '☀️'}
        </button>
      </div>
      <div
        dir={currentLanguage === 'ar' ? 'rtl' : 'ltr'}
        data-theme={isDark ? 'dark' : 'light'}
        className="pf-login-page grid place-items-center relative"
      >
        {/* Main Container - Split Layout */}
        <div className="pf-login-card" style={{ position: 'relative' }}>
        <div className="pf-login-surface" style={{
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
          <div style={{ position: 'relative', display: 'flex', alignItems: 'center' }} className="pf-login-form-panel">
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
                    gridTemplateColumns: isElectron ? '1fr' : '1fr 1fr',
                    gap: '7px',
                    padding: '5px',
                    borderRadius: '12px',
                    border: isDark ? '1px solid rgba(148, 163, 184, 0.16)' : '1px solid rgba(0, 0, 0, 0.09)',
                    background: isDark ? 'rgba(15, 23, 42, 0.72)' : 'rgba(243, 244, 246, 0.82)',
                  }}
                >
                  {connectionOptions.map(({ mode, label, description, Icon }) => {
                    const selected = connectionMode === mode;
                    return (
                      <button
                        key={mode}
                        type="button"
                        className="pf-login-method"
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
                        {mode === 'cloud' ? (
                          <span
                            aria-hidden="true"
                            style={{
                              width: '10px',
                              height: '10px',
                              borderRadius: '50%',
                              background: 'radial-gradient(circle, #a7f3d0 0%, #22c55e 42%, #15803d 100%)',
                              boxShadow: '0 0 0 1px rgba(34, 197, 94, 0.28), 0 0 10px rgba(34, 197, 94, 0.8), 0 0 18px rgba(34, 197, 94, 0.5)',
                              flexShrink: 0,
                              display: 'inline-block',
                            }}
                            className="pf-login-method-dot"
                          />
                        ) : (
                          <span
                            aria-hidden="true"
                            style={{
                              width: '10px',
                              height: '10px',
                              borderRadius: '50%',
                              background: 'radial-gradient(circle, #fde68a 0%, #f59e0b 42%, #b45309 100%)',
                              boxShadow: '0 0 0 1px rgba(245, 158, 11, 0.28), 0 0 10px rgba(245, 158, 11, 0.8), 0 0 18px rgba(245, 158, 11, 0.45)',
                              flexShrink: 0,
                              display: 'inline-block',
                            }}
                            className="pf-login-method-dot"
                          />
                        )}
                        <Icon aria-hidden="true" style={{ width: '15px', height: '15px', flexShrink: 0 }} />
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
    </>
  );
}
