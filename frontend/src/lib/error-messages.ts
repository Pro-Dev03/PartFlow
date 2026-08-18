// رسائل الأخطاء بالعربية لنظام PartFlow
// Arabic Error Messages for PartFlow System

export const errorMessages = {
  // Network Errors
  NETWORK_ERROR: 'مشكلة في الاتصال، جاري المحاولة...',
  TIMEOUT_ERROR: 'انتهت مهلة الاتصال',
  CONNECTION_FAILED: 'فشل الاتصال بالخادم',
  
  // API Errors
  ITEM_NOT_FOUND: 'المنتج غير موجود',
  ITEM_ALREADY_SOLD: 'هذه القطعة تم بيعها بالفعل',
  INSUFFICIENT_STOCK: 'المخزون غير كافٍ',
  DUPLICATE_SERIAL: 'الرقم التسلسلي مستخدم بالفعل',
  DUPLICATE_BARCODE: 'الباركود مستخدم بالفعل',
  INVALID_BARCODE: 'الباركود غير صالح',
  
  // Auth Errors
  UNAUTHORIZED: 'الجلسة منتهت، يرجى تسجيل الدخول',
  PERMISSION_DENIED: 'ليس لديك صلاحية للقيام بهذا الإجراء',
  INVALID_CREDENTIALS: 'البريد الإلكتروني أو كلمة المرور غير صحيحة',
  TOKEN_EXPIRED: 'انتهت صلاحية الجلسة',
  SESSION_EXPIRED: 'انتهت الجلسة، يرجى تسجيل الدخول مرة أخرى',
  
  // Validation Errors
  REQUIRED_FIELD: 'هذا الحقل مطلوب',
  INVALID_EMAIL: 'البريد الإلكتروني غير صالح',
  INVALID_PHONE: 'رقم الهاتف غير صالح',
  INVALID_NUMBER: 'القيمة يجب أن تكون رقماً',
  INVALID_DATE: 'التاريخ غير صالح',
  MIN_LENGTH: 'يجب أن يكون الحد الأدنى {min} أحرف',
  MAX_LENGTH: 'يجب أن يكون الحد الأقصى {max} أحرف',
  MIN_VALUE: 'يجب أن تكون القيمة على الأقل {min}',
  MAX_VALUE: 'يجب أن تكون القيمة على الأكثر {max}',
  POSITIVE_NUMBER: 'يجب أن تكون القيمة موجبة',
  
  // Business Logic Errors
  CUSTOMER_NOT_FOUND: 'العميل غير موجود',
  SUPPLIER_NOT_FOUND: 'المورد غير موجود',
  PRODUCT_NOT_FOUND: 'المنتج غير موجود',
  INVENTORY_ITEM_NOT_FOUND: 'القطعة غير موجودة',
  SALE_NOT_FOUND: 'عملية البيع غير موجودة',
  DEBT_NOT_FOUND: 'الدين غير موجود',
  PAYMENT_NOT_FOUND: 'الدفعة غير موجودة',
  
  INVENTORY_UPDATE_FAILED: 'فشل تحديث المخزون',
  SALE_CREATION_FAILED: 'فشل إنشاء عملية البيع',
  PAYMENT_FAILED: 'فشل عملية الدفع',
  DEBT_CREATION_FAILED: 'فشل إنشاء الدين',
  
  ALREADY_RESERVED: 'هذه القطعة محجوزة بالفعل',
  CANNOT_RESERVE_SOLD_ITEM: 'لا يمكن حجز قطعة تم بيعها',
  CANNOT_SOLD_RESERVED_ITEM: 'لا يمكن بيع قطعة محجوزة',
  
  RETURN_NOT_ALLOWED: 'لا يسمح بإرجاع هذه القطعة',
  WARRANTY_EXPIRED: 'الضمان منتهي',
  WARRANTY_NOT_FOUND: 'الضمان غير موجود',
  
  // Server Errors
  SERVER_ERROR: 'خطأ في الخادم، يرجى المحاولة لاحقاً',
  INTERNAL_SERVER_ERROR: 'خطأ داخلي في الخادم',
  BAD_REQUEST: 'طلب غير صالح',
  NOT_FOUND: 'المورد غير موجود',
  METHOD_NOT_ALLOWED: 'الطريقة غير مسموحة',
  CONFLICT: 'تعارض في البيانات',
  
  // Database Errors
  DATABASE_ERROR: 'خطأ في قاعدة البيانات',
  DUPLICATE_ENTRY: 'بيانات مكررة',
  FOREIGN_KEY_CONSTRAINT: 'تعارض في البيانات المرتبطة',
  
  // File Upload Errors
  FILE_TOO_LARGE: 'الملف كبير جداً',
  INVALID_FILE_TYPE: 'نوع الملف غير مسموح',
  UPLOAD_FAILED: 'فشل رفع الملف',
  
  // General Errors
  UNKNOWN_ERROR: 'حدث خطأ غير معروف',
  OPERATION_FAILED: 'فشلت العملية',
  INVALID_INPUT: 'إدخال غير صالح',
  VALIDATION_ERROR: 'خطأ في التحقق من البيانات',
  
  // Success Messages (for completeness)
  SUCCESS: 'تمت العملية بنجاح',
  CREATED_SUCCESSFULLY: 'تم الإنشاء بنجاح',
  UPDATED_SUCCESSFULLY: 'تم التحديث بنجاح',
  DELETED_SUCCESSFULLY: 'تم الحذف بنجاح',
  SAVED_SUCCESSFULLY: 'تم الحفظ بنجاح',
} as const;

export type ErrorMessageKey = keyof typeof errorMessages;

export function getArabicErrorMessage(error: any): string {
  // If error is already a string in Arabic, return it
  if (typeof error === 'string' && /[\u0600-\u06FF]/.test(error)) {
    return error;
  }
  
  // If error has a message property
  if (error?.message) {
    // Check if it's an error code
    if (errorMessages[error.message as ErrorMessageKey]) {
      return errorMessages[error.message as ErrorMessageKey];
    }
    
    // If message is in Arabic, return it
    if (/[\u0600-\u06FF]/.test(error.message)) {
      return error.message;
    }
  }
  
  // Map HTTP status codes to Arabic messages
  if (error?.status) {
    switch (error.status) {
      case 400:
        return errorMessages.BAD_REQUEST;
      case 401:
        return errorMessages.UNAUTHORIZED;
      case 403:
        return errorMessages.PERMISSION_DENIED;
      case 404:
        return errorMessages.NOT_FOUND;
      case 409:
        return errorMessages.CONFLICT;
      case 500:
        return errorMessages.SERVER_ERROR;
      case 502:
      case 503:
      case 504:
        return errorMessages.SERVER_ERROR;
      default:
        return errorMessages.UNKNOWN_ERROR;
    }
  }
  
  // Default fallback
  return errorMessages.UNKNOWN_ERROR;
}

export function isRetryableError(error: any): boolean {
  if (!error) return false;
  
  const retryableStatuses = [408, 429, 500, 502, 503, 504];
  const retryableTypes = ['NETWORK_ERROR', 'TIMEOUT_ERROR', 'CONNECTION_FAILED'];
  
  if (error?.status && retryableStatuses.includes(error.status)) {
    return true;
  }
  
  if (error?.type && retryableTypes.includes(error.type)) {
    return true;
  }
  
  if (error?.code === 'NETWORK_ERROR' || error?.code === 'TIMEOUT') {
    return true;
  }
  
  return false;
}