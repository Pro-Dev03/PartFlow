import { describe, it, expect } from 'vitest'
import { renderHook } from '@testing-library/react'
import { useTranslation } from '../../hooks/useTranslation'

describe('useTranslation Hook', () => {
  it('should return translation function', () => {
    const { result } = renderHook(() => useTranslation())
    
    expect(result.current).toBeDefined()
    expect(typeof result.current.t).toBe('function')
  })

  it('should return current language', () => {
    const { result } = renderHook(() => useTranslation())
    
    expect(result.current).toBeDefined()
    expect(typeof result.current.currentLanguage).toBe('string')
  })

  it('should return change language function', () => {
    const { result } = renderHook(() => useTranslation())
    
    expect(result.current).toBeDefined()
    expect(typeof result.current.changeLanguage).toBe('function')
  })
})