# PartFlow - Purchase Reversal Architecture Documentation

## نظرة عامة

تم تطبيق نهج معماري أكاديمي احترافي لإدارة عمليات الشراء في PartFlow، بناءً على مبدأ **"لا تحذف التاريخ التجاري، بل اعكسه"**. هذا النهج يحول النظام من نظام CRUD عادي إلى نظام تجاري احترافي مع audit trail كامل.

## المشكلة الأصلية

كان النظام السابق يستخدم "حذف كامل" لعمليات الشراء:
```go
// المشكلة: حذف الشراء = حذف المخزون تلقائياً
DeletePurchase() {
    if purchase.Status == "received" {
        // حذف عناصر المخزون المرتبطة
        DELETE FROM inventory_items WHERE item_code LIKE purchase_pattern
    }
    DELETE FROM purchases WHERE id = purchase_id
}
```

**العيوب:**
- فقدان التاريخ التجاري
- تدمير audit trail
- عدم حماية المخزون إذا تم بيع القطع
- إفساد التقارير والأرباح

## الحل الأكاديمي المطبق

### 1. Status-Based Logic بدلاً من Deletion Logic

#### حالات الشراء الجديدة
```go
const (
    StatusDraft            = "draft"             // مسودة - يمكن حذفها
    StatusPending          = "pending"           // قيد الانتظار - يمكن حذفها
    StatusReceived         = "received"          // تم الاستلام - لا يمكن حذفها، يجب عكسها
    StatusCancelled        = "cancelled"         // ملغي - لا يمكن حذفها
    StatusReversed         = "reversed"          // تم العكس - لا يمكن حذفها
    StatusPartiallyReceived = "partially_received" // استلام جزئي
)
```

#### منطق الحذف الجديد
```go
func (s *Service) DeletePurchase(ctx context.Context, id uuid.UUID) error {
    purchase, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return err
    }

    // Status-based deletion logic
    switch purchase.Status {
    case StatusDraft, StatusPending:
        // Allow full deletion for draft and pending purchases
        return s.deleteDraftPurchase(ctx, id)
    case StatusReceived, StatusPartiallyReceived:
        // Received purchases cannot be deleted directly - must be reversed
        return ErrCannotDeleteReceivedPurchase
    case StatusCancelled, StatusReversed:
        // Cancelled and reversed purchases cannot be deleted for audit trail
        return ErrInvalidPurchaseStatus
    default:
        return ErrInvalidPurchaseStatus
    }
}
```

### 2. ReversePurchase Function

الوظيفة الجديدة `ReversePurchase` تقوم بعكس العملية التجارية بدلاً من حذفها:

```go
func (s *Service) ReversePurchase(ctx context.Context, id uuid.UUID, userID uuid.UUID, reason string) (*PurchaseResponse, error) {
    // 1. التحقق من الحالة
    if purchase.Status == StatusReversed {
        return nil, ErrPurchaseAlreadyReversed
    }

    // 2. فحص إذا تم بيع أي قطعة (منع العكس إذا تم البيع)
    if soldCount > 0 {
        return nil, ErrItemsAlreadySold
    }

    // 3. عكس عناصر المخزون
    for _, item := range items {
        // تحديث حالة العنصر إلى RETURNED
        UPDATE inventory_items SET status = 'RETURNED' WHERE id = item_id

        // إنشاء حركة مخزون عكسية
        INSERT INTO inventory_movements (
            movement_type = 'PURCHASE_REVERSAL',
            quantity = -1,
            reason = 'Purchase reversal: {reason}'
        )

        // تحديث جدول المخزون التجميعي
        UPDATE inventory SET quantity = quantity - 1 WHERE product_id = product_id
    }

    // 4. تحديث حالة الشراء
    UPDATE purchases SET status = 'reversed' WHERE id = purchase_id

    // 5. إنشاء audit log
    INSERT INTO audit_logs (
        action = 'REVERSE_PURCHASE',
        entity_type = 'purchase',
        changes = 'Reversed purchase {invoice}, reason: {reason}, items affected: {count}'
    )
}
```

### 3. Status Transition Validation

```go
func isValidStatusTransition(currentStatus, newStatus string) bool {
    validTransitions := map[string][]string{
        StatusDraft:    {StatusPending, StatusCancelled},
        StatusPending:  {StatusReceived, StatusCancelled, StatusPartiallyReceived},
        StatusReceived: {StatusReversed}, // Can only reverse after receive
        StatusPartiallyReceived: {StatusReceived, StatusReversed, StatusCancelled},
        StatusCancelled: {}, // Cannot transition from cancelled
        StatusReversed:  {}, // Cannot transition from reversed
    }
    // منطق التحقق من الانتقالات المسموحة
}
```

## API Endpoints

### المسارات الجديدة

```go
// Reverse Purchase Endpoint
POST /api/v1/purchases/{id}/reverse
{
    "reason": "خطأ في الفاتورة - المورد خاطئ"
}
```

### المسارات المحدثة

```go
// Delete Purchase (الآن محدودة بحسب الحالة)
DELETE /api/v1/purchases/{id}
// فقط للحالات: draft, pending
// يعود خطأ للحالات: received, reversed, cancelled
```

## Frontend Changes

### 1. Updated UI Components

#### Purchase Status Badges
```typescript
const getStatusBadge = (status: string) => {
  const variants = {
    draft: { label: 'مسودة', variant: 'outline' },
    pending: { label: 'قيد الانتظار', variant: 'secondary' },
    received: { label: 'تم الاستلام', variant: 'default' },
    cancelled: { label: 'ملغي', variant: 'destructive' },
    reversed: { label: 'تم العكس', variant: 'destructive' },
    partially_received: { label: 'استلام جزئي', variant: 'secondary' },
  };
  return variants[status] || { label: status, variant: 'default' };
};
```

#### Action Buttons Logic
```typescript
{(purchase.status === 'pending' || purchase.status === 'draft') && (
  <>
    <Button onClick={() => handleDeletePurchase(purchase.id)}>
      <Trash2 /> حذف
    </Button>
  </>
)}
{purchase.status === 'received' && (
  <Button onClick={() => handleReversePurchase(purchase.id)}>
    <RotateCcw /> عكس العملية
  </Button>
)}
```

### 2. New API Functions

```typescript
// API Endpoint
reverse: (id: string, reason: string) =>
  apiClient.post(`/purchases/${id}/reverse`, { reason })

// Hook Function
const reversePurchaseMutation = useMutation({
  mutationFn: ({ purchaseId, reason }) =>
    purchasesApi.reverse(purchaseId, reason),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['purchases'] });
    queryClient.invalidateQueries({ queryKey: ['inventory'] });
  },
  onError: (error) => {
    if (error.message?.includes('some items have been sold')) {
      alert('لا يمكن عكس عملية الشراء لأن بعض القطع تم بيعها بالفعل.');
    }
  },
});
```

### 3. Updated Statistics

```typescript
const stats: PurchaseStats = {
  totalPurchases: purchases.length,
  pendingCount: purchases.filter((p) => p.status === 'pending' || p.status === 'draft').length,
  receivedCount: purchases.filter((p) => p.status === 'received' || p.status === 'partially_received').length,
  reversedCount: purchases.filter((p) => p.status === 'reversed').length,
  totalCost: purchases.filter((p) => p.status !== 'cancelled' && p.status !== 'reversed')
    .reduce((sum, p) => sum + (p.total_amount || 0), 0),
};
```

## الحماية والمصادقة

### منع الحذف غير المقصود

```go
// 1. فحص حالة الشراء
if purchase.Status == StatusReceived {
    return ErrCannotDeleteReceivedPurchase
}

// 2. فحص إذا تم بيع أي قطعة
soldItemsCheck := `
    SELECT COUNT(*)
    FROM inventory_items
    WHERE item_code LIKE $1 AND status = 'SOLD'
`
if soldCount > 0 {
    return ErrItemsAlreadySold
}
```

### Audit Trail كامل

```go
// Audit Log لكل عملية عكس
changes := fmt.Sprintf("Reversed purchase %s, reason: %s, items affected: %d",
    purchase.InvoiceNumber, reason, len(items))
_, err = tx.ExecContext(ctx, auditQuery,
    uuid.New(), userID, "REVERSE_PURCHASE", "purchase", purchase.ID,
    changes, time.Now())
```

## مثال عملي

### السيناريو قبل التحسين

```
شراء #1042 (تم الاستلام)
├── 10 قطع GPU في المخزون
└── المستخدم يضغط "حذف" بالخطأ

النتيجة:
❌ الشراء يختفي من قاعدة البيانات
❌ 10 قطع GPU تختفي من المخزون
❌ لا توجد سجلات للعملية
❌ التقارير والأرباح تفسد
```

### السيناريو بعد التحسين

```
شراء #1042 (تم الاستلام)
├── 10 قطع GPU في المخزون
└── المستخدم يضغط "حذف"

النتيجة:
✅ رسالة خطأ: "لا يمكن حذف عملية الشراء المستلمة. استخدم 'عكس العملية' بدلاً من الحذف."
✅ المستخدم يضغط "عكس العملية"
✅ يُطلب سبب العكس
✅ الشراء يصبح status = 'reversed'
✅ 10 قطع GPU تصبح status = 'RETURNED'
✅ حركات مخزون عكسية تُنشأ
✅ Audit log كامل يُسجل
✅ التقارير تبقى صحيحة
```

## الفوائد

### 1. Audit Trail كامل
- كل عملية عكس مسجلة مع السبب والمستخدم والوقت
- يمكن تتبع التاريخ الكامل لكل عنصر مخزون

### 2. حماية البيانات
- منع الحذف العرضي للعمليات التجارية
- حماية التكامل المالي للمخزون

### 3. تقارير دقيقة
- التقارير تبقى صحيحة حتى بعد عكس العمليات
- الأرباح والتكاليف تُحسب بشكل صحيح

### 4. تجربة مستخدم أفضل
- رسائل خطأ واضحة بالعربية
- إرشادات واضحة للعمليات المسموحة
- فصل واضح بين "حذف" و "عكس"

## الأخطاء الجديدة

```go
// أخطاء الحالات المحظورة
ErrPurchaseAlreadyReversed       = "purchase is already reversed"
ErrCannotDeleteReceivedPurchase  = "cannot delete received purchase - use reverse instead"
ErrItemsAlreadySold              = "cannot reverse purchase - some items have been sold"
ErrInvalidStatusTransition       = "invalid status transition"
```

## التوافق والترقية

### التوافق الخلفي
- الحالات القديمة (`pending`, `received`, `cancelled`) ما زالت مدعومة
- الـ legacy aliases مضافة للتوافق

### تحديث قاعدة البيانات
لا تتطلب تغييرات schema - الحالات الجديدة مدعومة في الحقل الموجود `status`.

## الاختبار

### Backend
```bash
cd backend
go build -o /tmp/partflow-test ./cmd/api/main.go
# ✅ Build ناجح
```

### Frontend
```bash
cd frontend
npm run build
# ✅ Build ناجح
```

## الخلاصة

تم تحويل PartFlow من نظام CRUD بسيط إلى نظام تجاري احترافي مع:

✅ **Status-Based Logic**: منطق قائم على الحالات بدلاً من الحذف الكامل
✅ **Reverse Operations**: عمليات عكس محمية بدلاً من الحذف
✅ **Audit Trail**: سجل تتبع كامل لكل العمليات
✅ **Status Validation**: التحقق من الانتقالات المسموحة بين الحالات
✅ **Inventory Protection**: حماية المخزون من الحذف غير المقصود
✅ **User Guidance**: إرشادات واضحة للمستخدمين بالعربية

هذا التطبيق يجعل PartFlow أقرب بكثير إلى أنظمة SaaS التجارية الاحترافية، مع الحفاظ على البساطة وسهولة الاستخدام.

---

**التاريخ**: 2026-08-26
**الإصدار**: v1.4.0
**المطور**: Devin AI Assistant