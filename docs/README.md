# PartFlow - Smart Computer Store Management System

## نظرة عامة

نظام إدارة ذكي لمحلات قطع الحاسوب والإلكترونيات، يدعم المنتجات الجديدة والمستعملة مع نظام متكامل لإدارة المخزون، المبيعات، العملاء، الديون، والموردين.

## التقنيات المستخدمة

### Backend
- **لغة البرمجة**: Go 1.21
- **Architecture**: Modular Monolith
- **قاعدة البيانات**: PostgreSQL via Supabase
- **الاستضافة**: Docker + Render

### Frontend
- **Framework**: React 18 + TypeScript
- **Architecture**: Feature-based
- **Styling**: Tailwind CSS
- **State Management**: Zustand + React Query
- **التدويل**: i18next (Arabic, Hebrew, English)
- **PWA**: Vite PWA Plugin

## بنية المشروع

```
PartFlow/
├── backend/              # Go Backend API
│   ├── cmd/            # Application entry points
│   ├── internal/       # Domain modules
│   ├── pkg/            # Shared packages
│   ├── migrations/     # Database migrations
│   └── tests/          # Backend tests
├── frontend/           # React Frontend
│   ├── src/
│   │   ├── app/       # App configuration
│   │   ├── components/ # Reusable components
│   │   ├── features/  # Feature modules
│   │   ├── layouts/   # Layout components
│   │   ├── hooks/     # Custom hooks
│   │   ├── services/  # API services
│   │   ├── stores/    # State management
│   │   ├── types/     # TypeScript types
│   │   ├── utils/     # Utilities
│   │   ├── i18n/      # Internationalization
│   │   └── styles/    # Global styles
│   └── public/        # Static assets
├── worker/            # Background workers
├── docker/            # Docker configurations
├── docs/              # Documentation
├── scripts/           # Utility scripts
└── docker-compose.yml # Local development
```

## الوثائق المتاحة

- **report.md**: تحليل وتصميم المشروع (المتطلبات والميزات)
- **report-backend.md**: المواصفات التقنية للـBackend
- **frontend.md**: تصميم وتطوير الـFrontend

## البدء بالتطوير

### المتطلبات
- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- Git

### إعداد البيئة المحلية

1. **استنساخ المشروع**
```bash
git clone <repository-url>
cd PartFlow
```

2. **إعداد البيئة**
```bash
cp .env.example .env
# تعديل المتغيرات في ملف .env
```

3. **تشغيل قاعدة البيانات**
```bash
docker-compose up postgres redis -d
```

4. **تشغيل Backend**
```bash
cd backend
go mod download
go run cmd/api/main.go
```

5. **تشغيل Frontend**
```bash
cd frontend
npm install
npm run dev
```

### التشغيل عبر Docker

```bash
docker-compose up
```

## الميزات الرئيسية

### إدارة المنتجات
- دعم المنتجات الجديدة والمستعملة
- نظام Barcode و Serial Number
- تتبع القطع الفردية
- إدارة فئات المنتجات والشركات المصنعة

### إدارة المخزون
- نظام تتبع حركة المخزون
- إدارة المواقع والتخزين
- تنبيهات المخزون المنخفض
- تتبع حالة القطع

### إدارة المبيعات
- نظام سريع للبيع عبر Barcode
- دعم طرق دفع متعددة
- إدارة السلة
- إنشاء الفواتير

### إدارة العملاء
- ملفات العملاء
- نظام الديون
- تتبع المدفوعات
- تاريخ الشراء

### التقارير
- تقارير المبيعات والأرباح
- تقارير المخزون
- تقارير الديون
- تقارير الموردين

## التطوير المستقبلي

- نظام Automation
- تنبيهات ذكية
- Dashboard ذكي
- دعم AI
- تطبيقات Mobile
- Multi-branch support

## الدعم

للمساعدة والدعم، يرجى مراجعة الوثائق في مجلد `docs/` أو التواصل مع فريق التطوير.

## الترخيص

[License information]
