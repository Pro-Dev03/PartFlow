# Smart Computer Store Management System

## Technical Architecture & Product Engineering Specification

### نظام ذكي لإدارة محلات قطع الحاسوب الجديدة والمستعملة

---

# 1. الرؤية التقنية

المشروع ليس POS تقليديًا، وليس Inventory CRUD، وليس مجرد نظام لإدارة الديون.

المشروع عبارة عن:

> **نظام تشغيل رقمي للمحل Store Operating System**

يقوم بربط دورة العمل كاملة:

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
Warranty / Returns
    ↓
Profit
    ↓
Reports
    ↓
Business Insights
```

لكن المستخدم لا يرى هذه التعقيدات كلها.

الهدف هو أن يرى المستخدم:

```text
ماذا لدي؟
ماذا بعت؟
من عليه مال؟
كم ربحت؟
ماذا يجب أن أفعل الآن؟
```

ويترك النظام بقية التفاصيل في الخلفية.

---

# 2. المبدأ الأول: النظام يعمل من أجل صاحبه

هذا هو المبدأ الذي يجب أن يحكم كل قرار تقني وUX في المشروع.

## النظام التقليدي

```text
المستخدم
   ↓
يدخل البيانات
   ↓
يفتح التقارير
   ↓
يحسب
   ↓
يبحث
   ↓
يحدث المخزون
   ↓
يراجع الديون
```

## النظام المقترح

```text
المستخدم
   ↓
يقوم بالعملية الطبيعية
   ↓
النظام يفهم العملية
   ↓
النظام يحدث البيانات المرتبطة
   ↓
النظام يحسب
   ↓
النظام ينبه
   ↓
النظام يعرض النتيجة
```

مثال:

الموظف يمسح Barcode فقط.

النظام يتولى:

```text
Identify Product
+
Identify Item
+
Check Stock
+
Load Price
+
Add To Cart
+
Update Inventory
+
Calculate Profit
+
Update Customer Balance
+
Create Audit Event
```

لا ينبغي أن يضطر المستخدم إلى القيام بهذه الخطوات يدويًا.

---

# 3. أهداف النظام

النظام يجب أن يكون:

* سريعًا.
* بسيطًا.
* واضحًا.
* آمنًا.
* قابلًا للتوسع.
* متعدد المؤسسات.
* مناسبًا للهاتف.
* مناسبًا لـWindows.
* مناسبًا للـTablet.
* يعمل عبر المتصفح.
* قابلًا للتغليف كتطبيق.
* Dockerized.
* Cloud Ready.
* Offline-aware.
* Automation Ready.
* AI Ready.

---

# 4. المنصات المستهدفة

يجب ألا يتم بناء ثلاثة أنظمة مستقلة.

المعمارية المقترحة:

```text
                    Core Platform
                         │
              React + TypeScript
                         │
                  Responsive UI
                         │
        ┌────────────────┼────────────────┐
        │                │                │
     Windows           Mobile           Tablet
      Browser          Browser          Browser
        │                │                │
        └────────────────┼────────────────┘
                         │
                        PWA
```

ثم يمكن لاحقًا تغليف نفس الواجهة كتطبيق Desktop/Mobile عند الحاجة.

---

# 5. استراتيجية التغليف Packaging Strategy

## الطبقة الأولى — Web

التطبيق يعمل مباشرة من:

```text
Chrome
Edge
Firefox
Safari
```

بدون تثبيت.

---

# 6. الطبقة الثانية — PWA

الواجهة ستكون:

```text
Progressive Web App
```

وبالتالي يستطيع المستخدم:

### Windows

اختيار:

```text
Install App
```

ليظهر التطبيق كتطبيق مستقل.

### Android

```text
Add to Home Screen
```

### iPhone

```text
Add to Home Screen
```

وهكذا يحصل المستخدم على تجربة قريبة جدًا من التطبيق الأصلي دون بناء واجهة منفصلة لكل نظام.

---

# 7. لماذا PWA؟

لأن طبيعة النظام تعتمد على:

* البحث.
* Barcode.
* المبيعات.
* العملاء.
* المخزون.
* التقارير.

ولا تحتاج في البداية إلى تطبيق Native كامل لكل منصة.

وبالتالي:

```text
One UI
+
One Codebase
+
Multiple Platforms
```

بدل:

```text
React Web
+
Android App
+
iOS App
+
Windows App
```

---

# 8. Desktop Packaging

إذا أصبح هناك احتياج فعلي إلى تطبيق Windows مستقل، يمكن تغليف الواجهة نفسها باستخدام:

```text
Tauri
```

بحيث تصبح:

```text
SmartStore.exe
```

بدون إعادة بناء النظام من الصفر.

المهم:

> Desktop Wrapper ليس Backend جديدًا.

بل:

```text
Tauri
   ↓
React App
   ↓
Go API
```

---

# 9. Mobile Packaging

المرحلة الأولى:

```text
React
+
PWA
```

المرحلة المستقبلية إذا احتجت تطبيقًا حقيقيًا:

```text
React
   ↓
Mobile Shell
   ↓
Android / iOS
```

لكن الـBackend يبقى نفسه.

---

# 10. القرار المعماري الأساسي

الـBackend:

# Go

الواجهة:

# React + TypeScript

قاعدة البيانات:

# PostgreSQL via Supabase

Storage:

# Supabase Storage

Authentication:

# Supabase Auth أو Authentication Layer مرتبطة بالـBackend

Deployment:

# Docker + Render

Frontend:

# PWA

---

# 11. Architecture Overview

```text
                         USERS
                           │
          ┌────────────────┼────────────────┐
          │                │                │
       Windows          Android            iOS
       Browser           PWA               PWA
          │                │                │
          └────────────────┼────────────────┘
                           │
                           ▼
                  React + TypeScript
                           │
                           │ HTTPS
                           ▼
                     Go Backend
                           │
         ┌─────────────────┼─────────────────┐
         │                 │                 │
         ▼                 ▼                 ▼
     Business           Security          Automation
      Logic              Layer              Layer
         │
         ▼
      PostgreSQL
      Supabase
         │
     ┌───┼──────────────┐
     │   │              │
     ▼   ▼              ▼
 Storage Auth           RLS
```

---

# 12. Backend Architecture

Go يجب أن يكون:

```text
API
+
Business Logic
+
Security
+
Transactions
+
Validation
+
Domain Services
```

وليس مجرد طبقة CRUD.

---

# 13. Modular Monolith

البنية المناسبة للنسخة الأولى:

# Modular Monolith

وليس Microservices.

أي تطبيق Go واحد، لكن داخله Modules مستقلة:

```text
Auth
Organizations
Products
Inventory
Sales
Customers
Debts
Payments
Purchases
Suppliers
Expenses
Returns
Warranty
Reports
Notifications
Audit
Automation
```

---

# 14. لماذا Modular Monolith؟

لأن Microservices في البداية ستضيف تعقيدًا غير ضروري:

```text
Networking
Service Discovery
Distributed Transactions
Multiple Deployments
Distributed Logging
Message Brokers
```

بينما المشروع يحتاج أولًا إلى:

```text
Correctness
Speed
Security
Simplicity
Maintainability
```

وعندما يكبر المشروع يمكن فصل Modules معينة إلى Services مستقلة.

---

# 15. Go Project Structure

```text
smart-store/
│
├── backend/
│   │
│   ├── cmd/
│   │   └── api/
│   │       └── main.go
│   │
│   ├── internal/
│   │   │
│   │   ├── auth/
│   │   ├── organizations/
│   │   ├── users/
│   │   ├── roles/
│   │   │
│   │   ├── products/
│   │   ├── categories/
│   │   ├── brands/
│   │   │
│   │   ├── inventory/
│   │   ├── barcode/
│   │   ├── locations/
│   │   │
│   │   ├── sales/
│   │   ├── payments/
│   │   ├── customers/
│   │   ├── debts/
│   │   │
│   │   ├── purchases/
│   │   ├── suppliers/
│   │   │
│   │   ├── expenses/
│   │   ├── returns/
│   │   ├── warranties/
│   │   ├── inspections/
│   │   │
│   │   ├── reports/
│   │   ├── notifications/
│   │   ├── automation/
│   │   └── audit/
│   │
│   ├── pkg/
│   │   ├── logger/
│   │   ├── validator/
│   │   ├── response/
│   │   ├── middleware/
│   │   └── errors/
│   │
│   ├── migrations/
│   ├── tests/
│   ├── Dockerfile
│   └── go.mod
│
├── frontend/
│
├── worker/
│
├── docker/
│
├── docs/
│
├── scripts/
│
├── docker-compose.yml
├── .env.example
└── README.md
```

---

# 16. Domain Architecture

كل Domain يكون مسؤولًا عن نفسه.

مثال:

```text
inventory/
├── model.go
├── repository.go
├── service.go
├── handler.go
├── dto.go
├── validator.go
└── errors.go
```

و:

```text
sales/
├── model.go
├── repository.go
├── service.go
├── handler.go
├── dto.go
└── errors.go
```

---

# 17. Frontend Architecture

```text
frontend/
│
├── src/
│   ├── app/
│   ├── routes/
│   ├── components/
│   ├── layouts/
│   │
│   ├── features/
│   │   ├── auth/
│   │   ├── dashboard/
│   │   ├── products/
│   │   ├── inventory/
│   │   ├── sales/
│   │   ├── customers/
│   │   ├── debts/
│   │   ├── suppliers/
│   │   ├── purchases/
│   │   ├── expenses/
│   │   ├── reports/
│   │   └── settings/
│   │
│   ├── services/
│   ├── hooks/
│   ├── types/
│   ├── lib/
│   └── styles/
│
├── public/
├── package.json
├── vite.config.ts
└── tsconfig.json
```

---

# 18. UX Architecture

أكبر خطر في المشروع ليس الأداء.

أكبر خطر هو:

# التعقيد.

إذا أصبح النظام يحتاج إلى شرح طويل، فقد فشل المشروع UX حتى لو كانت الـArchitecture ممتازة.

---

# 19. قاعدة UX الرئيسية

كل عملية يجب أن تسأل:

> ما أقل عدد خطوات يحتاجها المستخدم لإنجازها؟

مثال بيع قطعة:

```text
Scan
↓
Confirm
↓
Payment
↓
Done
```

وليس:

```text
Products
↓
Inventory
↓
Select Product
↓
Select Variant
↓
Select Item
↓
Customer
↓
Payment
↓
Confirm
↓
Save
```

النظام يجب أن يختصر هذه الخطوات.

---

# 20. الواجهة الرئيسية

بدل Dashboard مليء بالمخططات:

```text
Revenue
Graph
Pie Chart
Graph
Table
Graph
```

يجب أن يكون:

```text
صباح الخير 👋

اليوم

المبيعات       ₪ 7,450
الربح          ₪ 1,850
المخزون        ₪ 185,400
الديون         ₪ 24,800

يحتاج انتباهك

4 منتجات منخفضة
3 ديون متأخرة
1 ضمان ينتهي قريبًا
```

---

# 21. Navigation

القائمة الأساسية يجب أن تكون صغيرة.

```text
الرئيسية

المبيعات

المخزون

العملاء

الديون

المشتريات

الموردون

المصروفات

التقارير

الإعدادات
```

لا تظهر عشرات القوائم في المستوى الأول.

---

# 22. Smart Actions

في Dashboard:

```text
+ بيع
+ إضافة قطعة
+ إضافة عميل
+ تسجيل دفعة
+ إضافة مصروف
```

هذه هي العمليات الأكثر استخدامًا.

---

# 23. Global Search

زر بحث واحد يستطيع البحث في:

```text
Product
Barcode
Serial
Customer
Phone
Invoice
Sale
Supplier
```

مثال:

```text
RTX3060
```

يظهر:

```text
RTX 3060 ASUS
3 Items

2 Used
1 New
```

---

# 24. Global Barcode Action

زر دائم:

```text
SCAN
```

عند الضغط:

```text
Camera
```

أو انتظار Scanner خارجي.

بعد المسح:

```text
Product Found
```

ويقرر النظام تلقائيًا ماذا يفعل.

إذا كان في شاشة البيع:

```text
Add to Cart
```

إذا كان في المخزون:

```text
Open Item
```

إذا كان Barcode غير معروف:

```text
Create Product
```

هذه هي الـContext Awareness.

---

# 25. Context-Aware System

النظام يجب أن يعرف أين المستخدم.

مثلاً:

إذا كان المستخدم في:

```text
Sales
```

فمسح Barcode يعني:

```text
Add Item
```

أما إذا كان في:

```text
Inventory
```

فمسح Barcode يعني:

```text
View Item
```

وبالتالي لا يحتاج المستخدم إلى اختيار ما يريده كل مرة.

---

# 26. Product Model

يجب الفصل بين:

# Product

و:

# Inventory Item

مثال:

```text
Product:
RTX 3060 ASUS

Inventory:
ITEM-001
ITEM-002
ITEM-003
```

---

# 27. Product

```text
products
---------
id
organization_id
name
brand_id
category_id
model
sku
description
track_serial
track_individual
created_at
updated_at
```

---

# 28. Inventory Item

```text
inventory_items
---------------
id
organization_id
product_id
barcode
serial_number
condition
grade
purchase_cost
selling_price
status
location_id
supplier_id
created_at
updated_at
```

---

# 29. Product Conditions

```text
NEW
USED
REFURBISHED
DAMAGED
FOR_PARTS
```

---

# 30. Used Product Grade

```text
A
B
C
D
```

مع:

```text
condition_notes
inspection_status
inspection_date
inspected_by
```

---

# 31. Item Lifecycle

كل قطعة لها دورة حياة:

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
   ↓
WARRANTY
   ↓
RETURNED
```

ليس كل Item يمر بكل الحالات.

---

# 32. Barcode System

النظام يدعم:

```text
External Barcode
Internal Barcode
Serial Number
SKU
```

إذا لم يوجد Barcode:

```text
System
↓
Generate Internal Barcode
↓
Print
↓
Attach To Item
```

---

# 33. Individual Item Tracking

القطع الفردية مثل:

* GPU.
* CPU.
* Motherboard.
* SSD.
* HDD.
* PSU.
* Laptop.

يمكن تتبعها كقطعة مستقلة.

مثال:

```text
ITEM-000541

RTX 3070
Used
Grade B

Purchase:
900 ₪

Repair:
50 ₪

Total Cost:
950 ₪

Selling:
1,250 ₪

Profit:
300 ₪
```

---

# 34. Inventory Ledger

لا يتم تعديل المخزون بدون سبب.

كل حركة:

```text
PURCHASE
SALE
RETURN
ADJUSTMENT
DAMAGE
TRANSFER
RESERVATION
RELEASE
```

تخزن في:

```text
inventory_movements
```

---

# 35. Sales Engine

عملية البيع:

```text
Create Cart
↓
Scan Items
↓
Customer Optional
↓
Discount Optional
↓
Payment
↓
Confirm
```

النظام يقوم تلقائيًا بـ:

```text
Create Sale
Create Sale Items
Update Inventory
Create Payment
Update Customer Balance
Calculate COGS
Calculate Profit
Create Audit Event
```

---

# 36. Payment Engine

يجب فصل:

```text
Sale
```

عن:

```text
Payment
```

لأن:

```text
Sale = 2,000 ₪
```

يمكن أن تصبح:

```text
Payment 1 = 500
Payment 2 = 500
Payment 3 = 1,000
```

---

# 37. Customer Debt

الرصيد يجب أن يكون Ledger.

```text
Sale       +2,000
Payment     -500
Payment     -500
----------------
Balance     1,000
```

وليس مجرد حقل:

```text
debt = 1000
```

---

# 38. Debt Automation

إذا أصبح الدين متأخرًا:

```text
Debt
↓
Due Date
↓
Overdue
↓
Notification
```

في Dashboard:

```text
3 customers need attention
```

بدل أن يبحث صاحب المحل بنفسه.

---

# 39. Customer Profile

صفحة العميل:

```text
Ahmed

Phone
Email

Total Purchases
Total Paid
Outstanding

Recent Sales
Payments
Returns
Notes
```

ويجب أن يرى صاحب المحل الصورة المالية فورًا.

---

# 40. Supplier Management

المورد:

```text
Supplier
↓
Purchases
↓
Payments
↓
Outstanding
```

مع Ledger خاص به.

---

# 41. Purchase Management

عند وصول بضاعة:

```text
Purchase
↓
Supplier
↓
Scan / Add Items
↓
Purchase Cost
↓
Receive
```

ويقوم النظام بتحديث:

```text
Inventory
Supplier Balance
Cost
Reports
Audit
```

---

# 42. Expenses

النظام يدعم:

```text
Rent
Electricity
Internet
Salary
Shipping
Repair
Equipment
Other
```

والنظام يستخدمها في حساب صافي الربح.

---

# 43. Profit Engine

```text
Revenue
-
Cost Of Goods Sold
=
Gross Profit

Gross Profit
-
Operating Expenses
=
Net Profit
```

يجب فصل الحسابات وعدم خلطها.

---

# 44. Used Item Cost

القطعة المستعملة قد تحتوي على:

```text
Purchase Cost
+
Repair Cost
+
Testing Cost
+
Other Cost
```

وبالتالي:

```text
Total Item Cost
```

هو الرقم الذي يستخدم لحساب الربح الحقيقي.

---

# 45. Inspection

عند شراء قطعة مستعملة:

```text
Power
Temperature
Performance
Ports
Storage
Visual
Accessories
Serial
```

ثم:

```text
PASS
FAIL
PARTIAL
```

لكن هذه العملية يجب أن تكون اختيارية وسريعة.

لا تجعل الموظف يملأ 30 حقلًا لكل قطعة.

---

# 46. Smart Defaults

النظام يجب أن يتذكر:

```text
Last Supplier
Last Price
Default Category
Default Warranty
Default Location
```

ويملأها تلقائيًا عند الحاجة.

المستخدم يستطيع تعديلها.

---

# 47. Smart Forms

لا تعرض كل الحقول.

مثلاً عند إضافة Keyboard جديد:

اعرض:

```text
Name
Price
Quantity
Barcode
```

أما:

```text
Serial
Inspection
Repair
Warranty Claim
```

فلا تظهر إلا إذا كانت مطلوبة.

---

# 48. Progressive Disclosure

المعلومات المتقدمة تظهر عند الحاجة.

مثال:

```text
Product
────────────
Name
Price
Stock

[More Details]
```

عند الضغط:

```text
Serial
Warranty
Supplier
Cost
Location
Notes
```

وهكذا يبقى النظام بسيطًا.

---

# 49. Responsive Design

## Desktop

Sidebar:

```text
Dashboard
Sales
Inventory
Customers
...
```

## Mobile

Bottom Navigation:

```text
Home
Sales
Scan
Inventory
More
```

---

# 50. Mobile Priority

على الهاتف يجب أن تكون العمليات الأساسية:

```text
Scan
Search
Sell
Customer
Payment
```

متاحة بسرعة.

---

# 51. Windows Priority

على Windows:

* Keyboard shortcuts.
* Barcode Scanner.
* Large tables.
* Multi-column views.
* Fast search.
* Bulk actions.
* Printing.

---

# 52. Keyboard Shortcuts

مثلاً:

```text
F1  Search
F2  New Sale
F3  Scan
F4  Customer
F5  Refresh
ESC Close
```

يمكن تغييرها لاحقًا.

---

# 53. Printing

يجب دعم:

* Invoice.
* Barcode Label.
* Product Label.
* Customer Receipt.

وعلى Windows يمكن دعم الطباعة التقليدية.

---

# 54. PWA Service Worker

يجب استخدام Service Worker من أجل:

```text
Cache Static Assets
Offline UI
Fast Startup
Installability
```

لكن:

> لا يعني PWA أن كل العمليات المالية تعمل Offline تلقائيًا.

---

# 55. Offline Mode

يجب تقسيم العمليات:

## Safe Offline

مثل:

```text
View Cached Products
View Cached Customers
Open Recent Sales
```

## Sensitive Offline

مثل:

```text
Sale
Payment
Debt
Inventory Adjustment
```

هذه تحتاج استراتيجية Sync قوية.

---

# 56. Offline Queue

في المرحلة المتقدمة:

```text
User
 ↓
Offline Sale
 ↓
Local Queue
 ↓
Connection Restored
 ↓
Sync
 ↓
Server Validation
 ↓
Commit
```

إذا حدث Conflict:

```text
Conflict Detected
```

ولا يتم إخفاؤه عن النظام.

---

# 57. Multi-Tenant Architecture

النظام SaaS منذ البداية.

```text
Organization
│
├── Users
├── Products
├── Inventory
├── Customers
├── Suppliers
├── Sales
├── Purchases
└── Reports
```

كل سجل Business Data مرتبط بـ:

```text
organization_id
```

---

# 58. Database

```text
Supabase
    ↓
PostgreSQL
```

الجداول الأساسية:

```text
organizations

users
roles
permissions

products
categories
brands
product_barcodes

inventory_items
inventory_movements
inventory_locations

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

returns
return_items

inspections
inspection_items

warranties
warranty_claims

reservations

notifications
audit_logs
attachments
```

---

# 59. RLS

يجب تفعيل:

```text
Row Level Security
```

لحماية البيانات حسب المؤسسة والمستخدم والصلاحيات.

Supabase توصي بتفعيل RLS على الجداول الموجودة في الـexposed schema، ويمكن استخدامه لعزل بيانات المستخدمين والمؤسسات على مستوى الصفوف.

مثال منطقي:

```text
organization_id
=
current user's organization
```

وبذلك حتى لو حدث خطأ في التطبيق، توجد طبقة حماية إضافية على مستوى PostgreSQL.

---

# 60. Supabase Security

البنية المقترحة:

```text
React
   ↓
Go API
   ↓
Supabase
```

المفاتيح السرية لا تدخل إلى React.

أي Secret Server Key:

```text
Backend Only
```

أما مفاتيح العميل العامة فتستخدم وفق نموذج Supabase مع RLS وسياسات صحيحة. توضح وثائق Supabase أن المفاتيح السرية المخصصة للخادم يجب ألا تكشف للواجهة، وأن RLS هو طبقة الحماية الأساسية عند تعريض البيانات للعميل.

---

# 61. Storage

الصور:

```text
Product Images
Item Images
Inspection Photos
Invoices
Documents
```

تخزن في:

```text
Supabase Storage
```

مع سياسات وصول حسب:

```text
organization_id
```

---

# 62. Authentication

النظام يدعم:

```text
Login
Logout
Password Reset
Session
User Profile
```

ثم:

```text
User
↓
Organization
↓
Role
↓
Permissions
```

---

# 63. Authorization

لا تعتمد على Role فقط.

استخدم Permissions.

مثال:

```text
products.read
products.create
products.update

inventory.read
inventory.adjust

sales.create
sales.refund

customers.read
customers.update

debts.read
debts.payment

reports.read

expenses.create
```

---

# 64. Audit Log

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

مثال:

```text
Employee
changed selling price

RTX 3060

Old:
1,200 ₪

New:
1,150 ₪
```

---

# 65. Financial Immutability

لا نحذف المعاملات المالية.

إذا حدث خطأ:

```text
Reverse
```

بدل:

```text
Delete
```

مثلاً:

```text
Payment +500
```

تم إدخالها خطأ:

```text
Payment Reversal -500
```

وهذا يحافظ على تاريخ مالي قابل للتدقيق.

---

# 66. Transaction Safety

عملية البيع يجب أن تكون Atomic.

```text
BEGIN

Validate Items

Validate Stock

Create Sale

Create Sale Items

Update Inventory

Create Payment

Update Customer Ledger

Create Audit Event

COMMIT
```

إذا فشل شيء:

```text
ROLLBACK
```

---

# 67. Concurrency

إذا حاول موظفان بيع نفس القطعة:

```text
Employee A
   ↓
ITEM-001
```

و:

```text
Employee B
   ↓
ITEM-001
```

النتيجة:

```text
A → SUCCESS
B → ITEM ALREADY SOLD
```

وليس:

```text
A → SUCCESS
B → SUCCESS
```

---

# 68. Idempotency

الـAPI يجب أن يمنع تكرار العمليات الناتج عن:

* Double Click.
* Network Retry.
* Mobile Connection.
* Browser Retry.

خصوصًا:

```text
POST /sales
POST /payments
POST /returns
```

---

# 69. API

```text
/api/v1/auth

/api/v1/products
/api/v1/categories
/api/v1/brands

/api/v1/inventory
/api/v1/inventory/items
/api/v1/inventory/movements

/api/v1/sales
/api/v1/payments

/api/v1/customers
/api/v1/customers/{id}/ledger

/api/v1/debts

/api/v1/purchases
/api/v1/suppliers

/api/v1/expenses

/api/v1/returns
/api/v1/warranties
/api/v1/inspections

/api/v1/reports
/api/v1/dashboard

/api/v1/notifications
```

---

# 70. API Versioning

من البداية:

```text
/api/v1
```

حتى يمكن تطوير:

```text
/api/v2
```

مستقبلاً دون كسر العملاء الحاليين.

---

# 71. API Response

صيغة موحدة:

```json
{
  "success": true,
  "data": {},
  "meta": {},
  "error": null
}
```

الأخطاء:

```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "ITEM_ALREADY_SOLD",
    "message": "This item is no longer available."
  }
}
```

---

# 72. Error Codes

أمثلة:

```text
ITEM_NOT_FOUND
PRODUCT_NOT_FOUND
INSUFFICIENT_STOCK
ITEM_ALREADY_SOLD
DUPLICATE_SERIAL
INVALID_PAYMENT
CUSTOMER_NOT_FOUND
PERMISSION_DENIED
DEBT_NOT_FOUND
INVALID_OPERATION
```

---

# 73. Search

البحث يجب أن يكون سريعًا.

يبحث في:

```text
Name
SKU
Barcode
Serial
Model
Brand
Customer
Phone
Invoice
```

في البداية يمكن الاعتماد على PostgreSQL مع Indexing جيد.

لا تضف Elasticsearch إلى المشروع لمجرد أنه يبدو احترافيًا.

---

# 74. Database Indexing

Index على:

```text
organization_id
barcode
serial_number
sku
phone
product_id
customer_id
supplier_id
created_at
status
```

والـComposite Indexes حسب الاستعلامات الفعلية.

---

# 75. Pagination

لا ترسل آلاف السجلات للهاتف.

استخدم:

```text
limit
cursor
```

أو Pagination مناسبة.

---

# 76. Dashboard API

بدل 15 Request عند فتح Dashboard:

```text
GET /api/v1/dashboard
```

يعيد:

```text
sales
profit
inventory
debts
low_stock
alerts
top_products
```

وهذا يقلل زمن التحميل.

---

# 77. Background Worker

بعض الأعمال لا يجب أن تتم داخل HTTP Request.

مثلاً:

```text
Generate Large Report
Send Notification
Process Image
Daily Analysis
Warranty Scan
Debt Scan
```

تذهب إلى:

```text
Worker
```

---

# 78. Automation Engine

النظام يعمل تلقائيًا:

```text
Daily
↓
Check Low Stock

Check Overdue Debts

Check Warranty

Generate Insights
```

---

# 79. Notifications

أنواع:

```text
LOW_STOCK
OVERDUE_DEBT
WARRANTY_EXPIRING
ITEM_RESERVED
PAYMENT_RECEIVED
PURCHASE_RECEIVED
```

لكن لا تجعل الإشعارات مزعجة.

النظام يجب أن يرسل:

> ما يحتاج انتباه المستخدم فقط.

---

# 80. Smart Insights

بدل أن يطلب المستخدم التقرير:

```text
النظام:
لديك 4 منتجات منخفضة المخزون.
```

```text
النظام:
هناك 3 ديون متأخرة.
```

```text
النظام:
هذا المنتج لم يتحرك منذ 90 يومًا.
```

```text
النظام:
مبيعات القطع المستعملة ارتفعت هذا الشهر.
```

---

# 81. AI Future Layer

AI ليس جزءًا من Core Transaction Engine.

AI يجلس فوق النظام:

```text
Database
   ↓
Analytics Layer
   ↓
AI Assistant
```

مثال:

```text
"كم ربحت من GPU المستعملة هذا الشهر؟"
```

AI لا يخترع الإجابة.

بل:

```text
Question
↓
Intent
↓
Permission
↓
Safe Query
↓
Database
↓
Result
↓
Natural Language
```

---

# 82. Docker Architecture

المشروع يجب أن يكون:

```text
Dockerized
```

الخدمات:

```text
frontend
api
worker
```

وإذا احتجت:

```text
redis
```

لاحقًا.

أما Supabase في الإنتاج فيمكن استخدام Managed Supabase بدل تشغيل كامل Stack داخلي.

---

# 83. Development Docker

```text
docker-compose.yml

frontend
backend
worker
```

والـSupabase المحلي اختياري.

Supabase توفر أيضًا طريقة رسمية لتشغيل نسخة Self-Hosted باستخدام Docker Compose، لكن ذلك يضيف خدمات وموارد وصيانة ومسؤوليات تشغيلية أكبر؛ لذلك ليس من الضروري تشغيل كامل Supabase داخل Docker في Production إذا كان Managed Supabase مناسبًا لك.

---

# 84. Production Docker

```text
GitHub
   ↓
CI
   ↓
Docker Build
   ↓
Container Registry
   ↓
Render
```

---

# 85. Go Dockerfile

يستخدم:

```text
Multi-stage Build
```

النتيجة:

```text
Go Source
↓
Builder
↓
Compiled Binary
↓
Minimal Runtime Image
```

لا يجب وضع أدوات التطوير داخل Production Image.

---

# 86. Frontend Docker

```text
Node
↓
Install
↓
Build
↓
Static Assets
↓
Nginx
```

أو يمكن نشر الـFrontend كـStatic Service منفصل.

---

# 87. Container Security

كل Container:

* Non-root.
* Minimal.
* لا يحتوي Secrets.
* لا يفتح Ports غير ضرورية.
* Healthcheck.
* Read-only حيثما أمكن.
* Image Scan.

---

# 88. Render

Render مناسب كطبقة Deployment للـBackend والـFrontend/Services.

الهدف:

```text
Developer
↓
git push
↓
CI/CD
↓
Production
```

بدون إدارة Server يدويًا في كل مرة.

---

# 89. Environment Variables

Development:

```text
.env.local
```

Production:

```text
Render Environment
```

أمثلة:

```text
DATABASE_URL
SUPABASE_URL
SUPABASE_SECRET_KEY
JWT_SECRET
APP_ENV
APP_URL
```

ولا يتم Commit لهذه القيم.

---

# 90. CI/CD

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

عند نجاح Production Branch:

```text
Deploy
```

---

# 91. Testing

يجب اختبار:

## Inventory

```text
Purchase
Sale
Return
Adjustment
Transfer
```

## Finance

```text
Sale
Payment
Debt
Refund
Profit
```

## Security

```text
Organization Isolation
Permissions
Unauthorized Access
```

---

# 92. أهم Integration Tests

اختبار كامل:

```text
Create Product
↓
Receive Item
↓
Scan Item
↓
Sell Item
↓
Partial Payment
↓
Customer Debt
↓
Payment Later
↓
Profit Updated
↓
Inventory Updated
```

إذا نجح هذا الاختبار، فجزء كبير من Core Business يعمل بشكل صحيح.

---

# 93. Security Testing

اختبار:

```text
User A
```

يحاول الوصول إلى:

```text
Organization B
```

النتيجة:

```text
403 / No Data
```

ويجب أن تكون الحماية في:

```text
Backend
+
Database RLS
```

وليس Frontend فقط.

---

# 94. Performance

الأهداف الأولية:

```text
Barcode Lookup
< 300ms target

Normal API
< 500ms target

Dashboard
< 2s target
```

هذه أهداف تصميمية أولية وليست ضمانات.

---

# 95. Mobile Performance

يجب:

* تقليل JavaScript.
* Lazy Loading.
* Code Splitting.
* Image Optimization.
* Pagination.
* Cache.
* Avoid huge tables.
* Avoid unnecessary API calls.

---

# 96. Responsive Tables

الجداول الكبيرة على الهاتف لا يجب أن تصبح:

```text
← → ← → ← →
```

بدل ذلك:

Desktop:

```text
Product | SKU | Cost | Price | Stock | Status
```

Mobile:

```text
RTX 3060
1,150 ₪
Stock: 3
Used
```

ثم:

```text
View Details
```

---

# 97. Empty States

لا تعرض:

```text
No Data
```

فقط.

بل:

```text
لا توجد قطع حتى الآن.

[إضافة قطعة]
```

---

# 98. Error States

بدل:

```text
500 Internal Server Error
```

المستخدم يرى:

```text
حدث خطأ أثناء تحميل البيانات.

[إعادة المحاولة]
```

والتفاصيل التقنية تذهب إلى Logs.

---

# 99. Loading States

يجب استخدام:

```text
Skeleton
```

بدل تجميد الشاشة.

---

# 100. Confirmation UX

لا تسأل المستخدم:

```text
Are you sure?
```

في كل شيء.

العمليات الخطرة فقط:

```text
Delete
Refund
Inventory Adjustment
Cancel Sale
```

---

# 101. Smart Defaults

مثلاً عند إنشاء Sale:

```text
Payment Method:
Cash
```

إذا كان هذا هو الأكثر استخدامًا.

وعند إضافة Product:

```text
Currency:
ILS
```

حسب إعداد المؤسسة.

---

# 102. Localization

يجب أن يدعم النظام من البداية:

```text
Hebrew
Arabic
English
```

مع:

```text
RTL
LTR
```

خصوصًا لأن النظام مستهدف لسوق يمكن أن يحتاج واجهة عبرية وعربية وإنجليزية.

---

# 103. Currency

يجب تصميم الـFinance Engine ليكون قادرًا على التعامل مع:

```text
ILS
```

والعملات الأخرى مستقبلًا.

لا تستخدم Floating Point للحسابات المالية.

استخدم:

```text
Decimal
```

أو Minor Units بحسب تصميم قاعدة البيانات.

---

# 104. Timezone

التواريخ تخزن بشكل موحد، ثم تعرض حسب Timezone المؤسسة.

مثال:

```text
Asia/Jerusalem
```

---

# 105. Auditability

كل عملية حساسة يجب أن تكون قابلة للإجابة:

```text
من فعلها؟
متى؟
على ماذا؟
ما القيمة قبل؟
ما القيمة بعد؟
لماذا؟
```

---

# 106. Backup

يجب أن يكون هناك:

```text
Automated Backups
Retention
Restore Procedure
Recovery Test
```

لا يكفي وجود Backup دون اختبار الاستعادة.

---

# 107. Disaster Recovery

يجب تحديد:

```text
RPO
RTO
```

من البداية.

مثال أولي:

```text
RPO: 24h
RTO: 4h
```

ثم تحسينها مع نمو المنتج.

---

# 108. Business Data Model

أهم قاعدة:

> لا تخزن النتائج فقط؛ خزّن الأحداث التي أدت إلى النتيجة.

مثال سيئ:

```text
stock = 7
```

مثال صحيح:

```text
Purchase +10
Sale -2
Sale -1
Return +1
Adjustment -1
```

ثم:

```text
Current Stock = 7
```

---

# 109. Financial Data Model

لا تخزن:

```text
customer.debt = 500
```

فقط.

بل:

```text
Sale +2000
Payment -500
Payment -1000
```

والرصيد:

```text
500
```

---

# 110. النظام القابل للتفسير

كل رقم يظهر للمستخدم يجب أن يستطيع النظام تفسيره.

مثلاً:

```text
Inventory Value:
₪185,400
```

يجب أن يستطيع المستخدم الضغط عليه ومعرفة:

```text
Products
Items
Cost
Quantities
```

وبالنسبة للربح:

```text
Profit:
₪18,500
```

يمكن فتح:

```text
Revenue
COGS
Expenses
Net Profit
```

---

# 111. Dashboard ليس تقريرًا

Dashboard:

> ماذا يحدث الآن؟

Reports:

> ماذا حدث؟

Analytics:

> لماذا حدث؟

Automation:

> ماذا يجب أن يحدث؟

AI:

> ماذا يعني ذلك؟

هذه طبقات مختلفة.

---

# 112. Smart Store Assistant

يمكن أن يكون هناك قسم:

# يحتاج انتباهك

مثال:

```text
4 منتجات منخفضة المخزون

3 عملاء لديهم ديون متأخرة

2 قطع مستعملة لم يتم فحصها

1 ضمان ينتهي خلال 7 أيام
```

هذا القسم أهم من عشرات الرسوم البيانية.

---

# 113. Principle: Zero Unnecessary Input

إذا كان النظام يستطيع استنتاج معلومة صحيحة من عملية أخرى، لا تطلب من المستخدم إدخالها مرة ثانية.

مثال:

عند بيع Item:

لا تطلب:

```text
Product
Price
Cost
Stock
Customer
```

النظام يعرف معظمها.

المستخدم يؤكد فقط ما يحتاج تأكيدًا.

---

# 114. Principle: One Action, Many Updates

مثال:

```text
Sell RTX 3060
```

تؤدي تلقائيًا إلى:

```text
Inventory -
Revenue +
COGS +
Profit +
Customer Balance +
Payment +
Audit +
Analytics
```

هذه هي روح النظام.

---

# 115. Principle: Complexity Behind the Curtain

المستخدم يرى:

```text
بيع
```

لكن Backend ينفذ:

```text
Transaction
Inventory Lock
Sale Creation
Payment
Ledger
Profit
Audit
Events
Notifications
```

كل هذه التفاصيل يجب أن تختفي خلف زر واحد.

---

# 116. Offline/Online Awareness

المستخدم يجب أن يعرف حالة الاتصال:

```text
● Connected
```

أو:

```text
● Offline
```

ولكن لا تظهر رسائل تقنية مخيفة.

مثلاً:

```text
الاتصال بالإنترنت ضعيف.
سيتم حفظ العمليات المسموح بها ومزامنتها لاحقًا.
```

---

# 117. Hardware Integration

النظام يجب أن يدعم:

```text
USB Barcode Scanner
Camera
Printer
Receipt Printer
Label Printer
Keyboard
Mouse
Touch Screen
```

خصوصًا على Windows.

---

# 118. Receipt Flow

```text
Sale
↓
Payment
↓
Receipt
↓
Print / PDF / Share
```

على الهاتف:

```text
Share
```

وعلى Windows:

```text
Print
```

---

# 119. Barcode Label Flow

```text
Create Item
↓
Generate Barcode
↓
Preview Label
↓
Print
```

والـLabel يمكن أن يحتوي:

```text
Product Name
Price
Barcode
Item ID
```

---

# 120. Location Management

يمكن دعم:

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

وعند Scan:

```text
Location:
Shelf B / Box 04
```

---

# 121. Reservation System

يمكن حجز Item:

```text
Available
↓
Reserved
↓
Sold
```

مع:

```text
Customer
Expiration
```

والنظام يحرر الحجز تلقائيًا عند انتهاء المدة.

---

# 122. Return Flow

```text
Customer
↓
Sale
↓
Return Request
↓
Inspection
↓
Approved
↓
Refund / Exchange
↓
Inventory Update
```

ولا يتم حذف Sale الأصلية.

---

# 123. Warranty Flow

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

وهذا مهم جدًا للقطع المستعملة.

---

# 124. Reporting

التقارير الأساسية:

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

# 125. Business Questions

التقارير يجب أن تجيب:

```text
كم بعت اليوم؟
كم ربحت؟
ما أكثر منتج مبيعًا؟
ما أكثر منتج ربحًا؟
من عليه أكبر دين؟
ما قيمة المخزون؟
ما القطع الراكدة؟
ما القطع التي تحتاج إعادة طلب؟
كم صرفت؟
كم دفعت للموردين؟
```

---

# 126. Export

التقارير يمكن تصديرها إلى:

```text
PDF
CSV
Excel
```

لكن لا تجعل التصدير هو طريقة العمل الأساسية.

Dashboard أولًا.

Export عند الحاجة.

---

# 127. SaaS Readiness

المشروع يجب أن يكون قابلًا مستقبلًا:

```text
One Store
↓
Multiple Stores
↓
Multiple Branches
↓
Multiple Warehouses
```

بدون إعادة بناء قاعدة البيانات بالكامل.

---

# 128. Branch Architecture

مستقبلاً:

```text
Organization
│
├── Branch A
│   └── Inventory
│
├── Branch B
│   └── Inventory
│
└── Branch C
    └── Inventory
```

مع إمكانية:

```text
Transfer Item
```

بين الفروع.

---

# 129. Future E-Commerce

يمكن لاحقًا ربط:

```text
Online Store
```

بنفس:

```text
Products
Inventory
Prices
Orders
Customers
```

وبالتالي لا يصبح هناك مخزونان منفصلان.

---

# 130. Future Accounting Integration

يمكن لاحقًا إضافة تكامل مع أنظمة محاسبية خارجية.

لكن Core System يبقى مستقلًا.

---

# 131. Future Messaging

يمكن إضافة:

```text
WhatsApp
Email
SMS
Push Notifications
```

لكن لا تجعل هذه الخدمات جزءًا من Core Transaction.

---

# 132. Future AI

AI يمكن أن يقدم:

```text
Business Questions
Inventory Forecast
Purchase Suggestions
Slow Stock Detection
Customer Insights
Profit Analysis
```

لكن لا يسمح له بتجاوز:

```text
Authorization
Business Rules
Audit
```

---

# 133. Future Predictive Inventory

النظام يستطيع لاحقًا تحليل:

```text
Historical Sales
Seasonality
Current Stock
Supplier Lead Time
```

ثم:

```text
Recommended Purchase
```

مثال:

```text
RTX 3060

Current:
2

Average Monthly Sales:
8

Recommended Purchase:
6
```

---

# 134. AI Safety

AI لا يعدل:

```text
Sales
Payments
Debts
Inventory
```

مباشرة.

أي عملية تعديل تمر عبر:

```text
Command
↓
Permission
↓
Validation
↓
Business Rule
↓
Transaction
↓
Audit
```

---

# 135. Performance Architecture

المشروع يجب أن يكون سريعًا حتى على:

```text
Old Windows PC
Mid-range Android
Weak Internet
```

لذلك:

```text
Small Bundles
Lazy Loading
Pagination
Caching
Optimized Queries
Compressed Images
Efficient API
```

---

# 136. Scalability

البنية تبدأ:

```text
1 API
1 Worker
1 Database
```

ثم:

```text
Load Balancer
↓
API 1
API 2
API 3
↓
Database
```

والـWorker مستقل.

---

# 137. Monitoring

يجب مراقبة:

```text
API Errors
Latency
Database Errors
Failed Jobs
Login Failures
Inventory Conflicts
Payment Failures
```

---

# 138. Health Endpoints

```text
/health
/ready
```

مثلاً:

```json
{
  "status": "ok",
  "database": "ok",
  "version": "1.0.0"
}
```

---

# 139. Logging

Structured Logging:

```text
timestamp
level
request_id
user_id
organization_id
action
duration
status
```

لا تسجل:

```text
Passwords
Secrets
Tokens
Sensitive Payment Data
```

---

# 140. Request Tracing

كل Request له:

```text
request_id
```

ليسهل تتبع:

```text
Frontend
↓
Go API
↓
Database
↓
Worker
```

---

# 141. Security Layers

```text
HTTPS
 ↓
Rate Limiting
 ↓
Authentication
 ↓
Authorization
 ↓
Organization Isolation
 ↓
Business Rules
 ↓
Database Constraints
 ↓
RLS
 ↓
Audit
```

لا تعتمد على طبقة واحدة.

---

# 142. Database Integrity

استخدم:

```text
Foreign Keys
Unique Constraints
Check Constraints
Not Null
Indexes
Transactions
```

ولا تعتمد فقط على Go لمنع الأخطاء.

---

# 143. Soft Delete

البيانات المرجعية يمكن أن تستخدم:

```text
deleted_at
```

لكن العمليات المالية لا تحذف.

مثلاً:

```text
Customer
Product
Supplier
```

يمكن أرشفتها.

لكن:

```text
Sale
Payment
Return
Debt
```

تظل في التاريخ.

---

# 144. Archive

بدل:

```text
Delete Product
```

استخدم:

```text
Archive Product
```

حتى لا يختفي التاريخ.

---

# 145. User Experience Rule

إذا احتاجت الميزة إلى:

> "اضغط هنا ثم اذهب إلى Settings ثم اختر Advanced ثم..."

فهذا مؤشر أن UX يحتاج إعادة تصميم.

---

# 146. User Onboarding

عند إنشاء محل:

```text
Welcome
↓
Store Name
↓
Currency
↓
Categories
↓
First Product
↓
First User
↓
Ready
```

لا تطلب 50 إعدادًا.

---

# 147. Smart Setup

يمكن للنظام إنشاء افتراضيًا:

```text
PC Components
GPU
CPU
RAM
SSD
HDD
Motherboards
PSU
Cases
Cooling
Accessories
Cables
Peripherals
```

ويستطيع صاحب المحل تعديلها.

---

# 148. First Sale

يجب أن يستطيع المستخدم الوصول إلى أول عملية بيع بسرعة.

```text
Login
↓
Dashboard
↓
Scan
↓
Sell
```

بدون إعدادات معقدة.

---

# 149. Progressive Complexity

المستخدم الجديد يرى:

```text
Basic
```

المستخدم المتقدم يستطيع فتح:

```text
Advanced
```

وهكذا النظام لا يفرض تعقيدًا على الجميع.

---

# 150. Design Philosophy

التصميم:

```text
Professional
Clean
Minimal
Fast
Dense where useful
Spacious where needed
```

وليس:

```text
Huge Icons
Huge Buttons
Color Everywhere
Too Many Cards
Too Many Charts
```

---

# 151. Mobile UI Philosophy

الهاتف ليس نسخة مصغرة من Desktop.

بل:

```text
Mobile Workflow
```

يجب تصميمه حول:

```text
Scan
Search
Sell
Pay
Check
```

---

# 152. Desktop UI Philosophy

Windows يعطي مساحة أكبر:

```text
Tables
Filters
Keyboard Shortcuts
Multi-column
Bulk Operations
Reports
```

---

# 153. Shared Design System

نفس:

```text
Colors
Typography
Components
Icons
Spacing
Notifications
Forms
```

على كل المنصات.

---

# 154. Accessibility

يجب دعم:

* Keyboard navigation.
* Focus states.
* Readable contrast.
* Screen-reader friendly labels.
* Touch target مناسب.
* RTL.
* Font scaling حيث أمكن.

---

# 155. MVP الحقيقي

لا تبدأ بكل شيء.

النسخة الأولى يجب أن تركز على:

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
Dashboard
Audit
```

إذا كانت هذه ممتازة، لديك Core Product حقيقي.

---

# 156. المرحلة الثانية

```text
Purchases
Suppliers
Expenses
Returns
Warranty
Inspection
Locations
Reservations
Notifications
```

---

# 157. المرحلة الثالثة

```text
Automation
Advanced Reports
Analytics
Smart Alerts
Background Worker
Forecasting
```

---

# 158. المرحلة الرابعة

```text
AI Assistant
Desktop Wrapper
Advanced Mobile
Multi Branch
E-Commerce
Accounting Integrations
```

---

# 159. Definition of Done

أي Feature لا تعتبر مكتملة إلا إذا:

```text
✓ Backend
✓ Frontend
✓ Database
✓ Validation
✓ Authorization
✓ Error Handling
✓ Loading State
✓ Empty State
✓ Mobile UX
✓ Desktop UX
✓ Tests
✓ Audit
✓ Documentation
```

---

# 160. أهم معيار للنجاح

لا تقيس نجاح المشروع بعدد:

```text
Tables
Endpoints
Components
Lines of Code
Docker Containers
```

قسه بـ:

```text
كم ثانية يحتاج الموظف لإتمام البيع؟
كم ثانية يحتاج للعثور على قطعة؟
كم مرة يضطر لإدخال نفس المعلومة؟
كم خطأ تم منعه؟
كم وقت وفر النظام؟
```

---

# 161. المقياس الحقيقي

قبل النظام:

```text
البحث عن قطعة:
5 دقائق

حساب الدين:
يدوي

حساب الربح:
يدوي

تحديث المخزون:
يدوي

معرفة القطع الراكدة:
صعب
```

بعد النظام:

```text
البحث:
ثوانٍ

الدين:
تلقائي

الربح:
تلقائي

المخزون:
تلقائي

القطع الراكدة:
تنبيه
```

---

# 162. Final Architecture

```text
                         SMART STORE
                              │
                     React + TypeScript
                              │
                         PWA / Web
                              │
              ┌───────────────┼───────────────┐
              │               │               │
           Windows          Android          iOS
              │               │               │
              └───────────────┼───────────────┘
                              │
                             HTTPS
                              │
                              ▼
                         Go Backend
                              │
       ┌──────────────────────┼──────────────────────┐
       │                      │                      │
       ▼                      ▼                      ▼
   Business                Security              Worker
    Modules                  Layer                Jobs
       │                      │                      │
       └──────────────────────┼──────────────────────┘
                              │
                              ▼
                       Supabase PostgreSQL
                              │
               ┌──────────────┼──────────────┐
               │              │              │
               ▼              ▼              ▼
              RLS            Auth          Storage
```

---

# 163. المبدأ النهائي للمشروع

التقنية الموجودة خلف النظام يمكن أن تكون معقدة:

```text
Go
PostgreSQL
Supabase
RLS
Docker
Workers
Transactions
Queues
Audit
PWA
CI/CD
```

لكن المستخدم لا يجب أن يشعر بهذا التعقيد.

بالنسبة له:

```text
Scan
↓
Sell
↓
Done
```

أو:

```text
Customer
↓
Balance
↓
Payment
↓
Done
```

أو:

```text
Scan Item
↓
Know Everything
```

---

# 164. تعريف المنتج النهائي

هذا المشروع ليس:

> برنامج مخزون.

وليس:

> برنامج مبيعات.

وليس:

> برنامج ديون.

بل:

# Smart Operating System for Computer Stores

نظام يجعل صاحب المحل يرى عمله بالكامل من مكان واحد، بينما يقوم النظام في الخلفية بربط العمليات وحساب النتائج واكتشاف المشاكل وتنبيه المستخدم.

---

# 165. العبارة التي يجب أن تحكم التطوير

> **لا تجعل المستخدم يعمل من أجل النظام.**

بل:

> **اجعل النظام يفهم طريقة عمل المستخدم ويقوم بالأعمال المتكررة نيابةً عنه.**

كل Feature جديدة يجب أن تمر بهذا الاختبار:

```text
هل توفر وقتًا؟
هل تقلل إدخال البيانات؟
هل تمنع خطأ؟
هل تجعل القرار أسهل؟
هل تخفي التعقيد؟
```

إذا كانت الإجابة لا:

> لا تضفها لمجرد أنها تقنية.

---

# 166. النتيجة

إذا تم تنفيذ هذه المعمارية بشكل جيد، ستكون النتيجة منصة:

```text
✓ Web
✓ Windows Friendly
✓ Mobile Friendly
✓ PWA
✓ Barcode Ready
✓ Camera Ready
✓ Dockerized
✓ Go Backend
✓ React + TypeScript
✓ Supabase/PostgreSQL
✓ Multi-Tenant
✓ Secure
✓ Auditable
✓ Transaction Safe
✓ Offline-Aware
✓ Automation Ready
✓ AI Ready
✓ SaaS Ready
```

لكن الأهم:

```text
                    SIMPLE FOR USER
                           ▲
                           │
                           │
                    COMPLEX UNDER HOOD
                           │
                           ▼
                  Go + PostgreSQL
                  Supabase + Docker
                  Security + Automation
```

وهذا هو التصميم الذي يجب أن تستهدفه:

# **واجهة بسيطة جدًا للمستخدم، وبنية قوية جدًا خلفها.**

لأن صاحب المحل لا يهتم بأن النظام يستخدم Go أو React أو Supabase.

هو يهتم بأن يدخل إلى المحل صباحًا، يفتح النظام، ويعرف خلال ثوانٍ:

> **ماذا لدي؟ ماذا بعت؟ كم ربحت؟ من عليه دين؟ وما الذي يحتاج انتباهي الآن؟**

إذا استطاع النظام الإجابة عن هذه الأسئلة تلقائيًا، فقد نجحت في بناء المنتج الذي تريده.

