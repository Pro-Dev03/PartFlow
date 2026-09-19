import { useState, useEffect } from 'react';
import { ChevronDown, ChevronUp } from 'lucide-react';
import { cn } from '../../utils';

interface ScrollIndicatorProps {
  className?: string;
  showAt?: number; // Show indicator when scrolled past this percentage
  scrollContainer?: React.RefObject<HTMLDivElement | null>; // Custom scroll container
}

export function ScrollIndicator({ className, showAt = 10, scrollContainer }: ScrollIndicatorProps) {
  const [showTop, setShowTop] = useState(false);
  const [showBottom, setShowBottom] = useState(false);

  useEffect(() => {
    const container = scrollContainer?.current;
    const isWindowScroll = !container;

    const handleScroll = () => {
      let scrollTop: number;
      let scrollHeight: number;
      let clientHeight: number;

      if (isWindowScroll) {
        scrollTop = window.scrollY;
        scrollHeight = document.documentElement.scrollHeight;
        clientHeight = window.innerHeight;
      } else if (container) {
        scrollTop = container.scrollTop;
        scrollHeight = container.scrollHeight;
        clientHeight = container.clientHeight;
      } else {
        return; // Container is null, can't scroll
      }

      const docHeight = scrollHeight - clientHeight;
      const scrollPercent = docHeight > 0 ? (scrollTop / docHeight) * 100 : 0;

      setShowTop(scrollPercent > showAt);
      setShowBottom(scrollPercent < 90);
    };

    if (isWindowScroll) {
      window.addEventListener('scroll', handleScroll);
      handleScroll(); // Check initial state
      return () => window.removeEventListener('scroll', handleScroll);
    } else if (container) {
      container.addEventListener('scroll', handleScroll);
      handleScroll(); // Check initial state
      return () => container.removeEventListener('scroll', handleScroll);
    }
    return () => {};
  }, [showAt, scrollContainer]);

  const scrollToTop = () => {
    const container = scrollContainer?.current;
    if (container) {
      container.scrollTo({ top: 0, behavior: 'smooth' });
    } else {
      window.scrollTo({ top: 0, behavior: 'smooth' });
    }
  };

  const scrollToBottom = () => {
    const container = scrollContainer?.current;
    if (container) {
      container.scrollTo({ top: container.scrollHeight, behavior: 'smooth' });
    } else {
      window.scrollTo({ top: document.documentElement.scrollHeight, behavior: 'smooth' });
    }
  };

  return (
    <>
      {showTop && (
        <button
          onClick={scrollToTop}
          className={cn(
            'fixed bottom-6 left-6 z-50',
            'w-10 h-10 rounded-full',
            'bg-cyan/20 border border-cyan/30',
            'flex items-center justify-center',
            'hover:bg-cyan/30 transition-colors',
            'shadow-lg backdrop-blur-sm',
            className
          )}
          aria-label="العودة للأعلى"
        >
          <ChevronUp className="w-5 h-5 text-cyan" />
        </button>
      )}
      {showBottom && (
        <button
          onClick={scrollToBottom}
          className={cn(
            'fixed bottom-6 right-6 z-50',
            'w-10 h-10 rounded-full',
            'bg-cyan/20 border border-cyan/30',
            'flex items-center justify-center',
            'hover:bg-cyan/30 transition-colors',
            'shadow-lg backdrop-blur-sm',
            className
          )}
          aria-label="الذهاب للأسفل"
        >
          <ChevronDown className="w-5 h-5 text-cyan" />
        </button>
      )}
    </>
  );
}

interface ScrollProgressProps {
  className?: string;
  color?: string;
  scrollContainer?: React.RefObject<HTMLDivElement | null>; // Custom scroll container
}

export function ScrollProgress({ className, color = '#14b8a6', scrollContainer }: ScrollProgressProps) {
  const [scrollProgress, setScrollProgress] = useState(0);

  useEffect(() => {
    const container = scrollContainer?.current;
    const isWindowScroll = !container;

    const handleScroll = () => {
      let scrollTop: number;
      let scrollHeight: number;
      let clientHeight: number;

      if (isWindowScroll) {
        scrollTop = window.scrollY;
        scrollHeight = document.documentElement.scrollHeight;
        clientHeight = window.innerHeight;
      } else if (container) {
        scrollTop = container.scrollTop;
        scrollHeight = container.scrollHeight;
        clientHeight = container.clientHeight;
      } else {
        return; // Container is null, can't scroll
      }

      const docHeight = scrollHeight - clientHeight;
      const progress = docHeight > 0 ? (scrollTop / docHeight) * 100 : 0;
      setScrollProgress(progress);
    };

    if (isWindowScroll) {
      window.addEventListener('scroll', handleScroll);
      handleScroll();
      return () => window.removeEventListener('scroll', handleScroll);
    } else if (container) {
      container.addEventListener('scroll', handleScroll);
      handleScroll();
      return () => container.removeEventListener('scroll', handleScroll);
    }
    return () => {};
  }, [scrollContainer]);

  return (
    <div
      className={cn('fixed top-0 left-0 right-0 h-1 z-50', className)}
      style={{ background: 'rgba(7, 10, 18, 0.8)' }}
    >
      <div
        className="h-full transition-all duration-150 ease-out"
        style={{
          width: `${scrollProgress}%`,
          background: color,
          boxShadow: `0 0 10px ${color}`,
        }}
      />
    </div>
  );
}