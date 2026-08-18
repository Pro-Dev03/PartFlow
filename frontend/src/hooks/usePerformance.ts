import { useEffect, useState, useRef } from 'react'

// Performance monitoring hook
export function usePerformance() {
  const [metrics, setMetrics] = useState({
    fps: 0,
    memory: 0,
    loadTime: 0,
  })

  useEffect(() => {
    let animationFrameId: number
    let lastTime = performance.now()
    let frameCount = 0

    // Measure FPS
    const measureFPS = () => {
      const currentTime = performance.now()
      frameCount++

      if (currentTime - lastTime >= 1000) {
        const fps = Math.round((frameCount * 1000) / (currentTime - lastTime))
        setMetrics((prev) => ({ ...prev, fps }))
        frameCount = 0
        lastTime = currentTime
      }

      animationFrameId = requestAnimationFrame(measureFPS)
    }

    animationFrameId = requestAnimationFrame(measureFPS)

    // Measure memory (if available)
    const measureMemory = () => {
      if ('memory' in performance) {
        const memory = (performance as any).memory
        setMetrics((prev) => ({
          ...prev,
          memory: Math.round(memory.usedJSHeapSize / 1048576), // Convert to MB
        }))
      }
    }

    const memoryInterval = setInterval(measureMemory, 5000)

    // Measure page load time
    const loadTime = performance.timing.loadEventEnd - performance.timing.navigationStart
    if (loadTime > 0) {
      setMetrics((prev) => ({ ...prev, loadTime }))
    }

    return () => {
      cancelAnimationFrame(animationFrameId)
      clearInterval(memoryInterval)
    }
  }, [])

  return metrics
}

// Debounce hook for performance optimization
export function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value)

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value)
    }, delay)

    return () => {
      clearTimeout(handler)
    }
  }, [value, delay])

  return debouncedValue
}

// Throttle hook for performance optimization
export function useThrottle<T>(value: T, limit: number): T {
  const [throttledValue, setThrottledValue] = useState<T>(value)
  const lastRan = useRef(Date.now())

  useEffect(() => {
    const handler = setTimeout(() => {
      if (Date.now() - lastRan.current >= limit) {
        setThrottledValue(value)
        lastRan.current = Date.now()
      }
    }, limit - (Date.now() - lastRan.current))

    return () => {
      clearTimeout(handler)
    }
  }, [value, limit])

  return throttledValue
}

// Memoization hook for expensive computations
export function useMemoizedValue<T>(factory: () => T, deps: any[]): T {
  const [value, setValue] = useState<T>(factory)
  const prevDeps = useRef<any[]>(deps)

  useEffect(() => {
    const hasChanged = deps.some((dep, i) => dep !== prevDeps.current[i])
    if (hasChanged) {
      setValue(factory())
      prevDeps.current = deps
    }
  }, deps)

  return value
}

// Window size hook with debounce
export function useWindowSize(debounceMs: number = 250) {
  const [size, setSize] = useState({
    width: window.innerWidth,
    height: window.innerHeight,
  })

  useEffect(() => {
    let timeoutId: NodeJS.Timeout

    const handleResize = () => {
      clearTimeout(timeoutId)
      timeoutId = setTimeout(() => {
        setSize({
          width: window.innerWidth,
          height: window.innerHeight,
        })
      }, debounceMs)
    }

    window.addEventListener('resize', handleResize)
    return () => {
      window.removeEventListener('resize', handleResize)
      clearTimeout(timeoutId)
    }
  }, [debounceMs])

  return size
}

// Network status hook
export function useNetworkStatus() {
  const [isOnline, setIsOnline] = useState(navigator.onLine)
  const [connectionType, setConnectionType] = useState<string>('unknown')

  useEffect(() => {
    const handleOnline = () => setIsOnline(true)
    const handleOffline = () => setIsOnline(false)

    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)

    // Check connection type if available
    if ('connection' in navigator) {
      const connection = (navigator as any).connection
      setConnectionType(connection.effectiveType || 'unknown')

      connection.addEventListener('change', () => {
        setConnectionType(connection.effectiveType || 'unknown')
      })
    }

    return () => {
      window.removeEventListener('online', handleOnline)
      window.removeEventListener('offline', handleOffline)
    }
  }, [])

  return { isOnline, connectionType }
}