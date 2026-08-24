interface BrandPanelProps {
  isDark: boolean;
}

export function BrandPanel({ isDark }: BrandPanelProps) {
  return (
    <div
      style={{
        position: 'relative',
        minHeight: '590px',
        padding: '46px',
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'space-between',
        borderRight: isDark ? '1px solid rgba(148, 163, 184, 0.13)' : '1px solid rgba(0, 0, 0, 0.08)',
        background: isDark
          ? 'linear-gradient(145deg, rgba(34, 211, 238, 0.055), transparent 45%)'
          : 'linear-gradient(145deg, rgba(37, 99, 235, 0.03), transparent 45%)',
      }}
      className="md:block hidden"
    >
      {/* Brand */}
      <div style={{ display: 'flex', alignItems: 'center', gap: '13px' }}>
        <div
          style={{
            width: '44px',
            height: '44px',
            display: 'grid',
            placeItems: 'center',
            borderRadius: '13px',
            color: isDark ? '#22d3ee' : '#2563EB',
            border: isDark ? '1px solid rgba(34, 211, 238, 0.25)' : '1px solid rgba(37, 99, 235, 0.25)',
            background: isDark
              ? 'linear-gradient(145deg, rgba(34, 211, 238, 0.13), rgba(59, 130, 246, 0.06))'
              : 'linear-gradient(145deg, rgba(37, 99, 235, 0.08), rgba(37, 99, 235, 0.04))',
            boxShadow: isDark ? '0 0 30px rgba(34, 211, 238, 0.10)' : '0 0 30px rgba(37, 99, 235, 0.08)',
          }}
        >
          <span style={{ fontSize: '19px', fontWeight: '800', letterSpacing: '-1px' }}>PF</span>
        </div>
        <div>
          <div
            style={{
              fontSize: '17px',
              fontWeight: '750',
              letterSpacing: '-0.4px',
              color: isDark ? '#f1f7ff' : '#111827',
            }}
          >
            PartFlow
          </div>
          <div
            style={{
              fontSize: '10px',
              letterSpacing: '1.2px',
              textTransform: 'uppercase',
              color: isDark ? '#8290a7' : '#6B7280',
              marginTop: '-2px',
            }}
          >
            نظام إدارة المتاجر
          </div>
        </div>
      </div>

      {/* Hero Copy */}
      <div style={{ maxWidth: '340px' }}>
        <div
          style={{
            marginBottom: '14px',
            color: isDark ? '#22d3ee' : '#2563EB',
            fontSize: '10px',
            fontWeight: '700',
            textTransform: 'uppercase',
            letterSpacing: '2px',
          }}
        >
          نظام إدارة ذكي
        </div>
        <h1
          style={{
            fontSize: 'clamp(30px, 4vw, 42px)',
            lineHeight: '1.08',
            letterSpacing: '-1.8px',
            fontWeight: '800',
            color: isDark ? '#f1f7ff' : '#111827',
          }}
        >
          متجرك.
          <br />
          <span
            style={{
              color: isDark ? '#22d3ee' : '#2563EB',
              textShadow: isDark ? '0 0 25px rgba(34, 211, 238, 0.15)' : 'none',
            }}
          >
            تحت سيطرتك.
          </span>
        </h1>
        <p
          style={{
            marginTop: '18px',
            color: isDark ? '#8290a7' : '#6B7280',
            fontSize: '13px',
            lineHeight: '1.8',
          }}
        >
          إدارة المخزون، المبيعات، العملاء، الديون والعمليات اليومية من مساحة عمل ذكية واحدة.
        </p>
      </div>

      {/* Brand Footer */}
      <div
        style={{
          color: isDark ? '#56647a' : '#9CA3AF',
          fontSize: '10px',
          letterSpacing: '1px',
          textTransform: 'uppercase',
        }}
      >
        نظام PartFlow
      </div>
    </div>
  );
}
