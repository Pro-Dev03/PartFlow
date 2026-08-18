// Arabic error messages mapping
const errorMessages: Record<string, string> = {
  // Network errors
  NETWORK_ERROR: 'مشكلة في الاتصال، جاري المحاولة...',
  TIMEOUT_ERROR: 'انتهت مهلة الطلب، يرجى المحاولة مرة أخرى',
  CONNECTION_ERROR: 'لا يمكن الاتصال بالخادم',

  // Authentication errors
  UNAUTHORIZED: 'الجلسة منتهت، يرجى تسجيل الدخول',
  FORBIDDEN: 'ليس لديك صلاحية للقيام بهذا الإجراء',
  TOKEN_EXPIRED: 'انتهت صلاحية الجلسة',
  INVALID_CREDENTIALS: 'بيانات الدخول غير صحيحة',

  // Validation errors
  VALIDATION_ERROR: 'بيانات غير صحيحة، يرجى التحقق من المدخلات',
  REQUIRED_FIELD: 'هذا الحقل مطلوب',
  INVALID_FORMAT: 'تنسيق غير صحيح',
  INVALID_EMAIL: 'البريد الإلكتروني غير صحيح',
  INVALID_PHONE: 'رقم الهاتف غير صحيح',

  // Business logic errors
  ITEM_ALREADY_SOLD: 'هذه القطعة تم بيعها بالفعل',
  LOW_STOCK: 'المخزون منخفض جداً',
  OUT_OF_STOCK: 'المنتج غير متوفر',
  INSUFFICIENT_STOCK: 'الكمية المطلوبة غير متوفرة',
  PRODUCT_NOT_FOUND: 'المنتج غير موجود',
  CUSTOMER_NOT_FOUND: 'العميل غير موجود',
  SUPPLIER_NOT_FOUND: 'المورد غير موجود',

  // Payment errors
  PAYMENT_FAILED: 'فشلت عملية الدفع',
  INSUFFICIENT_FUNDS: 'رصيد غير كافٍ',
  PAYMENT_ALREADY_PROCESSED: 'تمت معالجة الدفع مسبقاً',

  // Debt errors
  DEBT_LIMIT_EXCEEDED: 'تم تجاوز حد الائتمان',
  DEBT_ALREADY_PAID: 'تم سداد هذا الدين مسبقاً',
  OVERDUE_DEBT: 'هذا الدين متأخر',

  // General errors
  UNKNOWN_ERROR: 'حدث خطأ غير متوقع',
  SERVER_ERROR: 'خطأ في الخادم، يرجى المحاولة لاحقاً',
  DATABASE_ERROR: 'خطأ في قاعدة البيانات',
};

export function getErrorMessage(error: any): string {
  if (typeof error === 'string') {
    return errorMessages[error] || error;
  }

  if (error?.response?.data?.error?.code) {
    const code = error.response.data.error.code;
    return errorMessages[code] || error.response.data.error.message || errorMessages.UNKNOWN_ERROR;
  }

  if (error?.message) {
    return errorMessages[error.message] || error.message;
  }

  if (error?.code) {
    return errorMessages[error.code] || errorMessages.UNKNOWN_ERROR;
  }

  return errorMessages.UNKNOWN_ERROR;
}

export function getErrorType(error: any): 'error' | 'warning' | 'info' {
  const message = getErrorMessage(error).toLowerCase();
  
  if (message.includes('منخفض') || message.includes('تنبيه')) {
    return 'warning';
  }
  
  if (message.includes('تم') || message.includes('نجح')) {
    return 'info';
  }
  
  return 'error';
}

export function isNetworkError(error: any): boolean {
  if (!error) return false;
  
  const message = getErrorMessage(error).toLowerCase();
  return message.includes('اتصال') || message.includes('شبكة') || message.includes('خادم');
}

export function isAuthError(error: any): boolean {
  if (!error) return false;
  
  const message = getErrorMessage(error).toLowerCase();
  return message.includes('جلسة') || message.includes('دخول') || message.includes('صلاحية');
}

export function shouldRetry(error: any): boolean {
  return isNetworkError(error);
}

export class AppError extends Error {
  code: string;
  originalError?: any;

  constructor(code: string, message?: string, originalError?: any) {
    super(message || getErrorMessage({ code }));
    this.name = 'AppError';
    this.code = code;
    this.originalError = originalError;
  }
}

export function handleApiError(error: any): string {
  console.error('API Error:', error);
  
  if (isAuthError(error)) {
    // Redirect to login or trigger re-auth
    setTimeout(() => {
      window.location.href = '/login';
    }, 2000);
  }
  
  return getErrorMessage(error);
}
