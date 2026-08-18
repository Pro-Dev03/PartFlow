/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // Semantic colors
        primary: {
          DEFAULT: '#3b82f6',
          50: '#eff6ff',
          100: '#dbeafe',
          500: '#3b82f6',
          600: '#2563eb',
        },
        success: {
          DEFAULT: '#10b981',
          50: '#ecfdf5',
          500: '#10b981',
          600: '#059669',
        },
        warning: {
          DEFAULT: '#f97316',
          50: '#fff7ed',
          500: '#f97316',
          600: '#ea580c',
        },
        danger: {
          DEFAULT: '#ef4444',
          50: '#fef2f2',
          500: '#ef4444',
          600: '#dc2626',
        },
        // PartFlow dark theme
        'bg-dark': '#0E1116',
        'surface-dark': '#151A21',
        'surface-elevated': '#1C222B',
        'border-dark': '#262E39',
        'text-dark': '#E8ECF1',
        'text-dim': '#8B96A5',
        'text-faint': '#5C6675',
        accent: '#00D9A3',
        // Light theme
        surface: '#ffffff',
        background: '#f6f7f9',
        text: '#111827',
        'text-secondary': '#4b5563',
        'text-tertiary': '#9ca3af',
        border: '#e5e7eb',
        // Legacy colors
        seal: '#B8863E',
        'seal-dark': '#D4A858',
      },
      fontFamily: {
        sans: ['IBM Plex Sans Arabic', 'Inter', 'ui-sans-serif', 'system-ui', 'sans-serif'],
        arabic: ['IBM Plex Sans Arabic', 'Cairo', 'Inter', 'ui-sans-serif', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace'],
      },
      borderRadius: {
        '4xl': '2rem',
      },
      animation: {
        'fade-in': 'fadeIn 150ms ease-in-out',
        'fade-out': 'fadeOut 150ms ease-in-out',
        'slide-in': 'slideIn 200ms ease-out',
        'slide-out': 'slideOut 200ms ease-in',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        fadeOut: {
          '0%': { opacity: '1' },
          '100%': { opacity: '0' },
        },
        slideIn: {
          '0%': { transform: 'translateY(-10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        slideOut: {
          '0%': { transform: 'translateY(0)', opacity: '1' },
          '100%': { transform: 'translateY(-10px)', opacity: '0' },
        },
      },
    },
  },
  plugins: [],
}