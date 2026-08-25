/**
 * Centralized Button Size Configuration
 * يوفر أحجام موحدة لجميع الأزرار في المشروع
 */

export const ButtonSizes = {
  // حجم احترافي صغير لجميع الأزرار - مثل زر التوصية
  sm: 'sm',

  // حجم متوسط للأزرار الرئيسية (نادر الاستخدام)
  md: 'md',

  // حجم كبير للأزرار المهمة جداً (نادر الاستخدام)
  lg: 'lg',

  // حجم للأيقونات فقط
  icon: 'icon'
} as const;

export type ButtonSize = typeof ButtonSizes[keyof typeof ButtonSizes];

/**
 * أحجام الأزرار حسب السياق
 * جميع الأزرار تستخدم الحجم الاحترافي الصغير مثل زر التوصية
 */
export const ButtonSizeByContext = {
  // أزرار التوصيات في AI Insights
  recommendation: ButtonSizes.sm,

  // أزرار الإجراءات الرئيسية في Page Headers
  pageHeaderAction: ButtonSizes.sm,

  // أزرار الجداول والحركات السريعة
  tableAction: ButtonSizes.sm,

  // أزرار النماذج والنوافذ المنبثقة
  modalAction: ButtonSizes.sm,

  // أزرار الأيقونات
  iconAction: ButtonSizes.icon,

  // أزرار الإجراءات المهمة
  primaryAction: ButtonSizes.sm
} as const;

/**
 * الأقسام التي تستخدم أحجام محددة
 * جميع الأزرار تستخدم الحجم الاحترافي الصغير مثل زر التوصية
 */
export const ButtonSizeBySection = {
  // قسم العملاء
  customers: {
    headerActions: ButtonSizes.sm,
    recommendation: ButtonSizes.sm,
    tableAction: ButtonSizes.sm,
    modalAction: ButtonSizes.sm
  },

  // قسم المخزون
  inventory: {
    headerActions: ButtonSizes.sm,
    recommendation: ButtonSizes.sm,
    tableAction: ButtonSizes.sm,
    modalAction: ButtonSizes.sm
  },

  // قسم الديون
  debts: {
    headerActions: ButtonSizes.sm,
    recommendation: ButtonSizes.sm,
    tableAction: ButtonSizes.sm,
    modalAction: ButtonSizes.sm
  },

  // قسم المبيعات
  sales: {
    headerActions: ButtonSizes.sm,
    tableAction: ButtonSizes.sm,
    modalAction: ButtonSizes.sm
  },

  // قسم التقارير
  reports: {
    headerActions: ButtonSizes.sm,
    recommendation: ButtonSizes.sm,
    tableAction: ButtonSizes.sm
  },

  // قسم المرتجعات
  returns: {
    headerActions: ButtonSizes.sm,
    tableAction: ButtonSizes.sm,
    modalAction: ButtonSizes.sm
  },

  // قسم الموردين
  suppliers: {
    headerActions: ButtonSizes.sm,
    tableAction: ButtonSizes.sm,
    modalAction: ButtonSizes.sm
  },

  // قسم القطع المستعملة
  usedparts: {
    headerActions: ButtonSizes.sm,
    tableAction: ButtonSizes.sm,
    modalAction: ButtonSizes.sm
  },

  // قسم أنواع القطع
  parttypes: {
    headerActions: ButtonSizes.sm,
    tableAction: ButtonSizes.sm,
    modalAction: ButtonSizes.sm
  },

  // قسم التصنيفات
  categories: {
    headerActions: ButtonSizes.sm,
    tableAction: ButtonSizes.sm,
    modalAction: ButtonSizes.sm
  }
} as const;

export type SectionName = keyof typeof ButtonSizeBySection;
export type SectionContext =
  | 'headerActions'
  | 'recommendation'
  | 'tableAction'
  | 'modalAction'
  | 'iconAction'
  | 'pageHeaderAction'
  | 'primaryAction';

/**
 * دالة للحصول على حجم الزر المناسب للقسم والسياق
 */
export function getButtonSize(section: SectionName, context: SectionContext): ButtonSize {
  const sectionConfig = ButtonSizeBySection[section] as Record<string, ButtonSize>;
  return sectionConfig[context] ?? ButtonSizes.sm;
}

/**
 * دالة للحصول على حجم الزر حسب السياق العام
 */
export function getButtonSizeByContext(context: keyof typeof ButtonSizeByContext): ButtonSize {
  return ButtonSizeByContext[context];
}