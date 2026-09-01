import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { VitePWA } from 'vite-plugin-pwa'
import path from 'path'

// https://vite.dev/config/
export default defineConfig({
  base: './',
  server: {
    port: 5174,
    strictPort: true,
  },
  resolve: {
    alias: {
      '@': path.resolve(process.cwd(), './src')
    }
  },
  plugins: [
    react(),
    // VitePWA completely disabled to avoid module loading issues in development
    // Only enable in production build
    ...(process.env.NODE_ENV === 'production' ? [
      VitePWA({
        registerType: 'autoUpdate',
        includeAssets: ['favicon.ico', 'apple-touch-icon.png', 'mask-icon.svg'],
        manifest: {
          name: 'PartFlow',
          short_name: 'PartFlow',
          description: 'نظام إدارة ذكي لمحل قطع الحاسوب',
          theme_color: '#2563eb',
          background_color: '#ffffff',
          display: 'standalone',
          orientation: 'portrait',
          scope: '/',
          start_url: '/',
          icons: [
            {
              src: 'pwa-192x192.png',
              sizes: '192x192',
              type: 'image/png',
              purpose: 'any'
            },
            {
              src: 'pwa-512x512.png',
              sizes: '512x512',
              type: 'image/png',
              purpose: 'any'
            },
            {
              src: 'pwa-192x192.png',
              sizes: '192x192',
              type: 'image/png',
              purpose: 'maskable'
            },
            {
              src: 'pwa-512x512.png',
              sizes: '512x512',
              type: 'image/png',
              purpose: 'maskable'
            }
          ],
          categories: ['business', 'productivity'],
          shortcuts: [
            {
              name: 'نقطة البيع',
              short_name: 'POS',
              description: 'بدء عملية بيع جديدة',
              url: '/pos',
              icons: [{ src: 'pwa-192x192.png', sizes: '192x192' }]
            },
            {
              name: 'المخزون',
              short_name: 'المخزون',
              description: 'عرض المخزون الحالي',
              url: '/inventory',
              icons: [{ src: 'pwa-192x192.png', sizes: '192x192' }]
            },
            {
              name: 'التقارير',
              short_name: 'التقارير',
              description: 'عرض التقارير والإحصائيات',
              url: '/reports',
              icons: [{ src: 'pwa-192x192.png', sizes: '192x192' }]
            }
          ]
        },
        workbox: {
          globPatterns: ['**/*.{js,css,html,ico,png,svg,woff,woff2}'],
          runtimeCaching: [
            {
              urlPattern: /\.(?:png|jpg|jpeg|svg|gif|webp|ico)$/i,
              handler: 'CacheFirst',
              options: {
                cacheName: 'image-cache',
                expiration: {
                  maxEntries: 100,
                  maxAgeSeconds: 60 * 60 * 24 * 30 // 30 days
                }
              }
            },
            {
              urlPattern: /\.(?:js|css)$/i,
              handler: 'StaleWhileRevalidate',
              options: {
                cacheName: 'static-cache',
                expiration: {
                  maxEntries: 100,
                  maxAgeSeconds: 60 * 60 * 24 * 7 // 7 days
                }
              }
            }
          ],
          cleanupOutdatedCaches: true,
          skipWaiting: true,
          clientsClaim: true
        },
        devOptions: {
          enabled: false
        }
      })
    ] : [])
  ],
  build: {
    rollupOptions: {
      output: {
        manualChunks: (id) => {
          if (id.includes('node_modules')) {
            if (id.includes('react') || id.includes('react-dom') || id.includes('react-router-dom')) {
              return 'vendor';
            }
            if (id.includes('lucide-react') || id.includes('clsx') || id.includes('tailwind-merge')) {
              return 'ui';
            }
            if (id.includes('recharts')) {
              return 'charts';
            }
            if (id.includes('react-hook-form') || id.includes('@hookform/resolvers') || id.includes('zod')) {
              return 'forms';
            }
            if (id.includes('i18next') || id.includes('react-i18next') || id.includes('i18next-browser-languagedetector')) {
              return 'i18n';
            }
            if (id.includes('@tanstack/react-query')) {
              return 'queries';
            }
            if (id.includes('date-fns') || id.includes('dayjs')) {
              return 'date';
            }
            if (id.includes('sonner')) {
              return 'notifications';
            }
          }
          // Feature-based chunking with better granularity
          if (id.includes('/features/dashboard/')) {
            return 'dashboard';
          }
          if (id.includes('/features/sales/')) {
            return 'sales';
          }
          if (id.includes('/features/inventory/')) {
            return 'inventory';
          }
          // Split customers into smaller chunks
          if (id.includes('/features/customers/components/FinancialTimeline')) {
            return 'customers-timeline';
          }
          if (id.includes('/features/customers/components/CustomerStats')) {
            return 'customers-stats';
          }
          if (id.includes('/features/customers/')) {
            return 'customers';
          }
          if (id.includes('/features/reports/')) {
            return 'reports';
          }
          if (id.includes('/features/debts/')) {
            return 'debts';
          }
          if (id.includes('/features/purchases/')) {
            return 'purchases';
          }
        }
      }
    },
    chunkSizeWarningLimit: 1000,
    // Enable minification and optimization
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true,
        pure_funcs: ['console.log', 'console.info', 'console.debug'],
      },
      mangle: {
        safari10: true,
      },
    },
    // Enable source maps in production for debugging
    sourcemap: false,
    // Optimize chunk size
    target: 'es2015',
    // Enable CSS code splitting
    cssCodeSplit: true,
  }
})
