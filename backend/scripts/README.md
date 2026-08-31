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
OWNER_PASSWORD=Owner123456
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

- تأكد من أن `DATABASE_URL` معرف في `.env` أو كمتغير بيئة
- السكريبتات تتصل بقاعدة البيانات المحددة في `DATABASE_URL`
- السكريبتات تتعامل مع Schema الموجود وتضيف التوافقيات اللازمة
- كلمات المرور مشفرة دائماً باستخدام bcrypt

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
