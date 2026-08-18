# دليل تصميم وتطوير Frontend
## PartFlow - Smart Computer Store Management System

> **الهدف:** بناء واجهة Frontend احترافية لنظام إدارة محلات قطع الحاسوب، تجعل النظام سريعًا، بسيطًا، ذكيًا، وقابلًا للتسويق تجاريًا، مع دعم Desktop/Windows وMobile والعربية والعبرية والإنجليزية.

---

## 📑 فهرس المحتويات

1. [الرؤية الأساسية](#1-الرؤية-الأساسية)
2. [أهداف التصميم](#2-أهداف-التصميم)
3. [التقنية المقترحة](#3-التقنية-المقترحة)
4. [Feature-Based Architecture](#4-feature-based-architecture)
5. [Design System](#5-design-system)
6. [نظام الألوان](#6-نظام-الألوان)
7. [الطباعة (Typography)](#7-الطباعة-typography)
8. [RTL/LTR Support](#8-rtlltr-support)
9. [Responsive Design](#9-responsive-design)
10. [الأيقونات](#10-الأيقونات)
11. [مكتبة المكوّنات](#11-مكتبة-المكوّنات)
12. [الشاشات المحورية](#12-الشاشات-المحورية)
13. [Barcode Experience](#13-barcode-experience)
14. [Dashboard Design](#14-dashboard-design)
15. [Navigation Design](#15-navigation-design)
16. [State Management](#16-state-management)
17. [API Layer](#17-api-layer)
18. [Validation](#18-validation)
19. [Security & Roles](#19-security--roles)
20. [Performance](#20-performance)
21. [Accessibility](#21-accessibility)
22. [Dark Mode](#22-dark-mode)
23. [Animations](#23-animations)
24. [Testing Strategy](#24-testing-strategy)
25. [معايير الجودة](#25-معايير-الجودة)

---

## 1. الرؤية الأساسية

المنتج لا يجب أن يبدو كبرنامج محاسبة تقليدي أو نظام ERP معقد. بل يجب أن يعطي المستخدم هذا الشعور:

> **"أنا أدير المحل، والنظام يدير التفاصيل."**

الـFrontend يجب أن يخفي التعقيد الموجود في الخلفية. المستخدم يرى:

```
Scan
↓
System Understands
↓
One Action
↓
Everything Updates
```

بدل أن يضطر إلى التنقل بين عدة صفحات وإدخال نفس البيانات أكثر من مرة.

### المبدأ الأساسي: Simple Outside — Powerful Inside

الخلفية يمكن أن تكون معقدة جدًا، لكن المستخدم لا يجب أن يشعر بذلك.

**مثال:** عند مسح قطعة، يحدث في الخلفية:
```
Barcode → Search → Product → Stock → Serial → Cost → Price → Profit → Warranty → History
```

أما المستخدم فيرى:
```
RTX 4070
Available
₪2,350
Cost       ₪1,850
Profit     ₪500
[ بيع ]
```

---

## 2. أهداف التصميم

الواجهة يجب أن تحقق:

1. **سهولة الاستخدام** - لا يتطلب تدريبًا طويلًا
2. **سرعة العمليات** - أقل عدد نقرات للمهام المتكررة
3. **وضوح المعلومات** - الأرقام المهمة بارزة دومًا
4. **تقليل الأخطاء** - Validation واضح وتوجيهات مباشرة
5. **دعم Barcode** - أساس التجربة في المتجر
6. **دعم القطع الفردية والمستعملة** - إدارة حالة كل قطعة
7. **دعم Windows/Desktop** - إنتاجية عالية للموظفين
8. **دعم الهواتف** - إمكانية الوصول للمالك
9. **دعم Touch & Keyboard** - تكيف مع طريقة الإدخال
10. **دعم RTL/LTR** - العربية والعبرية والإنجليزية
11. **Responsive Design حقيقي** - ليس مجرد تصغير
12. **Accessibility** - قابلية الوصول للجميع
13. **Performance عالي** - استجابة فورية
14. **تصميم قابل للتوسع** - نظام موحد
15. **مظهر Premium** - قابل للتسويق تجاريًا

### الفلسفة البصرية

الواجهة ليست "لوحة تحكم تقنية"، بل **أداة عمل يومية** يفتحها صاحب المحل أو موظفه عشرات المرات يوميًا تحت ضغط الوقت.

**القاعدة الحاكمة:**
> **كل بكسل يجب أن يخدم السرعة أو الوضوح، وإلا فهو زائد.**

**الإحساس المستهدف:** نظيف، سريع، احترافي — أقرب إلى Linear أو Stripe Dashboard منه إلى برنامج محاسبة كلاسيكي.

---

## 3. التقنية المقترحة

### Frontend Stack
```text
React
TypeScript
Vite
```

### البنية المقترحة
```text
frontend/
├── src/
│   ├── app/              # التطبيق الرئيسي والـrouter
│   ├── components/       # مكوّنات مشتركة
│   ├── features/         # Feature-based organization
│   ├── layouts/          # Layouts رئيسية
│   ├── hooks/            # Custom Hooks
│   ├── services/         # API Services
│   ├── stores/           # State Management
│   ├── types/            # TypeScript Types
│   ├── utils/            # Utilities
│   ├── i18n/             # Internationalization
│   ├── styles/           # Global Styles
│   └── assets/           # Static Assets
├── public/
└── package.json
```

---

## 4. Feature-Based Architecture

لا أنصح بجعل المشروع مجرد مجموعة ضخمة من Components. الأفضل تنظيمه حسب Features:

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

كل Feature يحتوي:
```text
components/    # مكوّنات خاصة بالـFeature
hooks/         # Custom Hooks
services/      # API Services
types/         # TypeScript Types
validation/    # Zod Schemas
```

**مثال - Feature Structure:**
```text
features/products/
├── components/
│   ├── ProductCard.tsx
│   ├── ProductTable.tsx
│   ├── ProductForm.tsx
│   └── ProductDrawer.tsx
├── hooks/
│   ├── useProducts.ts
│   ├── useProduct.ts
│   └── useCreateProduct.ts
├── services/
│   └── productService.ts
├── types/
│   └── product.types.ts
└── validation/
    └── product.schema.ts
```

---

## 5. Design System

قبل بناء عشرات الصفحات، يجب إنشاء Design System موحد.

### المكوّنات الأساسية
```text
Colors         # نظام ألوان موحد
Typography     # خطوط وأحجام
Spacing        # نظام تباعد
Buttons        # أنواع الأزرار
Inputs         # حقول الإدخال
Selects        # القوائم المنسدلة
Tables         # جداول البيانات
Cards          # البطاقات
Badges         # الشارات
Dialogs        # النوافذ المنبثقة
Drawers        # الأدراج الجانبية
Tabs           # التبويبات
Dropdowns      # القوائم المنسدلة
Tooltips       # التلميحات
Toasts         # الإشعارات
Charts         # الرسوم البيانية
Forms          # النماذج
Loading States # حالات التحميل
Empty States   # الحالات الفارغة
Error States   # حالات الخطأ
```

### Design Tokens

استخدام Design Tokens بدل القيم الثابتة:

```css
:root {
  /* الألوان */
  --color-primary: #3B82F6;
  --color-background: #FAFAFA;
  --color-surface: #FFFFFF;
  --color-text: #18181B;
  --color-muted: #71717A;
  --color-success: #16A34A;
  --color-warning: #D97706;
  --color-danger: #DC2626;

  /* التباعد */
  --space-xs: 4px;
  --space-sm: 8px;
  --space-md: 16px;
  --space-lg: 24px;
  --space-xl: 32px;

  /* الزوايا */
  --radius-sm: 6px;
  --radius-md: 10px;
  --radius-lg: 16px;

  /* الظلال */
  --shadow-sm: 0 1px 2px rgba(0,0,0,0.05);
  --shadow-md: 0 4px 12px rgba(0,0,0,0.08);
}
```

هذا يسمح بتغيير هوية النظام مستقبلًا بدون إعادة تصميم الواجهة كاملة.

---

## 6. نظام الألوان

### المنطق
لون أساسي واحد (Primary) للعلامة والإجراءات، رمادي محايد للخلفيات والنصوص، وألوان دلالية ثابتة المعنى.

### لوحة الألوان المقترحة

#### الألوان المحايدة (Neutral Scale)
```css
--gray-50:  #FAFAFA  /* خلفية فاتحة */
--gray-100: #F4F4F5  /* خلفية ثانوية */
--gray-200: #E4E4E7  /* حدود خفيفة */
--gray-300: #D4D4D8  /* حدود عادية */
--gray-400: #A1A1AA  /* نص ثانوي */
--gray-500: #71717A  /* نص عادي */
--gray-600: #52525B  /* نص داكن */
--gray-700: #3F3F46  /* عناوين */
--gray-800: #27272A  /* عناوين داكنة */
--gray-900: #18181B  /* نص أساسي */
```

#### اللون الأساسي (Primary)
```css
--primary-50:  #EFF6FF
--primary-100: #DBEAFE
--primary-400: #60A5FA
--primary-500: #3B82F6  /* اللون الأساسي */
--primary-600: #2563EB
--primary-700: #1D4ED8
```

#### الألوان الدلالية
| الاستخدام | اللون | الكود |
|-----------|-------|-------|
| نجاح / مكتمل / مدفوع | أخضر | `#16A34A` |
| تحذير / مخزون منخفض | برتقالي | `#D97706` |
| خطر / دين متأخر / خطأ | أحمر | `#DC2626` |
| معلومة / إشعار عام | أزرق فاتح | `#0284C7` |
| قطعة مستعملة | بنفسجي | `#7C3AED` |

### Dark Mode
```css
--bg-primary-dark:   #0F0F11
--bg-surface-dark:   #18181B
--border-dark:       #27272A
--text-primary-dark: #FAFAFA
--text-muted-dark:   #A1A1AA
```

### قواعد الاستخدام
- لا يُستخدم أكثر من لونين دلاليين في نفس الشاشة
- الألوان الدلالية للحالة فقط، لا للتزيين
- عدم الاعتماد على اللون وحده (أضف أيقونات/نص)

---

## 7. الطباعة (Typography)

### الخط المقترح
- **IBM Plex Sans Arabic** للعربية
- **Inter** للإنجليزية والأرقام
- دعم كامل للعربية والعبرية والإنجليزية

### السلم الطباعي (Type Scale)
| المستوى | الحجم | الوزن | الاستخدام |
|---------|-------|-------|-----------|
| Display | 32px | Bold | أرقام Dashboard الكبرى |
| H1 | 24px | Semibold | عنوان الصفحة |
| H2 | 20px | Semibold | عنوان قسم/بطاقة |
| H3 | 16px | Medium | عناوين فرعية |
| Body | 14px | Regular | النص العام، الجداول |
| Small | 12px | Regular | تسميات، Timestamps |
| Micro | 11px | Medium | Badges، حالات صغيرة |

### قواعد عملية
- الأرقام المالية بخط أثقل من النص المحيط
- الأرقام في الجداول بعرض ثابت (Tabular numerals)
- الأرقام والعملة تبقى LTR داخل جمل RTL

---

## 8. RTL/LTR Support

النظام يدعم العربية والعبرية (RTL) والإنجليزية (LTR).

### القواعد التقنية
- استخدام `margin-inline-start/end` بدل `margin-left/right`
- استخدام `padding-inline-start/end` بدل `padding-left/right`
- `flex-direction` يُقرأ من اتجاه الصفحة
- الأيقونات ذات الاتجاه تُعكس تلقائيًا في RTL

### العناصر التي تتغير
- Sidebar
- Tables
- Forms
- Icons
- Breadcrumbs
- Dropdowns
- Drawers
- Navigation
- Alignment

### التحقق
- اختبار كل شاشة في الاتجاهين قبل الاعتماد
- تبديل اللغة فوري دون إعادة تحميل كاملة

---

## 9. Responsive Design

### Breakpoints
```text
Mobile:   < 640px   → عمود واحد، Bottom Nav
Tablet:   640-1024  → عمودان، Sidebar قابل للطي
Desktop:  > 1024px  → تخطيط كامل، Sidebar ثابت
```

### نمط التنقل حسب المنصة
| المنصة | نمط التنقل |
|--------|-------------|
| Desktop | Sidebar ثابت بأيقونات + نص |
| Tablet | Sidebar قابل للطي |
| Mobile | Bottom Navigation بـ4-5 عناصر |

### كثافة المعلومات
- **Desktop**: كثافة عالية - جداول كثيفة، عرض متعدد الأعمدة
- **Mobile**: كثافة منخفضة - عنصر واحد يتصدر الشاشة

### إعادة التدفق لا إعادة التصغير
لا يُعاد استخدام نفس تخطيط Desktop مصغّرًا على الموبايل — بل يُعاد تصميم تدفق الشاشة.

---

## 10. الأيقونات

### المكتبة المقترحة
- **Lucide Icons** - مجموعة متسقة وخفيفة
- Outline style موحّد
- وزن خط ثابت

### قواعد الاستخدام
- الأيقونة دائمًا مرفقة بنص عند أول ظهور
- عدم الاعتماد على الأيقونة وحدها
- الأيقونات ذات الاتجاه تُعكس في RTL

---

## 11. مكتبة المكوّنات

### الأزرار (Buttons)
| النوع | الاستخدام |
|-------|-----------|
| Primary | إجراء رئيسي واحد فقط لكل شاشة |
| Secondary | إجراءات ثانوية (إلغاء، رجوع) |
| Destructive | حذف/إرجاع - أحمر مع تأكيد |
| Icon Button | إجراءات سريعة متكررة |

- ارتفاع الزر على الموبايل ≥ 44px
- الزر الأساسي في شاشة البيع يكون الأكبر والأوضح

### البطاقات (Cards)
- حواف دائرية خفيفة (8-12px)
- ظل خفيف جدًا
- حدود رفيعة (1px)

### الجداول (Data Tables)
- Desktop: أعمدة قابلة للفرز، فلاتر، تحديد متعدد
- Mobile: يتحول إلى قائمة بطاقات مصغّرة

### الشارات (Status Badges)
كل حالة لها لون وشكل ثابت:
```
[ متوفر ]   أخضر فاتح + نص أخضر داكن
[ نافد ]    رمادي
[ مستعمل ]  بنفسجي فاتح
[ متأخر ]   أحمر فاتح
```

### النماذج (Forms)
- حقل واحد لكل سطر على الموبايل
- Validation فوري أثناء الكتابة
- رسائل الخطأ أسفل الحقل مباشرة

### الحوارات (Modals / Sheets)
- Desktop: Modal مركزي
- Mobile: Bottom Sheet أسهل بالإبهام

### الإشعارات (Toasts)
- Toast صغير يختفي خلال 3 ثوانٍ
- مركز إشعارات دائم للتنبيهات المهمة

---

## 12. الشاشات المحورية

### Dashboard
أرقام كبيرة وواضحة أولًا (مبيعات اليوم، الربح، الديون)، ثم قسم "النظام يقترح" بتنبيهات قابلة للنقر.

```
صباح الخير، أحمد

اليوم
₪7,450 المبيعات
₪1,820 الربح
₪24,800 الديون

يحتاج انتباهك
🔴 3 ديون متأخرة
🟠 4 قطع منخفضة المخزون
🟠 2 قطع مستعملة تحتاج فحص

عمليات سريعة
[ بيع ] [ Scan ] [ إضافة قطعة ] [ إضافة عميل ] [ تسجيل دفعة ]
```

### شاشة البيع (POS)
أكبر مساحة للـBarcode/البحث، سلة جانبية واضحة، زر الدفع ثابت ومرئي دومًا.

```
بيع جديد
[ Search / Scan ]

Cart
----------------
RTX 4070       ₪2350
RAM 16GB        ₪250

Total          ₪2600

Customer: Cash Customer
Payment: Cash / Card / Transfer / Debt

[ إتمام البيع ]
```

### بطاقة القطعة
صورة/أيقونة، السعر والربح بارزين، Timeline زمني بسيط:

```
RTX 4070
ASUS
USED • GRADE A

₪2,350 Selling Price
₪1,850 Cost
₪500 Expected Profit

Stock: 1
Barcode: XXXXX
Serial: XXXXX
Location: B-03
Warranty: 30 days

[ بيع ] [ تعديل ]
```

### ملف الزبون
الرصيد المستحق في أعلى البطاقة بخط كبير:

```
Ahmed
Outstanding ₪1,250
Total Purchases ₪8,450
Last Purchase 2 days ago

Timeline:
Sale → Payment → Sale → Return → Payment
```

### التقارير
رسوم بيانية بسيطة مع خيار "عرض التفاصيل" عند الطلب:

```
كيف كان هذا الشهر؟
Sales ₪74,500
Profit ₪18,240
Expenses ₪7,300
Net Profit ₪10,940

أكثر المنتجات مبيعًا
أكثر المنتجات ربحًا
المخزون الراكد
الديون
```

---

## 13. Barcode Experience

Barcode يجب أن يكون أحد أهم عناصر تجربة الاستخدام.

### Workflow
```
Scan
↓
Camera Scanner أو External Scanner
↓
RTX 4070
USED • GRADE A
₪2,350
Stock: 1
[ بيع ]
```

### الأدوات
- Camera Scanner للموبايل
- External Barcode Scanner لـWindows
- Command Palette (Ctrl+K) للبحث السريع

### Performance
عملية Scan يجب أن تكون فورية بدون Loading screen كامل كل مرة.

---

## 14. Dashboard Design

### المبدأ
لا تجعل النظام ينتظر أن يسأل المستخدم "ما الذي يحدث؟"
بل النظام يخبره "هذا ما يحتاج انتباهك الآن."

### العناصر
1. **أرقام اليوم** - مبيعات، ربح، ديون
2. **تنبيهات قابلة للإجراء** - مخزون منخفض، ديون متأخرة
3. **عمليات سريعة** - بيع، Scan، إضافة قطعة
4. **رسوم بيانية** - اتجاهات بسيطة

---

## 15. Navigation Design

### Desktop Navigation
```
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

### Mobile Navigation
```
Home
Sales
Scan
More
```

- Scan بارز جدًا على الموبايل
- Sidebar واضح وغير عريض على Desktop
- يحافظ على الحالة

---

## 16. State Management

### فصل State

#### Server State
- Products
- Sales
- Customers
- Inventory
- Reports

#### UI State
- Modal
- Sidebar
- Filters
- Selected Product
- Theme
- Language

### الأدوات المقترحة
- **React Query** للـServer State
- **Zustand** للـUI State

### API Layer
```
Component
↓
Hook
↓
Service
↓
API
```

مثال:
```
ProductPage
↓
useProduct()
↓
productService
↓
Backend API
```

---

## 17. API Layer

### البنية
```text
services/
├── apiClient.ts        # API Client موحد
├── interceptors.ts     # Request/Response interceptors
├── productService.ts   # Products API
├── salesService.ts     # Sales API
└── ...
```

### المميزات
- Interceptors للـAuthentication
- Error handling موحد
- Type-safe requests
- Retry logic

---

## 18. Validation

### Frontend Validation
- Real-time validation أثناء الكتابة
- React Hook Form + Zod
- رسائل خطأ واضحة باللغة المحلية

### العناصر المطلوب التحقق منها
- السعر
- الكمية
- Barcode
- Serial
- Customer
- Payment

### المثال
```typescript
const productSchema = z.object({
  name: z.string().min(1, 'الاسم مطلوب'),
  price: z.number().min(0, 'السعر يجب أن يكون موجبًا'),
  barcode: z.string().min(1, 'الباركود مطلوب'),
});
```

---

## 19. Security & Roles

### Role-Based UI
| الدور | الصلاحيات |
|-------|-----------|
| Owner | كل شيء |
| Manager | المخزون والمبيعات والتقارير |
| Employee | البيع والمخزون والعمليات المسموحة |
| Accountant | التقارير والبيانات المالية |

### Security UX
- الموظف لا يرى Profit Margin و Cost Price
- Backend يفرض الصلاحيات أيضًا
- Audit Log واضح للتغييرات

### Audit Log
```
Ahmed changed price
Old: ₪1,150
New: ₪1,250
18 Aug 2026 14:32
```

---

## 20. Performance

### التقنيات
- Pagination
- Lazy Loading
- Code Splitting
- Virtualized Lists
- Image Optimization
- Caching
- Efficient API calls

### الأهداف
- استجابة فورية مع آلاف المنتجات
- Scan بدون Loading screen
- Dashboard أقل من ثانية

---

## 21. Accessibility

### المعايير
- Keyboard navigation
- Focus states واضحة
- Contrast WCAG AA
- Screen reader labels
- Touch targets ≥ 44×44px
- عدم الاعتماد على اللون فقط

### التحقق
- اختبار مع قارئ الشاشة
- اختبار بلوحة المفاتيح فقط
- اختبار التباين اللوني

---

## 22. Dark Mode

### الدعم
```text
Light
Dark
System
```

### التنفيذ
- CSS Variables من البداية
- ليس مجرد invert colors
- تصميم بعناية لكل وضع

---

## 23. Animations

### القواعد
- Animation تستخدم فقط عندما تفيد
- مدة مناسبة: 150-250ms
- تأكيد العمليات الحرجة بحركة بسيطة
- لا Animations زخرفية طويلة

### الأمثلة
- فتح Drawer
- انتقال Modal
- Success checkmark
- Loading states

---

## 24. Testing Strategy

### Unit Tests
- Utilities
- Validation
- Calculations

### Component Tests
- Forms
- Buttons
- Scanner states

### Integration Tests
- Sale flow
- Payment flow
- Product creation

### E2E Tests
```
Login → Scan → Add to Cart → Sell → Payment → Dashboard Update
```

### Mobile Testing
- اختبار حقيقي على أجهزة مختلفة
- Camera, Touch, Barcode, Keyboard
- Orientation, Network changes

---

## 25. معايير الجودة

### Design QA Checklist
قبل اعتماد أي شاشة:
```
☐ تعمل بصريًا في RTL و LTR
☐ تعمل على Mobile / Tablet / Desktop
☐ الحالات الأربع مصممة (Loading/Empty/Error/Populated)
☐ التباين اللوني يحقق WCAG AA
☐ لا يوجد أكثر من إجراء أساسي واحد بارز
☐ الأرقام المالية بارزة ومقروءة فورًا
☐ متسقة مع Design Tokens
☐ اختبار الاستخدام بإبهام واحد على الموبايل
```

### قواعد UX المهمة

#### قاعدة "3 Seconds"
عند فتح أي شاشة، يجب أن يعرف المستخدم خلال 3 ثوانٍ:
1. أين أنا؟
2. ماذا أرى؟
3. ماذا يمكنني أن أفعل؟

#### قاعدة "One Primary Action"
كل شاشة:
```
One Main Goal
One Primary CTA
```

#### قاعدة "No Dead Ends"
كل شاشة يجب أن تقود إلى إجراء منطقي.

#### قاعدة "Faster Over Time"
كل عملية متكررة يجب أن تصبح أسرع مع الوقت.

---

## الخلاصة

الـFrontend الناجح لهذا المشروع يجب أن يكون:

**بسيطًا للمستخدم، قويًا في الخلفية، سريعًا في العمليات، واضحًا في الأموال والمخزون، ومميزًا بصريًا دون مبالغة.**

الفكرة التي تحكم كل قرار تصميمي:

> **"النظام يعمل من أجل صاحب المحل، وليس صاحب المحل من أجل النظام."**

المنتج النهائي يجمع:
```
Modern UI
Fast UX
Barcode First
Mobile + Desktop
RTL/LTR
Smart Alerts
Automation
Business Intelligence
Simple Workflows
```

النتيجة: **Smart Store Operating System** - ليس مجرد POS.

---

## معايير إطلاق الشاشات

قبل إطلاق أي شاشة اسأل:

> **هل تجعل هذه الشاشة حياة صاحب المحل أسهل؟**

- نعم → احتفظ بها
- لا، لكنها تبدو جميلة → أعد تصميمها
- لا يحتاجها المستخدم أصلًا → احذفها

---

## الرؤية التجارية

إذا تم تنفيذ الـFrontend بهذه الطريقة، فالمنتج سيصبح أساس منتج SaaS يمكن تطويره لاحقًا لدعم:

- عدة محلات وفروع
- عدة مستخدمين وصلاحيات
- اشتراكات
- تقارير متقدمة
- Automation
- AI
- Integrations
- أجهزة Barcode وطابعات
- خدمات خارجية

---

## بنية المشروع النهائية

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

## أوامر التطوير

```bash
cd frontend
npm install              # تثبيت التبعيات
npm run dev              # تشغيل خادم التطوير
npm run build            # بناء للإنتاج
npm run lint             # فحص الكود
npm run test             # تشغيل الاختبارات
```

---

**هذه هي هوية الـFrontend التي أنصح أن تبني عليها المشروع بالكامل.**
