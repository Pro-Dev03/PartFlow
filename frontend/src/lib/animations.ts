/**
 * Animation Utilities for PartFlow
 * Following the Design System transition guidelines
 */

// Transition durations as per Design System
export const transitions = {
  fast: '150ms',
  normal: '200ms',
  slow: '300ms',
} as const;

// Easing functions as per Design System
export const easings = {
  easeIn: 'cubic-bezier(0.4, 0, 1, 1)',
  easeOut: 'cubic-bezier(0, 0, 0.2, 1)',
  easeInOut: 'cubic-bezier(0.4, 0, 0.2, 1)',
} as const;

// Common transition combinations
export const transitionPresets = {
  // Hover effects - fast, ease out
  hover: {
    duration: transitions.fast,
    easing: easings.easeOut,
  },
  // Modal open/close - normal, ease in out
  modal: {
    duration: transitions.normal,
    easing: easings.easeInOut,
  },
  // Page load - slow, ease out
  pageLoad: {
    duration: transitions.slow,
    easing: easings.easeOut,
  },
  // Button interactions - fast, ease out
  button: {
    duration: transitions.fast,
    easing: easings.easeOut,
  },
  // Card hover - normal, ease out
  card: {
    duration: transitions.normal,
    easing: easings.easeOut,
  },
  // Table row hover - fast, ease out
  tableRow: {
    duration: transitions.fast,
    easing: easings.easeOut,
  },
} as const;

// Animation classes for Tailwind
export const animationClasses = {
  // Fade in
  fadeIn: 'animate-fadeIn',
  fadeOut: 'animate-fadeOut',
  
  // Slide
  slideInUp: 'animate-slideInUp',
  slideInDown: 'animate-slideInDown',
  slideInLeft: 'animate-slideInLeft',
  slideInRight: 'animate-slideInRight',
  
  // Scale
  scaleIn: 'animate-scaleIn',
  scaleOut: 'animate-scaleOut',
  
  // Spin
  spin: 'animate-spin',
  pulse: 'animate-pulse',
  bounce: 'animate-bounce',
  
  // Shimmer
  shimmer: 'animate-shimmer',
  
  // Enhanced animations
  bounceIn: 'animate-bounceIn',
  flipIn: 'animate-flipIn',
  rotateIn: 'animate-rotateIn',
  zoomIn: 'animate-zoomIn',
  slideInFromLeftWithFade: 'animate-slideInFromLeftWithFade',
  slideInFromRightWithFade: 'animate-slideInFromRightWithFade',
} as const;

// Helper function to create transition string
export function createTransition(options?: {
  duration?: keyof typeof transitions;
  easing?: keyof typeof easings;
  delay?: string;
}): string {
  const duration = options?.duration || 'normal';
  const easing = options?.easing || 'easeInOut';
  const delay = options?.delay || '0ms';
  
  return `transition-all ${transitions[duration]} ${easings[easing]} ${delay}`;
}

// Helper function for combined transitions
export function createTransitionGroup(properties: string[]): string {
  return properties.map(prop => createTransition({ duration: 'normal', easing: 'easeInOut' })).join(', ');
}

// Micro-interaction utilities
export const microInteractions = {
  // Button press effect
  buttonPress: 'active:scale-95 transition-transform duration-100 ease-out',
  
  // Card hover lift
  cardHover: 'hover:-translate-y-1 hover:shadow-xl transition-all duration-200 ease-out',
  
  // Link underline animation
  linkUnderline: 'hover:underline decoration-2 underline-offset-4 transition-all duration-200',
  
  // Input focus ring
  inputFocus: 'focus:ring-2 focus:ring-primary focus:ring-offset-2 focus:ring-offset-background transition-all duration-200',
  
  // Scale on hover
  scaleOnHover: 'hover:scale-105 transition-transform duration-200 ease-out',
  
  // Brightness on hover
  brightnessOnHover: 'hover:brightness-110 transition-all duration-200',
} as const;

// Loading animation states
export const loadingStates = {
  // Skeleton loading
  skeleton: 'animate-pulse bg-surface-elevated',
  
  // Spinner
  spinner: 'animate-spin border-2 border-current border-t-transparent',
  
  // Dots
  dots: 'animate-bounce',
  
  // Progress bar
  progress: 'transition-all duration-300 ease-out',
} as const;

// Page transition animations
export const pageTransitions = {
  // Fade in from bottom
  fadeInUp: 'animate-fadeInUp',
  
  // Fade in with scale
  fadeInScale: 'animate-fadeInScale',
  
  // Slide from right
  slideInRight: 'animate-slideInRight',
  
  // Fade out
  fadeOut: 'animate-fadeOut',
} as const;

// Utility function to stagger animations
export function staggerAnimation(delay: number, index: number): string {
  return `animation-delay: ${delay * index}ms`;
}

// Enhanced staggered animation hook for lists
export function useStaggeredAnimation(count: number, delay: number = 100) {
  return Array.from({ length: count }, (_, i) => ({
    style: { animationDelay: `${i * delay}ms` }
  }));
}

// Animation variants for different use cases
export const animationVariants = {
  subtle: { duration: transitions.slow, easing: easings.easeOut },
  energetic: { duration: transitions.fast, easing: easings.easeIn },
  bouncy: { duration: transitions.normal, easing: easings.easeInOut },
  professional: { duration: transitions.normal, easing: easings.easeOut },
  playful: { duration: transitions.fast, easing: easings.easeInOut },
} as const;

// Spring physics presets for natural motion
export const springPresets = {
  gentle: { stiffness: 100, damping: 10 },
  bouncy: { stiffness: 150, damping: 5 },
  snappy: { stiffness: 200, damping: 15 },
  smooth: { stiffness: 120, damping: 12 },
} as const;

// Gesture-based animations
export const gestureAnimations = {
  swipe: { duration: transitions.fast, easing: easings.easeOut },
  pinch: { duration: transitions.normal, easing: easings.easeInOut },
  drag: { duration: transitions.fast, easing: easings.easeOut },
  scroll: { duration: transitions.normal, easing: easings.easeOut },
} as const;

// Animation sequencing types
export interface AnimationSequence {
  id: string;
  animation: string;
  delay?: number;
  duration?: number;
  onComplete?: () => void;
}

// Animation orchestration function
export function orchestrateAnimations(animations: AnimationSequence[]): void {
  animations.forEach((anim, index) => {
    const delay = anim.delay || 0;
    const totalDelay = delay + (index * 100); // Default stagger
    
    setTimeout(() => {
      // Apply animation logic here
      if (anim.onComplete) {
        setTimeout(anim.onComplete, anim.duration || 200);
      }
    }, totalDelay);
  });
}

// Keyframe animation definitions (for custom CSS)
export const keyframes = {
  fadeIn: {
    from: { opacity: 0 },
    to: { opacity: 1 },
  },
  fadeOut: {
    from: { opacity: 1 },
    to: { opacity: 0 },
  },
  slideInUp: {
    from: { transform: 'translateY(20px)', opacity: 0 },
    to: { transform: 'translateY(0)', opacity: 1 },
  },
  slideInDown: {
    from: { transform: 'translateY(-20px)', opacity: 0 },
    to: { transform: 'translateY(0)', opacity: 1 },
  },
  slideInLeft: {
    from: { transform: 'translateX(-20px)', opacity: 0 },
    to: { transform: 'translateX(0)', opacity: 1 },
  },
  slideInRight: {
    from: { transform: 'translateX(20px)', opacity: 0 },
    to: { transform: 'translateX(0)', opacity: 1 },
  },
  scaleIn: {
    from: { transform: 'scale(0.95)', opacity: 0 },
    to: { transform: 'scale(1)', opacity: 1 },
  },
  scaleOut: {
    from: { transform: 'scale(1)', opacity: 1 },
    to: { transform: 'scale(0.95)', opacity: 0 },
  },
  shimmer: {
    '0%': { backgroundPosition: '-1000px 0' },
    '100%': { backgroundPosition: '1000px 0' },
  },
} as const;