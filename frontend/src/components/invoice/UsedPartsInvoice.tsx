import { useRef } from 'react';
import { Printer, Download, Package, Layers, Clock, CheckCircle, XCircle } from 'lucide-react';
import { Button } from '../ui/button';

interface UsedPartsInvoiceProps {
  saleData: {
    id: string;
    customerName: string;
    customerPhone?: string;
    saleDate: string;
    items: Array<{
      name: string;
      partType?: string;
      partTypeColor?: string;
      condition: string;
      grade?: string;
      sellingPrice: number;
      quantity: number;
      total: number;
      specifications?: Record<string, any>;
      warranty?: string;
    }>;
    subtotal: number;
    total: number;
    paidAmount: number;
    remaining: number;
    paymentMethod: string;
    notes?: string;
  };
  storeInfo?: {
    name: string;
    address: string;
    phone: string;
    email?: string;
    taxNumber?: string;
  };
  onPrint?: () => void;
  onDownload?: () => void;
  onClose?: () => void;
}

export function UsedPartsInvoice({ saleData, storeInfo, onPrint, onDownload, onClose }: UsedPartsInvoiceProps) {
  const invoiceRef = useRef<HTMLDivElement>(null);

  const defaultStoreInfo = {
    name: 'متجر القطع المستعملة',
    address: 'عنوان المتجر',
    phone: 'رقم الهاتف',
    email: 'store@example.com',
    taxNumber: 'رقم الضريبة',
  };

  const store = storeInfo || defaultStoreInfo;

  const handlePrint = () => {
    if (onPrint) {
      onPrint();
    } else {
      window.print();
    }
  };

  const getConditionIcon = (condition: string) => {
    switch (condition) {
      case 'EXCELLENT':
      case 'USED':
        return <CheckCircle className="w-4 h-4 text-green" />;
      case 'GOOD':
        return <CheckCircle className="w-4 h-4 text-green" />;
      case 'FAIR':
        return <Clock className="w-4 h-4 text-yellow" />;
      case 'POOR':
        return <XCircle className="w-4 h-4 text-red" />;
      default:
        return <Package className="w-4 h-4 text-gray" />;
    }
  };

  const getConditionText = (condition: string) => {
    const conditionMap: Record<string, string> = {
      'EXCELLENT': 'ممتاز',
      'VERY_GOOD': 'جيد جداً',
      'GOOD': 'جيد',
      'FAIR': 'مقبول',
      'POOR': 'ضعيف',
      'USED': 'مستعمل',
      'NEW': 'جديد',
      'REFURBISHED': 'مجدد',
    };
    return conditionMap[condition] || condition;
  };

  const getPaymentMethodText = (method: string) => {
    const methodMap: Record<string, string> = {
      'cash': 'نقداً',
      'card': 'بطاقة',
      'bank_transfer': 'تحويل بنكي',
      'credit': 'دين',
    };
    return methodMap[method] || method;
  };

  return (
    <div className="space-y-4">
      {/* Action Buttons */}
      <div className="flex gap-2 print:hidden">
        <Button
          onClick={handlePrint}
          variant="primary"
        >
          <Printer className="w-4 h-4 mr-2" />
          طباعة الفاتورة
        </Button>
        {onDownload && (
          <Button
            onClick={onDownload}
            variant="secondary"
          >
            <Download className="w-4 h-4 mr-2" />
            تحميل PDF
          </Button>
        )}
        {onClose && (
          <Button
            onClick={onClose}
            variant="default"
          >
            إغلاق
          </Button>
        )}
      </div>

      {/* Invoice Container */}
      <div
        ref={invoiceRef}
        className="bg-white text-gray-900 p-8 max-w-4xl mx-auto border border-gray-200"
        style={{ direction: 'rtl' }}
      >
        {/* Header */}
        <div className="border-b-2 border-gray-800 pb-6 mb-6">
          <div className="flex justify-between items-start">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 mb-2">فاتورة بيع</h1>
              <p className="text-gray-600">رقم الفاتورة: {saleData.id}</p>
              <p className="text-gray-600">التاريخ: {new Date(saleData.saleDate).toLocaleDateString('ar-SA')}</p>
            </div>
            <div className="text-left">
              <h2 className="text-2xl font-bold text-gray-900 mb-2">{store.name}</h2>
              <p className="text-gray-600 text-sm">{store.address}</p>
              <p className="text-gray-600 text-sm">الهاتف: {store.phone}</p>
              {store.email && <p className="text-gray-600 text-sm">{store.email}</p>}
              {store.taxNumber && <p className="text-gray-600 text-sm">الرقم الضريبي: {store.taxNumber}</p>}
            </div>
          </div>
        </div>

        {/* Customer Information */}
        <div className="bg-gray-50 p-4 rounded-lg mb-6">
          <h3 className="font-bold text-lg mb-2 flex items-center gap-2">
            <Package className="w-5 h-5" />
            معلومات العميل
          </h3>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <p className="text-gray-600 text-sm">الاسم:</p>
              <p className="font-semibold">{saleData.customerName}</p>
            </div>
            {saleData.customerPhone && (
              <div>
                <p className="text-gray-600 text-sm">الهاتف:</p>
                <p className="font-semibold">{saleData.customerPhone}</p>
              </div>
            )}
          </div>
        </div>

        {/* Items Table */}
        <div className="mb-6">
          <h3 className="font-bold text-lg mb-4 flex items-center gap-2">
            <Layers className="w-5 h-5" />
            القطع المستعملة
          </h3>
          <table className="w-full border-collapse">
            <thead>
              <tr className="bg-gray-100 border-b-2 border-gray-800">
                <th className="p-3 text-right">القطعة</th>
                <th className="p-3 text-right">النوع</th>
                <th className="p-3 text-right">الحالة</th>
                <th className="p-3 text-right">الكمية</th>
                <th className="p-3 text-right">سعر البيع</th>
                <th className="p-3 text-right">الإجمالي</th>
              </tr>
            </thead>
            <tbody>
              {saleData.items.map((item, index) => (
                <tr key={index} className="border-b border-gray-200">
                  <td className="p-3">
                    <div className="font-semibold">{item.name}</div>
                    {item.specifications && Object.keys(item.specifications).length > 0 && (
                      <div className="text-xs text-gray-600 mt-1">
                        {Object.entries(item.specifications).map(([key, value]) => (
                          <span key={key} className="ml-2">
                            {key}: {value}
                          </span>
                        ))}
                      </div>
                    )}
                  </td>
                  <td className="p-3">
                    {item.partType && (
                      <div
                        className="inline-flex items-center gap-1 px-2 py-1 rounded text-sm"
                        style={{
                          backgroundColor: `${item.partTypeColor}20`,
                          color: item.partTypeColor,
                          border: `1px solid ${item.partTypeColor}40`,
                        }}
                      >
                        {item.partType}
                      </div>
                    )}
                  </td>
                  <td className="p-3">
                    <div className="flex items-center gap-1">
                      {getConditionIcon(item.condition)}
                      <span className="text-sm">{getConditionText(item.condition)}</span>
                    </div>
                    {item.grade && (
                      <div className="text-xs text-gray-600">{getConditionText(item.grade)}</div>
                    )}
                  </td>
                  <td className="p-3 text-center">{item.quantity}</td>
                  <td className="p-3 text-center font-semibold">{item.sellingPrice.toFixed(2)} ₪</td>
                  <td className="p-3 text-center font-bold">{item.total.toFixed(2)} ₪</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Warranty Information */}
        {saleData.items.some(item => item.warranty) && (
          <div className="bg-blue-50 p-4 rounded-lg mb-6 border border-blue-200">
            <h3 className="font-bold text-lg mb-2 flex items-center gap-2">
              <Clock className="w-5 h-5 text-blue" />
              معلومات الضمان
            </h3>
            <div className="space-y-2">
              {saleData.items.map((item, index) => (
                item.warranty && (
                  <div key={index} className="text-sm">
                    <span className="font-semibold">{item.name}:</span>
                    <span className="mr-2">{item.warranty}</span>
                  </div>
                )
              ))}
            </div>
          </div>
        )}

        {/* Totals */}
        <div className="bg-gray-50 p-4 rounded-lg mb-6">
          <div className="space-y-2">
            <div className="flex justify-between border-t-2 border-gray-800 pt-2">
              <span className="font-bold text-lg">الإجمالي:</span>
              <span className="font-bold text-lg">{saleData.total.toFixed(2)} ₪</span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-600">المدفوع:</span>
              <span className="font-semibold">{saleData.paidAmount.toFixed(2)} ₪</span>
            </div>
            {saleData.remaining > 0 && (
              <div className="flex justify-between text-red">
                <span className="font-semibold">المتبقي:</span>
                <span className="font-semibold">{saleData.remaining.toFixed(2)} ₪</span>
              </div>
            )}
          </div>
        </div>

        {/* Payment Information */}
        <div className="bg-gray-50 p-4 rounded-lg mb-6">
          <h3 className="font-bold text-lg mb-2">طريقة الدفع</h3>
          <p className="text-gray-600">{getPaymentMethodText(saleData.paymentMethod)}</p>
        </div>

        {/* Notes */}
        {saleData.notes && (
          <div className="bg-yellow-50 p-4 rounded-lg mb-6 border border-yellow-200">
            <h3 className="font-bold text-lg mb-2">ملاحظات</h3>
            <p className="text-gray-600">{saleData.notes}</p>
          </div>
        )}

        {/* Footer */}
        <div className="border-t-2 border-gray-800 pt-6 mt-6">
          <div className="grid grid-cols-2 gap-8">
            <div>
              <p className="text-gray-600 text-sm mb-2">توقيع البائع:</p>
              <div className="h-12 border-b border-gray-400"></div>
            </div>
            <div>
              <p className="text-gray-600 text-sm mb-2">توقيع العميل:</p>
              <div className="h-12 border-b border-gray-400"></div>
            </div>
          </div>
          <div className="mt-6 text-center text-gray-600 text-sm">
            <p>شكراً لتعاملكم معنا</p>
            <p className="mt-1">القطع المستعملة تباع بحالتها الحالية - لا يرجع بعد البيع</p>
            <p className="mt-1">جميع القطع خضعت لفحص شامل قبل البيع</p>
          </div>
        </div>
      </div>
    </div>
  );
}