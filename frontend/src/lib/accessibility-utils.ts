/**
 * Accessibility Utilities - أدوات إمكانية الوصول
 */

// إضافة ARIA labels للعناصر التفاعلية
export const getAriaProps = (label: string, description?: string) => ({
  'aria-label': label,
  ...(description && { 'aria-description': description }),
});

// إضافة ARIA labels للأزرار
export const getButtonAriaProps = (label: string, pressed?: boolean) => ({
  'aria-label': label,
  ...(pressed !== undefined && { 'aria-pressed': pressed }),
  role: 'button',
});

// إضافة ARIA labels للروابط
export const getLinkAriaProps = (label: string) => ({
  'aria-label': label,
  role: 'link',
});

// إضافة ARIA labels للحقول الإدخال
export const getInputAriaProps = (label: string, required?: boolean, invalid?: boolean) => ({
  'aria-label': label,
  ...(required && { 'aria-required': true }),
  ...(invalid && { 'aria-invalid': true }),
  'aria-describedby': required ? `${label}-error` : undefined,
});

// إضافة ARIA labels للجداول
export const getTableAriaProps = (label: string) => ({
  'aria-label': label,
  role: 'table',
});

// إضافة ARIA labels للمودال
export const getModalAriaProps = (label: string) => ({
  'aria-label': label,
  role: 'dialog',
  'aria-modal': true,
});

// إضافة ARIA labels للإشعارات
export const getNotificationAriaProps = (type: 'success' | 'error' | 'warning' | 'info') => ({
  'aria-live': 'polite',
  'aria-atomic': true,
  'role': 'status',
  'aria-label': type === 'success' ? 'إشعار نجاح' : 
              type === 'error' ? 'إشعار خطأ' : 
              type === 'warning' ? 'إشعار تحذير' : 'إشعار معلومات',
});

// إضافة ARIA labels للتنقل
export const getNavAriaProps = (label: string) => ({
  'aria-label': label,
  role: 'navigation',
});

// إضافة ARIA labels للبحث
export const getSearchAriaProps = (label: string) => ({
  'aria-label': label,
  role: 'search',
});

// إضافة ARIA labels للمحتوى الرئيسي
export const getMainAriaProps = (label: string) => ({
  'aria-label': label,
  role: 'main',
});

// إضافة ARIA labels للمناطق
export const getRegionAriaProps = (label: string) => ({
  'aria-label': label,
  role: 'region',
});

// إضافة ARIA labels للأيقونات
export const getIconAriaProps = (label: string, decorative?: boolean) => ({
  'aria-label': decorative ? undefined : label,
  'aria-hidden': decorative,
  role: decorative ? 'presentation' : 'img',
});

// إضافة ARIA labels للتحميل
export const getLoadingAriaProps = (label: string) => ({
  'aria-label': label,
  'aria-busy': true,
  'aria-live': 'polite',
});

// إضافة ARIA labels للتقدم
export const getProgressAriaProps = (label: string, value: number, max: number) => ({
  'aria-label': label,
  'aria-valuenow': value,
  'aria-valuemin': 0,
  'aria-valuemax': max,
  'role': 'progressbar',
});

// إضافة ARIA labels للتبويبات
export const getTabAriaProps = (label: string, selected?: boolean) => ({
  'aria-label': label,
  'aria-selected': selected,
  role: 'tab',
  tabIndex: selected ? 0 : -1,
});

// إضافة ARIA labels للقوائم
export const getMenuAriaProps = (label: string) => ({
  'aria-label': label,
  role: 'menu',
});

// إضافة ARIA labels لعناصر القائمة
export const getMenuItemAriaProps = (label: string) => ({
  'aria-label': label,
  role: 'menuitem',
});

// إضافة ARIA labels للـ tooltips
export const getTooltipAriaProps = (label: string) => ({
  'aria-label': label,
  'aria-describedby': `${label}-tooltip`,
});

// إضافة ARIA labels للـ alerts
export const getAlertAriaProps = (type: 'success' | 'error' | 'warning' | 'info') => ({
  'aria-live': 'assertive',
  'aria-atomic': true,
  'role': 'alert',
  'aria-label': type === 'success' ? 'تنبيه نجاح' : 
              type === 'error' ? 'تنبيه خطأ' : 
              type === 'warning' ? 'تنبيه تحذير' : 'تنبيه معلومات',
});