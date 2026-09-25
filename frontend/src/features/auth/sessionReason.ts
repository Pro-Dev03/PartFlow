export type AutoLogoutReason =
  | 'offline'
  | 'subscription'
  | 'account-deleted'
  | 'account-suspended'
  | 'session-expired'
  | 'missing-session'
  | 'cloud-rejected'
  | 'unknown';

export const AUTO_LOGOUT_REASON_KEY = 'partflow-auto-logout-reason';

const reasonMessages: Record<AutoLogoutReason, { title: string; message: string }> = {
  offline: {
    title: 'تم إيقاف الجلسة مؤقتًا',
    message: 'فُقد اتصال الإنترنت، لذلك تم إيقاف الجلسة لحماية حسابك. تحقق من الاتصال ثم سجّل الدخول مجددًا.',
  },
  subscription: {
    title: 'انتهى الاشتراك أو تم تعطيل الحساب',
    message: 'لم يعد الحساب مصرحًا له بالوصول. تواصل مع الإدارة لتجديد الاشتراك أو إعادة تفعيل الحساب.',
  },
  'account-deleted': {
    title: 'تم حذف هذا المستخدم',
    message: 'لم يعد هذا المستخدم موجودًا في النظام. تواصل مع المسؤول إذا كنت تعتقد أن ذلك حدث بالخطأ.',
  },
  'account-suspended': {
    title: 'تم إيقاف هذا الحساب',
    message: 'أوقف المسؤول هذا الحساب. تواصل مع المسؤول لإعادة تفعيله.',
  },
  'session-expired': {
    title: 'انتهت جلسة الدخول',
    message: 'انتهت صلاحية جلسة الدخول. سجّل الدخول مرة أخرى للمتابعة.',
  },
  'missing-session': {
    title: 'تحتاج إلى تسجيل الدخول',
    message: 'لم يتم العثور على جلسة دخول محفوظة. أدخل بياناتك للمتابعة.',
  },
  'cloud-rejected': {
    title: 'تعذر التحقق من الحساب',
    message: 'رفض الخادم السحابي التحقق من الجلسة الحالية. حاول تسجيل الدخول مرة أخرى.',
  },
  unknown: {
    title: 'تم إنهاء الجلسة',
    message: 'تعذر الحفاظ على جلسة الدخول. سجّل الدخول مرة أخرى للمتابعة.',
  },
};

export function getAutoLogoutMessage(reason: AutoLogoutReason) {
  return reasonMessages[reason] ?? reasonMessages.unknown;
}

export function saveAutoLogoutReason(reason: AutoLogoutReason): void {
  if (typeof window !== 'undefined') {
    sessionStorage.setItem(AUTO_LOGOUT_REASON_KEY, reason);
  }
}

export function readAutoLogoutReason(): AutoLogoutReason | null {
  if (typeof window === 'undefined') return null;
  const reason = sessionStorage.getItem(AUTO_LOGOUT_REASON_KEY);
  return reason && reason in reasonMessages ? reason as AutoLogoutReason : null;
}

export function clearAutoLogoutReason(): void {
  if (typeof window !== 'undefined') {
    sessionStorage.removeItem(AUTO_LOGOUT_REASON_KEY);
  }
}
