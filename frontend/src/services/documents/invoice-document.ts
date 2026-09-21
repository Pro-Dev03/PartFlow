import { printHtmlDocument } from './print-html';
import html2canvas from 'html2canvas';
import { jsPDF } from 'jspdf';

export interface InvoiceDocumentItem {
  name: string;
  sku?: string;
  barcode?: string;
  quantity: number;
  unitPrice: number;
  discount?: number;
  tax?: number;
  total: number;
}

export interface InvoiceDocument {
  kind: 'sale' | 'purchase';
  title: string;
  invoiceNumber: string;
  transactionId?: string;
  date: string;
  partyLabel: string;
  partyName: string;
  partyPhone?: string;
  status: string;
  paymentMethod?: string;
  items: InvoiceDocumentItem[];
  subtotal: number;
  discount: number;
  tax: number;
  total: number;
  paid: number;
  remaining: number;
  notes?: string;
  storeName?: string;
  storePhone?: string;
  storeAddress?: string;
  storeWebsite?: string;
}

const formatAmount = (value: unknown) => Number(value || 0).toLocaleString('en-US', { maximumFractionDigits: 2 });
const escapeHtml = (value: unknown) => String(value ?? '').replaceAll('&', '&amp;').replaceAll('<', '&lt;').replaceAll('>', '&gt;').replaceAll('"', '&quot;').replaceAll("'", '&#039;');
const bidi = (value: unknown, direction: 'rtl' | 'ltr' = 'rtl') => `<bdi dir="${direction}">${escapeHtml(value)}</bdi>`;
const formatInvoiceProductName = (value: unknown) => String(value ?? '').replace(/(\p{L})(?=\d)/gu, '$1 ').replace(/(\d)(?=\p{L})/gu, '$1 ');

const paymentMethodLabel = (value?: string) => ({
  cash: 'نقدًا', card: 'بطاقة', debt: 'دين', credit: 'دين', transfer: 'تحويل بنكي', bank_transfer: 'تحويل بنكي', checks: 'شيك',
}[String(value || '').toLowerCase()] || value || '-');

const paymentStatusLabel = (value: unknown, remaining: number, paid: number) => ({
  paid: 'مدفوعة',
  partial: 'مدفوعة جزئيًا',
  debt: 'مدفوعة جزئيًا',
  pending: 'غير مدفوعة',
}[String(value || '').toLowerCase()] || (remaining <= 0 ? 'مدفوعة' : paid > 0 ? 'مدفوعة جزئيًا' : 'غير مدفوعة'));

export function invoiceDocumentFromSale(saleData: any, storeInfo?: Partial<InvoiceDocument>): InvoiceDocument {
  const total = Number(saleData?.total ?? saleData?.total_amount ?? 0);
  const paid = Number(saleData?.paidAmount ?? saleData?.paid_amount ?? 0);
  const remaining = Math.max(0, Number(saleData?.remaining ?? total - paid));
  const status = paymentStatusLabel(saleData?.paymentStatus || saleData?.payment_status, remaining, paid);
  return {
    kind: 'sale', title: 'فاتورة بيع', invoiceNumber: String(saleData?.invoiceNumber || saleData?.invoice_number || saleData?.id || '-'),
    transactionId: saleData?.id, date: saleData?.saleDate || saleData?.sale_date || saleData?.created_at || new Date().toISOString(),
    partyLabel: 'العميل', partyName: saleData?.customerName || saleData?.customer_name || 'عميل نقدي', partyPhone: saleData?.customerPhone || saleData?.customer_phone,
    status, paymentMethod: paymentMethodLabel(saleData?.paymentMethod || saleData?.payment_method),
    items: Array.isArray(saleData?.items) ? saleData.items.map((item: any) => ({
      name: item.name || item.product_name || item.product?.name || 'منتج', sku: item.sku || item.product?.sku, barcode: item.barcode || item.product?.barcode,
      quantity: Number(item.quantity || 0), unitPrice: Number(item.sellingPrice ?? item.unit_price ?? 0), discount: Number(item.discountAmount ?? item.discount_amount ?? 0),
      tax: Number(item.taxAmount ?? item.tax_amount ?? 0), total: Number(item.total ?? item.total_amount ?? 0),
    })) : [],
    subtotal: Number(saleData?.subtotal ?? 0), discount: Number(saleData?.discountAmount ?? saleData?.discount_amount ?? 0), tax: Number(saleData?.taxAmount ?? saleData?.tax_amount ?? 0), total, paid, remaining, notes: saleData?.notes,
    storeName: storeInfo?.storeName || saleData?.storeName || saleData?.store_name, storePhone: storeInfo?.storePhone || saleData?.storePhone || saleData?.store_phone,
    storeAddress: storeInfo?.storeAddress || saleData?.storeAddress || saleData?.store_address, storeWebsite: storeInfo?.storeWebsite || saleData?.storeWebsite || saleData?.store_website,
  };
}

export function invoiceDocumentFromPurchase({ purchase, supplier, items, storeInfo }: { purchase: any; supplier?: any; items: any[]; storeInfo?: Partial<InvoiceDocument> }): InvoiceDocument {
  const total = Number(purchase?.total_amount || 0);
  const paid = Number(purchase?.paid_amount || 0);
  const remaining = Math.max(0, Number(purchase?.remaining ?? total - paid));
  return {
    kind: 'purchase', title: 'فاتورة شراء من مورد', invoiceNumber: String(purchase?.invoice_number || purchase?.purchase_number || purchase?.id || '-'), transactionId: purchase?.id,
    date: purchase?.purchase_date || purchase?.created_at || new Date().toISOString(), partyLabel: 'المورد', partyName: supplier?.name || purchase?.supplier_name || 'غير محدد',
    status: remaining <= 0 ? 'مدفوعة' : paid > 0 ? 'مدفوعة جزئيًا' : 'غير مدفوعة',
    items: Array.isArray(items) ? items.map((item: any) => ({ name: item.product_name || item.product?.name || 'قطعة', sku: item.sku || item.product?.sku, barcode: item.barcode || item.product?.barcode, quantity: Number(item.quantity || 0), unitPrice: Number(item.unit_cost ?? item.unit_price ?? 0), total: Number(item.total_amount ?? Number(item.quantity || 0) * Number(item.unit_cost ?? item.unit_price ?? 0)) })) : [],
    subtotal: Number(purchase?.subtotal ?? 0), discount: Number(purchase?.discount_amount || 0), tax: Number(purchase?.tax_amount || 0), total, paid, remaining, notes: purchase?.notes,
    storeName: storeInfo?.storeName || purchase?.storeName || purchase?.store_name, storePhone: storeInfo?.storePhone || purchase?.storePhone || purchase?.store_phone,
    storeAddress: storeInfo?.storeAddress || purchase?.storeAddress || purchase?.store_address, storeWebsite: storeInfo?.storeWebsite || purchase?.storeWebsite || purchase?.store_website,
  };
}

export function renderInvoiceHtml(document: InvoiceDocument): string {
  const date = new Date(document.date);
  const dateText = Number.isNaN(date.valueOf()) ? '-' : `${date.toLocaleDateString('ar-SA').replace(/\s*\/\s*/g, ' \u00a0/\u00a0 ')}\u00a0\u00a0`;
  const timeText = Number.isNaN(date.valueOf()) ? '-' : `\u00a0\u00a0${date.toLocaleTimeString('ar-SA', { hour: '2-digit', minute: '2-digit' })}`;
  const rows = document.items.map((item) => `<tr><td class="product"><strong>${escapeHtml(formatInvoiceProductName(item.name))}</strong>${item.sku ? `<small>SKU ${bidi(item.sku, 'ltr')}</small>` : ''}${item.barcode ? `<small>Barcode ${bidi(item.barcode, 'ltr')}</small>` : ''}</td><td class="center">${bidi(formatAmount(item.quantity), 'ltr')}</td><td class="numeric">₪${bidi(formatAmount(item.unitPrice), 'ltr')}</td><td class="numeric">${item.discount ? `-₪${bidi(formatAmount(item.discount), 'ltr')}` : '-'}</td><td class="numeric">${item.tax ? `₪${bidi(formatAmount(item.tax), 'ltr')}` : '-'}</td><td class="numeric strong">₪${bidi(formatAmount(item.total), 'ltr')}</td></tr>`).join('');
  const storeName = document.storeName || 'PartFlow';
  const storeDetails = [document.storePhone, document.storeAddress, document.storeWebsite].filter(Boolean).map(escapeHtml).join(' · ');
  const discountRow = document.discount ? `<div class="amount-row"><span>الخصم</span><strong>-₪${bidi(formatAmount(document.discount), 'ltr')}</strong></div>` : '';
  return `<!doctype html><html lang="ar" dir="rtl"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>${escapeHtml(document.invoiceNumber)}</title><style>
@font-face{font-family:InvoiceArabic;src:url('/fonts/NotoNaskhArabic.ttf') format('truetype');font-weight:400;font-display:swap}@page{size:A4;margin:12mm}*{box-sizing:border-box}html,body{margin:0;background:#f3f5f7;color:#17212b;font-family:InvoiceArabic,"Noto Naskh Arabic",serif;font-size:12px;line-height:1.55;direction:rtl}body{padding:20px}.invoice{width:186mm;margin:auto;background:#fff;padding:15mm 14mm;box-shadow:0 1px 10px #17212b18}.header{display:flex;justify-content:space-between;gap:24px;border-bottom:2px solid #17212b;padding-bottom:14px}.brand{font-family:"Segoe UI",Arial,sans-serif;font-size:23px;font-weight:700;direction:ltr;text-align:right}.brand small{display:block;font-family:InvoiceArabic,serif;font-size:12px;font-weight:400;color:#536170;direction:rtl}.title{text-align:left}.title h1{margin:0;font-size:24px;line-height:1.2}.title p{margin:3px 0 0;color:#536170}.meta{display:grid;grid-template-columns:1.25fr 1fr 1fr;margin:18px 0 20px;border:1px solid #cbd3da}.meta-cell{min-height:57px;padding:9px 11px;border-left:1px solid #cbd3da}.meta-cell:last-child{border-left:0}.label{display:block;color:#536170;font-size:10px;margin-bottom:2px}.value{font-weight:700}.section-title{margin:0 0 7px;font-size:14px;border-right:3px solid #17212b;padding-right:8px}table{width:100%;border-collapse:collapse;table-layout:fixed}thead{display:table-header-group}th{padding:8px 7px;background:#e9edf0;border-top:1px solid #9eabb6;border-bottom:1px solid #9eabb6;text-align:right;font-weight:700}td{padding:8px 7px;border-bottom:1px solid #d7dde2;vertical-align:top}th:nth-child(1){width:34%}th:nth-child(2){width:9%}th:nth-child(3){width:14%}th:nth-child(4),th:nth-child(5){width:12%}th:nth-child(6){width:19%}.product strong{display:block;font-size:12px}.product small{display:block;color:#536170;font-family:"Segoe UI",Arial,sans-serif;font-size:9px;direction:ltr;text-align:right;unicode-bidi:isolate}.center{text-align:center}.numeric{text-align:left;direction:ltr;unicode-bidi:isolate;white-space:nowrap}.strong{font-weight:700}.bottom{display:flex;justify-content:space-between;gap:22px;align-items:flex-start;margin-top:21px;break-inside:avoid}.summary{width:88mm;border-top:2px solid #17212b}.payment{width:67mm;border:1px solid #cbd3da;padding:10px 12px}.payment h3{margin:0 0 5px;font-size:12px}.amount-row{display:flex;justify-content:space-between;gap:14px;padding:6px 0;border-bottom:1px solid #d7dde2}.amount-row strong{direction:ltr;unicode-bidi:isolate;display:inline-block;white-space:nowrap}.amount-row:last-child{border-bottom:0}.final{font-size:16px;padding:9px 0;border-bottom:2px solid #17212b}.footer{margin-top:30px;padding-top:9px;border-top:1px solid #cbd3da;color:#536170;text-align:center;font-size:11px}.notes{margin-top:17px;color:#536170;font-size:10px;white-space:pre-wrap}@media print{html,body{background:#fff}body{padding:0}.invoice{width:auto;padding:0;box-shadow:none}.no-print{display:none!important}}
}</style></head><body><main class="invoice"><header class="header"><div class="brand">${escapeHtml(storeName)}<small>نظام إدارة المتجر</small>${storeDetails ? `<small>${storeDetails}</small>` : ''}</div><div class="title"><h1>${escapeHtml(document.title)}</h1><p>رقم الفاتورة: ${bidi(document.invoiceNumber, 'ltr')}</p><p>${escapeHtml(dateText)} · ${escapeHtml(timeText)}</p></div></header><section class="meta"><div class="meta-cell"><span class="label">بيانات ${escapeHtml(document.partyLabel)}</span><span class="value">${escapeHtml(document.partyName)}${document.partyPhone ? ` · ${bidi(document.partyPhone, 'ltr')}` : ''}</span></div><div class="meta-cell"><span class="label">طريقة الدفع</span><span class="value">${escapeHtml(document.paymentMethod || '-')}</span></div><div class="meta-cell"><span class="label">حالة الدفع</span><span class="value">${escapeHtml(document.status)}</span></div></section><h2 class="section-title">تفاصيل المنتجات</h2><table><thead><tr><th>المنتج</th><th>الكمية</th><th>سعر الوحدة</th><th>الخصم</th><th>الضريبة</th><th>الإجمالي</th></tr></thead><tbody>${rows || '<tr><td colspan="6">لا توجد منتجات</td></tr>'}</tbody></table><section class="bottom"><section class="summary"><div class="amount-row"><span>الإجمالي قبل الضريبة</span><strong>₪${bidi(formatAmount(document.subtotal), 'ltr')}</strong></div>${discountRow}<div class="amount-row"><span>الضريبة</span><strong>₪${bidi(formatAmount(document.tax), 'ltr')}</strong></div><div class="amount-row final"><span>الإجمالي النهائي</span><strong>₪${bidi(formatAmount(document.total), 'ltr')}</strong></div></section><section class="payment"><h3>ملخص الدفع</h3><div class="amount-row"><span>المدفوع</span><strong>₪${bidi(formatAmount(document.paid), 'ltr')}</strong></div><div class="amount-row"><span>المتبقي</span><strong>₪${bidi(formatAmount(document.remaining), 'ltr')}</strong></div><div class="amount-row"><span>الحالة</span><strong>${escapeHtml(document.status)}</strong></div></section></section>${document.notes ? `<div class="notes">${escapeHtml(document.notes)}</div>` : ''}<footer class="footer">شكرًا لتعاملكم معنا</footer></main></body></html>`;
}

const fileNameFor = (document: InvoiceDocument) => `${document.kind === 'sale' ? 'فاتورة-بيع' : 'فاتورة-مورد'}-${document.invoiceNumber.replace(/[<>:"/\\|?*]/g, '-')}.pdf`;

const saveBrowserPdf = async (document: InvoiceDocument, fileName: string): Promise<void> => {
  const frame = window.document.createElement('iframe');
  frame.setAttribute('aria-hidden', 'true');
  frame.style.position = 'fixed';
  frame.style.left = '-10000px';
  frame.style.top = '0';
  frame.style.width = '210mm';
  frame.style.height = '297mm';
  frame.style.border = '0';
  frame.srcdoc = renderInvoiceHtml(document);
  window.document.body.appendChild(frame);

  try {
    await new Promise<void>((resolve, reject) => {
      const frameWindow = frame.contentWindow;
      if (!frameWindow) {
        reject(new Error('تعذر إنشاء مستند الفاتورة'));
        return;
      }
      frame.addEventListener('load', () => resolve(), { once: true });
      window.setTimeout(() => reject(new Error('انتهت مهلة تجهيز الفاتورة')), 15000);
    });

    const frameDocument = frame.contentDocument;
    const invoice = frameDocument?.querySelector<HTMLElement>('.invoice');
    if (!invoice) throw new Error('تعذر العثور على محتوى الفاتورة');
    if (frameDocument?.fonts) await frameDocument.fonts.ready;

    const canvas = await html2canvas(invoice, {
      backgroundColor: '#ffffff',
      scale: 2,
      useCORS: true,
      logging: false,
    });
    const pdf = new jsPDF({ orientation: 'portrait', unit: 'mm', format: 'a4' });
    const pageWidth = pdf.internal.pageSize.getWidth();
    const pageHeight = pdf.internal.pageSize.getHeight();
    const imageHeight = (canvas.height * pageWidth) / canvas.width;
    const imageData = canvas.toDataURL('image/jpeg', 0.95);

    let offset = 0;
    while (offset < imageHeight) {
      if (offset > 0) pdf.addPage();
      pdf.addImage(imageData, 'JPEG', 0, -offset, pageWidth, imageHeight);
      offset += pageHeight;
    }
    pdf.save(fileName);
  } finally {
    frame.remove();
  }
};

/* Legacy jsPDF export intentionally removed: Browser PDF must use the shared HTML document. */
/*
async function saveBrowserPdf(document: InvoiceDocument, fileName: string): Promise<void> {
  const pdf = new jsPDF({ unit: 'mm', format: 'a4' });
  pdf.setR2L(true);
  try {
    const fontData = await fetch('/fonts/NotoNaskhArabic.ttf').then((response) => response.arrayBuffer());
    let binary = '';
    for (const byte of new Uint8Array(fontData)) binary += String.fromCharCode(byte);
    pdf.addFileToVFS('NotoNaskhArabic.ttf', btoa(binary));
    pdf.addFont('NotoNaskhArabic.ttf', 'NotoNaskhArabic', 'normal');
    pdf.setFont('NotoNaskhArabic');
  } catch {
    pdf.setFont('helvetica');
  }

  const pageWidth = pdf.internal.pageSize.getWidth();
  const right = pageWidth - 14;
  let y = 18;
  const shape = (value: unknown) => {
    const processArabic = (pdf as jsPDF & { processArabic?: (input: string) => string }).processArabic;
    return processArabic ? processArabic(String(value ?? '')) : String(value ?? '');
  };
  const write = (value: unknown, size = 10, bold = false) => {
    pdf.setFontSize(size);
    pdf.setFont('NotoNaskhArabic', bold ? 'normal' : 'normal');
    pdf.text(shape(value), right, y, { align: 'right' });
  };
  write('PartFlow', 20, true); y += 8;
  write(document.title, 13, true); y += 7;
  write(`رقم الفاتورة: ${document.invoiceNumber}`, 10); y += 6;
  write(`رقم العملية: ${document.transactionId || '-'}`, 10); y += 6;
  write(`${document.partyLabel}: ${document.partyName}`, 10); y += 6;
  write(`التاريخ: ${new Date(document.date).toLocaleDateString('ar-SA')}`, 10); y += 6;
  write(`حالة الدفع: ${document.status}`, 10, true); y += 10;
  pdf.setFillColor(238, 242, 247);
  pdf.rect(14, y - 6, pageWidth - 28, 9, 'F');
  pdf.setFontSize(9);
  pdf.text(shape('المنتج'), right - 4, y, { align: 'right' });
  pdf.text(shape('الكمية'), right - 92, y, { align: 'center' });
  pdf.text(shape('سعر الوحدة'), right - 125, y, { align: 'center' });
  pdf.text(shape('الإجمالي'), 18, y, { align: 'center' });
  y += 10;
  for (const item of document.items) {
    if (y > 270) { pdf.addPage(); y = 18; }
    pdf.setFontSize(9);
    pdf.text(shape(`${item.name}${item.sku ? ` | SKU: ${item.sku}` : ''}`), right - 4, y, { align: 'right', maxWidth: 80 });
    pdf.text(String(item.quantity), right - 92, y, { align: 'center' });
    pdf.text(`₪${formatAmount(item.unitPrice)}`, right - 125, y, { align: 'center' });
    pdf.text(`₪${formatAmount(item.total)}`, 18, y, { align: 'center' });
    pdf.setDrawColor(210, 216, 225); pdf.line(14, y + 3, right, y + 3); y += 8;
  }
  y += 10;
  for (const [label, value] of [['الإجمالي قبل الضريبة', document.subtotal], ['الخصم', -document.discount], ['الضريبة', document.tax], ['الإجمالي النهائي', document.total], ['المدفوع', document.paid], ['المتبقي', document.remaining]] as const) {
    if (label === 'الخصم' && !document.discount) continue;
    write(`${label}: ₪${formatAmount(value)}`, label === 'الإجمالي النهائي' ? 12 : 10, label === 'الإجمالي النهائي'); y += 7;
  }

  const browserWindow = window as Window & { showSaveFilePicker?: (options: unknown) => Promise<{ createWritable: () => Promise<{ write: (data: Blob) => Promise<void>; close: () => Promise<void> }> }> };
  const blob = pdf.output('blob');
  if (browserWindow.showSaveFilePicker) {
    const handle = await browserWindow.showSaveFilePicker({ suggestedName: fileName, types: [{ description: 'PDF', accept: { 'application/pdf': ['.pdf'] } }] });
    const writable = await handle.createWritable(); await writable.write(blob); await writable.close();
  } else {
    const url = URL.createObjectURL(blob); const link = window.document.createElement('a'); link.href = url; link.download = fileName; link.click(); URL.revokeObjectURL(url);
  }
}
*/

export function printInvoiceDocument(document: InvoiceDocument): Promise<boolean> {
  return printHtmlDocument(renderInvoiceHtml(document));
}

export async function saveInvoiceDocumentPdf(document: InvoiceDocument): Promise<string | null> {
  const fileName = fileNameFor(document);
  if (window.partflowDesktop?.invoice) {
    const result = await window.partflowDesktop.invoice.savePdf({ html: renderInvoiceHtml(document), fileName });
    return result.canceled ? null : result.filePath || fileName;
  }

  await saveBrowserPdf(document, fileName);
  return fileName;

  /* Removed legacy jsPDF export. The shared HTML template is the only Browser PDF path.
  const pdf = new jsPDF({ orientation: 'portrait', unit: 'mm', format: 'a4' });
  pdf.addFileToVFS('NotoNaskhArabic.ttf', await loadArabicFont());
  pdf.addFont('NotoNaskhArabic.ttf', 'NotoNaskhArabic', 'normal');
  pdf.setFont('NotoNaskhArabic', 'normal');
  pdf.setLanguage('ar');
  pdf.setR2L(true);
  const right = 196;
  let y = 18;
  const write = (value: unknown, size = 10, x = right) => {
    pdf.setFontSize(size);
    pdf.text(shapeAndReorder(value), x, y, { align: x === right ? 'right' : 'left' });
    y += size >= 14 ? 8 : 6;
  };
  write(document.storeName || 'PartFlow', 20);
  write(document.title, 13);
  write(`رقم الفاتورة: ${document.invoiceNumber}`);
  write(`التاريخ: ${new Date(document.date).toLocaleDateString('ar-SA')}`);
  write(`${document.partyLabel}: ${document.partyName}`);
  write(`حالة الدفع: ${document.status}`);
  y += 5;
  pdf.setFillColor(233, 237, 240);
  pdf.rect(14, y - 6, 182, 9, 'F');
  write('المنتج', 9, right - 4);
  pdf.text(shapeAndReorder('الكمية'), 145, y - 6, { align: 'center' });
  pdf.text(shapeAndReorder('سعر الوحدة'), 112, y - 6, { align: 'center' });
  pdf.text(shapeAndReorder('الإجمالي'), 30, y - 6, { align: 'center' });
  y += 5;
  for (const item of document.items) {
    if (y > 270) { pdf.addPage(); y = 18; }
    pdf.setFontSize(9);
    pdf.text(shapeAndReorder(`${item.name}${item.sku ? ` | SKU ${item.sku}` : ''}`), right - 4, y, { align: 'right', maxWidth: 82 });
    pdf.text(shapeAndReorder(String(item.quantity)), 145, y, { align: 'center' });
    pdf.text(shapeAndReorder(`₪${formatAmount(item.unitPrice)}`), 112, y, { align: 'center' });
    pdf.text(shapeAndReorder(`₪${formatAmount(item.total)}`), 30, y, { align: 'center' });
    pdf.setDrawColor(215, 221, 226);
    pdf.line(14, y + 3, 196, y + 3);
    y += 8;
  }
  y += 10;
  for (const [label, value] of [['الإجمالي قبل الضريبة', document.subtotal], ['الخصم', -document.discount], ['الضريبة', document.tax], ['الإجمالي النهائي', document.total], ['المدفوع', document.paid], ['المتبقي', document.remaining]] as const) {
    if (label === 'الخصم' && !document.discount) continue;
    write(`${label}: ₪${formatAmount(value)}`, label === 'الإجمالي النهائي' ? 13 : 10);
  }
  const blob = pdf.output('blob');
  const url = URL.createObjectURL(blob);
  const link = window.document.createElement('a');
  link.href = url;
  link.download = fileName;
  link.style.display = 'none';
  window.document.body.appendChild(link);
  link.click();
  link.remove();
  URL.revokeObjectURL(url);
  return fileName;
  */
}
