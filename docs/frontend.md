# تقرير شامل — تصميم ومظهر وواجهة Frontend

## Smart Computer Store Management System

> **الهدف:** بناء واجهة Frontend احترافية لنظام إدارة محلات قطع الحاسوب، تجعل النظام سريعًا، بسيطًا، ذكيًا، وقابلًا للتسويق تجاريًا، مع دعم Desktop/Windows وMobile والعربية والعبرية والإنجليزية.

---

# 1. الرؤية الأساسية

المنتج لا يجب أن يبدو كبرنامج محاسبة تقليدي أو نظام ERP معقد.

بل يجب أن يعطي المستخدم هذا الشعور:

> **"أنا أدير المحل، والنظام يدير التفاصيل."**

الـFrontend يجب أن يخفي التعقيد الموجود في الخلفية.

المستخدم يرى:

```text
Scan
↓
System Understands
↓
One Action
↓
Everything Updates
```

بدل أن يضطر إلى التنقل بين عدة صفحات وإدخال نفس البيانات أكثر من مرة.

---

# 2. أهداف الـFrontend

الواجهة يجب أن تحقق:

* سهولة الاستخدام.
* سرعة العمليات.
* وضوح المعلومات.
* تقليل عدد النقرات.
* تقليل أخطاء الموظفين.
* دعم Barcode.
* دعم القطع الفردية والمستعملة.
* دعم Windows/Desktop.
* دعم الهواتف.
* دعم Touch.
* دعم Keyboard.
* دعم RTL/LTR.
* دعم العربية والعبرية والإنجليزية.
* Responsive Design حقيقي.
* Accessibility.
* Performance عالي.
* تصميم قابل للتوسع.
* Design System موحد.
* مظهر Premium قابل للتسويق.

---

# 3. المبدأ الأساسي

## Simple Outside — Powerful Inside

الخلفية يمكن أن تكون معقدة جدًا.

لكن المستخدم لا يجب أن يشعر بذلك.

مثلاً، عندما يقوم بمسح قطعة:

```text
Barcode
   ↓
Search
   ↓
Product
   ↓
Stock
   ↓
Serial
   ↓
Cost
   ↓
Price
   ↓
Profit
   ↓
Warranty
   ↓
History
```

كل ذلك يمكن أن يحدث في الخلفية.

أما المستخدم فيرى:

```text
RTX 4070

Available
₪2,350

Cost       ₪1,850
Profit     ₪500

[ بيع ]
```

---

# 4. التقنية المقترحة

## Frontend

```text
React
TypeScript
```

مع بنية Components منظمة وقابلة لإعادة الاستخدام.

مثال:

```text
frontend/
├── app/
├── components/
├── features/
├── layouts/
├── pages/
├── hooks/
├── services/
├── stores/
├── types/
├── utils/
├── i18n/
├── styles/
└── assets/
```

---

# 5. Feature-Based Architecture

لا أنصح بجعل المشروع مجرد مجموعة ضخمة من Components.

الأفضل تنظيمه حسب Features.

مثلاً:

```text
features/
├── auth/
├── dashboard/
├── products/
├── inventory/
├── sales/
├── customers/
├── debts/
├── purchases/
├── suppliers/
├── expenses/
├── reports/
├── barcode/
├── warranties/
└── settings/
```

كل Feature تحتوي ما تحتاجه من:

```text
components
hooks
services
types
validation
```

وهذا يجعل المشروع قابلًا للتوسع.

---

# 6. Design System

قبل بناء عشرات الصفحات، يجب إنشاء Design System.

يشمل:

```text
Colors
Typography
Spacing
Buttons
Inputs
Selects
Tables
Cards
Badges
Dialogs
Drawers
Tabs
Dropdowns
Tooltips
Toasts
Charts
Forms
Loading States
Empty States
Error States
```

الهدف:

> لا تقوم بتصميم كل صفحة من الصفر.

---

# 7. Design Tokens

يجب استخدام Design Tokens.

مثلاً:

```text
--color-primary
--color-background
--color-surface
--color-text
--color-muted
--color-success
--color-warning
--color-danger
```

والأبعاد:

```text
--space-xs
--space-sm
--space-md
--space-lg
--space-xl
```

والـRadius:

```text
--radius-sm
--radius-md
--radius-lg
```

وهذا يسمح بتغيير هوية النظام مستقبلًا بدون إعادة تصميم الواجهة كاملة.

---

# 8. الألوان

الواجهة لا تحتاج عشرات الألوان.

النظام المقترح:

```text
Primary
Secondary
Success
Warning
Danger
Neutral
```

استخدام اللون يجب أن يكون وظيفيًا.

مثلاً:

```text
Green  → جيد
Amber  → يحتاج انتباه
Red    → يحتاج إجراء
Gray   → معلومات ثانوية
```

ولا يجب الاعتماد على اللون وحده.

مثلاً:

```text
🔴 منخفض المخزون
```

أفضل من اللون الأحمر وحده.

---

# 9. Typography

الخط يجب أن يكون:

* واضحًا.
* حديثًا.
* احترافيًا.
* سهل القراءة.
* مناسبًا للأرقام.
* يدعم العربية.
* يدعم العبرية.
* يدعم الإنجليزية.

ويجب الانتباه خصوصًا إلى:

* أسعار المنتجات.
* أرقام Barcode.
* Serial Numbers.
* الكميات.
* التواريخ.

---

# 10. RTL / LTR

النظام يستهدف سوقًا يمكن أن يحتوي على:

```text
Arabic
Hebrew
English
```

لذلك يجب تصميم الـFrontend من البداية لدعم:

```text
RTL
LTR
```

وليس إضافة RTL في نهاية المشروع.

يجب أن تتغير:

* Sidebar.
* Tables.
* Forms.
* Icons.
* Breadcrumbs.
* Dropdowns.
* Drawers.
* Navigation.
* Alignment.

حسب اتجاه اللغة.

---

# 11. Responsive Design

الواجهة ليست Desktop يتم تصغيرها على الهاتف.

يجب تصميم:

## Desktop Experience

للمحل:

* Keyboard.
* Mouse.
* Barcode Scanner.
* Tables.
* Multi-column layouts.
* Bulk actions.

## Mobile Experience

لصاحب المحل والموظف:

* Touch.
* Camera Barcode Scanner.
* Quick actions.
* Notifications.
* Fast lookup.

---

# 12. Breakpoints

يجب اختبار النظام على:

```text
320px
375px
390px
430px
768px
1024px
1280px
1440px
1920px
```

---

# 13. Dashboard

لا نريد Dashboard تقليديًا مليئًا بالـCards.

بل:

# Today's Control Center

مثال:

```text
صباح الخير، أحمد

اليوم

₪7,450
المبيعات

₪1,820
الربح

₪24,800
الديون
```

ثم:

```text
يحتاج انتباهك

🔴 3 ديون متأخرة
🟠 4 قطع منخفضة المخزون
🟠 2 قطع مستعملة تحتاج فحص
```

ثم:

```text
عمليات سريعة

[ بيع ]
[ Scan ]
[ إضافة قطعة ]
[ إضافة عميل ]
[ تسجيل دفعة ]
```

---

# 14. أهم فكرة في Dashboard

لا تجعل النظام ينتظر أن يسأل المستخدم:

> "ما الذي يحدث؟"

بل النظام يخبره:

> **"هذا ما يحتاج انتباهك الآن."**

هذه من أهم نقاط التميز.

---

# 15. Navigation

Desktop:

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

Sidebar يجب أن يكون:

* واضحًا.
* غير عريض.
* قابلًا للطي.
* يحتوي Icons + Labels.
* يحافظ على الحالة.

---

# 16. Mobile Navigation

على الهاتف:

```text
Home
Sales
Scan
More
```

ويجب أن يكون:

# Scan

بارزًا جدًا.

---

# 17. Barcode Experience

Barcode يجب أن يكون أحد أهم عناصر تجربة الاستخدام.

عند الضغط على:

```text
Scan
```

يمكن فتح:

```text
Camera Scanner
```

أو استخدام:

```text
External Barcode Scanner
```

في Windows.

بعد القراءة:

```text
RTX 4070

USED
GRADE A

₪2,350

Stock: 1

[ بيع ]
```

---

# 18. Product Page

صفحة المنتج يجب ألا تكون Form ضخمة.

مثال:

```text
RTX 4070
ASUS

USED • GRADE A

₪2,350
Selling Price

₪1,850
Cost

₪500
Expected Profit

Stock: 1
Barcode: XXXXX
Serial: XXXXX
Location: B-03
Warranty: 30 days
```

ثم:

```text
[ بيع ]
[ تعديل ]
```

---

# 19. المنتجات المستعملة

هذه فرصة UX مهمة.

عند اختيار:

```text
Used
```

تظهر عناصر إضافية:

```text
Condition
A / B / C

Inspection
Start Inspection

Photos
Add Photos

Notes
...

Warranty
...
```

---

# 20. Inspection UX

بدل Form طويل:

```text
GPU
✓ Display
✓ Ports
✓ Temperature
✓ Stress Test
✓ Fans

Result:
PASSED
```

ويمكن حفظ النتيجة في تاريخ القطعة.

---

# 21. Product Timeline

كل قطعة يمكن أن تحتوي:

```text
Purchased
↓
Inspected
↓
Price Changed
↓
Reserved
↓
Sold
```

هذا مهم جدًا خصوصًا للقطع المستعملة والفردية.

---

# 22. صفحة المبيعات

يجب أن تكون سريعة.

```text
بيع جديد

[ Search / Scan ]

Cart
----------------
RTX 4070       ₪2350
RAM 16GB        ₪250

Total          ₪2600

Customer:
Cash Customer

Payment:
Cash / Card / Transfer / Debt

[ إتمام البيع ]
```

---

# 23. قاعدة One-Screen Sale

البيع الطبيعي يجب أن يتم من شاشة واحدة قدر الإمكان.

لا:

```text
Product
↓
Customer
↓
Payment
↓
Confirmation
```

إلا عندما تحتاج العملية إلى تفاصيل إضافية.

---

# 24. Customers

صفحة العميل:

```text
Ahmed

Outstanding
₪1,250

Total Purchases
₪8,450

Last Purchase
2 days ago
```

ثم Timeline:

```text
Sale
Payment
Sale
Return
Payment
```

---

# 25. Debts

صفحة الديون يجب أن تكون عملية.

```text
Ahmed

Outstanding
₪1,250

Last Payment
₪500

Remaining
₪750

[ تسجيل دفعة ]
```

---

# 26. Payment Flow

```text
المبلغ المستحق
₪1,250

المبلغ المدفوع
[ ₪500 ]

طريقة الدفع
Cash
Card
Transfer

[ تسجيل الدفعة ]
```

بعدها:

```text
✓ تم تسجيل الدفعة

الرصيد المتبقي:
₪750
```

---

# 27. Inventory

الواجهة يجب أن تدعم:

```text
All
Low Stock
Out of Stock
Slow Moving
Used
New
Reserved
```

مع:

```text
Search
Filter
Sort
Pagination
Bulk Actions
```

---

# 28. Inventory Table

مثال:

```text
Product      Status    Stock   Cost   Price
------------------------------------------------
RTX 3060     Used      2       900    1250
RTX 4070     New       1       1850   2350
RAM 16GB     New       8       180    250
```

الجداول يجب أن تكون:

* Compact.
* واضحة.
* قابلة للفرز.
* قابلة للتصفية.
* Responsive.

---

# 29. Quick Preview

عند الضغط على المنتج داخل الجدول:

يفتح Drawer بدل مغادرة الصفحة.

يعرض:

```text
Product
Stock
Price
Status
History
```

ثم:

```text
[ Open Full ]
```

---

# 30. Search

يجب أن يكون البحث سريعًا جدًا.

يدعم:

```text
Product
Barcode
Serial
SKU
Customer
```

ويجب توفير بحث عام.

---

# 31. Command Palette

اختصار:

```text
Ctrl + K
```

يفتح:

```text
Search or command...
```

مثلاً:

```text
> RTX 4070
> Ahmed
> Low Stock
> Today's Sales
> New Sale
```

هذه ميزة تجعل النظام يشعر بأنه حديث جدًا.

---

# 32. Smart Alerts

الإشعارات يجب ألا تكون مجرد رقم.

مثلاً:

```text
🔴 RTX 3060 وصل إلى الحد الأدنى.

[ عرض القطعة ]
[ طلب شراء ]
```

أو:

```text
🟠 Ahmed لديه دين متأخر.

[ فتح العميل ]
```

---

# 33. Reports

التقارير يجب أن تبدأ بالأسئلة وليس الجداول.

مثلاً:

```text
كيف كان هذا الشهر؟

Sales
₪74,500

Profit
₪18,240

Expenses
₪7,300

Net Profit
₪10,940
```

ثم:

```text
أكثر المنتجات مبيعًا
أكثر المنتجات ربحًا
المخزون الراكد
الديون
المشتريات
المصروفات
```

---

# 34. Business Intelligence

ميزة مستقبلية قوية:

```text
Money Locked in Inventory
```

مثال:

> لديك ₪32,400 في مخزون لم يتحرك منذ أكثر من 90 يومًا.

هذا يعطي صاحب المحل معلومة قابلة لاتخاذ قرار.

---

# 35. Smart Pricing

يمكن للواجهة عرض:

```text
Current Price
₪1,150

Suggested Price
₪1,250

Reason:
هامش الربح الحالي أقل من الحد المستهدف.
```

ولا يتم تغيير السعر إلا بقرار المستخدم.

---

# 36. Onboarding

المستخدم الجديد لا يجب أن يرى 30 إعدادًا.

البداية:

```text
مرحبًا 👋

لنجهز محلك خلال دقائق.

اسم المحل
[______]

نوع النشاط
[ Computer Store ]

العملة
[ ₪ ILS ]

[ ابدأ ]
```

ثم النظام يجهز Defaults.

---

# 37. Import

يجب دعم:

```text
Excel
CSV
```

Workflow:

```text
Upload
↓
Detect Columns
↓
Map
↓
Preview
↓
Validate
↓
Import
```

---

# 38. Demo Mode

مهم جدًا إذا كان الهدف تسويق المنتج.

يوجد:

```text
Demo Store
```

مع بيانات جاهزة.

يمكن عرض:

```text
Scan
Sale
Customer
Debt
Payment
Inventory
Reports
```

بدون إدخال بيانات حقيقية.

---

# 39. Empty States

لا تستخدم:

```text
No records found.
```

استخدم:

```text
لا توجد مبيعات اليوم.

ابدأ أول عملية بيع من هنا.

[ + بيع جديد ]
```

---

# 40. Loading States

لا تجعل الصفحة فارغة.

استخدم:

```text
Skeleton Loading
```

خصوصًا:

* Dashboard.
* Tables.
* Product Details.
* Reports.

---

# 41. Error States

رسائل الخطأ يجب أن تكون بلغة المستخدم.

بدل:

```text
HTTP 409
```

استخدم:

> هذه القطعة لم تعد متوفرة.

ثم:

```text
آخر عملية:
تم بيعها اليوم 14:32
```

---

# 42. Confirmation

العمليات الحساسة يجب أن تحتوي تأكيدًا واضحًا.

مثال:

> هل تريد استرجاع هذه القطعة؟

```text
[ إلغاء ] [ استرجاع ]
```

---

# 43. Undo

في العمليات الآمنة:

```text
تم تغيير السعر.

[ تراجع ]
```

---

# 44. Tables على الهاتف

لا تحاول إجبار جدول Desktop كامل على شاشة الهاتف.

يمكن تحويله إلى:

```text
Product Card

RTX 4070
Used • Grade A

Stock: 1
Price: ₪2,350

[ View ]
```

---

# 45. Forms

Forms يجب أن تكون قصيرة.

لا:

```text
20 field
```

في شاشة واحدة.

استخدم:

```text
Basic Information
Pricing
Inventory
Advanced
```

أو Progressive Disclosure.

---

# 46. Cards

لا تستخدم Card لكل شيء.

Card تستخدم عندما تساعد على فصل المعلومات.

أما البيانات الكثيفة فالأفضل:

* Table.
* List.
* Grid.

---

# 47. Visual Density

النظام تجاري وليس Landing Page.

لذلك يجب أن تكون المعلومات كثيفة لكن منظمة.

الهدف:

```text
High Information Density
+
Strong Visual Hierarchy
```

---

# 48. Desktop Productivity

على Windows يجب توفير:

* Keyboard Shortcuts.
* Barcode Scanner.
* Fast Search.
* Tables.
* Bulk Operations.
* Multi-column Layout.
* Quick Actions.

الهدف:
موظف معتاد على النظام يستطيع تنفيذ العمليات المتكررة بسرعة كبيرة.

---

# 49. Mobile Productivity

على الهاتف:

* Scan.
* Search.
* Sell.
* Pay.
* View Stock.
* View Customer.
* Notifications.

لا تحاول نقل كل تفاصيل Desktop إلى Mobile.

---

# 50. Accessibility

يجب دعم:

* Keyboard navigation.
* Focus states.
* Contrast.
* Screen readers قدر الإمكان.
* Touch targets.
* عدم الاعتماد على اللون فقط.

---

# 51. Dark Mode

يمكن دعم:

```text
Light
Dark
System
```

لكن Dark Mode يجب أن يصمم بعناية.

ليس:

```text
invert colors
```

فقط.

---

# 52. Animations

Animation تستخدم فقط عندما تفيد.

أمثلة:

* فتح Drawer.
* انتقال Modal.
* Success.
* Loading.
* تغيير حالة.

المدة المناسبة غالبًا:

```text
150–250ms
```

---

# 53. لا تستخدم Gaming UI

رغم أن النظام مخصص لمحل قطع حاسوب، لا يجب أن يتحول إلى:

```text
Neon
RGB
Cyberpunk
Glowing Cards
```

إلا إذا كان Branding المحل نفسه يحتاج ذلك.

المنتج الأساسي:

# Business First

---

# 54. Design Personality

الواجهة يجب أن تكون:

```text
Modern
Professional
Fast
Calm
Intelligent
Reliable
```

وليس:

```text
Cold
Technical
Complicated
Corporate
```

---

# 55. Frontend State Management

يجب فصل:

### Server State

مثل:

* Products.
* Sales.
* Customers.
* Inventory.
* Reports.

عن:

### UI State

مثل:

* Modal.
* Sidebar.
* Filters.
* Selected Product.
* Theme.
* Language.

هذا يمنع فوضى State Management.

---

# 56. API Layer

لا تجعل Components تتصل مباشرة بكل API.

الأفضل:

```text
Component
↓
Hook
↓
Service
↓
API
```

مثال:

```text
ProductPage
↓
useProduct()
↓
productService
↓
Backend API
```

هذا يسهل الاختبار والصيانة.

---

# 57. Validation

يجب أن تكون Validation على مستوى Frontend لتجربة المستخدم، مع Validation حقيقية في Backend.

مثلاً:

* السعر.
* الكمية.
* Barcode.
* Serial.
* Customer.
* Payment.

---

# 58. Security UX

الواجهة يجب أن تحترم صلاحيات المستخدم.

مثلاً الموظف لا يرى:

```text
Profit Margin
Cost Price
Financial Reports
```

إذا لم تكن لديه الصلاحية.

ولا يكفي إخفاء الزر فقط؛ Backend يجب أن يفرض الصلاحيات أيضًا.

---

# 59. Role-Based UI

أمثلة:

### Owner

كل شيء.

### Manager

المخزون والمبيعات والتقارير حسب الصلاحيات.

### Employee

البيع والمخزون والعمليات المسموحة.

### Accountant

التقارير والبيانات المالية المناسبة.

---

# 60. Audit UX

يجب توفير سجل واضح:

```text
Ahmed changed price

Old:
₪1,150

New:
₪1,250

18 Aug 2026
14:32
```

وهذا مهم للثقة والإدارة.

---

# 61. Performance

الـFrontend يجب أن يكون سريعًا حتى مع:

```text
Thousands of products
Thousands of sales
Large customers list
Large inventory
```

يجب استخدام:

* Pagination.
* Lazy Loading.
* Code Splitting.
* Virtualized Lists عند الحاجة.
* Image Optimization.
* Caching.
* Efficient API calls.

---

# 62. Barcode Performance

عملية:

```text
Scan
```

يجب ألا تسبب:

```text
Loading screen كامل
```

كل مرة.

الأفضل:

```text
Scan
↓
Quick lookup
↓
Immediate result
```

---

# 63. Offline / Network Resilience

بما أن المحل يعتمد على النظام في العمليات اليومية، يجب التفكير في حالات:

```text
Slow Internet
Temporary Disconnect
Backend Unavailable
```

الواجهة يجب أن تعرض حالة الاتصال بوضوح.

مثلاً:

```text
● Connected
```

أو:

```text
● Reconnecting...
```

لكن لا تخدع المستخدم بأن العملية تمت إذا لم يتم تأكيدها من الخادم.

---

# 64. Toasts

يجب ألا تكون مزعجة.

مثال:

```text
✓ تم حفظ القطعة
```

ثم تختفي.

لكن العمليات المالية المهمة تحتاج Confirmation أو Receipt واضح.

---

# 65. Notifications Center

يمكن توفير مركز إشعارات:

```text
Today

3 Actions Required
4 Reviews
2 Updates
```

لكن الإشعارات يجب أن تكون قابلة للفتح مباشرة.

---

# 66. Frontend Testing

يجب اختبار:

### Unit Tests

* Utilities.
* Validation.
* Calculations.

### Component Tests

* Forms.
* Buttons.
* Scanner states.

### Integration Tests

* Sale flow.
* Payment flow.
* Product creation.

### E2E

مثلاً:

```text
Login
↓
Scan
↓
Add to Cart
↓
Sell
↓
Payment
↓
Dashboard Update
```

---

# 67. Mobile Testing

اختبار حقيقي على أجهزة مختلفة، وليس فقط Browser Resize.

خصوصًا:

* Camera.
* Touch.
* Barcode.
* Keyboard.
* Orientation.
* Network changes.

---

# 68. Browser/Desktop Testing

يجب اختبار:

* Chrome.
* Edge.
* Windows WebView/Packaging حسب الحل النهائي.

مع:

* Barcode Scanner.
* Keyboard.
* Printer إن تم دعمها.
* Large screens.

---

# 69. Frontend Folder Structure

هيكل مقترح:

```text
src/
├── app/
│   ├── router/
│   ├── providers/
│   └── config/
│
├── components/
│   ├── ui/
│   ├── forms/
│   ├── tables/
│   ├── feedback/
│   └── navigation/
│
├── features/
│   ├── auth/
│   ├── dashboard/
│   ├── products/
│   ├── inventory/
│   ├── sales/
│   ├── customers/
│   ├── debts/
│   ├── purchases/
│   ├── suppliers/
│   ├── expenses/
│   ├── reports/
│   ├── barcode/
│   └── settings/
│
├── layouts/
│   ├── DesktopLayout/
│   └── MobileLayout/
│
├── hooks/
├── services/
├── stores/
├── types/
├── utils/
├── i18n/
├── assets/
└── styles/
```

---

# 70. Component Hierarchy

مثال:

```text
App
└── AuthProvider
    └── Router
        └── AppLayout
            ├── Sidebar
            ├── TopBar
            └── MainContent
                └── Dashboard
```

---

# 71. Feature Example

```text
features/products/

components/
ProductCard
ProductTable
ProductForm
ProductDrawer

hooks/
useProducts
useProduct
useCreateProduct

services/
productService

types/
product.types.ts

validation/
product.schema.ts
```

هذا يجعل كل Feature مستقلة وقابلة للصيانة.

---

# 72. UX Rule

كل عملية متكررة يجب أن تصبح:

# Faster Over Time

مثلاً الموظف في أول يوم:

```text
Scan
→ Product
→ Sell
```

بعد أسبوع:

```text
Scan
→ Enter
```

بعد شهر:

```text
Scanner
→ Keyboard
→ Enter
```

النظام يصبح جزءًا طبيعيًا من العمل.

---

# 73. Marketing Differentiation

لا تسوق:

> لدينا 150 ميزة.

بل:

> **النظام يعرف ما تحتاجه قبل أن تبحث عنه.**

مثلاً:

> لديك 3 ديون متأخرة.

> لديك 4 قطع أوشكت على النفاد.

> لديك ₪32,400 في مخزون بطيء الحركة.

هذه رسائل أقوى من قائمة Features.

---

# 74. القيمة الحقيقية للـFrontend

الـFrontend ليس مجرد:

```text
Buttons
Tables
Forms
```

بل هو طبقة تجعل النظام المعقد:

```text
Easy
Fast
Understandable
Predictable
```

---

# 75. قاعدة "3 Seconds"

عند فتح أي شاشة، يجب أن يعرف المستخدم خلال حوالي 3 ثوانٍ:

1. أين أنا؟
2. ماذا أرى؟
3. ماذا يمكنني أن أفعل؟

إذا لم يعرف ذلك، فالتصميم يحتاج تحسينًا.

---

# 76. قاعدة "One Primary Action"

كل شاشة:

```text
One Main Goal
One Primary CTA
```

مثال:

Inventory:

```text
[ + إضافة قطعة ]
```

Customer:

```text
[ تسجيل دفعة ]
```

Product:

```text
[ بيع ]
```

---

# 77. قاعدة "No Dead Ends"

كل شاشة يجب أن تقود إلى إجراء منطقي.

مثلاً:

```text
Low Stock
↓
Product
↓
Purchase
```

أو:

```text
Overdue Debt
↓
Customer
↓
Record Payment
```

---

# 78. Smart Workflow

الهدف النهائي:

```text
Event
↓
System understands
↓
System updates related data
↓
System generates notification if needed
↓
Dashboard updates
```

مثلاً بيع قطعة:

```text
Sale
↓
Stock -1
↓
Revenue updated
↓
Profit updated
↓
Customer balance updated if debt
↓
Product history updated
↓
Low-stock alert if needed
↓
Dashboard updated
```

المستخدم لا يقوم بهذه الخطوات يدويًا.

---

# 79. هذا هو الفرق الأساسي

النظام التقليدي:

```text
User
↓
Input everything
↓
System stores it
```

النظام المقترح:

```text
User
↓
One simple action
↓
System handles the consequences
```

وهذا هو جوهر المنتج.

---

# 80. التصميم النهائي المقترح

المنتج يجب أن يجمع:

```text
Modern UI
+
Fast UX
+
Barcode First
+
Mobile
+
Desktop
+
RTL/LTR
+
Smart Alerts
+
Automation
+
Business Intelligence
+
Simple Workflows
```

والنتيجة:

# Smart Store Operating System

وليس مجرد:

# POS.

---

# 81. الرؤية التجارية النهائية

إذا تم تنفيذ الـFrontend بهذه الطريقة، فالمنتج لن يكون مجرد مشروع تخرج أو CRUD Dashboard.

سيصبح أساس منتج SaaS يمكن تطويره لاحقًا لدعم:

* عدة محلات.
* عدة فروع.
* عدة مستخدمين.
* صلاحيات.
* اشتراكات.
* تقارير متقدمة.
* Automation.
* AI.
* Integrations.
* أجهزة Barcode.
* طابعات.
* Receipts.
* خدمات خارجية.

---

# 82. المعيار النهائي

قبل إطلاق أي شاشة اسأل:

> هل تجعل هذه الشاشة حياة صاحب المحل أسهل؟

إذا كانت الإجابة:

**نعم → احتفظ بها.**

إذا كانت:

**لا، لكنها تبدو جميلة → أعد تصميمها.**

إذا كانت:

**لا يحتاجها المستخدم أصلًا → احذفها.**

---

# 83. الخلاصة

الـFrontend الناجح لهذا المشروع يجب أن يكون:

**بسيطًا للمستخدم، قويًا في الخلفية، سريعًا في العمليات، واضحًا في الأموال والمخزون، ومميزًا بصريًا دون مبالغة.**

والفكرة التي يجب أن تحكم كل قرار تصميمي هي:

> ## **"النظام يعمل من أجل صاحب المحل، وليس صاحب المحل من أجل النظام."**

ومن هذه الفكرة تبني كل شيء:

```text
Scan
↓
Understand
↓
Automate
↓
Notify
↓
Recommend
↓
User Decides
```# تقرير التصميم والمظهر — Frontend Design System
## Smart Computer Store Management System

> وثيقة مرجعية لهوية النظام البصرية، نظام التصميم، وسلوك الواجهة عبر المنصات (Desktop / Tablet / Mobile).

---

# 1. الفلسفة البصرية

الواجهة ليست "لوحة تحكم تقنية"، بل **أداة عمل يومية** يفتحها صاحب المحل أو موظفه عشرات المرات يوميًا تحت ضغط الوقت (زبون واقف أمامه).

القاعدة الحاكمة:

> **كل بكسل يجب أن يخدم السرعة أو الوضوح، وإلا فهو زائد.**

هذا يعني عمليًا:

| بدل | استخدم |
|---|---|
| شاشات مزدحمة بكل البيانات الممكنة | شاشة تُظهر ما يحتاجه المستخدم *الآن* فقط |
| ألوان كثيرة للتزيين | لون واحد أساسي + ألوان دلالية محدودة (نجاح/خطر/تنبيه) |
| أيقونات وأزرار ضخمة "تجميلية" | كثافة معلومات عالية في الأماكن التي يحتاجها الموظف الخبير |
| تنقل عبر عدة صفحات لإتمام مهمة | إجراء واحد يكفي (Scan → Sell) |

**الإحساس المستهدف**: نظيف، سريع، احترافي — أقرب إلى Linear أو Stripe Dashboard منه إلى برنامج محاسبة كلاسيكي أو ERP مثقل.

---

# 2. أهداف التصميم (Design Goals)

1. **السرعة الإدراكية**: يفهم المستخدم الشاشة خلال أقل من ثانيتين.
2. **أقل عدد نقرات**: كل مهمة متكررة (بيع، بحث، تسجيل دفعة) بأقل خطوات ممكنة.
3. **وضوح الحالة المالية**: الأرقام المهمة (ربح، دين، مخزون) بارزة دومًا، لا تحتاج بحث.
4. **ثقة بصرية**: تصميم يبدو مدفوعًا (Premium) وليس مشروعًا تجريبيًا — لأن هذا يبني ثقة صاحب المحل بالنظام كأداة مالية.
5. **الاتساق**: نفس المكوّن يبدو ويتصرف بنفس الطريقة في كل الشاشات.
6. **التكيّف الكامل**: نفس النظام التصميمي يعمل بصريًا وعمليًا على شاشة 27 بوصة وعلى موبايل بشاشة 6 بوصات.

---

# 3. الهوية البصرية (Brand Identity)

## 3.1 الطابع العام
- **Modern Tech-Utility**: مستوحى من أدوات SaaS الاحترافية (Linear, Notion, Stripe) لا من قوالب لوحات التحكم الجاهزة (Bootstrap Admin).
- ميل نحو **الحياد اللوني** (رمادي/أبيض/أسود) مع لون هوية واحد فاتح الحضور، بحيث تبرز الألوان الدلالية (أخضر/أحمر/برتقالي) بوضوح عند الحاجة فقط.

## 3.2 الشخصية
- **Precise** — دقيق، لا يترك مجالًا للغموض في الأرقام.
- **Calm** — لا صراخ بصري، حتى التنبيهات الحرجة تكون واضحة دون فوضى.
- **Fast** — كل شيء يوحي بالخفة والاستجابة الفورية.

---

# 4. نظام الألوان (Color System)

## 4.1 المنطق
لون أساسي واحد (Primary) للعلامة والإجراءات، رمادي محايد للخلفيات والنصوص، وألوان دلالية ثابتة المعنى في كل النظام (لا يجوز أن يعني الأخضر "خطر" في شاشة و"نجاح" في أخرى).

## 4.2 لوحة الألوان المقترحة

### الألوان الأساسية (Neutral Scale)
```
--gray-50:  #FAFAFA
--gray-100: #F4F4F5
--gray-200: #E4E4E7
--gray-300: #D4D4D8
--gray-400: #A1A1AA
--gray-500: #71717A
--gray-600: #52525B
--gray-700: #3F3F46
--gray-800: #27272A
--gray-900: #18181B
```

### اللون الأساسي (Primary — الهوية)
```
--primary-50:  #EFF6FF
--primary-100: #DBEAFE
--primary-400: #60A5FA
--primary-500: #3B82F6   ← اللون الأساسي (أزرار، روابط، تحديد)
--primary-600: #2563EB
--primary-700: #1D4ED8
```
> يمكن استبدال الأزرق بلون هوية آخر (بنفسجي/أخضر داكن) حسب اسم العلامة النهائي — الأهم أن يبقى واحدًا ومتسقًا.

### الألوان الدلالية (Semantic Colors)
| الاستخدام | اللون | الكود |
|---|---|---|
| نجاح / مكتمل / مدفوع | أخضر | `#16A34A` |
| تحذير / مخزون منخفض / ضمان قارب على الانتهاء | برتقالي/كهرماني | `#D97706` |
| خطر / دين متأخر / خطأ | أحمر | `#DC2626` |
| معلومة / إشعار عام | أزرق فاتح | `#0284C7` |
| قطعة مستعملة (Badge خاص) | بنفسجي رمادي | `#7C3AED` |

### الوضع الداكن (Dark Mode)
النظام يُبنى مع دعم Dark Mode من البداية عبر CSS Variables وليس كطبقة لاحقة، لأن كثيرًا من الموظفين يعملون في إضاءة منخفضة مساءً:
```
--bg-primary-dark:   #0F0F11
--bg-surface-dark:   #18181B
--border-dark:       #27272A
--text-primary-dark: #FAFAFA
--text-muted-dark:   #A1A1AA
```

## 4.3 قاعدة استخدام اللون
- **لا يُستخدم أكثر من لونين دلاليين في نفس الشاشة الفرعية** إلا للضرورة (مثلاً لا تُلوّن 6 أشياء مختلفة بـ6 ألوان في نفس البطاقة).
- الألوان الدلالية تُستخدم للحالة (Status) فقط، لا للتزيين.

---

# 5. الطباعة (Typography)

## 5.1 الخط
- **واجهة عربية/عبرية/إنجليزية موحّدة**: خط يدعم الثلاث لغات بجودة عالية (اقتراح: `IBM Plex Sans Arabic` أو `Noto Sans Arabic` للعربية مع `Inter` أو `IBM Plex Sans` للإنجليزية/الأرقام اللاتينية — Font Stack مزدوج).
- الأرقام دائمًا **Tabular / Latin numerals** بخط ثابت العرض في الجداول المالية، لتفادي "قفز" الأرقام عند التحديث المباشر.

## 5.2 السلم الطباعي (Type Scale)
| المستوى | الحجم | الوزن | الاستخدام |
|---|---|---|---|
| Display | 32px | Bold | أرقام Dashboard الكبرى (المبيعات اليوم) |
| H1 | 24px | Semibold | عنوان الصفحة |
| H2 | 20px | Semibold | عنوان قسم/بطاقة |
| H3 | 16px | Medium | عناوين فرعية |
| Body | 14px | Regular | النص العام، الجداول |
| Small | 12px | Regular | تسميات، Timestamps، Captions |
| Micro | 11px | Medium | Badges، حالات صغيرة |

## 5.3 قاعدة عملية
- الأرقام المالية المهمة (السعر، الربح، الدين) تُكتب دائمًا بوزن أثقل من النص المحيط بها، حتى لو كانت بنفس الحجم — العين يجب أن تجدها فورًا.

---

# 6. نظام الشبكة والتباعد (Grid & Spacing)

## 6.1 وحدة القياس
نظام تباعد قائم على مضاعفات **4px**:
```
4 · 8 · 12 · 16 · 24 · 32 · 48 · 64
```
هذا يمنع القرارات العشوائية في التباعد ويجعل كل شيء "يبدو مصممًا" لا عشوائيًا.

## 6.2 الشبكة
- **Desktop**: حاوية بعرض أقصى مرن (Fluid) مع أعمدة (12-column grid)، هوامش جانبية 24-32px.
- **Tablet**: 8 أعمدة، هوامش 16-20px.
- **Mobile**: عمود واحد، هوامش 16px، وحدات كاملة العرض.

## 6.3 كثافة المعلومات حسب المنصة
- **Desktop**: كثافة عالية — جداول بأعمدة كثيرة، عرض متعدد الأقسام في نفس الشاشة (مثال: تفاصيل الزبون + قائمة مشترياته جنبًا إلى جنب).
- **Mobile**: كثافة منخفضة — عنصر واحد يتصدر الشاشة، تمرير رأسي بسيط، لا جداول أفقية معقدة.

---

# 7. مكتبة المكوّنات (Component Library)

## 7.1 الأزرار (Buttons)
| النوع | الاستخدام |
|---|---|
| Primary | إجراء رئيسي واحد فقط لكل شاشة (بيع، حفظ، تأكيد) |
| Secondary | إجراءات ثانوية (إلغاء، رجوع) |
| Destructive | حذف/إرجاع — أحمر، مع تأكيد إضافي دائمًا |
| Icon Button | إجراءات سريعة متكررة (بحث، مسح Barcode) |

- ارتفاع الزر على الموبايل ≥ 44px (Touch Target قياسي).
- الزر الأساسي في شاشة البيع يكون **الأكبر والأوضح عنصر في الشاشة** — لأنه العملية الأكثر تكرارًا في اليوم.

## 7.2 البطاقات (Cards)
تُستخدم لعرض: مؤشرات Dashboard، بطاقة قطعة، بطاقة زبون.
- حواف دائرية خفيفة (8-12px)، ظل خفيف جدًا (لا Skeuomorphism)، حدود رفيعة (1px) بدل ظلال ثقيلة في الوضع الفاتح.

## 7.3 الجداول (Data Tables)
- Desktop: أعمدة قابلة للفرز، فلاتر أعلى الجدول، تحديد متعدد للعمليات الجماعية.
- Mobile: الجدول يتحول تلقائيًا إلى **قائمة بطاقات مصغّرة** (كل صف → بطاقة) بدل تمرير أفقي — لأن الجداول الأفقية على الموبايل تجربة سيئة دائمًا.

## 7.4 الشارات (Status Badges)
كل حالة (متوفر/نافد/مستعمل/جديد/متأخر/مدفوع) لها لون وشكل ثابت في كل الشاشات:
```
[ متوفر ]   أخضر فاتح خلفية + نص أخضر داكن
[ نافد ]    رمادي
[ مستعمل ]  بنفسجي فاتح
[ متأخر ]   أحمر فاتح
```

## 7.5 النماذج (Forms)
- حقل واحد لكل سطر على الموبايل، حقلان جنبًا إلى جنب مسموح على Desktop فقط عند الترابط المنطقي (مثال: السعر + العملة).
- التحقق من الإدخال (Validation) فوري أثناء الكتابة، لا بعد الإرسال فقط.
- رسائل الخطأ توضع أسفل الحقل مباشرة بلون أحمر، بدون Popup منفصل.

## 7.6 الحوارات والنوافذ المنبثقة (Modals / Sheets)
- Desktop: Modal مركزي للعمليات القصيرة (تأكيد، تعديل سريع).
- Mobile: **Bottom Sheet** بدل Modal مركزي — أسهل بالإبهام، وأقرب لتوقعات تطبيقات الموبايل.

## 7.7 الإشعارات (Toasts / Notifications)
- Toast صغير أعلى/أسفل الشاشة لتأكيد العمليات ("تم البيع بنجاح") يختفي تلقائيًا خلال 3 ثوانٍ.
- مركز إشعارات دائم (🔔) للتنبيهات التي تحتاج متابعة لاحقة (دين متأخر، مخزون منخفض).

---

# 8. الأيقونات (Iconography)

- مجموعة أيقونات واحدة متسقة الخط (Outline style موحّد، وزن خط ثابت) — يفضّل مكتبة جاهزة مثل **Lucide** لتفادي عدم الاتساق.
- الأيقونات دائمًا مرفقة بنص عند أول ظهور لها في القسم (لا يُعتمد على الأيقونة وحدها لتوصيل المعنى، خصوصًا لموظفين جدد).

---

# 9. دعم تعدد اللغات (i18n) و RTL/LTR

هذه نقطة حرجة تقنيًا وتصميميًا لأن النظام يدعم العربية والعبرية (RTL) والإنجليزية (LTR) معًا.

## 9.1 القواعد
- **كل مكوّن يُبنى Direction-Agnostic**: استخدام `margin-inline-start/end` بدل `margin-left/right`، و`flex-direction` يُقرأ من اتجاه الصفحة تلقائيًا.
- الأيقونات ذات الاتجاه (أسهم، رجوع/تقدم) تُعكس تلقائيًا في RTL.
- الأرقام والعملة تبقى بصيغتها القياسية (LTR داخل جملة RTL) لتفادي التباس القراءة، وهذا سلوك قياسي معتمد في تطبيقات مالية عربية.
- اختبار كل شاشة جديدة في الاتجاهين قبل اعتمادها ضمن Definition of Done.

## 9.2 تبديل اللغة
- مبدّل لغة واضح في الإعدادات وشاشة الدخول، يُطبَّق فوريًا دون إعادة تحميل كاملة للتطبيق قدر الإمكان.

---

# 10. التصميم المتجاوب (Responsive Breakpoints)

```
Mobile:   < 640px   → عمود واحد، تنقل سفلي (Bottom Nav)
Tablet:   640–1024  → عمودان، تنقل جانبي قابل للطي
Desktop:  > 1024px  → تخطيط كامل، Sidebar ثابت، كثافة بيانات عالية
```

## 10.1 نمط التنقل حسب المنصة
| المنصة | نمط التنقل |
|---|---|
| Desktop | Sidebar ثابت على اليمين (لواجهة RTL) بأيقونات + نص |
| Tablet | Sidebar قابل للطي (Icons فقط عند الطي) |
| Mobile | Bottom Navigation بـ4-5 عناصر أساسية فقط (لوحة، بيع، بحث، إشعارات، المزيد) |

## 10.2 مبدأ إعادة التدفق لا إعادة التصغير
لا يُعاد استخدام نفس تخطيط Desktop مصغّرًا على الموبايل — بل يُعاد تصميم تدفق الشاشة (Reflow) حول الإجراء الأهم لتلك الشاشة على الموبايل تحديدًا.

---

# 11. تجربة الموبايل (Mobile-First Workflows)

الموبايل ليس نسخة "طوارئ" من Desktop، بل مصمم حول 3 سيناريوهات مكثفة الاستخدام:

```
Scan → Sell         (البيع السريع)
Search → Know        (معرفة حالة قطعة فورًا)
Customer → Pay        (تسجيل دفعة دين في الميدان)
```

- دعم كاميرا الهاتف لمسح الـBarcode مباشرة بدون جهاز خارجي.
- أزرار وحقول كبيرة كافية للاستخدام بإبهام واحد (One-hand usage).
- أهم إجراء في كل شاشة يكون في **منطقة الإبهام السفلية** لا أعلى الشاشة.

---

# 12. تجربة سطح المكتب (Desktop Workflows)

Desktop يُستغل لما لا يوفره الموبايل بطبيعته:

- جداول كثيفة قابلة للفرز والفلترة.
- اختصارات لوحة المفاتيح (مثال: `/` للبحث السريع، `N` لبيع جديد).
- عرض متعدد الأعمدة (قائمة + تفاصيل جنبًا إلى جنب) بدل التنقل بين صفحات.
- عمليات جماعية (تحديد عدة عناصر وتنفيذ إجراء واحد عليها).

---

# 13. الحالات الأساسية لكل شاشة (UI States)

كل شاشة أو مكوّن بيانات **يجب** أن يُصمَّم بأربع حالات، لا حالة "المحتوى المثالي" فقط:

1. **Loading** — هيكل عظمي (Skeleton) بدل دوّارة تحميل عامة، يعطي إحساسًا بالسرعة.
2. **Empty** — رسالة ودّية + إجراء واضح ("لا توجد منتجات بعد — إضافة أول منتج").
3. **Error** — رسالة مفهومة بلغة بشرية + خيار إعادة المحاولة، لا كود خطأ تقني للمستخدم النهائي.
4. **Populated** — الحالة الطبيعية بالبيانات الكاملة.

---

# 14. الحركة والانتقالات (Motion)

- انتقالات قصيرة جدًا (150-200ms) بين الشاشات — الهدف الإحساس بالاستجابة الفورية لا "العرض الجمالي".
- تأكيد العمليات الحرجة (بيع، حذف) بحركة بصرية بسيطة (Checkmark مختصر) بدل نص فقط، لأن الموظف يعمل بسرعة ولا يقرأ كل رسالة.
- لا Animations زخرفية طويلة تُبطئ سير العمل المتكرر.

---

# 15. إمكانية الوصول (Accessibility)

- تباين ألوان يحقق معيار WCAG AA على الأقل لكل نص أساسي.
- كل عنصر تفاعلي قابل للوصول عبر لوحة المفاتيح مع Focus State واضح.
- تسميات مناسبة لقارئات الشاشة (Screen Reader labels) خاصة في النماذج والأيقونات المستقلة.
- حجم أهداف اللمس (Touch Targets) لا يقل عن 44×44px على الموبايل.

---

# 16. الشاشات المحورية (Key Screens Overview)

## 16.1 Dashboard
أرقام كبيرة وواضحة أولًا (مبيعات اليوم، الربح، الديون المستحقة)، ثم قسم "النظام يقترح" أسفلها بتنبيهات قابلة للنقر مباشرة للإجراء.

## 16.2 شاشة البيع (POS Screen)
أكبر مساحة للـBarcode/البحث، سلة جانبية واضحة، زر الدفع النهائي ثابت ومرئي دومًا دون الحاجة للتمرير لأسفل.

## 16.3 بطاقة القطعة (Item Detail)
صورة/أيقونة، السعر والربح بارزين، Timeline زمني بسيط لحياة القطعة (شراء → فحص → تخزين → بيع).

## 16.4 ملف الزبون
الرصيد المستحق في أعلى البطاقة بخط كبير وواضح دائمًا — هذا الرقم هو أهم معلومة في هذه الشاشة تحديدًا.

## 16.5 التقارير
رسوم بيانية بسيطة (Line/Bar) لا تفاصيل زائدة، مع خيار "عرض التفاصيل" عند الطلب فقط — التقرير الافتراضي يُجيب السؤال مباشرة بدون تحليل من المستخدم.

---

# 17. Design Tokens (مثال تقني للتطبيق)

```css
:root {
  --radius-sm: 6px;
  --radius-md: 10px;
  --radius-lg: 16px;

  --space-1: 4px;
  --space-2: 8px;
  --space-3: 12px;
  --space-4: 16px;
  --space-6: 24px;
  --space-8: 32px;

  --shadow-sm: 0 1px 2px rgba(0,0,0,0.05);
  --shadow-md: 0 4px 12px rgba(0,0,0,0.08);

  --font-body: 'IBM Plex Sans Arabic', 'Inter', sans-serif;
}
```

استخدام Design Tokens بدل قيم ثابتة متفرقة في الكود يضمن أن أي تعديل مستقبلي في الهوية (تغيير لون أساسي مثلاً) يتم من مكان واحد فقط.

---

# 18. معايير جودة تصميم (Design QA Checklist)

قبل اعتماد أي شاشة جديدة كمكتملة:

```
☐ تعمل بصريًا في RTL و LTR
☐ تعمل على Mobile / Tablet / Desktop
☐ الحالات الأربع مصممة (Loading/Empty/Error/Populated)
☐ التباين اللوني يحقق WCAG AA
☐ لا يوجد أكثر من إجراء أساسي واحد بارز في الشاشة
☐ الأرقام المالية بارزة ومقروءة فورًا
☐ متسقة مع Design Tokens المعتمدة (لا قيم عشوائية)
☐ اختبار الاستخدام بإبهام واحد على الموبايل
```

---

# 19. الخلاصة

نظام التصميم المقترح يقوم على مبدأ واحد يحكم كل قرار بصري:

> **بساطة من الخارج، قوة من الداخل.**

الواجهة لا تُقاس بجمالها فقط، بل بقدرتها على جعل موظف جديد يبيع أول قطعة خلال دقائق دون تدريب طويل — وجعل صاحب المحل يفهم وضعه المالي بنظرة واحدة إلى Dashboard. كل مكوّن، لون، وحركة في هذا النظام يجب أن يخدم هذا الهدف مباشرة.

**هذه هي هوية الـFrontend التي أنصح أن تبني عليها المشروع بالكامل.**

