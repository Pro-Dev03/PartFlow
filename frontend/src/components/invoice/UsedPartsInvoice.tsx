import { useState } from 'react';
import { Printer, Download, Package, Layers, Clock, CheckCircle, XCircle } from 'lucide-react';
import { Button } from '../../design-system/components/button';
import { invoiceDocumentFromSale, printInvoiceDocument, renderInvoiceHtml, saveInvoiceDocumentPdf } from '../../services/documents/invoice-document';

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
      sku?: string;
      barcode?: string;
      discountAmount?: number;
      taxAmount?: number;
      specifications?: Record<string, any>;
      warranty?: string;
    }>;
    subtotal: number;
    discountAmount?: number;
    total: number;
    paidAmount: number;
    cashReceived?: number;
    changeAmount?: number;
    remaining: number;
    paymentMethod: string;
    paymentStatus?: string;
    invoiceNumber?: string;
    discountAmount?: number;
    taxAmount?: number;
    notes?: string;
    paymentAllocations?: Array<{
      amount: number;
      method: string;
      check_number?: string;
      bank_name?: string;
      check_date?: string;
    }>;
    notes?: string;
  };
  storeInfo?: {
    name: string;
    address: string;
    phone: string;
    email?: string;
    taxNumber?: string;
  };
  onClose?: () => void;
}

export function UsedPartsInvoice({ saleData, storeInfo, onClose }: UsedPartsInvoiceProps) {
  const [isDownloading, setIsDownloading] = useState(false);

  const persistedStoreName = typeof window !== 'undefined' ? localStorage.getItem('partflow-store-name') : null;
  const store = storeInfo || { name: persistedStoreName || 'PartFlow' };
  const documentStoreInfo = {
    storeName: store.name,
    storeAddress: storeInfo?.address,
    storePhone: storeInfo?.phone,
    storeWebsite: storeInfo?.email,
  };
  const invoiceDocument = invoiceDocumentFromSale(saleData, documentStoreInfo);

  const handlePrint = () => {
    void printInvoiceDocument(invoiceDocument);
  };

  const handleDownload = async () => {
    if (isDownloading) return;

    setIsDownloading(true);
    try {
      await saveInvoiceDocumentPdf(invoiceDocument);
    } finally {
      setIsDownloading(false);
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

  const formatInvoiceAmount = (value: number | undefined | null) => {
    const amount = Number(value ?? 0);
    return Number.isFinite(amount)
      ? amount.toLocaleString('en-US', { minimumFractionDigits: 0, maximumFractionDigits: 2 })
      : '0';
  };

  return (
    <div className="space-y-4">
      <div className="flex gap-2 print:hidden">
        <Button onClick={handlePrint} variant="primary"><Printer className="w-4 h-4 mr-2" />طباعة الفاتورة</Button>
        <Button onClick={handleDownload} variant="secondary" disabled={isDownloading}><Download className="w-4 h-4 mr-2" />{isDownloading ? 'جاري إنشاء PDF...' : 'حفظ كـ PDF'}</Button>
        {onClose && <Button onClick={onClose} variant="default">إغلاق</Button>}
      </div>
      <iframe
        title="قالب الفاتورة"
        srcDoc={renderInvoiceHtml(invoiceDocument)}
        className="h-[760px] w-full border-0 bg-white"
      />
    </div>
  );

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
        <Button
          onClick={handleDownload}
          variant="secondary"
          disabled={isDownloading}
        >
          <Download className="w-4 h-4 mr-2" />
          {isDownloading ? 'جاري إنشاء PDF...' : 'حفظ كـ PDF'}
        </Button>
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
        className="bg-white text-gray-900 p-8 max-w-4xl mx-auto border border-gray-200"
        style={{ direction: 'rtl' }}
      >
        {/* Header */}
        <div className="border-b-2 border-gray-800 pb-6 mb-6">
          <div className="flex justify-between items-start">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 mb-2">فاتورة بيع</h1>
              <p className="text-gray-600">رقم الفاتورة: {saleData.invoiceNumber || saleData.id}</p>
              <p className="text-gray-600">التاريخ: {new Date(saleData.saleDate).toLocaleString('ar-SA')}</p>
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
            تفاصيل المنتجات
          </h3>
          <table className="w-full border-collapse text-sm" dir="rtl">
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
                    {(item.sku || item.barcode) && (
                      <div className="text-xs text-gray-600 mt-1">
                        {item.sku && <span className="ml-2">SKU: {item.sku}</span>}
                        {item.barcode && <span>Barcode: {item.barcode}</span>}
                      </div>
                    )}
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
                  <td className="whitespace-nowrap p-3 text-center align-middle">{item.quantity}</td>
                  <td className="whitespace-nowrap p-3 text-center align-middle font-semibold">{formatInvoiceAmount(item.sellingPrice)} ₪</td>
                  <td className="whitespace-nowrap p-3 text-center align-middle font-bold">{formatInvoiceAmount(item.total)} ₪</td>
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
            <div className="flex justify-between">
              <span className="text-gray-600">الإجمالي قبل الضريبة:</span>
              <span className="font-semibold">{formatInvoiceAmount(saleData.subtotal)} ₪</span>
            </div>
            {(saleData.discountAmount ?? 0) > 0 && (
              <div className="flex justify-between">
                <span className="text-gray-600">الخصم:</span>
                <span className="font-semibold">{formatInvoiceAmount(saleData.discountAmount)} ₪</span>
              </div>
            )}
            {(saleData.taxAmount ?? 0) > 0 && (
              <div className="flex justify-between">
                <span className="text-gray-600">الضريبة:</span>
                <span className="font-semibold">{formatInvoiceAmount(saleData.taxAmount)} ₪</span>
              </div>
            )}
            <div className="flex justify-between border-t-2 border-gray-800 pt-2">
              <span className="font-bold text-lg">الإجمالي:</span>
              <span className="font-bold text-lg">{formatInvoiceAmount(saleData.total)} ₪</span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-600">المدفوع:</span>
              <span className="font-semibold">{formatInvoiceAmount(saleData.paidAmount)} ₪</span>
            </div>
            {(saleData.cashReceived ?? 0) > saleData.paidAmount && (
              <>
                <div className="flex justify-between">
                  <span className="text-gray-600">المبلغ المستلم:</span>
                  <span className="font-semibold">{formatInvoiceAmount(saleData.cashReceived)} ₪</span>
                </div>
                <div className="flex justify-between text-green-700">
                  <span className="font-semibold">المردود:</span>
                  <span className="font-semibold">{formatInvoiceAmount(saleData.changeAmount ?? saleData.cashReceived! - saleData.paidAmount)} ₪</span>
                </div>
              </>
            )}
            <div className={`flex justify-between ${saleData.remaining > 0 ? 'text-red' : 'text-green-700'}`}>
                <span className="font-semibold">المتبقي:</span>
                <span className="font-semibold">{formatInvoiceAmount(saleData.remaining)} ₪</span>
            </div>
          </div>
        </div>

        {/* Payment Information */}
        <div className="bg-gray-50 p-4 rounded-lg mb-6">
          <h3 className="font-bold text-lg mb-2">طريقة الدفع</h3>
          <p className="text-gray-600">{getPaymentMethodText(saleData.paymentMethod)}</p>
          {saleData.paymentAllocations && saleData.paymentAllocations.length > 0 && (
            <div className="mt-2 space-y-1 border-t border-gray-200 pt-2 text-sm">
              {saleData.paymentAllocations.map((allocation, index) => (
                <div key={`${allocation.method}-${index}`} className="flex justify-between gap-3">
                  <span>{getPaymentMethodText(allocation.method)}</span>
                  <span className="font-semibold">{formatInvoiceAmount(allocation.amount)} ₪</span>
                </div>
              ))}
            </div>
          )}
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