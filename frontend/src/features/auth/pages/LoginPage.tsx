import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from '../../../hooks/useTranslation';
import { useAuthStore } from '../../../stores/authStore';
import { Input } from '../../../components/ui/input';
import { Button } from '../../../components/ui/button';
import { Card, CardContent } from '../../../components/ui/card';
import { 
  Mail, 
  Lock, 
  Eye, 
  EyeOff, 
  ArrowRight,
  Sparkles,
  Shield,
  ChevronRight,
  Fingerprint,
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
    <div dir="rtl" className="min-h-screen flex items-center justify-center bg-bg p-4 relative overflow-hidden">
      {/* Futuristic Background */}
      <div className="absolute inset-0 bg-bg-gradient" />
      
      {/* Animated background elements */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        {/* Glowing orbs */}
        <div className="absolute top-20 right-20 w-96 h-96 bg-cyan/5 rounded-full blur-3xl animate-float" style={{ animationDelay: '0s' }} />
        <div className="absolute bottom-20 left-20 w-80 h-80 bg-blue/5 rounded-full blur-3xl animate-float" style={{ animationDelay: '1s' }} />
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] bg-cyan/3 rounded-full blur-3xl animate-glow-pulse" />
        
        {/* Grid pattern overlay */}
        <div className="absolute inset-0 opacity-[0.02]" style={{
          backgroundImage: `
            linear-gradient(rgba(34, 211, 238, 0.1) 1px, transparent 1px),
            linear-gradient(90deg, rgba(34, 211, 238, 0.1) 1px, transparent 1px)
          `,
          backgroundSize: '50px 50px'
        }} />
      </div>

      <div className="relative w-full max-w-md z-10">
        {/* Back button */}
        <button
          onClick={() => navigate('/')}
          className="absolute -top-16 left-0 flex items-center gap-sm text-text-muted hover:text-text transition-colors duration-normal group"
        >
          <ArrowRight className="w-4 h-4 group-hover:-translate-x-1 transition-transform duration-normal" />
          <span className="text-small">العودة</span>
        </button>

        {/* Main Card using Design System */}
        <Card variant="featured" className="animate-scale-in">
          <CardContent className="p-2xl">
            {/* Logo and Header */}
            <div className="text-center mb-2xl animate-slide-in-from-top">
              <div className="inline-flex items-center justify-center w-20 h-20 mb-xl bg-button-primary-gradient rounded-lg shadow-glow animate-float">
                <Sparkles className="w-10 h-10 text-cyan" />
              </div>
              <h1 className="text-h1 font-bold bg-gradient-to-r from-cyan to-blue bg-clip-text text-transparent mb-md">
                PartFlow
              </h1>
              <p className="text-text-muted text-body">
                نظام إدارة المخزون والمبيعات
              </p>
            </div>

            {/* Welcome message */}
            <div className="text-center mb-xl animate-slide-in-from-top" style={{ animationDelay: '0.1s' }}>
              <h2 className="text-h2 font-bold text-text mb-md">
                {t('auth.welcomeBack')}
              </h2>
              <p className="text-small text-text-muted">
                {t('auth.signInToAccount')}
              </p>
            </div>

            {/* Error message using Design System */}
            {error && (
              <div className="mb-xl bg-red/10 border border-red/30 rounded-sm p-lg flex items-start gap-md animate-fade-in">
                <div className="flex-shrink-0 w-6 h-6 bg-red/20 rounded-full flex items-center justify-center">
                  <AlertCircle className="w-3 h-3 text-red" />
                </div>
                <p className="text-small text-red flex-1">{error}</p>
              </div>
            )}

            <form onSubmit={handleSubmit} className="space-y-xl">
              {/* Email Input using Design System */}
              <div className="animate-slide-in-from-top" style={{ animationDelay: '0.2s' }}>
                <Input
                  type="email"
                  placeholder="البريد الإلكتروني"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  size="lg"
                  leftIcon={<Mail className="w-5 h-5" />}
                />
              </div>

              {/* Password Input using Design System */}
              <div className="animate-slide-in-from-top" style={{ animationDelay: '0.3s' }}>
                <Input
                  type={showPassword ? 'text' : 'password'}
                  placeholder="كلمة المرور"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  size="lg"
                  leftIcon={<Lock className="w-5 h-5" />}
                  rightIcon={
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="text-text-muted hover:text-cyan transition-colors duration-normal"
                    >
                      {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                    </button>
                  }
                />
              </div>

              {/* Remember me & Forgot password */}
              <div className="flex items-center justify-between animate-slide-in-from-top" style={{ animationDelay: '0.4s' }}>
                <label className="flex items-center gap-md cursor-pointer group">
                  <div className="relative">
                    <input
                      type="checkbox"
                      className="peer sr-only"
                    />
                    <div className="w-5 h-5 border-2 border-border rounded-md bg-surface peer-checked:bg-cyan/20 peer-checked:border-cyan transition-all duration-normal flex items-center justify-center">
                      <Fingerprint className="w-3 h-3 text-cyan opacity-0 peer-checked:opacity-100 transition-opacity duration-normal" />
                    </div>
                  </div>
                  <span className="text-small text-text-muted group-hover:text-text transition-colors duration-normal">
                    {t('auth.rememberMe')}
                  </span>
                </label>
                <button
                  type="button"
                  className="text-small text-cyan hover:text-cyan/80 transition-colors duration-normal flex items-center gap-sm group"
                >
                  {t('auth.forgotPassword')}
                  <ChevronRight className="w-4 h-4 group-hover:-translate-x-1 transition-transform duration-normal" />
                </button>
              </div>

              {/* Login Button using Design System */}
              <Button
                type="submit"
                variant="primary"
                size="lg"
                fullWidth
                isLoading={isLoading}
                className="animate-slide-in-from-top"
                style={{ animationDelay: '0.5s' }}
              >
                <span className="flex items-center gap-md">
                  {t('auth.login')}
                  <Shield className="w-5 h-5" />
                </span>
              </Button>
            </form>

            {/* Footer with security badge */}
            <div className="mt-2xl pt-xl border-t border-border/30 text-center animate-slide-in-from-top" style={{ animationDelay: '0.6s' }}>
              <div className="flex items-center justify-center gap-md text-tiny text-text-muted">
                <Fingerprint className="w-3 h-3 text-cyan/60" />
                <span>نظام آمن ومحمي بتشفير متقدم</span>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Additional info */}
        <div className="mt-xl text-center animate-slide-in-from-top" style={{ animationDelay: '0.7s' }}>
          <p className="text-small text-text-muted">
            ليس لديك حساب؟{' '}
            <button className="text-cyan hover:text-cyan/80 font-medium transition-colors duration-normal flex items-center gap-sm inline-flex group">
              تواصل مع الإدارة
              <ChevronRight className="w-4 h-4 group-hover:-translate-x-1 transition-transform duration-normal" />
            </button>
          </p>
        </div>
      </div>
    </div>
  );
}