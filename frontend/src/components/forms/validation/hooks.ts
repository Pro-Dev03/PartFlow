import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

export function useValidatedForm<T extends z.ZodType>(
  schema: T,
  defaultValues?: Partial<z.infer<T>>
) {
  return useForm<z.infer<T>>({
    resolver: zodResolver(schema),
    defaultValues,
    mode: 'onBlur',
  })
}

export function useFormFieldError(form: any, name: string) {
  const error = form.formState.errors[name]
  return error ? String(error.message) : undefined
}

// Common validation patterns
export const createValidationSchema = (fields: Record<string, z.ZodTypeAny>) => {
  return z.object(fields)
}

export const validateField = (schema: z.ZodTypeAny, value: any): string | null => {
  try {
    schema.parse(value)
    return null
  } catch (error) {
    if (error instanceof z.ZodError) {
      return error.errors[0]?.message || 'Invalid value'
    }
    return 'Validation error'
  }
}

export const validateForm = <T extends Record<string, any>>(
  schema: z.ZodSchema<T>,
  data: T
): { valid: boolean; errors: Record<string, string> } => {
  try {
    schema.parse(data)
    return { valid: true, errors: {} }
  } catch (error) {
    if (error instanceof z.ZodError) {
      const errors: Record<string, string> = {}
      error.errors.forEach((err) => {
        if (err.path.length > 0) {
          errors[err.path[0] as string] = err.message
        }
      })
      return { valid: false, errors }
    }
    return { valid: false, errors: {} }
  }
}

// Re-export from schemas.ts for convenience
export * from './schemas'
