import { useState, useEffect } from 'react';

interface LoginBackgroundProps {
  isDark: boolean;
}

export function LoginBackground({ isDark }: LoginBackgroundProps) {
  return (
    <>
      {/* Futuristic Background */}
      <div
        style={{
          position: 'absolute',
          inset: 0,
          background: isDark
            ? `radial-gradient(circle at 80% 0%, rgba(34, 211, 238, 0.11), transparent 30%), radial-gradient(circle at 15% 85%, rgba(59, 130, 246, 0.08), transparent 32%), var(--bg-background)`
            : 'radial-gradient(circle at 80% 0%, rgba(37, 99, 235, 0.05), transparent 30%), radial-gradient(circle at 15% 85%, rgba(37, 99, 235, 0.03), transparent 32%), var(--bg-background)',
        }}
      />

      {/* Subtle futuristic grid */}
      <div
        style={{
          position: 'absolute',
          inset: 0,
          pointerEvents: 'none',
          backgroundImage: `linear-gradient(rgba(148, 163, 184, 0.025) 1px, transparent 1px), linear-gradient(90deg, rgba(148, 163, 184, 0.025) 1px, transparent 1px)`,
          backgroundSize: '45px 45px',
          maskImage: 'linear-gradient(to bottom, black, transparent 85%)',
          WebkitMaskImage: 'linear-gradient(to bottom, black, transparent 85%)',
        }}
      />

      {/* Ambient glow orbs */}
      <div
        style={{
          position: 'absolute',
          top: '-220px',
          right: '-120px',
          width: '420px',
          height: '420px',
          borderRadius: '50%',
          background: isDark
            ? 'radial-gradient(circle, rgba(34, 211, 238, 0.08), transparent 68%)'
            : 'radial-gradient(circle, rgba(37, 99, 235, 0.04), transparent 68%)',
          filter: 'blur(20px)',
          pointerEvents: 'none',
        }}
      />
      <div
        style={{
          position: 'absolute',
          bottom: '-260px',
          left: '-180px',
          width: '420px',
          height: '420px',
          borderRadius: '50%',
          background: isDark
            ? 'radial-gradient(circle, rgba(59, 130, 246, 0.07), transparent 68%)'
            : 'radial-gradient(circle, rgba(37, 99, 235, 0.03), transparent 68%)',
          filter: 'blur(20px)',
          pointerEvents: 'none',
        }}
      />
    </>
  );
}
