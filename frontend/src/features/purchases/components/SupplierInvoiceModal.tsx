import { FileDown, Printer } from 'lucide-react';
import { Modal } from '../../../design-system/components/modal';
import { Button } from '../../../design-system/components/button';
import { toast } from 'sonner';
import { useState } from 'react';
import { invoiceDocumentFromPurchase, printInvoiceDocument, renderInvoiceHtml, saveInvoiceDocumentPdf } from '../../../services/documents/invoice-document';

const withTimeout = <T,>(operation: Promise<T>, message: string, timeoutMs = 15000): Promise<T> =>
  Promise.race([
    operation,
    new Promise<T>((_, reject) => {
      window.setTimeout(() => reject(new Error(message)), timeoutMs);
    }),
  ]);

interface SupplierInvoiceModalProps {
  isOpen: boolean;
  onClose: () => void;
  purchase: any;
  supplier?: any;
  items: any[];
}

export function SupplierInvoiceModal({ isOpen, onClose, purchase, supplier, items }: SupplierInvoiceModalProps) {
  const [isPrinting, setIsPrinting] = useState(false);
  const [isSavingPdf, setIsSavingPdf] = useState(false);

  if (!purchase) return null;

  const storeName = typeof window !== 'undefined' ? localStorage.getItem('partflow-store-name') || undefined : undefined;
  const invoicePayload = { purchase, supplier, items, storeInfo: { storeName } };
  const invoiceDocument = invoiceDocumentFromPurchase(invoicePayload);

  const handlePrint = async () => {
    setIsPrinting(true);
    try {
      await withTimeout(printInvoiceDocument(invoiceDocument), 'انتهت مهلة الطباعة');
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'تعذر فتح نظام الطباعة');
    } finally {
      setIsPrinting(false);
    }
  };

  const handleSavePdf = async () => {
    setIsSavingPdf(true);
    try {
      const fileName = await withTimeout(saveInvoiceDocumentPdf(invoiceDocument), 'انتهت مهلة حفظ PDF');
      if (fileName) toast.success(`تم حفظ ${fileName}`);
    } catch (error) {
      if ((error as { name?: string })?.name !== 'AbortError') {
        toast.error(error instanceof Error ? error.message : 'تعذر حفظ ملف PDF');
      }
    } finally {
      setIsSavingPdf(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="فاتورة التاجر" size="xl">
      <div className="supplier-invoice-print-root space-y-5">
        <div className="flex flex-wrap items-start justify-between gap-4 border-b border-border pb-4">
          <div>
            <p className="text-xs font-semibold uppercase tracking-wide text-text-muted">فاتورة التاجر</p>
            <h2 className="mt-1 text-xl font-bold text-text-primary">{purchase.invoice_number || '-'}</h2>
          </div>
          <div className="flex flex-wrap gap-2 supplier-invoice-print-button">
            <Button type="button" variant="secondary" size="sm" onClick={() => { void handlePrint(); }} disabled={isPrinting || isSavingPdf} className="gap-2" aria-label="طباعة الفاتورة">
              <Printer className="h-4 w-4" />
              {isPrinting ? 'جاري الطباعة...' : 'طباعة'}
            </Button>
            <Button type="button" variant="secondary" size="sm" onClick={() => { void handleSavePdf(); }} disabled={isPrinting || isSavingPdf} className="gap-2" aria-label="حفظ الفاتورة كـ PDF">
              <FileDown className="h-4 w-4" />
              {isSavingPdf ? 'جاري حفظ PDF...' : 'حفظ PDF'}
            </Button>
          </div>
        </div>

            <iframe title="قالب فاتورة التاجر" srcDoc={renderInvoiceHtml(invoiceDocument)} className="h-[760px] w-full border-0 bg-white" />
      </div>
    </Modal>
  );
}