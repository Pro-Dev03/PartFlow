# PartFlow Backend Scripts

هذا المجلد يحتوي على سكريبتات مساعدة لإدارة قاعدة البيانات وإنشاء الحسابات.

## السكريبتات المتاحة

### 1. create_owner.go
سكريبت لإنشاء أو تحديث حساب owner بسهولة.

#### الاستخدام:
```bash
cd backend
go run scripts/create_owner.go
```

#### المتغيرات البيئية (اختياري):
```bash
OWNER_EMAIL=owner@partflow.com
OWNER_PASSWORD=<strong-random-password>
OWNER_FIRST_NAME=Admin
OWNER_LAST_NAME=Owner
OWNER_PHONE=+970599000000
```

#### الميزات:
- ✅ إنشاء أو تحديث حساب owner
- ✅ تشفير كلمة المرور تلقائياً باستخدام bcrypt
- ✅ تفعيل الاشتراك لمدة سنة
- ✅ التوافق مع Schema قاعدة البيانات الموجودة

### 2. seed_data.go
سكريبت لإضافة البيانات الأولية الأساسية للنظام.

#### الاستخدام:
```bash
cd backend
go run scripts/seed_data.go
```

#### البيانات المضافة:
- ✅ Roles افتراضية (Admin, owner, Staff)
- ✅ Categories افتراضية (إلكترونيات، قطع غيار، إكسسوارات، أدوات)
- ✅ Warehouse افتراضي (المخزن الرئيسي)

## قاعدة البيانات

### مدير الاشتراكات عبر Supabase

```powershell
$env:SUPABASE_DATABASE_URL = '<رابط PostgreSQL المباشر من Supabase>'
go run .\scripts\subscription-manager.go summary
go run .\scripts\subscription-manager.go disable --email user@example.com
```

يمكن بدلاً من تعيين المتغير في كل جلسة وضعه في ملف `.env` داخل `backend` أو جذر المشروع؛
سيقرأه مدير الاشتراكات تلقائياً عند تشغيله من `backend\scripts`. متغيرات البيئة الموجودة
في PowerShell لها الأولوية، ولا تُحفظ أي كلمة مرور داخل الكود أو Git.

إذا أغلق Supabase اتصال المنفذ `5432`، استخدم Transaction Pooler على المنفذ `6543`
مع إضافة `?sslmode=require` إلى نهاية الرابط.

يُستخدم `DATABASE_URL_DIRECT` كاسم بديل. لا تضع الرابط أو كلمة المرور داخل
الملفات أو Git.

### Migration Files
- `migrations/001_fix_users_schema.sql`: إصلاح Schema جدول users

### تشغيل Migrations:
```bash
psql $DATABASE_URL -f migrations/001_fix_users_schema.sql
```

## تسلسل العمليات الموصى به

### للإعداد الأولي:
1. تشغيل migration fix
2. تشغيل seed data
3. إنشاء حساب owner

```bash
# 1. Fix schema
psql $DATABASE_URL -f migrations/001_fix_users_schema.sql

# 2. Seed basic data
go run scripts/seed_data.go

# 3. Create owner account
go run scripts/create_owner.go
```

### لإنشاء حساب جديد:
```bash
go run scripts/create_owner.go
```

## API Registration

بعد تحسينات API، يمكنك أيضاً إنشاء حسابات عبر API:

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "Password123",
    "first_name": "John",
    "last_name": "Doe",
    "phone": "+970599000000"
  }'
```

## ملاحظات هامة

- عرّف `SUPABASE_DATABASE_URL` لرابط PostgreSQL المباشر من Supabase.
- يمكن استخدام `DATABASE_URL_DIRECT` كاسم بديل؛ لا يستخدم مدير الاشتراكات قاعدة محلية.
- مدير الاشتراكات يتصل مباشرة بعنوان Supabase المحدد في `SUPABASE_DATABASE_URL` أو `DATABASE_URL_DIRECT`
- السكريبتات تتعامل مع Schema الموجود وتضيف التوافقيات اللازمة
- كلمات المرور مشفرة دائماً باستخدام bcrypt
- لا يُنشأ Owner محلي افتراضي في نسخة Desktop؛ استخدم `PARTFLOW_BOOTSTRAP_OWNER_PASSWORD` للاختبارات أو الإعداد المحلي المقصود فقط.

## استكشاف الأخطاء

### مشاكل الاتصال بقاعدة البيانات:
```bash
# تحقق من DATABASE_URL
echo $DATABASE_URL

# جرب الاتصال المباشر
psql $DATABASE_URL
```

### مشاكل Schema:
```bash
# شغّل migration fix
psql $DATABASE_URL -f migrations/001_fix_users_schema.sql
```

### مشاكل البيانات المكررة:
السكريبتات تتعامل مع البيانات المكررة تلقائياً:
- `create_owner.go`: تحديث الحساب الموجود بدلاً من إنشاء جديد
- `seed_data.go`: يتخطى البيانات الموجودة

## الأرشيف

تم نقل سكريبتات الإصلاحات المؤقتة، واختبارات التشخيص، وعمليات الترحيل الخاصة
بمشاكل سابقة إلى `backend/scripts/archive/2026-08`. لا تُشغّل أي سكريبت من
الأرشيف على بيئة الإنتاج إلا بعد مراجعة محتواه وهدفه.
