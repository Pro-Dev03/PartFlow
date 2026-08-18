import { z } from 'zod'

// Common validation schemas
export const emailSchema = z
  .string()
  .min(1, 'البريد الإلكتروني مطلوب')
  .email('البريد الإلكتروني غير صالح')

export const phoneSchema = z
  .string()
  .min(1, 'رقم الهاتف مطلوب')
  .regex(/^05\d{8}$/, 'رقم الهاتف يجب أن يبدأ بـ 05 ويحتوي 10 أرقام')

export const passwordSchema = z
  .string()
  .min(8, 'كلمة المرور يجب أن تكون 8 أحرف على الأقل')
  .regex(/[A-Z]/, 'كلمة المرور يجب أن تحتوي على حرف كبير')
  .regex(/[a-z]/, 'كلمة المرور يجب أن تحتوي على حرف صغير')
  .regex(/[0-9]/, 'كلمة المرور يجب أن تحتوي على رقم')

export const nameSchema = z
  .string()
  .min(2, 'الاسم يجب أن يكون حرفين على الأقل')
  .max(100, 'الاسم طويل جداً')

export const amountSchema = z
  .number()
  .min(0, 'المبلغ يجب أن يكون رقمًا موجبًا')
  .max(9999999.99, 'المبلغ كبير جداً')

export const quantitySchema = z
  .number()
  .int('الكمية يجب أن تكون رقمًا صحيحًا')
  .min(0, 'الكمية يجب أن تكون رقمًا موجبًا')
  .max(999999, 'الكمية كبيرة جداً')

export const barcodeSchema = z
  .string()
  .min(1, 'الباركود مطلوب')
  .max(50, 'الباركود طويل جداً')
  .regex(/^[A-Za-z0-9-]+$/, 'الباركود يجب أن يحتوي على أحرف وأرقام فقط')

export const dateSchema = z
  .string()
  .min(1, 'التاريخ مطلوب')
  .regex(/^\d{4}-\d{2}-\d{2}$/, 'التاريخ يجب أن يكون بصيغة YYYY-MM-DD')

// Product validation schema
export const productSchema = z.object({
  name: nameSchema,
  barcode: barcodeSchema.optional(),
  category: z.string().min(1, 'الفئة مطلوبة'),
  brand: z.string().min(1, 'العلامة التجارية مطلوبة'),
  model: z.string().optional(),
  price: amountSchema,
  cost: amountSchema,
  stock: quantitySchema,
  condition: z.enum(['new', 'used']),
  grade: z.enum(['A', 'B', 'C']).optional(),
  location: z.string().min(1, 'الموقع مطلوب'),
  warrantyDays: z.number().int().min(0).optional(),
  description: z.string().max(500, 'الوصف طويل جداً').optional(),
})

// Customer validation schema
export const customerSchema = z.object({
  name: nameSchema,
  phone: phoneSchema,
  email: emailSchema.optional().or(z.literal('')),
  address: z.string().max(200, 'العنوان طويل جداً').optional(),
  notes: z.string().max(500, 'الملاحظات طويلة جداً').optional(),
})

// Supplier validation schema
export const supplierSchema = z.object({
  name: nameSchema,
  phone: phoneSchema,
  email: emailSchema.optional().or(z.literal('')),
  address: z.string().max(200, 'العنوان طويل جداً').optional(),
  city: z.string().max(100, 'المدينة طويلة جداً').optional(),
  notes: z.string().max(500, 'الملاحظات طويلة جداً').optional(),
})

// Sale validation schema
export const saleItemSchema = z.object({
  productId: z.string().min(1, 'المنتج مطلوب'),
  quantity: quantitySchema,
  price: amountSchema,
})

export const saleSchema = z.object({
  customerId: z.string().min(1, 'العميل مطلوب'),
  items: z.array(saleItemSchema).min(1, 'يجب إضافة منتج واحد على الأقل'),
  paymentMethod: z.enum(['cash', 'card', 'transfer', 'debt']),
  notes: z.string().max(500, 'الملاحظات طويلة جداً').optional(),
})

// Purchase validation schema
export const purchaseItemSchema = z.object({
  productId: z.string().min(1, 'المنتج مطلوب'),
  quantity: quantitySchema,
  cost: amountSchema,
})

export const purchaseSchema = z.object({
  supplierId: z.string().min(1, 'المورد مطلوب'),
  items: z.array(purchaseItemSchema).min(1, 'يجب إضافة منتج واحد على الأقل'),
  notes: z.string().max(500, 'الملاحظات طويلة جداً').optional(),
})

// Expense validation schema
export const expenseSchema = z.object({
  categoryId: z.string().min(1, 'الفئة مطلوبة'),
  amount: amountSchema,
  description: z.string().min(1, 'الوصف مطلوب').max(200, 'الوصف طويل جداً'),
  date: dateSchema,
  paymentMethod: z.string().min(1, 'طريقة الدفع مطلوبة'),
  receipt: z.string().optional(),
  notes: z.string().max(500, 'الملاحظات طويلة جداً').optional(),
})

// User validation schema
export const userSchema = z.object({
  name: nameSchema,
  email: emailSchema,
  phone: phoneSchema.optional(),
  role: z.enum(['owner', 'manager', 'employee', 'accountant']),
  password: passwordSchema.optional(),
})

// Organization settings validation schema
export const organizationSettingsSchema = z.object({
  name: nameSchema,
  type: z.enum(['computer_store', 'electronics', 'repair', 'trading']),
  currency: z.enum(['ILS', 'USD', 'EUR', 'GBP']),
  language: z.enum(['ar', 'he', 'en']),
  timezone: z.string().min(1, 'التوقيت مطلوب'),
  address: z.string().max(200, 'العنوان طويل جداً').optional(),
  phone: z.string().optional(),
  email: emailSchema.optional().or(z.literal('')),
  taxId: z.string().max(50, 'الرقم الضريبي طويل جداً').optional(),
})

// Export inference types
export type ProductFormData = z.infer<typeof productSchema>
export type CustomerFormData = z.infer<typeof customerSchema>
export type SupplierFormData = z.infer<typeof supplierSchema>
export type SaleFormData = z.infer<typeof saleSchema>
export type PurchaseFormData = z.infer<typeof purchaseSchema>
export type ExpenseFormData = z.infer<typeof expenseSchema>
export type UserFormData = z.infer<typeof userSchema>
export type OrganizationSettingsFormData = z.infer<typeof organizationSettingsSchema>
