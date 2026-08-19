# PartFlow Design Tokens Documentation

## نظرة عامة

هذا المستند يوثق نظام Design Tokens الخاص بـ PartFlow. هذه القيم البصرية الأساسية تضمن الاتساق والاحترافية في جميع أنحاء التطبيق.

## استخدام Design Tokens

### في CSS
```css
.my-component {
  background-color: var(--bg-surface);
  color: var(--text-primary);
  padding: var(--spacing-4);
  border-radius: var(--radius-md);
}
```

### في Tailwind
```jsx
<div className="bg-surface text-primary p-4 rounded-md">
  المحتوى هنا
</div>
```

### في React Components
```jsx
const style = {
  backgroundColor: 'var(--bg-surface)',
  color: 'var(--text-primary)',
  padding: 'var(--spacing-4)',
  borderRadius: 'var(--radius-md)',
};
```

---

## الألوان (Colors)

### الخلفيات (Backgrounds)
- `--bg-background`: الخلفية الرئيسية (Dark: #0B0F14, Light: #F7F9FC)
- `--bg-surface`: خلفية العناصر (Dark: #11171F, Light: #FFFFFF)
- `--bg-surface-elevated`: خلفية العناصر المرتفعة (Dark: #171F29, Light: #FFFFFF)
- `--bg-border`: لون الحدود (Dark: #24303D, Light: #E2E8F0)

### الألوان الأساسية (Primary)
- `--primary`: اللون الأساسي (Dark: #3B82F6, Light: #2563EB)
- `--primary-hover`: لون التحويم (Dark: #2563EB, Light: #1D4ED8)
- `--primary-active`: لون النشط (Dark: #1D4ED8, Light: #1E40AF)

### ألوان الحالات (Status)
- `--success`: نجاح (Dark: #22C55E, Light: #16A34A)
- `--warning`: تحذير (Dark: #F59E0B, Light: #D97706)
- `--danger`: خطر (Dark: #EF4444, Light: #DC2626)
- `--info`: معلومة (Dark: #06B6D4, Light: #0891B2)

### ألوان النصوص (Text)
- `--text-primary`: النص الأساسي (Dark: #F9FAFB, Light: #111827)
- `--text-secondary`: النص الثانوي (Dark: #9CA3AF, Light: #6B7280)
- `--text-tertiary`: النص الثالث (Dark: #6B7280, Light: #9CA3AF)
- `--text-disabled`: النص المعطل (Dark: #4B5563, Light: #D1D5DB)

---

## الخطوط (Typography)

### أحجام الخطوط (Font Sizes)
- `--font-size-page-title`: 32px - عناوين الصفحات
- `--font-size-section-title`: 28px - عناوين الأقسام
- `--font-size-card-title`: 24px - عناوين البطاقات
- `--font-size-body`: 16px - النص العادي
- `--font-size-secondary`: 14px - النص الثانوي
- `--font-size-caption`: 12px - النص الصغير

### وزن الخطوط (Font Weights)
- `--font-weight-regular`: 400
- `--font-weight-medium`: 500
- `--font-weight-semibold`: 600
- `--font-weight-bold`: 700

### ارتفاع السطر (Line Heights)
- `--line-height-tight`: 1.25
- `--line-height-normal`: 1.5
- `--line-height-relaxed`: 1.75

### عائلات الخطوط (Font Families)
- `--font-family-arabic`: 'Cairo', 'Tajawal', system-ui, sans-serif
- `--font-family-latin`: 'Inter', system-ui, sans-serif

---

## المسافات (Spacing)

- `--spacing-1`: 4px
- `--spacing-2`: 8px
- `--spacing-3`: 12px
- `--spacing-4`: 16px
- `--spacing-5`: 20px
- `--spacing-6`: 24px
- `--spacing-8`: 32px
- `--spacing-10`: 40px
- `--spacing-12`: 48px
- `--spacing-16`: 64px

---

## زوايا الحدود (Border Radius)

- `--radius-sm`: 6px
- `--radius-md`: 8px
- `--radius-lg`: 12px
- `--radius-xl`: 16px
- `--radius-full`: 9999px

---

## الظلال (Shadows)

- `--shadow-sm`: ظل صغير
- `--shadow-md`: ظل متوسط
- `--shadow-lg`: ظل كبير
- `--shadow-xl`: ظل جداً كبير

---

## الانتقالات (Transitions)

### المدة (Duration)
- `--transition-fast`: 150ms
- `--transition-normal`: 200ms
- `--transition-slow`: 300ms

### التوقيت (Timing Functions)
- `--transition-timing-default`: cubic-bezier(0.4, 0, 0.2, 1)
- `--transition-timing-in`: cubic-bezier(0.4, 0, 1, 1)
- `--transition-timing-out`: cubic-bezier(0, 0, 0.2, 1)

---

## الطبقات (Z-Index)

- `--z-dropdown`: 1000
- `--z-sticky`: 1020
- `--z-fixed`: 1030
- `--z-modal-backdrop`: 1040
- `--z-modal`: 1050
- `--z-popover`: 1060
- `--z-tooltip`: 1070

---

## RTL/LTR Support

### المتغيرات المنطقية
- `--direction-start`: البداية (RTL: right, LTR: left)
- `--direction-end`: النهاية (RTL: left, LTR: right)
- `--font-family-base`: الخط الأساسي (RTL: Arabic, LTR: Latin)

---

## المكونات الخاصة (Component-Specific Tokens)

### الأزرار (Buttons)
- `--button-height-sm`: 32px
- `--button-height-md`: 40px
- `--button-height-lg`: 48px

### حقول الإدخال (Inputs)
- `--input-height-sm`: 32px
- `--input-height-md`: 40px
- `--input-height-lg`: 48px

### البطاقات (Cards)
- `--card-padding-sm`: var(--spacing-4)
- `--card-padding-md`: var(--spacing-6)
- `--card-padding-lg`: var(--spacing-8)

### الجداول (Tables)
- `--table-row-height`: 48px
- `--table-cell-padding`: var(--spacing-4)

### الشريط الجانبي (Sidebar)
- `--sidebar-width-expanded`: 260px
- `--sidebar-width-collapsed`: 64px

### الشريط العلوي (Topbar)
- `--topbar-height`: 64px

---

## أفضل الممارسات

### 1. استخدام Design Tokens بدلاً من القيم الثابتة
❌ **سيء:**
```css
background-color: #11171F;
padding: 16px;
border-radius: 8px;
```

✅ **جيد:**
```css
background-color: var(--bg-surface);
padding: var(--spacing-4);
border-radius: var(--radius-md);
```

### 2. استخدام فئات Tailwind المتكاملة مع Design Tokens
```jsx
<div className="bg-surface p-4 rounded-md">
  المحتوى هنا
</div>
```

### 3. الحفاظ على الاتساق
استخدم نفس القيم في جميع المكونات لضمان الاتساق البصري.

### 4. التجاوب مع السمات
جميع Design Tokens تدعم الوضع الليلي والنهاري تلقائياً.

---

## قواعد التسمية

- استخدم `--` كبادئة لجميع CSS Variables
- استخدم `kebab-case` للأسماء الطويلة
- كن واضحاً ووصفياً في التسمية
- استخدم البادئات لتجميع القيم المتشابهة (مثل `--color-`, `--spacing-`)

---

## الصيانة والتحديث

### إضافة قيمة جديدة
1. أضف القيمة في الملف المناسب (tokens.css أو themes.css)
2. أضف التوثيق في هذا الملف
3. حدث tailwind.config.js إذا لزم الأمر
4. اختبر القيمة في جميع السمات

### تعديل قيمة موجودة
1. عدل القيمة في ملف tokens.css
2. تأكد من تحديث جميع القيم المرتبطة
3. اختبر التغيير في جميع السمات
4. حدث التوثيق

---

## الأدوات والموارد

### اختبار Design Tokens
```css
.test-token {
  background-color: var(--bg-surface);
  color: var(--text-primary);
  padding: var(--spacing-4);
  border-radius: var(--radius-md);
  border: 1px solid var(--bg-border);
}
```

### Debugging
استخدم أدوات المطور في المتصفح لفحص قيم CSS Variables:
```javascript
getComputedStyle(document.documentElement).getPropertyValue('--bg-surface');
```

---

## المراجع

- [CSS Custom Properties](https://developer.mozilla.org/en-US/docs/Web/CSS/Using_CSS_custom_properties)
- [Design Tokens W3C Community Group](https://www.w3.org/community/design-tokens/)
- [Tailwind CSS Custom Properties](https://tailwindcss.com/docs/customizing-colors#using-css-variables)