import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useToast, useToastStore } from '../../hooks/useToast'

describe('useToast Hook', () => {
  beforeEach(() => {
    // Clear toasts before each test
    useToastStore.getState().clearToasts()
  })

  it('should show success toast', () => {
    const { result } = renderHook(() => useToast())
    
    act(() => {
      result.current.success('Success message')
    })
    
    const toasts = useToastStore.getState().toasts
    expect(toasts).toHaveLength(1)
    expect(toasts[0].type).toBe('success')
    expect(toasts[0].message).toBe('Success message')
  })

  it('should show error toast', () => {
    const { result } = renderHook(() => useToast())
    
    act(() => {
      result.current.error('Error message')
    })
    
    const toasts = useToastStore.getState().toasts
    expect(toasts).toHaveLength(1)
    expect(toasts[0].type).toBe('error')
    expect(toasts[0].message).toBe('Error message')
  })

  it('should show warning toast', () => {
    const { result } = renderHook(() => useToast())
    
    act(() => {
      result.current.warning('Warning message')
    })
    
    const toasts = useToastStore.getState().toasts
    expect(toasts).toHaveLength(1)
    expect(toasts[0].type).toBe('warning')
    expect(toasts[0].message).toBe('Warning message')
  })

  it('should show info toast', () => {
    const { result } = renderHook(() => useToast())
    
    act(() => {
      result.current.info('Info message')
    })
    
    const toasts = useToastStore.getState().toasts
    expect(toasts).toHaveLength(1)
    expect(toasts[0].type).toBe('info')
    expect(toasts[0].message).toBe('Info message')
  })

  it('should remove toast', () => {
    const { result } = renderHook(() => useToast())
    
    act(() => {
      result.current.success('Test message')
    })
    
    const toasts = useToastStore.getState().toasts
    const toastId = toasts[0].id
    
    act(() => {
      result.current.remove(toastId)
    })
    
    const updatedToasts = useToastStore.getState().toasts
    expect(updatedToasts).toHaveLength(0)
  })

  it('should clear all toasts', () => {
    const { result } = renderHook(() => useToast())
    
    act(() => {
      result.current.success('Message 1')
      result.current.error('Message 2')
      result.current.warning('Message 3')
    })
    
    act(() => {
      result.current.clear()
    })
    
    const toasts = useToastStore.getState().toasts
    expect(toasts).toHaveLength(0)
  })
})