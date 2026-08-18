import { useState, useCallback } from 'react'

export interface UndoAction {
  id: string
  description: string
  undo: () => void | Promise<void>
  timestamp: number
}

interface UndoOptions {
  description: string
  undo: () => void | Promise<void>
  timeout?: number // Auto-undo after this many ms (optional)
}

export function useUndo(maxHistory = 10) {
  const [history, setHistory] = useState<UndoAction[]>([])

  const addUndo = useCallback((options: UndoOptions): UndoAction => {
    const action: UndoAction = {
      id: Math.random().toString(36).substr(2, 9),
      description: options.description,
      undo: options.undo,
      timestamp: Date.now(),
    }

    setHistory(prev => {
      const newHistory = [action, ...prev].slice(0, maxHistory)
      return newHistory
    })

    // Auto-undo if timeout is specified
    if (options.timeout) {
      setTimeout(() => {
        executeUndo(action.id)
      }, options.timeout)
    }

    return action
  }, [maxHistory])

  const executeUndo = useCallback(async (id: string) => {
    const action = history.find(a => a.id === id)
    if (action) {
      await action.undo()
      setHistory(prev => prev.filter(a => a.id !== id))
    }
  }, [history])

  const clearHistory = useCallback(() => {
    setHistory([])
  }, [])

  const removeAction = useCallback((id: string) => {
    setHistory(prev => prev.filter(a => a.id !== id))
  }, [])

  return {
    history,
    addUndo,
    executeUndo,
    clearHistory,
    removeAction,
    canUndo: history.length > 0,
  }
}
