# PartFlow Frontend - Development Guide

## 📋 نظرة عامة

هذا الدليل يشرح البنية الجديدة والمحسنة لـ Frontend بعد إعادة التنظيم والصيانة.

## 🏗️ البية التحتية

### Directory Structure

```
frontend/
├── src/
│   ├── app/                    # Application-level code
│   │   ├── App.tsx            # Main App component
│   │   └── router/            # Routing configuration
│   │       ├── index.tsx      # Main router with lazy loading
│   │       └── lazyRoutes.tsx # Lazy route definitions
│   ├── components/            # Shared UI components
│   │   ├── ui/               # Basic UI components (Button, Card, etc.)
│   │   ├── accessibility/    # Accessibility features
│   │   ├── auth/             # Authentication components
│   │   ├── barcode/          # Barcode scanning components
│   │   ├── charts/           # Chart components
│   │   ├── command-palette/  # Command palette
│   │   ├── dashboard/        # Dashboard-specific components
│   │   ├── dialog/           # Dialog/Modal components
│   │   ├── feedback/         # Feedback components (Toast, EmptyState)
│   │   ├── forms/            # Form components
│   │   ├── inventory/        # Inventory-specific components
│   │   ├── mobile/           # Mobile-specific components
│   │   ├── navigation/       # Navigation components
│   │   ├── print/            # Print templates
│   │   ├── search/           # Search components
│   │   ├── shortcuts/        # Keyboard shortcuts
│   │   ├── tables/           # Table components
│   │   ├── theme/            # Theme provider
│   │   ├── toast/            # Toast notifications
│   │   └── undo/             # Undo functionality
│   ├── features/             # Feature-based architecture
│   │   ├── audit/           # Audit logs feature
│   │   ├── auth/            # Authentication feature
│   │   ├── barcode/         # Barcode scanning feature
│   │   ├── customers/       # Customer management
│   │   ├── dashboard/       # Dashboard feature
│   │   ├── debts/           # Debt management
│   │   ├── expenses/        # Expense tracking
│   │   ├── import-export/   # Import/Export functionality
│   │   ├── inventory/       # Inventory management
│   │   ├── notifications/   # Notifications
│   │   ├── onboarding/      # User onboarding
│   │   ├── purchases/       # Purchase management
│   │   ├── reports/         # Reports and analytics
│   │   ├── returns/         # Returns management
│   │   ├── sales/           # Sales/POS feature
│   │   ├── search/          # Global search
│   │   ├── settings/        # Application settings
│   │   ├── shortcuts/       # Keyboard shortcuts
│   │   ├── suppliers/       # Supplier management
│   │   └── warranties/      # Warranty management
│   ├── hooks/               # Custom React hooks
│   │   ├── api/            # API-related hooks
│   │   ├── useServiceWorker.ts
│   │   └── usePerformance.ts
│   ├── layouts/             # Layout components
│   │   ├── DesktopLayout/  # Desktop layout
│   │   └── MobileLayout/   # Mobile layout
│   ├── services/            # External services
│   │   └── api/           # API client and configuration
│   ├── shared/             # Shared utilities and types
│   │   └── types/         # Global type definitions
│   ├── stores/             # Zustand state management
│   │   ├── uiStore.ts     # UI state (theme, language, modals, etc.)
│   │   ├── cartStore.ts   # Shopping cart state
│   │   └── index.ts       # Centralized exports
│   ├── styles/            # Global styles
│   ├── utils/             # Utility functions
│   ├── i18n/              # Internationalization
│   └── assets/            # Static assets
├── public/                # Public files
├── dist/                  # Build output
├── package.json          # Dependencies
├── tsconfig.json         # TypeScript configuration
├── tailwind.config.js    # Tailwind CSS configuration
├── vite.config.ts        # Vite configuration
└── AGENTS.md            # This file
```

## 🔧 Configuration Files

### TypeScript Configuration

- **File**: `tsconfig.json`
- **Strict Mode**: Disabled (can be enabled gradually)
- **Path Aliases**: Configured for cleaner imports
  - `@/*` → `./src/*`
  - `@components/*` → `./src/components/*`
  - `@features/*` → `./src/features/*`
  - `@stores/*` → `./src/stores/*`
  - `@shared/*` → `./src/shared/*`
  - And more...

### Vite Configuration

- **File**: `vite.config.ts`
- **Plugins**: React, PWA
- **Code Splitting**: Manual chunks for vendor, UI, forms, charts, utils
- **Path Aliases**: Same as TypeScript for consistency
- **Dev Server**: Port 3000, host 0.0.0.0

### Tailwind Configuration

- **File**: `tailwind.config.js`
- **Theme**: Extended with custom colors, spacing, animations
- **Dark Mode**: Class-based
- **Custom Colors**:
  - `primary`, `success`, `warning`, `danger`, `info`
  - Semantic colors: `background`, `surface`, `text`, `border`
  - Dark theme colors: `bg-dark`, `surface-dark`, `text-dark`
  - Legacy colors: `ink`, `paper`, `seal` (for compatibility)

## 📦 State Management

### Zustand Stores

#### UI Store (`src/stores/uiStore.ts`)

Manages UI-related state:
- **Theme**: `light | dark | system`
- **Language**: `ar | he | en`
- **Sidebar**: Collapsed state
- **Modals**: Active modal and data
- **Notifications**: Toast notifications
- **Loading**: Global loading state
- **Filters**: Active filters
- **Selection**: Selected items

**Usage**:
```typescript
import { useUIStore } from '@stores'

const { theme, setTheme, addNotification } = useUIStore()
```

#### Cart Store (`src/stores/cartStore.ts`)

Manages shopping cart state:
- **Items**: Cart items with product details
- **Customer**: Selected customer
- **Notes**: Order notes
- **Computed**: Subtotal, total, discount, profit

**Usage**:
```typescript
import { useCartStore } from '@stores'

const { items, addItem, getTotal, getTotalProfit } = useCartStore()
```

## 🎨 Component Architecture

### Feature-Based Structure

Each feature follows this pattern:
```
feature-name/
├── components/       # Feature-specific components
├── pages/           # Feature pages
├── hooks/           # Feature-specific hooks
├── services/        # Feature API services
├── types/           # Feature types (re-exports from shared)
├── validation/      # Form validation schemas
└── index.tsx        # Main feature export
```

### Shared Components

Located in `src/components/`:
- **UI Components**: Reusable basic components
- **Feature Components**: Feature-specific but shared components

## 🔄 Routing

### Lazy Loading

All routes use lazy loading for better performance:
```typescript
const Dashboard = lazy(() => import('../../features/dashboard').then(m => ({ default: m.Dashboard })))
```

### Loading State

Custom loading component shown during route transitions:
```typescript
<Suspense fallback={<LoadingFallback />}>
  <Routes>
    {/* Routes */}
  </Routes>
</Suspense>
```

## 🌐 Shared Types

### Global Types (`src/shared/types/index.ts`)

Centralized type definitions:
- **Customer**: Customer data structure
- **Product**: Product data structure
- **CartItem**: Shopping cart item
- **Sale**: Sale transaction
- **API**: API response types
- **Filters**: Filter and sort types

**Usage**:
```typescript
import type { Customer, Product, CartItem } from '@shared/types'
```

### Feature Types

Feature type files re-export from shared types:
```typescript
// src/features/customers/types/customer.ts
export type { Customer, CustomerStats, CustomerTimelineEvent } from '@shared/types'
```

## 🚀 Development Commands

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Run linter
npm run lint

# Run tests
npm run test
```

## 🎯 Best Practices

### Component Development

1. **Use Feature-Based Structure**: Create features in `src/features/`
2. **Shared Components**: Put reusable components in `src/components/`
3. **Types**: Use shared types from `@shared/types`
4. **State**: Use Zustand stores from `@stores`
5. **Styling**: Use Tailwind CSS classes
6. **Lazy Loading**: Use lazy loading for routes and heavy components

### State Management

1. **UI State**: Use `useUIStore` for theme, language, modals, notifications
2. **Cart State**: Use `useCartStore` for shopping cart
3. **Server State**: Use React Query for API data
4. **Form State**: Use React Hook Form for forms

### TypeScript

1. **Types**: Import from `@shared/types` when possible
2. **Strict Mode**: Currently disabled, enable gradually
3. **Path Aliases**: Use configured aliases for cleaner imports

### Styling

1. **Tailwind**: Use Tailwind classes for styling
2. **Components**: Use shared UI components (Button, Card, etc.)
3. **Responsive**: Use Tailwind responsive prefixes
4. **Dark Mode**: Use dark mode classes when needed

## 🔍 Troubleshooting

### Build Errors

**Error**: "Could not resolve entry module 'date-fns'"
- **Solution**: Already fixed - removed from vite.config.ts

**Error**: "Module not found: Can't resolve '@/shared/types'"
- **Solution**: Ensure path aliases are configured in both tsconfig.json and vite.config.ts

### Runtime Errors

**Error**: "useUIStore is not defined"
- **Solution**: Import from `@stores` instead of individual files

**Error**: "Type 'Customer' is not assignable"
- **Solution**: Ensure types are imported from `@shared/types`

## 📝 Recent Changes

### Fixed Issues

1. **Removed duplicate state management system**
   - Deleted `src/store/useStore.ts`
   - Using only `src/stores/` directory

2. **Created shared types directory**
   - Added `src/shared/types/index.ts`
   - Centralized all common type definitions
   - Feature types now re-export from shared

3. **Enabled lazy loading**
   - Updated router to use lazy loading
   - Added loading fallback component
   - Improved initial load performance

4. **Fixed Tailwind config**
   - Removed duplicate color definitions
   - Organized colors properly
   - Added opacity utilities

5. **Converted Dashboard to Tailwind**
   - Removed inline styles
   - Now uses Tailwind classes
   - Consistent with other components

6. **Updated path aliases**
   - Added `@shared/*` alias
   - Configured in both tsconfig.json and vite.config.ts

### Architecture Improvements

1. **Better Code Organization**
   - Feature-based structure
   - Shared utilities and types
   - Clear separation of concerns

2. **Improved Performance**
   - Lazy loading for routes
   - Code splitting in build
   - Optimized bundle size

3. **Type Safety**
   - Centralized type definitions
   - Reduced type duplication
   - Better TypeScript support

4. **Developer Experience**
   - Cleaner imports with aliases
   - Consistent component structure
   - Better documentation

## 🚦 Next Steps

### Short Term

1. Enable TypeScript strict mode gradually
2. Add more shared types as needed
3. Implement missing feature routes
4. Add error boundaries

### Long Term

1. Add comprehensive testing
2. Implement performance monitoring
3. Add storybook for components
4. Improve accessibility

## 📚 Additional Resources

- **React Documentation**: https://react.dev
- **TypeScript**: https://www.typescriptlang.org
- **Tailwind CSS**: https://tailwindcss.com
- **Zustand**: https://zustand-demo.pmnd.rs
- **React Query**: https://tanstack.com/query/latest
- **Vite**: https://vitejs.dev

---

**Last Updated**: 2026-08-18
**Version**: 1.0.0
