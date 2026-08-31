# PartFlow Offline-Online Repository Architecture

## نظرة عامة

تم بناء نظام إدارة المستودعات (Repository Pattern) الذي يدعم كل من العمل بدون إنترنت (Offline) والعمل مع السحابة (Online) بدون الحاجة لتغيير الأكواد.

## البنية المعمارية

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP Handlers                        │
│              (Products, Sales, Customers, etc)          │
└────────────────┬────────────────────────────────────────┘
                 │
                 ↓
┌─────────────────────────────────────────────────────────┐
│         RepositoryFactoryMiddleware                     │
│   • قراءة operating_mode من local_metadata             │
│   • إنشاء Repository Factory                           │
│   • حقن المصنع في Gin Context                         │
└────────────────┬────────────────────────────────────────┘
                 │
                 ↓
┌─────────────────────────────────────────────────────────┐
│         RepositoryFactory (نمط المصنع)                 │
│   GetProductRepository()                               │
│   GetCustomerRepository()                              │
│   GetSupplierRepository()                              │
│   GetInventoryRepository()                             │
│   GetSalesRepository()                                 │
└────┬──────────────────────────────────────┬─────────────┘
     │                                      │
     ↓ (offline)                    ↓ (online)
┌──────────────────┐      ┌──────────────────┐
│  SQLite Repos    │      │ PostgreSQL Repos │
│  • Products      │      │  • Products      │
│  • Customers     │      │  • Customers     │
│  • Suppliers     │      │  • Suppliers     │
│  • Inventory     │      │  • Inventory     │
│  • Sales         │      │  • Sales         │
└──────────────────┘      └──────────────────┘
     ↓                             ↓
┌──────────────────┐      ┌──────────────────┐
│  SQLite Database │      │  PostgreSQL DB   │
│  ~/.config/      │      │  Cloud Server    │
│  PartFlow/       │      │                  │
│  partflow.db     │      │                  │
└──────────────────┘      └──────────────────┘
```

## المكونات الرئيسية

### 1. Repository Interfaces (واجهات المستودعات)
**الموقع**: `backend/internal/repositories/interfaces.go`

تعريف الواجهات البرمجية الموحدة:
```go
type ProductRepository interface {
    Create(ctx context.Context, product interface{}) (string, error)
    GetByID(ctx context.Context, id string) (interface{}, error)
    GetBySKU(ctx context.Context, sku string) (interface{}, error)
    List(ctx context.Context, limit, offset int) ([]interface{}, error)
    Update(ctx context.Context, id string, product interface{}) error
    Delete(ctx context.Context, id string) error
    Search(ctx context.Context, query string, limit, offset int) ([]interface{}, error)
    Count(ctx context.Context) (int, error)
}
```

### 2. SQLite Implementations (تطبيقات SQLite)
**الموقع**: `backend/internal/repositories/sqlite/`

الملفات:
- `products.go` - إدارة المنتجات
- `customers.go` - إدارة العملاء
- `suppliers.go` - إدارة الموردين
- `inventory.go` - إدارة المخزون
- `sales.go` - إدارة المبيعات

كل ملف يحتوي على:
- هيكل البيانات (Struct)
- التطبيق الكامل للواجهة

### 3. Repository Factory (مصنع المستودعات)
**الموقع**: `backend/internal/repositories/factory.go`

```go
type RepositoryFactoryImpl struct {
    postgresDB *sqlx.DB
    sqliteDB   *sql.DB
    mode       string // "offline" أو "online"
}
```

وظائف رئيسية:
- `NewRepositoryFactory()` - إنشاء مصنع جديد
- `GetProductRepository()` - الحصول على Repository المناسب
- `GetCustomerRepository()`
- `GetSupplierRepository()`
- `GetInventoryRepository()`
- `GetSalesRepository()`
- `SetOperatingMode()` - تغيير وضع التشغيل
- `GetOperatingMode()` - الحصول على وضع التشغيل الحالي

### 4. Repository Factory Middleware (وسيط المصنع)
**الموقع**: `backend/internal/middleware/repository_factory.go`

وظائف:
- قراءة `operating_mode` من جدول `local_metadata` في SQLite
- إنشاء `RepositoryFactory` بالوضع المناسب
- حقن المصنع في Gin Context

استخدام:
```go
// في main.go
router.Use(middleware.RepositoryFactoryMiddleware(postgresDB, sqliteDB))
```

### 5. SQLite Schema (مخطط قاعدة البيانات)
**الموقع**: `backend/internal/localdb/localdb.go`

الجداول المُنشأة تلقائياً:
- `categories` - الفئات
- `products` - المنتجات
- `customers` - العملاء
- `suppliers` - الموردين
- `inventory_items` - أصناف المخزون
- `sales` - المبيعات
- `sale_items` - تفاصيل المبيعات
- `purchases` - المشتريات
- `purchase_items` - تفاصيل المشتريات
- `payments` - الدفعات
- `debts` - الديون
- `local_metadata` - البيانات الوصفية المحلية (تخزين operating_mode)
- `sync_queue` - قائمة المزامنة (للعمليات المعلقة)

## كيفية الاستخدام

### من جانب المعالج (Handler)

```go
func (h *ProductHandler) ListProducts(c *gin.Context) {
    factory := middleware.GetRepositoryFactory(c)
    repo := factory.GetProductRepository()
    
    products, err := repo.List(c.Request.Context(), 10, 0)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, products)
}
```

النظام يختار تلقائياً:
- **SQLite** إذا كان `operating_mode = "offline"`
- **PostgreSQL** إذا كان `operating_mode = "online"` (في المستقبل)

### تغيير وضع التشغيل

```go
// في handler تغيير الإعدادات
mode := "offline" // أو "online"
localdb.SetMetadata(sqliteDB, "operating_mode", mode)

// إعادة تحميل المصنع (سيحدث تلقائياً عند الطلب التالي)
```

## عملية المزامنة (Sync)

عندما يُرغب الانتقال من **Offline** إلى **Online**:

1. قراءة جميع البيانات من SQLite
2. إرسال العمليات إلى PostgreSQL
3. تسجيل كل عملية في `sync_queue` مع `idempotency_key`
4. تطبيق تتبع `synced_at` لتجنب التكرار
5. إعادة المحاولة التلقائية للعمليات الفاشلة

## جدول الحالات

### Operating Mode Transitions

```
┌──────────┐         ┌──────────┐
│ OFFLINE  │────────▶│  ONLINE  │
│(SQLite)  │ (sync)  │(PostgreSQL)
└──────────┘         └──────────┘
     ▲                    │
     │                    │
     └────────────────────┘
```

- **الانتقال من Online إلى Offline**: تنزيل البيانات من PostgreSQL إلى SQLite
- **الانتقال من Offline إلى Online**: مزامنة البيانات من SQLite إلى PostgreSQL

## معالجة الأخطاء

كل Repository implementation يعالج:
- `sql.ErrNoRows` - إرجاع nil (لا يوجد بيانات)
- Constraints violations - إرجاع خطأ واضح
- Connection errors - إرجاع خطأ مغلّف مع context

## الأداء والفهرسة

SQLite Indexes:
- `idx_products_category` - للبحث السريع عن المنتجات حسب الفئة
- `idx_products_sku` - للبحث السريع عن المنتجات حسب SKU
- `idx_inventory_product` - للبحث السريع عن المخزون
- `idx_inventory_barcode` - للمسح الضوئي السريع
- `idx_sales_customer` - للبحث السريع عن مبيعات العميل
- وغيرها...

## المميزات

✅ **التبديل السلس**: بدون إعادة تشغيل  
✅ **عزل البيانات**: كل وضع تشغيل له قاعدة بيانات مستقلة  
✅ **معايرة الموضوعية**: توحيد العمليات عبر كلا الوضعين  
✅ **سهولة التوسع**: إضافة Repository جديد سهل جداً  
✅ **الاختبار**: يمكن اختبار كلا الوضعين بشكل مستقل  

## الخطوات التالية

1. ✅ إنشاء Repository Interfaces
2. ✅ تطبيق SQLite Repositories
3. ✅ بناء Repository Factory
4. ✅ إنشاء Middleware
5. ⏳ تحديث Handlers لاستخدام الـ Factory
6. ⏳ تطبيق PostgreSQL Repositories
7. ⏳ إضافة Sync mechanism
8. ⏳ اختبار شامل للمزامنة
