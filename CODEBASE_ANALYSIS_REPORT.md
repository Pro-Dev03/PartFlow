# تقرير تحليل أكواد PartFlow الشامل
**التاريخ:** 31 أغسطس 2026
**النطاق:** بحث شامل عن الأخطاء والتناقضات والمشاكل المحتملة

---

## الملخص التنفيذي

تم اكتشاف **42+ مشكلة محتملة** عبر الأكواد تشمل تناقضات في قاعدة البيانات، عدم توافق نقاط نهاية الـ API، فجوات معالجة الأخطاء، مشاكل تدفق البيانات، والتحقق المفقود. تم تصنيف المشاكل حسب الخطورة.

---

## 1. 🔴 المشاكل الحرجة

### 1.1 API التجميعات - بيانات مزيفة موضحة مسبقاً
**الموقع:** [backend/internal/api/aggregations.go](backend/internal/api/aggregations.go#L119-L270)

**المشكلة:** جميع 8 نقاط نهاية للتجميع ترجع بيانات مزيفة موضحة مسبقاً بدلاً من الاستعلام عن جداول قاعدة البيانات الفعلية.

**نقاط النهاية المتأثرة:**
- `GET /aggregations/daily-sales` - ترجع 500 عنصر ثابت
- `GET /aggregations/monthly-sales` - ترجع بيانات 2026-08 ثابتة
- `GET /aggregations/daily-inventory` - 500 عنصر موضح مسبقاً، قيمة 250000
- `GET /aggregations/monthly-inventory` - بيانات شهرية موضحة مسبقاً
- `GET /aggregations/daily-debt` - 15000 دين موضح مسبقاً
- `GET /aggregations/monthly-debt` - دين شهري موضح مسبقاً
- `GET /aggregations/daily-profit` - 1500 ربح موضح مسبقاً
- `GET /aggregations/monthly-profit` - 45000 ربح شهري موضح مسبقاً
- `GET /aggregations/status` - يرجع طوابع زمنية مزيفة
- `POST /aggregations/update` - لا توجد منطق تحديث فعلي

**التأثير:** لوحة البيانات ستعرض بيانات غير صحيحة؛ التقارير غير موثوقة؛ القرارات التجارية المبنية على التجميعات ستكون خاطئة.

**الملفات المتأثرة:**
- [backend/internal/api/aggregations.go](backend/internal/api/aggregations.go)

---

### 1.2 المشتريات - منطق دفتر الموردين معطل
**الموقع:** [backend/internal/purchases/service.go](backend/internal/purchases/service.go#L121-L170)

**المشكلة:** منطق دفتر الموردين وسجل التدقيق كاملاً معلق مع علامات TODO.

**الأكواد المعلقة:**
- تتبع رصيد الموردين معطل
- تسجيل معاملات دفتر الموردين معطل
- إنشاء سجل التدقيق معطل (مؤقتاً)

**التأثير:** تتبع الديون للموردين لن يعمل؛ لا يوجد سجل معاملات؛ لا يمكن تسوية حسابات الموردين.

**الملفات المتأثرة:**
- [backend/internal/purchases/service.go](backend/internal/purchases/service.go)

---

### 1.3 خدمة Worker - مهام الخلفية غير كاملة
**الموقع:** [backend/worker/main.go](backend/worker/main.go#L61-L115)

**المشكلة:** عدة workers خلفية ليس لها تطبيق (تعليقات TODO فقط).

**Unimplemented Workers:**
- منطق انتهاء الحجوزات (السطر 61)
- عامل فحص الديون (السطر 70)
- عامل فحص المخزون المنخفض (السطر 83)
- توليد الرؤى اليومية (السطر 96)

**التأثير:** الحجوزات لن تنتهي؛ الديون المتأخرة لن يتم تجاهلها؛ تنبيهات المخزون المنخفض لن تعمل.

**الملفات المتأثرة:**
- [backend/worker/main.go](backend/worker/main.go)

---

### 1.4 تحديد معدل الطلب - لم يتم التنفيذ
**الموقع:** [backend/pkg/middleware/middleware.go](backend/pkg/middleware/middleware.go#L360-L371)

**المشكلة:** middleware تحديد معدل الطلب هو عنصر نائب لا يفعل شيئاً.

**الكود:**
```go
// TODO: تطبيق تحديد معدل طلب مناسب مع Redis
// في الوقت الحالي، هذا عنصر نائب
return func(c *gin.Context) {
    c.Next()
}
```

**التأثير:** لا توجد حماية ضد سوء استخدام الـ API؛ لا توجد تخفيف من هجمات DDoS؛ قد يؤدي إلى إساءة استخدام الخدمة.

---

## 2. 🟠 مشاكل أولوية عالية

### 2.1 عدم توافق قاعدة البيانات - مشاكل متعددة

#### 2.1.1 عدم توافق قاعدة بيانات المخزون
**الملفات:**
- [backend/migrations/001_initial_schema.sql](backend/migrations/001_initial_schema.sql) - يستخدم جدول `inventory`
- [backend/internal/localdb/localdb.go](backend/internal/localdb/localdb.go#L550-L620) - يتوقع جدول `inventory_items`
- [backend/internal/inventory/service.go](backend/internal/inventory/service.go) - يشير إلى عناصر المخزون

**المشكلة:** عدم اتساق بين schema PostgreSQL (يستخدم `inventory`) و SQLite محلي (يستخدم `inventory_items`).

**التأثير:** المزامنة بين قاعدة البيانات المحلية والسحابية ستفشل؛ مشاكل في هجرة البيانات.

#### 2.1.2 جدول دفتر الموردين المفقود
**الموقع:** [backend/internal/purchases/service.go](backend/internal/purchases/service.go#L121-L145)

**المشكلة:** الكود يشير إلى جدول `supplier_ledger` لكنه لم يتم إنشاؤه في الترحيلات.

**الحالة:** يوجد فقط customer_ledger؛ إنشاء supplier_ledger معلق.

**التأثير:** تتبع دفعات الموردين لن ينقذ؛ لا يوجد مسار تدقيق.

#### 2.1.3 مشكلة schema سجلات التدقيق
**الموقع:** [backend/internal/purchases/service.go](backend/internal/purchases/service.go#L147-L168)

**المشكلة:** إنشاء سجل التدقيق معلق مع TODO؛ غير مؤكد إذا كان جدول `audit_logs` موجوداً ويطابق schema المتوقع.

---

### 2.2 المكتسبات - تدفق إنشاء المخزون معطل
**الموقع:**
- [backend/internal/acquisitions/service.go](backend/internal/acquisitions/service.go#L43-L130)
- [backend/internal/acquisitions/handler.go](backend/internal/acquisitions/handler.go#L20-L80)
- [backend/internal/inspections/repository.go](backend/internal/inspections/repository.go#L60-L140)

**المشاكل:**

1. **لم يتم إنشاء عناصر مخزون** - عند إنشاء اكتساب، يتم إدراج عناصر الاستحواذ لكن لم يتم ربطها بجدول `inventory_items`
2. **LinkItemToInventory موجودة** لكنها تحدث الحالة فقط، لا تنشئ العنصر
3. **تحديث حالة الفحص** يحدث عناصر الاستحواذ لكن ربط المخزون غير واضح
4. **لا يوجد تتبع للكمية** - عناصر الاستحواذ لا تؤثر على كميات المخزون الفعلية

**مشكلة التدفق:**
```
تم إنشاء الاستحواذ
    ↓
تم إنشاء عناصر الاستحواذ (في جدول acquisition_items)
    ↓
تم إجراء الفحص (تم تحديث الحالة)
    ↓
❌ لم يتم إنشاء inventory_items تلقائياً
❌ لم يتم تحديث كميات المخزون
❌ لا توجد طريقة لبيع هذه العناصر من المخزون
```

**التأثير:** تم استحواذ قطع مستعملة لكن غير متاحة للبيع؛ تعداد المخزون يبقى صفراً؛ يظهر كمعاملة مفقودة.

**الملفات المتأثرة:**
- [backend/internal/acquisitions/service.go](backend/internal/acquisitions/service.go)
- [backend/internal/inventory/service.go](backend/internal/inventory/service.go)
- [backend/internal/inspections/repository.go](backend/internal/inspections/repository.go)

---

### 2.3 نظام المزامنة - فجوات متعددة

#### 2.3.1 تدفق مزامنة غير مكتمل
**الموقع:**
- [backend/internal/sync/service.go](backend/internal/sync/service.go#L1-L220)
- [backend/internal/sync/worker.go](backend/internal/sync/worker.go#L1-L83)

**المشاكل:**
1. قائمة الانتظار غير المتصل في الواجهة الأمامية (`localStorage`) منفصلة عن `sync_queue` الخلفي
2. لا توجد ضمانة لحفظ العمليات في `sync_queue` أثناء الإنشاء
3. اكتشاف التضارب موجود لكن الحل غير مرتبط بتدفق المزامنة
4. تبديل نمط التشغيل (غير متصل/متصل) لا ينشئ مزامنة بيانات أولية

#### 2.3.2 جدول تضارب المزامنة غير المستخدم
**الموقع:** [backend/internal/localdb/localdb.go](backend/internal/localdb/localdb.go#L1200-L1400)

**المشكلة:** تم إنشاء جدول `sync_conflicts` واكتشاف التضارب موجود، لكن:
- لا توجد حل تضارب تلقائي
- نقطة نهاية الحل اليدوي موجودة لكن منطق العمل غير مكتمل
- لا توجد واجهة مستخدم لحل التضارب في الواجهة الأمامية

**التأثير:** إذا حدث تضارب، يلزم التدخل اليدوي؛ قد تضيع البيانات.

#### 2.3.3 التحقق من استيثاق السحابة فقط في Middleware
**الموقع:** [backend/pkg/middleware/middleware.go](backend/pkg/middleware/middleware.go#L40-L95)

**المشكلة:** التحقق من السحابة يتحقق من الاشتراك لكن فقط على مستوى المسار، وليس لقرارات منطق العمل.

**المفقود:** حالة الاشتراك لم يتم التحقق منها قبل إنشاء المبيعات/المشتريات/الدفعات.

**التأثير:** يمكن للمستخدم إجراء عمليات بعد انتهاء الاشتراك إذا عمل بسرعة.

---

### 2.4 معالجة الأخطاء - أنماط غير متسقة

#### 2.4.1 معالج الاستحواذ - عدم التغليف الخطأ المفقود
**الموقع:** [backend/internal/acquisitions/handler.go](backend/internal/acquisitions/handler.go#L20-L252)

**المشكلة:** يستخدم أخطاء `gin.H` عادية بدلاً من أنواع أخطاء مناسبة.

**الحالي:**
```go
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
```

**المتوقع:** يجب أن يستخدم `errors.HandleError()` مثل المعالجات الأخرى.

**المقارنة مع:**
- [backend/internal/customers/handler.go](backend/internal/customers/handler.go#L60-L120) - يستخدم معالجة خطأ مناسبة
- [backend/internal/products/handler.go](backend/internal/products/handler.go#L70-L120) - يستخدم error.HandleError()

**التأثير:** استجابات أخطاء API غير متسقة؛ لا يمكن للعملاء تحليل الأخطاء بموثوقية.

#### 2.4.2 معالج الفحص - نفس المشكلة
**الموقع:** [backend/internal/inspections/handler.go](backend/internal/inspections/handler.go#L30-L90)

**المشكلة:** يستخدم أخطاء `gin.H` عادية بدلاً من برنامج معالجة الأخطاء.

#### 2.4.3 معالج المرتجعات - نفس المشكلة
**الموقع:** [backend/internal/returns/handler.go](backend/internal/returns/handler.go#L30-L90)

**المشكلة:** نمط متسق من استخدام `gin.H{"error": ...}` بدلاً من معالجات الأخطاء.

---

## 3. 🟡 مشاكل أولوية متوسطة

### 3.1 التحقق المفقود من الطلبات

#### 3.1.1 الاستحواذ - عدم التحقق من نوع
**الموقع:** [backend/internal/acquisitions/handler.go](backend/internal/acquisitions/handler.go#L20-L45)

**المشكلة:** تم التحقق من النوع "SUPPLIER" مقابل "CUSTOMER" في الخدمة لكن ليس برسائل خطأ واضحة.

**المشكلة:** إذا قدم المستخدم نوعاً غير صحيح، رسالة الخطأ غير واضحة.

#### 3.1.2 المخزون - انتقالات الحالة غير المدققة
**الموقع:** [backend/internal/inventory/service.go](backend/internal/inventory/service.go#L600-L658)

**انتقالات الحالة المحددة لكن:**
- لا يوجد التحقق على الواجهة الأمامية قبل إرسال الطلب
- التحقق من الخلفية يحدث متأخراً في طبقة الخدمة
- رسائل الخطأ ليست صديقة للمستخدم

**الانتقالات الصحيحة:**
- PURCHASED → RECEIVED, INSPECTION
- RECEIVED → INSPECTION, AVAILABLE
- INSPECTION → AVAILABLE, DAMAGED, IN_REPAIR
- إلخ.

#### 3.1.3 عمليات قائمة المزامنة - التحقق المفقود
**الموقع:** [backend/internal/sync/service.go](backend/internal/sync/service.go#L107-L165)

**المشكلة:** دالة `syncOneItem` لا تدقق payload قبل استخدامه.

**المخاطر:**
- معرفات فارغة قد تسبب فشل صامت
- يجب أن تفشل أنواع العمليات غير الصحيحة بسرعة
- حقول الحمل المفقود غير محتاجة

### 3.2 مشاكل تدفق البيانات

#### 3.2.1 إنشاء Trade-In - تطبيق غير مكتمل
**الموقع:** [backend/internal/inventory/handler.go](backend/internal/inventory/handler.go#L500-L580)

**المشكلة:** تقبل نقطة نهاية Trade-in أسماء العملاء/المنتجات اليدوية لكن تخزنها كملاحظات فقط.

**المشكلة:**
- إذا لم يتم توفير customer_id، يتم تعيينه إلى nil واسم العميل مخزن في الملاحظات
- إذا لم يتم توفير product_id، تعيين إلى nil واسم المنتج مخزن في الملاحظات
- قد تفشل الاستعلامات اللاحقة أو ترجع نتائج فارغة
- لا يوجد ربط معاملة للاستحواذات

**مثال الكود:**
```go
if req.CustomerID != nil {
    customerID = *req.CustomerID
} else {
    customerID = uuid.Nil  // ❌ تعيين إلى nil بدلاً من إنشاء عميل
    existingNotes := ""
    if req.Notes != nil {
        existingNotes = *req.Notes
    }
    customerNote := fmt.Sprintf("زبون: %s - %s", req.CustomerName, existingNotes)
    req.Notes = &customerNote
}
```

**التأثير:** عناصر Trade-in غير مرتبطة بشكل صحيح بالعملاء؛ التقارير لن تعرض معلومات العميل.

#### 3.2.2 إنشاء المخزون - سجل الحركة المفقود
**الموقع:** [backend/internal/inventory/service.go](backend/internal/inventory/service.go#L180-L260)

**المشكلة:** منطق ReceiveItem ينشئ سجل حركة لكن `currentQuantity` لم يتم تعيينه أبداً.

**الكود:**
```go
var currentQuantity int
// ❌ لم يتم التعيين من قاعدة البيانات
movement := &InventoryMovement{
    BeforeQuantity: currentQuantity - 1,  // سيكون -1
    AfterQuantity:  currentQuantity,      // سيكون 0
}
```

**التأثير:** سجل الحركة سيعرض كميات غير صحيحة قبل/بعد.

### 3.3 عدم توافق الواجهة الأمامية والخلفية

#### 3.3.1 نقاط نهاية الـ API محددة لكن التطبيق مفقود
**الملفات:**
- [frontend/src/services/api/endpoints.ts](frontend/src/services/api/endpoints.ts#L120-L170) - يحدد نقاط النهاية
- معالجات الخلفية موجودة لكن بعضها يحتوي على بيانات موضحة مسبقاً

**أمثلة:**
- نقاط نهاية التجميع (الملخصات اليومية/الشهرية)
- نقطة نهاية سجل العنصر
- نقطة نهاية فحص الباركود

#### 3.3.2 عدم اتساق ترجمة رسالة الخطأ
**الموقع:**
- [frontend/src/lib/error-handling.ts](frontend/src/lib/error-handling.ts) - مفاتيح إنجليزية
- [frontend/src/lib/error-messages.ts](frontend/src/lib/error-messages.ts) - رسائل عربية
- الخلفية ترجع رموز خطأ إنجليزية

**المشكلة:** تتوقع الواجهة الأمامية رسائل عربية لكن الخلفية ترجع رموز إنجليزية؛ طبقة الترجمة موجودة لكن قد لا تغطي جميع الحالات.

### 3.4 مشاكل استعلام قاعدة البيانات

#### 3.4.1 استعلام المزامنة - لا يوجد خطأ على جداول فارغة
**الموقع:** [backend/internal/sync/handler.go](backend/internal/sync/handler.go#L20-L80)

**المشكلة:** إذا لم يكن الجدول موجوداً، يتم التعامل مع الخطأ بصمت:

```go
rows, err := h.getSnapshotTable(item.table)
if err != nil {
    // يرجع مصفوفة فارغة، وليس خطأ
    return fmt.Errorf("fetch %s from online db: %w", item.key, err)
}
```

**السلوك المتوقع:** يجب التمييز بين "لا توجد صفوف" و "الجدول غير موجود"

#### 3.4.2 تحديث حالة الفحص - قد تحدث حالة تنافسية
**الموقع:** [backend/internal/inspections/repository.go](backend/internal/inspections/repository.go#L60-L140)

**المشكلة:** يتم إجراء التحديثات بالتسلسل بدون معاملة:
1. تحديث حالة عناصر الاستحواذ
2. تحديث حالة عناصر المخزون
3. تحديث حالة الاستحواذ

**المشكلة:** إذا فشلت الخطوة 2 أو 3، تم تحديث عناصر الاستحواذ بالفعل.

---

## 4. 🟢 مشاكل أولوية منخفضة

### 4.1 جودة الكود

#### 4.1.1 أنماط الخطأ المكررة
**الملفات:**
- [backend/internal/acquisitions/handler.go](backend/internal/acquisitions/handler.go)
- [backend/internal/inspections/handler.go](backend/internal/inspections/handler.go)
- [backend/internal/returns/handler.go](backend/internal/returns/handler.go)
- [backend/internal/purchases/handler.go](backend/internal/purchases/handler.go)

**المشكلة:** الكل يستخدم `c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})` بدلاً من حزمة الأخطاء.

**الحل:** إنشاء دالة مساعدة للاتساق.

#### 4.1.2 تعليقات TODO بدون تتبع
**الملفات:**
- [backend/internal/api/aggregations.go](backend/internal/api/aggregations.go) - 7 TODOs
- [backend/internal/purchases/service.go](backend/internal/purchases/service.go) - 2 TODOs
- [backend/pkg/middleware/middleware.go](backend/pkg/middleware/middleware.go) - 1 TODO
- [backend/worker/main.go](backend/worker/main.go) - 4 TODOs

**المشكلة:** لا يوجد نظام تتبع لـ TODOs؛ غير واضح أي منها حرج مقابل لطيف.

#### 4.1.3 الأرقام السحرية والقيم الموضحة مسبقاً
**الموقع:** [backend/internal/sync/worker.go](backend/internal/sync/worker.go#L10)

**الكود:**
```go
const defaultOfflineSyncInterval = 30 * time.Second
```

**المشكلة:** فترة المزامنة صعبة التعديل؛ يجب أن تكون قابلة للتكوين.

### 4.2 مشاكل الأداء

#### 4.2.1 مشكلة الاستعلام N+1 في الاستحواذات
**الموقع:** [backend/internal/acquisitions/service.go](backend/internal/acquisitions/service.go#L90-L130)

**المشكلة:** GetAcquisitionWithItems تجلب العناصر، ثم تجري استدعاءات منفصلة لمعلومات المشتري:

```go
acquisition, err := s.GetAcquisition(ctx, id)  // الاستعلام 1
var items []AcquisitionItem
err = s.db.Select(&items, query, id)           // الاستعلام 2
var seller *SellerInfo
if acquisition.Type == TypeSupplier {
    seller, err = s.getSupplierInfo(ctx, *acquisition.SupplierID) // الاستعلام 3
}
```

**الحل:** استخدام JOIN في الاستعلام الرئيسي.

#### 4.2.2 الذاكرة المؤقتة غير المستخدمة باستمرار
**الموقع:** [backend/internal/customers/handler.go](backend/internal/customers/handler.go#L60-L120)

**المشكلة:** فقط الصفحة الأولى بدون عوامل تصفية يتم تخزينها مؤقتاً؛ لم يتم توضيح إبطال الذاكرة المؤقتة.

**المشكلة:** إذا تم إضافة عميل جديد، لم يتم إبطال الذاكرة المؤقتة؛ يتم إرجاع البيانات القديمة.

---

## 5. 📋 فجوات التحقق من البيانات

### 5.1 الحقول التي يجب أن يكون لها التحقق لكن لا تملكها

| الحقل | الحالي | المشكلة |
|--------|---------|--------|
| Acquisition.AcquisitionDate | تاريخ نصي | عدم التحقق من نطاق التاريخ |
| Purchase.ExpectedDeliveryDate | اختياري | يمكن أن يكون في الماضي |
| Inspection.InspectionDate | تاريخ نصي | عدم التحقق من نطاق التاريخ |
| Return.ReturnDate | تاريخ نصي | عدم التحقق من نطاق التاريخ |
| Expense.Amount | عائم | عدم التحقق من التحقق من الحد الأقصى |
| Debt.DueDate | تاريخ نصي | يمكن أن يكون في الماضي عند الإنشاء |

### 5.2 مشاكل القيد الفريد

**الملفات:**
- [backend/internal/products/dto.go](backend/internal/products/dto.go) - يجب أن يكون SKU الخاص بالمنتج فريداً، لكن التحقق فقط في قاعدة البيانات
- [backend/internal/customers/handler.go](backend/internal/customers/handler.go) - التحقق من رمز العميل غير واضح

**المشكلة:** انتهاكات القيد الفريد ترجع خطأ عام "موجود بالفعل" بدون إرشادات.

---

## 6. 📊 مشاكل المزامنة/الإصدار

### 6.1 تبديل نمط التشغيل لا ينشئ مزامنة
**الموقع:** [backend/internal/settings/local_database_handler.go](backend/internal/settings/local_database_handler.go#L200-L330)

**المشكلة:** عند الانتقال من وضع غير متصل إلى وضع متصل، لا تحدث مزامنة بيانات أولية تلقائياً.

**تدفق متوقع:**
1. ينقر المستخدم على "المزامنة مع السحابة"
2. يجلب النظام جميع الجداول
3. يدمج مع التغييرات المحلية
4. يبلغ عن التضاربات

**التدفق الفعلي:**
1. ينقر المستخدم على "المزامنة مع السحابة"
2. يستبدل النظام بيانات محلية برمز سحابي
3. أي تغييرات محلية غير متصلة يتم فقدانها

### 6.2 لا توجد مزامنة حالة أولية
**الموقع:** [backend/internal/sync/service.go](backend/internal/sync/service.go#L1-L60)

**المشكلة:** يتم استدعاء `SeedLocalDatabaseFromOnline` لكن التوقيت غير واضح.

**المشكلة:** إذا تم استدعاؤها أثناء التشغيل العادي، يستبدل جميع البيانات المحلية.

### 6.3 اكتشاف التضارب بدون حل
**الموقع:** [backend/internal/sync/service.go](backend/internal/sync/service.go#L127-L165)

**المشكلة:** يتم اكتشاف التضاربات وتسجيلها لكن:
- لا توجد استراتيجية حل تضارب تلقائية
- نقطة نهاية الحل اليدوي موجودة لكن منطق العمل غير مكتمل
- لا يوجد رد فعل واجهة مستخدم للمستخدم

---

## 7. 🔑 الملفات الرئيسية التي تحتاج إلى مراجعة

| الملف | المشاكل | الأولوية |
|------|--------|----------|
| [backend/internal/api/aggregations.go](backend/internal/api/aggregations.go) | بيانات موضحة مسبقاً | حرج |
| [backend/internal/purchases/service.go](backend/internal/purchases/service.go) | منطق دفتر معطل | حرج |
| [backend/worker/main.go](backend/worker/main.go) | workers غير مطبق | حرج |
| [backend/internal/acquisitions/service.go](backend/internal/acquisitions/service.go) | تدفق مخزون معطل | حرج |
| [backend/internal/sync/service.go](backend/internal/sync/service.go) | فجوات المزامنة | عالي |
| [backend/pkg/middleware/middleware.go](backend/pkg/middleware/middleware.go) | لا يوجد تحديد معدل | عالي |
| [backend/internal/inspections/repository.go](backend/internal/inspections/repository.go) | قد تحدث حالات تنافسية | عالي |
| [backend/internal/inventory/service.go](backend/internal/inventory/service.go) | تتبع الحركة معطل | متوسط |
| [backend/internal/acquisitions/handler.go](backend/internal/acquisitions/handler.go) | معالجة أخطاء غير متسقة | متوسط |

---

## 8. 🎯 الإصلاحات الموصى بها (ترتيب الأولوية)

### المرحلة 1: حرج (يحجب الاستخدام الحالي)
1. ✅ تطبيق استعلامات التجميع (إزالة البيانات المزيفة)
2. ✅ إلغاء التعليق والاختبار منطق دفتر الموردين
3. ✅ تطبيق workers الخلفية
4. ✅ إصلاح تدفق الاستحواذ → إنشاء المخزون
5. ✅ إصلاح تتبع كمية حركة المخزون

### المرحلة 2: عالي (يحجب الميزات الجديدة)
1. ✅ تطبيق تحديد معدل الطلب
2. ✅ إصلاح حل تضارب المزامنة
3. ✅ إضافة معاملة إلى تحديثات الفحص
4. ✅ توحيد معالجة الأخطاء
5. ✅ التحقق من جميع مدخلات المستخدم

### المرحلة 3: متوسط (تحسين الموثوقية)
1. ✅ إصلاح استعلامات N+1
2. ✅ تطبيق إبطال الذاكرة المؤقتة المناسب
3. ✅ إضافة نطاقات التحقق من البيانات
4. ✅ إصلاح ربط Trade-in → المخزون
5. ✅ إكمال تدفق مزامنة نمط التشغيل

---

## 9. 📝 توصيات الاختبار

### الاختبارات التي يجب أن تفشل (تكشف الأخطاء)
```go
// الاختبار 1: دقة التجميع
func TestDailySalesSummaryAgainstActualData(t *testing.T) {
    // حالياً يرجع 500 عنصر موضح مسبقاً دائماً
    // يجب أن يرجع التعداد الفعلي
}

// الاختبار 2: الاستحواذ ينشئ مخزون
func TestAcquisitionCreatesInventoryItems(t *testing.T) {
    // ينشئ acquisition_items لكن بدون inventory_items
    // يجب إنشاء كليهما
}

// الاختبار 3: تحديث دفتر الموردين
func TestSupplierLedgerUpdatedOnPurchase(t *testing.T) {
    // حالياً معطل/معلق
    // يجب إنشاء إدخال دفتر
}

// الاختبار 4: حل تضارب المزامنة
func TestConflictResolutionApplied(t *testing.T) {
    // التضاربات مسجلة لكن لم يتم حلها
    // يجب تطبيق سياسة الحل
}

// الاختبار 5: تتبع كمية حركة المخزون
func TestInventoryMovementQuantitiesCorrect(t *testing.T) {
    // currentQuantity لم يتم تعيينه، يعرض دائماً -1 قبل
    // يجب عرض القيم الصحيحة
}
```

---

## 10. 📞 الخطوات التالية

1. **إنشاء مشاكل GitHub** لكل عنصر حرج مع خطوات إعادة الإنتاج
2. **تخصيص Sprint** لإصلاحات المرحلة 1
3. **إضافة اختبارات الوحدة** لجميع الوظائف المصححة
4. **مراجعة معالجة الأخطاء** عبر جميع المعالجات
5. **تدقيق استعلامات قاعدة البيانات** للحصول على مشاكل N+1
6. **اختبار تدفق المزامنة** مع سيناريوهات غير متصلة حقيقية
7. **التحقق من جميع المدخلات** مع مجموعة اختبارات شاملة

---

**نهاية التقرير**
