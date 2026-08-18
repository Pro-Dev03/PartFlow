import { useEffect, useState } from 'react'

/**
 * Hook to detect if the user is navigating with keyboard
 * This helps show focus rings only for keyboard navigation
 */
export function useFocusVisible() {
  const [isFocused, setIsFocused] = useState(false)

  useEffect(() => {
    let isKeyboardNavigation = false

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Tab' || e.key === 'Shift') {
        isKeyboardNavigation = true
      }
    }

    const handleMouseDown = () => {
      isKeyboardNavigation = false
    }

    const handleFocus = () => {
      if (isKeyboardNavigation) {
        setIsFocused(true)
      }
    }

    const handleBlur = () => {
      setIsFocused(false)
    }

    window.addEventListener('keydown', handleKeyDown)
    window.addEventListener('mousedown', handleMouseDown)
    
    return () => {
      window.removeEventListener('keydown', handleKeyDown)
      window.removeEventListener('mousedown', handleMouseDown)
    }
  }, [])

  return isFocused
}
