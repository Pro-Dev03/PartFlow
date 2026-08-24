import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from '../../../hooks/useTranslation';
import { useAuthStore } from '../../../stores/authStore';
import {
  Eye,
  EyeOff,
  ArrowRight,
  Shield,
  CheckCircle,
  AlertCircle
} from 'lucide-react';

export function LoginPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { login, isLoading } = useAuthStore();
  
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [rememberMe, setRememberMe] = useState(false);
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    
    try {
      await login(email, password);
      navigate('/app');
    } catch (err) {
      setError(t('auth.invalidCredentials'));
    }
  };

  return (
    <div dir="rtl" className="min-h-screen grid place-items-center relative overflow-hidden" style={{ background: 'var(--bg-background)' }}>
      {/* Futuristic Background - مطابق demo-logim.html */}
      <div style={{
        position: 'absolute',
        inset: 0,
        background: isDark ? `
          radial-gradient(circle at 80% 0%, rgba(34, 211, 238, 0.11), transparent 30%),
          radial-gradient(circle at 15% 85%, rgba(59, 130, 246, 0.08), transparent 32%),
          var(--bg-background)
        ` : 'radial-gradient(circle at 80% 0%, rgba(37, 99, 235, 0.05), transparent 30%), radial-gradient(circle at 15% 85%, rgba(37, 99, 235, 0.03), transparent 32%), var(--bg-background)'
      }} />
      
      {/* Subtle futuristic grid */}
      <div style={{
        position: 'absolute',
        inset: 0,
        pointerEvents: 'none',
        backgroundImage: `
          linear-gradient(rgba(148, 163, 184, 0.025) 1px, transparent 1px),
          linear-gradient(90deg, rgba(148, 163, 184, 0.025) 1px, transparent 1px)
        `,
        backgroundSize: '45px 45px',
        maskImage: 'linear-gradient(to bottom, black, transparent 85%)',
        WebkitMaskImage: 'linear-gradient(to bottom, black, transparent 85%)'
      }} />

      {/* Ambient glow orbs */}
      <div style={{
        position: 'absolute',
        top: '-220px',
        right: '-120px',
        width: '420px',
        height: '420px',
        borderRadius: '50%',
        background: isDark ? 'radial-gradient(circle, rgba(34, 211, 238, 0.08), transparent 68%)' : 'radial-gradient(circle, rgba(37, 99, 235, 0.04), transparent 68%)',
        filter: 'blur(20px)',
        pointerEvents: 'none'
      }} />
      <div style={{
        position: 'absolute',
        bottom: '-260px',
        left: '-180px',
        width: '420px',
        height: '420px',
        borderRadius: '50%',
        background: isDark ? 'radial-gradient(circle, rgba(59, 130, 246, 0.07), transparent 68%)' : 'radial-gradient(circle, rgba(37, 99, 235, 0.03), transparent 68%)',
        filter: 'blur(20px)',
        pointerEvents: 'none'
      }} />

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
          backdropFilter: 'blur(18px)'
        }}>
          
          {/* BRAND PANEL */}
          <div style={{
            position: 'relative',
            minHeight: '590px',
            padding: '46px',
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'space-between',
            borderRight: isDark ? '1px solid rgba(148, 163, 184, 0.13)' : '1px solid rgba(0, 0, 0, 0.08)',
            background: isDark ? 'linear-gradient(145deg, rgba(34, 211, 238, 0.055), transparent 45%)' : 'linear-gradient(145deg, rgba(37, 99, 235, 0.03), transparent 45%)'
          }} className="md:block hidden">
            
            {/* Brand */}
            <div style={{ display: 'flex', alignItems: 'center', gap: '13px' }}>
              <div style={{
                width: '44px',
                height: '44px',
                display: 'grid',
                placeItems: 'center',
                borderRadius: '13px',
                color: isDark ? '#22d3ee' : '#2563EB',
                border: isDark ? '1px solid rgba(34, 211, 238, 0.25)' : '1px solid rgba(37, 99, 235, 0.25)',
                background: isDark ? 'linear-gradient(145deg, rgba(34, 211, 238, 0.13), rgba(59, 130, 246, 0.06))' : 'linear-gradient(145deg, rgba(37, 99, 235, 0.08), rgba(37, 99, 235, 0.04))',
                boxShadow: isDark ? '0 0 30px rgba(34, 211, 238, 0.10)' : '0 0 30px rgba(37, 99, 235, 0.08)'
              }}>
                <span style={{ fontSize: '19px', fontWeight: '800', letterSpacing: '-1px' }}>PF</span>
              </div>
              <div>
                <div style={{ fontSize: '17px', fontWeight: '750', letterSpacing: '-0.4px', color: isDark ? '#f1f7ff' : '#111827' }}>
                  PartFlow
                </div>
                <div style={{ fontSize: '10px', letterSpacing: '1.2px', textTransform: 'uppercase', color: isDark ? '#8290a7' : '#6B7280', marginTop: '-2px' }}>
                  نظام إدارة المتاجر
                </div>
              </div>
            </div>

            {/* Hero Copy */}
            <div style={{ maxWidth: '340px' }}>
              <div style={{ marginBottom: '14px', color: isDark ? '#22d3ee' : '#2563EB', fontSize: '10px', fontWeight: '700', textTransform: 'uppercase', letterSpacing: '2px' }}>
                نظام إدارة ذكي
              </div>
              <h1 style={{ fontSize: 'clamp(30px, 4vw, 42px)', lineHeight: '1.08', letterSpacing: '-1.8px', fontWeight: '800', color: isDark ? '#f1f7ff' : '#111827' }}>
                متجرك.
                <br />
                <span style={{ color: isDark ? '#22d3ee' : '#2563EB', textShadow: isDark ? '0 0 25px rgba(34, 211, 238, 0.15)' : 'none' }}>
                  تحت سيطرتك.
                </span>
              </h1>
              <p style={{ marginTop: '18px', color: isDark ? '#8290a7' : '#6B7280', fontSize: '13px', lineHeight: '1.8' }}>
                إدارة المخزون، المبيعات، العملاء، الديون والعمليات اليومية من مساحة عمل ذكية واحدة.
              </p>
            </div>

            {/* Brand Footer */}
            <div style={{ color: isDark ? '#56647a' : '#9CA3AF', fontSize: '10px', letterSpacing: '1px', textTransform: 'uppercase' }}>
              نظام PartFlow
            </div>
          </div>

          {/* LOGIN PANEL */}
          <div style={{ padding: '46px', display: 'flex', alignItems: 'center' }} className="md:p-[30px]">
            <div style={{ width: '100%', maxWidth: '360px', margin: 'auto' }}>
              
              {/* Mobile Brand */}
              <div style={{ display: 'flex', alignItems: 'center', gap: '13px', marginBottom: '32px' }} className="md:hidden">
                <div style={{
                  width: '44px',
                  height: '44px',
                  display: 'grid',
                  placeItems: 'center',
                  borderRadius: '13px',
                  color: isDark ? '#22d3ee' : '#2563EB',
                  border: isDark ? '1px solid rgba(34, 211, 238, 0.25)' : '1px solid rgba(37, 99, 235, 0.25)',
                  background: isDark ? 'linear-gradient(145deg, rgba(34, 211, 238, 0.13), rgba(59, 130, 246, 0.06))' : 'linear-gradient(145deg, rgba(37, 99, 235, 0.08), rgba(37, 99, 235, 0.04))',
                  boxShadow: isDark ? '0 0 30px rgba(34, 211, 238, 0.10)' : '0 0 30px rgba(37, 99, 235, 0.08)'
                }}>
                  <span style={{ fontSize: '19px', fontWeight: '800', letterSpacing: '-1px' }}>PF</span>
                </div>
                <div>
                  <div style={{ fontSize: '17px', fontWeight: '750', letterSpacing: '-0.4px', color: isDark ? '#f1f7ff' : '#111827' }}>
                    PartFlow
                  </div>
                  <div style={{ fontSize: '10px', letterSpacing: '1.2px', textTransform: 'uppercase', color: isDark ? '#8290a7' : '#6B7280', marginTop: '-2px' }}>
                    نظام إدارة المتاجر
                  </div>
                </div>
              </div>
              
              {/* Login Header */}
              <div style={{ marginBottom: '30px' }}>
                <h2 style={{ fontSize: '25px', letterSpacing: '-0.8px', fontWeight: '750', color: isDark ? '#f1f7ff' : '#111827' }}>
                  {t('auth.welcomeBack')}
                </h2>
                <p style={{ marginTop: '6px', color: isDark ? '#8290a7' : '#6B7280', fontSize: '12px' }}>
                  {t('auth.signInToAccount')}
                </p>
              </div>

              {/* Error message */}
              {error && (
                <div style={{
                  marginBottom: '18px',
                  padding: '11px 12px',
                  display: 'flex',
                  gap: '10px',
                  border: '1px solid rgba(251, 113, 133, 0.3)',
                  borderRadius: '10px',
                  background: 'rgba(251, 113, 133, 0.1)',
                  color: '#fb7185',
                  fontSize: '10px',
                  lineHeight: '1.5'
                }}>
                  <span style={{ color: '#fb7185', flexShrink: 0 }}>
                    <AlertCircle style={{ width: '12px', height: '12px' }} />
                  </span>
                  <span>{error}</span>
                </div>
              )}

              <form onSubmit={handleSubmit}>
                {/* Email Input */}
                <div style={{ marginBottom: '18px' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
                    <label style={{ color: isDark ? '#cbd5e1' : '#374151', fontSize: '11px', fontWeight: '650' }} htmlFor="email">
                      البريد الإلكتروني
                    </label>
                  </div>
                  <input
                    id="email"
                    type="email"
                    placeholder="you@example.com"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    required
                    autoComplete="email"
                    style={{
                      width: '100%',
                      padding: '12px 13px',
                      color: isDark ? '#f1f7ff' : '#111827',
                      border: isDark ? '1px solid rgba(148, 163, 184, 0.13)' : '1px solid rgba(0, 0, 0, 0.08)',
                      borderRadius: '10px',
                      outline: 'none',
                      background: isDark ? 'rgba(17, 24, 39, 0.72)' : 'rgba(255, 255, 255, 0.8)',
                      transition: 'border-color 180ms ease, box-shadow 180ms ease, background 180ms ease',
                      fontSize: '13px'
                    }}
                    className="placeholder:text-[#4f5c70]"
                    onMouseEnter={(e) => {
                      e.currentTarget.style.borderColor = isDark ? 'rgba(148, 163, 184, 0.22)' : 'rgba(0, 0, 0, 0.12)';
                    }}
                    onMouseLeave={(e) => {
                      e.currentTarget.style.borderColor = isDark ? 'rgba(148, 163, 184, 0.13)' : '1px solid rgba(0, 0, 0, 0.08)';
                    }}
                    onFocus={(e) => {
                      e.currentTarget.style.borderColor = isDark ? 'rgba(34, 211, 238, 0.45)' : 'rgba(37, 99, 235, 0.5)';
                      e.currentTarget.style.background = isDark ? 'rgba(17, 24, 39, 0.95)' : 'rgba(255, 255, 255, 0.95)';
                      e.currentTarget.style.boxShadow = isDark ? '0 0 0 3px rgba(34, 211, 238, 0.07), 0 0 25px rgba(34, 211, 238, 0.05)' : '0 0 0 3px rgba(37, 99, 235, 0.1), 0 0 25px rgba(37, 99, 235, 0.08)';
                    }}
                    onBlur={(e) => {
                      e.currentTarget.style.borderColor = isDark ? 'rgba(148, 163, 184, 0.13)' : '1px solid rgba(0, 0, 0, 0.08)';
                      e.currentTarget.style.background = isDark ? 'rgba(17, 24, 39, 0.72)' : 'rgba(255, 255, 255, 0.8)';
                      e.currentTarget.style.boxShadow = 'none';
                    }}
                  />
                </div>

                {/* Password Input */}
                <div style={{ marginBottom: '18px' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
                    <label style={{ color: isDark ? '#cbd5e1' : '#374151', fontSize: '11px', fontWeight: '650' }} htmlFor="password">
                      كلمة المرور
                    </label>
                  </div>
                  <div style={{ position: 'relative' }}>
                    <input
                      id="password"
                      type={showPassword ? 'text' : 'password'}
                      placeholder="أدخل كلمة المرور"
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      required
                      autoComplete="current-password"
                      style={{
                        width: '100%',
                        padding: '12px 13px',
                        color: isDark ? '#f1f7ff' : '#111827',
                        border: isDark ? '1px solid rgba(148, 163, 184, 0.13)' : '1px solid rgba(0, 0, 0, 0.08)',
                        borderRadius: '10px',
                        outline: 'none',
                        background: isDark ? 'rgba(17, 24, 39, 0.72)' : 'rgba(255, 255, 255, 0.8)',
                        transition: 'border-color 180ms ease, box-shadow 180ms ease, background 180ms ease',
                        fontSize: '13px'
                      }}
                      className="placeholder:text-[#4f5c70]"
                      onMouseEnter={(e) => {
                        e.currentTarget.style.borderColor = isDark ? 'rgba(148, 163, 184, 0.22)' : 'rgba(0, 0, 0, 0.12)';
                      }}
                      onMouseLeave={(e) => {
                        e.currentTarget.style.borderColor = isDark ? 'rgba(148, 163, 184, 0.13)' : '1px solid rgba(0, 0, 0, 0.08)';
                      }}
                      onFocus={(e) => {
                        e.currentTarget.style.borderColor = isDark ? 'rgba(34, 211, 238, 0.45)' : 'rgba(37, 99, 235, 0.5)';
                        e.currentTarget.style.background = isDark ? 'rgba(17, 24, 39, 0.95)' : 'rgba(255, 255, 255, 0.95)';
                        e.currentTarget.style.boxShadow = isDark ? '0 0 0 3px rgba(34, 211, 238, 0.07), 0 0 25px rgba(34, 211, 238, 0.05)' : '0 0 0 3px rgba(37, 99, 235, 0.1), 0 0 25px rgba(37, 99, 235, 0.08)';
                      }}
                      onBlur={(e) => {
                        e.currentTarget.style.borderColor = isDark ? 'rgba(148, 163, 184, 0.13)' : '1px solid rgba(0, 0, 0, 0.08)';
                        e.currentTarget.style.background = isDark ? 'rgba(17, 24, 39, 0.72)' : 'rgba(255, 255, 255, 0.8)';
                        e.currentTarget.style.boxShadow = 'none';
                      }}
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      style={{
                        position: 'absolute',
                        top: '50%',
                        left: '-35px',
                        transform: 'translateY(-50%)',
                        padding: '6px',
                        color: isDark ? '#8290a7' : '#6B7280',
                        background: isDark ? 'rgba(17, 24, 39, 0.8)' : 'rgba(255, 255, 255, 0.9)',
                        border: isDark ? '1px solid rgba(148, 163, 184, 0.13)' : '1px solid rgba(0, 0, 0, 0.08)',
                        cursor: 'pointer',
                        transition: 'all 180ms ease',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        borderRadius: '8px',
                        boxShadow: isDark ? '0 2px 8px rgba(0, 0, 0, 0.2)' : '0 2px 8px rgba(0, 0, 0, 0.1)'
                      }}
                      onMouseEnter={(e) => {
                        e.currentTarget.style.color = isDark ? '#22d3ee' : '#2563EB';
                        e.currentTarget.style.borderColor = isDark ? 'rgba(34, 211, 238, 0.3)' : 'rgba(37, 99, 235, 0.3)';
                        e.currentTarget.style.transform = 'translateY(-50%) scale(1.05)';
                      }}
                      onMouseLeave={(e) => {
                        e.currentTarget.style.color = isDark ? '#8290a7' : '#6B7280';
                        e.currentTarget.style.borderColor = isDark ? 'rgba(148, 163, 184, 0.13)' : '1px solid rgba(0, 0, 0, 0.08)';
                        e.currentTarget.style.transform = 'translateY(-50%) scale(1)';
                      }}
                      aria-label={showPassword ? 'إخفاء كلمة المرور' : 'إظهار كلمة المرور'}
                    >
                      {showPassword ? <EyeOff style={{ width: '16px', height: '16px' }} /> : <Eye style={{ width: '16px', height: '16px' }} />}
                    </button>
                  </div>
                </div>

                {/* Form Options */}
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', margin: '8px 0 22px' }}>
                  <label style={{ display: 'flex', alignItems: 'center', gap: '8px', color: isDark ? '#8290a7' : '#6B7280', fontSize: '11px', cursor: 'pointer' }}>
                    <input
                      type="checkbox"
                      checked={rememberMe}
                      onChange={(e) => setRememberMe(e.target.checked)}
                      style={{ accentColor: isDark ? '#22d3ee' : '#2563EB' }}
                    />
                    {t('auth.rememberMe')}
                  </label>
                  <button
                    type="button"
                    style={{
                      color: isDark ? '#22d3ee' : '#2563EB',
                      fontSize: '11px',
                      background: 'none',
                      border: 'none',
                      cursor: 'pointer',
                      textDecoration: 'none'
                    }}
                    onMouseEnter={(e) => {
                      e.currentTarget.style.textDecoration = 'underline';
                    }}
                    onMouseLeave={(e) => {
                      e.currentTarget.style.textDecoration = 'none';
                    }}
                  >
                    {t('auth.forgotPassword')}
                  </button>
                </div>

                {/* Login Button */}
                <button
                  type="submit"
                  disabled={isLoading}
                  style={{
                    width: '100%',
                    padding: '12px 14px',
                    border: isDark ? '1px solid rgba(34, 211, 238, 0.35)' : '1px solid rgba(37, 99, 235, 0.35)',
                    borderRadius: '10px',
                    color: isDark ? '#eaffff' : '#FFFFFF',
                    fontWeight: '700',
                    cursor: isLoading ? 'not-allowed' : 'pointer',
                    background: isDark ? 'linear-gradient(135deg, rgba(34, 211, 238, 0.17), rgba(59, 130, 246, 0.12))' : 'linear-gradient(135deg, rgba(37, 99, 235, 0.9), rgba(37, 99, 235, 0.8))',
                    boxShadow: isDark ? '0 0 25px rgba(34, 211, 238, 0.07)' : '0 0 25px rgba(37, 99, 235, 0.15)',
                    transition: 'transform 180ms ease, border-color 180ms ease, box-shadow 180ms ease',
                    opacity: isLoading ? 0.5 : 1,
                    fontSize: '13px'
                  }}
                  onMouseEnter={(e) => {
                    if (!isLoading) {
                      e.currentTarget.style.transform = 'translateY(-1px)';
                      e.currentTarget.style.borderColor = isDark ? 'rgba(34, 211, 238, 0.55)' : 'rgba(37, 99, 235, 0.6)';
                      e.currentTarget.style.boxShadow = isDark ? '0 0 35px rgba(34, 211, 238, 0.12)' : '0 0 35px rgba(37, 99, 235, 0.2)';
                    }
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.transform = 'translateY(0)';
                    e.currentTarget.style.borderColor = isDark ? 'rgba(34, 211, 238, 0.35)' : '1px solid rgba(37, 99, 235, 0.35)';
                    e.currentTarget.style.boxShadow = isDark ? '0 0 25px rgba(34, 211, 238, 0.07)' : '0 0 25px rgba(37, 99, 235, 0.15)';
                  }}
                  onMouseDown={(e) => {
                    e.currentTarget.style.transform = 'translateY(0)';
                  }}
                  onFocus={(e) => {
                    e.currentTarget.style.outline = isDark ? '2px solid #22d3ee' : '2px solid #2563EB';
                    e.currentTarget.style.outlineOffset = '3px';
                  }}
                  onBlur={(e) => {
                    e.currentTarget.style.outline = 'none';
                  }}
                >
                  {isLoading ? (
                    <span style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '8px' }}>
                      <svg style={{ animation: 'spin 1s linear infinite', width: '16px', height: '16px' }} xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                      </svg>
                      جاري تسجيل الدخول...
                    </span>
                  ) : (
                    <span style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '8px' }}>
                      {t('auth.login')}
                      <Shield style={{ width: '16px', height: '16px' }} />
                    </span>
                  )}
                </button>
              </form>

              {/* Additional info */}
              <div style={{ marginTop: '24px', textAlign: 'center' }}>
                <p style={{ fontSize: '13px', color: '#8290a7' }}>
                  ليس لديك حساب؟{' '}
                  <button 
                    onClick={() => navigate('/')}
                    style={{
                      color: '#22d3ee',
                      background: 'none',
                      border: 'none',
                      cursor: 'pointer',
                      fontWeight: '500',
                      fontSize: '13px',
                      display: 'inline-flex',
                      alignItems: 'center',
                      gap: '8px',
                      margin: '0 auto'
                    }}
                    onMouseEnter={(e) => {
                      e.currentTarget.style.color = 'rgba(34, 211, 238, 0.8)';
                    }}
                    onMouseLeave={(e) => {
                      e.currentTarget.style.color = '#22d3ee';
                    }}
                  >
                    تواصل مع الإدارة
                    <ArrowRight style={{ width: '16px', height: '16px' }} />
                  </button>
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Responsive Styles */}
      <style>{`
        @media (max-width: 760px) {
          .md\\:block { display: none !important; }
          .md\\:p-\\[30px\\] { padding: 30px !important; }
        }
        @keyframes spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
      `}</style>
    </div>
  );
}