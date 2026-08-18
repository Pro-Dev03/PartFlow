# PartFlow - Agent Documentation

## نظرة عامة على المشروع
PartFlow هو نظام إدارة مخزون ومبيعات شامل مصمم للمتاجر الصغيرة والمتوسطة. يوفر المشروع واجهة مستخدم حديثة مع دعم كامل للغة العربية، إدارة المخزون، نقاط البيع، إدارة الديون، والتقارير.

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

## المستقبل

### الميزات المخطط لها
1. تطبيق موبايل (React Native)
2. تكامل مع بوابات الدفع
3. تقارير متقدمة
4. نظام نقاط الولاء
5. تكامل مع منصات التجارة الإلكترونية

## الدعم

للدعم والاستفسارات، يرجى مراجعة:
- وثائق API
- كود المشروع
- فريق التطوير