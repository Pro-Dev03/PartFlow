export interface ProductCsvRow {
  name: string;
  barcode: string;
  costPrice: string;
  sellingPrice: string;
  quantity: string;
  minStockLevel: string;
  supplierId: string;
}

const requiredHeaders = ['name', 'barcode'];

function parseCsvLine(line: string): string[] {
  const values: string[] = [];
  let value = '';
  let quoted = false;

  for (let index = 0; index < line.length; index += 1) {
    const character = line[index];
    if (character === '"') {
      if (quoted && line[index + 1] === '"') {
        value += '"';
        index += 1;
      } else {
        quoted = !quoted;
      }
    } else if (character === ',' && !quoted) {
      values.push(value.trim());
      value = '';
    } else {
      value += character;
    }
  }

  if (quoted) throw new Error('CSV يحتوي على اقتباس غير مغلق');
  values.push(value.trim());
  return values;
}

function normalizeHeader(header: string): string {
  return header.replace(/^\ufeff/, '').trim().toLowerCase().replace(/[\s-]+/g, '_');
}

export function parseProductCsv(csv: string): ProductCsvRow[] {
  const lines = csv.split(/\r?\n/).filter((line) => line.trim());
  if (lines.length < 2) throw new Error('ملف CSV فارغ أو لا يحتوي على صفوف');

  const headers = parseCsvLine(lines[0]).map(normalizeHeader);
  const missing = requiredHeaders.filter((header) => !headers.includes(header));
  if (missing.length) throw new Error(`أعمدة CSV المطلوبة مفقودة: ${missing.join(', ')}`);

  const indexOf = (header: string) => headers.indexOf(header);
  return lines.slice(1).map((line, rowIndex) => {
    const values = parseCsvLine(line);
    const read = (header: string) => values[indexOf(header)] || '';
    const row = {
      name: read('name'),
      barcode: read('barcode'),
      costPrice: read('cost_price'),
      sellingPrice: read('selling_price'),
      quantity: read('quantity') || '0',
      minStockLevel: read('min_stock_level') || read('min_stock') || '0',
      supplierId: read('supplier_id') || read('preferred_supplier_id'),
    };
    if (!row.name || !row.barcode) throw new Error(`الصف ${rowIndex + 2} يحتاج اسمًا وباركودًا`);
    return row;
  });
}

export function createProductCsvTemplate(): string {
  return 'name,barcode,cost_price,selling_price,quantity,min_stock_level,supplier_id\nTasco Iced Coffee,8854419001507,,,0,0,\n';
}