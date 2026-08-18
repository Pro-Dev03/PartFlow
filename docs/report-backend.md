على الرحب والسعة. بناءً على `report.md` الذي أرسلته، سأعتبره **المرجع الوظيفي الأعلى**، وأحوّل متطلباته إلى تقرير Backend تنفيذي شامل، بحيث يستطيع مطور Go أو AI Coding Agent استخدامه كخطة بناء فعلية.

# PartFlow

## Backend Architecture & Implementation Specification

### نظام الإدارة الذكي لمحل قطع الحاسوب

---

# 1. مقدمة

PartFlow ليس نظام POS تقليديًا، وليس مجرد نظام لإدارة المخزون.

الهدف من الـBackend هو بناء **المحرك التشغيلي للمحل** الذي يربط جميع العمليات ببعضها:

```text
Products
    ↓
Inventory
    ↓
Purchasing
    ↓
Inspection
    ↓
Storage
    ↓
Sales
    ↓
Payments
    ↓
Customer Debt
    ↓
Returns
    ↓
Warranty
    ↓
Expenses
    ↓
Profit
    ↓
Reports
    ↓
Insights
    ↓
Automation
```

المبدأ الأساسي:

> **المستخدم يدخل المعلومة مرة واحدة، والـBackend يتولى بقية العمليات المرتبطة بها تلقائيًا.**

---

# 2. الهدف من الـBackend

يجب أن يحقق الـBackend خمسة أهداف رئيسية:

### 1. إدارة البيانات

حفظ جميع بيانات المحل بشكل منظم وآمن.

### 2. تنفيذ Business Logic

تطبيق قواعد العمل الفعلية للمحل.

### 3. الأتمتة

تنفيذ العمليات المترابطة تلقائيًا.

### 4. الحماية

ضمان أن كل مستخدم يصل فقط إلى البيانات والعمليات المسموح بها.

### 5. الذكاء

تجهيز البيانات لإنتاج:

* تقارير.
* تنبيهات.
* Insights.
* توصيات.
* AI Assistant مستقبلًا.

---

# 3. المعمارية العامة

المعمارية المقترحة:

```text
                   React / PWA
                       │
                       │ HTTPS
                       ▼
                ┌──────────────┐
                │   Go API     │
                └──────┬───────┘
                       │
        ┌──────────────┼──────────────┐
        │              │              │
        ▼              ▼              ▼
   Business        Security       Background
    Modules          Layer           Worker
        │              │              │
        └──────────────┼──────────────┘
                       │
                       ▼
                PostgreSQL
                 / Supabase
                       │
              ┌────────┴────────┐
              │                 │
              ▼                 ▼
         Database           Storage
```

---

# 4. التقنية

## Backend

```text
Go
```

## Database

```text
PostgreSQL
```

ويفضل استخدام:

```text
Supabase PostgreSQL
```

## Authentication

```text
Supabase Auth
```

مع قيام Go Backend بتطبيق Authorization وBusiness Rules.

## Storage

```text
Supabase Storage
```

للصور والمرفقات.

## Worker

```text
Go Worker
```

للمهام الخلفية.

## Deployment

```text
Docker
Render
```

---

# 5. لماذا Modular Monolith؟

لا يوصى باستخدام Microservices في النسخة الأولى.

البنية المناسبة:

```text
PartFlow Backend
│
├── Auth
├── Organizations
├── Users
├── Products
├── Inventory
├── Customers
├── Sales
├── Payments
├── Debts
├── Suppliers
├── Purchases
├── Expenses
├── Returns
├── Warranty
├── Inspections
├── Reports
├── Notifications
├── Audit
└── Insights
```

كل Module مستقل منطقيًا، لكن جميعها تعمل داخل تطبيق Go واحد.

هذا يوفر:

* سرعة التطوير.
* سهولة الاختبار.
* Transactions أبسط.
* Deployment أسهل.
* تكلفة أقل.
* إمكانية فصل Modules لاحقًا إذا احتاج المشروع.

---

# 6. هيكل المشروع

البنية المقترحة:

```text
backend/
│
├── cmd/
│   ├── api/
│   │   └── main.go
│   │
│   └── worker/
│       └── main.go
│
├── internal/
│   │
│   ├── auth/
│   ├── organizations/
│   ├── users/
│   ├── roles/
│   ├── permissions/
│   │
│   ├── products/
│   ├── categories/
│   ├── brands/
│   ├── barcodes/
│   │
│   ├── inventory/
│   ├── locations/
│   ├── inspections/
│   │
│   ├── customers/
│   ├── customer_ledger/
│   ├── payments/
│   ├── debts/
│   │
│   ├── suppliers/
│   ├── supplier_ledger/
│   ├── purchases/
│   │
│   ├── sales/
│   ├── returns/
│   ├── warranties/
│   ├── expenses/
│   │
│   ├── dashboard/
│   ├── reports/
│   ├── notifications/
│   ├── audit/
│   ├── search/
│   └── insights/
│
├── pkg/
│   ├── database/
│   ├── http/
│   ├── middleware/
│   ├── validation/
│   ├── errors/
│   ├── logger/
│   ├── pagination/
│   ├── money/
│   └── response/
│
├── migrations/
│
├── tests/
│
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

---

# 7. طبقات النظام

يجب فصل المسؤوليات.

```text
HTTP Handler
      ↓
Validation
      ↓
Authorization
      ↓
Service
      ↓
Repository
      ↓
Database
```

## Handler

مسؤول عن HTTP فقط.

## Service

يحتوي Business Logic.

## Repository

يتعامل مع قاعدة البيانات.

## Database

تفرض:

* Foreign Keys.
* Unique Constraints.
* Check Constraints.
* Transactions.
* Indexes.

---

# 8. Multi-Tenant Architecture

يجب بناء النظام منذ البداية ليكون SaaS-ready.

المفهوم الأساسي:

```text
Organization
```

كل محل يمثل Organization.

مثال:

```text
Organization A
├── Users
├── Products
├── Inventory
├── Customers
└── Sales
```

و:

```text
Organization B
├── Users
├── Products
├── Inventory
├── Customers
└── Sales
```

لا يمكن لأي Organization الوصول إلى بيانات Organization أخرى.

---

# 9. organization_id

كل جدول Business مهم يجب أن يحتوي على:

```text
organization_id
```

مثل:

```text
products
inventory_items
customers
sales
suppliers
purchases
expenses
```

ويجب أن يكون هذا جزءًا من جميع Queries.

---

# 10. Row Level Security

إذا تم استخدام Supabase، يجب تفعيل RLS.

لكن لا يجب الاعتماد على RLS وحده.

الحماية تكون:

```text
Authentication
      ↓
Organization
      ↓
Role
      ↓
Permission
      ↓
Business Rule
      ↓
RLS
      ↓
Database
```

---

# 11. Authentication

النظام يحتاج:

```text
Login
Logout
Session
Password Reset
Current User
```

Backend يجب أن يعرف:

```text
User ID
Organization ID
Role
Permissions
```

---

# 12. Authorization

لا يكفي:

```text
role = admin
```

يجب وجود Permissions.

مثال:

```text
products.read
products.create
products.update
products.archive

inventory.read
inventory.adjust
inventory.transfer

sales.read
sales.create
sales.cancel
sales.refund

customers.read
customers.create
customers.update

debts.read
debts.payment

purchases.read
purchases.create
purchases.receive

reports.read
reports.export

expenses.read
expenses.create

users.manage
settings.manage
audit.read
```

---

# 13. Roles

## Owner

صلاحيات كاملة.

## Manager

إدارة العمليات الأساسية والتقارير.

## Employee

البيع والبحث والمخزون المسموح.

## Accountant

الدفعات والمصروفات والتقارير المالية.

ويجب أن تكون الصلاحيات قابلة للتخصيص لاحقًا.

---

# 14. نموذج البيانات الأساسي

الجداول الأساسية:

```text
organizations

users
roles
permissions
user_roles
role_permissions

products
categories
brands
barcodes

inventory_items
inventory_movements
locations

customers
customer_ledger
customer_payments

suppliers
supplier_ledger
supplier_payments

sales
sale_items
payments

purchases
purchase_items

expenses
expense_categories

inspections
inspection_items

returns
return_items

warranties
warranty_claims

reservations

notifications
audit_logs

attachments
```

---

# 15. Product وInventory Item

يجب الفصل بين:

```text
Product
```

و:

```text
Inventory Item
```

## Product

يمثل النوع أو الموديل.

مثال:

```text
RTX 3060 ASUS
```

## Inventory Item

يمثل القطعة الفعلية.

مثال:

```text
ITEM-000421
```

وهذه نقطة أساسية في PartFlow.

---

# 16. Product

يمكن أن يحتوي:

```text
id
organization_id
name
brand_id
category_id
model
sku
description
product_type
default_cost
default_price
minimum_stock
warranty_policy
created_at
updated_at
```

---

# 17. Product Types

## Quantity

مثل:

```text
Cable
Mouse
Keyboard
USB
Thermal Paste
```

يتم التعامل معها بالكمية.

## Individual

مثل:

```text
GPU
CPU
RAM
SSD
HDD
Laptop
Motherboard
PSU
```

كل قطعة يمكن أن تكون Inventory Item مستقلة.

---

# 18. Inventory Item

يمكن أن يحتوي:

```text
id
organization_id
product_id
item_code
barcode
serial_number
condition
grade
purchase_cost
selling_price
status
location_id
supplier_id
purchase_date
sold_at
created_at
updated_at
```

---

# 19. Condition

القيم المقترحة:

```text
NEW
USED
REFURBISHED
DAMAGED
FOR_PARTS
```

والـGrade:

```text
EXCELLENT
VERY_GOOD
GOOD
FAIR
POOR
```

---

# 20. Item Lifecycle

كل قطعة يجب أن تمتلك Lifecycle واضحًا:

```text
PURCHASED
    ↓
RECEIVED
    ↓
INSPECTION
    ↓
AVAILABLE
    ↓
RESERVED
    ↓
SOLD
```

ويمكن أن تنتقل إلى:

```text
DAMAGED
IN_REPAIR
RETURNED
WARRANTY
FOR_PARTS
ARCHIVED
```

---

# 21. Barcode System

يدعم النظام:

```text
External Barcode
Internal Barcode
SKU
Serial Number
Item Code
```

إذا لم يوجد Barcode:

```text
Generate Internal Barcode
```

مثال:

```text
PF-GPU-000001
PF-CPU-000002
PF-RAM-000003
```

---

# 22. Barcode Lookup

Endpoint أساسي:

```http
GET /api/v1/barcodes/{code}
```

يجب أن يعيد:

```text
Product
Item
Condition
Price
Status
Location
Warranty
```

ويجب أن يكون سريعًا جدًا لأنه يمثل واحدة من أهم عمليات النظام.

---

# 23. Context-Aware Barcode

الـBackend يجب أن يفهم Context العملية.

مثلاً:

```text
context=sale
```

يعني:

```text
Scan
→ Add to Cart
```

أما:

```text
context=inventory
```

فيعني:

```text
Scan
→ Open Item
```

وهذا يسمح للـFrontend ببناء تجربة Scan موحدة.

---

# 24. Inventory Engine

لا يجب الاعتماد فقط على:

```text
stock_quantity
```

بل يجب تسجيل Inventory Movements.

مثال:

```text
Purchase +10
Sale -2
Return +1
Damage -1
```

والرصيد النهائي:

```text
Current Stock = 8
```

---

# 25. Inventory Movement

كل حركة يجب أن تحتوي:

```text
id
organization_id
item_id / product_id
movement_type
quantity
before_quantity
after_quantity
reference_type
reference_id
reason
created_by
created_at
```

الأنواع:

```text
PURCHASE
SALE
RETURN
ADJUSTMENT
TRANSFER
RESERVATION
RELEASE
DAMAGE
REPAIR
```

---

# 26. Locations

دعم:

```text
Warehouse
Shelf
Box
Display
```

مثال:

```text
Warehouse A
└── Shelf B
    └── Box 04
```

ويجب أن يستطيع الموظف معرفة مكان القطعة من خلال Scan.

---

# 27. Reservations

يمكن حجز قطعة:

```text
AVAILABLE
   ↓
RESERVED
   ↓
SOLD
```

إذا انتهى وقت الحجز:

```text
RESERVED
   ↓
AVAILABLE
```

والـWorker مسؤول عن انتهاء الحجوزات.

---

# 28. Used Items

المنتجات المستعملة ليست مجرد:

```text
condition = USED
```

بل يمكن أن تحتوي على:

```text
Inspection
Grade
Purchase Cost
Repair Cost
Testing
Notes
Photos
Warranty
```

---

# 29. Inspection

يمكن إنشاء:

```text
inspection
```

يحتوي:

```text
item_id
inspected_by
status
notes
inspected_at
```

و:

```text
inspection_items
```

مثل:

```text
Power
Temperature
Performance
Ports
Visual
Serial Verification
```

---

# 30. Inspection Status

```text
PENDING
IN_PROGRESS
PASSED
PARTIAL
FAILED
```

ويجب حفظ تاريخ كل فحص ومن قام به.

---

# 31. تكلفة القطعة المستعملة

التكلفة الحقيقية يمكن أن تكون:

```text
Purchase Cost
+
Repair Cost
+
Testing Cost
+
Other Costs
```

مثال:

```text
Purchase = 900
Repair = 50
Testing = 20

True Cost = 970
```

إذا بيعت بـ:

```text
1,150
```

فالربح:

```text
180
```

وليس:

```text
250
```

---

# 32. Customers

Customer:

```text
id
organization_id
name
phone
email
notes
created_at
updated_at
```

ويرتبط بـ:

```text
Sales
Payments
Debt
Returns
Warranty
Ledger
```

---

# 33. Customer Ledger

يجب عدم تخزين الدين كرقم فقط.

بل يجب أن يكون هناك Ledger.

مثال:

```text
Sale +2500
Payment -1000
Payment -500
```

الرصيد:

```text
1000
```

وبذلك يمكن معرفة سبب كل رقم.

---

# 34. Debt Management

يجب دعم:

```text
Outstanding
Due Date
Overdue
Payment History
```

مثال:

```text
Customer:
Ahmed

Total:
5000 ₪

Paid:
2500 ₪

Outstanding:
2500 ₪
```

---

# 35. Debt Aging

تصنيف:

```text
CURRENT
DUE
OVERDUE
```

ويظهر:

```text
Customer
Amount
Due Date
Days Overdue
Last Payment
```

---

# 36. Payments

طرق الدفع:

```text
CASH
CARD
BANK_TRANSFER
DEBT
OTHER
```

ويجب دعم Split Payment.

مثال:

```text
Sale = 2000

Cash = 500
Card = 1000
Debt = 500
```

---

# 37. Financial Immutability

لا يجب حذف:

```text
Sale
Payment
Refund
Debt
Return
```

إذا كانت هناك عملية خاطئة:

```text
Reverse
```

بدل:

```text
Delete
```

---

# 38. Sales Engine

عملية البيع يجب أن تكون Atomic Transaction.

```text
BEGIN
 ↓
Validate Customer
 ↓
Validate Items
 ↓
Lock Inventory
 ↓
Calculate Totals
 ↓
Create Sale
 ↓
Create Sale Items
 ↓
Create Payments
 ↓
Update Inventory
 ↓
Update Customer Ledger
 ↓
Create Audit
 ↓
COMMIT
```

إذا فشل شيء:

```text
ROLLBACK
```

---

# 39. Concurrency

يجب منع بيع نفس القطعة مرتين.

سيناريو:

```text
Employee A → Scan GPU
Employee B → Scan GPU
```

في نفس الوقت.

يجب استخدام:

```text
Transaction
+
Row Lock
+
Status Validation
```

---

# 40. Idempotency

إذا أرسل Frontend نفس طلب البيع مرتين بسبب:

* Double Click.
* Slow Network.
* Retry.

لا يجب إنشاء عمليتي بيع.

العمليات الحساسة يجب أن تدعم:

```text
Idempotency-Key
```

---

# 41. Sale Model

```text
sales

id
organization_id
customer_id
subtotal
discount
tax
total
paid_amount
debt_amount
status
created_by
created_at
```

---

# 42. Sale Items

```text
sale_items

sale_id
product_id
inventory_item_id
quantity
unit_price
unit_cost
discount
total
```

يجب حفظ `unit_cost` لحظة البيع حتى لا تتغير الأرباح التاريخية إذا تغيرت تكلفة المنتج لاحقًا.

---

# 43. Profit Engine

يجب فصل:

```text
Revenue
COGS
Gross Profit
Expenses
Net Profit
```

المعادلة:

```text
Revenue - COGS = Gross Profit

Gross Profit - Expenses = Net Profit
```

---

# 44. Purchases

Workflow:

```text
Create Purchase
 ↓
Select Supplier
 ↓
Add Items
 ↓
Enter Cost
 ↓
Receive
 ↓
Inventory Increase
 ↓
Supplier Ledger
 ↓
Audit
```

عملية الاستلام يجب أن تكون Transaction.

---

# 45. Suppliers

Supplier:

```text
id
organization_id
name
phone
email
notes
```

ويرتبط بـ:

```text
Purchases
Payments
Ledger
Products
```

---

# 46. Supplier Ledger

مثال:

```text
Purchase +5000
Payment -2000
Payment -1000
```

الرصيد:

```text
2000
```

وهذا يتيح معرفة ما يجب دفعه لكل مورد.

---

# 47. Expenses

دعم:

```text
Rent
Electricity
Internet
Salary
Shipping
Maintenance
Equipment
Other
```

كل Expense يحتوي:

```text
organization_id
category_id
amount
description
date
created_by
```

---

# 48. Returns

Return ليست حذفًا للبيع.

Workflow:

```text
Sale
 ↓
Return Request
 ↓
Inspection
 ↓
Approve / Reject
 ↓
Refund / Exchange
 ↓
Inventory Update
 ↓
Financial Adjustment
 ↓
Audit
```

---

# 49. Refund

يدعم:

```text
FULL
PARTIAL
```

ويجب ألا يسمح النظام برد مبلغ أكبر من المبلغ القابل للاسترجاع.

---

# 50. Warranty

كل Sale يمكن أن تنتج Warranty.

مثال:

```text
Warranty Start
Warranty End
Duration
Type
Status
```

---

# 51. Warranty Claim

```text
Customer
 ↓
Warranty Claim
 ↓
Inspection
 ↓
Repair / Replace / Reject
 ↓
Resolution
```

ويجب حفظ تاريخ كامل للعملية.

---

# 52. Dashboard Backend

لا يجب أن يضطر Frontend إلى إرسال 15 Request للحصول على Dashboard.

يفضل:

```http
GET /api/v1/dashboard
```

ويرجع:

```json
{
  "sales": {},
  "profit": {},
  "inventory": {},
  "debts": {},
  "low_stock": [],
  "alerts": [],
  "top_products": [],
  "insights": []
}
```

---

# 53. Dashboard Philosophy

Dashboard لا يجب أن يكون مجرد Charts.

السؤال الأساسي:

> ماذا يحتاج صاحب المحل أن يعرف الآن؟

لذلك يجب أن يعرض:

```text
Today's Sales
Today's Profit
Inventory Value
Outstanding Debts
Low Stock
Overdue Debts
Warranty Alerts
Slow Moving Items
Important Insights
```

---

# 54. Reports

الـBackend يجب أن يدعم:

```text
Sales
Profit
Inventory
Customers
Debts
Purchases
Suppliers
Expenses
Returns
Warranty
Used Products
```

---

# 55. Business Questions

التقارير يجب أن تجيب:

```text
كم بعت اليوم؟
كم ربحت؟
ما أكثر منتج مبيعًا؟
ما أكثر منتج ربحًا؟
من عليه أكبر دين؟
ما قيمة المخزون؟
ما المنتجات الراكدة؟
كم دفعت للموردين؟
كم صرفت؟
كم ربحت من المستعمل؟
```

---

# 56. Global Search

Endpoint:

```http
GET /api/v1/search?q=3060
```

يبحث في:

```text
Products
Items
Barcode
Serial
SKU
Customers
Phone
Sales
Suppliers
```

ويجب استخدام PostgreSQL Indexes قبل إدخال Search Engine خارجي.

---

# 57. Notifications

أنواع التنبيهات:

```text
LOW_STOCK
OVERDUE_DEBT
WARRANTY_EXPIRING
INSPECTION_REQUIRED
RESERVATION_EXPIRING
PAYMENT_RECEIVED
PURCHASE_RECEIVED
```

لكن النظام يجب ألا يغرق المستخدم بالتنبيهات.

الهدف:

> عرض الأشياء التي تحتاج إلى انتباه.

---

# 58. Smart Insights

الـBackend يجب أن يستطيع استخراج مؤشرات مثل:

```text
Low Stock
Slow Moving Products
Top Sellers
Top Profit Products
Overdue Customers
Used Product Performance
Sales Trends
Inventory Trends
```

مثال:

```text
"8 منتجات لم تتحرك منذ 90 يومًا."
```

---

# 59. Automation Worker

خدمة Worker منفصلة:

```text
backend
worker
```

المهام:

```text
Reservation Expiration
Debt Scan
Warranty Scan
Low Stock Scan
Daily Insights
Notifications
Large Reports
Image Processing
```

---

# 60. AI Ready

AI لا يجب أن يكون مسؤولًا عن العمليات المالية الأساسية.

البنية:

```text
Database
 ↓
Analytics / Business Services
 ↓
AI Layer
```

مثال:

```text
User:
كم ربحت من القطع المستعملة هذا الشهر؟
```

AI لا يخمن الرقم.

بل:

```text
Question
 ↓
Intent
 ↓
Permission
 ↓
Safe Business Query
 ↓
Database
 ↓
Verified Result
 ↓
AI Explanation
```

---

# 61. AI لا يملك صلاحية SQL حرة

يجب عدم السماح لـAI بإنشاء وتنفيذ SQL غير مقيد على قاعدة البيانات.

بدل ذلك:

```text
AI
 ↓
Approved Tool
 ↓
Business Service
 ↓
Database
```

وهذا يحافظ على:

* الأمان.
* الصلاحيات.
* صحة البيانات.

---

# 62. Audit Log

كل عملية حساسة تسجل:

```text
Who
What
When
Target
Before
After
Result
```

أمثلة:

```text
CREATE_PRODUCT
UPDATE_PRODUCT
CHANGE_PRICE

CREATE_SALE
CANCEL_SALE

CREATE_PAYMENT
REVERSE_PAYMENT

CREATE_PURCHASE
RECEIVE_PURCHASE

CREATE_RETURN
APPROVE_RETURN

CREATE_EXPENSE

ADJUST_INVENTORY

CHANGE_PERMISSION
```

---

# 63. Soft Delete

يمكن استخدام:

```text
deleted_at
```

لـ:

```text
Products
Customers
Suppliers
```

لكن لا يتم حذف السجلات المالية الأساسية.

---

# 64. Database Integrity

قاعدة البيانات نفسها يجب أن تمنع البيانات الخاطئة.

استخدام:

```text
Foreign Keys
Unique Constraints
Check Constraints
NOT NULL
Indexes
Transactions
```

---

# 65. Money

لا تستخدم:

```text
float64
```

للأموال.

استخدم:

```text
Decimal
```

أو Integer Minor Units.

العملة الأساسية:

```text
ILS
```

مع إمكانية دعم عملات أخرى مستقبلًا.

---

# 66. Timezone

يجب تخزين التواريخ بطريقة موحدة.

ويتم عرضها حسب Timezone المؤسسة.

بالنسبة للسوق الإسرائيلي:

```text
Asia/Jerusalem
```

لكن لا يجب Hardcode الـTimezone داخل Business Logic.

---

# 67. Pagination

لا يتم إرسال آلاف السجلات.

القوائم تستخدم:

```text
limit
cursor
has_more
next_cursor
```

---

# 68. API Versioning

كل API:

```text
/api/v1/
```

مثال:

```text
/api/v1/products
/api/v1/inventory
/api/v1/sales
/api/v1/customers
/api/v1/purchases
/api/v1/reports
```

---

# 69. API — Products

```http
GET    /api/v1/products
POST   /api/v1/products
GET    /api/v1/products/{id}
PATCH  /api/v1/products/{id}
POST   /api/v1/products/{id}/archive
```

---

# 70. API — Inventory

```http
GET   /api/v1/inventory
GET   /api/v1/inventory/{id}
POST  /api/v1/inventory
PATCH /api/v1/inventory/{id}

POST /api/v1/inventory/{id}/adjust
POST /api/v1/inventory/{id}/transfer

GET /api/v1/inventory/{id}/history
```

---

# 71. API — Barcode

```http
GET  /api/v1/barcodes/{code}
POST /api/v1/barcodes/generate
POST /api/v1/barcodes/labels
```

---

# 72. API — Sales

```http
GET  /api/v1/sales
POST /api/v1/sales
GET  /api/v1/sales/{id}

POST /api/v1/sales/{id}/cancel
POST /api/v1/sales/{id}/refund
GET  /api/v1/sales/{id}/receipt
```

---

# 73. API — Customers

```http
GET   /api/v1/customers
POST  /api/v1/customers
GET   /api/v1/customers/{id}
PATCH /api/v1/customers/{id}

GET /api/v1/customers/{id}/ledger
GET /api/v1/customers/{id}/sales
```

---

# 74. API — Payments

```http
GET  /api/v1/payments
POST /api/v1/payments

POST /api/v1/payments/{id}/reverse
```

---

# 75. API — Suppliers

```http
GET   /api/v1/suppliers
POST  /api/v1/suppliers
GET   /api/v1/suppliers/{id}
PATCH /api/v1/suppliers/{id}

GET /api/v1/suppliers/{id}/ledger
```

---

# 76. API — Purchases

```http
GET  /api/v1/purchases
POST /api/v1/purchases
GET  /api/v1/purchases/{id}

POST /api/v1/purchases/{id}/receive
```

---

# 77. API — Returns

```http
GET  /api/v1/returns
POST /api/v1/returns
GET  /api/v1/returns/{id}

POST /api/v1/returns/{id}/approve
POST /api/v1/returns/{id}/reject
```

---

# 78. API — Warranty

```http
GET  /api/v1/warranties
POST /api/v1/warranties
GET  /api/v1/warranties/{id}

POST /api/v1/warranties/{id}/claims
```

---

# 79. API — Inspection

```http
POST  /api/v1/inspections
GET   /api/v1/inspections/{id}
PATCH /api/v1/inspections/{id}
```

---

# 80. API — Dashboard

```http
GET /api/v1/dashboard
```

---

# 81. API — Reports

```http
GET /api/v1/reports/sales
GET /api/v1/reports/profit
GET /api/v1/reports/inventory
GET /api/v1/reports/debts
GET /api/v1/reports/purchases
GET /api/v1/reports/expenses
GET /api/v1/reports/returns
GET /api/v1/reports/warranty
```

مع Filters:

```text
from
to
category
brand
supplier
customer
condition
status
```

---

# 82. API — Notifications

```http
GET  /api/v1/notifications
POST /api/v1/notifications/{id}/read
POST /api/v1/notifications/read-all
```

---

# 83. API — Search

```http
GET /api/v1/search?q=
```

---

# 84. Error Handling

كل API يستخدم Error Structure موحدة:

```json
{
  "error": {
    "code": "INSUFFICIENT_STOCK",
    "message": "The requested item is not available.",
    "request_id": "..."
  }
}
```

لا يجب إرسال Stack Traces للمستخدم.

---

# 85. Request ID

كل Request يحصل على:

```text
request_id
```

ويستخدم في:

```text
Logs
Errors
Audit
Debugging
```

---

# 86. Logging

يجب استخدام Structured Logging.

مثال:

```json
{
  "level": "error",
  "request_id": "...",
  "organization_id": "...",
  "user_id": "...",
  "operation": "create_sale",
  "error": "insufficient_stock"
}
```

ولا يتم تسجيل:

```text
Passwords
Tokens
Secrets
Sensitive payment data
```

---

# 87. Rate Limiting

يجب حماية:

```text
Login
Password Reset
Search
Barcode
Public APIs
File Upload
```

خصوصًا العمليات التي يمكن إساءة استخدامها.

---

# 88. File Storage

يمكن استخدام Supabase Storage لـ:

```text
Product Images
Item Images
Inspection Photos
Invoices
Documents
```

ويجب ربط الملفات بـOrganization.

---

# 89. File Upload Security

يجب التحقق من:

```text
MIME Type
File Size
Extension
Organization
```

ويفضل استخدام UUID بدل اسم الملف الأصلي.

---

# 90. Caching

لا يجب إدخال Redis بلا حاجة.

يمكن استخدام Cache مستقبلًا لـ:

```text
Categories
Brands
Settings
Dashboard Aggregates
Frequently Used Data
```

لكن العمليات المالية يجب ألا تعتمد على Cache قديم.

---

# 91. Offline/PWA Readiness

إذا كان الـFrontend يعمل كـPWA، يجب أن يكون الـBackend مصممًا مستقبلًا لدعم:

```text
Offline Queue
Sync
Retry
Idempotency
Conflict Resolution
```

لكن لا ينبغي اعتبار Offline الكامل متطلبًا للنسخة الأولى إذا لم تكن هناك حاجة تشغيلية له.

الأهم هو أن API تكون:

* Idempotent.
* قابلة لإعادة المحاولة.
* واضحة في حالات Conflict.

---

# 92. Smart Defaults

Backend يمكن أن يوفر:

```text
Default Category
Default Location
Default Supplier
Default Warranty
Last Used Price
Last Supplier
```

حتى يقل إدخال البيانات.

---

# 93. Onboarding

عند إنشاء Organization:

```text
Create Organization
 ↓
Store Name
 ↓
Currency
 ↓
Timezone
 ↓
Categories
 ↓
First Product
 ↓
First User
 ↓
Ready
```

لا ينبغي إجبار صاحب المحل على إعداد نظام ضخم قبل أول عملية بيع.

---

# 94. First Sale Principle

المسار المثالي:

```text
Login
 ↓
Scan
 ↓
Sell
```

يجب أن يكون الـBackend مصممًا لدعم هذا السيناريو بأقل عدد ممكن من الخطوات والطلبات.

---

# 95. أهم Business Workflows

## إضافة قطعة

```text
Scan
 ↓
Identify
 ↓
Condition
 ↓
Cost
 ↓
Price
 ↓
Location
 ↓
Inspection
 ↓
Available
```

---

## بيع قطعة

```text
Scan
 ↓
Product
 ↓
Customer
 ↓
Payment
 ↓
Confirm
 ↓
Inventory
 ↓
Ledger
 ↓
Profit
 ↓
Receipt
 ↓
Audit
```

---

## بيع بالدين

```text
Sale
 ↓
Partial Payment
 ↓
Customer Ledger
 ↓
Outstanding Debt
```

---

## دفع الدين

```text
Customer
 ↓
Outstanding Balance
 ↓
Payment
 ↓
Ledger
 ↓
New Balance
 ↓
Audit
```

---

## شراء

```text
Supplier
 ↓
Purchase
 ↓
Receive
 ↓
Inventory
 ↓
Supplier Ledger
```

---

## Return

```text
Sale
 ↓
Return
 ↓
Inspection
 ↓
Approve
 ↓
Refund
 ↓
Inventory
 ↓
Ledger
```

---

## Warranty

```text
Sale
 ↓
Warranty
 ↓
Claim
 ↓
Inspection
 ↓
Repair / Replace / Reject
```

---

# 96. مثال متكامل

تم شراء:

```text
RTX 3070 Used
```

بتكلفة:

```text
1200 ₪
```

ثم:

```text
Repair = 50
Testing = 20
```

إذن:

```text
True Cost = 1270 ₪
```

تم وضع سعر البيع:

```text
1550 ₪
```

ثم تم البيع:

```text
1550 ₪
```

الـBackend يحدث تلقائيًا:

```text
Inventory -1

Revenue +1550

COGS +1270

Gross Profit +280

Sale Created

Payment Created

Customer Ledger Updated

Warranty Created

Audit Event Created
```

إذا دفع العميل:

```text
1000 ₪
```

فإن:

```text
Paid = 1000
Outstanding = 550
```

ولا يحتاج الموظف إلى إدخال هذه المعلومات في عدة أماكن.

---

# 97. Database Transactions

العمليات التالية يجب أن تكون Transactions:

```text
Sale
Purchase Receive
Payment
Refund
Return
Inventory Adjustment
Inventory Transfer
Reservation
```

الهدف:

> إما أن تنجح العملية كاملة، أو لا يحدث أي تغيير جزئي.

---

# 98. Business Rules

يجب أن تكون جميع القواعد الحساسة داخل Backend.

أمثلة:

```text
لا يمكن بيع Item غير Available.

لا يمكن بيع نفس Individual Item مرتين.

لا يمكن Refund مبلغ أكبر من المبلغ القابل للاسترجاع.

لا يمكن تعديل عملية مالية مغلقة مباشرة.

لا يمكن حذف Payment.

لا يمكن لموظف تعديل COGS.

لا يمكن تعديل Inventory بدون تسجيل Movement.

لا يمكن الوصول إلى Organization أخرى.

لا يمكن إنشاء Sale بدون وجود Inventory فعلي للمنتج الفردي.

لا يمكن اعتماد Inspection من مستخدم غير مصرح.

```

---

# 99. Testing

يجب بناء:

## Unit Tests

لـ:

```text
Pricing
Profit
Debt
Validation
Permissions
Inventory Rules
```

## Integration Tests

لـ:

```text
Sales
Payments
Inventory
Purchases
Returns
Database Transactions
```

## API Tests

لـ:

```text
Authentication
Authorization
Validation
Responses
Errors
```

---

# 100. أهم اختبارات النظام

يجب اختبار:

```text
Two employees sell same item simultaneously.

Payment submitted twice.

Return submitted twice.

Unauthorized user attempts to access another organization.

Employee attempts to modify protected financial data.

Purchase receives same item twice.

Inventory adjustment with invalid quantity.

Refund greater than sale amount.

Expired reservation.

Expired warranty.
```

---

# 101. Performance Targets

أهداف التصميم:

```text
Barcode Lookup:
< 300ms

Normal API:
< 500ms

Dashboard:
< 2 seconds
```

هذه أهداف هندسية وليست ضمانات مطلقة.

---

# 102. Health Checks

يجب توفير:

```http
GET /health
GET /ready
```

مثال:

```json
{
  "status": "ok",
  "database": "ok",
  "version": "1.0.0"
}
```

---

# 103. Docker

يجب استخدام Multi-stage Build.

```text
Go Source
   ↓
Builder
   ↓
Compiled Binary
   ↓
Minimal Runtime
```

ويجب أن يكون Production Container:

* Minimal.
* Non-root.
* بدون Secrets داخل Image.
* بدون Development Tools.

---

# 104. Environment Variables

مثل:

```text
DATABASE_URL
SUPABASE_URL
SUPABASE_SECRET_KEY
JWT_SECRET
APP_ENV
APP_URL
```

لا يتم Commit للقيم السرية.

---

# 105. CI/CD

كل Pull Request:

```text
Lint
 ↓
Unit Tests
 ↓
Integration Tests
 ↓
Build
 ↓
Docker Build
 ↓
Security Scan
```

ثم Production:

```text
Build
 ↓
Deploy
 ↓
Health Check
```

---

# 106. Backup

يجب وجود:

```text
Automated Backups
Retention Policy
Restore Procedure
Restore Testing
```

ولا يكفي وجود Backup إذا لم يتم اختبار استعادته.

---

# 107. Auditability

كل رقم مهم يجب أن يكون قابلًا للتفسير.

إذا ظهر:

```text
Inventory Value:
185,400 ₪
```

يجب أن يستطيع النظام معرفة:

```text
Products
Items
Quantities
Costs
```

وإذا ظهر:

```text
Profit:
18,500 ₪
```

يمكن الوصول إلى:

```text
Revenue
COGS
Expenses
Net Profit
```

---

# 108. Localization

الـBackend يخزن Codes وليس نصوص الواجهة.

مثلاً:

```text
LOW_STOCK
OVERDUE_DEBT
SALE_COMPLETED
```

والـFrontend يستطيع عرضها:

```text
العربية
עברית
English
```

وهذا يسمح بدعم RTL/LTR.

---

# 109. التوسع المستقبلي

يجب أن تسمح البنية مستقبلًا بـ:

```text
Multiple Branches
Multiple Warehouses
Multiple POS Devices
Online Store
Mobile App
Customer Portal
Shipping Integration
Accounting Integration
WhatsApp Notifications
AI Assistant
Advanced Analytics
```

بدون إعادة بناء Core Architecture.

---

# 110. مراحل التنفيذ

## Phase 1 — Foundation

```text
Go Project
Database
Migrations
HTTP
Config
Logging
Errors
Docker
Health
```

## Phase 2 — Identity

```text
Authentication
Organizations
Users
Roles
Permissions
RLS
```

## Phase 3 — Product Core

```text
Categories
Brands
Products
Barcodes
Inventory Items
Locations
```

## Phase 4 — Inventory

```text
Movements
Stock
Transfers
Adjustments
Reservations
Barcode Lookup
```

## Phase 5 — Customers & Finance

```text
Customers
Ledger
Payments
Debts
Expenses
```

## Phase 6 — Sales

```text
Sales
Sale Items
Checkout
Payments
Inventory
Profit
Receipt
Audit
```

## Phase 7 — Suppliers

```text
Suppliers
Purchases
Receiving
Supplier Ledger
```

## Phase 8 — Used Products

```text
Inspection
Grades
Repair Costs
Warranty
```

## Phase 9 — Returns

```text
Returns
Refunds
Exchanges
Warranty Claims
```

## Phase 10 — Intelligence

```text
Dashboard
Reports
Notifications
Insights
Automation
```

## Phase 11 — AI

```text
Business Assistant
Natural Language Queries
Recommendations
Forecasting
```

---

# 111. الأولويات

## P0 — يجب أن يعمل

```text
Authentication
Organizations
Products
Inventory
Barcode
Customers
Sales
Payments
Debts
Audit
```

## P1 — العمليات المتقدمة

```text
Purchases
Suppliers
Expenses
Returns
Warranty
Inspection
Reports
Dashboard
```

## P2 — الذكاء

```text
Notifications
Automation
Insights
Advanced Search
```

## P3 — المستقبل

```text
AI
Forecasting
Multi-Branch
External Integrations
Advanced BI
```

---

# 112. Definition of Done

لا يعتبر Module مكتملًا بمجرد كتابة Endpoint.

يجب أن يحتوي:

```text
Database Schema
Migration
Model
Repository
Service
Handler
Routes
Validation
Authorization
Business Rules
Transactions
Tests
Audit
Logging
Error Handling
```

---

# 113. معايير نجاح Backend

يجب أن ينجح السيناريو التالي بالكامل:

```text
Create Product
 ↓
Create Item
 ↓
Generate Barcode
 ↓
Assign Location
 ↓
Inspect
 ↓
Make Available
 ↓
Scan
 ↓
Sell
 ↓
Receive Payment
 ↓
Update Inventory
 ↓
Update Ledger
 ↓
Calculate Profit
 ↓
Generate Receipt
 ↓
Create Audit
 ↓
Update Dashboard
```

بدون إدخال نفس المعلومة أكثر من مرة.

---

# 114. المبدأ الذي يجب أن يحكم كل API

كل Endpoint يجب أن يسأل:

> ما الذي يجب أن يحدث تلقائيًا نتيجة هذه العملية؟

مثال:

```text
POST /sales
```

ليس مجرد:

```text
INSERT sale
```

بل:

```text
Validate
→ Lock
→ Calculate
→ Create Sale
→ Create Items
→ Create Payment
→ Update Inventory
→ Update Ledger
→ Update Financial Data
→ Create Warranty
→ Audit
→ Commit
```

وهذا هو الفرق بين Backend حقيقي لنظام PartFlow وبين CRUD Backend.

---

# 115. ما يجب تجنبه

لا يجب:

```text
وضع Business Logic في React.

حساب الأرباح في Frontend.

تعديل المخزون مباشرة بدون Movement.

حذف العمليات المالية.

استخدام float للأموال.

السماح ببيع Item مرتين.

الاعتماد على Role فقط.

الاعتماد على Frontend Security.

السماح لـAI بتنفيذ SQL حر.

إضافة Microservices مبكرًا.

إضافة Elasticsearch دون حاجة.

إرسال آلاف السجلات للواجهة.

جعل Dashboard يعتمد على عشرات Requests.

إجبار المستخدم على إدخال نفس البيانات عدة مرات.
```

---

# 116. النتيجة المعمارية النهائية

```text
                         USER
                          │
                          ▼
                   React / PWA
                          │
                         HTTPS
                          │
                          ▼
                  ┌───────────────┐
                  │    Go API     │
                  └───────┬───────┘
                          │
       ┌──────────────────┼──────────────────┐
       │                  │                  │
       ▼                  ▼                  ▼
   Products           Inventory           Sales
   Customers          Purchases           Payments
   Suppliers          Inspection          Debts
   Warranty           Returns             Expenses
       │                  │                  │
       └──────────────────┼──────────────────┘
                          │
                          ▼
                  Business Services
                          │
              ┌───────────┼───────────┐
              │           │           │
              ▼           ▼           ▼
         PostgreSQL    Storage      Worker
              │                       │
              ▼                       ▼
          RLS/Audit             Automation
              │                       │
              └───────────┬───────────┘
                          ▼
                    Insights / AI
```

---

# 117. التعريف النهائي للـBackend

PartFlow Backend يجب ألا يُبنى على أساس:

> "ما هي الصفحات الموجودة في Frontend؟"

بل على أساس:

> **"ما هي العمليات التي تحدث داخل المحل، وما الذي يجب أن يفعله النظام تلقائيًا عندما تحدث؟"**

ولهذا تكون النواة:

```text
Identity
+
Products
+
Individual Items
+
Inventory
+
Barcode
+
Purchasing
+
Inspection
+
Sales
+
Payments
+
Customer Ledger
+
Supplier Ledger
+
Expenses
+
Returns
+
Warranty
+
Profit
+
Audit
+
Reports
+
Notifications
+
Automation
```

ثم فوقها:

```text
Insights
        ↓
AI Assistant
```

---

# 118. الرؤية النهائية

عندما يفتح صاحب المحل PartFlow، لا ينبغي أن يشعر أنه فتح قاعدة بيانات.

بل يرى:

```text
Today's Sales       ₪ 7,450
Today's Profit      ₪ 1,850

Inventory Value     ₪ 185,400

Outstanding Debt    ₪ 24,800

Low Stock           4

Overdue Customers   3

Warranty Alerts     1
```

ثم يخبره النظام:

```text
⚠ 4 منتجات تحتاج إعادة طلب.

⚠ 3 عملاء لديهم ديون متأخرة.

📦 7 قطع لم تتحرك منذ 90 يومًا.

📈 RTX 3060 هو الأكثر مبيعًا هذا الشهر.

💰 متوسط هامش ربح القطع المستعملة ارتفع.
```

وهنا يتحقق جوهر المشروع:

# PartFlow لا يسجل ما يحدث في المحل فقط.

# PartFlow يفهم ما يحدث في المحل.

والمرحلة النهائية من المشروع هي:

> **من نظام إدارة إلى مساعد تشغيلي ذكي.**

---

# 119. الخلاصة التنفيذية

الـBackend النهائي يجب أن يكون:

```text
Go
+
PostgreSQL/Supabase
+
Modular Monolith
+
Multi-Tenant
+
RLS
+
RBAC
+
Transactions
+
Inventory Ledger
+
Customer/Supplier Ledgers
+
Audit
+
Worker
+
Automation
+
Reports
+
Insights
+
AI-ready
```

ويجب أن تكون كل عملية Business مهمة:

```text
Atomic
Auditable
Secure
Idempotent
Tenant-safe
Traceable
```

والهدف النهائي ليس بناء أكبر Backend ممكن.

الهدف هو بناء **Backend يجعل تجربة المستخدم بسيطة جدًا لأن كل التعقيد موجود في المكان الصحيح: داخل النظام نفسه.**

---

# 120. المبدأ الأخير

يجب أن يبقى هذا المبدأ حاضرًا طوال عملية التطوير:

> **لا تجعل صاحب المحل يعمل من أجل النظام؛ اجعل النظام يعمل من أجل صاحب المحل.**

إذا احتاج المستخدم إلى إدخال نفس المعلومة مرتين، فهناك مشكلة في التصميم.

إذا احتاج إلى فتح خمس صفحات لإتمام عملية بسيطة، فهناك مشكلة في الـWorkflow.

إذا احتاج إلى حساب شيء يدويًا يستطيع النظام حسابه، فهناك فرصة للأتمتة.

إذا كان النظام يعرف معلومة مهمة ولا يخبر المستخدم بها، فهناك فرصة للـInsights.

وإذا كان المستخدم يستطيع أن يسأل النظام سؤالًا عن أعماله ويحتاج إلى البحث يدويًا، فهناك فرصة مستقبلية للـAI Assistant.

**هذه هي فلسفة PartFlow التي يجب أن يترجمها الـBackend إلى نظام حقيقي، وليس مجرد مجموعة APIs.**

