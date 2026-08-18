import { useRef } from 'react'
import { Print } from 'lucide-react'

export interface BarcodeLabelData {
  productName: string
  barcode: string
  price?: number
  sku?: string
  storeName?: string
  quantity?: number
}

interface BarcodeLabelProps {
  data: BarcodeLabelData[]
  onPrint?: () => void
  showPrintButton?: boolean
  labelWidth?: number // in mm
  labelHeight?: number // in mm
}

export function BarcodeLabel({
  data,
  onPrint,
  showPrintButton = true,
  labelWidth = 50,
  labelHeight = 30
}: BarcodeLabelProps) {
  const printRef = useRef<HTMLDivElement>(null)

  const handlePrint = () => {
    if (printRef.current) {
      const printContent = printRef.current.innerHTML
      const originalContent = document.body.innerHTML

      document.body.innerHTML = printContent
      window.print()
      document.body.innerHTML = originalContent

      onPrint?.()
    }
  }

  return (
    <div className="bg-white p-8">
      {showPrintButton && (
        <button
          onClick={handlePrint}
          className="mb-4 flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-90 transition-colors"
        >
          <Print className="w-4 h-4" />
          طباعة الباركود
        </button>
      )}

      <div ref={printRef} className="flex flex-wrap gap-4 justify-start" dir="rtl">
        {data.map((item, index) => (
          <div
            key={index}
            className="border border-gray-300 p-2 text-center"
            style={{
              width: `${labelWidth}mm`,
              height: `${labelHeight}mm`,
              fontSize: '8pt'
            }}
          >
            {/* Store Name */}
            {item.storeName && (
              <div className="font-bold text-xs mb-1 truncate">{item.storeName}</div>
            )}

            {/* Product Name */}
            <div className="font-semibold text-xs mb-1 truncate" title={item.productName}>
              {item.productName}
            </div>

            {/* SKU */}
            {item.sku && (
              <div className="text-xs text-gray-600 mb-1">{item.sku}</div>
            )}

            {/* Price */}
            {item.price && (
              <div className="font-bold text-sm mb-1">{item.price.toFixed(2)}</div>
            )}

            {/* Barcode */}
            <div className="text-xs font-mono font-bold mb-1">{item.barcode}</div>

            {/* Quantity */}
            {item.quantity && item.quantity > 1 && (
              <div className="text-xs text-gray-600">الكمية: {item.quantity}</div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}

// Single barcode label component
export function SingleBarcodeLabel({
  data,
  onPrint,
  showPrintButton = true
}: {
  data: BarcodeLabelData
  onPrint?: () => void
  showPrintButton?: boolean
}) {
  const printRef = useRef<HTMLDivElement>(null)

  const handlePrint = () => {
    if (printRef.current) {
      const printContent = printRef.current.innerHTML
      const originalContent = document.body.innerHTML

      document.body.innerHTML = printContent
      window.print()
      document.body.innerHTML = originalContent

      onPrint?.()
    }
  }

  return (
    <div className="bg-white p-8">
      {showPrintButton && (
        <button
          onClick={handlePrint}
          className="mb-4 flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-90 transition-colors"
        >
          <Print className="w-4 h-4" />
          طباعة الباركود
        </button>
      )}

      <div
        ref={printRef}
        className="border-2 border-black p-4 text-center mx-auto"
        style={{ width: '100mm', height: '60mm' }}
        dir="rtl"
      >
        {/* Store Name */}
        {data.storeName && (
          <div className="font-bold text-lg mb-2">{data.storeName}</div>
        )}

        {/* Product Name */}
        <div className="font-bold text-xl mb-2 truncate" title={data.productName}>
          {data.productName}
        </div>

        {/* SKU */}
        {data.sku && (
          <div className="text-sm text-gray-600 mb-2">SKU: {data.sku}</div>
        )}

        {/* Price */}
        {data.price && (
          <div className="font-bold text-3xl mb-4">{data.price.toFixed(2)}</div>
        )}

        {/* Barcode */}
        <div className="text-2xl font-mono font-bold mb-2">{data.barcode}</div>

        {/* Quantity */}
        {data.quantity && data.quantity > 1 && (
          <div className="text-sm text-gray-600">الكمية: {data.quantity}</div>
        )}
      </div>
    </div>
  )
}