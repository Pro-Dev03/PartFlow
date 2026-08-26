# PartFlow - مبادئ التصميم المعماري

## نظرة عامة

هذا الملف يوثق الفلسفة المعمارية الأساسية لـ PartFlow، مبنية على مبادئ الاستدامة وقابلية التوسع للسنوات الطويلة.

---

## المبدأ الأساسي: لا تحذف السجل التجاري

### المشكلة
الحذف المباشر للسجلات التجارية (شراء، بيع، دفع، إرجاع) يؤدي إلى:
- فقدان التاريخ التجاري
- تعقيد المحاسبة والتدقيق
- عدم القدرة على تتبع الأخطاء
- صعوبة في إعادة بناء الحالة المالية

### الحل
فصل **"السجل التاريخي"** عن **"البيانات التشغيلية الحالية"**.

---

## 1. Current State + Immutable History

### المفهوم
النظام يحتفظ بحالتين من البيانات:
1. **Current State**: الحالة الحالية (ما نحتاجه للعمليات اليومية)
2. **Immutable History**: السجل التاريخي (للمراجعة والتدقيق)

### تطبيق على المخزون

#### التصميم الحالي (مشكلة)
```go
type InventoryItem struct {
    ID        uuid.UUID
    ProductID *uuid.UUID
    Quantity  int  // يُحسب من كل الحركات التاريخية
    Status    string
}
```

#### التصميم المقترح
```go
// الحالة الحالية - للعمليات اليومية
type InventoryItem struct {
    ID             uuid.UUID
    ProductID      *uuid.UUID
    Status         string

    // Current State fields
    CurrentQuantity    int     // الكمية الحالية
    ReservedQuantity   int     // المحجوز
    AvailableQuantity  int     // المتاح = Current - Reserved
    CurrentCost        float64 // التكلفة الحالية
    CurrentValue       float64 // القيمة الحالية = Quantity * Cost
    LastMovementID     *uuid.UUID // آخر حركة
    UpdatedAt          time.Time
}

// السجل التاريخي - للمراجعة والتدقيق
type InventoryMovement struct {
    ID              uuid.UUID
    ItemID          uuid.UUID
    Type            string // purchase, sale, return, adjustment, transfer, damage, repair, reserve, cancel_reserve
    QuantityBefore  int
    QuantityAfter   int
    QuantityChange  int
    CostBefore      float64
    CostAfter       float64
    ReferenceType   string // purchase, sale, adjustment
    ReferenceID     uuid.UUID
    Reason          string
    CreatedBy       uuid.UUID
    CreatedAt       time.Time
}
```

### الفوائد
- لا حاجة لإعادة حساب المخزون من كل التاريخ في كل مرة
- سجل واضح لكل حركة
- يمكن تتبع أي تغيير في المخزون
- أداء أفضل للعمليات اليومية

---

## 2. Reverse بدلاً من Delete (آلية داخلية فقط)

### النهج العملي للمستخدم
المستخدم يرى فقط **"حذف العملية"** والنظام يقرر داخلياً ماذا يفعل:

#### الحالة 1 — مسودة (Draft)
```
حذف → DELETE فعلي
```

#### الحالة 2 — مستلمة ولم يحدث عليها شيء
```
حذف → Delete with inventory adjustment
```
النظام يقول: "سيتم حذف عملية الشراء وإزالة الكمية المرتبطة بها من المخزون."

#### الحالة 3 — هناك عمليات لاحقة
```
حذف → Block Delete مع رسالة واضحة
```
النظام يقول: "لا يمكن حذف عملية الشراء لأنها مرتبطة بعمليات بيع. يمكنك عرض التفاصيل أو معالجة الإرجاع."

### القاعدة الذهبية
> **النظام يجب أن يكون معقدًا من الداخل حتى لا يكون معقدًا على صاحب المحل.**

صاحب المحل لا يحتاج فهم:
- Reverse / Ledger / Transaction / Inventory Movement

هو يفكر فقط:
- شراء → استلام → مخزون → بيع

### المفهوم
لا نحذف السجلات التجارية المؤكدة، بل نعمل على عكسها (Reverse).

### قواعد الحذف والعكس

#### يُسمح بالحذف (DELETE)
- Draft purchases (مسودات الشراء)
- Draft sales (مسودات البيع)
- أي سجل في حالة "Draft"

#### يُستخدم العكس (REVERSE)
- Confirmed purchases (شراء مؤكد)
- Received purchases (شراء مستلم)
- Confirmed sales (بيع مؤكد)
- Payments (دفعات)
- Returns (مرتجعات)
- Inventory adjustments (تعديلات المخزون)

### مثال: شراء مكرر

#### التسلسل الزمني
```
2027-01-15: Purchase #104 - $2,400 (Received)
2027-01-20: اكتشف أن الفاتورة مكررة
2027-01-20: Purchase #104 - Reversed - Reason: Duplicate invoice
```

#### تنفيذ العكس
```go
type Purchase struct {
    ID          uuid.UUID
    Status      string // draft, confirmed, received, reversed
    Total       float64
    // ... fields
}

type PurchaseReversal struct {
    ID              uuid.UUID
    PurchaseID      uuid.UUID
    Reason          string
    ReversedBy      uuid.UUID
    ReversedAt      time.Time
    OriginalTotal   float64
}
```

### الفوائد
- تاريخ واضح حتى بعد سنوات
- سهولة التدقيق والمحاسبة
- يمكن تتبع الأخطاء
- لا فقدان للبيانات

### التنفيذ العملي
المستخدم يرى فقط زر **"حذف"** والنظام يقرر داخلياً:

```go
// Frontend calls this simple endpoint
DELETE /api/purchases/{id}

// Backend SmartDeleteService decides:
// - Draft → DELETE فعلي
// - Received + لا توجد عمليات لاحقة → REVERSE داخلي
// - Received + توجد عمليات لاحقة → BLOCK مع رسالة واضحة
```

#### الملفات المُحدّثة
- `backend/internal/purchases/model.go`: إضافة حقول Reversal و PurchaseReversal struct
- `backend/internal/purchases/reversal.go`: ReversalService للشراء (آلية داخلية)
- `backend/internal/purchases/smart_delete.go`: SmartDeleteService - واجهة بسيطة للمستخدم
- `backend/internal/payments/model.go`: إضافة حقول Reversal و PaymentReversal struct
- `backend/internal/payments/reversal.go`: ReversalService للدفعات (آلية داخلية)
- `backend/internal/payments/smart_delete.go`: SmartDeleteService - واجهة بسيطة للمستخدم
- `backend/internal/sales/model.go`: إضافة حقول Reversal و SaleReversal struct
- `backend/internal/sales/reversal.go`: ReversalService للمبيعات (آلية داخلية)
- `backend/internal/sales/smart_delete.go`: SmartDeleteService - واجهة بسيطة للمستخدم

---

## 3. Aggregations للـ Dashboard

### المشكلة
إذا كان لديك 10 سنوات من البيانات:
- 5,000,000 sales
- 2,000,000 purchases
- 20,000,000 inventory movements

لا تجعل Dashboard يقرأ ملايين السجلات في كل مرة.

### الحل
استخدم نظام التجميع (Aggregations):

```
Transactions
    ↓
Aggregations
    ↓
Daily / Monthly Statistics
    ↓
Dashboard
```

### الجداول المقترحة

#### daily_sales_summary
```sql
CREATE TABLE daily_sales_summary (
    date DATE PRIMARY KEY,
    total_sales BIGINT,
    total_revenue DECIMAL(15,2),
    total_profit DECIMAL(15,2),
    total_customers INT,
    average_order_value DECIMAL(15,2),
    updated_at TIMESTAMP
);
```

#### monthly_sales_summary
```sql
CREATE TABLE monthly_sales_summary (
    year INT,
    month INT,
    total_sales BIGINT,
    total_revenue DECIMAL(15,2),
    total_profit DECIMAL(15,2),
    total_customers INT,
    average_order_value DECIMAL(15,2),
    updated_at TIMESTAMP,
    PRIMARY KEY (year, month)
);
```

#### daily_inventory_summary
```sql
CREATE TABLE daily_inventory_summary (
    date DATE PRIMARY KEY,
    total_items INT,
    total_value DECIMAL(15,2),
    low_stock_count INT,
    out_of_stock_count INT,
    updated_at TIMESTAMP
);
```

### استراتيجية التحديث
- تحديث يومي في نهاية كل يوم
- تحديث فوري للـ Dashboard الحالي
- استخدم Workers للتحديث في الخلفية

### الفوائد
- Dashboard سريع جداً
- لا ضغط على قاعدة البيانات
- يمكن توسيع النظام لسنوات
- تقارير سريعة

---

## 4. Archive Layer (Future-Ready)

### المفهوم
تصميم النظام بحيث يمكن تطبيق Archiving لاحقاً دون إعادة البنية بالكامل.

### استراتيجية Archiving المقترحة

```
0 - 2 years
    ↓
Active Database
    ↓
2 - 5 years
    ↓
Archive (أبطأ لكن رخيص)
    ↓
5+ years
    ↓
Long-term Archive / Cold Storage (سحابة، أرخص جداً)
```

### البيانات المرشحة للـ Archiving
- Sales History (أقدم من سنتين)
- Purchase History (أقدم من سنتين)
- Inventory Movements (أقدم من سنتين)
- Audit Logs (أقدم من سنة)

### البيانات التي تبقى في Active Database
- Current Inventory
- Current Customers
- Current Suppliers
- Current Balances
- Recent Transactions (آخر سنتين)

### الفوائد
- أداء أفضل للقاعدة النشطة
- تكلفة أقل للتخزين
- يمكن الوصول للتاريخ القديم عند الحاجة
- قابلية التوسع للسنوات الطويلة

---

## 5. Audit Log - لا تبالغ

### الفرق بين Business Data و Audit Data

#### Business Data (مهم جداً)
```go
- Purchase created
- Purchase received
- Purchase reversed
- Sale created
- Payment received
- Debt changed
- Inventory adjusted
- User changed financial data
```

#### Audit Data (يمكن التخلص منه أو الاحتفاظ لفترة قصيرة)
```go
- User moved mouse
- User opened modal
- User changed search filter
- User opened inventory page
- UI interactions
```

### استراتيجية Audit Log
- احتفظ بـ Business Data للأبد
- احتفظ بـ Audit Data لمدة 90 يوماً
- Audit Log يمكن أن يكون في نظام منفصل

---

## 6. Indexing Strategy

### الفهارس الأساسية
```sql
-- Sales
CREATE INDEX idx_sales_date ON sales(created_at);
CREATE INDEX idx_sales_customer ON sales(customer_id);
CREATE INDEX idx_sales_status ON sales(status);

-- Inventory Movements
CREATE INDEX idx_movements_item ON inventory_movements(item_id);
CREATE INDEX idx_movements_date ON inventory_movements(created_at);
CREATE INDEX idx_movements_type ON inventory_movements(type);

-- Payments
CREATE INDEX idx_payments_debt ON payments(debt_id);
CREATE INDEX idx_payments_date ON payments(created_at);
```

### الفهارس المركبة
```sql
-- للـ Dashboard
CREATE INDEX idx_sales_date_status ON sales(created_at, status);
CREATE INDEX idx_movements_item_date ON inventory_movements(item_id, created_at);
```

---

## 7. البنية المعمارية الكاملة

```
PARTFLOW
    │
    ┌────────────┴────────────┐
    │                         │
 OPERATIONAL DATA         HISTORICAL DATA
    │                         │
Current Inventory      Sales History
Current Customers      Purchase History
Current Suppliers      Inventory Ledger
Current Balances       Payments
    │                     Returns
    │                         │
    └──────────┬──────────────┘
               │
        ARCHIVE LAYER
               │
    Old historical data
               │
        Cold Storage
```

---

## 8. خطة التطبيق

### المرحلة 1: الأساسيات (الآن)
1. تطبيق Current State + Immutable History على المخزون
2. تطبيق Reverse بدلاً من Delete للعمليات المالية
3. إضافة جداول Aggregation للـ Dashboard

### المرحلة 2: التحسينات (لاحقاً)
1. إضافة Worker لتحديث Aggregations
2. تحسين الفهارس
3. إضافة Audit Log system

### المرحلة 3: Archiving (مستقبلاً)
1. تصميم استراتيجية Archiving
2. تطبيق Archiving للبيانات القديمة
3. إضافة Cold Storage

---

## 9. PostgreSQL ملاحظات عن

### PostgreSQL قوي جداً
- يمكنه التعامل مع ملايين السجلات
- التصميم الصحيح والفهارس المهمة جداً
- لا تخف من البيانات الكبيرة

### المشكلة الحقيقية
ليست: "ماذا لو أصبحت قاعدة البيانات ضخمة؟"

بل: "هل صممت النظام بحيث لا يحتاج كل طلب إلى قراءة تاريخ 10 سنوات؟"

---

## الخلاصة

### مبادئ PartFlow المعمارية
1. **Current State + Immutable History**: احتفظ بحالة حالية + سجل تاريخي
2. **Reverse بدلاً من Delete**: آلية داخلية فقط - المستخدم يرى "حذف" والنظام يقرر
3. **Aggregations للـ Dashboard**: لا تجعل Dashboard يقرأ ملايين السجلات
4. **Archive Layer Future-ready**: صمم للـ Archiving لاحقاً
5. **Audit Log Balanced**: لا تبالغ في تسجيل كل شيء
6. **Indexing Strategy**: فهارس ذكية للأداء

### النتيجة
نظام يستطيع العمل لسنوات طويلة بدون أن تضطر إلى حذف التاريخ التجاري المهم.

---

## المراجع
- AGENTS.md: التوثيق الرئيسي للمشروع
- ARCHITECTURE_ANALYSIS.md: تحليل البنية المعمارية الحالية
