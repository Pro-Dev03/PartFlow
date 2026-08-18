# تقرير بنية وهيكلية مشروع PartFlow

## نظرة عامة

PartFlow هو نظام إدارة مخازن متكامل مبني ببنية Modular Monolith، يدعم Multi-tenant، متعدد اللغات، ومصمم للعمل على منصات متعددة (Web, PWA, Mobile).

---

## البنية العامة للمشروع

```
PartFlow/
├── backend/              # Backend API (Go)
├── frontend/             # Frontend Application (React)
├── worker/               # Background Jobs (Go)
├── docker/               # Docker configurations
├── docs/                 # Documentation
├── scripts/              # Utility scripts
├── render.yaml          # Render deployment config
├── docker-compose.yml   # Local development
└── AGENTS.md           # Developer guide
```

---

## Backend Architecture

### التقنيات المستخدمة

- **اللغة**: Go 1.21
- **Framework**: Gin (HTTP Web Framework)
- **قاعدة البيانات**: PostgreSQL + Supabase
- **ORM**: sqlx (Type-safe SQL)
- **المصادقة**: JWT (golang-jwt/jwt/v5)
- **Validation**: go-playground/validator
- **Logging**: zerolog
- **التشفير**: golang.org/x/crypto

### هيكل المشروع

```
backend/
├── cmd/
│   └── api/
│       └── main.go           # نقطة الدخول الرئيسية
├── internal/                 # كود التطبيق الداخلي
│   ├── auth/                 # نظام المصادقة
│   ├── organizations/        # إدارة المؤسسات
│   ├── users/                # إدارة المستخدمين
│   ├── roles/                # إدارة الصلاحيات
│   ├── products/             # إدارة المنتجات
│   ├── inventory/            # إدارة المخزون
│   ├── sales/                # إدارة المبيعات
│   ├── payments/             # إدارة المدفوعات
│   ├── customers/            # إدارة العملاء
│   ├── debts/                # إدارة الديون
│   ├── purchases/            # إدارة المشتريات
│   ├── suppliers/            # إدارة الموردين
│   ├── expenses/             # إدارة المصروفات
│   ├── returns/              # إدارة المرتجعات
│   ├── warranties/           # إدارة الضمانات
│   ├── inspections/          # إدارة الفحوصات
│   ├── reports/              # إدارة التقارير
│   ├── notifications/        # إدارة الإشعارات
│   ├── automation/           # الأتمتة
│   └── audit/                # تدقيق العمليات
├── pkg/                      # حزم مشتركة
│   ├── errors/               # معالجة الأخطاء
│   ├── logger/               # نظام التسجيل
│   ├── middleware/           # البرمجيات الوسيطة
│   ├── response/             # استجابات HTTP موحدة
│   └── validator/            # التحقق من البيانات
├── migrations/               # ترحيل قاعدة البيانات
├── tests/                    # الاختبارات
├── Dockerfile               # صورة Docker
├── .dockerignore           # ملفات مستثناة من Docker
└── go.mod                  # تبعيات Go
```

### نمط البنية المعمارية (Architecture Pattern)

يتبع Backend نمط **Repository-Service-Handler** لكل Domain:

#### هيكل كل Domain

```
internal/{domain}/
├── model/           # نماذج البيانات (Database Models)
├── repository/      # طبقة الوصول للبيانات (Data Access Layer)
├── service/         # منطق العمل (Business Logic Layer)
├── handler/         # معالجة HTTP (HTTP Handler Layer)
├── dto/             # كائنات نقل البيانات (Data Transfer Objects)
├── validator/       # التحقق من البيانات (Validation Logic)
└── errors/          # معالجة الأخطاء المخصصة (Custom Errors)
```

#### توزيع المسؤوليات

| الطبقة | المسؤولية |
|--------|-----------|
| **Model** | تعريف هيكل البيانات وتعيين قاعدة البيانات |
| **Repository** | العمليات CRUD والاستعلامات المعقدة |
| **Service** | منطق العمل، المعاملات، والتحقق |
| **Handler** | استقبال الطلبات HTTP والاستجابة |
| **DTO** | تحويل البيانات بين الطبقات |
| **Validator** | التحقق من صحة البيانات المدخلة |
| **Errors** | تعريف ومعالجة الأخطاء المخصصة |

### تدفق الطلب (Request Flow)

```
HTTP Request
    ↓
Middleware (CORS, Auth, Logging)
    ↓
Handler (Parse Request)
    ↓
Validator (Validate DTO)
    ↓
Service (Business Logic)
    ↓
Repository (Database Operations)
    ↓
Database (PostgreSQL)
    ↓
Response (JSON)
```

### نظام Multi-tenant

كل بيانات الأعمال مرتبطة بـ `organization_id`:

```go
type BaseModel struct {
    ID             uuid.UUID `json:"id"`
    OrganizationID uuid.UUID `json:"organization_id"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
}
```

### قاعدة البيانات

#### التكنولوجيا
- **PostgreSQL 15** كقاعدة بيانات رئيسية
- **Supabase** للاستضافة والخدمات الإضافية
- **Row Level Security (RLS)** لعزل البيانات

#### Extensions
- `uuid-ossp` لتوليد UUIDs
- `pgcrypto` للتشفير

#### Connection Pooling
- استخدام Connection Pooling لتحسين الأداء
- إدارة الاتصالات عبر sqlx

### الأمان

#### المصادقة (Authentication)
- JWT Token-based authentication
- HMAC signing algorithm
- Token expiration management

#### التفويض (Authorization)
- Role-based access control (RBAC)
- Permission system
- Organization-level isolation

#### حماية البيانات
- التشفير للبيانات الحساسة
- SQL Injection prevention (via sqlx)
- XSS protection

---

## Frontend Architecture

### التقنيات المستخدمة

- **Framework**: React 18.2
- **Language**: TypeScript 5.0
- **Build Tool**: Vite 4.4
- **Routing**: React Router DOM 6.15
- **State Management**: Zustand 4.4
- **Server State**: React Query 3.39
- **Forms**: React Hook Form 7.45
- **Validation**: Zod 3.22
- **HTTP Client**: Axios 1.5
- **Styling**: Tailwind CSS 3.3
- **Internationalization**: i18next 23.4
- **Charts**: Recharts 2.8
- **Icons**: Lucide React 0.279
- **PWA**: vite-plugin-pwa 0.16
- **Testing**: Vitest 0.34

### هيكل المشروع

```
frontend/
├── src/
│   ├── app/                  # إعدادات التطبيق
│   │   ├── App.tsx          # المكون الرئيسي
│   │   └── router/          # تكوين المسارات
│   ├── features/            # الميزات (Feature-based)
│   │   ├── auth/            # المصادقة
│   │   ├── dashboard/       # لوحة التحكم
│   │   ├── sales/           # المبيعات
│   │   ├── purchases/       # المشتريات
│   │   ├── inventory/       # المخزون
│   │   ├── products/        # المنتجات
│   │   ├── customers/       # العملاء
│   │   ├── suppliers/       # الموردين
│   │   ├── debts/           # الديون
│   │   ├── expenses/        # المصروفات
│   │   ├── reports/         # التقارير
│   │   ├── settings/        # الإعدادات
│   │   └── barcode/         # الباركود
│   ├── components/         # مكونات مشتركة
│   │   ├── navigation/      # التنقل
│   │   ├── forms/           # النماذج
│   │   ├── tables/          # الجداول
│   │   ├── ui/              # عناصر UI
│   │   └── feedback/        # Feedback UI
│   ├── layouts/            # التخطيطات
│   │   ├── DesktopLayout/   # تخطيط سطح المكتب
│   │   └── MobileLayout/    # تخطيط الموبايل
│   ├── hooks/              # Custom Hooks
│   ├── services/           # API Services
│   ├── stores/             # State Stores (Zustand)
│   ├── types/              # TypeScript Types
│   ├── utils/              # Utility Functions
│   ├── i18n/               # Internationalization
│   ├── styles/             # Global Styles
│   └── assets/             # Static Assets
├── public/                 # Static Files
├── nginx.conf             # Nginx Configuration
├── Dockerfile            # Docker Image
├── .dockerignore         # Docker Exclusions
└── package.json          # Dependencies
```

### نمط البنية المعمارية (Architecture Pattern)

يتبع Frontend نمط **Feature-based Architecture**:

#### هيكل كل Feature

```
src/features/{feature}/
├── components/      # مكونات UI الخاصة بالـFeature
├── hooks/          # Custom Hooks الخاصة بالـFeature
├── services/       # API Services للـFeature
├── types/          # TypeScript Types للـFeature
├── validation/     # Zod Schemas للتحقق
└── index.tsx       # نقطة الدخول للـFeature
```

#### توزيع المسؤوليات

| المجلد | المسؤولية |
|--------|-----------|
| **components** | مكونات UI قابلة لإعادة الاستخدام |
| **hooks** | Custom Hooks للمنطق القابل لإعادة الاستخدام |
| **services** | اتصالات API وطلبات HTTP |
| **types** | تعريفات TypeScript |
| **validation** | Schemas للتحقق من النماذج |

### إدارة الحالة (State Management)

#### أنواع الحالة

```
┌─────────────────────────────────────────┐
│           State Management              │
├─────────────────────────────────────────┤
│  1. Server State (React Query)          │
│     - بيانات من API                    │
│     - caching & synchronization        │
│                                        │
│  2. Global State (Zustand)             │
│     - إعدادات التطبيق                  │
│     - حالة المستخدم                   │
│     - موضوع الواجهة                    │
│                                        │
│  3. Local State (useState)             │
│     - حالة المكونات                   │
│     - بيانات مؤقتة                    │
│                                        │
│  4. Form State (React Hook Form)       │
│     - بيانات النماذج                  │
│     - حالة التحقق                      │
└─────────────────────────────────────────┘
```

### التوجيه (Routing)

#### بنية المسارات

```typescript
/                          # Login/Register
/dashboard                 # Dashboard
/sales                     # Sales Management
  /sales/new              # New Sale
  /sales/:id              # Sale Details
/purchases                # Purchases Management
/inventory                 # Inventory Management
/products                  # Products Management
/customers                 # Customers Management
/suppliers                 # Suppliers Management
/debts                     # Debts Management
/expenses                  # Expenses Management
/reports                   # Reports
/settings                  # Settings
```

### التصميم المتجاوب (Responsive Design)

#### التخطيطات

```
┌─────────────────────────────────────┐
│         Desktop Layout              │
├─────────────────────────────────────┤
│  ┌──────┬──────────────────────┐   │
│  │      │                      │   │
│  │ Side │      Main Content    │   │
│  │ bar  │                      │   │
│  │      │                      │   │
│  └──────┴──────────────────────┘   │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│         Mobile Layout               │
├─────────────────────────────────────┤
│  ┌──────────────────────────────┐   │
│  │         Top Bar              │   │
│  ├──────────────────────────────┤   │
│  │                              │   │
│  │      Main Content            │   │
│  │                              │   │
│  ├──────────────────────────────┤   │
│  │       Bottom Nav             │   │
│  └──────────────────────────────┘   │
└─────────────────────────────────────┘
```

### دعم اللغات (Internationalization)

#### اللغات المدعومة
- العربية (AR) - RTL
- العبرية (HE) - RTL
- الإنجليزية (EN) - LTR

#### التنفيذ
```typescript
// i18next configuration
i18n.init({
  lng: 'ar',
  fallbackLng: 'en',
  resources: {
    ar: { translation: arabicTranslations },
    he: { translation: hebrewTranslations },
    en: { translation: englishTranslations }
  }
})
```

### Progressive Web App (PWA)

#### المزايا
- تثبيت التطبيق على الجهاز
- العمل بدون اتصال (Offline)
- إشعارات Push
- تحديثات تلقائية

#### الإعداد
```typescript
// vite.config.ts
import { VitePWA } from 'vite-plugin-pwa'

export default defineConfig({
  plugins: [
    VitePWA({
      registerType: 'autoUpdate',
      workbox: {
        globPatterns: ['**/*.{js,css,html,ico,png,svg}']
      }
    })
  ]
})
```

---

## تكامل Backend-Frontend

### API Communication

#### Architecture
```
Frontend (React)
    ↓ HTTP/HTTPS
    ↓ JSON
Backend (Go/Gin)
    ↓
PostgreSQL
```

#### API Client (Axios)

```typescript
// src/services/api.ts
import axios from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Request interceptor
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Response interceptor
api.interceptors.response.use(
  (response) => response,
  (error) => {
    // Handle errors
    return Promise.reject(error)
  }
)
```

### Authentication Flow

```
1. User Login
   ↓
2. Frontend sends credentials
   ↓
3. Backend validates and returns JWT
   ↓
4. Frontend stores token
   ↓
5. Frontend includes token in requests
   ↓
6. Backend validates token
   ↓
7. Access granted/denied
```

---

## Worker Architecture

### الغرض
- معالجة المهام في الخلفية
- إرسال الإشعارات
- معالجة التقارير المخططة
- مزامنة البيانات

### التقنيات
- Go 1.21
- Redis (Queue)
- PostgreSQL

### الهيكل
```
worker/
├── main.go          # نقطة الدخول
├── jobs/            # تعريف المهام
├── processors/      # معالجات المهام
└── Dockerfile      # صورة Docker
```

---

## البنية التحتية (Infrastructure)

### Docker Configuration

#### Backend Dockerfile
```dockerfile
# Multi-stage build
FROM golang:1.21-alpine AS builder
# Build stage
FROM alpine:latest
# Runtime stage
```

#### Frontend Dockerfile
```dockerfile
# Multi-stage build
FROM node:18-alpine AS builder
# Build stage
FROM nginx:alpine
# Serve with nginx
```

### Docker Compose (Local Development)

```yaml
services:
  backend:      # Go API
  frontend:     # React App
  postgres:     # Database
  redis:        # Cache/Queue
  worker:       # Background jobs
```

### Render Deployment

```yaml
services:
  - type: pserv    # PostgreSQL
  - type: pserv    # Redis
  - type: web      # Backend
  - type: web      # Frontend
  - type: worker   # Background jobs
```

---

## أنماط التصميم المستخدمة (Design Patterns)

### Backend Patterns

1. **Repository Pattern**
   - فصل منطق الوصول للبيانات
   - تسهيل الاختبار
   - إمكانية تبديل مصدر البيانات

2. **Service Layer Pattern**
   - عزل منطق العمل
   - إعادة استخدام الكود
   - سهولة الصيانة

3. **Dependency Injection**
   - حقن التبعيات عبر constructors
   - تحسين قابلية الاختبار
   - فصل المسؤوليات

4. **Middleware Pattern**
   - معالجة الطلبات بشكل متسلسل
   - إعادة استخدام البرمجيات الوسيطة
   - فصل الاهتمامات

### Frontend Patterns

1. **Container/Presentational Pattern**
   - فصل المنطق عن العرض
   - إعادة استخدام المكونات
   - سهولة الاختبار

2. **Custom Hooks Pattern**
   - إعادة استخدام المنطق
   - فصل الاهتمامات
   - تنظيم الكود

3. **Feature-based Pattern**
   - تنظيم حسب الميزات
   - سهولة الإضافة والحذف
   - عزل المكونات

4. **Composition Pattern**
   - بناء مكونات معقدة من بسيطة
   - إعادة الاستخدام
   - المرونة

---

## الأداء والتحسين (Performance)

### Backend Optimization

1. **Database Optimization**
   - Indexes على الأعمدة المهمة
   - Connection Pooling
   - Query Optimization

2. **Caching Strategy**
   - Redis caching
   - Response caching
   - Database query caching

3. **Code Optimization**
   - Goroutines للعمليات المتوازية
   - Efficient data structures
   - Memory management

### Frontend Optimization

1. **Code Splitting**
   - Lazy loading
   - Route-based splitting
   - Dynamic imports

2. **Asset Optimization**
   - Image optimization
   - Minification
   - Gzip compression

3. **Rendering Optimization**
   - React.memo
   - useMemo
   - useCallback

4. **Network Optimization**
   - API caching (React Query)
   - Request batching
   - Optimistic updates

---

## الأمان (Security)

### Backend Security

1. **Authentication**
   - JWT tokens
   - Secure token storage
   - Token expiration

2. **Authorization**
   - RBAC
   - Permission checks
   - Organization isolation

3. **Data Protection**
   - Encryption at rest
   - Encryption in transit (TLS)
   - Input validation

4. **API Security**
   - Rate limiting
   - CORS configuration
   - SQL injection prevention

### Frontend Security

1. **XSS Prevention**
   - React's built-in escaping
   - Content Security Policy
   - Input sanitization

2. **CSRF Protection**
   - SameSite cookies
   - CSRF tokens
   - Origin validation

3. **Data Validation**
   - Client-side validation (Zod)
   - Server-side validation
   - Type safety (TypeScript)

---

## الاختبار (Testing)

### Backend Testing

```go
// Unit tests
func TestService_Create(t *testing.T) {
    // Test business logic
}

// Integration tests
func TestHandler_CreateSale(t *testing.T) {
    // Test HTTP handlers
}
```

### Frontend Testing

```typescript
// Component tests
describe('SalesComponent', () => {
  it('renders correctly', () => {})
})

// Integration tests
describe('SalesFlow', () => {
  it('creates sale successfully', () => {})
})
```

---

## المراقبة والتسجيل (Monitoring & Logging)

### Backend Monitoring

1. **Logging**
   - Structured logging (zerolog)
   - Log levels
   - Contextual information

2. **Metrics**
   - Request metrics
   - Database metrics
   - Error tracking

3. **Audit Trail**
   - User actions
   - Data changes
   - System events

### Frontend Monitoring

1. **Error Tracking**
   - Runtime errors
   - Network errors
   - User feedback

2. **Performance Monitoring**
   - Page load time
   - API response time
   - User interactions

---

## قابلية التوسع (Scalability)

### Horizontal Scaling

1. **Backend**
   - Stateless services
   - Load balancing
   - Database connection pooling

2. **Frontend**
   - CDN distribution
   - Static asset caching
   - PWA offline support

### Vertical Scaling

1. **Database**
   - Read replicas
   - Connection pooling
   - Query optimization

2. **Application**
   - Resource optimization
   - Memory management
   - CPU utilization

---

## الخلاصة

PartFlow مبني ببنية معمارية حديثة وقابلة للتوسع:

### المزايا الرئيسية

1. **Modular Architecture**
   - فصل واضح بين المكونات
   - سهولة الصيانة
   - إمكانية التوسع

2. **Multi-tenant Support**
   - عزل البيانات
   - إدارة متعددة المحلات
   - تكامل مرن

3. **Modern Tech Stack**
   - Go للـ Backend (أداء عالي)
   - React للـ Frontend (تجربة مستخدم ممتازة)
   - PostgreSQL (قاعدة بيانات قوية)

4. **Production Ready**
   - Docker support
   - Render deployment
   - PWA support

5. **Developer Experience**
   - TypeScript safety
   - Clear structure
   - Comprehensive documentation

### الجاهزية للإنتاج

المشروع جاهز للنشر مع:
- إعداد Docker كامل
- تكوين Render
- دليل نشر شامل
- أفضل ممارسات الأمان
- استراتيجيات الاختبار

---

## المستقبل

### تحسينات محتملة

1. **Backend**
   - إضافة GraphQL API
   - WebSocket support
   - Advanced caching

2. **Frontend**
   - Mobile app (React Native)
   - Desktop app (Electron)
   - Advanced analytics

3. **Infrastructure**
   - Kubernetes deployment
   - Microservices migration
   - Advanced monitoring

---

**التقرير المعده بواسطة**: Devin AI
**التاريخ**: 2026-08-18
**الإصدار**: 1.0.0
