export interface Customer {
  id: string;
  name: string;
  phone: string;
  email?: string;
  totalPurchases: number;
  outstanding: number;
  financial_timeline?: any[];
}

export interface CustomerFormData {
  name: string;
  phone: string;
  email?: string;
  address?: string;
  notes?: string;
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