import { useQuery } from '@tanstack/react-query';
import { reportsApi, inventoryApi, productsApi, acquisitionsApi, customersApi } from '../../../services/api/endpoints';
import { addStoreDays, getStoreDateKey } from '../../../utils/store-time';

function getDateRangeParams(dateRange: string) {
  const storeDate = (date: Date) => getStoreDateKey(date) || '';
  const today = storeDate(new Date());
  const shiftStoreDate = (date: string, days: number) => addStoreDays(date, days);

  switch (dateRange) {
    case 'today':
      return { start_date: today, end_date: shiftStoreDate(today, 1) };
    case 'thisWeek': {
      const todayDate = new Date(`${today}T12:00:00Z`);
      const weekStart = shiftStoreDate(today, -todayDate.getUTCDay());
      return { start_date: weekStart, end_date: shiftStoreDate(today, 1) };
    }
    case 'thisMonth':
      return { start_date: `${today.slice(0, 7)}-01`, end_date: shiftStoreDate(today, 1) };
    case 'thisYear':
      return { start_date: `${today.slice(0, 4)}-01-01`, end_date: shiftStoreDate(today, 1) };
    default:
      return {};
  }
}

export function useReports(selectedReport: string, dateRange: string, customStartDate?: string, customEndDate?: string) {
  // Calculate date range based on selection
  const getSelectedDateRangeParams = () => {
    if (dateRange !== 'custom') return getDateRangeParams(dateRange);
    const shiftStoreDate = (date: string, days: number) => addStoreDays(date, days);
    switch (dateRange) {
      case 'custom':
        return customStartDate && customEndDate
          ? { start_date: customStartDate, end_date: shiftStoreDate(customEndDate, 1) }
          : {};
      default:
        return {};
    }
  };

  const dateParams = getSelectedDateRangeParams();

  const { data: reportData, isLoading: reportLoading, isError: reportError, error: reportRequestError, refetch } = useQuery({
    queryKey: ['reports', selectedReport, dateRange, customStartDate, customEndDate],
    enabled: dateRange !== 'custom' || Boolean(customStartDate && customEndDate),
    queryFn: async () => {
      switch (selectedReport) {
        case 'sales-profit': {
          const [salesResponse, profitResponse, netSalesResponse] = await Promise.all([
            reportsApi.sales(dateParams),
            reportsApi.profit(dateParams),
            reportsApi.netSales(dateParams),
          ]);
          const sales = salesResponse?.data ?? salesResponse ?? {};
          const profit = profitResponse?.data ?? profitResponse ?? {};
          const netSales = netSalesResponse?.data ?? netSalesResponse ?? {};
          const salesDays = Array.isArray(sales.by_day) ? sales.by_day : [];
          const profitDays = Array.isArray(profit.by_day) ? profit.by_day : [];
          const dayMap = new Map<string, { date: string; revenue: number; net_profit: number }>();
          salesDays.forEach((day: any) => {
            const date = String(day.date || '').slice(0, 10);
            dayMap.set(date, { date, revenue: Number(day.revenue || 0), net_profit: 0 });
          });
          profitDays.forEach((day: any) => {
            const date = String(day.date || '').slice(0, 10);
            const current = dayMap.get(date) || { date, revenue: 0, net_profit: 0 };
            current.revenue = Number(day.revenue || 0);
            current.net_profit = Number(day.net_profit || 0);
            dayMap.set(date, current);
          });
          return {
            ...sales,
            ...profit,
            ...netSales,
            by_day: Array.from(dayMap.values()).sort((left, right) => left.date.localeCompare(right.date)),
            sales_report: sales,
            profit_report: profit,
            net_sales_report: netSales,
          };
        }
        case 'purchases-suppliers': {
          const [purchasesResponse, suppliersResponse] = await Promise.all([
            reportsApi.purchases(dateParams),
            reportsApi.suppliers(),
          ]);
          const purchases = purchasesResponse?.data ?? purchasesResponse ?? {};
          const suppliers = suppliersResponse?.data ?? suppliersResponse ?? {};
          return { ...purchases, ...suppliers, purchases_report: purchases, suppliers_report: suppliers };
        }
        case 'sales':
          return reportsApi.sales(dateParams);
        case 'net-sales':
          return reportsApi.netSales(dateParams);
        case 'tax':
          return reportsApi.tax(dateParams);
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
        case 'purchases':
          return reportsApi.purchases(dateParams);
        case 'expenses':
          return reportsApi.expenses(dateParams);
        case 'returns':
          return reportsApi.returns(dateParams);
        case 'returns-analysis':
          return reportsApi.returnsAnalysis();
        default:
          return reportsApi.sales(dateParams);
      }
    },
  });

  return {
    reportData,
    reportLoading,
    reportError,
    reportRequestError,
    refetch,
  };
}
