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