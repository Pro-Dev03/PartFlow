# تحسينات مكون Modal

## الميزات الجديدة

تم تحديث مكون `Modal` ليشمل ميزات متقدمة للتنقل السريع وتحسين تجربة المستخدم:

### 1. التركيز التلقائي (Auto Focus)
- عند فتح أي نافذة منبثقة، يتم التركيز تلقائياً على أول حقل إدخال (input/select/textarea)
- يمكن تعطيل هذه الميزة عبر prop `autoFocus={false}`

### 2. التنقل بـ Enter (Enter Navigation)
- الضغط على زر Enter في حقل إدخال ينقلك تلقائياً للحقل التالي
- يتخطى الأزرار وينتقل فقط بين حقول الإدخال
- يمكن تعطيل هذه الميزة عبر prop `enableEnterNavigation={false}`

### 3. تحسين التنقل بـ TAB
- دعم محسّن للتنقل بـ TAB + Shift للرجوع للخلف
- حلقة مستمرة داخل النافذة (من آخر عنصر يرجع لأول عنصر)

### 4. التصميم المحسّن
- تحديث تصميم `modern` ليطابق أحدث معايير التصميم
- ظلال متعددة الطبقات لعمق أفضل
- حدود ناعمة وواضحة

## الاستخدام

### الاستخدام الأساسي
```tsx
import { Modal } from '../../components/ui/modal';

<Modal
  isOpen={isOpen}
  onClose={() => setIsOpen(false)}
  title="إضافة قطعة جديدة"
  variant="modern"
  size="lg"
>
  <form>
    <input placeholder="اسم المنتج" />
    <input placeholder="الباركود" />
    <select>
      <option>اختر التصنيف</option>
    </select>
  </form>
</Modal>
```

### مع التحكم في الميزات
```tsx
<Modal
  isOpen={isOpen}
  onClose={() => setIsOpen(false)}
  title="إضافة قطعة جديدة"
  variant="modern"
  size="lg"
  autoFocus={true}           // التركيز التلقائي (افتراضي: true)
  enableEnterNavigation={true} // التنقل بـ Enter (افتراضي: true)
>
  {/* المحتوى */}
</Modal>
```

### تعطيل ميزة معينة
```tsx
<Modal
  isOpen={isOpen}
  onClose={() => setIsOpen(false)}
  title="عرض تفاصيل"
  variant="modern"
  size="md"
  autoFocus={false}          // لا تركز تلقائياً
  enableEnterNavigation={false} // لا تستخدم Enter للتنقل
>
  {/* محتوى للقراءة فقط */}
</Modal>
```

## اختصارات لوحة المفاتيح

| المفتاح | الوظيفة |
|---------|---------|
| `ESC` | إغلاق النافذة |
| `TAB` | الانتقال للحقل التالي |
| `Shift + TAB` | الرجوع للحقل السابق |
| `Enter` | الانتقال للحقل التالي (في حقول الإدخال) |

## المتغيرات المتاحة (Variants)

- `default`: التصميم الافتراضي البسيط
- `modern`: التصميم الحديث مع ظلال متعددة (الأكثر استخداماً)
- `elegant`: تصميم أنيق مع تدرج لوني
- `glass`: تصميم زجاجي مع تأثير blur

## الأحجام المتاحة (Sizes)

- `sm`: 400px
- `md`: 500px
- `lg`: 600px
- `xl`: 800px
- `2xl`: 1000px
- `full`: عرض كامل

## أفضل الممارسات

1. **للنماذج (Forms)**: استخدم `autoFocus={true}` و `enableEnterNavigation={true}`
2. **للعرض فقط**: استخدم `autoFocus={false}` و `enableEnterNavigation={false}`
3. **للنوافذ المعقدة**: استخدم حجم `lg` أو `xl`
4. **للنوافذ البسيطة**: استخدم حجم `sm` أو `md`

## التوافق

المكون متوافق مع:
- React 18+
- TypeScript
- جميع المتصفحات الحديثة
- الوضع الفاتح والداكن
- الأجهزة المحمولة (responsive)
