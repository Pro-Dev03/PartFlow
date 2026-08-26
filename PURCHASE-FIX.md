# إصلاح مشكلة إنشاء المشتريات (Purchase Creation Fix)

## المشكلة الأصلية
عند محاولة إنشاء شراء جديد، كان النظام يرجع خطأ:
```
HTTP 400 Bad Request
"failed to create purchase: failed to create purchase: pq: INSERT has more expressions than target columns"
```

## التشخيص

### 1. عدم تطابق Schema بين Database و Go Code

**Database Schema (purchases table):**
- `user_id` (NOT NULL) - المستخدم الذي أنشأ الشراء
- `payment_method` - طريقة الدفع
- `payment_status` - حالة الدفع
- `status` (default 'completed')

**Go Code Expects:**
- `created_by` (nullable) - المستخدم الذي أنشأ الشراء
- لا يتوقع `payment_method` أو `payment_status`
- `status` (pending, received, cancelled)

### 2. خطأ في SQL INSERT

في `repository.go` كان هناك 12 placeholders ($1-$12) لكن فقط 11 حقول في INSERT statement.

## الإصلاحات المطبقة

### Backend Changes

#### 1. إصلاح Model (`internal/purchases/model.go`)
```go
// قبل
type Purchase struct {
    ...
    CreatedBy      uuid.UUID  `json:"created_by" db:"created_by"`
    ...
}

// بعد
type Purchase struct {
    ...
    UserID         uuid.UUID  `json:"user_id" db:"user_id"` // Maps to created_by in API
    ...
}
```

#### 2. إصلاح Repository (`internal/purchases/repository.go`)
- تحديث SQL INSERT ليستخدم `user_id` بدلاً من `created_by`
- إصلاح عدد placeholders من 12 إلى 11
- تحديث جميع SELECT queries لاستخدام `user_id`

#### 3. إصلاح Service (`internal/purchases/service.go`)
- تحديث إنشاء Purchase لاستخدام `UserID` بدلاً من `CreatedBy`
- تحديث جميع ledger queries لاستخدام `user_id` بدلاً من `created_by`
- تحديث payment queries لاستخدام `user_id` بدلاً من `created_by`

#### 4. إضافة API Compatibility (`internal/purchases/dto.go`)
```go
// ToAPIMap converts Purchase to a map with created_by for API compatibility
func (p *Purchase) ToAPIMap() map[string]interface{} {
    return map[string]interface{}{
        ...
        "created_by": p.UserID, // Map user_id to created_by for API compatibility
        ...
    }
}
```

### Frontend Changes

#### 1. إصلاح البيانات المرسلة (`features/purchases/pages/CreatePurchasePage.tsx`)
- إزالة `expected_date` (غير موجود في Backend)
- تحويل الأنواع بشكل صريح:
  - `quantity`: `parseInt(item.quantity.toString()) || 0`
  - `unit_cost`: `parseFloat(item.unit_cost.toString()) || 0`
- إضافة validation شامل قبل الإرسال

#### 2. تحسين معالجة الأخطاء (`features/purchases/hooks/usePurchases.ts`)
- إضافة logging مفصل للأخطاء
- عرض رسائل خطأ واضحة للمستخدم
- طباعة error status, response, و code

## النتائج

✅ **البناء نجح** بدون أخطاء
✅ **تم إصلاح عدم تطابق Schema** بين Database و Go Code
✅ **تم إصلاح SQL INSERT error**
✅ **تم تحسين معالجة الأخطاء** في Frontend
✅ **تم إضافة validation** للبيانات قبل الإرسال

## الملفات المُحدّثة

### Backend:
- `backend/internal/purchases/model.go`
- `backend/internal/purchases/repository.go`
- `backend/internal/purchases/service.go`
- `backend/internal/purchases/dto.go`

### Frontend:
- `frontend/src/features/purchases/pages/CreatePurchasePage.tsx`
- `frontend/src/features/purchases/hooks/usePurchases.ts`

## الاختبار

الآن يمكن اختبار إنشاء شراء جديد:
1. أدخل صفحة `/app/purchases/create`
2. اختر مورد
3. أضف عناصر للشراء
4. اضغط حفظ

يجب أن يعمل بشكل صحيح بدون أخطاء.
