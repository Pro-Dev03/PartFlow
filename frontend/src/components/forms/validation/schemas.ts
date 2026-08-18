import { z } from 'zod'

// Common validation schemas
export const commonSchemas = {
  email: z.string().email('invalidEmail'),
  phone: z.string().min(10, 'minLength').max(15, 'maxLength'),
  url: z.string().url('invalidUrl'),
  date: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, 'invalidDate'),
  positiveNumber: z.number().positive('positiveNumber'),
  nonNegativeNumber: z.number().min(0, 'minValue'),
}

// Product validation schema
export const productSchema = z.object({
  name: z.string().min(2, 'minLength').max(100, 'maxLength'),
  barcode: z.string().min(3, 'minLength').max(50, 'maxLength'),
  sku: z.string().optional(),
  category: z.string().min(1, 'required'),
  manufacturer: z.string().min(1, 'required'),
  model: z.string().optional(),
  condition: z.enum(['new', 'used']),
  cost: commonSchemas.positiveNumber,
  price: commonSchemas.positiveNumber,
  stock: commonSchemas.nonNegativeNumber,
  minStock: commonSchemas.nonNegativeNumber.optional(),
  location: z.string().optional(),
  serialNumber: z.string().optional(),
  warranty: z.object({
    enabled: z.boolean(),
    duration: z.number().optional(),
    type: z.string().optional(),
  }).optional(),
  description: z.string().max(500, 'maxLength').optional(),
  images: z.array(z.string()).optional(),
  supplierId: z.string().optional(),
})

// Customer validation schema
export const customerSchema = z.object({
  name: z.string().min(2, 'minLength').max(100, 'maxLength'),
  phone: commonSchemas.phone,
  email: commonSchemas.email.optional(),
  address: z.string().max(200, 'maxLength').optional(),
  city: z.string().optional(),
  notes: z.string().max(500, 'maxLength').optional(),
})

// Sale validation schema
export const saleSchema = z.object({
  customerId: z.string().optional(),
  items: z.array(z.object({
    productId: z.string(),
    quantity: commonSchemas.positiveNumber,
    price: commonSchemas.positiveNumber,
  })).min(1, 'atLeastOneItem'),
  paymentMethod: z.enum(['cash', 'card', 'transfer', 'debt']),
  paidAmount: commonSchemas.nonNegativeNumber,
  discount: commonSchemas.nonNegativeNumber.optional(),
  tax: commonSchemas.nonNegativeNumber.optional(),
  notes: z.string().max(500, 'maxLength').optional(),
})

// Payment validation schema
export const paymentSchema = z.object({
  customerId: z.string(),
  amount: commonSchemas.positiveNumber,
  method: z.enum(['cash', 'card', 'transfer']),
  reference: z.string().optional(),
  notes: z.string().max(200, 'maxLength').optional(),
})

// Purchase validation schema
export const purchaseSchema = z.object({
  supplierId: z.string(),
  items: z.array(z.object({
    productId: z.string(),
    quantity: commonSchemas.positiveNumber,
    cost: commonSchemas.positiveNumber,
  })).min(1, 'atLeastOneItem'),
  totalCost: commonSchemas.positiveNumber,
  paymentStatus: z.enum(['paid', 'partial', 'pending']),
  paidAmount: commonSchemas.nonNegativeNumber,
  notes: z.string().max(500, 'maxLength').optional(),
})

// Expense validation schema
export const expenseSchema = z.object({
  categoryId: z.string(),
  amount: commonSchemas.positiveNumber,
  date: commonSchemas.date,
  description: z.string().min(1, 'required').max(200, 'maxLength'),
  receipt: z.string().optional(),
  notes: z.string().max(500, 'maxLength').optional(),
})

// User validation schema
export const userSchema = z.object({
  name: z.string().min(2, 'minLength').max(100, 'maxLength'),
  email: commonSchemas.email,
  phone: commonSchemas.phone.optional(),
  role: z.enum(['owner', 'manager', 'employee', 'accountant']),
  permissions: z.array(z.string()).optional(),
  active: z.boolean(),
})

// Inspection validation schema
export const inspectionSchema = z.object({
  productId: z.string(),
  inspector: z.string(),
  date: commonSchemas.date,
  powerTest: z.enum(['passed', 'failed', 'skipped']),
  temperatureTest: z.enum(['passed', 'failed', 'skipped']),
  performanceTest: z.enum(['passed', 'failed', 'skipped']),
  portsTest: z.enum(['passed', 'failed', 'skipped']),
  visualInspection: z.enum(['passed', 'failed', 'skipped']),
  serialVerification: z.enum(['passed', 'failed', 'skipped']),
  overallResult: z.enum(['passed', 'failed']),
  notes: z.string().max(500, 'maxLength').optional(),
  photos: z.array(z.string()).optional(),
})