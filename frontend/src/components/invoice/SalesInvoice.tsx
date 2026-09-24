import { useState } from 'react';
import { Printer, Download } from 'lucide-react';
import { Button } from '../../design-system/components/button';
import { invoiceDocumentFromSale, printInvoiceDocument, renderInvoiceHtml, saveInvoiceDocumentPdf } from '../../services/documents/invoice-document';

interface SalesInvoiceProps {
  saleData: {
    id: string;
    customerName: string;
    customerPhone?: string;
    saleDate: string;
    items: Array<{
      name: string;
      partType?: string;
      partTypeColor?: string;
      condition?: string;
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
    taxAmount?: number;
    notes?: string;
    paymentAllocations?: Array<{
      amount: number;
      method: string;
      check_number?: string;
      bank_name?: string;
      check_date?: string;
    }>;
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

export function SalesInvoice({ saleData, storeInfo, onClose }: SalesInvoiceProps) {
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
}
