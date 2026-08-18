import { useForm, UseFormReturn } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useState, useCallback } from 'react'

interface UseFormValidationProps<T extends z.ZodSchema> {
  schema: T
  defaultValues?: z.infer<T>
  onSubmit: (data: z.infer<T>) => Promise<void> | void
  mode?: 'onSubmit' | 'onBlur' | 'onChange' | 'onTouched' | 'all'
}

interface UseFormValidationReturn<T extends z.ZodSchema> extends UseFormReturn<z.infer<T>> {
  isSubmitting: boolean
  submitError: string | null
  handleSubmit: (e?: React.BaseSyntheticEvent) => Promise<void>
  resetForm: () => void
}

export function useFormValidation<T extends z.ZodSchema>({
  schema,
  defaultValues,
  onSubmit,
  mode = 'onSubmit',
}: UseFormValidationProps<T>): UseFormValidationReturn<T> {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [submitError, setSubmitError] = useState<string | null>(null)

  const methods = useForm<z.infer<T>>({
    resolver: zodResolver(schema),
    defaultValues,
    mode,
  })

  const handleSubmit = useCallback(
    async (e?: React.BaseSyntheticEvent) => {
      e?.preventDefault()
      setSubmitError(null)

      const isValid = await methods.trigger()
      if (!isValid) {
        return
      }

      setIsSubmitting(true)
      try {
        const data = methods.getValues()
        await onSubmit(data)
        methods.reset(defaultValues)
      } catch (error: any) {
        setSubmitError(error.message || 'حدث خطأ أثناء إرسال النموذج')
        console.error('Form submission error:', error)
      } finally {
        setIsSubmitting(false)
      }
    },
    [methods, onSubmit, defaultValues]
  )

  const resetForm = useCallback(() => {
    methods.reset(defaultValues)
    setSubmitError(null)
  }, [methods, defaultValues])

  return {
    ...methods,
    isSubmitting,
    submitError,
    handleSubmit,
    resetForm,
  }
}

// Custom hook for real-time validation
export function useRealTimeValidation<T extends z.ZodSchema>(
  schema: T,
  data: any
) {
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [isValid, setIsValid] = useState(false)

  const validate = useCallback(() => {
    try {
      schema.parse(data)
      setErrors({})
      setIsValid(true)
      return true
    } catch (error) {
      if (error instanceof z.ZodError) {
        const fieldErrors: Record<string, string> = {}
        error.errors.forEach((err) => {
          if (err.path.length > 0) {
            fieldErrors[err.path.join('.')] = err.message
          }
        })
        setErrors(fieldErrors)
        setIsValid(false)
      }
      return false
    }
  }, [schema, data])

  return { errors, isValid, validate }
}

// Hook for async validation (e.g., checking uniqueness)
export function useAsyncValidation<T>(
  validator: (value: T) => Promise<boolean>,
  debounceMs: number = 500
) {
  const [isValidating, setIsValidating] = useState(false)
  const [isValid, setIsValid] = useState<boolean | null>(null)
  const [error, setError] = useState<string | null>(null)

  const validate = useCallback(
    debounce(async (value: T) => {
      setIsValidating(true)
      setError(null)
      
      try {
        const result = await validator(value)
        setIsValid(result)
        if (!result) {
          setError('القيمة موجودة بالفعل')
        }
      } catch (err) {
        setError('حدث خطأ أثناء التحقق')
        setIsValid(false)
      } finally {
        setIsValidating(false)
      }
    }, debounceMs),
    [validator, debounceMs]
  )

  return { isValidating, isValid, error, validate }
}

// Debounce utility
function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: NodeJS.Timeout | null = null
  return (...args: Parameters<T>) => {
    if (timeout) clearTimeout(timeout)
    timeout = setTimeout(() => func(...args), wait)
  }
}