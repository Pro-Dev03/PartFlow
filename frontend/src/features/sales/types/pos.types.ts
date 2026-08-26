export interface CartItem {
  id: string;
  name: string;
  barcode: string;
  price: number;
  quantity: number;
  total: number;
  stock?: number;
  isTradeIn?: boolean;
  purchaseCost?: number;
  condition?: string;
  partType?: string;
  partTypeColor?: string;
  grade?: string;
}

export type PaymentMethod = 'cash' | 'card' | 'credit'; // Simplified for POS (SALES-PHILOSOPHY.md)

export type ItemInputMethodType = 'barcode' | 'camera' | 'manual';

export interface InvoiceData {
  id: string;
  customerName: string;
  customerPhone?: string;
  saleDate: string;
  items: InvoiceItem[];
  subtotal: number;
  total: number;
  paidAmount: number;
  remaining: number;
  paymentMethod: PaymentMethod;
}

export interface InvoiceItem {
  name: string;
  partType?: string;
  partTypeColor?: string;
  condition?: string;
  grade?: string;
  sellingPrice: number;
  quantity: number;
  total: number;
}

export interface TradeInFormData {
  customerId?: string;
  customerName?: string;
  productId?: string;
  productName?: string;
  partTypeId: string;
  purchaseCost: number;
  specifications: any[];
}

export interface POSState {
  cart: CartItem[];
  barcodeInput: string;
  searchQuery: string;
  selectedCustomer: string;
  paymentMethod: PaymentMethod;
  paidAmount: string;
  isProcessing: boolean;
  soundEnabled: boolean;
  quickAddMode: boolean;
  inputMethod: ItemInputMethodType;
}