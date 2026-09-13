import { useEffect, useState } from 'react';
import { useTranslation } from '../../../hooks/useTranslation';
import { Eye, EyeOff, AlertCircle, Headset, X } from 'lucide-react';

interface LoginFormProps {
  isDark: boolean;
  isLoading: boolean;
  externalError?: string;
  onSubmit: (email: string, password: string) => Promise<void>;
}

export function LoginForm({ isDark, isLoading, externalError, onSubmit }: LoginFormProps) {
  const { t } = useTranslation();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [rememberMe, setRememberMe] = useState(false);
  const [showForgotPassword, setShowForgotPassword] = useState(false);

  useEffect(() => {
    if (!showForgotPassword) return;

    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setShowForgotPassword(false);
    };
    document.addEventListener('keydown', handleEscape);
    return () => document.removeEventListener('keydown', handleEscape);
  }, [showForgotPassword]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    
    try {
      await onSubmit(email, password);
    } catch (err) {
      setError(t('auth.invalidCredentials'));
    }
  };

  const inputStyle = {
    width: '100%',
    boxSizing: 'border-box' as const,
    paddingTop: '12px',
    paddingRight: '13px',
    paddingBottom: '12px',
    paddingLeft: '13px',
    color: isDark ? '#f1f7ff' : '#111827',
    border: isDark ? '1px solid rgba(148, 163, 184, 0.13)' : '1px solid rgba(0, 0, 0, 0.08)',
    borderRadius: '10px',
    outline: 'none',
    background: isDark ? 'rgba(17, 24, 39, 0.72)' : 'rgba(255, 255, 255, 0.8)',
    transition: 'border-color 180ms ease, box-shadow 180ms ease, background 180ms ease',
    fontSize: '13px',
  };

  const passwordInputStyle = {
    ...inputStyle,
    paddingRight: '52px',
  };

  const handleFocus = (e: React.FocusEvent<HTMLInputElement>) => {
    e.currentTarget.style.borderColor = isDark ? 'rgba(34, 211, 238, 0.45)' : 'rgba(37, 99, 235, 0.5)';
    e.currentTarget.style.background = isDark ? 'rgba(17, 24, 39, 0.95)' : 'rgba(255, 255, 255, 0.95)';
    e.currentTarget.style.boxShadow = isDark
      ? '0 0 0 3px rgba(34, 211, 238, 0.07), 0 0 25px rgba(34, 211, 238, 0.05)'
      : '0 0 0 3px rgba(37, 99, 235, 0.1), 0 0 25px rgba(37, 99, 235, 0.08)';
  };

  const handleBlur = (e: React.FocusEvent<HTMLInputElement>) => {
    e.currentTarget.style.borderColor = isDark ? 'rgba(148, 163, 184, 0.13)' : '1px solid rgba(0, 0, 0, 0.08)';
    e.currentTarget.style.background = isDark ? 'rgba(17, 24, 39, 0.72)' : 'rgba(255, 255, 255, 0.8)';
    e.currentTarget.style.boxShadow = 'none';
  };

  return (
    <div style={{ width: '100%', maxWidth: '360px', margin: 'auto' }}>
      {/* Login Header */}
      <div style={{ marginBottom: '30px' }}>
        <h2
          style={{
            fontSize: '25px',
            letterSpacing: '-0.8px',
            fontWeight: '750',
            color: isDark ? '#f1f7ff' : '#111827',
          }}
        >
          {t('auth.welcomeBack')}
        </h2>
        <p style={{ marginTop: '6px', color: isDark ? '#8290a7' : '#6B7280', fontSize: '12px' }}>
          {t('auth.signInToAccount')}
        </p>
      </div>

      {/* Error message */}
      {(error || externalError) && (
        <div
          style={{
            marginBottom: '18px',
            padding: '11px 12px',
            display: 'flex',
            gap: '10px',
            border: '1px solid rgba(251, 113, 133, 0.3)',
            borderRadius: '10px',
            background: 'rgba(251, 113, 133, 0.1)',
            color: 'var(--color-danger)',
            fontSize: '10px',
            lineHeight: '1.5',
          }}
        >
          <span style={{ color: 'var(--color-danger)', flexShrink: 0 }}>
            <AlertCircle style={{ width: '12px', height: '12px' }} />
          </span>
          <span>{externalError || error}</span>
        </div>
      )}

      {isLoading && (
        <div
          aria-live="polite"
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '12px',
            marginBottom: '18px',
            padding: '14px 16px',
            borderRadius: '12px',
            border: isDark ? '1px solid rgba(34, 211, 238, 0.22)' : '1px solid rgba(37, 99, 235, 0.2)',
            background: isDark ? 'rgba(14, 116, 144, 0.12)' : 'rgba(37, 99, 235, 0.08)',
            color: isDark ? '#d8f4ff' : '#1e3a8a',
            fontSize: '12px',
            fontWeight: '700',
          }}
        >
          <div
            style={{
              width: '16px',
              height: '16px',
              borderRadius: '9999px',
              border: '2px solid rgba(255,255,255,0.35)',
              borderTopColor: isDark ? '#67e8f9' : '#2563eb',
              animation: 'spin 0.85s linear infinite',
            }}
          />
          <span>جاري تسجيل الدخول...</span>
        </div>
      )}

      <form onSubmit={handleSubmit}>
        {/* Email Input */}
        <div style={{ marginBottom: '18px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
            <label
              style={{ color: isDark ? '#cbd5e1' : '#374151', fontSize: '11px', fontWeight: '650' }}
              htmlFor="email"
            >
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
            style={inputStyle}
            className="placeholder:text-[#4f5c70]"
            onFocus={handleFocus}
            onBlur={handleBlur}
          />
        </div>

        {/* Password Input */}
        <div style={{ marginBottom: '18px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
            <label
              style={{ color: isDark ? '#cbd5e1' : '#374151', fontSize: '11px', fontWeight: '650' }}
              htmlFor="password"
            >
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
              style={passwordInputStyle}
              className="placeholder:text-[#4f5c70]"
              onFocus={handleFocus}
              onBlur={handleBlur}
            />
            <button
              type="button"
              onClick={() => setShowPassword(!showPassword)}
              style={{
                position: 'absolute',
                top: '50%',
                insetInlineStart: '4px',
                transform: 'translateY(-50%)',
                width: '40px',
                height: '40px',
                minWidth: '40px',
                minHeight: '40px',
                padding: '0',
                color: isDark ? '#8290a7' : '#6B7280',
                background: 'transparent',
                border: 'none',
                cursor: 'pointer',
                transition: 'all 180ms ease',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                borderRadius: '8px',
                boxShadow: 'none',
              }}
              aria-label={showPassword ? 'إخفاء كلمة المرور' : 'إظهار كلمة المرور'}
            >
              {showPassword ? <EyeOff style={{ width: '16px', height: '16px' }} /> : <Eye style={{ width: '16px', height: '16px' }} />}
            </button>
          </div>
        </div>

        {/* Form Options */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', margin: '8px 0 22px' }}>
          <label
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              minHeight: '44px',
              color: isDark ? '#8290a7' : '#6B7280',
              fontSize: '11px',
              cursor: 'pointer',
            }}
          >
            <input
              type="checkbox"
              checked={rememberMe}
              onChange={(e) => setRememberMe(e.target.checked)}
              style={{
                width: '16px',
                height: '16px',
                minWidth: '16px',
                minHeight: '16px',
                flexShrink: 0,
                margin: 0,
                accentColor: isDark ? '#14b8a6' : '#2563EB',
              }}
            />
            {t('auth.rememberMe')}
          </label>
          <button
            type="button"
            onClick={() => setShowForgotPassword(true)}
            style={{
              color: isDark ? '#14b8a6' : '#2563EB',
              fontSize: '11px',
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              textDecoration: 'none',
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
            padding: '13px',
            color: 'var(--text-on-primary)',
            border: 'none',
            borderRadius: '10px',
            fontSize: '13px',
            fontWeight: '700',
            cursor: isLoading ? 'not-allowed' : 'pointer',
            background: 'linear-gradient(135deg, var(--primary) 0%, var(--primary-hover) 100%)',
            boxShadow: '0 4px 20px rgba(37, 99, 235, 0.3), 0 0 30px rgba(37, 99, 235, 0.15)',
            transition: 'all 200ms ease',
            opacity: isLoading ? 0.7 : 1,
          }}
        >
          {isLoading ? 'جاري تسجيل الدخول...' : t('auth.signIn')}
        </button>
      </form>

      {showForgotPassword && (
        <div
          role="presentation"
          onMouseDown={(event) => {
            if (event.target === event.currentTarget) setShowForgotPassword(false);
          }}
          style={{
            position: 'fixed',
            inset: 0,
            zIndex: 50,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            padding: '20px',
            background: 'rgba(2, 6, 23, 0.62)',
            backdropFilter: 'blur(8px)',
          }}
        >
          <div
            role="dialog"
            aria-modal="true"
            aria-labelledby="forgot-password-title"
            style={{
              width: 'min(100%, 430px)',
              position: 'relative',
              overflow: 'hidden',
              border: isDark ? '1px solid rgba(148, 163, 184, 0.18)' : '1px solid rgba(0, 0, 0, 0.1)',
              borderRadius: '18px',
              padding: '28px',
              background: isDark
                ? 'linear-gradient(145deg, rgba(17, 24, 39, 0.98), rgba(9, 14, 24, 0.98))'
                : 'linear-gradient(145deg, rgba(255, 255, 255, 0.99), rgba(249, 250, 251, 0.99))',
              boxShadow: '0 24px 80px rgba(0, 0, 0, 0.35)',
              color: isDark ? '#f1f7ff' : '#111827',
            }}
          >
            <button
              type="button"
              onClick={() => setShowForgotPassword(false)}
              aria-label="إغلاق"
              style={{
                position: 'absolute',
                top: '16px',
                insetInlineStart: '16px',
                width: '34px',
                height: '34px',
                display: 'grid',
                placeItems: 'center',
                border: 'none',
                borderRadius: '9px',
                background: isDark ? 'rgba(148, 163, 184, 0.1)' : 'rgba(15, 23, 42, 0.06)',
                color: isDark ? '#a8b4c7' : '#64748b',
                cursor: 'pointer',
              }}
            >
              <X size={17} />
            </button>

            <div style={{ display: 'flex', alignItems: 'flex-start', gap: '14px', paddingInlineStart: '4px' }}>
              <div
                style={{
                  width: '46px',
                  height: '46px',
                  flexShrink: 0,
                  display: 'grid',
                  placeItems: 'center',
                  borderRadius: '13px',
                  background: isDark ? 'rgba(20, 184, 166, 0.14)' : 'rgba(37, 99, 235, 0.1)',
                  color: isDark ? '#5eead4' : '#2563eb',
                }}
              >
                <Headset size={23} />
              </div>
              <div style={{ paddingTop: '2px' }}>
                <h2 id="forgot-password-title" style={{ margin: 0, fontSize: '19px', fontWeight: 750 }}>
                  نسيت كلمة المرور؟
                </h2>
                <p style={{ margin: '8px 0 0', color: isDark ? '#9aa8bc' : '#64748b', fontSize: '12px', lineHeight: 1.8 }}>
                  لا تقلق، يمكن للإدارة مساعدتك في استعادة الوصول إلى حسابك.
                </p>
              </div>
            </div>

            <div
              style={{
                marginTop: '22px',
                padding: '15px 16px',
                border: isDark ? '1px solid rgba(20, 184, 166, 0.2)' : '1px solid rgba(37, 99, 235, 0.16)',
                borderRadius: '12px',
                background: isDark ? 'rgba(20, 184, 166, 0.07)' : 'rgba(37, 99, 235, 0.05)',
                color: isDark ? '#d8f4ff' : '#1e3a8a',
                fontSize: '12px',
                lineHeight: 1.9,
              }}
            >
              تواصل مع مسؤول النظام أو الإدارة لتغيير كلمة المرور والتحقق من بيانات حسابك.
            </div>

            <button
              type="button"
              onClick={() => setShowForgotPassword(false)}
              style={{
                width: '100%',
                marginTop: '22px',
                padding: '12px',
                border: 'none',
                borderRadius: '10px',
                background: 'linear-gradient(135deg, var(--primary) 0%, var(--primary-hover) 100%)',
                color: 'var(--text-on-primary)',
                fontSize: '12px',
                fontWeight: 700,
                cursor: 'pointer',
                boxShadow: '0 4px 18px rgba(37, 99, 235, 0.24)',
              }}
            >
              فهمت
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
