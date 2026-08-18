# دليل النشر على Render - PartFlow

## نظرة عامة

هذا الدليل يشرح كيفية نشر مشروع PartFlow على منصة Render السحابية.

## المتطلبات

- حساب على Render (https://render.com)
- حساب GitHub مع الكود المصدري للمشروع
- Git مثبت على جهازك

## خطوات النشر

### 1. إعداد المستودع على GitHub

```bash
# إضافة ملف render.yaml إلى المستودع
git add render.yaml
git commit -m "Add Render configuration"
git push origin main
```

### 2. إنشاء حساب Render

1. اذهب إلى https://render.com
2. سجل حساب جديد أو سجل الدخول
3. قم بربط حساب GitHub الخاص بك

### 3. نشر المشروع باستخدام render.yaml

#### الطريقة الأسرع (Blueprint)

1. في لوحة تحكم Render، اضغط على "New +"
2. اختر "Blueprint"
3. اختر المستودع الذي يحتوي على مشروع PartFlow
4. سيكتشف Render تلقائياً ملف `render.yaml`
5. راجع الإعدادات واضغط "Apply"

#### الطريقة اليدوية

##### 3.1 إنشاء قاعدة البيانات

1. اضغط "New +"
2. اختر "PostgreSQL"
3. أدخل الإعدادات:
   - Name: `partflow-db`
   - Database: `partflow`
   - Version: `15`
   - Plan: Free
4. اضغط "Create Database"

##### 3.2 إنشاء Redis

1. اضغط "New +"
2. اختر "Redis"
3. أدخل الإعدادات:
   - Name: `partflow-redis`
   - Plan: Free
4. اضغط "Create Redis"

##### 3.3 نشر Backend

1. اضغط "New +"
2. اختر "Web Service"
3. أدخل الإعدادات:
   - Name: `partflow-backend`
   - Runtime: Docker
   - Dockerfile Path: `./backend/Dockerfile`
   - Docker Context: `./backend`
   - Region: Oregon (أو الأقرب لك)
   - Branch: `main`
4. أضف Environment Variables:
   - `DATABASE_URL`: (اتصل بقاعدة البيانات)
   - `REDIS_URL`: (اتصل بـ Redis)
   - `JWT_SECRET`: (قم بتوليد قيمة عشوائية)
   - `SUPABASE_URL`: (إذا كنت تستخدم Supabase)
   - `SUPABASE_KEY`: (إذا كنت تستخدم Supabase)
   - `APP_ENV`: `production`
   - `APP_PORT`: `8080`
5. اضغط "Create Web Service"

##### 3.4 نشر Frontend

1. اضغط "New +"
2. اختر "Web Service"
3. أدخل الإعدادات:
   - Name: `partflow-frontend`
   - Runtime: Docker
   - Dockerfile Path: `./frontend/Dockerfile`
   - Docker Context: `./frontend`
   - Region: نفس الـ Backend
   - Branch: `main`
4. أضف Environment Variables:
   - `VITE_API_URL`: `https://partflow-backend.onrender.com`
   - `NODE_ENV`: `production`
5. اضغط "Create Web Service"

##### 3.5 نشر Worker

1. اضغط "New +"
2. اختر "Worker"
3. أدخل الإعدادات:
   - Name: `partflow-worker`
   - Runtime: Docker
   - Dockerfile Path: `./worker/Dockerfile`
   - Docker Context: `./worker`
   - Region: نفس الـ Backend
   - Branch: `main`
4. أضف Environment Variables:
   - `DATABASE_URL`: (نفس Backend)
   - `REDIS_URL`: (نفس Backend)
   - `JWT_SECRET`: (نفس Backend)
5. اضغط "Create Worker"

### 4. تشغيل Database Migrations

بعد نشر Backend، قم بتشغيل migrations:

```bash
# عبر Render Shell
# 1. افتح خدمة Backend في Render
# 2. اضغط على "SSH"
# 3. شغل الأمر:
go run cmd/migrate/main.go up
```

أو أضف ملف `scripts/migrate.sh` وشغله تلقائياً في Dockerfile.

### 5. التحقق من النشر

1. افتح URL الخاص بالـ Frontend: `https://partflow-frontend.onrender.com`
2. تحقق من أن التطبيق يعمل بشكل صحيح
3. راجع الـ Logs في لوحة تحكم Render

## إدارة Environment Variables

### في Render Dashboard

1. افتح الخدمة المطلوبة
2. اضغط على "Environment"
3. أضف أو عدل المتغيرات
4. اضغط "Save Changes"
5. سيتم إعادة تشغيل الخدمة تلقائياً

### المتغيرات المهمة

```bash
# Backend
DATABASE_URL=postgresql://...
REDIS_URL=redis://...
JWT_SECRET=your-secret-key
SUPABASE_URL=your-supabase-url
SUPABASE_KEY=your-supabase-key
APP_ENV=production
APP_PORT=8080

# Frontend
VITE_API_URL=https://partflow-backend.onrender.com
NODE_ENV=production
```

## مراقبة الأداء

### Logs

- Dashboard > Service > Logs
- يمكنك تصفية Logs حسب الوقت أو المستوى

### Metrics

- Dashboard > Service > Metrics
- راقب CPU, Memory, Response Time

### Health Checks

- Backend: `/health`
- Frontend: `/`
- يقوم Render بالتحقق تلقائياً

## تحديث التطبيق

### Auto-deploy

Render يقوم بالنشر التلقائي عند:
- Push إلى branch الرئيسي
- Merge Pull Request

### Manual Deploy

1. Dashboard > Service > Manual Deploy
2. اختر branch
3. اضغط "Deploy"

## استكشاف الأخطاء

### الخدمة لا تعمل

1. راجع Logs
2. تحقق من Environment Variables
3. تأكد من Health Check
4. تحقق من الاتصال بقاعدة البيانات

### أخطاء البناء

1. راجع Build Logs
2. تحقق من Dockerfile
3. تأكد من وجود جميع الملفات المطلوبة

### مشاكل قاعدة البيانات

1. تحقق من DATABASE_URL
2. تأكد من تشغيل Migrations
3. راجع Logs لمعرفة الأخطاء

## التكلفة

### Free Tier

- PostgreSQL: 90 يوم تجريبي
- Redis: Free limited
- Web Services: 750 ساعة/شهر
- Worker: Limited

### Starter Plan

- مناسب للمشاريع الصغيرة
- starts at $7/month
- أداء أفضل

## النسخ الاحتياطي

### Database Backups

Render يقوم بـ:
- Daily backups تلقائي
- 7 أيام retention

### Manual Backup

```bash
# Export database
pg_dump $DATABASE_URL > backup.sql

# Import database
psql $DATABASE_URL < backup.sql
```

## الأمان

### Best Practices

1. استخدم Strong Secrets
2. فعّل SSL/TLS (مفعّل تلقائياً)
3. راجع Environment Variables بانتظام
4. استخدم Private Services للخدمات الداخلية

### IP Allowlist

1. Dashboard > Database > Settings
2. أضيف IPs المسموح بها
3. أو استخدم Private Traffic

## دعم Multi-tenant

PartFlow يدعم Multi-tenant من البداية:

1. كل business مرتبط بـ `organization_id`
2. استخدم Row Level Security في قاعدة البيانات
3. عزل البيانات على مستوى التطبيق

## الموارد

- [Render Documentation](https://render.com/docs)
- [Docker on Render](https://render.com/docs/docker)
- [PostgreSQL on Render](https://render.com/docs/postgresql)
- [Redis on Render](https://render.com/docs/redis)

## الدعم

إذا واجهت مشاكل:
1. راجع Render Status Page
2. تحقق من Logs
3. راجع هذا الدليل
4. تواصل مع فريق الدعم
