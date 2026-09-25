import { describe, expect, it } from 'vitest';
import { invoiceDocumentFromSale, renderInvoiceHtml } from '../../services/documents/invoice-document';

const document = invoiceDocumentFromSale({
  id: 'sale-1',
  invoiceNumber: 'INV-1',
  customerName: 'عميل اختبار',
  saleDate: '2026-09-25T11:00:00Z',
  items: [{ name: 'منتج اختبار', quantity: 1, sellingPrice: 10, total: 10 }],
  subtotal: 10,
  total: 10,
  paidAmount: 10,
  remaining: 0,
  paymentMethod: 'cash',
});

describe('invoice document formats', () => {
  it('renders thermal print documents at 80mm', () => {
    const html = renderInvoiceHtml(document);

    expect(html).toContain('@page{size:80mm auto;margin:0}');
    expect(html).toContain('.invoice{width:80mm');
  });

  it('keeps A4 available for PDF export', () => {
    const html = renderInvoiceHtml(document, 'a4');

    expect(html).toContain('@page{size:A4;margin:12mm}');
    expect(html).not.toContain('@page{size:80mm auto;margin:0}');
  });
});