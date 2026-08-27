# PartFlow - Agent Documentation

## نظرة عامة على المشروع
PartFlow هو نظام إدارة مخزون ومبيعات شامل مصمم للمتاجر الصغيرة والمتوسطة. يوفر المشروع واجهة مستخدم حديثة مع دعم كامل للغة العربية، إدارة المخزون، نقاط البيع، إدارة الديون، والتقارير.

## الفلسفة الأساسية (PRODUCT-PHILOSOPHY.md)
القاعدة الذهبية التي تحكم PartFlow:

> **النظام يعمل من أجل صاحب المتجر، وليس صاحب المتجر يعمل من أجل النظام.**

المبادئ الأساسية:
- **التعقيد يجب أن يكون داخل النظام**: Backend معقد لكن الواجهة بسيطة
- **لا تسأل المستخدم ما يعرفه النظام**: النظام يحسب تلقائياً
- **لغة صاحب المتجر**: استخدم مصطلحات العمل اليومية بدلاً من المصطلحات التقنية
- **خطوة واحدة بدل ثلاث**: النظام ينفذ العمليات المرتبطة تلقائياً
- **الأخطاء ليست مسؤولية المستخدم**: النظام يساعد على التصحيح

## البنية المعمارية المحسّنة
تم تحسين البنية المعمارية بناءً على مبادئ ناجحة من مشروع Fynexa، مع الحفاظ على هوية PartFlow الخاصة بالمتاجر.

### المبادئ المعمارية المطبقة
1. **Centralized Routing Pattern**: ملف router مركزي لإدارة جميع المسارات
2. **Integration Management System**: إطار عمل منظم لإدارة التكاملات الخارجية
3. **Service-Based Architecture**: فصل واضح بين business logic والتكاملات
4. **Current State + Immutable History**: فصل الحالة الحالية عن السجل التاريخي (ARCHITECTURE-PRINCIPLES.md)
5. **Reverse instead of Delete**: عكس العمليات بدلاً من الحذف للحفاظ على التاريخ (ARCHITECTURE-PRINCIPLES.md)
6. **Aggregation Tables**: جداول تجميع لتسريع Dashboard (ARCHITECTURE-PRINCIPLES.md)

### الملفات المعمارية الجديدة
- `backend/internal/api/router.go`: المسار المركزي (جاهز للتفعيل)
- `backend/internal/integrations/`: نظام إدارة التكاملات الخارجية
- `ARCHITECTURE_ANALYSIS.md`: تحليل شامل للتحسينات المعمارية
- `ARCHITECTURE-PRINCIPLES.md`: مبادئ التصميم المعماري الجديدة (2026-08-26)
- `backend/migrations/002_architecture_principles.sql`: ترحيل قاعدة البيانات للمبادئ الجديدة

## المميزات الرئيسية

### 1. نظام إدارة المخزون
- تتبع المخزون في الوقت الفعلي
- تنبيهات المخزون المنخفض
- دعم المنتجات المتعددة
- إدارة حركات المخزون

### 2. نظام نقاط البيع (POS)
- مسح الباركود
- إدارة المبيعات
- دعم الديون والدفعات
- واجهة سريعة وسهلة الاستخدام

### 3. إدارة الديون
- تتبع ديون العملاء
- نظام تقادم الديون (aging)
- تنبيهات الديون المتأخرة
- تسجيل الدفعات

### 4. التقارير والتحليلات
- تقارير المبيعات
- تقارير الأرباح
- تقارير المخزون
- تقارير الديون

### 5. الإشعارات
- إشعارات النظام
- إشعارات المتصفح
- تنبيهات فورية
- إدارة تفضيلات الإشعارات

## البنية التقنية

### الواجهة الأمامية (Frontend)
- **Framework**: React 18 مع TypeScript
- **Routing**: React Router
- **State Management**: Zustand
- **Data Fetching**: TanStack Query
- **Styling**: Tailwind CSS
- **UI Components**: مكونات مخصصة
- **Internationalization**: i18next
- **Build Tool**: Vite

### الواجهة الخلفية (Backend)
- **Language**: Go
- **Framework**: Gin
- **Database**: PostgreSQL
- **Architecture**: بنية قائمة على الخدمات (Service-based)

### العمليات الخلفية (Worker)
- **Language**: Go
- **Tasks**: 
  - فحص الديون المتأخرة
  - تنبيهات المخزون المنخفض
  - فحص الضمانات المنتهية
  - توليد الرؤى اليومية

## تعليمات البناء والتشغيل

### البناء والتشغيل (Frontend)
```bash
cd frontend
npm install
npm run dev       # للتطوير
npm run build     # للإنتاج
npm run preview   # لمعاينة الإنتاج
```

### البناء والتشغيل (Backend)
```bash
cd backend
go mod download
go run cmd/api/main.go
```

### البناء والتشغيل (Worker)
```bash
cd worker
go mod download
go run main.go
```

## البيئة المطلوبة

### المتطلبات الأساسية
- Node.js 18+
- Go 1.21+
- PostgreSQL 14+
- نظام تشغيل يدعم Docker (اختياري)

### متغيرات البيئة
```env
# Frontend
VITE_API_BASE_URL=http://localhost:8080/api/v1

# Backend
DB_HOST=localhost
DB_PORT=5432
DB_NAME=partflow
DB_USER=postgres
DB_PASSWORD=your_password
JWT_SECRET=your_jwt_secret
```

## الميزات المحسنة

### 1. تحسينات الأداء
- **Lazy Loading**: تحميل بطيء للصفحات لتقليل حجم التطبيق الأولي
- **API Caching**: تخزين مؤقت لطلبات API لتقليل الاستهلاك
- **Code Splitting**: تقسيم الكود إلى chunks لتحسين التحميل
- **Service Worker**: دعم PWA للعمل بدون إنترنت

### 2. تحسينات الإشعارات
- **Browser Notifications**: إشعارات المتصفح الأصلية
- **Sound Alerts**: تنبيهات صوتية للإشعارات الجديدة
- **Real-time Updates**: تحديثات فورية للإشعارات
- **Smart Filtering**: تصفية ذكية للإشعارات حسب النوع

### 3. تحسينات معالجة الأخطاء
- **Arabic Error Messages**: رسائل خطأ مترجمة للعربية
- **Retry Logic**: إعادة المحاولة التلقائية للطلبات الفاشلة
- **Error Boundaries**: حدود أخطاء React لمنع تعطل التطبيق
- **User-friendly Errors**: رسائل خطأ سهلة الفهم

### 4. تحسينات الديون
- **Debt Worker Integration**: تكامل مع Debt Worker للتنبيهات التلقائية
- **Aging Display**: عرض واضح لتقادم الديون
- **Quick Actions**: إجراءات سريعة لتسجيل الدفعات
- **Dashboard Alerts**: تنبيهات على لوحة التحكم

### 5. تحسينات جديدة بناءً على التقارير (2026-08-20)
بناءً على تحليل التقارير في مجلد `docs/`، تم تطبيق تحسينات لتتوافق مع الهدف الأساسي: **"النظام يعمل لصالح صاحب المحل"**

#### أ. حل مشكلة الوضع الفاتح
- **المشكلة**: خلفية تظهر خلف الخلفية عند تفعيل الوضع الفاتح
- **الحل**: تحديث `themes.css` باستخدام `!important` لضمان تطبيق الألوان الصحيحة
- **الملف**: `frontend/src/styles/themes.css`

#### ب. تحويل Dashboard إلى "ماذا يحدث الآن؟"
- **الهدف**: تحويل Dashboard من مجرد charts إلى نظام يخبر صاحب المحل بما يحدث الآن
- **التحسينات**:
  - إضافة قسم "يحتاج انتباهك" - أهم قسم في Dashboard
  - عرض منتجات منخفضة المخزون مع أزرار مباشرة
  - عرض ديون متأخرة مع تفاصيل
  - عرض ضمانات تنتهي قريباً
  - إضافة Smart Actions (العمليات اليومية الأكثر استخداماً)
- **الملف**: `frontend/src/features/dashboard/pages/DashboardPage.tsx`

#### ج. Customer Financial Timeline
- **الهدف**: تتبع السجل المالي لكل عميل بشكل واضح
- **التحسينات**:
  - إنشاء مكون `FinancialTimeline` جديد
  - عرض جميع الحركات المالية (بيع، دفع، مرتجع، استرجاع، تعديل)
  - عرض الرصيد بعد كل حركة
  - واجهة بصرية واضحة مع ألوان رمزية
- **الملفات**:
  - `frontend/src/components/ui/financial-timeline.tsx` (جديد)
  - `frontend/src/features/customers/pages/CustomersPage.tsx` (مُحدّث)

#### د. Debt Aging System
- **الهدف**: تصنيف ديون حسب العمر لتحديد الأولويات
- **التحسينات**:
  - إضافة `getDebtAging` function لتصنيف الديون
  - التصنيفات: PAID, DUE_SOON, CURRENT, OVERDUE_1_7, OVERDUE_8_14, OVERDUE_15_30, OVERDUE_30_PLUS
  - عرض عدد الأيام المتأخرة
  - تحديث إحصائيات الديون بناءً على التصنيف
- **الملف**: `frontend/src/features/debts/pages/DebtsPage.tsx`

#### هـ. Inventory Movements Ledger
- **الهدف**: تتبع حركات المخزون بدلاً من مجرد عرض الرقم الحالي
- **التحسينات**:
  - إنشاء مكون `InventoryLedger` جديد
  - عرض جميع الحركات (شراء، بيع، مرتجع، تعديل، نقل، تالف، إصلاح، حجز، إلغاء حجز)
  - عرض الكمية قبل وبعد كل حركة
  - واجهة بصرية واضحة مع ألوان رمزية
- **الملفات**:
  - `frontend/src/components/ui/inventory-ledger.tsx` (جديد)
  - `frontend/src/features/inventory/pages/InventoryPage.tsx` (للتحديث)

#### و. Financial Immutability
- **الهدف**: منع حذف السجلات المالية (دفعات، مبيعات، إلخ)
- **التحسينات**:
  - استبدال `handleDeleteDebt` بـ `handleReversePayment`
  - عند عكس الدفعة، يتم إنشاء سجل عكس بدلاً من الحذف
  - رسالة توضيحية للمستخدم
- **الملف**: `frontend/src/features/debts/pages/DebtsPage.tsx`

#### ز. Smart Actions في Dashboard
- **الهدف**: توفير وصول سريع للعمليات اليومية
- **التحسينات**:
  - إضافة أزرار كبيرة وواضحة في Dashboard
  - العمليات: بيع جديد، إضافة قطعة، إضافة عميل، تسجيل دفعة، إضافة مصروف
  - تصميم يسهل النقر السريع
- **الملف**: `frontend/src/features/dashboard/pages/DashboardPage.tsx`

#### ح. Modal Navigation Enhancements (2026-08-26)
- **الهدف**: تحسين التنقل في النوافذ المنبثقة (Modals) باستخدام لوحة المفاتيح
- **التحسينات**:
  - **Auto Focus**: التركيز التلقائي على أول حقل إدخال عند فتح النافذة
  - **Enter Navigation**: الضغط على Enter للانتقال للحقل التالي تلقائياً
  - **Enhanced TAB Navigation**: تحسين التنقل بـ TAB و Shift+TAB
  - **Modern Design Update**: تحديث تصميم variant "modern" ليطابق المعايير الحديثة
  - **Configurable Props**: إضافة `autoFocus` و `enableEnterNavigation` للتحكم بالميزات
- **الملفات**:
  - `frontend/src/components/ui/modal.tsx` (مُحدّث)
  - `frontend/src/features/inventory/components/InventoryModals.tsx` (مُحدّث)
  - `frontend/MODAL-ENHANCEMENTS.md` (جديد - التوثيق)
- **اختصارات لوحة المفاتيح**:
  - `ESC`: إغلاق النافذة
  - `TAB`: الانتقال للحقل التالي
  - `Shift + TAB`: الرجوع للحقل السابق
  - `Enter`: الانتقال للحقل التالي (في حقول الإدخال)

## اختبار المشروع

### اختبار البناء
```bash
cd frontend
npm run build
```

### اختبار التطوير
```bash
cd frontend
npm run dev
```

### اختبار PWA
```bash
cd frontend
npm run build
npm run preview
```

## المشاكل المعروفة والحلول

### 1. مشاكل TypeScript
- **المشكلة**: أخطاء TypeScript في Service Worker
- **الحل**: تم إضافة تعريفات النوع المخصصة

### 2. مشاكل Build
- **المشكلة**: أخطاء في vite.config.ts
- **الحل**: تم تحديث manualChunks لتكون دالة

### 3. مشاكل API
- **المشكلة**: أخطاء في الاتصال بالـ API
- **الحل**: تم إضافة retry logic و caching

## التحسينات المعمارية الحديثة

### 1. Frontend Design System (النهج الجديد)
تم تبني نهج جديد كلياً بناءً على فحص معماري دقيق:
- **الاستراتيجية**: صقل الموجود بدلاً من إعادة البناء
- **التشخيص**: المشكلة ليست في Architecture، بل في Visual System Consistency
- **الملفات الجديدة**:
  - `frontend/FRONTEND-DESIGN-SYSTEM.md`: القانون البصري للمشروع
  - `frontend/COMPONENT-AUDIT.md`: تقرير فحص المكونات
  - `frontend/src/components/tables/data-table.tsx`: Business Table System

### 2. Component Audit Results
- **تقييم المكونات الموجودة**: 8.5/10
- **المكونات الممتازة**: Button, Badge, Page Header, Card, Input, Table
- **المكونات المحدثة**: Modal (للتوافق مع Design System)
- **المكونات الجديدة**: Business Table System

### 3. Centralized Routing System
- **الملف**: `backend/internal/api/router.go`
- **الوضع الحالي**: جاهز للتفعيل (مُعطل مؤقتاً بسبب أخطاء في بعض الموديولات)
- **الفوائد**:
  - سهولة الصيانة
  - وضوح البنية
  - تقليل تكرار الكود
  - سهولة التتبع
- **الخطوات التالية**: إصلاح الأخطاء في الموديولات الحالية ثم تفعيل المسار المركزي

### 4. Integration Management System
- **المجلد**: `backend/internal/integrations/`
- **الوضع الحالي**: مُطبق وجاهز للاستخدام
- **الفوائد**:
  - إدارة مركزية لجميع التكاملات الخارجية
  - واجهة موحدة لجميع الخدمات
  - سهولة إضافة تكاملات جديدة
  - فصل واضح بين business logic والتكاملات
- **الأنواع المدعومة**:
  - Payment gateways
  - Notification services
  - Storage services
  - Messaging services
  - Analytics services
  - Shipping services

### 5. التوثيق المعماري
- **الملف**: `ARCHITECTURE_ANALYSIS.md`
- **المحتوى**: تحليل شامل للمبادئ المعمارية من Fynexa وكيفية تطبيقها على PartFlow
- **يشمل**:
  - مقارنة بين البنيتين
  - فجوات النضج المعماري
  - خطة التطبيق

### 6. Responsive Design Mobile Improvements (2026-08-20)
- **الهدف**: تحسين تجربة المستخدم على الأجهزة المحمولة
- **التحسينات المطبقة**:
  - **Mobile Sidebar**: إضافة drawer متحرك للموبايل مع backdrop
  - **Mobile Header**: إضافة زر القائمة للموبايل وإزالة MobileMenu القديم
  - **Mobile CSS**: إنشاء `mobile.css` مع تحسينات للموبايل
  - **Touch Targets**: زيادة حجم عناصر اللمس إلى 44px
  - **Responsive Grids**: تحسين الشبكات في جميع الصفحات الرئيسية
  - **Typography**: تحسين الخطوط للموبايل
  - **Landscape Mode**: تحسينات للوضع الأفقي
- **الملفات المحدثة**:
  - `frontend/src/components/navigation/sidebar.tsx`
  - `frontend/src/components/navigation/header.tsx`
  - `frontend/src/layouts/AppLayout/AppLayout.tsx`
  - `frontend/src/styles/mobile.css` (جديد)
  - `frontend/src/index.css`
  - `frontend/src/features/dashboard/pages/DashboardPage.tsx`
  - `frontend/src/features/customers/pages/CustomersPage.tsx`
  - `frontend/src/features/inventory/pages/InventoryPage.tsx`
  - `frontend/src/features/sales/pages/POSPage.tsx`

## الميزات الجديدة (2026-08-27)

### 1. Held Sales System
- **الهدف**: حفظ عملية البيع مؤقتاً واستكمالها لاحقاً
- **التحسينات**:
  - إضافة functionality لحفظ البيع مؤقتاً
  - عرض قائمة المبيعات المحفوظة
  - استعادة وحذف المبيعات المحفوظة
  - دعم إدارة المبيعات المحفوظة في POS
- **الملفات**:
  - `backend/internal/sales/held_sales.go` (جديد)
  - `backend/migrations/038_held_sales.sql` (جديد)
  - تحديثات في مكونات POS

### 2. E2E Testing with Playwright
- **الهدف**: اختبار المسارات الحرجة للمستخدم
- **التحسينات**:
  - إعداد Playwright للاختبار
  - اختبارات critical-path.spec.ts
  - اختبارات dashboard.spec.ts
  - اختبارات مبيعات الكاش المباشرة
- **الملفات**:
  - `frontend/e2e/critical-path.spec.ts` (جديد)
  - `frontend/e2e/dashboard.spec.ts` (جديد)
  - `frontend/playwright.config.ts` (جديد)
  - `frontend/pos-cash-sale-test.mjs` (جديد)

### 3. Purchase Details Page
- **الهدف**: عرض تفاصيل المشتريات بشكل واضح
- **التحسينات**:
  - صفحة جديدة لعرض تفاصيل المشتريات
  - عرض معلومات البائع والمشتريات
  - واجهة واضحة وسهلة الاستخدام
- **الملفات**:
  - `frontend/src/features/purchases/pages/PurchaseDetailsPage.tsx` (جديد)

### 4. Used Parts Stock Page
- **الهدف**: إدارة مخزون القطع المستعملة
- **التحسينات**:
  - صفحة جديدة لإدارة القطع المستعملة
  - عرض وتتبع القطع المستعملة
  - واجهة متكاملة لإدارة المخزون
- **الملفات**:
  - `frontend/src/features/usedparts/pages/UsedPartsStockPage.tsx` (جديد)

### 5. Performance Optimizations
- **الهدف**: تحسين أداء قاعدة البيانات والاستعلامات
- **التحسينات**:
  - إضافة فهرسة لتحسين الاستعلامات
  - تحسينات في migrations
  - تحسينات في API client
- **الملفات**:
  - `backend/migrations/037_performance_indexes.sql` (جديد)
  - تحديثات في `frontend/src/services/api/client.ts`

### 6. New UI Components
- **الهدف**: توسيع مكتبة المكونات
- **التحسينات**:
  - confirm-dialog component
  - form-group component
  - loading-state component
- **الملفات**:
  - `frontend/src/components/ui/confirm-dialog.tsx` (جديد)
  - `frontend/src/components/ui/form-group.tsx` (جديد)
  - `frontend/src/components/ui/loading-state.tsx` (جديد)

### 7. Custom Hooks
- **الهدف**: إعادة استخدام منطق مشترك
- **التحسينات**:
  - useDebounce hook للتحسين في البحث
  - useIsMobile hook للتحقق من حجم الشاشة
- **الملفات**:
  - `frontend/src/hooks/useDebounce.ts` (جديد)
  - `frontend/src/hooks/useIsMobile.ts` (جديد)

## اختبار المشروع

### اختبار البناء
```bash
cd frontend
npm run build
```

### اختبار التطوير
```bash
cd frontend
npm run dev
```

### اختبار E2E
```bash
cd frontend
npx playwright test
```

### اختبار PWA
```bash
cd frontend
npm run build
npm run preview
```

## البيئة المطلوبة

### المتطلبات الأساسية
- Node.js 18+
- Go 1.21+
- PostgreSQL 14+
- نظام تشغيل يدعم Docker (اختياري)
- Playwright للاختبار (اختياري)

### متغيرات البيئة
```env
# Frontend
VITE_API_BASE_URL=http://localhost:8080/api/v1

# Backend
DB_HOST=localhost
DB_PORT=5432
DB_NAME=partflow
DB_USER=postgres
DB_PASSWORD=your_password
JWT_SECRET=your_jwt_secret
```

## التوثيق الإضافي

### ملفات التوثيق المتاحة
- `PRODUCT-PHILOSOPHY.md`: الفلسفة الأساسية للمشروع
- `ARCHITECTURE_ANALYSIS.md`: تحليل معماري شامل
- `ARCHITECTURE-PRINCIPLES.md`: مبادئ التصميم المعماري
- `FRONTEND-DESIGN-SYSTEM.md`: القانون البصري للواجهة الأمامية
- `COMPONENT-AUDIT.md`: تقرير فحص المكونات
- `MODAL-ENHANCEMENTS.md`: تحسينات النوافذ المنبثقة
- `IMPROVEMENTS-SUMMARY.md`: ملخص التحسينات
- `FRONTEND-IMPROVEMENTS-REPORT.md`: تقرير تحسينات الواجهة الأمامية
- `FRONTEND-OPTIMIZATION-REPORT.md`: تقرير تحسينات الأداء
- `FRONTEND-QA-REPORT.md`: تقرير الجودة
- `PERFORMANCE-TEST-REPORT.md`: تقرير اختبار الأداء
- `USER-TESTING-LOG.md`: سجل اختبار المستخدم

## السياسات والإرشادات

### التطوير
- اتبع مبادئ الفلسفة الأساسية دائماً
- استخدم المكونات الموجودة قدر الإمكان
- اتبع معايير التصميم المعمارية
- احتفظ بالتوثيق محدثاً

### الاختبار
- اختبار المسارات الحرجة دائماً
- التحقق من الأداء قبل الرفع
- التأكد من التوافق مع الموبايل
- اختبار الإشعارات والتنبيهات

### الرفع
- إنشاء commits واضحة ومفصلة
- عدم رفع الأسرار والمفاتيح
- مراجعة التغييرات قبل الرفع
- التأكد من عدم كسر الوظائف الموجودة