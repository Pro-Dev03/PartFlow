بالتأكيد. وبناءً على **`report.md` + `report-backend.md`** الموجودين في مشروع PartFlow، سأتعامل مع الفرونت إند هنا كأنه **غير موجود إطلاقًا**، وأضع مواصفات Frontend كاملة من الصفر حتى الإنتاج، وليس مجرد قائمة صفحات.

راجعت أيضًا المستودع الحالي؛ المشروع محدد حاليًا بـ React 18 + TypeScript + Tailwind + Zustand + React Query + i18next، والـBackend موصوف كـ Go Modular Monolith مع PostgreSQL/Supabase وواجهة API تحت `/api/v1`. ([GitHub][1])

# PartFlow

## Frontend Architecture & UX Engineering Specification

### Smart Computer Store Management System

> **الهدف:** بناء Frontend يجعل صاحب المحل يدير عمله بأقل عدد ممكن من الخطوات، بينما يتولى النظام الحسابات، الترابط، التحديثات، التنبيهات، والتتبع في الخلفية.

---

# 1. الهدف من هذا التقرير

هذا التقرير يحدد **الفرونت إند الكامل لـ PartFlow من الصفر إلى الإنتاج**.

لا يفترض وجود أي واجهة سابقة.

ولا يفترض أن المستخدم يعرف كيف يعمل النظام.

الواجهة يجب أن تُبنى حول الـBackend والـBusiness Model الموجودين في `report.md` و`report-backend.md`، وليس العكس.

المبدأ الأساسي:

> **The system works for the owner — not the owner for the system.**

وهذا يعني أن Frontend ليس مجرد شاشة لعرض بيانات الـAPI.

بل هو:

**Business Operating Interface**

تُخفي التعقيد الموجود في الـBackend وتعرض للمستخدم فقط ما يحتاجه لإنجاز عمله.

التقرير الأساسي للمشروع يحدد صراحة أن النظام يجب أن يربط المنتجات والمخزون والمشتريات والفحص والتخزين والمبيعات والمدفوعات والديون والضمانات والمرتجعات والأرباح والتقارير، بينما لا يرى المستخدم هذه التعقيدات الداخلية. ([GitHub][2])

---

# 2. أهم قرار في تصميم Frontend

لن نبني:

```text
Dashboard
Products
Inventory
Sales
Customers
Reports
...
```

ثم نضع CRUD داخل كل صفحة.

هذا سيكون خطأ.

سنبدأ من:

```text
كيف يعمل المحل؟
        ↓
ما العمليات المتكررة؟
        ↓
ما الذي يحتاجه صاحب المحل فورًا؟
        ↓
ما الذي يستطيع النظام استنتاجه؟
        ↓
ما الذي يجب أن يظهر للمستخدم؟
```

ثم نبني الواجهة.

---

# 3. تجربة المستخدم الأساسية

يجب أن يشعر صاحب المحل أن PartFlow يفهم سياق العمل.

مثلاً:

## بيع قطعة

المطلوب من المستخدم:

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
Search
↓
Inventory
↓
Select Product
↓
Select Item
↓
Customer
↓
Price
↓
Payment
↓
Confirm
↓
Save
```

الـBackend نفسه يحدد أن تقليل الخطوات هو قاعدة UX أساسية، وأن عملية البيع المثالية يجب أن تختصر إلى Scan → Confirm → Payment → Done. ([GitHub][2])

---

# 4. هوية الواجهة

## التصميم العام

PartFlow يجب أن يكون:

* Modern SaaS
* Professional
* Clean
* Fast
* Practical
* Business-oriented
* Dense enough for desktop
* Touch-friendly
* Responsive
* RTL/LTR
* Dark Mode
* Accessible

لكن:

> **لا نريد واجهة "استعراضية".**

هذا نظام يستخدمه موظف محل طوال اليوم.

لذلك الأولوية:

```text
Clarity
↓
Speed
↓
Accuracy
↓
Efficiency
↓
Visual polish
```

وليس:

```text
Animations
↓
Gradients
↓
Fancy cards
↓
Complex dashboards
```

---

# 5. المستخدمون

الواجهة يجب أن تدعم:

### Owner

يرى:

* كل شيء.
* الأرباح.
* المخزون.
* الديون.
* التقارير.
* المصروفات.
* الموردين.
* الموظفين.
* Audit.

### Manager

يرى العمليات الإدارية والتشغيلية حسب الصلاحيات.

### Employee

التركيز:

```text
Sales
Inventory Search
Barcode
Customers
```

ولا يرى العمليات الحساسة التي لا يملك صلاحيتها.

### Accountant

التركيز:

```text
Payments
Debts
Expenses
Reports
Financial Data
```

الـBackend يفرض Permissions وليس Role فقط، لذلك يجب أن تكون الواجهة Permission-aware أيضًا. ([GitHub][2])

---

# 6. Architecture

الـFrontend المقترح:

```text
React 18
TypeScript
Vite
Tailwind CSS
React Query
Zustand
React Router
i18next
PWA
```

وهو متوافق مع التقنية المحددة حاليًا للمشروع في المستودع. ([GitHub][1])

---

# 7. Frontend Architecture

الهيكل:

```text
frontend/
│
├── src/
│
│   ├── app/
│   │   ├── App.tsx
│   │   ├── providers/
│   │   ├── router/
│   │   ├── config/
│   │   └── permissions/
│
│   ├── layouts/
│   │   ├── AppLayout/
│   │   ├── AuthLayout/
│   │   ├── PublicLayout/
│   │   └── MobileLayout/
│
│   ├── components/
│   │   ├── ui/
│   │   ├── forms/
│   │   ├── tables/
│   │   ├── dialogs/
│   │   ├── feedback/
│   │   ├── navigation/
│   │   └── business/
│
│   ├── features/
│   │
│   │   ├── auth/
│   │   ├── dashboard/
│   │   ├── products/
│   │   ├── inventory/
│   │   ├── barcode/
│   │   ├── sales/
│   │   ├── customers/
│   │   ├── debts/
│   │   ├── suppliers/
│   │   ├── purchases/
│   │   ├── expenses/
│   │   ├── returns/
│   │   ├── warranties/
│   │   ├── inspections/
│   │   ├── reports/
│   │   ├── notifications/
│   │   └── settings/
│
│   ├── services/
│   │   ├── api/
│   │   ├── auth/
│   │   └── storage/
│
│   ├── stores/
│   │   ├── authStore.ts
│   │   ├── uiStore.ts
│   │   ├── scannerStore.ts
│   │   └── organizationStore.ts
│
│   ├── hooks/
│   ├── types/
│   ├── utils/
│   ├── i18n/
│   ├── styles/
│   └── main.tsx
│
├── public/
├── package.json
├── vite.config.ts
└── tsconfig.json
```

وهذا يتماشى مع الهيكل المقترح للـFrontend في تقرير الـBackend، الذي يفصل features مثل auth/dashboard/products/inventory/sales/customers/debts/suppliers/purchases/expenses/reports/settings. ([GitHub][2])

---

# 8. قاعدة مهمة جدًا: Server State ≠ UI State

لن نضع كل شيء في Zustand.

## React Query

لـ:

* Products
* Inventory
* Customers
* Sales
* Debts
* Suppliers
* Purchases
* Reports
* Dashboard
* Notifications

## Zustand

لـ:

* Sidebar state
* Theme
* Scanner state
* Current organization
* UI preferences
* POS temporary UI state إذا لزم

البيانات القادمة من الـBackend يجب أن تكون Server State.

---

# 9. API Layer

كل التواصل مع الـBackend يكون من خلال طبقة API موحدة.

مثلاً:

```text
services/api/
```

وتتعامل مع:

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

وهي الـAPI domains المحددة في `report-backend.md`. ([GitHub][2])

---

# 10. API Response Handling

الـBackend يستخدم Response موحدًا:

```json
{
  "success": true,
  "data": {},
  "meta": {},
  "error": null
}
```

والـFrontend يجب أن يبني API Client يفهم هذا الشكل مركزيًا. ([GitHub][2])

عند الخطأ:

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

الواجهة لا تعرض:

```text
500 Internal Server Error
```

بل:

> **هذه القطعة تم بيعها بالفعل.**

مع الإجراء المناسب:

```text
[عرض القطعة]
```

أو:

```text
[تحديث]
```

---

# 11. Authentication

صفحات:

```text
/login
/forgot-password
/reset-password
```

بعد تسجيل الدخول:

```text
User
↓
Organization
↓
Role
↓
Permissions
↓
Application
```

وهو نفس نموذج الصلاحيات المحدد في الـBackend. ([GitHub][2])

---

# 12. Login UX

لا نحتاج شاشة تسجيل دخول معقدة.

```text
PartFlow

إدارة محلك بطريقة أذكى

البريد الإلكتروني
[________________]

كلمة المرور
[________________]

[ تسجيل الدخول ]

نسيت كلمة المرور؟
```

مع:

* Loading state
* Error state
* Session restoration
* Redirect حسب الصلاحية

---

# 13. App Shell

بعد الدخول:

```text
┌──────────────────────────────────────────────┐
│ TopBar                                       │
├──────────────┬───────────────────────────────┤
│              │                               │
│ Sidebar      │        Main Content           │
│              │                               │
│              │                               │
│              │                               │
└──────────────┴───────────────────────────────┘
```

لكن الـSidebar لا يكون ضخمًا.

---

# 14. Navigation

المستوى الأول:

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

وهذا مطابق للمبدأ الموجود في تقرير الـBackend: القائمة الأساسية صغيرة ولا تعرض عشرات الخيارات في المستوى الأول. ([GitHub][2])

---

# 15. Smart Navigation

بعض الأشياء لا تحتاج إلى صفحة مستقلة في القائمة.

مثلاً:

```text
Product
↓
Item
↓
Inspection
↓
Warranty
```

يمكن أن تكون داخل صفحة القطعة.

لا نحول كل Entity إلى Menu Item.

---

# 16. Top Bar

يحتوي على:

```text
[ Search ]

[ Scan ]

[ Notifications ]

[ Organization ]

[ User ]
```

والبحث والـScanner يجب أن يكونا دائمين وسريعين.

---

# 17. Global Search

بحث واحد يستطيع البحث في:

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

وهو مطلوب صراحة في الـBackend specification. ([GitHub][2])

مثلاً:

```text
RTX3060
```

النتيجة:

```text
RTX 3060 ASUS

3 Items
2 Used
1 New

[فتح]
```

---

# 18. Global Barcode

زر:

```text
SCAN
```

يظهر بشكل دائم.

يمكن أن يعمل مع:

* USB Scanner
* Camera
* Keyboard input

والـBackend specification يحدد دعم هذه الأجهزة، خصوصًا بيئة Windows. ([GitHub][2])

---

# 19. Context-Aware Scanner

هذه من أهم خصائص PartFlow.

إذا كان المستخدم داخل:

### Sales

المسح:

```text
Barcode
↓
Add to Cart
```

إذا كان داخل:

### Inventory

المسح:

```text
Barcode
↓
Open Item
```

إذا كان Barcode غير معروف:

```text
Barcode Unknown

[Create Product]
```

هذا مذكور صراحة في الـUX architecture الخاصة بالمشروع. ([GitHub][2])

---

# 20. Dashboard

الـDashboard ليس Report.

الـDashboard يجيب:

> **ماذا يحدث الآن؟**

وليس:

> ماذا حدث خلال السنة؟

وهذا الفصل موجود في تقرير الـBackend:

```text
Dashboard = ماذا يحدث الآن؟
Reports = ماذا حدث؟
Analytics = لماذا حدث؟
Automation = ماذا يجب أن يحدث؟
AI = ماذا يعني ذلك؟
```

([GitHub][2])

---

# 21. Dashboard النهائي

أعلى الصفحة:

```text
صباح الخير 👋

إليك أهم ما يحدث في محلك اليوم.
```

ثم:

```text
المبيعات اليوم
₪ 7,450

الربح اليوم
₪ 1,850

قيمة المخزون
₪ 185,400

ديون العملاء
₪ 24,800
```

---

# 22. قسم "يحتاج انتباهك"

هذا أهم قسم في الصفحة.

مثلاً:

```text
يحتاج انتباهك

⚠ 4 منتجات منخفضة المخزون
   [عرض المنتجات]

⚠ 3 عملاء لديهم ديون متأخرة
   [عرض الديون]

⚠ 2 قطع مستعملة لم يتم فحصها
   [فحص القطع]

⚠ ضمان ينتهي خلال 7 أيام
   [عرض الضمانات]
```

الـBackend specification يعتبر هذا القسم أهم من عشرات الرسوم البيانية. ([GitHub][2])

---

# 23. Smart Actions

في Dashboard:

```text
+ بيع

+ إضافة قطعة

+ إضافة عميل

+ تسجيل دفعة

+ إضافة مصروف
```

وهي العمليات اليومية الأكثر استخدامًا. ([GitHub][2])

---

# 24. Dashboard API

لا نريد أن يفتح Dashboard ويطلب 15 API requests.

الـBackend يوفر:

```text
GET /api/v1/dashboard
```

ويعيد:

```text
sales
profit
inventory
debts
low_stock
alerts
top_products
```

لذلك يجب أن يستخدم Frontend هذه الاستجابة الواحدة لبناء الصفحة. ([GitHub][2])

---

# 25. Inventory

صفحة:

```text
/ inventory
```

ليست مجرد جدول.

أعلى الصفحة:

```text
المخزون

[ Search ]

[ Scan ]

[ + إضافة ]

[ Filters ]
```

ثم:

```text
إجمالي القطع
1,284

قيمة المخزون
₪185,400

منخفض المخزون
12

مستعمل
84
```

---

# 26. Inventory Views

يمكن التبديل بين:

```text
Products
Items
Movements
Locations
```

لكن بدون إرهاق المستخدم.

---

# 27. Product vs Item

يجب أن تكون الواجهة واضحة جدًا بشأن الفرق.

## Product

مثلاً:

```text
RTX 3060 ASUS
```

هو النموذج العام.

## Item

مثلاً:

```text
ITEM-000421
Serial: XXXX
Used
```

هي القطعة الفعلية.

الـBackend specification يطلب الفصل بين Product وInventory Item. ([GitHub][2])

---

# 28. Product List

Desktop:

```text
Product | SKU | Condition | Stock | Price | Status
```

Mobile:

```text
RTX 3060 ASUS

₪1,150
Stock: 3
Used

[View Details]
```

ولا يجب جعل الهاتف عبارة عن جدول أفقي قابل للتمرير.

هذا مذكور صراحة ضمن responsive table rules. ([GitHub][2])

---

# 29. Product Details

صفحة:

```text
/products/:id
```

تحتوي:

```text
Product Header

RTX 3060 ASUS
GPU
Used / New

[Edit]
[Scan]
[Add Stock]
```

ثم:

```text
Overview
Inventory
Items
Sales
Purchases
Warranty
Activity
```

---

# 30. Item Details

صفحة القطعة الفردية:

```text
RTX 3060 ASUS

ITEM-000421

Used

Available
```

ثم:

```text
Barcode
Serial
Purchase Cost
Selling Price
Supplier
Location
Warranty
Condition
```

ثم Timeline:

```text
Purchased
↓
Inspected
↓
Stored
↓
Sold
```

---

# 31. Used Item Inspection

عند إضافة قطعة مستعملة:

```text
فحص القطعة

Power Test
[ ✓ ]

Temperature Test
[ ✓ ]

Performance Test
[ ✓ ]

Ports Test
[ ✓ ]

Visual Inspection
[ ✓ ]

Serial Verification
[ ✓ ]
```

ثم:

```text
Inspection Status

PASSED
```

ويظهر:

```text
Inspected by:
Employee Name

Date:
18/08/2026
```

---

# 32. Condition UI

الحالة يجب أن تكون واضحة بصريًا:

```text
New
Used
Refurbished
Needs Repair
Parts Only
```

ولا نعتمد على اللون فقط.

يجب أن يكون هناك:

* Label
* Icon عند الحاجة
* Text
* Color semantic

---

# 33. Locations

داخل Inventory:

```text
Warehouse A
Shelf B
Box 04
```

وعند فتح Item:

```text
Location

Warehouse A
Shelf B
Box 04

[Find Location]
```

---

# 34. Stock Movements

لا نعرض:

```text
Stock: 7
```

فقط.

بل:

```text
Purchase       +10
Sale            -2
Sale            -1
Return          +1
Adjustment      -1

Current Stock:
7
```

لأن الـBackend مبني على Event/Ledger model وليس مجرد تخزين الرقم النهائي. ([GitHub][2])

---

# 35. Sales / POS

هذه أهم شاشة تشغيلية.

يجب تصميمها لتكون:

# Barcode First

وليس Product Catalog First.

---

# 36. POS Layout

Desktop:

```text
┌─────────────────────────────┬──────────────────┐
│                             │                  │
│ Scan / Search               │ Cart             │
│                             │                  │
│ [ Scan Barcode ]            │ RTX 3060         │
│                             │ ₪1,150           │
│                             │                  │
│ Product Results             │ Keyboard         │
│                             │ ₪200             │
│                             │                  │
│                             │------------------│
│                             │ Total ₪1,350     │
│                             │                  │
│                             │ [ Checkout ]     │
└─────────────────────────────┴──────────────────┘
```

---

# 37. POS Workflow

```text
Scan
↓
Identify
↓
Check availability
↓
Add to cart
↓
Customer optional/required حسب الحالة
↓
Payment
↓
Confirm
↓
Receipt
↓
Done
```

---

# 38. No Duplicate Input

إذا عرف النظام:

```text
Product
Price
Cost
Stock
Item
```

لا يطلبها مرة أخرى.

هذه قاعدة أساسية في `report-backend.md`:

> **Zero Unnecessary Input**

إذا كان النظام يستطيع استنتاج المعلومة من العملية، لا يطلبها من المستخدم مرة ثانية. ([GitHub][2])

---

# 39. Payment

الواجهة:

```text
المجموع

₪ 1,350

طريقة الدفع:

○ نقد
○ بطاقة
○ تحويل
○ دين
```

يمكن لاحقًا دعم:

```text
Split Payment
```

إذا كان الـBackend يسمح بذلك.

---

# 40. Debt Sale

إذا اختار:

```text
دين
```

لا يفتح النظام 5 صفحات.

بل:

```text
Customer
[ Search customer ]

Total
₪1,350

Paid Now
₪500

Remaining
₪850

[ Confirm Sale ]
```

بعد التأكيد:

```text
Sale
+
Payment
+
Customer Ledger
+
Debt
+
Inventory
+
Profit
+
Audit
```

عملية واحدة.

---

# 41. نجاح البيع

بعد نجاح العملية:

```text
✓ تمت عملية البيع

RTX 3060
₪1,150

Paid:
₪1,000

Remaining:
₪150

Invoice #INV-1042
```

ثم:

```text
[ طباعة الفاتورة ]
[ إرسال ]
[ عملية بيع جديدة ]
```

لا نرسل المستخدم إلى Dashboard بلا سبب.

---

# 42. حماية من Double Click

الـBackend يدعم Idempotency لمنع تكرار العمليات عند:

* Double Click
* Network Retry
* Browser Retry
* Mobile connection

خصوصًا sales/payments/returns. ([GitHub][2])

والFrontend يجب أن يعكس ذلك:

```text
Confirming...
```

مع تعطيل الزر أثناء الطلب.

---

# 43. Customers

صفحة:

```text
/ customers
```

أعلى:

```text
العملاء

[ Search ]

[ + إضافة عميل ]
```

إحصائيات:

```text
Total Customers
Active Customers
Customers With Debt
Total Outstanding
```

---

# 44. Customer Profile

```text
Ahmed

Phone
05XXXXXXXX

Outstanding
₪2,500
```

ثم:

```text
Overview
Sales
Payments
Debt
Returns
Activity
```

---

# 45. Customer Financial Timeline

مثال:

```text
18 Aug
Sale +₪2,000

18 Aug
Payment -₪500

20 Aug
Payment -₪1,000

Balance:
₪500
```

وهذا يتطابق مع النموذج المالي في الـBackend. ([GitHub][2])

---

# 46. Debts

صفحة مستقلة:

```text
الديون
```

Cards:

```text
إجمالي الديون
₪24,800

متأخرة
₪8,200

مستحقة
₪16,600
```

ثم قائمة:

```text
Customer
Outstanding
Due Date
Status
Action
```

---

# 47. Debt Detail

```text
Ahmed

Outstanding
₪2,500

[ تسجيل دفعة ]
```

ثم Ledger.

وعند تسجيل دفعة:

```text
Amount
₪500

Method
Cash

[ Confirm Payment ]
```

بعدها:

```text
Outstanding
₪2,000
```

---

# 48. Financial Immutability

لا تعرض:

```text
Delete Payment
```

لأن الـBackend لا يحذف المعاملات المالية.

بدلًا من ذلك:

```text
Reverse Payment
```

مثلاً:

```text
Payment +500

↓ Reverse

Payment Reversal -500
```

وهذا مبدأ أساسي في الـBackend. ([GitHub][2])

---

# 49. Suppliers

صفحة:

```text
/suppliers
```

تحتوي:

```text
Supplier
Phone
Purchases
Paid
Outstanding
Last Purchase
```

---

# 50. Supplier Profile

```text
Supplier A

Total Purchases
₪82,000

Paid
₪70,000

Outstanding
₪12,000
```

ثم:

```text
Purchases
Payments
Products
Activity
```

---

# 51. Purchases

Workflow:

```text
New Purchase
↓
Select Supplier
↓
Add / Scan Items
↓
Purchase Cost
↓
Condition
↓
Serial / Barcode
↓
Location
↓
Confirm
```

بعد التأكيد:

```text
Inventory +
Supplier Ledger +
Cost +
Audit
```

---

# 52. Add Product UX

لا نريد Form من 30 حقلًا.

يجب أن تكون العملية ذكية.

المرحلة الأولى:

```text
Scan Barcode

أو

Search Product
```

إذا المنتج موجود:

```text
Product Found

Add Quantity
Purchase Cost
```

إذا غير موجود:

```text
New Product
```

ثم تظهر الحقول اللازمة فقط.

---

# 53. Progressive Disclosure

مثلاً:

إذا اختار:

```text
Condition: New
```

لا تظهر كل حقول الفحص.

إذا اختار:

```text
Condition: Used
```

تظهر:

```text
Condition Grade
Inspection
Serial
Notes
Photos
```

وهذا يقلل حجم الـForm.

---

# 54. Expenses

صفحة:

```text
/expenses
```

Cards:

```text
This Month
₪12,400

Rent
₪4,000

Salaries
₪5,000

Utilities
₪1,200

Other
₪2,200
```

ثم:

```text
[ + Add Expense ]
```

---

# 55. Returns

صفحة:

```text
/returns
```

لا نبدأ Return من الصفر.

يجب البحث عن:

```text
Invoice
Customer
Item
```

ثم:

```text
Sale Found

RTX 3060
₪1,150

[ Return ]
```

---

# 56. Return Workflow

```text
Sale
↓
Return Request
↓
Inspection
↓
Approve / Reject
↓
Inventory Adjustment
↓
Financial Adjustment
```

وهو نفس الـworkflow المحدد في `report.md`. ([GitHub][3])

---

# 57. Warranty

صفحة:

```text
/warranties
```

Cards:

```text
Active
Expiring Soon
Expired
```

مثلاً:

```text
RTX 3060

Customer:
Ahmed

Expires:
25/08/2026

7 days remaining
```

---

# 58. Notifications

لا نريد Notification Center مليئًا بكل حدث.

الـBackend يحدد أن الإشعارات يجب أن تكون للأشياء التي تحتاج انتباه المستخدم فقط. ([GitHub][2])

أنواع مهمة:

```text
LOW_STOCK
OVERDUE_DEBT
WARRANTY_EXPIRING
ITEM_RESERVED
PAYMENT_RECEIVED
PURCHASE_RECEIVED
```

---

# 59. Notification UX

Badge:

```text
🔔 4
```

عند الفتح:

```text
يحتاج انتباهك

4 منتجات منخفضة
3 ديون متأخرة
1 ضمان ينتهي
```

كل إشعار قابل للنقر ويذهب مباشرة إلى المكان المناسب.

---

# 60. Reports

التقارير يجب ألا تكون مجرد Data Tables.

يجب أن تجيب عن الأسئلة.

---

# 61. Reports Home

```text
التقارير

المبيعات
الأرباح
المخزون
الديون
المنتجات
الموردون
المصروفات
```

---

# 62. Sales Report

Filters:

```text
Today
This Week
This Month
Custom
```

ثم:

```text
Revenue
Orders
Items Sold
Average Sale
```

ثم التفاصيل.

---

# 63. Profit Report

```text
Revenue
₪100,000

COGS
₪72,000

Gross Profit
₪28,000

Expenses
₪8,000

Net Profit
₪20,000
```

ويجب أن يكون كل رقم قابلًا للفتح لمعرفة مصدره.

الـBackend يحدد أن كل رقم مهم يجب أن يكون **Explainable**. ([GitHub][2])

---

# 64. Inventory Report

```text
Inventory Value

By Category
By Condition
By Brand
By Location
```

ثم:

```text
Low Stock
Dead Stock
Fast Moving
```

---

# 65. Dead Stock

مثلاً:

```text
Products not sold for 90+ days

RTX 2080
Last Sale:
112 days ago

GTX 1080
Last Sale:
143 days ago
```

ثم:

```text
[ View Item ]
```

---

# 66. Used Products Report

```text
Used Items Sold
Used Revenue
Used Cost
Used Profit
Average Margin
```

وهذا مهم لأن المنتجات المستعملة عنصر أساسي في PartFlow وليس مجرد Product Type بسيط. ([GitHub][3])

---

# 67. Reports Export

كل تقرير يحتاج:

```text
Export
```

والخيارات تعتمد على ما يدعمه الـBackend لاحقًا:

```text
PDF
CSV
Print
```

لكن التصدير الكبير يمكن أن يكون Background Job لأن الـBackend يحدد أن التقارير الكبيرة لا يجب أن تنفذ داخل HTTP request. ([GitHub][2])

---

# 68. Settings

لا نضع كل شيء في صفحة واحدة ضخمة.

الأقسام:

```text
Organization
Profile
Users & Permissions
Store
Inventory
Sales
Payments
Notifications
Localization
Appearance
Security
Audit
```

---

# 69. Organization

```text
Store Name
Phone
Email
Address
Timezone
Currency
```

الـBackend يستخدم Timezone للمؤسسة، مع مثال `Asia/Jerusalem`. ([GitHub][2])

---

# 70. Localization

PartFlow يجب أن يدعم من البداية:

```text
العربية
עברית
English
```

مع:

```text
RTL
LTR
```

وهذا منصوص عليه في المواصفات. ([GitHub][2])

---

# 71. RTL/LTR Architecture

لا نكتب:

```css
margin-left
margin-right
```

بشكل عشوائي.

يفضل استخدام logical properties:

```css
margin-inline-start
margin-inline-end
padding-inline
inset-inline
```

حتى تعمل الواجهة بشكل صحيح في:

```text
Arabic RTL
Hebrew RTL
English LTR
```

---

# 72. Currency

الواجهة يجب أن تتعامل مع:

```text
ILS
₪
```

وفق إعداد المؤسسة.

ولا تقوم بأي حساب مالي حساس في JavaScript باستخدام floating-point.

الـBackend هو مصدر الحقيقة المالية، والـFrontend يعرض النتائج. الـBackend specification يشدد على Decimal/Minor Units بدل Floating Point للحسابات المالية. ([GitHub][2])

---

# 73. Date & Time

التخزين من الـBackend موحد.

لكن العرض حسب:

```text
Organization Timezone
```

مثلاً:

```text
18/08/2026
14:32
```

---

# 74. Loading States

لا نستخدم:

```text
Loading...
```

في كل مكان.

بل:

```text
Skeleton
```

للصفحات والجداول والبطاقات.

وهذا مطلوب في مواصفات الـBackend/UX. ([GitHub][2])

---

# 75. Empty States

لا نعرض:

```text
No Data
```

فقط.

بل:

```text
لا توجد قطع حتى الآن.

ابدأ بإضافة أول قطعة إلى مخزونك.

[ + إضافة قطعة ]
```

وهذا أيضًا منصوص عليه في التقرير. ([GitHub][2])

---

# 76. Error States

لا نعرض:

```text
500 Internal Server Error
```

للمستخدم.

نعرض:

```text
حدث خطأ أثناء تحميل البيانات.

[ إعادة المحاولة ]
```

أما التفاصيل التقنية فتذهب إلى Logs. ([GitHub][2])

---

# 77. Confirmation UX

لا نطلب:

> Are you sure?

عند كل نقرة.

Confirmation فقط للعمليات الخطرة مثل:

```text
Delete
Refund
Inventory Adjustment
Cancel Sale
```

كما يحدد التقرير. ([GitHub][2])

---

# 78. Smart Defaults

مثلاً:

إذا كان:

```text
Cash
```

هو الأكثر استخدامًا:

```text
Payment Method:
Cash
```

محدد مسبقًا.

وعند إضافة منتج:

```text
Currency:
ILS
```

حسب إعداد المؤسسة. ([GitHub][2])

---

# 79. Permission-Aware UI

إذا لم يمتلك المستخدم:

```text
products.update
```

لا نعرض:

```text
Edit Product
```

وإذا لم يمتلك:

```text
sales.refund
```

لا نعرض:

```text
Refund
```

لكن:

> **Frontend permissions ليست طبقة الأمان الأساسية.**

الـBackend وRLS هما مصدر الحماية الحقيقي. ([GitHub][2])

---

# 80. Audit UI

Owner/authorized users يمكنهم رؤية:

```text
Activity
```

مثال:

```text
Employee changed selling price

RTX 3060

Old:
₪1,200

New:
₪1,150

18/08/2026
14:32
```

لأن الـBackend يسجل:

```text
Who
What
When
Target
Before
After
Result
```

([GitHub][2])

---

# 81. AI Interface — مستقبلًا

AI لا يدخل في Core Transaction UI.

بل يكون Layer فوق النظام:

```text
Database
↓
Analytics
↓
AI Assistant
```

كما يحدد التقرير. ([GitHub][2])

واجهة مستقبلية:

```text
اسأل PartFlow

"كم ربحت من GPU المستعملة هذا الشهر؟"

[ اسأل ]
```

والإجابة:

```text
حققت هذا الشهر:

Revenue: ₪18,400
Cost: ₪13,200
Profit: ₪5,200
Margin: 28.3%
```

لكن AI لا يخترع البيانات.

---

# 82. Smart Assistant

حتى قبل AI، يمكن للنظام إنشاء Insights حقيقية من الـBackend:

```text
💡 مبيعات القطع المستعملة ارتفعت هذا الشهر.

⚠ RTX 3060 وصل للحد الأدنى.

⚠ يوجد 3 عملاء لديهم ديون متأخرة.

📦 يوجد 7 منتجات لم تتحرك منذ 90 يومًا.
```

وهذه تعتمد على Automation/Smart Insights في الـBackend. ([GitHub][2])

---

# 83. Offline Awareness

بما أن المشروع PWA، يجب أن يعرف المستخدم حالة الاتصال.

مثلاً:

```text
● متصل
```

أو:

```text
● الاتصال ضعيف
```

بدل رسائل تقنية مخيفة.

الـBackend specification يطلب أن يعرف المستخدم حالة الاتصال مع رسائل بسيطة. ([GitHub][2])

---

# 84. لا نعد Offline الكامل لكل شيء

هذه نقطة مهمة.

لا ينبغي أن نجعل Frontend يتظاهر بأنه يستطيع تنفيذ كل العمليات المالية Offline.

العمليات الحساسة مثل:

```text
Sale
Payment
Refund
Debt
Inventory
```

تحتاج مصدر حقيقة من الـBackend.

لذلك Offline UX يجب أن يكون واضحًا بشأن ما يمكن فعله وما ينتظر الاتصال.

---

# 85. PWA

PartFlow يجب أن يكون:

```text
Installable
Responsive
Fast
Offline-aware
```

ويدعم:

```text
Desktop
Tablet
Mobile
```

والـREADME الحالي يحدد بالفعل PWA وResponsive وDark Mode ضمن أهداف الواجهة. ([GitHub][1])

---

# 86. Mobile

الهاتف ليس Desktop مصغرًا.

يجب إعادة ترتيب التجربة.

مثلاً Dashboard:

```text
Sales
₪7,450

Profit
₪1,850

Debt
₪24,800

Attention
4 Alerts
```

بدل 4 أعمدة صغيرة.

---

# 87. Mobile Navigation

يمكن استخدام:

```text
Bottom Navigation
```

للعمليات الأكثر استخدامًا:

```text
الرئيسية
المبيعات
المخزون
العملاء
المزيد
```

مع:

```text
SCAN
```

كإجراء سريع واضح.

---

# 88. Tablet

الـTablet مهم جدًا لموظف المحل.

يمكن أن يستخدم:

```text
Sidebar collapsed
+
Large touch targets
+
POS optimized
```

---

# 89. Desktop

Desktop هو بيئة الإدارة الرئيسية.

يمكن استخدام:

```text
Sidebar
+
Dense Tables
+
Keyboard Shortcuts
+
POS
```

---

# 90. Keyboard Shortcuts

مهمة جدًا في POS.

مثلاً:

```text
F2 → Search
F4 → Scan
F8 → Checkout
Esc → Close Modal
Enter → Confirm
```

يمكن تخصيصها لاحقًا.

---

# 91. Accessibility

يجب دعم:

* Keyboard navigation
* Focus states
* Screen readers
* Labels
* Contrast
* Reduced motion
* Semantic HTML
* ARIA عند الحاجة

ولا نعتمد على اللون وحده للحالة.

---

# 92. Design System

قبل بناء الصفحات يجب بناء Design System.

## Buttons

```text
Primary
Secondary
Ghost
Danger
Success
```

## Inputs

```text
Text
Search
Number
Currency
Select
Date
Date Range
Scanner
```

## Feedback

```text
Toast
Alert
Dialog
Drawer
Skeleton
Empty State
Error State
```

---

# 93. Cards

لا نستخدم Card لكل شيء.

Card تستخدم عندما يكون هناك:

```text
Summary
Grouping
Important Context
```

ولا نحول كل عنصر إلى Card لأن ذلك يجعل النظام يبدو كـDashboard template.

---

# 94. Tables

Desktop:

```text
Dense
Sortable
Filterable
Paginated
Selectable
```

Mobile:

```text
Card/List
```

وليس horizontal scrolling دائمًا.

---

# 95. Filters

الفلاتر يجب أن تكون context-aware.

Inventory:

```text
Condition
Brand
Category
Stock
Location
```

Sales:

```text
Date
Payment
Customer
Employee
```

Debts:

```text
Status
Due Date
Customer
```

---

# 96. Search Debouncing

البحث النصي:

```text
User typing
↓
Debounce
↓
API
```

لكن Barcode:

```text
Immediate
```

لأن السرعة أساسية.

---

# 97. Pagination

لا نحمل آلاف السجلات.

الـBackend يوصي باستخدام:

```text
limit
cursor
```

أو Pagination مناسبة. ([GitHub][2])

والـFrontend يجب أن يبني ذلك في Data Table component مشترك.

---

# 98. Caching

React Query يجب أن يستخدم:

```text
Query Cache
Stale Time
Invalidation
Prefetch
Optimistic UI
```

بحذر.

مثلاً بعد:

```text
Create Product
```

يتم تحديث:

```text
Products Query
Inventory Query
Dashboard Query
```

حسب الحاجة.

---

# 99. One Action → Many UI Updates

مثلاً:

```text
Sell RTX 3060
```

بعد النجاح:

```text
POS Cart
↓
Inventory
↓
Dashboard
↓
Customer Balance
↓
Sales History
↓
Notifications
```

يجب أن تصبح الواجهة متزامنة مع نتيجة العملية.

وهذا يعكس مبدأ:

> One Action, Many Updates. ([GitHub][2])

---

# 100. Backend Errors → Business UX

الـFrontend يجب أن يحول Error Codes إلى إجراءات مفهومة.

مثلاً:

```text
ITEM_NOT_FOUND
```

→

> لم يتم العثور على القطعة.

```text
ITEM_ALREADY_SOLD
```

→

> هذه القطعة لم تعد متاحة للبيع.

```text
INSUFFICIENT_STOCK
```

→

> الكمية المتوفرة غير كافية.

```text
DUPLICATE_SERIAL
```

→

> Serial Number مستخدم بالفعل.

```text
PERMISSION_DENIED
```

→

> ليس لديك صلاحية لتنفيذ هذه العملية.

هذه الأكواد محددة في مواصفات الـAPI. ([GitHub][2])

---

# 101. Security Frontend

يجب:

* عدم تخزين Secrets.
* عدم وضع Server Keys في React.
* حماية Routes.
* إخفاء UI حسب Permissions.
* التعامل الصحيح مع Session.
* تنظيف Input.
* منع تسريب بيانات حساسة في logs.
* عدم الاعتماد على Frontend authorization وحده.

الـBackend يحدد صراحة أن Server Secret Keys تبقى Backend-only. ([GitHub][2])

---

# 102. Images & Files

الواجهة تحتاج دعم:

```text
Product Images
Item Images
Inspection Photos
Invoices
Documents
```

والـBackend يخطط لـSupabase Storage مع سياسات حسب `organization_id`. ([GitHub][2])

---

# 103. Image UX

عند Product:

```text
[ Main Image ]

+ Add Photos
```

عند Used Item:

```text
Inspection Photos
```

ويمكن ربط الصور بتاريخ الفحص.

---

# 104. Store Hardware

الواجهة يجب أن تتعامل مع:

```text
USB Barcode Scanner
Camera
Printer
Receipt Printer
Label Printer
Keyboard
Mouse
Touch
```

وهي الأجهزة التي يحددها Backend specification. ([GitHub][2])

---

# 105. Barcode Printing

بعد إنشاء Barcode داخلي:

```text
FNX-GPU-000421
```

يمكن عرض:

```text
┌───────────────────┐
│ RTX 3060          │
│                   │
│ |||||||||||||||   │
│ FNX-GPU-000421    │
│                   │
│ ₪1,150            │
└───────────────────┘
```

ثم:

```text
[ Print Label ]
```

---

# 106. Global UX Principle

كل صفحة يجب أن تجيب:

### ماذا أستطيع أن أفعل هنا؟

مثلاً Inventory:

```text
ابحث
Scan
أضف
انقل
افتح
```

Sales:

```text
Scan
Add
Checkout
```

Customers:

```text
Find
Open
Sell
Collect Payment
```

Reports:

```text
Understand
Filter
Export
```

---

# 107. الصفحة لا يجب أن تكون مجرد CRUD

مثلاً:

## Product Page

ليست:

```text
Edit
Delete
Save
```

فقط.

بل:

```text
Product
↓
Stock
↓
Items
↓
Sales
↓
Purchases
↓
Warranty
↓
History
```

---

# 108. Dashboard ليس مكان كل شيء

لا نضع:

```text
10 Charts
20 Tables
15 Metrics
```

الـDashboard يجب أن يكون:

```text
Current State
+
Attention
+
Actions
```

فقط.

---

# 109. النظام الاستباقي

الواجهة يجب أن تسمح للنظام بأن يقول:

```text
"لديك مشكلة."
```

بدل أن ينتظر:

```text
المستخدم → يفتح التقرير
المستخدم → يحدد التاريخ
المستخدم → يفلتر
المستخدم → يحلل
```

مثلاً:

```text
⚠ 3 ديون متأخرة

[مراجعة]
```

ثم يذهب مباشرة إلى القائمة المناسبة.

---

# 110. Smart Action Routing

كل Insight يجب أن يحتوي على Action.

مثلاً:

```text
4 Low Stock

[Review]
```

→ Inventory filtered.

```text
3 Overdue Debts

[Review]
```

→ Debts filtered.

```text
1 Warranty Expiring

[Review]
```

→ Warranty filtered.

لا تجعل المستخدم يبحث عن المكان بنفسه.

---

# 111. Owner Mode

عند دخول Owner، Dashboard يجب أن يركز على:

```text
Revenue
Profit
Inventory Value
Debts
Expenses
Alerts
Insights
```

---

# 112. Employee Mode

عند دخول Employee:

```text
Sales
Scanner
Inventory Search
Customers
```

بدل إغراقه بالأرباح والبيانات الحساسة.

---

# 113. Manager Mode

```text
Sales
Inventory
Purchases
Customers
Reports
Alerts
```

حسب Permissions.

---

# 114. Accountant Mode

```text
Debts
Payments
Expenses
Profit
Reports
Supplier Balances
```

---

# 115. Organization Isolation

إذا كان النظام Multi-Tenant:

```text
Organization A
```

لا يمكن أن تظهر بيانات:

```text
Organization B
```

حتى في حالة وجود خطأ في UI.

الحماية الحقيقية في:

```text
Backend
+
RLS
```

وليس Frontend فقط. ([GitHub][2])

---

# 116. Performance Targets

يجب تصميم Frontend ليستهدف:

```text
Barcode Lookup
< 300ms target

Normal API
< 500ms target

Dashboard
< 2s target
```

وهذه أهداف تصميمية أولية مذكورة في التقرير. ([GitHub][2])

---

# 117. Performance Techniques

يجب استخدام:

```text
Lazy Loading
Code Splitting
Image Optimization
Pagination
React Query Cache
Prefetch
Memoization where useful
Virtualization for very large lists
```

مع تجنب:

```text
Huge JS bundle
Huge tables
Unnecessary API calls
```

وهي مبادئ متوافقة مع مواصفات الأداء الموجودة في Backend report. ([GitHub][2])

---

# 118. Routing

المسارات المقترحة:

```text
/login
/forgot-password

/app
/app/dashboard

/app/sales
/app/sales/:id

/app/inventory
/app/inventory/products
/app/inventory/items
/app/inventory/movements

/app/products/:id

/app/customers
/app/customers/:id

/app/debts
/app/debts/:id

/app/suppliers
/app/suppliers/:id

/app/purchases
/app/purchases/:id

/app/expenses

/app/returns
/app/warranties
/app/inspections

/app/reports
/app/reports/sales
/app/reports/profit
/app/reports/inventory
/app/reports/debts

/app/settings
```

---

# 119. Route Guards

كل Route يمر عبر:

```text
Authenticated?
↓
Organization selected?
↓
Permission?
↓
Render
```

---

# 120. Feature Module

كل Feature يحتوي على:

```text
feature/
├── pages/
├── components/
├── hooks/
├── api/
├── schemas/
├── types/
└── utils/
```

مثلاً:

```text
features/sales/
├── pages/
│   └── SalesPage.tsx
├── components/
│   ├── Cart.tsx
│   ├── Scanner.tsx
│   ├── Checkout.tsx
│   └── PaymentSelector.tsx
├── api/
│   └── salesApi.ts
├── hooks/
│   └── useCreateSale.ts
└── types/
    └── sales.types.ts
```

---

# 121. Shared Components

يجب عدم إعادة بناء:

```text
Button
Modal
Table
Input
Select
Toast
Drawer
Skeleton
Badge
```

في كل Feature.

كلها تكون Design System مشتركة.

---

# 122. Business Components

إضافة طبقة مختلفة عن UI Components:

```text
Money
BarcodeBadge
StockStatus
DebtStatus
ConditionBadge
PaymentStatus
WarrantyStatus
ActivityTimeline
PermissionGate
```

هذه مكونات تفهم Business Domain.

---

# 123. Form Validation

الـFrontend يتحقق من:

```text
Required
Format
Range
Invalid values
```

لكن Backend هو المصدر النهائي للتحقق.

لا نفترض أن Frontend validation كافية.

---

# 124. Forms

Forms يجب أن تكون:

```text
Short
Contextual
Progressive
Smart defaults
Keyboard friendly
```

ولا نعرض كل الحقول دفعة واحدة.

---

# 125. Unsaved Changes

إذا كان المستخدم يعدل Product ثم حاول الخروج:

```text
لديك تغييرات غير محفوظة.

[متابعة التحرير]
[الخروج بدون حفظ]
```

فقط عند الحاجة.

---

# 126. Toasts

بعد العملية:

```text
✓ تمت إضافة القطعة
```

أو:

```text
✓ تم تسجيل الدفعة
```

لكن لا نستخدم Toast للأخطاء المعقدة التي تحتاج قرارًا.

---

# 127. Drawers بدل Pages عندما يكون مناسبًا

مثلاً:

```text
Quick Add Customer
Quick Payment
Quick Expense
```

يمكن أن تكون Drawer.

لكن:

```text
Customer Profile
Reports
Inventory
```

تحتاج صفحات كاملة.

---

# 128. Modal Discipline

لا نحول التطبيق إلى:

```text
Modal داخل Modal داخل Modal
```

Modal للعمليات القصيرة.

Page للعمليات المهمة.

---

# 129. Search-first UX

عند فتح:

```text
Customers
```

يكون البحث جاهزًا.

عند:

```text
Inventory
```

يكون Search/Scan واضحًا.

لأن الموظف غالبًا يعرف ما يبحث عنه.

---

# 130. POS-first UX

في المبيعات:

```text
Scanner
```

يكون Focus تلقائيًا عند فتح POS إذا كان ذلك مناسبًا.

وبذلك يمكن للموظف:

```text
Scan
Scan
Scan
Checkout
```

بدون لمس الفأرة لكل عملية.

---

# 131. Success Flow

بعد كل عملية ناجحة، النظام يجب أن يخبر المستخدم:

```text
What happened?
```

وليس فقط:

```text
Success
```

مثلاً:

```text
تم بيع RTX 3060.

تم تحديث المخزون.
تم تسجيل الدفعة.
تم تحديث رصيد العميل.
تم احتساب الربح.
```

بدون إغراقه بالتفاصيل التقنية.

---

# 132. Explainability

إذا رأى:

```text
Profit:
₪1,850
```

يمكنه الضغط.

يفتح:

```text
Revenue
₪7,450

COGS
₪5,100

Expenses
₪500

Net Profit
₪1,850
```

هذه نقطة مهمة جدًا لأن النظام يجب أن يكون قابلًا للتفسير. ([GitHub][2])

---

# 133. لا تخفي المعلومات المهمة

UX البسيط لا يعني:

> إخفاء كل شيء.

بل:

> إظهار الشيء الصحيح في الوقت الصحيح.

مثلاً Employee لا يحتاج Profit في POS.

Owner يحتاج Profit في Dashboard.

---

# 134. Dark Mode

Dark Mode يجب أن يكون:

```text
Professional
Readable
Low contrast fatigue
```

وليس Cyberpunk.

لأن النظام تشغيلي ويستخدم لساعات طويلة.

---

# 135. Animations

Animations فقط حيث تساعد المستخدم:

```text
Page transition
Toast
Drawer
Modal
State transition
```

لا نستخدم animation لكل Card.

---

# 136. Responsive Rule

كل Feature يجب اختباره على:

```text
Desktop
Tablet
Mobile
```

وليس:

> نصلح Mobile في النهاية.

---

# 137. Testing Strategy

Frontend يجب أن يختبر:

### Unit

```text
formatMoney
formatDate
permissions
validation
business helpers
```

### Component

```text
Checkout
Cart
ProductForm
DebtPayment
```

### Integration

```text
Login
Create Product
Sell
Payment
Return
```

### E2E

أهم سيناريو:

```text
Create Product
↓
Receive Item
↓
Scan
↓
Sell
↓
Partial Payment
↓
Debt
↓
Pay Later
↓
Profit Updated
↓
Inventory Updated
```

وهو نفس الـintegration scenario الذي يحدده تقرير الـBackend. ([GitHub][2])

---

# 138. Security Testing

اختبار:

```text
User A
↓
tries Organization B
```

النتيجة:

```text
403 / No Data
```

لكن يجب أن تكون الحماية:

```text
Backend
+
Database RLS
```

وليس UI فقط. ([GitHub][2])

---

# 139. Testing Matrix

يجب اختبار:

```text
Arabic RTL
Hebrew RTL
English LTR

Desktop
Tablet
Mobile

Dark
Light

Owner
Manager
Employee
Accountant
```

---

# 140. Development Phases

## Phase 0 — Foundation

```text
React
TypeScript
Vite
Tailwind
Router
React Query
Zustand
i18n
Design System
```

---

# 141. Phase 1 — Application Shell

```text
Auth
App Layout
Sidebar
TopBar
Responsive Layout
Theme
Language
Permissions
Global Search
Scanner
```

---

# 142. Phase 2 — Dashboard

```text
Dashboard API
Metrics
Attention
Smart Actions
Recent Activity
Alerts
Responsive
```

---

# 143. Phase 3 — Products & Inventory

```text
Products
Items
Barcode
Serial
Conditions
Locations
Stock Movements
Inspection
```

---

# 144. Phase 4 — Sales

```text
POS
Scanner
Cart
Customer
Payment
Debt
Receipt
Sale History
```

---

# 145. Phase 5 — Customers & Debts

```text
Customers
Profiles
Ledger
Payments
Debt
Overdue
```

---

# 146. Phase 6 — Purchasing

```text
Suppliers
Purchases
Receiving
Inventory Update
Supplier Ledger
```

---

# 147. Phase 7 — Finance

```text
Expenses
Profit
Financial Reports
Payments
```

---

# 148. Phase 8 — Returns & Warranty

```text
Returns
Inspection
Warranty
Warranty Alerts
```

---

# 149. Phase 9 — Reports & Analytics

```text
Sales
Profit
Inventory
Debts
Products
Suppliers
Expenses
```

---

# 150. Phase 10 — Smart System

```text
Notifications
Smart Insights
Automation
Attention Center
```

---

# 151. Phase 11 — PWA & Hardware

```text
Camera Scanner
USB Scanner
Receipt Printing
Label Printing
PWA
Offline Awareness
```

---

# 152. Phase 12 — AI

بعد استقرار الـCore:

```text
AI Assistant
Natural Language Queries
Business Questions
Recommendations
```

وليس قبل ذلك.

فالـBackend report نفسه يضع AI كطبقة فوق Analytics وليس داخل Core Transaction Engine. ([GitHub][2])

---

# 153. المرحلة الأخيرة — Production

قبل الإطلاق:

```text
Build
↓
Lint
↓
Type Check
↓
Unit Tests
↓
Integration Tests
↓
E2E
↓
Security Checks
↓
Performance
↓
Docker Build
↓
Production
```

وهذا متوافق مع CI/CD flow المحدد في Backend specification. ([GitHub][2])

---

# 154. Definition of Done

لا نعتبر Feature مكتملة لأنها "تعمل".

Feature لا تعتبر Done إلا إذا:

```text
UI
✓

API
✓

Loading
✓

Empty State
✓

Error State
✓

Permissions
✓

Mobile
✓

RTL
✓

LTR
✓

Dark Mode
✓

Accessibility
✓

Tests
✓
```

---

# 155. أهم سيناريو في النظام

يجب أن يعمل من البداية إلى النهاية:

```text
Supplier
   ↓
Purchase
   ↓
Receive Item
   ↓
Barcode
   ↓
Inspection
   ↓
Inventory
   ↓
Sale
   ↓
Customer
   ↓
Payment
   ↓
Debt
   ↓
Payment Later
   ↓
Profit
   ↓
Dashboard
   ↓
Reports
```

إذا كانت هذه الدورة تعمل بشكل صحيح، فلدينا أساس قوي جدًا للنظام.

---

# 156. مثال كامل لتجربة المستخدم

وصلت RTX 3060 مستعملة.

الموظف:

```text
SCAN
```

النظام:

```text
RTX 3060

Product exists.

[Receive Item]
```

يضغط.

النظام يعرف:

```text
Supplier
Product
Category
```

الموظف يدخل فقط:

```text
Purchase Cost
Condition
Serial
```

ثم:

```text
[Confirm]
```

النظام:

```text
Create Item
+
Inventory
+
Barcode
+
Supplier Ledger
+
Audit
```

---

# 157. بعد ذلك البيع

الموظف:

```text
SCAN
```

PartFlow يعرف أن المستخدم في POS.

فيضيف القطعة تلقائيًا.

```text
RTX 3060
Used
₪1,150
```

ثم:

```text
[Checkout]
```

اختيار:

```text
Cash
```

ثم:

```text
[Complete Sale]
```

انتهى.

---

# 158. إذا كان البيع بالدين

الموظف:

```text
Checkout
↓
Debt
```

يختار العميل.

يدفع:

```text
₪500
```

النظام يسجل:

```text
Sale
₪1,150

Payment
₪500

Debt
₪650
```

ثم يحدث تلقائيًا:

```text
Inventory
Profit
Customer Ledger
Debt
Dashboard
Reports
Audit
```

وهذا بالضبط جوهر PartFlow: عملية واحدة تؤدي إلى تحديثات متعددة دون إدخال البيانات مرة أخرى. ([GitHub][2])

---

# 159. اليوم التالي

صاحب المحل يفتح Dashboard:

```text
صباح الخير 👋

المبيعات
₪7,450

الربح
₪1,850

المخزون
₪185,400

ديون العملاء
₪24,800
```

ثم:

```text
يحتاج انتباهك

⚠ 4 منتجات منخفضة
⚠ 3 ديون متأخرة
⚠ 1 ضمان قريب الانتهاء
```

ثم:

```text
اقتراح PartFlow

RTX 3060 يتحرك بسرعة هذا الشهر.
```

هنا لا يحتاج صاحب المحل إلى:

```text
فتح Inventory
فتح Reports
حساب Stock
فتح Debts
حساب Profit
```

النظام قام بذلك مسبقًا.

---

# 160. المبدأ النهائي للتصميم

كل قرار في Frontend يجب أن يمر بهذا السؤال:

> **هل هذا يجعل حياة صاحب المحل أسهل أم يجعل النظام أسهل؟**

إذا جعل النظام أسهل على حساب المستخدم:

**نرفضه.**

إذا جعل المستخدم يقوم بخطوات يستطيع النظام تنفيذها:

**نرفضه.**

إذا جعل الصفحة أجمل ولكن أبطأ:

**نرفضه.**

إذا أضاف ميزة فقط لأنها "احترافية":

**نرفضها.**

إذا اختصر:

```text
5 خطوات → خطوتين
```

فهي ميزة حقيقية.

---

# 161. النتيجة المعمارية النهائية

PartFlow Frontend يجب أن يصبح:

```text
                  PARTFLOW FRONTEND
                         │
           ┌─────────────┴─────────────┐
           │                           │
       USER ACTION                SYSTEM STATE
           │                           │
           ↓                           ↓
       Scan / Sell /              Dashboard /
       Receive / Pay              Notifications
           │                           │
           └─────────────┬─────────────┘
                         ↓
                     Go API
                         ↓
                 Business Logic
                         ↓
                  PostgreSQL
                         ↓
                  Events / Worker
                         ↓
          ┌──────────────┼──────────────┐
          ↓              ↓              ↓
       Inventory       Finance       Insights
          │              │              │
          └──────────────┼──────────────┘
                         ↓
                    FRONTEND
                         ↓
               "What needs attention?"
```

---

# 162. الخلاصة التنفيذية

الـFrontend المطلوب لـPartFlow **ليس Dashboard + CRUD**.

بل يجب أن يكون طبقة ذكية فوق الـBackend.

الـBackend يتعامل مع:

```text
Transactions
Inventory
Ledger
Payments
Permissions
RLS
Audit
Events
Automation
```

بينما الـFrontend يتعامل مع:

```text
Intent
Context
Simplicity
Speed
Visibility
Guidance
```

وهذا الفصل مهم جدًا.

فالـBackend يمكن أن ينفذ عشرات العمليات عند بيع قطعة، لكن المستخدم يرى:

> **بيع**

والـBackend يستطيع إنشاء:

```text
Sale
Sale Items
Payment
Inventory Movement
Customer Ledger
Profit
Audit Event
Notification
Analytics Event
```

لكن المستخدم لا يحتاج إلى رؤية هذه التعقيدات. وهذا هو المبدأ المركزي في مواصفات المشروع: **Complexity Behind the Curtain**. ([GitHub][2])

---

# 163. الشكل النهائي الذي أريد أن يصل إليه PartFlow

ليس:

> "هذا برنامج فيه مخزون ومبيعات."

بل عندما يدخل صاحب المحل إلى النظام يشعر:

> **"PartFlow يعرف ماذا يحدث في محلي."**

يدخل فيجد:

```text
اليوم
──────────────

المبيعات       ₪7,450
الربح          ₪1,850
المخزون        ₪185,400
الديون         ₪24,800

يحتاج انتباهك
──────────────

⚠ 4 منتجات تحتاج إعادة طلب
⚠ 3 عملاء متأخرون
⚠ 2 قطع مستعملة تحتاج فحص
⚠ ضمان واحد ينتهي قريبًا

إجراءات سريعة
──────────────

[ + بيع ]
[ + إضافة قطعة ]
[ + تسجيل دفعة ]
[ + إضافة عميل ]
[ + مصروف ]

[ 🔍 بحث ]
[ 📷 Scan ]
```

ثم يقوم بعمله.

**لا يبحث عن المعلومات.**

**لا يحسب الأرباح يدويًا.**

**لا يتذكر الديون.**

**لا يتتبع القطعة في دفتر.**

**لا يدخل نفس البيانات مرتين.**

**لا يحتاج إلى فهم بنية النظام.**

بل:

# هو يدير المحل.

# وPartFlow يدير التفاصيل.


**هذا هو الـFrontend الذي أنصح ببنائه من الصفر، وليس محاولة ترقيع الواجهة الحالية.**

