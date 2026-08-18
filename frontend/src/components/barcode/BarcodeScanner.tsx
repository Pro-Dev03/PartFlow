import { useState, useEffect } from 'react'
import { CameraScanner } from './CameraScanner'

interface BarcodeScannerProps {
  onScan: (barcode: string) => void
  onClose?: () => void
  mode?: 'camera' | 'manual' | 'auto'
  className?: string
}

export function BarcodeScanner({ onScan, onClose, mode = 'auto', className }: BarcodeScannerProps) {
  const [scannerMode, setScannerMode] = useState<'camera' | 'manual'>(mode === 'manual' ? 'manual' : 'camera')
  const [manualBarcode, setManualBarcode] = useState('')

  useEffect(() => {
    // Listen for external barcode scanner input (keyboard emulation)
    const handleKeyPress = (e: KeyboardEvent) => {
      if (scannerMode === 'auto' || scannerMode === 'manual') {
        // Barcode scanners typically send ENTER after the barcode
        if (e.key === 'Enter' && manualBarcode.length > 0) {
          onScan(manualBarcode)
          setManualBarcode('')
        } else if (e.key.length === 1) {
          // Build barcode from keypresses
          setManualBarcode(prev => prev + e.key)
        }
      }
    }

    window.addEventListener('keypress', handleKeyPress)
    return () => window.removeEventListener('keypress', handleKeyPress)
  }, [manualBarcode, scannerMode, onScan])

  const handleManualSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (manualBarcode.trim()) {
      onScan(manualBarcode.trim())
      setManualBarcode('')
      if (onClose) onClose()
    }
  }

  if (scannerMode === 'camera') {
    return (
      <CameraScanner
        onScan={(barcode) => {
          onScan(barcode)
          if (onClose) onClose()
        }}
        onClose={() => {
          setScannerMode('manual')
          if (onClose) onClose()
        }}
        className={className}
      />
    )
  }

  return (
    <div className={clsx('bg-surface rounded-lg shadow-lg p-6', className)}>
      {/* Header */}
      <div className="mb-4 flex items-center justify-between">
        <h3 className="font-semibold text-text">مسح الباركود</h3>
        {onClose && (
          <button
            onClick={onClose}
            className="text-muted hover:text-text p-2 rounded-lg hover:bg-muted-10"
          >
            ✕
          </button>
        )}
      </div>

      {/* Mode Toggle */}
      <div className="flex gap-2 mb-4">
        <button
          onClick={() => setScannerMode('camera')}
          className={clsx(
            'flex-1 px-4 py-2 rounded-lg text-sm font-medium transition-colors flex items-center justify-center gap-2',
            scannerMode === 'camera' ? 'bg-primary text-white' : 'bg-muted text-muted hover:bg-muted-80'
          )}
        >
          📷 الكاميرا
        </button>
        <button
          onClick={() => setScannerMode('manual')}
          className={clsx(
            'flex-1 px-4 py-2 rounded-lg text-sm font-medium transition-colors flex items-center justify-center gap-2',
            scannerMode === 'manual' ? 'bg-primary text-white' : 'bg-muted text-muted hover:bg-muted-80'
          )}
        >
          ⌨️ يدوي
        </button>
      </div>

      {/* Manual Input */}
      <form onSubmit={handleManualSubmit} className="space-y-4">
        <div>
          <input
            type="text"
            value={manualBarcode}
            onChange={(e) => setManualBarcode(e.target.value)}
            placeholder="أدخل الباركود أو استخدم قارئ خارجي..."
            className="w-full px-4 py-3 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent text-lg"
            autoFocus
          />
        </div>
        <button
          type="submit"
          disabled={!manualBarcode.trim()}
          className="w-full px-4 py-3 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed font-medium"
        >
          تأكيد
        </button>
      </form>

      {/* Instructions */}
      <div className="mt-4 p-4 bg-muted-10 rounded-lg">
        <p className="text-sm text-muted text-center">
          يمكنك استخدام قارئ الباركود الخارجي أو إدخال الرقم يدوياً
        </p>
      </div>
    </div>
  )
}
