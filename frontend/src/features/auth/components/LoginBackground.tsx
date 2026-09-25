import { useEffect, useRef } from 'react';

interface LoginBackgroundProps {
  isDark: boolean;
}

interface NetworkPoint {
  x: number;
  y: number;
  vx: number;
  vy: number;
  radius: number;
}

interface TrailPoint {
  x: number;
  y: number;
  life: number;
}

export function LoginBackground({ isDark }: LoginBackgroundProps) {
  const backgroundRef = useRef<HTMLDivElement>(null);
  const networkCanvasRef = useRef<HTMLCanvasElement>(null);
  const trailCanvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const background = backgroundRef.current;
    const networkCanvas = networkCanvasRef.current;
    const trailCanvas = trailCanvasRef.current;
    const networkContext = networkCanvas?.getContext('2d');
    const trailContext = trailCanvas?.getContext('2d');
    if (!background || !networkCanvas || !trailCanvas || !networkContext || !trailContext) return;

    const motionPreference = window.matchMedia('(prefers-reduced-motion: reduce)');
    const networkPoints: NetworkPoint[] = [];
    const trailPoints: TrailPoint[] = [];
    let width = 0;
    let height = 0;
    let networkDpr = 1;
    let trailDpr = 1;
    let networkFrame = 0;
    let trailFrame = 0;
    let glowFrame = 0;
    let glowActive = false;
    let pointerX = 50;
    let pointerY = 50;
    let glowX = 50;
    let glowY = 50;

    const cssValue = (name: string, fallback: string) =>
      getComputedStyle(background).getPropertyValue(name).trim() || fallback;

    const resize = () => {
      width = window.innerWidth;
      height = window.innerHeight;
      networkDpr = Math.min(window.devicePixelRatio || 1, 2);
      trailDpr = networkDpr;

      networkCanvas.width = Math.round(width * networkDpr);
      networkCanvas.height = Math.round(height * networkDpr);
      trailCanvas.width = Math.round(width * trailDpr);
      trailCanvas.height = Math.round(height * trailDpr);
      networkContext.setTransform(networkDpr, 0, 0, networkDpr, 0, 0);
      trailContext.setTransform(trailDpr, 0, 0, trailDpr, 0, 0);

      networkPoints.length = 0;
      const pointCount = Math.max(36, Math.min(70, Math.floor((width * height) / 22000)));
      for (let index = 0; index < pointCount; index += 1) {
        networkPoints.push({
          x: Math.random() * width,
          y: Math.random() * height,
          vx: (Math.random() - 0.5) * 0.22,
          vy: (Math.random() - 0.5) * 0.22,
          radius: Math.random() * 1.8 + 0.8,
        });
      }
    };

    const drawNetwork = () => {
      networkContext.clearRect(0, 0, width, height);
      const maxDistance = Math.min(180, Math.max(130, width / 7));
      const lineColor = cssValue('--pf-login-net-line', '120, 150, 190');
      const dotColor = cssValue('--pf-login-net-dot', '180, 200, 230');

      for (const point of networkPoints) {
        point.x += point.vx;
        point.y += point.vy;
        if (point.x < -20) point.x = width + 20;
        if (point.x > width + 20) point.x = -20;
        if (point.y < -20) point.y = height + 20;
        if (point.y > height + 20) point.y = -20;
      }

      for (let index = 0; index < networkPoints.length; index += 1) {
        const first = networkPoints[index];
        if (!first) continue;
        for (let nextIndex = index + 1; nextIndex < networkPoints.length; nextIndex += 1) {
          const second = networkPoints[nextIndex];
          if (!second) continue;
          const dx = first.x - second.x;
          const dy = first.y - second.y;
          const distance = Math.sqrt(dx * dx + dy * dy);
          if (distance < maxDistance) {
            networkContext.strokeStyle = `rgba(${lineColor},${(1 - distance / maxDistance) * 0.55})`;
            networkContext.lineWidth = 1;
            networkContext.beginPath();
            networkContext.moveTo(first.x, first.y);
            networkContext.lineTo(second.x, second.y);
            networkContext.stroke();
          }
        }
      }

      for (const point of networkPoints) {
        networkContext.beginPath();
        networkContext.fillStyle = `rgba(${dotColor},1)`;
        networkContext.arc(point.x, point.y, point.radius, 0, Math.PI * 2);
        networkContext.fill();
      }

      if (!motionPreference.matches && !document.hidden) {
        networkFrame = window.requestAnimationFrame(drawNetwork);
      }
    };

    const drawTrail = () => {
      trailContext.clearRect(0, 0, width, height);
      if (trailPoints.length > 1) {
        const indigo = cssValue('--pf-login-indigo', '#2563eb');
        const teal = cssValue('--pf-login-teal', '#2dd4bf');
        for (let index = 1; index < trailPoints.length; index += 1) {
          const previous = trailPoints[index - 1];
          const point = trailPoints[index];
          if (!previous || !point) continue;
          point.life *= 0.965;
          const progress = index / trailPoints.length;
          const gradient = trailContext.createLinearGradient(previous.x, previous.y, point.x, point.y);
          gradient.addColorStop(0, indigo);
          gradient.addColorStop(1, teal);
          trailContext.strokeStyle = gradient;
          trailContext.globalAlpha = Math.max(0, point.life * progress);
          trailContext.lineWidth = Math.max(1.5, 10 * progress * (0.4 + point.life * 0.6));
          trailContext.lineCap = 'round';
          trailContext.shadowBlur = 14;
          trailContext.shadowColor = teal;
          trailContext.beginPath();
          trailContext.moveTo(previous.x, previous.y);
          trailContext.lineTo(point.x, point.y);
          trailContext.stroke();
        }
        trailContext.shadowBlur = 0;
        trailContext.globalAlpha = 1;
        while (trailPoints.length && (trailPoints[0]?.life ?? 0) <= 0.05) trailPoints.shift();
      }
      if (!document.hidden) trailFrame = window.requestAnimationFrame(drawTrail);
    };

    const animateGlow = () => {
      glowX += (pointerX - glowX) * 0.12;
      glowY += (pointerY - glowY) * 0.12;
      background.style.setProperty('--pf-login-mx', `${glowX}%`);
      background.style.setProperty('--pf-login-my', `${glowY}%`);
      if (Math.abs(pointerX - glowX) > 0.05 || Math.abs(pointerY - glowY) > 0.05) {
        glowFrame = window.requestAnimationFrame(animateGlow);
      } else {
        glowActive = false;
      }
    };

    const handlePointerMove = (event: PointerEvent) => {
      pointerX = (event.clientX / Math.max(window.innerWidth, 1)) * 100;
      pointerY = (event.clientY / Math.max(window.innerHeight, 1)) * 100;
      trailPoints.push({ x: event.clientX, y: event.clientY, life: 1 });
      if (trailPoints.length > 42) trailPoints.shift();
      if (!glowActive) {
        glowActive = true;
        glowFrame = window.requestAnimationFrame(animateGlow);
      }
    };

    const handleVisibilityChange = () => {
      window.cancelAnimationFrame(networkFrame);
      window.cancelAnimationFrame(trailFrame);
      if (document.hidden) return;
      drawNetwork();
      if (!motionPreference.matches) drawTrail();
    };

    const handleResize = () => {
      resize();
      window.cancelAnimationFrame(networkFrame);
      drawNetwork();
    };

    const handleMotionPreferenceChange = () => {
      window.cancelAnimationFrame(networkFrame);
      window.cancelAnimationFrame(trailFrame);
      if (motionPreference.matches) {
        window.removeEventListener('pointermove', handlePointerMove);
        trailContext.clearRect(0, 0, width, height);
        trailPoints.length = 0;
      } else {
        window.addEventListener('pointermove', handlePointerMove, { passive: true });
      }
      drawNetwork();
      if (!motionPreference.matches && !document.hidden) drawTrail();
    };

    resize();
    drawNetwork();
    if (!motionPreference.matches) {
      window.addEventListener('pointermove', handlePointerMove, { passive: true });
      drawTrail();
    }
    window.addEventListener('resize', handleResize);
    document.addEventListener('visibilitychange', handleVisibilityChange);
    motionPreference.addEventListener('change', handleMotionPreferenceChange);

    return () => {
      window.cancelAnimationFrame(networkFrame);
      window.cancelAnimationFrame(trailFrame);
      window.cancelAnimationFrame(glowFrame);
      window.removeEventListener('pointermove', handlePointerMove);
      document.removeEventListener('visibilitychange', handleVisibilityChange);
      motionPreference.removeEventListener('change', handleMotionPreferenceChange);
      window.removeEventListener('resize', handleResize);
    };
  }, [isDark]);

  return (
    <>
      <div
        ref={backgroundRef}
        className="pf-login-background"
        data-theme={isDark ? 'dark' : 'light'}
        aria-hidden="true"
      >
        <div className="pf-login-wash" />
        <div className="pf-login-grid" />
        <div className="pf-login-blob pf-login-blob-one" />
        <div className="pf-login-blob pf-login-blob-two" />
        <div className="pf-login-blob pf-login-blob-three" />
        <canvas ref={networkCanvasRef} className="pf-login-network" />
        <div className="pf-login-vignette" />
        <div className="pf-login-cursor-glow" />
      </div>
      <canvas
        ref={trailCanvasRef}
        className="pf-login-trail"
        data-theme={isDark ? 'dark' : 'light'}
        aria-hidden="true"
      />
    </>
  );
}
