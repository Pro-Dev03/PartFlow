import { useEffect, useId, useMemo, useState } from 'react'

type FaceState = 'idle' | 'sleepy' | 'hello' | 'thinking' | 'surprise' | 'focus' | 'joy' | 'confidence' | 'creativity' | 'energy'
type AvatarState = 'idle' | 'greeting' | 'listening' | 'thinking' | 'speaking' | 'attention' | 'error' | 'offline'

const faceStateByAvatarState: Record<AvatarState, FaceState> = {
  idle: 'idle',
  greeting: 'hello',
  listening: 'focus',
  thinking: 'thinking',
  speaking: 'joy',
  attention: 'surprise',
  error: 'surprise',
  offline: 'sleepy',
};

export default function NeonAIBot({ size = 64, state = 'idle' }: { size?: number; state?: AvatarState }) {
  const gradientId = useId().replace(/:/g, '')
  const botBgId = `bot-bg-${gradientId}`
  const botFaceId = `bot-face-${gradientId}`
  
  // Animation states
  const [blink, setBlink] = useState(false)
  const faceState = faceStateByAvatarState[state];
  const isSleepy = state === 'offline';
  const isWaving = state === 'greeting';
  const pulse = state === 'thinking' || state === 'attention' || state === 'error';
  const float = state === 'greeting' || state === 'speaking';
  const wave = state === 'greeting';

  // Particle positions
  const particles = useMemo(() => 
    Array.from({ length: 8 }).map((_, i) => ({
      id: i,
      x: (i * 37 + 11) % 100,
      y: (i * 61 + 23) % 100,
      size: (i % 3) + 1,
      delay: (i * 1.3) % 5,
      duration: (i % 3) + 4,
    })), []
  )

  // Basic animations (blink, pulse, float)
  useEffect(() => {
    if (state !== 'idle') return;
    const blinkTimer = setInterval(() => {
      if (faceState === 'idle') {
        setBlink(true)
        window.setTimeout(() => setBlink(false), 180)
      }
    }, 3200)

    return () => {
      clearInterval(blinkTimer)
    }
  }, [faceState, state])

  // Eye rendering based on state
  const renderEyes = () => {
    if (isSleepy) {
      return (
        <>
          <path d="M38 39 Q42 42 46 39" fill="none" stroke="#67e8f9" strokeWidth="2" strokeLinecap="round" />
          <path d="M54 39 Q58 42 62 39" fill="none" stroke="#67e8f9" strokeWidth="2" strokeLinecap="round" />
        </>
      )
    }

    switch (faceState) {
      case 'thinking':
        // Eyes looking up
        return (
          <>
            <circle cx="42" cy="37" r="3.2" fill="#67e8f9" />
            <circle cx="58" cy="37" r="3.2" fill="#67e8f9" />
            <circle cx="43" cy="36" r="1.2" fill="#fff" />
            <circle cx="59" cy="36" r="1.2" fill="#fff" />
          </>
        )
      
      case 'surprise':
        // Big round eyes
        return (
          <>
            <circle cx="42" cy="39" r="4.5" fill="#67e8f9" />
            <circle cx="58" cy="39" r="4.5" fill="#67e8f9" />
            <circle cx="43" cy="38" r="1.5" fill="#fff" />
            <circle cx="59" cy="38" r="1.5" fill="#fff" />
          </>
        )
      
      case 'focus':
        // Focused narrow eyes
        return (
          <>
            <ellipse cx="42" cy="39" rx="3" ry="2" fill="#67e8f9" />
            <ellipse cx="58" cy="39" rx="3" ry="2" fill="#67e8f9" />
            {/* Eyebrows */}
            <path d="M37 34 L47 35" fill="none" stroke="#67e8f9" strokeWidth="1.5" strokeLinecap="round" />
            <path d="M53 35 L63 34" fill="none" stroke="#67e8f9" strokeWidth="1.5" strokeLinecap="round" />
          </>
        )
      
      case 'joy':
        // Happy arc eyes
        return (
          <>
            <path d="M38 41 Q42 36 46 41" fill="none" stroke="#67e8f9" strokeWidth="2" strokeLinecap="round" />
            <path d="M54 41 Q58 36 62 41" fill="none" stroke="#67e8f9" strokeWidth="2" strokeLinecap="round" />
          </>
        )
      
      case 'confidence':
        // Determined eyes
        return (
          <>
            <ellipse cx="42" cy="39" rx="3.2" ry="2.8" fill="#67e8f9" />
            <ellipse cx="58" cy="39" rx="3.2" ry="2.8" fill="#67e8f9" />
            {/* Determined eyebrows */}
            <path d="M37 35 L47 33" fill="none" stroke="#67e8f9" strokeWidth="2" strokeLinecap="round" />
            <path d="M53 33 L63 35" fill="none" stroke="#67e8f9" strokeWidth="2" strokeLinecap="round" />
          </>
        )
      
      case 'creativity':
        // One eye winking
        return (
          <>
            <circle cx="42" cy="39" r="3.2" fill="#67e8f9" />
            <path d="M54 39 Q58 42 62 39" fill="none" stroke="#67e8f9" strokeWidth="2" strokeLinecap="round" />
          </>
        )
      
      case 'energy':
        // Sharp energetic eyes
        return (
          <>
            <polygon points="42,35 46,39 42,43 38,39" fill="#67e8f9" />
            <polygon points="58,35 62,39 58,43 54,39" fill="#67e8f9" />
          </>
       )
      
      default:
        // Normal eyes
        return (
          <>
            <circle
              cx="42"
              cy="39"
              r="3.2"
              fill="#67e8f9"
              transform={blink ? 'scale(0.15, 0.15)' : 'scale(1,1)'}
              style={{ transformOrigin: '42px 39px', transition: 'transform 0.12s ease-out' }}
            />
            <circle
              cx="58"
              cy="39"
              r="3.2"
              fill="#67e8f9"
              transform={blink ? 'scale(0.15, 0.15)' : 'scale(1,1)'}
              style={{ transformOrigin: '58px 39px', transition: 'transform 0.12s ease-out' }}
            />
          </>
        )
    }
  }

  // Mouth rendering based on state
  const renderMouth = () => {
    if (isSleepy) {
      return <circle cx="50" cy="47" r="2.5" fill="none" stroke="#67e8f9" strokeWidth="1.5" />
    }

    switch (faceState) {
      case 'thinking':
        return <path d="M45 47 L55 47" fill="none" stroke="#67e8f9" strokeWidth="2" strokeLinecap="round" />
      
      case 'surprise':
        return <ellipse cx="50" cy="48" rx="4" ry="5" fill="#0f172a" stroke="#67e8f9" strokeWidth="1.5" />
      
      case 'focus':
        return <path d="M44 47 L56 47" fill="none" stroke="#67e8f9" strokeWidth="2.5" strokeLinecap="round" />
      
      case 'joy':
        return <path d="M42 44 Q50 53 58 44" fill="none" stroke="#67e8f9" strokeWidth="2" strokeLinecap="round" />
      
      case 'confidence':
        return <path d="M43 46 L57 46" fill="none" stroke="#67e8f9" strokeWidth="2.5" strokeLinecap="round" />
      
      case 'creativity':
        return <path d="M44 47 Q50 44 56 47" fill="none" stroke="#67e8f9" strokeWidth="2" strokeLinecap="round" />
      
      case 'energy':
        return <path d="M42 45 Q50 52 58 45" fill="none" stroke="#67e8f9" strokeWidth="2.5" strokeLinecap="round" />
      
      default:
        return <path d="M44 46 Q50 50 56 46" fill="none" stroke="#67e8f9" strokeWidth="2" strokeLinecap="round" />
    }
  }

  // Extra face elements (blush, sweat, etc.)
  const renderFaceExtras = () => {
    switch (faceState) {
      case 'joy':
        return (
          <>
            <circle cx="36" cy="42" r="3" fill="#f472b6" opacity="0.3" />
            <circle cx="64" cy="42" r="3" fill="#f472b6" opacity="0.3" />
          </>
        )
      
      case 'surprise':
        return (
          <>
            <circle cx="50" cy="28" r="2" fill="#fbbf24" className="sparkle-star" />
            <circle cx="30" cy="32" r="1.5" fill="#fbbf24" className="sparkle-star" style={{ animationDelay: '0.2s' }} />
            <circle cx="70" cy="32" r="1.5" fill="#fbbf24" className="sparkle-star" style={{ animationDelay: '0.4s' }} />
          </>
        )
      
      case 'creativity':
        return (
          <g className="lightbulb">
            <circle cx="50" cy="22" r="6" fill="#fbbf24" opacity="0.8" />
            <path d="M47 26 L53 26" fill="none" stroke="#0f172a" strokeWidth="1.5" />
            <path d="M47 28 L53 28" fill="none" stroke="#0f172a" strokeWidth="1.5" />
            <circle cx="50" cy="22" r="8" fill="#fbbf24" opacity="0.2" className="bulb-glow" />
          </g>
        )
      
      case 'energy':
        return (
          <>
            <path d="M25 25 L30 30" stroke="#fbbf24" strokeWidth="2" className="energy-line" />
            <path d="M75 25 L70 30" stroke="#fbbf24" strokeWidth="2" className="energy-line" style={{ animationDelay: '0.1s' }} />
            <path d="M20 35 L28 35" stroke="#fbbf24" strokeWidth="2" className="energy-line" style={{ animationDelay: '0.2s' }} />
            <path d="M72 35 L80 35" stroke="#fbbf24" strokeWidth="2" className="energy-line" style={{ animationDelay: '0.3s' }} />
          </>
        )
      
      default:
        return null
    }
  }

  return (
    <div
      style={{
        width: size,
        height: size,
        transform: float ? 'translateY(-3px)' : 'translateY(0px)',
        transition: 'transform 0.55s ease-out',
        position: 'relative',
        animation: state === 'idle'
          ? 'bot-breathe 4.5s ease-in-out infinite'
          : state === 'thinking'
            ? 'bot-think 1.8s ease-in-out infinite'
            : state === 'speaking'
              ? 'bot-speak 0.9s ease-in-out infinite'
              : state === 'attention'
                ? 'bot-attention 0.8s ease-in-out 2'
                : state === 'error'
                  ? 'bot-error 0.6s ease-in-out 2'
                  : 'none',
      }}
      className="relative flex items-center justify-center"
    >
      {/* Outer Glow Pulse */}
      <div
        className="absolute inset-0 rounded-full"
        style={{
          background: `radial-gradient(circle, rgba(14,165,233,0.22) 0%, rgba(14,165,233,0.06) 38%, transparent 75%)`,
          transform: pulse ? 'scale(1.18)' : 'scale(1)',
          opacity: pulse ? 1 : 0.7,
          transition: 'all 0.8s ease-out',
        }}
      />

      {/* Particles */}
      {particles.map((p) => (
        <div
          key={p.id}
          className="absolute rounded-full pointer-events-none"
          style={{
            width: p.size,
            height: p.size,
            left: `${p.x}%`,
            top: `${p.y}%`,
            background: 'rgba(103, 232, 249, 0.6)',
            boxShadow: '0 0 4px rgba(103, 232, 249, 0.8)',
            animation: `particle-float ${p.duration}s ease-in-out ${p.delay}s infinite`,
          }}
        />
      ))}

      {/* SVG Bot */}
      <svg viewBox="0 0 100 100" className="relative z-10 h-full w-full drop-shadow-[0_8px_18px_rgba(14,165,233,0.18)]">
        <defs>
          <linearGradient id={botBgId} x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stopColor="#38bdf8" />
            <stop offset="100%" stopColor="#0ea5e9" />
          </linearGradient>
          <linearGradient id={botFaceId} x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stopColor="#f8fbff" />
            <stop offset="100%" stopColor="#dbeafe" />
          </linearGradient>
        </defs>

        <circle cx="50" cy="50" r="42" fill={`url(#${botBgId})`} opacity={pulse ? 0.15 : 0.1} />
        <circle cx="50" cy="50" r="30" fill={`url(#${botBgId})`} opacity={pulse ? 0.2 : 0.12} />

        {/* Head */}
        <rect x="26" y="24" width="48" height="32" rx="12" fill={`url(#${botFaceId})`} stroke="#dbeafe" strokeWidth="2" />
        
        {/* Face screen */}
        <rect x="33" y="30" width="34" height="18" rx="7" fill="#0f172a" opacity="0.94" />

        {/* Eyes - Dynamic */}
        {renderEyes()}

        {/* Mouth - Dynamic */}
        {renderMouth()}

        {/* Extra face elements */}
        {renderFaceExtras()}

        {/* Body */}
        <path d="M35 56 Q50 63 65 56 L68 74 Q50 80 32 74 Z" fill={`url(#${botFaceId})`} stroke="#dbeafe" strokeWidth="2" />
        <circle cx="50" cy="65" r="4" fill="#0f172a" />
        <circle cx="50" cy="65" r="1.6" fill="#67e8f9" />

        {/* Right Hand */}
        <g 
          transform={isWaving ? 'rotate(-30 68 60)' : (wave && faceState === 'idle' ? 'rotate(-22 68 60)' : 'rotate(0 68 60)')} 
          style={{ 
            transformOrigin: '68px 60px', 
            transition: 'transform 0.3s ease-out',
          }}
        >
          <path 
            d={isWaving ? "M68 60 Q78 50 82 42" : "M68 60 Q78 57 82 64"} 
            fill="none" 
            stroke="#dbeafe" 
            strokeWidth="3" 
            strokeLinecap="round"
          />
          {isWaving && (
            <>
              <circle cx="82" cy="40" r="2" fill="#dbeafe" />
              <line x1="82" y1="42" x2="84" y2="38" stroke="#dbeafe" strokeWidth="1.5" strokeLinecap="round" />
            </>
          )}
        </g>

        {/* Left Hand */}
        <path 
          d="M22 63 Q32 58 35 68" 
          fill="none" 
          stroke="#dbeafe" 
          strokeWidth="3" 
          strokeLinecap="round"
        />

        {/* Sleepy sweat drop */}
        {isSleepy && (
          <path 
            d="M72 28 Q74 34 72 36 Q70 34 72 28" 
            fill="#67e8f9"
            opacity="0.7"
            style={{ animation: 'drop-fall 1s ease-in infinite' }}
          />
        )}
      </svg>

      {/* Inline Styles */}
      <style>{`
        @keyframes particle-float {
          0%, 100% { transform: translateY(0) translateX(0) scale(1); opacity: 0.3; }
          25% { transform: translateY(-8px) translateX(5px) scale(1.2); opacity: 0.8; }
          50% { transform: translateY(-15px) translateX(-3px) scale(0.8); opacity: 0.5; }
          75% { transform: translateY(-8px) translateX(-5px) scale(1.1); opacity: 0.7; }
        }

        @keyframes bubble-pop {
          0% { transform: translateX(-50%) scale(0) translateY(10px); opacity: 0; }
          15% { transform: translateX(-50%) scale(1.1) translateY(0); opacity: 1; }
          25% { transform: translateX(-50%) scale(1) translateY(0); opacity: 1; }
          80% { transform: translateX(-50%) scale(1) translateY(0); opacity: 1; }
          100% { transform: translateX(-50%) scale(0.8) translateY(-10px); opacity: 0; }
        }

        @keyframes drop-fall {
          0% { transform: translateY(0); opacity: 0.7; }
          50% { transform: translateY(5px); opacity: 0.3; }
          100% { transform: translateY(10px); opacity: 0; }
        }

        @keyframes sparkle {
          0%, 100% { transform: scale(0); opacity: 0; }
          50% { transform: scale(1); opacity: 1; }
        }

        @keyframes glow-pulse {
          0%, 100% { transform: scale(1); opacity: 0.2; }
          50% { transform: scale(1.3); opacity: 0.5; }
        }

        @keyframes energy-zap {
          0%, 100% { opacity: 0; }
          50% { opacity: 1; }
        }

        @keyframes bot-breathe {
          0%, 100% { transform: translateY(0) scale(1); }
          50% { transform: translateY(-1px) scale(1.015); }
        }

        @keyframes bot-think {
          0%, 100% { transform: translateY(0) rotate(0deg); }
          50% { transform: translateY(-2px) rotate(-2deg); }
        }

        @keyframes bot-speak {
          0%, 100% { transform: translateY(0); }
          50% { transform: translateY(-2px); }
        }

        @keyframes bot-attention {
          0%, 100% { transform: scale(1); }
          50% { transform: scale(1.06); }
        }

        @keyframes bot-error {
          0%, 100% { transform: translateX(0); }
          35% { transform: translateX(-2px); }
          70% { transform: translateX(2px); }
        }

        .sparkle-star {
          animation: sparkle 0.8s ease-in-out infinite;
        }

        .bulb-glow {
          animation: glow-pulse 1s ease-in-out infinite;
        }

        .energy-line {
          animation: energy-zap 0.3s ease-in-out infinite;
        }
      `}</style>
    </div>
  )
}
