import { useState, useEffect, useRef } from 'react';
import { cn } from '../../utils';
import { Camera, X, RefreshCw, Check } from 'lucide-react';

interface CameraScannerProps {
  onScan: (barcode: string) => void;
  onClose: () => void;
  onError?: (error: string) => void;
}

export function CameraScanner({ onScan, onClose, onError }: CameraScannerProps) {
  const [isScanning, setIsScanning] = useState(false);
  const [scannedBarcodes, setScannedBarcodes] = useState<string[]>([]);
  const [error, setError] = useState<string | null>(null);
  const videoRef = useRef<HTMLVideoElement>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    startCamera();
    return () => {
      stopCamera();
    };
  }, []);

  const startCamera = async () => {
    try {
      setIsScanning(true);
      setError(null);
      
      // Request camera access
      const stream = await navigator.mediaDevices.getUserMedia({
        video: { facingMode: 'environment' }
      });
      
      if (videoRef.current) {
        videoRef.current.srcObject = stream;
        videoRef.current.play();
      }
      
      // Start barcode detection (simplified version)
      startBarcodeDetection();
    } catch (err) {
      const errorMessage = 'تعذر الوصول للكاميرا. يرجى التأكد من الإذن.';
      setError(errorMessage);
      setIsScanning(false);
      onError?.(errorMessage);
    }
  };

  const stopCamera = () => {
    if (videoRef.current?.srcObject) {
      const stream = videoRef.current.srcObject as MediaStream;
      stream.getTracks().forEach(track => track.stop());
      videoRef.current.srcObject = null;
    }
    setIsScanning(false);
  };

  const startBarcodeDetection = () => {
    // Simplified barcode detection logic
    // In production, you would use a library like zxing or quagga
    const detectionInterval = setInterval(() => {
      if (!isScanning || !videoRef.current || !canvasRef.current) {
        clearInterval(detectionInterval);
        return;
      }

      // Capture frame
      const canvas = canvasRef.current;
      const video = videoRef.current;
      canvas.width = video.videoWidth;
      canvas.height = video.videoHeight;
      
      const ctx = canvas.getContext('2d');
      if (ctx) {
        ctx.drawImage(video, 0, 0);
        
        // Here you would implement actual barcode detection
        // For now, this is a placeholder
      }
    }, 100);
  };

  const handleManualScan = (barcode: string) => {
    if (barcode && !scannedBarcodes.includes(barcode)) {
      setScannedBarcodes([...scannedBarcodes, barcode]);
      onScan(barcode);
    }
  };

  const handleRetry = () => {
    setError(null);
    startCamera();
  };

  return (
    <div className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4">
      <div className="bg-surface border border-border-default rounded-2xl w-full max-w-2xl overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-border-default">
          <div className="flex items-center gap-2">
            <Camera className="w-5 h-5 text-cyan-400" />
            <h3 className="text-text font-semibold">مسح الباركود بالكاميرا</h3>
          </div>
          <button
            onClick={onClose}
            className="p-2 rounded-lg hover:bg-surface/50 transition-colors"
          >
            <X className="w-5 h-5 text-text-muted" />
          </button>
        </div>

        {/* Camera View */}
        <div className="relative aspect-video bg-black">
          <video
            ref={videoRef}
            className="w-full h-full object-cover"
            autoPlay
            playsInline
            muted
          />
          <canvas ref={canvasRef} className="hidden" />
          
          {/* Scanner Overlay */}
          {isScanning && (
            <div className="absolute inset-0 flex items-center justify-center">
              <div className="w-64 h-32 border-2 border-cyan-400 rounded-lg relative">
                <div className="absolute top-0 left-0 w-4 h-4 border-t-4 border-l-4 border-cyan-400 -mt-1 -ml-1" />
                <div className="absolute top-0 right-0 w-4 h-4 border-t-4 border-r-4 border-cyan-400 -mt-1 -mr-1" />
                <div className="absolute bottom-0 left-0 w-4 h-4 border-b-4 border-l-4 border-cyan-400 -mb-1 -ml-1" />
                <div className="absolute bottom-0 right-0 w-4 h-4 border-b-4 border-r-4 border-cyan-400 -mb-1 -mr-1" />
                
                {/* Scanning Line */}
                <div className="absolute top-0 left-0 right-0 h-0.5 bg-cyan-400 animate-pulse" />
              </div>
            </div>
          )}

          {/* Error State */}
          {error && (
            <div className="absolute inset-0 flex items-center justify-center bg-black/90">
              <div className="text-center p-6">
                <div className="w-16 h-16 rounded-full bg-red-500/20 flex items-center justify-center mx-auto mb-4">
                  <X className="w-8 h-8 text-red-400" />
                </div>
                <p className="text-text mb-4">{error}</p>
                <button
                  onClick={handleRetry}
                  className="flex items-center gap-2 px-4 py-2 bg-cyan-500/20 text-cyan-400 rounded-lg hover:bg-cyan-500/30 transition-colors"
                >
                  <RefreshCw className="w-4 h-4" />
                  إعادة المحاولة
                </button>
              </div>
            </div>
          )}
        </div>

        {/* Scanned Items */}
        {scannedBarcodes.length > 0 && (
          <div className="p-4 border-t border-border-default">
            <div className="flex items-center gap-2 mb-3">
              <Check className="w-4 h-4 text-green-400" />
              <span className="text-sm font-medium text-text">
                الباركودات الممسوحة ({scannedBarcodes.length})
              </span>
            </div>
            <div className="flex flex-wrap gap-2">
              {scannedBarcodes.map((barcode, index) => (
                <div
                  key={index}
                  className="px-3 py-1.5 bg-green-500/10 border border-green-500/30 rounded-lg text-green-400 text-sm"
                >
                  {barcode}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Manual Input Fallback */}
        <div className="p-4 border-t border-border-default">
          <div className="flex gap-2">
            <input
              type="text"
              placeholder="أو أدخل الباركود يدوياً..."
              className="flex-1 px-4 py-2 bg-surface/50 border border-border-default rounded-lg text-text placeholder:text-text-muted/50 focus:outline-none focus:border-cyan-500/50"
              onKeyPress={(e) => {
                if (e.key === 'Enter') {
                  handleManualScan(e.currentTarget.value);
                  e.currentTarget.value = '';
                }
              }}
            />
            <button
              onClick={() => {
                const input = document.querySelector('input') as HTMLInputElement;
                if (input?.value) {
                  handleManualScan(input.value);
                  input.value = '';
                }
              }}
              className="px-4 py-2 bg-cyan-500/20 text-cyan-400 rounded-lg hover:bg-cyan-500/30 transition-colors"
            >
              مسح
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}