# PartFlow - Agent Development Guide

## 📋 معلومات مهمة للمطورين

### أوامر التطوير

#### Backend (Go)
```bash
cd backend
go mod download          # تثبيت التبعيات
go run cmd/api/main.go   # تشغيل الخادم
go test ./...            # تشغيل الاختبارات
go build ./cmd/api       # بناء التطبيق
```

#### Frontend (React)
```bash
cd frontend
npm install              # تثبيت التبعيات
npm run dev              # تشغيل خادم التطوير
npm run build            # بناء للإنتاج
npm run lint             # فحص الكود
npm run test             # تشغيل الاختبارات
```

#### Docker
```bash
docker-compose up              # تشغيل جميع الخدمات
docker-compose up -d           # تشغيل في الخلفية
docker-compose down            # إيقاف الخدمات
docker-compose logs backend    # عرض سجلات Backend
docker-compose logs frontend   # عرض سجلات Frontend
```

### بنية المشروع

#### Backend (Modular Monolith)
كل Domain يحتوي على:
- `model.go` - نماذج البيانات
- `repository.go` - طبقة الوصول للبيانات
- `service.go` - منطق العمل
- `handler.go` - معالجة HTTP
- `dto.go` - كائنات نقل البيانات
- `validator.go` - التحقق من البيانات
- `errors.go` - معالجة الأخطاء

#### Frontend (Feature-based)
كل Feature يحتوي على:
- `components/` - مكونات UI الخاصة بالـFeature
- `hooks/` - Custom Hooks
- `services/` - API Services
- `types/` - TypeScript Types
- `validation/` - Schemas للتحقق

### قواعد التطوير

1. **Backend**
   - اتبع نمط Repository-Service-Handler
   - استخدم Transactions للعمليات المالية
   - قم بتسجيل كل العمليات المهمة في Audit Log
   - طبق RLS على جميع الاستعلامات

2. **Frontend**
   - استخدم Feature-based Architecture
   - افصل Server State عن UI State
   - استخدم TypeScript بشكل صارم
   - دعم RTL/LTR في جميع المكونات

3. **Database**
   - استخدم Migrations للتغييرات
   - طبق Row Level Security
   - استخدم Indexes للتحسين

### متغيرات البيئة المهمة

```bash
DATABASE_URL=postgresql://...
SUPABASE_URL=your-supabase-url
SUPABASE_KEY=your-supabase-key
JWT_SECRET=your-jwt-secret
REDIS_URL=redis://localhost:6379
```

### اختبار النظام

1. تأكد من تشغيل PostgreSQL و Redis
2. ابدأ Backend: `cd backend && go run cmd/api/main.go`
3. ابدأ Frontend: `cd frontend && npm run dev`
4. افتح المتصفح على `http://localhost:3000`

### خطوات إضافة Feature جديد

#### Backend
1. أنشئ مجلد جديد في `internal/`
2. أضف model, repository, service, handler
3. سجل الـroutes في `cmd/api/main.go`
4. أضف migrations للجداول الجديدة

#### Frontend
1. أنشئ مجلد جديد في `src/features/`
2. أضف components, hooks, services, types
3. أضف الـroute في `src/app/router/index.tsx`
4. أضف التنقل في Sidebar/Nav

### الوثائق

- **المتطلبات**: `docs/report.md`
- **Backend**: `docs/report-backend.md`
- **Frontend**: `docs/frontend.md`
- **القراءة**: `docs/README.md`

### النشر

1. بناء Docker images
2. دفع إلى Container Registry
3. نشر على Render أو أي منصة سحابية
4. إعداد environment variables
5. تشغيل database migrations

### النسخ الاحتياطي

- قاعدة البيانات: استخدم Supabase Backup
- الملفات: Supabase Storage
- Logs: Cloud logging service

### النشر على Render

PartFlow جاهز للنشر على Render السحابية:

#### الملفات المطلوبة
- `render.yaml` - تكوين جميع الخدمات
- `backend/Dockerfile` - صورة Backend
- `frontend/Dockerfile` - صورة Frontend
- `worker/Dockerfile` - صورة Worker
- `scripts/deploy-render.sh` - سكريبت النشر

#### خطوات النشر
1. راجع `docs/DEPLOYMENT_RENDER.md` للتفاصيل الكاملة
2. شغل `./scripts/deploy-render.sh --all` للإعداد
3. اذهب إلى Render Dashboard وأنشئ Blueprint
4. أو أنشئ الخدمات يدوياً: Database, Redis, Backend, Frontend, Worker

#### الخدمات على Render
- PostgreSQL Database
- Redis
- Backend Web Service
- Frontend Web Service
- Background Worker

### الملاحظات

- النظام يدعم Multi-tenant منذ البداية
- كل بيانات business مرتبطة بـ `organization_id`
- استخدام PWA لتجربة Mobile أفضل
- دعم RTL/LTR مطلوب لجميع المكونات

### الميزات المتقدمة في Frontend

تم إضافة الميزات المتقدمة التالية:

1. **Advanced Search & Global Search**
   - نظام بحث شامل يدعم Barcode, Serial, SKU, Product, Customer
   - فلاتر متقدمة مع عمليات مختلفة
   - تنقل بالكيبورد

2. **Keyboard Shortcuts System**
   - اختصارات لوحة المفاتيح (F1-F5, Ctrl+K, ESC, etc.)
   - Modal لعرض جميع الاختصارات
   - تسجيل ديناميكي للاختصارات

3. **Form Validation Integration**
   - React Hook Form + Zod integration
   - Real-time و Async validation
   - إدارة حالة الإرسال

4. **State Management**
   - Zustand stores للـ UI State (uiStore, cartStore)
   - React Query للـ Server State
   - API Client موحد مع interceptors
   - نظام state management موحد ومنظم

5. **Offline Support**
   - Service Worker متقدم مع استراتيجيات caching متعددة
   - Background Sync للمبيعات والمدفوعات
   - IndexedDB للبيانات Offline

6. **PWA Configuration**
   - Vite PWA Plugin متقدم
   - Manifest مع shortcuts
   - Workbox runtime caching

7. **Advanced Charts**
   - Recharts integration
   - Line, Bar, Pie, Area charts
   - Multi-line و Stacked charts

8. **Print Templates**
   - Invoice Template احترافي
   - Barcode Labels قابلة للتخصيص
   - دعم RTL

9. **Performance Optimization**
   - Lazy loading للمكونات والمسارات
   - Virtual scrolling
   - Performance monitoring hooks
   - Code splitting في Vite
   - تحسينات في حجم الباندل

### البنية المحسنة للـ Frontend

تم إعادة تنظيم Frontend بشكل احترافي ليسهل التوسع والصيانة:

1. **إصلاح المشاكل الحرجة**
   - إزالة `date-fns` من vite.config.ts (كانت تسبب فشل البناء)
   - حل تضارب State Management - دمج النظامين في نظام واحد
   - توحيد تعريفات الـ Types في shared directory

2. **Shared Types Directory**
   - إنشاء `src/shared/types/index.ts` لتوحيد تعريفات البيانات
   - إعادة تصدير الـ types من الميزات المختلفة
   - تقليل التكرار وتحسين الاتساق

3. **تحسين State Management**
   - استخدام نظام Zustand موحد في `src/stores/`
   - حذف النظام القديم في `src/store/useStore.ts`
   - تنظيم أفضل مع UI Store و Cart Store

4. **تفعيل Lazy Loading**
   - تطبيق lazy loading على جميع المسارات
   - إضافة loading fallback component
   - تحسين أداء التحميل الأولي

5. **إصلاح Tailwind Config**
   - إزالة التكرار في تعريفات الألوان
   - تنظيم أفضل للـ colors و opacity utilities
   - دعم أفضل للـ dark mode

6. **تحويل Dashboard إلى Tailwind**
   - إزالة inline styles
   - استخدام Tailwind classes بشكل كامل
   - اتساق مع باقي المكونات

7. **تحسين TypeScript Configuration**
   - إضافة path alias `@shared/*`
   - تحديث vite.config.ts للتوافق
   - تحسين دعم path aliases

8. **توثيق شامل**
   - إنشاء `frontend/AGENTS.md` مع شرح كامل للبنية
   - توثيق أفضل للمكونات والـ stores
   - أدلة للمطورين

### التحقق من الميزات المتقدمة

```bash
cd frontend
npm run dev
```

ثم افتح المتصفح وتحقق من:
- اختصارات لوحة المفاتيح (Ctrl+K)
- Service Worker registration
- Print functionality
- Charts rendering
- Performance metrics
- Lazy loading للمسارات
- Shared types consistency
