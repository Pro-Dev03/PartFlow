import { useRef, useEffect } from 'react'
import { Print } from 'lucide-react'

export interface InvoiceItem {
  name: string
  barcode?: string
  quantity: number
  price: number
  total: number
}

export interface InvoiceData {
  invoiceNumber: string
  date: string
  customer: {
    name: string
    phone?: string
    address?: string
  }
  items: InvoiceItem[]
  subtotal: number
  tax?: number
  discount?: number
  total: number
  paid: number
  remaining: number
  paymentMethod: string
  store: {
    name: string
    address: string
    phone: string
    logo?: string
  }
  notes?: string
}

interface InvoiceTemplateProps {
  data: InvoiceData
  onPrint?: () => void
  showPrintButton?: boolean
}

export function InvoiceTemplate({ data, onPrint, showPrintButton = true }: InvoiceTemplateProps) {
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
          طباعة الفاتورة
        </button>
      )}

      <div ref={printRef} className="max-w-2xl mx-auto bg-white p-8" dir="rtl">
        {/* Header */}
        <div className="flex justify-between items-start mb-8 border-b pb-4">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">{data.store.name}</h1>
            <p className="text-sm text-gray-600 mt-1">{data.store.address}</p>
            <p className="text-sm text-gray-600">{data.store.phone}</p>
          </div>
          <div className="text-left">
            <h2 className="text-xl font-bold text-gray-900">فاتورة</h2>
            <p className="text-sm text-gray-600">رقم: {data.invoiceNumber}</p>
            <p className="text-sm text-gray-600">التاريخ: {data.date}</p>
          </div>
        </div>

        {/* Customer Info */}
        <div className="mb-6 p-4 bg-gray-50 rounded-lg">
          <h3 className="font-semibold text-gray-900 mb-2">معلومات العميل</h3>
          <p className="text-sm text-gray-700">
            <span className="font-medium">الاسم:</span> {data.customer.name}
          </p>
          {data.customer.phone && (
            <p className="text-sm text-gray-700">
              <span className="font-medium">الهاتف:</span> {data.customer.phone}
            </p>
          )}
          {data.customer.address && (
            <p className="text-sm text-gray-700">
              <span className="font-medium">العنوان:</span> {data.customer.address}
            </p>
          )}
        </div>

        {/* Items Table */}
        <table className="w-full mb-6">
          <thead>
            <tr className="border-b-2 border-gray-300">
              <th className="text-right py-2 text-sm font-semibold text-gray-900">المنتج</th>
              <th className="text-center py-2 text-sm font-semibold text-gray-900">الكمية</th>
              <th className="text-center py-2 text-sm font-semibold text-gray-900">السعر</th>
              <th className="text-left py-2 text-sm font-semibold text-gray-900">الإجمالي</th>
            </tr>
          </thead>
          <tbody>
            {data.items.map((item, index) => (
              <tr key={index} className="border-b border-gray-200">
                <td className="py-3 text-sm text-gray-700">
                  {item.name}
                  {item.barcode && (
                    <span className="block text-xs text-gray-500">{item.barcode}</span>
                  )}
                </td>
                <td className="py-3 text-sm text-gray-700 text-center">{item.quantity}</td>
                <td className="py-3 text-sm text-gray-700 text-center">{item.price.toFixed(2)}</td>
                <td className="py-3 text-sm text-gray-700 text-left">{item.total.toFixed(2)}</td>
              </tr>
            ))}
          </tbody>
        </table>

        {/* Totals */}
        <div className="space-y-2 mb-6">
          <div className="flex justify-between text-sm">
            <span className="text-gray-600">المجموع الفرعي:</span>
            <span className="font-medium">{data.subtotal.toFixed(2)}</span>
          </div>
          {data.tax && (
            <div className="flex justify-between text-sm">
              <span className="text-gray-600">الضريبة:</span>
              <span className="font-medium">{data.tax.toFixed(2)}</span>
            </div>
          )}
          {data.discount && (
            <div className="flex justify-between text-sm">
              <span className="text-gray-600">الخصم:</span>
              <span className="font-medium text-red-600">-{data.discount.toFixed(2)}</span>
            </div>
          )}
          <div className="flex justify-between text-lg font-bold border-t pt-2">
            <span className="text-gray-900">الإجمالي:</span>
            <span className="text-gray-900">{data.total.toFixed(2)}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-gray-600">المدفوع:</span>
            <span className="font-medium">{data.paid.toFixed(2)}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-gray-600">المتبقي:</span>
            <span className={`font-medium ${data.remaining > 0 ? 'text-red-600' : 'text-green-600'}`}>
              {data.remaining.toFixed(2)}
            </span>
          </div>
        </div>

        {/* Payment Method */}
        <div className="mb-6 p-4 bg-gray-50 rounded-lg">
          <p className="text-sm text-gray-700">
            <span className="font-medium">طريقة الدفع:</span> {data.paymentMethod}
          </p>
        </div>

        {/* Notes */}
        {data.notes && (
          <div className="mb-6 p-4 bg-gray-50 rounded-lg">
            <h3 className="font-semibold text-gray-900 mb-2">ملاحظات</h3>
            <p className="text-sm text-gray-700">{data.notes}</p>
          </div>
        )}

        {/* Footer */}
        <div className="text-center text-sm text-gray-600 pt-4 border-t">
          <p>شكراً لتعاملكم معنا</p>
          <p className="mt-1">{data.store.name} - {data.store.phone}</p>
        </div>
      </div>
    </div>
  )
}