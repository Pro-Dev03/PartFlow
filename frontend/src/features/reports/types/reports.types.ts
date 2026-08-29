export interface ReportData {
  date?: string;
  value?: number;
  amount?: number;
  description?: string;
  name?: string;
  status?: string;
}

export interface ReportType {
  id: string;
  label: string;
  icon: any;
  group?: 'period' | 'current';
}

export interface DateRange {
  value: string;
  label: string;
}

export interface ReportStats {
  totalSales: number;
  totalProfit: number;
  transactionCount: number;
  averageOrder: number;
}
