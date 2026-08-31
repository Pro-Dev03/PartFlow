# خطة الاختبار الشاملة - وضع الأوفلاين

## ✅ الحل المطبق

### 1. Backend Changes (Go)
- ✅ **dashboard/handler.go**: تحسين معالجة الأخطاء
  - GetDashboardStats: إرجاع stats فارغة بدلاً من 500
  - GetLowStockItems: إرجاع قائمة فارغة بدلاً من 500
  - GetOverdueDebts: إرجاع قائمة فارغة بدلاً من 500

### 2. Frontend Changes (TypeScript/React)
- ✅ **DatabaseSettings.tsx**: تحسين معالجة التبديل بين الأوضاع
  - إضافة رسائل توجيهية
  - معالجة أفضل للأخطاء
  - تأخير قبل reload

- ✅ **app.ts**: تحسين API configuration
  - توضيح logic الـ API URL selection
  - إضافة دوال مساعدة
  - معالجة server-side rendering

## 📋 خطوات الاختبار

### المرحلة 1: إعداد البيئة

```bash
# 1. تشغيل البيانات (PostgreSQL)
# ✅ تأكد أن PostgreSQL يعمل على localhost:5432

# 2. تشغيل Backend
cd backend
go run cmd/api/main.go

# 3. تشغيل Frontend (في نافذة منفصلة)
cd frontend
npm run dev
```

### المرحلة 2: اختبار Online Mode (الأساس)

```
1. افتح http://localhost:5173/#/app/login
2. سجل الدخول بحسابك
3. تحقق من:
   - ✅ Dashboard يحمل البيانات
   - ✅ المنتجات تظهر
   - ✅ العملاء يظهرون
   - ✅ أي أخطاء في console؟
```

### المرحلة 3: اختبار Offline Mode Switch

```
1. اذهب إلى Settings (#/app/settings)
2. ابحث عن "Database Settings"
3. انقر على "أوفلاين فقط"
4. تحقق من:
   - ✅ رسالة نجاح = "تم التبديل إلى الأوفلاين"
   - ✅ الصفحة تعيد تحميل (reload)
   - ✅ لا توجد أخطاء في console
   - ✅ localStorage يحتوي على "partflow-operating-mode": "offline"
```

### المرحلة 4: اختبار Offline Mode Pages

بعد التبديل للأوفلاين:

```
1. Dashboard:
   - ✅ تحمل بدون أخطاء
   - ✅ تظهر إحصائيات = 0
   - ✅ لا توجد alerts

2. نقطة البيع (POS):
   - ✅ تحمل بدون أخطاء
   - ✅ قائمة المنتجات = فارغة
   - ✅ لا توجد رسائل خطأ

3. المخزون:
   - ✅ تحمل بدون أخطاء
   - ✅ قائمة المخزون = فارغة
   - ✅ جدول المخزون = فارغ

4. العملاء:
   - ✅ تحمل بدون أخطاء
   - ✅ قائمة العملاء = فارغة

5. الديون:
   - ✅ تحمل بدون أخطاء
   - ✅ قائمة الديون = فارغة

6. المشتريات:
   - ✅ تحمل بدون أخطاء
   - ✅ قائمة المشتريات = فارغة
```

### المرحلة 5: اختبار الحوارات والمدال

```
1. افتح Modal (مثل "إضافة منتج"):
   - ✅ تحمل بدون أخطاء
   - ✅ يمكنك إدخال البيانات
   - ✅ الأزرار تعمل

2. اختبر عمليات CRUD:
   - ✅ إضافة منتج جديد
   - ✅ تحديث المنتج
   - ✅ حذف المنتج
   - ✅ تحقق من البيانات في localStorage
```

### المرحلة 6: اختبار العودة للـ Online Mode

```
1. اذهب إلى Settings
2. انقر على "مزامنة البيانات السحابية"
3. تحقق من:
   - ✅ رسالة نجاح
   - ✅ الصفحة تعيد تحميل
   - ✅ البيانات تعود
   - ✅ localStorage يحتوي على "partflow-operating-mode": "online"
```

## 🔍 نقاط الفحص المهمة

### في Browser Console (F12):

```javascript
// تحقق من الـ mode
localStorage.getItem('partflow-operating-mode')
// يجب أن ترجع "offline" أو "online"

// تحقق من الـ API URL
// افتح Network tab وانظر إلى الطلبات
// يجب أن تبدأ بـ http://localhost:8080 في الأوفلاين
```

### في Network Tab:

```
Offline Mode:
- API requests تذهب إلى: http://localhost:8080/api/v1/...
- ✅ responses = 200 مع بيانات فارغة
- ❌ NO 500 errors

Online Mode:
- API requests تذهب إلى: https://partflow-api.onrender.com/api/v1/...
- ✅ responses = 200 مع البيانات الفعلية
```

## 🐛 الأخطاء الشائعة والحلول

### ❌ الخطأ: "Cannot GET /api/v1/..."
- السبب: Backend لم يبدأ
- الحل: تأكد من تشغيل `go run cmd/api/main.go`

### ❌ الخطأ: "CORS error"
- السبب: Backend CORS configuration
- الحل: تأكد من أن CORS مفعل للـ localhost:5173

### ❌ الخطأ: "503 Service Unavailable"
- السبب: API على Render غير متاح
- الحل: تجاهل عند الاختبار المحلي

### ❌ 500 Error في Dashboard (الأوفلاين)
- السبب: Handler لم يتم تحديثه
- الحل: تأكد من أن dashboard/handler.go محدث

## ✅ قائمة المراجعة النهائية

قبل بناء Windows:

- [ ] ✅ Frontend بناء بنجاح (npm run build)
- [ ] ✅ Backend بناء بنجاح (go build)
- [ ] ✅ Online mode يعمل بدون أخطاء
- [ ] ✅ Offline mode يعمل بدون 500 errors
- [ ] ✅ Dashboard يحمل في الأوفلاين (بدون بيانات)
- [ ] ✅ جميع الصفحات تحمل في الأوفلاين بدون أخطاء
- [ ] ✅ التبديل بين الأوضاع يعمل بسلاسة
- [ ] ✅ لا توجد أخطاء في console
- [ ] ✅ localStorage يتابع التغييرات بشكل صحيح

## 📦 بناء Windows

بعد التأكد من جميع النقاط أعلاه:

```bash
cd frontend
npm run build:windows:full
```

## 📝 ملاحظات مهمة

1. **البيانات الفارغة في الأوفلاين**: هذا متوقع، لأن LocalDB فارغ
2. **عدم وجود مزامنة تلقائية**: يتم العمل عليها، حالياً يدوية
3. **الجلسة الأوفلاين**: يتم حفظها بحيث لا تحتاج لتسجيل الدخول مرة أخرى
4. **لا حذف البيانات**: عند التبديل للأوفلاين، البيانات الأصلية تبقى في السحاب
