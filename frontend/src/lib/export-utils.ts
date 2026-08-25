/**
 * Export Utilities - أدوات التصدير والطباعة
 */

// تصدير البيانات إلى CSV
export const exportToCSV = (data: any[], filename: string, headers?: string[]) => {
  if (!data || data.length === 0) {
    console.warn('No data to export');
    return;
  }

  const csvContent = generateCSV(data, headers);
  downloadCSV(csvContent, filename);
};

// توليد محتوى CSV
const generateCSV = (data: any[], headers?: string[]): string => {
  const csvRows: string[] = [];

  // إضافة الرؤوس
  if (headers) {
    csvRows.push(headers.join(','));
  } else {
    const keys = Object.keys(data[0]);
    csvRows.push(keys.join(','));
  }

  // إضافة البيانات
  for (const row of data) {
    const values = Object.values(row).map((value) => {
      const stringValue = String(value ?? '');
      // Escape quotes and wrap in quotes if contains comma
      if (stringValue.includes(',') || stringValue.includes('"')) {
        return `"${stringValue.replace(/"/g, '""')}"`;
      }
      return stringValue;
    });
    csvRows.push(values.join(','));
  }

  return csvRows.join('\n');
};

// تحميل ملف CSV
const downloadCSV = (csvContent: string, filename: string) => {
  const blob = new Blob(['\ufeff' + csvContent], { type: 'text/csv;charset=utf-8;' });
  const link = document.createElement('a');
  const url = URL.createObjectURL(blob);
  
  link.setAttribute('href', url);
  link.setAttribute('download', `${filename}.csv`);
  link.style.visibility = 'hidden';
  
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  
  URL.revokeObjectURL(url);
};

// تصدير البيانات إلى JSON
export const exportToJSON = (data: any[], filename: string) => {
  if (!data || data.length === 0) {
    console.warn('No data to export');
    return;
  }

  const jsonContent = JSON.stringify(data, null, 2);
  const blob = new Blob([jsonContent], { type: 'application/json;charset=utf-8;' });
  const link = document.createElement('a');
  const url = URL.createObjectURL(blob);
  
  link.setAttribute('href', url);
  link.setAttribute('download', `${filename}.json`);
  link.style.visibility = 'hidden';
  
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  
  URL.revokeObjectURL(url);
};

// طباعة المحتوى
export const printContent = (elementId: string) => {
  const element = document.getElementById(elementId);
  if (!element) {
    console.warn('Element not found for printing');
    return;
  }

  const printWindow = window.open('', '_blank');
  if (!printWindow) {
    console.warn('Failed to open print window');
    return;
  }

  const printContent = element.innerHTML;
  const printStyles = `
    <style>
      body { 
        font-family: Arial, sans-serif; 
        direction: rtl; 
        margin: 20px;
      }
      table { 
        width: 100%; 
        border-collapse: collapse; 
        margin-bottom: 20px;
      }
      th, td { 
        border: 1px solid #ddd; 
        padding: 8px; 
        text-align: right; 
      }
      th { 
        background-color: #f5f5f5; 
        font-weight: bold;
      }
      .no-print { 
        display: none; 
      }
    </style>
  `;

  printWindow.document.write(`
    <!DOCTYPE html>
    <html>
    <head>
      <title>طباعة</title>
      ${printStyles}
    </head>
    <body>
      ${printContent}
    </body>
    </html>
  `);

  printWindow.document.close();
  printWindow.print();
};

// طباعة الجدول مباشرة
export const printTable = (data: any[], headers: string[], title: string) => {
  const printWindow = window.open('', '_blank');
  if (!printWindow) {
    console.warn('Failed to open print window');
    return;
  }

  const tableRows = data.map(row => {
    const cells = Object.values(row).map(value => 
      `<td>${String(value ?? '')}</td>`
    ).join('');
    return `<tr>${cells}</tr>`;
  }).join('');

  const headerRow = headers.map(header => 
    `<th>${header}</th>`
  ).join('');

  const printContent = `
    <h1 style="text-align: center; margin-bottom: 20px;">${title}</h1>
    <table>
      <thead>
        <tr>${headerRow}</tr>
      </thead>
      <tbody>
        ${tableRows}
      </tbody>
    </table>
    <p style="text-align: center; margin-top: 20px; font-size: 12px;">
      تم التوليد بواسطة PartFlow - ${new Date().toLocaleDateString('ar-SA')}
    </p>
  `;

  const printStyles = `
    <style>
      body { 
        font-family: Arial, sans-serif; 
        direction: rtl; 
        margin: 20px;
      }
      table { 
        width: 100%; 
        border-collapse: collapse; 
        margin-bottom: 20px;
      }
      th, td { 
        border: 1px solid #ddd; 
        padding: 8px; 
        text-align: right; 
      }
      th { 
        background-color: #f5f5f5; 
        font-weight: bold;
      }
      h1 {
        color: #333;
        font-size: 24px;
      }
    </style>
  `;

  printWindow.document.write(`
    <!DOCTYPE html>
    <html>
    <head>
      <title>${title}</title>
      ${printStyles}
    </head>
    <body>
      ${printContent}
    </body>
    </html>
  `);

  printWindow.document.close();
  printWindow.print();
};

// طباعة إيصال الدفع - تصميم احترافي وعصري
export const printPaymentReceipt = (paymentData: {
  amount: number;
  method: string;
  customerName: string;
  date: string;
  language?: string;
}) => {
  const { amount, method, customerName, date, language = 'ar' } = paymentData;
  const isRTL = language === 'ar';

  // تحويل طريقة الدفع للعربية
  const methodArabic: Record<string, string> = {
    'cash': 'نقدي',
    'card': 'بطاقة',
    'bank_transfer': 'تحويل بنكي',
    'check': 'شيك',
  };

  const paymentMethod = isRTL ? (methodArabic[method] || method) : method;

  // إنشاء محتوى HTML للإيصال بتصميم احترافي
  const htmlContent = `
    <!DOCTYPE html>
    <html dir="${isRTL ? 'rtl' : 'ltr'}">
    <head>
      <meta charset="UTF-8">
      <title>${isRTL ? 'إيصال دفع' : 'Payment Receipt'}</title>
      <style>
        * {
          margin: 0;
          padding: 0;
          box-sizing: border-box;
        }
        body {
          font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
          direction: ${isRTL ? 'rtl' : 'ltr'};
          padding: 20px;
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          min-height: 100vh;
        }
        .receipt-container {
          max-width: 320px;
          margin: 0 auto;
          background: white;
          border-radius: 16px;
          box-shadow: 0 20px 60px rgba(0,0,0,0.3);
          overflow: hidden;
        }
        .header {
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          padding: 30px 25px;
          text-align: center;
          position: relative;
        }
        .header::before {
          content: '';
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background: url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23ffffff' fill-opacity='0.1'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E");
          opacity: 0.5;
        }
        .logo {
          font-size: 32px;
          font-weight: 700;
          color: white;
          margin-bottom: 8px;
          position: relative;
          letter-spacing: 2px;
        }
        .system-name {
          font-size: 11px;
          color: rgba(255,255,255,0.9);
          font-weight: 500;
          letter-spacing: 1px;
          text-transform: uppercase;
          position: relative;
        }
        .receipt-number {
          font-size: 10px;
          color: rgba(255,255,255,0.7);
          margin-top: 12px;
          position: relative;
          font-family: 'Courier New', monospace;
        }
        .content {
          padding: 25px;
        }
        .section {
          margin-bottom: 20px;
        }
        .section-title {
          font-size: 11px;
          font-weight: 600;
          color: #667eea;
          margin-bottom: 12px;
          text-transform: uppercase;
          letter-spacing: 1px;
          padding-bottom: 8px;
          border-bottom: 2px solid #f0f0f0;
        }
        .info-row {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 10px;
          font-size: 13px;
        }
        .info-label {
          color: #6b7280;
          font-weight: 500;
        }
        .info-value {
          font-weight: 600;
          color: #1f2937;
        }
        .amount-section {
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          padding: 25px;
          border-radius: 12px;
          margin: 20px 0;
          text-align: center;
          position: relative;
          overflow: hidden;
        }
        .amount-section::before {
          content: '';
          position: absolute;
          top: -50%;
          left: -50%;
          width: 200%;
          height: 200%;
          background: radial-gradient(circle, rgba(255,255,255,0.1) 0%, transparent 70%);
          animation: shine 3s infinite;
        }
        @keyframes shine {
          0% { transform: translate(-30%, -30%) rotate(0deg); }
          100% { transform: translate(30%, 30%) rotate(360deg); }
        }
        .amount-label {
          font-size: 11px;
          color: rgba(255,255,255,0.9);
          margin-bottom: 8px;
          font-weight: 500;
          text-transform: uppercase;
          letter-spacing: 1px;
          position: relative;
        }
        .amount-value {
          font-size: 36px;
          font-weight: 700;
          color: white;
          line-height: 1;
          position: relative;
        }
        .amount-currency {
          font-size: 20px;
          margin-left: 4px;
        }
        .divider {
          height: 1px;
          background: linear-gradient(to right, transparent, #e0e0e0, transparent);
          margin: 20px 0;
        }
        .footer {
          text-align: center;
          padding: 20px 25px;
          background: #f9fafb;
          border-top: 1px solid #e5e7eb;
        }
        .footer-text {
          font-size: 11px;
          color: #6b7280;
          margin-bottom: 8px;
          font-weight: 500;
        }
        .terms {
          font-size: 9px;
          color: #9ca3af;
          line-height: 1.6;
        }
        .status-badge {
          display: inline-block;
          padding: 4px 12px;
          background: #10b981;
          color: white;
          border-radius: 20px;
          font-size: 10px;
          font-weight: 600;
          text-transform: uppercase;
          letter-spacing: 0.5px;
          margin-top: 8px;
        }
        @media print {
          body {
            background: white;
            padding: 0;
          }
          .receipt-container {
            box-shadow: none;
            border: 1px solid #e5e7eb;
            margin: 0;
          }
          .amount-section::before {
            display: none;
          }
        }
      </style>
    </head>
    <body>
      <div class="receipt-container">
        <!-- Header -->
        <div class="header">
          <div class="logo">PartFlow</div>
          <div class="system-name">${isRTL ? 'نظام إدارة المتاجر' : 'Store Management System'}</div>
          <div class="receipt-number">
            ${isRTL ? 'رقم الإيصال: ' : 'Receipt #: '}${new Date().getTime().toString().slice(-8)}
          </div>
          <div class="status-badge">${isRTL ? 'مدفوع' : 'PAID'}</div>
        </div>

        <div class="content">
          <!-- Customer Info -->
          <div class="section">
            <div class="section-title">${isRTL ? 'معلومات العميل' : 'Customer Information'}</div>
            <div class="info-row">
              <span class="info-label">${isRTL ? 'الاسم:' : 'Name:'}</span>
              <span class="info-value">${customerName}</span>
            </div>
          </div>

          <!-- Payment Details -->
          <div class="section">
            <div class="section-title">${isRTL ? 'تفاصيل الدفع' : 'Payment Details'}</div>
            <div class="info-row">
              <span class="info-label">${isRTL ? 'طريقة الدفع:' : 'Method:'}</span>
              <span class="info-value">${paymentMethod}</span>
            </div>
            <div class="info-row">
              <span class="info-label">${isRTL ? 'التاريخ:' : 'Date:'}</span>
              <span class="info-value">${date}</span>
            </div>
          </div>

          <!-- Amount -->
          <div class="amount-section">
            <div class="amount-label">${isRTL ? 'المبلغ المدفوع' : 'Amount Paid'}</div>
            <div class="amount-value">
              <span class="amount-currency">₪</span>${amount.toFixed(2)}
            </div>
          </div>
        </div>

        <div class="divider"></div>

        <!-- Footer -->
        <div class="footer">
          <div class="footer-text">${isRTL ? 'شكراً لتعاملكم معنا' : 'Thank you for your business'}</div>
          <div class="terms">
            ${isRTL ? '• هذا الإيصال إثبات للدفع المسجل' : '• This receipt is proof of recorded payment'}
            <br>
            ${isRTL ? '• يرجى الاحتفاظ به للمراجعة' : '• Please keep for reference'}
            <br>
            ${isRTL ? '• لأي استفسار، يرجى التواصل مع الإدارة' : '• For inquiries, contact management'}
          </div>
        </div>
      </div>
    </body>
    </html>
  `;

  // إنشاء نافذة جديدة للطباعة
  const printWindow = window.open('', '_blank');
  if (!printWindow) {
    console.warn('Failed to open print window');
    return;
  }

  printWindow.document.write(htmlContent);
  printWindow.document.close();

  // انتظار تحميل المحتوى ثم الطباعة
  printWindow.onload = function() {
    printWindow.print();
  };
};