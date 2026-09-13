import { CloudCog, LockKeyhole, ShieldCheck } from 'lucide-react';
import { PartFlowLogo } from '../../../components/branding/PartFlowLogo';

export function SubscriptionVerificationScreen() {
  return (
    <div
      dir="rtl"
      role="status"
      aria-live="polite"
      aria-busy="true"
      style={{
        minHeight: '100vh',
        display: 'grid',
        placeItems: 'center',
        padding: '24px',
        color: '#e5eef7',
        background: 'linear-gradient(145deg, #08121d 0%, #0b1d2b 55%, #102b37 100%)',
        overflow: 'hidden',
        position: 'relative',
      }}
    >
      <style>{`
        @keyframes pf-verification-pulse {
          0%, 100% { transform: scale(0.96); opacity: 0.7; }
          50% { transform: scale(1.04); opacity: 1; }
        }
        @keyframes pf-verification-spin {
          to { transform: rotate(360deg); }
        }
        @keyframes pf-verification-progress {
          0% { transform: translateX(160%); }
          100% { transform: translateX(-220%); }
        }
        @keyframes pf-verification-rise {
          from { opacity: 0; transform: translateY(12px); }
          to { opacity: 1; transform: translateY(0); }
        }
      `}</style>

      <div
        aria-hidden="true"
        style={{
          position: 'absolute',
          width: 'min(720px, 95vw)',
          height: 'min(720px, 95vw)',
          borderRadius: '50%',
          border: '1px solid rgba(103, 232, 249, 0.12)',
          boxShadow: '0 0 0 80px rgba(103, 232, 249, 0.025), 0 0 0 160px rgba(103, 232, 249, 0.018)',
          animation: 'pf-verification-pulse 5s ease-in-out infinite',
        }}
      />

      <div
        style={{
          width: 'min(520px, 100%)',
          padding: '42px 34px 30px',
          textAlign: 'center',
          position: 'relative',
          zIndex: 1,
          borderRadius: '26px',
          background: 'rgba(11, 29, 43, 0.88)',
          border: '1px solid rgba(148, 163, 184, 0.2)',
          boxShadow: '0 28px 90px rgba(0, 0, 0, 0.38), inset 0 1px 0 rgba(255, 255, 255, 0.06)',
          backdropFilter: 'blur(18px)',
          animation: 'pf-verification-rise 500ms ease-out both',
        }}
      >
        <div
          style={{
            width: '92px',
            height: '92px',
            margin: '0 auto 28px',
            borderRadius: '28px',
            display: 'grid',
            placeItems: 'center',
            color: '#67e8f9',
            background: 'linear-gradient(145deg, rgba(34, 211, 238, 0.2), rgba(14, 116, 144, 0.14))',
            border: '1px solid rgba(103, 232, 249, 0.28)',
            boxShadow: '0 20px 60px rgba(34, 211, 238, 0.16)',
          }}
        >
          <PartFlowLogo size={66} priority />
        </div>

        <p style={{ margin: '0 0 10px', color: '#67e8f9', fontSize: '12px', fontWeight: 800, letterSpacing: '0.16em' }}>
          PARTFLOW · ACCESS CONTROL
        </p>
        <h1 style={{ margin: '0', color: '#f8fbff', fontSize: 'clamp(26px, 5vw, 38px)', lineHeight: 1.25, fontWeight: 800 }}>
          نتحقق من اشتراكك
        </h1>
        <p style={{ maxWidth: '390px', margin: '14px auto 0', color: '#9fb2c5', fontSize: '14px', lineHeight: 1.9 }}>
          نتحقق من صلاحية اشتراكك قبل فتح مساحة العمل الخاصة بك.
        </p>

        <div style={{ display: 'flex', justifyContent: 'center', marginTop: '20px' }}>
          <span style={{ padding: '6px 12px', borderRadius: '999px', color: '#8af3cf', background: 'rgba(52, 211, 153, 0.1)', border: '1px solid rgba(52, 211, 153, 0.24)', fontSize: '11px', fontWeight: 750 }}>
            اشتراكك قيد التحقق
          </span>
        </div>

        <div
          style={{
            marginTop: '34px',
            padding: '17px 18px',
            display: 'flex',
            alignItems: 'center',
            gap: '14px',
            textAlign: 'right',
            borderRadius: '16px',
            background: 'rgba(7, 20, 31, 0.62)',
            border: '1px solid rgba(148, 163, 184, 0.16)',
          }}
        >
          <span
            aria-hidden="true"
            style={{
              width: '28px',
              height: '28px',
              flexShrink: 0,
              display: 'grid',
              placeItems: 'center',
              borderRadius: '50%',
              color: '#67e8f9',
              border: '2px solid rgba(103, 232, 249, 0.25)',
              borderTopColor: '#67e8f9',
              animation: 'pf-verification-spin 850ms linear infinite',
            }}
          />
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', color: '#e5eef7', fontSize: '13px', fontWeight: 750 }}>
              <CloudCog size={17} color="#67e8f9" />
              جارٍ الاتصال بخدمة الاشتراكات
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '7px', marginTop: '5px', color: '#7890a5', fontSize: '11px' }}>
              <LockKeyhole size={13} />
              التحقق مشفّر ولا يتم فتح النظام قبل اكتماله
            </div>
          </div>
        </div>

        <div style={{ height: '4px', marginTop: '24px', overflow: 'hidden', borderRadius: '999px', background: 'rgba(148, 163, 184, 0.14)' }}>
          <div style={{ width: '58%', height: '100%', borderRadius: 'inherit', background: 'linear-gradient(90deg, #22d3ee, #67e8f9)', animation: 'pf-verification-progress 2.2s ease-in-out infinite' }} />
        </div>
        <p style={{ margin: '12px 0 0', color: '#678096', fontSize: '11px' }}>
          يتم تأمين بيانات حسابك أثناء التحقق
        </p>
      </div>
    </div>
  );
}