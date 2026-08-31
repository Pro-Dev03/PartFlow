import { useNavigate } from 'react-router-dom';
import { AlertTriangle, LogOut, ShieldAlert } from 'lucide-react';
import { useAuthStore } from '../../../stores/authStore';

export default function SubscriptionExpiredPage() {
  const navigate = useNavigate();
  const logout = useAuthStore((state) => state.logout);

  const handleLogout = () => {
    logout();
    navigate('/login', { replace: true });
  };

  return (
    <div dir="rtl" className="min-h-screen flex items-center justify-center px-4" style={{ background: 'var(--bg-background)' }}>
      <div
        className="w-full max-w-xl rounded-2xl border"
        style={{
          background: 'var(--bg-surface)',
          borderColor: 'var(--border-color)',
          boxShadow: '0 24px 80px rgba(15, 23, 42, 0.18)',
          padding: '32px 28px',
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'center', marginBottom: '18px' }}>
          <div
            style={{
              width: '80px',
              height: '80px',
              borderRadius: '20px',
              background: 'rgba(239, 68, 68, 0.12)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: '#ef4444',
            }}
          >
            <ShieldAlert size={34} />
          </div>
        </div>

        <h1 style={{ textAlign: 'center', fontSize: '30px', fontWeight: 800, marginBottom: '12px', color: 'var(--color-text-primary)' }}>
          اشتراكك منتهي
        </h1>

        <p style={{ textAlign: 'center', lineHeight: 1.8, color: 'var(--color-text-secondary)', marginBottom: '18px' }}>
          تم إيقاف الوصول إلى النظام لأن الحساب غير نشط أو أن الاشتراك منتهي. لمتابعة استخدام PartFlow، يرجى التواصل مع الإدارة لتجديد الخدمة.
        </p>

        <div
          style={{
            border: '1px solid rgba(239, 68, 68, 0.18)',
            background: 'rgba(239, 68, 68, 0.05)',
            borderRadius: '14px',
            padding: '16px 18px',
            marginBottom: '24px',
            color: 'var(--color-text-primary)',
            textAlign: 'center',
            lineHeight: 1.8,
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '8px', marginBottom: '6px', fontWeight: 700 }}>
            <AlertTriangle size={18} />
            <span>تواصل مع الإدارة</span>
          </div>
          <div>
            تم تعطيل الحساب مؤقتًا حتى يتم تجديد الخدمة.
          </div>
        </div>

        <div style={{ display: 'flex', justifyContent: 'center' }}>
          <button
            type="button"
            onClick={handleLogout}
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: '8px',
              borderRadius: '12px',
              padding: '12px 24px',
              background: 'var(--color-primary)',
              color: '#fff',
              border: 'none',
              fontWeight: 700,
              cursor: 'pointer',
            }}
          >
            <LogOut size={18} />
            تسجيل الخروج
          </button>
        </div>
      </div>
    </div>
  );
}
