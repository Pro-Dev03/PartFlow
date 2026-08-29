import { useState } from 'react';
import { useTranslation } from '../../../hooks/useTranslation';
import { Eye, EyeOff, AlertCircle } from 'lucide-react';

interface LoginFormProps {
  isDark: boolean;
  isLoading: boolean;
  onSubmit: (email: string, password: string) => Promise<void>;
}

export function LoginForm({ isDark, isLoading, onSubmit }: LoginFormProps) {
  const { t } = useTranslation();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [rememberMe, setRememberMe] = useState(false);

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
      {error && (
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
          <span>{error}</span>
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
    </div>
  );
}
