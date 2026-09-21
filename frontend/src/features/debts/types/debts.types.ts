export interface Debt {
  id: string;
  customer_id: string;
  amount: number;
  remaining_amount: number;
  due_date: string;
  notes?: string;
  debt_reason?: string;
  status: 'pending' | 'partial' | 'paid' | 'overdue';
  created_at: string;
  customer?: {
    id: string;
    name: string;
    code: string;
    notes?: string;
  };
}

export interface DebtPayment {
  id: string;
  debt_id: string;
  amount: number;
  payment_date: string;
  notes?: string;
}

export interface DebtStats {
  totalDebt: number;
  paidAmount: number;
  remainingAmount: number;
  overdueCount: number;
  customerCount: number;
}

export interface DebtSummaryResponse {
  total_debt: number;
  paid_amount: number;
  remaining_amount: number;
  customer_count: number;
}

export interface DebtAging {
  status: 'PAID' | 'DUE_SOON' | 'CURRENT' | 'OVERDUE_1_7' | 'OVERDUE_8_14' | 'OVERDUE_15_30' | 'OVERDUE_30_PLUS';
  daysOverdue?: number;
  label: string;
  color: string;
}
