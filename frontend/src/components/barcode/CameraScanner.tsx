import { useState, useEffect, useRef } from 'react'
import { clsx } from 'clsx'

interface CameraScannerProps {
  onScan: (barcode: string) => void
  onClose?: () => void
  className?: string
}

export function CameraScanner({ onScan, onClose, className }: CameraScannerProps) {
  const [stream, setStream] = useState<MediaStream | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [scanning, setScanning] = useState(false)
  const videoRef = useRef<HTMLVideoElement>(null)
  const canvasRef = useRef<HTMLCanvasElement>(null)

  useEffect(() => {
    startCamera()
    return () => {
      stopCamera()
    }
  }, [])

  const startCamera = async () => {
    try {
      const mediaStream = await navigator.mediaDevices.getUserMedia({
        video: { facingMode: 'environment' },
      })
      setStream(mediaStream)
      setError(null)
      
      if (videoRef.current) {
        videoRef.current.srcObject = mediaStream
        videoRef.current.play()
      }
      
      setScanning(true)
      scanFrame()
    } catch (err) {
      setError('غير قادر على الوصول للكاميرا')
      console.error('Camera error:', err)
    }
  }

  const stopCamera = () => {
    if (stream) {
      stream.getTracks().forEach(track => track.stop())
      setStream(null)
    }
    setScanning(false)
  }

  const scanFrame = () => {
    if (!scanning || !videoRef.current || !canvasRef.current) return

    const video = videoRef.current
    const canvas = canvasRef.current
    const context = canvas.getContext('2d')

    if (context && video.readyState === video.HAVE_ENOUGH_DATA) {
      canvas.width = video.videoWidth
      canvas.height = video.videoHeight
      context.drawImage(video, 0, 0, canvas.width, canvas.height)

      // TODO: Implement actual barcode detection using a library like zxing or quagga
      // For now, this is a placeholder that simulates scanning
      // In production, integrate with a proper barcode scanning library
      
      // Simulated scan (remove this in production)
      // setTimeout(() => {
      //   onScan('1234567890123')
      //   stopCamera()
      // }, 2000)
    }

    if (scanning) {
      requestAnimationFrame(scanFrame)
    }
  }

  const handleManualInput = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const formData = new FormData(e.currentTarget)
    const barcode = formData.get('barcode') as string
    if (barcode) {
      onScan(barcode)
      stopCamera()
    }
  }

  return (
    <div className={clsx('bg-surface rounded-lg shadow-lg overflow-hidden', className)}>
      {/* Header */}
      <div className="p-4 border-b border-border flex items-center justify-between">
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

      {/* Camera View */}
      <div className="relative">
        {error ? (
          <div className="p-8 text-center">
            <div className="text-4xl mb-4">📷</div>
            <p className="text-danger mb-4">{error}</p>
            <button
              onClick={startCamera}
              className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
            >
              إعادة المحاولة
            </button>
          </div>
        ) : (
          <>
            <video
              ref={videoRef}
              className="w-full h-64 object-cover bg-black"
              playsInline
              muted
            />
            {/* Scanning overlay */}
            <div className="absolute inset-0 pointer-events-none">
              <div className="absolute inset-0 flex items-center justify-center">
                <div className="w-48 h-48 border-2 border-primary rounded-lg relative">
                  <div className="absolute top-0 left-0 w-4 h-4 border-t-4 border-l-4 border-primary" />
                  <div className="absolute top-0 right-0 w-4 h-4 border-t-4 border-r-4 border-primary" />
                  <div className="absolute bottom-0 left-0 w-4 h-4 border-b-4 border-l-4 border-primary" />
                  <div className="absolute bottom-0 right-0 w-4 h-4 border-b-4 border-r-4 border-primary" />
                  <div className="absolute top-1/2 left-0 right-0 h-0.5 bg-primary animate-pulse" />
                </div>
              </div>
            </div>
          </>
        )}
      </div>

      {/* Manual Input */}
      <div className="p-4 border-t border-border">
        <form onSubmit={handleManualInput} className="flex gap-2">
          <input
            type="text"
            name="barcode"
            placeholder="أدخل الباركود يدوياً..."
            className="flex-1 px-3 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          />
          <button
            type="submit"
            className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-600 transition-colors"
          >
            تأكيد
          </button>
        </form>
      </div>

      {/* Instructions */}
      <div className="p-4 bg-muted-10 border-t border-border">
        <p className="text-sm text-muted text-center">
          ضع الباركود داخل الإطار للمسح التلقائي
        </p>
      </div>

      {/* Hidden canvas for frame capture */}
      <canvas ref={canvasRef} className="hidden" />
    </div>
  )
}
