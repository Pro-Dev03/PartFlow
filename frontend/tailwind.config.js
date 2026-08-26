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
        // PartFlow Design System Colors - Use CSS variables for theme support
        background: {
          DEFAULT: 'var(--bg-background)',
          surface: 'var(--bg-surface)',
          'surface-elevated': 'var(--bg-surface-elevated)',
          'surface-3': 'var(--bg-surface-3)',
        },
        surface: {
          DEFAULT: 'var(--bg-surface)',
          elevated: 'var(--bg-surface-elevated)',
          '3': 'var(--bg-surface-3)',
        },
        border: {
          DEFAULT: 'var(--border-default)',
          primary: 'var(--border-primary)',
        },
        text: {
          DEFAULT: 'var(--text-primary)',
          secondary: 'var(--text-secondary)',
          tertiary: 'var(--text-tertiary)',
          muted: 'var(--text-muted)',
          disabled: 'var(--text-disabled)',
          'on-primary': 'var(--text-on-primary)',
        },
        // Primary/Accent colors - use CSS variables
        primary: {
          DEFAULT: 'var(--primary)',
          hover: 'var(--primary-hover)',
          active: 'var(--primary-active)',
        },
        // Status colors - use CSS variables
        success: {
          DEFAULT: 'var(--success)',
          hover: 'var(--success-hover)',
        },
        warning: {
          DEFAULT: 'var(--warning)',
          hover: 'var(--warning-hover)',
        },
        danger: {
          DEFAULT: 'var(--danger)',
          hover: 'var(--danger-hover)',
        },
        info: {
          DEFAULT: 'var(--info)',
          hover: 'var(--info-hover)',
        },
        // Accent Colors (static - don't change between themes)
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
      },
      fontFamily: {
        sans: ['Inter', 'ui-sans-serif', 'system-ui', '-apple-system', 'BlinkMacSystemFont', '"Segoe UI"', 'sans-serif'],
        arabic: ['var(--font-family-arabic)', 'system-ui', 'sans-serif'],
        technical: ['var(--font-family-technical)', 'monospace'],
      },
      fontSize: {
        // PartFlow Design System Font Sizes (حسب التقرير النهائي)
        'h1': ['30px', { lineHeight: '1.2', letterSpacing: '-0.5px', fontWeight: '700' }], // Page Title: 28-32px / 700
        'h2': ['20px', { lineHeight: '1.3', letterSpacing: '-0.3px', fontWeight: '600' }], // Section Title: 18-22px / 600
        'h3': ['16px', { lineHeight: '1.4', letterSpacing: '-0.2px', fontWeight: '600' }], // Card Title: 15-17px / 600
        'metric': ['34px', { lineHeight: '1.2', fontWeight: '700' }], // Metric: 28-40px / 700
        'body': ['15px', { lineHeight: '1.5' }], // Body: 14-16px
        'small': ['13px', { lineHeight: '1.5' }], // Secondary: 13-14px
        'tiny': ['11px', { lineHeight: '1.4', letterSpacing: '0.5px' }], // Caption: 11-12px
        'eyebrow': ['10px', { lineHeight: '1.4', letterSpacing: '2px', textTransform: 'uppercase' }],
        'section-title': ['10px', { lineHeight: '1.4', letterSpacing: '1.7px', textTransform: 'uppercase' }],
      },
      spacing: {
        // PartFlow Design System Spacing - تطابق tokens.css (حسب التقرير النهائي)
        'xs': '4px',
        'sm': '8px',
        'md': '12px',
        'lg': '16px',
        'xl': '24px',
        '2xl': '32px',
        '3xl': '48px',
        '4xl': '64px',
        // Spacing Numeric - لدعم الـcomponents القديمة
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
      padding: {
        // PartFlow Design System Padding - تطابق tokens.css (حسب التقرير النهائي)
        'xs': '4px',
        'sm': '8px',
        'md': '12px',
        'lg': '16px',
        'xl': '24px',
        '2xl': '32px',
        '3xl': '48px',
        '4xl': '64px',
      },
      gap: {
        // Tailwind gap utilities - تطابق tokens.css (حسب التقرير النهائي)
        'xs': '4px',
        'sm': '8px',
        'md': '12px',
        'lg': '16px',
        'xl': '24px',
        '2xl': '32px',
        '3xl': '48px',
        '4xl': '64px',
      },
      space: {
        // Tailwind space utilities - تطابق tokens.css (حسب التقرير النهائي)
        'xs': '4px',
        'sm': '8px',
        'md': '12px',
        'lg': '16px',
        'xl': '24px',
        '2xl': '32px',
        '3xl': '48px',
        '4xl': '64px',
      },
      borderRadius: {
        // PartFlow Design System Radius (حسب التقرير النهائي)
        'sm': '8px',
        'md': '12px',
        'lg': '16px',
        'xl': '20px',
        'full': '9999px',
      },
      boxShadow: {
        // PartFlow Design System Shadows - تطابق tokens.css (حسب التقرير النهائي)
        'sm': '0 1px 2px rgba(0, 0, 0, 0.20)',
        'md': '0 8px 24px rgba(0, 0, 0, 0.20)',
        'lg': '0 16px 40px rgba(0, 0, 0, 0.25)',
        'xl': '0 20px 25px rgba(0, 0, 0, 0.25)',
        'card': '0 15px 50px rgba(0, 0, 0, 0.20)',
        'glow': '0 0 24px rgba(37, 99, 235, 0.08)', // Subtle glow فقط (أزرق موحّد)
        'glow-strong': '0 0 30px rgba(37, 99, 235, 0.10)', // Stronger glow نادر
        'glow-soft': '0 0 20px rgba(37, 99, 235, 0.06)', // Soft glow
        'nav-active': 'inset 2px 0 0 #2563eb, 0 0 20px rgba(37, 99, 235, 0.04)',
      },
      transitionDuration: {
        // PartFlow Design System Transitions - تطابق tokens.css
        'fast': '150ms',
        'normal': '200ms',
        'slow': '300ms',
        'slower': '250ms',
      },
      transitionTimingFunction: {
        // PartFlow Design System Easing - تطابق tokens.css
        'default': 'cubic-bezier(0.4, 0, 0.2, 1)',
        'in': 'cubic-bezier(0.4, 0, 1, 1)',
        'out': 'cubic-bezier(0, 0, 0.2, 1)',
      },
      zIndex: {
        // PartFlow Design System Z-Index
        'base': '1',
        'dropdown': '10',
        'sticky': '20',
        'modal': '100',
        'tooltip': '200',
      },
      backgroundImage: {
        // PartFlow Design System Gradients (حسب التقرير النهائي - خفيفة جدًا ومحدودة)
        'background-gradient': 'radial-gradient(circle at 80% 0%, rgba(37, 99, 235, 0.12), transparent 30%), radial-gradient(circle at 20% 80%, rgba(59, 130, 246, 0.08), transparent 30%), #090d12',
        'sidebar-gradient': 'linear-gradient(180deg, rgba(12, 17, 28, 0.92), rgba(7, 10, 18, 0.86))',
        'card-gradient': 'linear-gradient(145deg, rgba(17, 24, 39, 0.92), rgba(9, 14, 24, 0.92))',
        'card-ai-gradient': 'linear-gradient(145deg, rgba(37, 99, 235, 0.08), rgba(17, 24, 39, 0.92))',
        'logo-gradient': 'linear-gradient(135deg, rgba(37, 99, 235, 0.12), rgba(59, 130, 246, 0.04))',
        'button-primary-gradient': 'linear-gradient(135deg, rgba(37, 99, 235, 0.18), rgba(59, 130, 246, 0.12))',
        'nav-active-gradient': 'linear-gradient(90deg, rgba(37, 99, 235, 0.13), rgba(59, 130, 246, 0.03))',
        'chart-bar-gradient': 'linear-gradient(180deg, rgba(37, 99, 235, 0.90), rgba(59, 130, 246, 0.12))',
        // Gradients الجديدة حسب التقرير النهائي
        'gradient-subtle': 'linear-gradient(135deg, rgba(37, 99, 235, 0.10), rgba(52, 211, 153, 0.04))',
        'gradient-metric': 'linear-gradient(135deg, rgba(37, 99, 235, 0.15), rgba(59, 130, 246, 0.08))',
      },
      animation: {
        // PartFlow Design System Animations (kebab-case only)
        'fade-in': 'fadeIn 0.2s ease-out',
        'fade-out': 'fadeOut 0.2s ease-in',
        'fade-in-up': 'fadeInUp 0.3s ease-out',
        'fade-in-scale': 'fadeInScale 0.3s ease-out',
        'scale-in': 'scaleIn 0.2s ease-out',
        'scale-out': 'scaleOut 0.2s ease-in',
        'slide-in-up': 'slideInUp 0.3s ease-out',
        'slide-in-down': 'slideInDown 0.3s ease-out',
        'slide-in-left': 'slideInLeft 0.3s ease-out',
        'slide-in-right': 'slideInRight 0.3s ease-out',
        'slide-in-from-top': 'slideInFromTop 0.3s ease-out',
        'slide-in-from-left-with-fade': 'slideInFromLeftWithFade 0.4s ease-out',
        'slide-in-from-right-with-fade': 'slideInFromRightWithFade 0.4s ease-out',
        'glow-pulse': 'glowPulse 2s ease-in-out infinite',
        'float': 'float 3s ease-in-out infinite',
        'bounce-in': 'bounceIn 0.6s cubic-bezier(0.68, -0.55, 0.265, 1.55)',
        'flip-in': 'flipIn 0.6s ease-in-out',
        'rotate-in': 'rotateIn 0.6s ease-out',
        'zoom-in': 'zoomIn 0.3s ease-out',
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
        fadeInUp: {
          '0%': { transform: 'translateY(20px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        fadeInScale: {
          '0%': { transform: 'scale(0.95)', opacity: '0' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
        scaleIn: {
          '0%': { transform: 'scale(0.95)', opacity: '0' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
        scaleOut: {
          '0%': { transform: 'scale(1)', opacity: '1' },
          '100%': { transform: 'scale(0.95)', opacity: '0' },
        },
        slideInUp: {
          '0%': { transform: 'translateY(20px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        slideInDown: {
          '0%': { transform: 'translateY(-20px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        slideInLeft: {
          '0%': { transform: 'translateX(-20px)', opacity: '0' },
          '100%': { transform: 'translateX(0)', opacity: '1' },
        },
        slideInRight: {
          '0%': { transform: 'translateX(20px)', opacity: '0' },
          '100%': { transform: 'translateX(0)', opacity: '1' },
        },
        slideInFromTop: {
          '0%': { transform: 'translateY(-10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        slideInFromLeftWithFade: {
          '0%': { transform: 'translateX(-100%)', opacity: '0' },
          '100%': { transform: 'translateX(0)', opacity: '1' },
        },
        slideInFromRightWithFade: {
          '0%': { transform: 'translateX(100%)', opacity: '0' },
          '100%': { transform: 'translateX(0)', opacity: '1' },
        },
        glowPulse: {
          '0%, 100%': { boxShadow: '0 0 25px rgba(37, 99, 235, 0.07)' },
          '50%': { boxShadow: '0 0 35px rgba(37, 99, 235, 0.12)' },
        },
        float: {
          '0%, 100%': { transform: 'translateY(0)' },
          '50%': { transform: 'translateY(-5px)' },
        },
        bounceIn: {
          '0%': { transform: 'scale(0.3)', opacity: '0' },
          '50%': { transform: 'scale(1.05)', opacity: '1' },
          '70%': { transform: 'scale(0.9)' },
          '100%': { transform: 'scale(1)' },
        },
        flipIn: {
          '0%': { transform: 'perspective(400px) rotateY(90deg)', opacity: '0' },
          '100%': { transform: 'perspective(400px) rotateY(0deg)', opacity: '1' },
        },
        rotateIn: {
          '0%': { transform: 'rotate(-200deg)', opacity: '0' },
          '100%': { transform: 'rotate(0deg)', opacity: '1' },
        },
        zoomIn: {
          '0%': { transform: 'scale(0.5)', opacity: '0' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
      },
    },
  },
  plugins: [],
}