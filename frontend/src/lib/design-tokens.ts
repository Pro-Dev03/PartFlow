/**
 * PartFlow Design Tokens
 * 
 * Based on demo.html - Futuristic Professional Design System
 * Extracted and converted to TypeScript for React usage
 */

// =========================
// COLORS
// =========================

export const colors = {
  // Backgrounds
  bg: '#070a12',
  surface: '#0c111c',
  surface2: '#111827',
  surface3: '#151e2d',

  // Borders
  border: 'rgba(148, 163, 184, 0.13)',

  // Text
  text: '#f1f7ff',
  muted: '#8290a7',

  // Accent Colors
  blue: '#38bdf8',
  cyan: '#22d3ee',
  green: '#34d399',
  yellow: '#fbbf24',
  red: '#fb7185',

  // Special
  white: '#ffffff',
  black: '#000000',
} as const;

// =========================
// TYPOGRAPHY
// =========================

export const typography = {
  fontFamily: [
    'Inter',
    'ui-sans-serif',
    'system-ui',
    '-apple-system',
    'BlinkMacSystemFont',
    '"Segoe UI"',
    'sans-serif',
  ].join(', '),

  fontSizes: {
    h1: '30px',
    h2: '24px',
    h3: '18px',
    body: '13px',
    small: '11px',
    tiny: '9px',
    metric: '27px',
  },

  fontWeights: {
    regular: 400,
    medium: 500,
    semibold: 650,
    bold: 700,
    extrabold: 750,
    black: 800,
  },

  letterSpacings: {
    tight: '-1px',
    normal: '0px',
    wide: '0.5px',
    wider: '1.4px',
    widest: '2px',
  },

  lineHeights: {
    tight: 1.4,
    normal: 1.5,
    relaxed: 1.6,
  },
} as const;

// =========================
// SPACING
// =========================

export const spacing = {
  xs: '4px',
  sm: '9px',
  md: '14px',
  lg: '18px',
  xl: '22px',
  '2xl': '28px',
  '3xl': '32px',
  '4xl': '50px',
} as const;

// =========================
// RADIUS
// =========================

export const radius = {
  sm: '10px',
  md: '16px',
  lg: '999px',
} as const;

// =========================
// SHADOWS
// =========================

export const shadows = {
  card: '0 15px 50px rgba(0, 0, 0, 0.20)',
  glow: '0 0 25px rgba(34, 211, 238, 0.07)',
  glowStrong: '0 0 30px rgba(34, 211, 238, 0.10)',
  glowSoft: '0 0 20px rgba(34, 211, 238, 0.08)',
  navActive: 'inset 2px 0 0 #22d3ee, 0 0 20px rgba(34, 211, 238, 0.04)',
} as const;

// =========================
// GRADIENTS
// =========================

export const gradients = {
  // Background
  bg: `
    radial-gradient(circle at 80% 0%, rgba(34, 211, 238, 0.10), transparent 30%),
    radial-gradient(circle at 20% 80%, rgba(59, 130, 246, 0.08), transparent 30%),
    #070a12
  `,

  // Sidebar
  sidebar: `
    linear-gradient(180deg, rgba(12, 17, 28, 0.92), rgba(7, 10, 18, 0.86))
  `,

  // Card
  card: `
    linear-gradient(145deg, rgba(17, 24, 39, 0.92), rgba(9, 14, 24, 0.92))
  `,

  // Card AI
  cardAI: `
    linear-gradient(145deg, rgba(34, 211, 238, 0.07), rgba(17, 24, 39, 0.92))
  `,

  // Logo
  logo: `
    linear-gradient(135deg, rgba(34, 211, 238, 0.12), rgba(59, 130, 246, 0.04))
  `,

  // Button Primary
  buttonPrimary: `
    linear-gradient(135deg, rgba(34, 211, 238, 0.17), rgba(59, 130, 246, 0.12))
  `,

  // Nav Active
  navActive: `
    linear-gradient(90deg, rgba(34, 211, 238, 0.13), rgba(59, 130, 246, 0.03))
  `,

  // Chart Bar
  chartBar: `
    linear-gradient(180deg, rgba(34, 211, 238, 0.90), rgba(59, 130, 246, 0.12))
  `,
} as const;

// =========================
// TRANSITIONS
// =========================

export const transitions = {
  fast: '150ms ease',
  normal: '180ms ease',
  slow: '200ms ease',
  slower: '250ms ease',
} as const;

// =========================
// Z-INDEX
// =========================

export const zIndex = {
  base: 1,
  dropdown: 10,
  sticky: 20,
  modal: 100,
  tooltip: 200,
} as const;

// =========================
// BREAKPOINTS
// =========================

export const breakpoints = {
  sm: '480px',
  md: '720px',
  lg: '1050px',
  xl: '1280px',
  '2xl': '1536px',
} as const;

// =========================
// COMPONENT STYLES
// =========================

export const componentStyles = {
  // Button
  button: {
    base: {
      border: `1px solid ${colors.border}`,
      background: 'rgba(17, 24, 39, 0.75)',
      color: '#dbeafe',
      borderRadius: radius.sm,
      padding: '10px 14px',
      transition: transitions.normal,
    },
    hover: {
      transform: 'translateY(-1px)',
      borderColor: 'rgba(34, 211, 238, 0.3)',
    },
    primary: {
      borderColor: 'rgba(34, 211, 238, 0.35)',
      background: gradients.buttonPrimary,
      boxShadow: shadows.glow,
    },
  },

  // Card
  card: {
    base: {
      border: `1px solid ${colors.border}`,
      borderRadius: radius.md,
      background: gradients.card,
      boxShadow: shadows.card,
      padding: spacing.lg,
      transition: transitions.slow,
    },
    hover: {
      borderColor: 'rgba(148, 163, 184, 0.22)',
      transform: 'translateY(-2px)',
    },
    ai: {
      borderColor: 'rgba(34, 211, 238, 0.18)',
      background: gradients.cardAI,
    },
  },

  // Badge
  badge: {
    base: {
      padding: '4px 8px',
      borderRadius: radius.lg,
      fontSize: typography.fontSizes.tiny,
      letterSpacing: typography.letterSpacings.wide,
    },
    warning: {
      color: colors.yellow,
      background: 'rgba(251, 191, 36, 0.08)',
    },
    danger: {
      color: colors.red,
      background: 'rgba(251, 113, 133, 0.08)',
    },
    success: {
      color: colors.green,
      background: 'rgba(52, 211, 153, 0.08)',
    },
  },

  // Navigation
  nav: {
    link: {
      base: {
        display: 'flex',
        alignItems: 'center',
        gap: spacing.sm,
        padding: '11px 12px',
        borderRadius: radius.sm,
        color: '#94a3b8',
        textDecoration: 'none',
        fontSize: typography.fontSizes.body,
        transition: transitions.normal,
      },
      hover: {
        color: colors.white,
        background: 'rgba(148, 163, 184, 0.06)',
        transform: 'translateX(2px)',
      },
      active: {
        color: '#eaffff',
        background: gradients.navActive,
        border: '1px solid rgba(34, 211, 238, 0.14)',
        boxShadow: shadows.navActive,
      },
    },
  },

  // Sidebar
  sidebar: {
    base: {
      borderRight: `1px solid ${colors.border}`,
      background: gradients.sidebar,
      backdropFilter: 'blur(20px)',
      padding: `${spacing.xl} ${spacing.md}`,
    },
  },
} as const;

// =========================
// THEME CONFIGURATION
// =========================

export const theme = {
  colors,
  typography,
  spacing,
  radius,
  shadows,
  gradients,
  transitions,
  zIndex,
  breakpoints,
  componentStyles,
} as const;

// =========================
// CSS VARIABLES (for global styles)
// =========================

export const cssVariables = {
  '--bg': colors.bg,
  '--surface': colors.surface,
  '--surface-2': colors.surface2,
  '--surface-3': colors.surface3,
  '--border': colors.border,
  '--text': colors.text,
  '--muted': colors.muted,
  '--blue': colors.blue,
  '--cyan': colors.cyan,
  '--green': colors.green,
  '--yellow': colors.yellow,
  '--red': colors.red,
  '--radius': radius.md,
} as const;

// =========================
// UTILITY FUNCTIONS
// =========================

/**
 * Apply CSS variables to document
 */
export function applyCSSVariables() {
  const root = document.documentElement;
  Object.entries(cssVariables).forEach(([key, value]) => {
    root.style.setProperty(key, value);
  });
}

/**
 * Get theme value by path
 */
export function getThemeValue(path: string): any {
  const keys = path.split('.');
  let value: any = theme;
  
  for (const key of keys) {
    value = value?.[key];
    if (value === undefined) return undefined;
  }
  
  return value;
}

// =========================
// EXPORT DEFAULT
// =========================

export default theme;