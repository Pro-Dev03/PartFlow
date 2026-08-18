const CACHE_VERSION = 'v1'
const CACHE_NAME = `partflow-${CACHE_VERSION}`

// Assets to cache during install
const STATIC_CACHE_URLS = [
  '/',
  '/index.html',
  '/manifest.json',
]

// API routes that should be cached with network-first strategy
const NETWORK_FIRST_ROUTES = [
  '/api/v1/products',
  '/api/v1/customers',
  '/api/v1/inventory',
]

// API routes that should be cached with cache-first strategy
const CACHE_FIRST_ROUTES = [
  '/api/v1/categories',
  '/api/v1/brands',
]

// Install event - cache static assets
self.addEventListener('install', (event) => {
  console.log('[SW] Installing service worker...')
  
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      console.log('[SW] Caching static assets')
      return cache.addAll(STATIC_CACHE_URLS)
    })
  )
  
  // Force the waiting service worker to become active
  self.skipWaiting()
})

// Activate event - clean up old caches
self.addEventListener('activate', (event) => {
  console.log('[SW] Activating service worker...')
  
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames
          .filter((cacheName) => cacheName !== CACHE_NAME)
          .map((cacheName) => {
            console.log('[SW] Deleting old cache:', cacheName)
            return caches.delete(cacheName)
          })
      )
    })
  )
  
  // Take control of all pages immediately
  return self.clients.claim()
})

// Network-first strategy for dynamic content
async function networkFirst(request) {
  try {
    const networkResponse = await fetch(request)
    
    // Cache the successful response
    if (networkResponse.ok) {
      const cache = await caches.open(CACHE_NAME)
      cache.put(request, networkResponse.clone())
    }
    
    return networkResponse
  } catch (error) {
    console.log('[SW] Network failed, trying cache')
    const cachedResponse = await caches.match(request)
    
    if (cachedResponse) {
      return cachedResponse
    }
    
    // Return offline fallback for API requests
    if (request.url.includes('/api/')) {
      return new Response(
        JSON.stringify({
          success: false,
          error: 'offline',
          message: 'أنت غير متصل بالإنترنت. البيانات المعروضة قد تكون قديمة.',
        }),
        {
          status: 503,
          headers: { 'Content-Type': 'application/json' },
        }
      )
    }
    
    throw error
  }
}

// Cache-first strategy for static content
async function cacheFirst(request) {
  const cachedResponse = await caches.match(request)
  
  if (cachedResponse) {
    // Update cache in background
    fetch(request).then((networkResponse) => {
      if (networkResponse.ok) {
        const cache = caches.open(CACHE_NAME)
        cache.then((c) => c.put(request, networkResponse))
      }
    })
    
    return cachedResponse
  }
  
  return networkFirst(request)
}

// Stale-while-revalidate strategy
async function staleWhileRevalidate(request) {
  const cachedResponse = await caches.match(request)
  
  const networkFetch = fetch(request).then((networkResponse) => {
    if (networkResponse.ok) {
      const cache = caches.open(CACHE_NAME)
      cache.then((c) => c.put(request, networkResponse.clone()))
    }
    return networkResponse
  })
  
  return cachedResponse || networkFetch
}

// Determine caching strategy based on request
function getCacheStrategy(request) {
  const url = new URL(request.url)
  
  // Network-first for API routes that need fresh data
  if (NETWORK_FIRST_ROUTES.some(route => url.pathname.includes(route))) {
    return networkFirst(request)
  }
  
  // Cache-first for static assets and reference data
  if (CACHE_FIRST_ROUTES.some(route => url.pathname.includes(route)) ||
      request.destination === 'image' ||
      request.destination === 'script' ||
      request.destination === 'style') {
    return cacheFirst(request)
  }
  
  // Stale-while-revalidate for other requests
  return staleWhileRevalidate(request)
}

// Fetch event - handle all requests
self.addEventListener('fetch', (event) => {
  const { request } = event
  
  // Skip non-GET requests
  if (request.method !== 'GET') {
    return
  }
  
  // Skip chrome-extension and other protocols
  if (!request.url.startsWith('http')) {
    return
  }
  
  event.respondWith(getCacheStrategy(request))
})

// Background sync for offline actions
self.addEventListener('sync', (event) => {
  console.log('[SW] Background sync:', event.tag)
  
  if (event.tag === 'sync-sales') {
    event.waitUntil(syncOfflineSales())
  }
  
  if (event.tag === 'sync-payments') {
    event.waitUntil(syncOfflinePayments())
  }
})

// Sync offline sales
async function syncOfflineSales() {
  try {
    const offlineSales = await getOfflineData('offline-sales')
    
    for (const sale of offlineSales) {
      await fetch('/api/v1/sales', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(sale),
      })
    }
    
    // Clear synced data
    await clearOfflineData('offline-sales')
    console.log('[SW] Offline sales synced successfully')
  } catch (error) {
    console.error('[SW] Failed to sync offline sales:', error)
  }
}

// Sync offline payments
async function syncOfflinePayments() {
  try {
    const offlinePayments = await getOfflineData('offline-payments')
    
    for (const payment of offlinePayments) {
      await fetch('/api/v1/payments', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payment),
      })
    }
    
    // Clear synced data
    await clearOfflineData('offline-payments')
    console.log('[SW] Offline payments synced successfully')
  } catch (error) {
    console.error('[SW] Failed to sync offline payments:', error)
  }
}

// IndexedDB helpers for offline storage
async function getOfflineData(storeName) {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('partflow-offline', 1)
    
    request.onerror = () => reject(request.error)
    request.onsuccess = () => {
      const db = request.result
      const transaction = db.transaction(storeName, 'readonly')
      const store = transaction.objectStore(storeName)
      const getAll = store.getAll()
      
      getAll.onsuccess = () => resolve(getAll.result)
      getAll.onerror = () => reject(getAll.error)
    }
    
    request.onupgradeneeded = () => {
      const db = request.result
      if (!db.objectStoreNames.contains(storeName)) {
        db.createObjectStore(storeName, { keyPath: 'id' })
      }
    }
  })
}

async function clearOfflineData(storeName) {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('partflow-offline', 1)
    
    request.onerror = () => reject(request.error)
    request.onsuccess = () => {
      const db = request.result
      const transaction = db.transaction(storeName, 'readwrite')
      const store = transaction.objectStore(storeName)
      const clear = store.clear()
      
      clear.onsuccess = () => resolve()
      clear.onerror = () => reject(clear.error)
    }
  })
}

// Push notification handling
self.addEventListener('push', (event) => {
  console.log('[SW] Push notification received')
  
  const options = {
    body: event.data ? event.data.text() : 'New notification',
    icon: '/icons/icon-192x192.png',
    badge: '/icons/icon-96x96.png',
    vibrate: [200, 100, 200],
    data: {
      dateOfArrival: Date.now(),
      primaryKey: 1,
    },
  }
  
  event.waitUntil(
    self.registration.showNotification('PartFlow', options)
  )
})

// Notification click handling
self.addEventListener('notificationclick', (event) => {
  console.log('[SW] Notification clicked')
  
  event.notification.close()
  
  event.waitUntil(
    clients.openWindow('/')
  )
})