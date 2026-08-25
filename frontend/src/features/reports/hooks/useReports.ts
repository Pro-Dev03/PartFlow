import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { reportsApi, inventoryApi } from '../../../services/api/endpoints';
import { ReportData } from '../types/reports.types';

export function useReports(selectedReport: string, dateRange: string) {
  // Calculate date range based on selection
  const getDateRangeParams = () => {
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    
    switch (dateRange) {
      case 'today':
        return {
          start_date: today.toISOString().split('T')[0],
          end_date: today.toISOString().split('T')[0]
        };
      case 'thisWeek':
        const weekStart = new Date(today);
        weekStart.setDate(today.getDate() - today.getDay());
        return {
          start_date: weekStart.toISOString().split('T')[0],
          end_date: today.toISOString().split('T')[0]
        };
      case 'thisMonth':
        const monthStart = new Date(today.getFullYear(), today.getMonth(), 1);
        return {
          start_date: monthStart.toISOString().split('T')[0],
          end_date: today.toISOString().split('T')[0]
        };
      case 'thisYear':
        const yearStart = new Date(today.getFullYear(), 0, 1);
        return {
          start_date: yearStart.toISOString().split('T')[0],
          end_date: today.toISOString().split('T')[0]
        };
      default:
        return {};
    }
  };

  const dateParams = getDateRangeParams();

  const { data: reportData, isLoading: reportLoading, refetch } = useQuery({
    queryKey: ['reports', selectedReport, dateRange],
    queryFn: () => {
      switch (selectedReport) {
        case 'sales':
          return reportsApi.sales(dateParams);
        case 'profit':
          return reportsApi.profit(dateParams);
        case 'inventory':
          return reportsApi.inventory();
        case 'debts':
          return reportsApi.debts();
        case 'products':
          return reportsApi.products();
        case 'suppliers':
          return reportsApi.suppliers();
        case 'expenses':
          return reportsApi.expenses(dateParams);
        case 'returns':
          return reportsApi.returns(dateParams);
        case 'used-items':
          return inventoryApi.list({ condition: 'USED' });
        default:
          return reportsApi.sales(dateParams);
      }
    },
  });

  return {
    reportData,
    reportLoading,
    refetch,
  };
}
