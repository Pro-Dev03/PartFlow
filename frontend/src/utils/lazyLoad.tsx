import { lazy, Suspense, useState, useEffect, useRef } from 'react'
import { Skeleton } from '../components/ui/Skeleton'

// Generic lazy loading wrapper with loading state
export function withLazyLoad<T extends React.ComponentType<any>>(
  importFunc: () => Promise<{ default: T }>,
  fallback?: React.ReactNode
) {
  const LazyComponent = lazy(importFunc)

  return function WithLazyLoad(props: React.ComponentProps<T>) {
    return (
      <Suspense
        fallback={
          fallback || (
            <div className="p-4 space-y-4">
              <Skeleton className="h-8 w-1/3" />
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-4 w-5/6" />
              <Skeleton className="h-32 w-full" />
            </div>
          )
        }
      >
        <LazyComponent {...props} />
      </Suspense>
    )
  }
}

// Preload a component without rendering it
export function preloadComponent(importFunc: () => Promise<any>) {
  importFunc()
}

// Intersection Observer lazy loading for images
export function useLazyImage() {
  const [isLoaded, setIsLoaded] = useState(false)
  const [isInView, setIsInView] = useState(false)
  const imgRef = useRef<HTMLImageElement>(null)

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsInView(true)
          observer.disconnect()
        }
      },
      { threshold: 0.1 }
    )

    if (imgRef.current) {
      observer.observe(imgRef.current)
    }

    return () => observer.disconnect()
  }, [])

  const handleLoad = () => {
    setIsLoaded(true)
  }

  return { imgRef, isInView, isLoaded, handleLoad }
}

// Lazy image component
export function LazyImage({
  src,
  alt,
  className,
  ...props
}: React.ImgHTMLAttributes<HTMLImageElement>) {
  const { imgRef, isInView, isLoaded, handleLoad } = useLazyImage()

  return (
    <img
      ref={imgRef}
      src={isInView ? src : undefined}
      alt={alt}
      className={className}
      onLoad={handleLoad}
      style={{
        opacity: isLoaded ? 1 : 0,
        transition: 'opacity 0.3s ease-in-out',
      }}
      {...props}
    />
  )
}

// Virtual scrolling for large lists
export function useVirtualScroll(options: {
  itemCount: number
  itemHeight: number
  containerHeight: number
  overscan?: number
}) {
  const { itemCount, itemHeight, containerHeight, overscan = 5 } = options
  const [scrollTop, setScrollTop] = useState(0)

  const startIndex = Math.max(0, Math.floor(scrollTop / itemHeight) - overscan)
  const endIndex = Math.min(
    itemCount - 1,
    Math.floor((scrollTop + containerHeight) / itemHeight) + overscan
  )

  const visibleItems = Array.from(
    { length: endIndex - startIndex + 1 },
    (_, i) => startIndex + i
  )

  const totalHeight = itemCount * itemHeight
  const offsetY = startIndex * itemHeight

  return {
    visibleItems,
    totalHeight,
    offsetY,
    onScroll: (e: React.UIEvent<HTMLDivElement>) => {
      setScrollTop(e.currentTarget.scrollTop)
    },
  }
}