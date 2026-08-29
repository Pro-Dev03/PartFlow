import { useEffect, useId, useState } from 'react'

export default function NeonAIBot({ size = 64 }: { size?: number }) {
  const gradientId = useId().replace(/:/g, '')
  const botBgId = `bot-bg-${gradientId}`
  const botFaceId = `bot-face-${gradientId}`
  const [blink, setBlink] = useState(false)
  const [pulse, setPulse] = useState(false)
  const [float, setFloat] = useState(false)
  const [wave, setWave] = useState(false)

  useEffect(() => {
    const blinkTimer = setInterval(() => {
      setBlink(true)
      window.setTimeout(() => setBlink(false), 180)
    }, 3200)

    const pulseTimer = setInterval(() => {
      setPulse(true)
      window.setTimeout(() => setPulse(false), 1200)
    }, 2500)

    const floatTimer = setInterval(() => {
      setFloat(true)
      window.setTimeout(() => setFloat(false), 700)
    }, 4200)

    const waveTimer = setInterval(() => {
      setWave(true)
      window.setTimeout(() => setWave(false), 1100)
    }, 7000)

    return () => {
      clearInterval(blinkTimer)
      clearInterval(pulseTimer)
      clearInterval(floatTimer)
      clearInterval(waveTimer)
    }
  }, [])

  return (
    <div
      style={{
        width: size,
        height: size,
        transform: float ? 'translateY(-3px)' : 'translateY(0px)',
        transition: 'transform 0.55s ease-out',
      }}
      className="relative flex items-center justify-center"
    >
      <div
        className="absolute inset-0 rounded-full"
        style={{
          background: `radial-gradient(circle, rgba(14,165,233,0.22) 0%, rgba(14,165,233,0.06) 38%, transparent 75%)`,
          transform: pulse ? 'scale(1.18)' : 'scale(1)',
          opacity: pulse ? 1 : 0.7,
          transition: 'all 0.8s ease-out',
        }}
      />

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

        <rect x="26" y="24" width="48" height="32" rx="12" fill={`url(#${botFaceId})`} stroke="#dbeafe" strokeWidth="2" />
        <rect x="33" y="30" width="34" height="18" rx="7" fill="#0f172a" opacity="0.94" />

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

        <path d="M42 46 Q50 51 58 46" fill="none" stroke="#67e8f9" strokeWidth="2" strokeLinecap="round" />

        <path d="M35 56 Q50 63 65 56 L68 74 Q50 80 32 74 Z" fill={`url(#${botFaceId})`} stroke="#dbeafe" strokeWidth="2" />
        <circle cx="50" cy="65" r="4" fill="#0f172a" />
        <circle cx="50" cy="65" r="1.6" fill="#67e8f9" />

        <g transform={wave ? 'rotate(-22 68 60)' : 'rotate(0 68 60)'} style={{ transformOrigin: '68px 60px', transition: 'transform 0.35s ease-out' }}>
          <path d="M68 60 Q78 57 82 64" fill="none" stroke="#dbeafe" strokeWidth="3" strokeLinecap="round" />
        </g>

        <path d="M22 63 Q32 58 35 68" fill="none" stroke="#dbeafe" strokeWidth="3" strokeLinecap="round" />
      </svg>
    </div>
  )
}
