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
        // PartFlow Design System Colors (from demo.html)
        bg: {
          DEFAULT: '#070a12',
          surface: '#0c111c',
          surface2: '#111827',
          surface3: '#151e2d',
        },
        surface: {
          DEFAULT: '#0c111c',
          2: '#111827',
          3: '#151e2d',
        },
        border: {
          DEFAULT: 'rgba(148, 163, 184, 0.13)',
        },
        text: {
          DEFAULT: '#f1f7ff',
          muted: '#8290a7',
        },
        // Accent Colors
        cyan: {
          DEFAULT: '#22d3ee',
          50: '#cffafe',
          100: '#a5f3fc',
          200: '#67e8f9',
          300: '#22d3ee',
          400: '#06b6d4',
          500: '#0891b2',
          600: '#0e7490',
          700: '#155e75',
          800: '#164e63',
          900: '#083344',
        },
        blue: {
          DEFAULT: '#38bdf8',
          50: '#f0f9ff',
          100: '#e0f2fe',
          200: '#bae6fd',
          300: '#7dd3fc',
          400: '#38bdf8',
          500: '#0ea5e9',
          600: '#0284c7',
          700: '#0369a1',
          800: '#075985',
          900: '#0c4a6e',
        },
        green: {
          DEFAULT: '#34d399',
          50: '#ecfdf5',
          100: '#d1fae5',
          200: '#a7f3d0',
          300: '#6ee7b7',
          400: '#34d399',
          500: '#10b981',
          600: '#059669',
          700: '#047857',
          800: '#065f46',
          900: '#064e3b',
        },
        yellow: {
          DEFAULT: '#fbbf24',
          50: '#fffbeb',
          100: '#fef3c7',
          200: '#fde68a',
          300: '#fcd34d',
          400: '#fbbf24',
          500: '#f59e0b',
          600: '#d97706',
          700: '#b45309',
          800: '#92400e',
          900: '#78350f',
        },
        red: {
          DEFAULT: '#fb7185',
          50: '#fff1f2',
          100: '#ffe4e6',
          200: '#fecdd3',
          300: '#fda4af',
          400: '#fb7185',
          500: '#f43f5e',
          600: '#e11d48',
          700: '#be123c',
          800: '#9f1239',
          900: '#881337',
        },
        // Legacy color names for compatibility
        primary: {
          DEFAULT: '#22d3ee',
          hover: '#06b6d4',
          active: '#0891b2',
        },
        success: {
          DEFAULT: '#34d399',
          hover: '#10b981',
        },
        warning: {
          DEFAULT: '#fbbf24',
          hover: '#f59e0b',
        },
        danger: {
          DEFAULT: '#fb7185',
          hover: '#f43f5e',
        },
        info: {
          DEFAULT: '#38bdf8',
          hover: '#0ea5e9',
        },
      },
      fontFamily: {
        sans: ['Inter', 'ui-sans-serif', 'system-ui', '-apple-system', 'BlinkMacSystemFont', '"Segoe UI"', 'sans-serif'],
        arabic: ['var(--font-family-arabic)', 'system-ui', 'sans-serif'],
      },
      fontSize: {
        // PartFlow Design System Font Sizes
        'h1': ['30px', { lineHeight: '1.2', letterSpacing: '-1px', fontWeight: '800' }],
        'h2': ['24px', { lineHeight: '1.3', letterSpacing: '-0.5px', fontWeight: '700' }],
        'h3': ['18px', { lineHeight: '1.4', letterSpacing: '-0.3px', fontWeight: '650' }],
        'body': ['13px', { lineHeight: '1.6' }],
        'small': ['11px', { lineHeight: '1.5' }],
        'tiny': ['9px', { lineHeight: '1.4', letterSpacing: '1.4px' }],
        'metric': ['27px', { lineHeight: '1.2', fontWeight: '750' }],
        'eyebrow': ['10px', { lineHeight: '1.4', letterSpacing: '2px', textTransform: 'uppercase' }],
        'section-title': ['10px', { lineHeight: '1.4', letterSpacing: '1.7px', textTransform: 'uppercase' }],
        // Legacy for compatibility
        'page-title': ['30px', { lineHeight: '1.2', letterSpacing: '-1px', fontWeight: '800' }],
        'card-title': ['14px', { lineHeight: '1.4', fontWeight: '650' }],
        'caption': ['11px', { lineHeight: '1.5' }],
      },
      spacing: {
        // PartFlow Design System Spacing
        'xs': '4px',
        'sm': '9px',
        'md': '14px',
        'lg': '18px',
        'xl': '22px',
        '2xl': '28px',
        '3xl': '32px',
        '4xl': '50px',
        // Legacy for compatibility
        '1': '4px',
        '2': '8px',
        '3': '12px',
        '4': '16px',
        '5': '20px',
        '6': '24px',
        '8': '32px',
        '10': '40px',
        '12': '48px',
        '16': '64px',
      },
      borderRadius: {
        // PartFlow Design System Radius
        'sm': '10px',
        'md': '16px',
        'lg': '999px',
        'full': '999px',
      },
      boxShadow: {
        // PartFlow Design System Shadows
        'card': '0 15px 50px rgba(0, 0, 0, 0.20)',
        'glow': '0 0 25px rgba(34, 211, 238, 0.07)',
        'glow-strong': '0 0 30px rgba(34, 211, 238, 0.10)',
        'glow-soft': '0 0 20px rgba(34, 211, 238, 0.08)',
        'nav-active': 'inset 2px 0 0 #22d3ee, 0 0 20px rgba(34, 211, 238, 0.04)',
        // Legacy for compatibility
        'sm': '0 1px 2px rgba(0, 0, 0, 0.05)',
        'md': '0 4px 6px rgba(0, 0, 0, 0.1)',
        'lg': '0 10px 15px rgba(0, 0, 0, 0.1)',
        'xl': '0 20px 25px rgba(0, 0, 0, 0.1)',
      },
      transitionDuration: {
        // PartFlow Design System Transitions
        'fast': '150ms',
        'normal': '180ms',
        'slow': '200ms',
        'slower': '250ms',
        // Legacy for compatibility
        'DEFAULT': '180ms',
      },
      transitionTimingFunction: {
        'default': 'ease',
        'in': 'ease-in',
        'out': 'ease-out',
      },
      zIndex: {
        // PartFlow Design System Z-Index
        'base': '1',
        'dropdown': '10',
        'sticky': '20',
        'modal': '100',
        'tooltip': '200',
        // Legacy for compatibility
        'fixed': '50',
        'modal-backdrop': '90',
        'popover': '150',
      },
      backgroundImage: {
        // PartFlow Design System Gradients
        'bg-gradient': 'radial-gradient(circle at 80% 0%, rgba(34, 211, 238, 0.10), transparent 30%), radial-gradient(circle at 20% 80%, rgba(59, 130, 246, 0.08), transparent 30%), #070a12',
        'sidebar-gradient': 'linear-gradient(180deg, rgba(12, 17, 28, 0.92), rgba(7, 10, 18, 0.86))',
        'card-gradient': 'linear-gradient(145deg, rgba(17, 24, 39, 0.92), rgba(9, 14, 24, 0.92))',
        'card-ai-gradient': 'linear-gradient(145deg, rgba(34, 211, 238, 0.07), rgba(17, 24, 39, 0.92))',
        'logo-gradient': 'linear-gradient(135deg, rgba(34, 211, 238, 0.12), rgba(59, 130, 246, 0.04))',
        'button-primary-gradient': 'linear-gradient(135deg, rgba(34, 211, 238, 0.17), rgba(59, 130, 246, 0.12))',
        'nav-active-gradient': 'linear-gradient(90deg, rgba(34, 211, 238, 0.13), rgba(59, 130, 246, 0.03))',
        'chart-bar-gradient': 'linear-gradient(180deg, rgba(34, 211, 238, 0.90), rgba(59, 130, 246, 0.12))',
      },
      animation: {
        // PartFlow Design System Animations
        'slide-in-from-top': 'slideInFromTop 0.3s ease-out',
        'fade-in': 'fadeIn 0.2s ease-out',
        'scale-in': 'scaleIn 0.2s ease-out',
        'glow-pulse': 'glowPulse 2s ease-in-out infinite',
        'float': 'float 3s ease-in-out infinite',
        'slide-in-right': 'slideInRight 0.3s ease-out',
        'slide-in-left': 'slideInLeft 0.3s ease-out',
      },
      keyframes: {
        slideInFromTop: {
          '0%': { transform: 'translateY(-10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        scaleIn: {
          '0%': { transform: 'scale(0.95)', opacity: '0' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
        glowPulse: {
          '0%, 100%': { boxShadow: '0 0 25px rgba(34, 211, 238, 0.07)' },
          '50%': { boxShadow: '0 0 35px rgba(34, 211, 238, 0.12)' },
        },
        float: {
          '0%, 100%': { transform: 'translateY(0)' },
          '50%': { transform: 'translateY(-5px)' },
        },
        slideInRight: {
          '0%': { transform: 'translateX(10px)', opacity: '0' },
          '100%': { transform: 'translateX(0)', opacity: '1' },
        },
        slideInLeft: {
          '0%': { transform: 'translateX(-10px)', opacity: '0' },
          '100%': { transform: 'translateX(0)', opacity: '1' },
        },
      },
    },
  },
  plugins: [],
}