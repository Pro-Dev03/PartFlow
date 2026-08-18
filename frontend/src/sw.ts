// Service Worker for PartFlow PWA
// This file is automatically registered by Vite PWA plugin

// Type declarations for Service Worker
interface ExtendableEvent extends Event {
  waitUntil(promise: Promise<any>): void;
}

interface Client {
  url: string;
  id: string;
  frameType: string;
}

interface Clients {
  claim(): Promise<any>;
  matchAll(options?: { includeUncontrolled?: boolean }): Promise<Client[]>;
}

interface ServiceWorkerGlobalScope {
  skipWaiting(): Promise<void>;
  clients: Clients;
  addEventListener(type: 'install', listener: (event: ExtendableEvent) => void): void;
  addEventListener(type: 'activate', listener: (event: ExtendableEvent) => void): void;
  addEventListener(type: 'fetch', listener: (event: FetchEvent) => void): void;
}

interface FetchEvent extends Event {
  request: Request;
  respondWith(promise: Promise<Response> | Response): void;
}

declare const self: ServiceWorkerGlobalScope;

// Precache important assets
self.addEventListener('install', (event: ExtendableEvent) => {
  console.log('[Service Worker] Install');
  self.skipWaiting();
});

self.addEventListener('activate', (event: ExtendableEvent) => {
  console.log('[Service Worker] Activate');
  event.waitUntil(
    Promise.all([
      self.clients.claim(),
      caches.keys().then((cacheNames: string[]) => {
        return Promise.all(
          cacheNames
            .filter(cacheName => cacheName !== 'api-cache' && 
                                cacheName !== 'image-cache' && 
                                cacheName !== 'static-cache')
            .map(cacheName => caches.delete(cacheName))
        );
      })
    ])
  );
});

// Handle API requests with Network First strategy
self.addEventListener('fetch', (event: FetchEvent) => {
  const url = new URL(event.request.url);
  
  // API calls - Network First with timeout
  if (url.pathname.startsWith('/api/')) {
    event.respondWith(
      networkFirst(event.request)
    );
  }
  // Static assets - Cache First
  else if (url.pathname.match(/\.(png|jpg|jpeg|svg|gif|webp|ico)$/)) {
    event.respondWith(
      cacheFirst(event.request)
    );
  }
  // JS/CSS files - Stale While Revalidate
  else if (url.pathname.match(/\.(js|css)$/)) {
    event.respondWith(
      staleWhileRevalidate(event.request)
    );
  }
});

async function networkFirst(request: Request): Promise<Response> {
  const cache = await caches.open('api-cache');
  
  try {
    // Try network first with timeout
    const networkPromise = fetch(request);
    const timeoutPromise = new Promise((_, reject) => 
      setTimeout(() => reject(new Error('Timeout')), 3000)
    );
    
    const response = await Promise.race([
      networkPromise,
      timeoutPromise
    ]) as Response;
    
    // Cache successful responses
    if (response.ok) {
      cache.put(request, response.clone());
    }
    
    return response;
  } catch (error) {
    // Fallback to cache
    const cachedResponse = await cache.match(request);
    if (cachedResponse) {
      return cachedResponse;
    }
    
    // Return offline fallback
    return new Response(JSON.stringify({
      success: false,
      error: {
        code: 'OFFLINE',
        message: 'أنت غير متصل بالإنترنت'
      }
    }), {
      status: 503,
      headers: { 'Content-Type': 'application/json' }
    });
  }
}

async function cacheFirst(request: Request): Promise<Response> {
  const cache = await caches.open('image-cache');
  const cachedResponse = await cache.match(request);
  
  if (cachedResponse) {
    return cachedResponse;
  }
  
  try {
    const networkResponse = await fetch(request);
    if (networkResponse.ok) {
      cache.put(request, networkResponse.clone());
    }
    return networkResponse;
  } catch (error) {
    return new Response('Image not available', { status: 404 });
  }
}

async function staleWhileRevalidate(request: Request): Promise<Response> {
  const cache = await caches.open('static-cache');
  const cachedResponse = await cache.match(request);
  
  const networkPromise = fetch(request).then(networkResponse => {
    if (networkResponse.ok) {
      cache.put(request, networkResponse.clone());
    }
    return networkResponse;
  });
  
  return cachedResponse || (await networkPromise);
}