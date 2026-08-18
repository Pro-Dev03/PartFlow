// API Response Types
export interface ApiResponse<T> {
  success: boolean;
  data: T;
  meta?: Record<string, any>;
  error?: {
    code: string;
    message: string;
  };
}

// Auth Types
export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  user: User;
  token: string;
  organization: Organization;
}

export interface User {
  id: string;
  email: string;
  name: string;
  role: string;
  permissions: string[];
  createdAt: string;
}

export interface Organization {
  id: string;
  name: string;
  settings: OrganizationSettings;
}

export interface OrganizationSettings {
  currency: string;
  language: string;
  timezone: string;
}

// Dashboard Types
export interface DashboardStats {
  todaySales: number;
  todayProfit: number;
  inventoryValue: number;
  outstandingDebts: number;
  lowStock: number;
  alerts: DashboardAlert[];
  topProducts: Product[];
}

export interface DashboardAlert {
  type: 'LOW_STOCK' | 'OVERDUE_DEBT' | 'WARRANTY_EXPIRING' | 'PENDING_INSPECTION';
  message: string;
  count: number;
  actionUrl?: string;
}

// Product Types
export interface Product {
  id: string;
  name: string;
  sku: string;
  category: string;
  brand: string;
  condition: 'new' | 'used' | 'refurbished' | 'parts_only';
  sellingPrice: number;
  purchasePrice: number;
  description?: string;
  specifications?: Record<string, any>;
  images?: string[];
  warrantyPeriod?: number;
  stock: number;
  status: 'active' | 'inactive' | 'discontinued';
  createdAt: string;
  updatedAt: string;
}

// Inventory Types
export interface InventoryItem {
  id: string;
  productId: string;
  product?: Product;
  barcode: string;
  serial?: string;
  condition: 'new' | 'used' | 'refurbished' | 'parts_only';
  purchaseCost: number;
  sellingPrice: number;
  supplierId?: string;
  supplier?: Supplier;
  location: {
    warehouse: string;
    shelf: string;
    box: string;
  };
  warranty?: {
    startDate: string;
    endDate: string;
  };
  status: 'available' | 'reserved' | 'sold' | 'in_repair';
  inspection?: Inspection;
  createdAt: string;
  updatedAt: string;
}

export interface StockMovement {
  id: string;
  itemId: string;
  type: 'purchase' | 'sale' | 'return' | 'adjustment' | 'transfer';
  quantity: number;
  notes?: string;
  createdAt: string;
}

// Sales Types
export interface Sale {
  id: string;
  invoiceNumber: string;
  customerId?: string;
  customer?: Customer;
  items: SaleItem[];
  subtotal: number;
  discount: number;
  tax: number;
  total: number;
  paidAmount: number;
  remainingAmount: number;
  paymentMethod: 'cash' | 'card' | 'bank_transfer' | 'credit';
  status: 'completed' | 'pending' | 'cancelled';
  notes?: string;
  createdAt: string;
  updatedAt: string;
}

export interface SaleItem {
  id: string;
  itemId: string;
  item?: InventoryItem;
  quantity: number;
  price: number;
  discount: number;
  total: number;
}

export interface Payment {
  id: string;
  saleId?: string;
  customerId?: string;
  amount: number;
  method: 'cash' | 'card' | 'bank_transfer';
  notes?: string;
  createdAt: string;
}

// Customer Types
export interface Customer {
  id: string;
  name: string;
  phone: string;
  email?: string;
  address?: string;
  creditLimit?: number;
  totalPurchases: number;
  outstanding: number;
  lastPurchase?: string;
  createdAt: string;
  updatedAt: string;
}

export interface LedgerEntry {
  id: string;
  type: 'sale' | 'payment' | 'refund' | 'adjustment';
  amount: number;
  balance: number;
  description: string;
  referenceId?: string;
  createdAt: string;
}

// Debt Types
export interface Debt {
  id: string;
  customerId: string;
  customer?: Customer;
  amount: number;
  paidAmount: number;
  remainingAmount: number;
  dueDate: string;
  status: 'pending' | 'overdue' | 'paid';
  createdAt: string;
  updatedAt: string;
}

// Supplier Types
export interface Supplier {
  id: string;
  name: string;
  phone: string;
  email?: string;
  address?: string;
  totalPurchases: number;
  paidAmount: number;
  outstanding: number;
  lastPurchase?: string;
  createdAt: string;
  updatedAt: string;
}

// Purchase Types
export interface Purchase {
  id: string;
  supplierId: string;
  supplier?: Supplier;
  items: PurchaseItem[];
  totalCost: number;
  paidAmount: number;
  remainingAmount: number;
  status: 'pending' | 'ordered' | 'received' | 'cancelled';
  expectedDate?: string;
  receivedDate?: string;
  notes?: string;
  createdAt: string;
  updatedAt: string;
}

export interface PurchaseItem {
  id: string;
  productId: string;
  product?: Product;
  quantity: number;
  cost: number;
  total: number;
}

// Expense Types
export interface Expense {
  id: string;
  category: 'rent' | 'salaries' | 'utilities' | 'supplies' | 'maintenance' | 'marketing' | 'shipping' | 'other';
  amount: number;
  date: string;
  description: string;
  receipt?: string;
  recurring?: boolean;
  recurringPeriod?: 'weekly' | 'monthly' | 'yearly';
  createdAt: string;
  updatedAt: string;
}

// Return Types
export interface Return {
  id: string;
  saleId: string;
  sale?: Sale;
  itemId: string;
  item?: InventoryItem;
  reason: string;
  condition: string;
  refundAmount: number;
  refundMethod: 'cash' | 'card' | 'bank_transfer';
  status: 'pending' | 'approved' | 'rejected' | 'completed';
  inspectionRequired: boolean;
  inspection?: Inspection;
  createdAt: string;
  updatedAt: string;
}

// Warranty Types
export interface Warranty {
  id: string;
  itemId: string;
  item?: InventoryItem;
  customerId: string;
  customer?: Customer;
  startDate: string;
  endDate: string;
  duration: number;
  coverage: string;
  terms?: string;
  status: 'active' | 'expired' | 'claimed';
  claims: WarrantyClaim[];
  createdAt: string;
  updatedAt: string;
}

export interface WarrantyClaim {
  id: string;
  warrantyId: string;
  reason: string;
  status: 'pending' | 'approved' | 'rejected' | 'completed';
  createdAt: string;
  updatedAt: string;
}

// Inspection Types
export interface Inspection {
  id: string;
  itemId: string;
  item?: InventoryItem;
  inspectorId: string;
  inspector?: User;
  date: string;
  type: 'power_test' | 'temperature_test' | 'performance_test' | 'ports_test' | 'visual_inspection' | 'serial_verification';
  results: InspectionResult[];
  status: 'passed' | 'failed' | 'pending';
  notes?: string;
  photos?: string[];
  createdAt: string;
  updatedAt: string;
}

export interface InspectionResult {
  test: string;
  passed: boolean;
  notes?: string;
}

// Report Types
export interface SalesReport {
  revenue: number;
  orders: number;
  itemsSold: number;
  averageSale: number;
  startDate: string;
  endDate: string;
}

export interface ProfitReport {
  revenue: number;
  cogs: number;
  grossProfit: number;
  expenses: number;
  netProfit: number;
  margin: number;
  startDate: string;
  endDate: string;
}

export interface InventoryReport {
  totalValue: number;
  byCategory: Record<string, number>;
  byCondition: Record<string, number>;
  byBrand: Record<string, number>;
  byLocation: Record<string, number>;
  lowStock: number;
  deadStock: number;
  fastMoving: number;
}

export interface DebtsReport {
  totalDebts: number;
  overdue: number;
  dueSoon: number;
  paid: number;
  startDate: string;
  endDate: string;
}

// Notification Types
export interface Notification {
  id: string;
  type: 'LOW_STOCK' | 'OVERDUE_DEBT' | 'WARRANTY_EXPIRING' | 'ITEM_RESERVED' | 'PAYMENT_RECEIVED' | 'PURCHASE_RECEIVED';
  title: string;
  message: string;
  read: boolean;
  actionUrl?: string;
  createdAt: string;
}

// Search Types
export interface SearchResult {
  type: 'product' | 'item' | 'customer' | 'supplier' | 'sale' | 'invoice';
  id: string;
  title: string;
  subtitle?: string;
  url: string;
}