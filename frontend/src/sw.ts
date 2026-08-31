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
self.addEventListener('install', (_event: ExtendableEvent) => {
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
            .filter(cacheName => cacheName !== 'image-cache' &&
                                cacheName !== 'static-cache')
            .map(cacheName => caches.delete(cacheName))
        );
      })
    ])
  );
});

// Handle API requests without caching. API responses contain authenticated
// business data and must never be replayed from a shared service-worker cache.
self.addEventListener('fetch', (event: FetchEvent) => {
  const url = new URL(event.request.url);
  
  // API calls - network only
  if (url.pathname.startsWith('/api/')) {
    event.respondWith(
      fetch(event.request).catch(() => new Response(JSON.stringify({
        success: false,
        error: { code: 'OFFLINE', message: 'Offline' }
      }), {
        status: 503,
        headers: { 'Content-Type': 'application/json' }
      }))
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
