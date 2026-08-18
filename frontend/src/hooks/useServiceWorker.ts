import { useEffect, useState } from 'react'

interface ServiceWorkerRegistration {
  waiting: ServiceWorker | null
  active: ServiceWorker | null
  update: () => Promise<void>
}

export function useServiceWorker(): ServiceWorkerRegistration | null {
  const [registration, setRegistration] = useState<ServiceWorkerRegistration | null>(null)

  useEffect(() => {
    if ('serviceWorker' in navigator) {
      // Register service worker
      navigator.serviceWorker
        .register('/sw.js')
        .then((reg) => {
          console.log('[SW] Service worker registered:', reg.scope)
          
          setRegistration({
            waiting: reg.waiting || null,
            active: reg.active || null,
            update: () => reg.update(),
          })

          // Check for updates
          reg.addEventListener('updatefound', () => {
            const newWorker = reg.installing
            if (newWorker) {
              newWorker.addEventListener('statechange', () => {
                if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
                  // New version available
                  console.log('[SW] New version available')
                  setRegistration((prev) => ({
                    ...prev!,
                    waiting: newWorker,
                  }))
                }
              })
            }
          })
        })
        .catch((error) => {
          console.error('[SW] Service worker registration failed:', error)
        })

      // Handle controller change
      navigator.serviceWorker.addEventListener('controllerchange', () => {
        console.log('[SW] Controller changed - reloading page')
        window.location.reload()
      })
    }

    return () => {
      // Cleanup if needed
    }
  }, [])

  return registration
}

// Hook to check online/offline status
export function useOnlineStatus() {
  const [isOnline, setIsOnline] = useState(navigator.onLine)

  useEffect(() => {
    const handleOnline = () => setIsOnline(true)
    const handleOffline = () => setIsOnline(false)

    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)

    return () => {
      window.removeEventListener('online', handleOnline)
      window.removeEventListener('offline', handleOffline)
    }
  }, [])

  return isOnline
}

// Hook to register background sync
export function useBackgroundSync() {
  const registerSync = async (tag: string) => {
    if ('serviceWorker' in navigator && 'sync' in ServiceWorkerRegistration.prototype) {
      const reg = await navigator.serviceWorker.ready
      try {
        await reg.sync.register(tag)
        console.log(`[SW] Background sync registered: ${tag}`)
      } catch (error) {
        console.error(`[SW] Background sync registration failed: ${tag}`, error)
      }
    }
  }

  return { registerSync }
}