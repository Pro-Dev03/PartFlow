# PartFlow - نظام المبيعات والعملاء والديون

**التاريخ:** 2026-08-26  
**الهدف:** تطبيق فلسفة المبيعات والعملاء والديون على PartFlow

---

## ✅ ما تم إنجازه

### 1. فلسفة المبيعات المطبقة (SALES-PHILOSOPHY.md)

#### المبدأ الأساسي:
> **النظام يعمل من أجل صاحب المتجر، وليس صاحب المتجر يعمل من أجل النظام.**

#### التطبيق العملي:
- ✅ **فصل مسار البيع**: مدفوع بالكامل vs بالدين
- ✅ **البيع المدفوع**: لا يتطلب عميل (أسرع مسار)
- ✅ **البيع بالدين**: يتطلب تحديد العميل (ضروري)
- ✅ **الحساب التلقائي**: النظام يحسب الدين تلقائياً
- ✅ **حد الدين**: تنبيهات ذكية مع قرار صاحب المتجر

---

### 2. تحديث نظام الدفع (PaymentSection)

#### أ. إضافة طريقة الدفع "آجل"
```typescript
// Before: PaymentMethod = 'cash' | 'card' | 'bank_transfer' | 'credit'
// After: PaymentMethod = 'cash' | 'card' | 'credit' (simplified for POS)
```

#### ب. منطق منع البيع بالدين بدون عميل
```typescript
// For credit sales, customer is required (SALES-PHILOSOPHY.md)
const isCreditSaleWithoutCustomer = isCreditSale && !selectedCustomer;
```

#### ج. تنبيهات حد الدين
```typescript
// Check if customer will exceed credit limit (SALES-PHILOSOPHY.md)
const projectedDebt = customerBalance + remaining;
const willExceedCreditLimit = customerCreditLimit && projectedDebt > customerCreditLimit;
```

#### د. زر إنشاء عميل سريع
```typescript
// Quick Customer Creation for Credit Sales (SALES-PHILOSOPHY.md)
{isCreditSale && !selectedCustomer && onQuickCustomerCreate && (
  <Button onClick={onQuickCustomerCreate}>
    + إنشاء عميل سريع
  </Button>
)}
```

---

### 3. تحديث صفحة البيع (POSPage)

#### أ. إضافة حالة للعميل السريع
```typescript
// Quick customer creation modal (SALES-PHILOSOPHY.md)
const [isQuickCustomerModalOpen, setIsQuickCustomerModalOpen] = useState(false);
const [customerBalance, setCustomerBalance] = useState(0);
const [customerCreditLimit, setCustomerCreditLimit] = useState<number | undefined>();
```

#### ب. معالج اختيار العميل
```typescript
// Handle customer selection to load balance info (SALES-PHILOSOPHY.md)
const handleCustomerChange = (customerId: string) => {
  setSelectedCustomer(customerId);
  if (customerId) {
    const customer = customers.find((c: any) => c.id === customerId);
    if (customer) {
      setCustomerBalance(customer.balance || 0);
      setCustomerCreditLimit(customer.credit_limit);
    }
  }
};
```

#### ج. نافذة إنشاء عميل سريع
```typescript
// Quick Customer Creation Modal (SALES-PHILOSOPHY.md)
<Modal
  isOpen={isQuickCustomerModalOpen}
  title="إضافة عميل سريع"
  variant="modern"
  size="md"
>
  <Input placeholder="أدخل اسم العميل..." autoFocus />
  <Input type="tel" placeholder="05xxxxxxxx" />
  <Input type="number" placeholder="حد الدين" />
  <Button>حفظ ومتابعة البيع</Button>
</Modal>
```

---

### 4. تحديث خدمة المبيعات (Sales Service)

#### أ. الحساب التلقائي للدين
```go
// Calculate debt amount (SALES-PHILOSOPHY.md)
debtAmount := totalAmount - req.PaymentAmount
newBalance := currentBalance + debtAmount

// Create ledger entry for the sale
ledgerQuery := `
    INSERT INTO customer_ledger (customer_id, transaction_type, amount, balance, ...)
    VALUES ($1, "SALE", $2, $3, ...)
`
```

#### ب. فحص حد الدين
```go
// Check credit limit (SALES-PHILOSOPHY.md)
var creditLimit *float64
limitQuery := `SELECT credit_limit FROM customers WHERE id = $1`
if newBalance > *creditLimit {
    // Log warning but don't fail the sale (store owner's decision)
    fmt.Printf("Warning: Customer will exceed credit limit...")
}
```

#### ج. المعاملة الذرية
```go
// Single transaction handles:
// - Create sale
// - Update inventory
// - Calculate profit
// - Record debt
// - Update customer balance
// - Create payment record
// - Create warranty
// - Create audit log
```

---

### 5. القواعد المطبقة

#### ✅ القاعدة 1: ليس كل بيع يحتاج إلى ملف عميل
- البيع المدفوع بالكامل: لا يتطلب عميل
- البيع بالدين: يتطلب عميل (إلزامي)

#### ✅ القاعدة 2: كل بيع بالدين يجب أن يرتبط بزبون
```typescript
const isCreditSaleWithoutCustomer = isCreditSale && !selectedCustomer;
// Checkout button disabled if true
```

#### ✅ القاعدة 3: إنشاء الزبون يجب أن يكون سريعًا
- نافذة منبثقة من شاشة البيع
- حقول أساسية فقط (الاسم، الهاتف، حد الدين)
- زر "حفظ ومتابعة البيع"

#### ✅ القاعدة 4: الرصيد يحسب تلقائيًا
```go
debtAmount := totalAmount - req.PaymentAmount
newBalance := currentBalance + debtAmount
```

#### ✅ القاعدة 5: حد الدين اختياري مع تنبيهات
```typescript
if (willExceedCreditLimit) {
  <AlertTriangle />
  <span>تنبيه: هذا البيع سيتجاوز حد الدين المحدد</span>
}
```

#### ✅ القاعدة 6: المعاملة الواحدة
```go
tx, err := s.db.BeginTxx(ctx, nil)
// All operations in one transaction
tx.Commit()
```

---

### 6. تجربة المستخدم المحسنة

#### قبل التطبيق:
```
Sale → Select Customer (mandatory) → Payment → Complete
```

#### بعد التطبيق:
```
Sale (Cash/Card) → Payment → Complete
Sale (Credit) → Select Customer → Payment → Complete
```

#### الفرق:
- البيع المدفوع: أصبح أسرع (لا حاجة لعميل)
- البيع بالدين: أصبح أوضح (تنبيهات واضحة)
- إنشاء عميل: أصبح سريع (من داخل البيع)

---

### 7. الرسائل المحسنة (لغة صاحب المتجر)

#### ❌ قبل:
```
"Customer required for credit sale"
"Credit limit exceeded"
```

#### ✅ بعد:
```
"البيع بالدين يتطلب تحديد العميل"
"تنبيه: هذا البيع سيتجاوز حد الدين المحدد"
```

---

### 8. التوثيق المحدث

#### أ. التعليقات في الكود
- جميع التعديلات تحتوي على `// SALES-PHILOSOPHY.md`
- توضح المبادئ المطبقة

#### ب. التوثيق الشامل
- `SALES-IMPLEMENTATION.md` (هذا الملف)
- يربط كل تغيير بالمبدأ المعماري المناسب

---

## 🎯 النموذج الكامل المطبق

```text
                       SALE
                         │
                    Scan/Search
                         │
                         ↓
                       CART
                         │
                         ↓
                  Choose Payment
                         │
             ┌───────────┴───────────┐
             │                       │
          PAID                    CREDIT
             │                       │
             ↓                       ↓
        Complete Sale          Select Customer
                                     │
                              Existing / New
                                     │
                                     ↓
                                Record Debt
                                     │
                                     ↓
                               Complete Sale
```

---

## 📊 النتائج المتوقعة

### الأداء
- **البيع المدفوع**: أسرع بسبب إزالة خطوة اختيار العميل
- **البيع بالدين**: نفس السرعة لكن مع حماية أفضل
- **المعاملات**: آمنة ومتسقة

### تجربة المستخدم
- **البساطة**: مسار واضح لكل نوع بيع
- **الوضوح**: رسائل واضحة بلغة صاحب المتجر
- **الأمان**: حماية من الأخطاء (حد الدين، عميل مطلوب)

### البيانات
- **التاريخ**: محفوظ بشكل كامل في ledger
- **الحساب**: تلقائي ودقيق
- **الاتساق**: محموم بـ transactions

---

## 🚀 الخطوات التالية المقترحة

### المرحلة القصيرة
1. اختبار مسار البيع المدفوع بالكامل
2. اختبار مسار البيع بالدين
3. اختبار إنشاء العميل السريع
4. اختبار تنبيهات حد الدين

### المرحلة المتوسطة
1. تحسين صفحة العميل لعرض المعلومات العملية
2. تحسين صفحة الديون لتكون أكثر وضوحاً
3. إضافة timeline للعميل (سجل الحركات)
4. تحسين واجهة تسجيل الدفعات

### المرحلة الطويلة
1. إضافة تنبيهات ذكية للديون المتأخرة
2. تحسين التقارير المرتبطة بالمبيعات
3. إضافة تحليلات للأرباح
4. تحسين البحث عن العملاء

---

## 📝 الملفات المعدلة/المضافة

### Frontend
- ✅ `frontend/src/features/sales/types/pos.types.ts` (معدل)
- ✅ `frontend/src/features/sales/components/PaymentSection.tsx` (معدل)
- ✅ `frontend/src/features/sales/pages/POSPage.tsx` (معدل)

### Backend
- ✅ `backend/internal/sales/service.go` (معدل)

### التوثيق
- ✅ `SALES-IMPLEMENTATION.md` (جديد - هذا الملف)

---

## ✅ الخلاصة

تم تطبيق فلسفة المبيعات والعملاء والديون بنجاح:

> **النظام يعمل من أجل صاحب المتجر، وليس صاحب المتجر يعمل من أجل النظام.**

النتيجة:
- ✅ مسارين واضحين للبيع (مدفوع vs بالدين)
- ✅ إنشاء عميل سريع من داخل البيع
- ✅ حساب تلقائي للديون
- ✅ تنبيهات حد الدين الذكية
- ✅ رسائل بلغة صاحب المتجر
- ✅ معاملات آمنة ومتسقة
- ✅ تاريخ محفوظ بشكل كامل

النظام الآن يدعم الواقع الحقيقي للمتاجر:
- زبون يدفع فورًا → بيع سريع
- زبون يشتري بالدين → حماية ووضوح

المستخدم يقول: "بعت هذه القطعة لمحمد، ودفع 500"  
النظام يحسب: الدين = الباقي، الرصيد الجديد، ويحدث كل شيء تلقائياً.

**المستخدم يقوم بالعمل التجاري، وPartFlow يتولى العمل الإداري.**