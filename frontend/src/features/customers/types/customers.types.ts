export interface Customer {
  id: string;
  code: string;
  name: string;
  phone: string;
  email?: string;
  address?: string;
  totalPurchases: number;
  paidAmount?: number;
  outstanding: number;
  lastPurchase?: string;
  credit_limit?: number;
  is_active?: boolean;
  financial_timeline?: any[];
}

export interface CustomerFormData {
  code: string;
  name: string;
  phone: string;
  email?: string;
  address?: string;
  notes?: string;
  credit_limit?: number;
  is_active?: boolean;
}

export interface CustomerStats {
  totalCustomers: number;
  activeCustomers: number;
  customersWithDebt: number;
  totalOutstanding: number;
}

export interface SortConfig {
  key: string;
  direction: 'asc' | 'desc' | null;
}
