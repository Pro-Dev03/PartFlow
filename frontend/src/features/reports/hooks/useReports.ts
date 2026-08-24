import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { reportsApi, inventoryApi } from '../../../services/api/endpoints';
import { ReportData } from '../types/reports.types';

export function useReports(selectedReport: string, dateRange: string) {
  const { data: reportData, isLoading: reportLoading, refetch } = useQuery({
    queryKey: ['reports', selectedReport, dateRange],
    queryFn: () => {
      switch (selectedReport) {
        case 'sales':
          return reportsApi.sales();
        case 'profit':
          return reportsApi.profit();
        case 'inventory':
          return reportsApi.inventory();
        case 'debts':
          return reportsApi.debts();
        case 'products':
          return reportsApi.products();
        case 'suppliers':
          return reportsApi.suppliers();
        case 'expenses':
          return reportsApi.expenses();
        case 'returns':
          return reportsApi.returns();
        case 'used-items':
          return inventoryApi.list({ condition: 'USED' });
        default:
          return reportsApi.sales();
      }
    },
  });

  return {
    reportData,
    reportLoading,
    refetch,
  };
}
