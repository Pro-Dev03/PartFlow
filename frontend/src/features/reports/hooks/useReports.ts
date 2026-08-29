import { useQuery } from '@tanstack/react-query';
import { reportsApi, inventoryApi, productsApi } from '../../../services/api/endpoints';

export function useReports(selectedReport: string, dateRange: string, customStartDate?: string, customEndDate?: string) {
  // Calculate date range based on selection
  const getDateRangeParams = () => {
    const now = new Date();
    const today = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate()));
    const formatDate = (date: Date) => date.toISOString().split('T')[0];
    
    switch (dateRange) {
      case 'today':
        return {
          start_date: today.toISOString().split('T')[0],
          end_date: formatDate(today)
        };
      case 'thisWeek':
        const weekStart = new Date(today);
        weekStart.setUTCDate(today.getUTCDate() - today.getUTCDay());
        return {
          start_date: formatDate(weekStart),
          end_date: formatDate(today)
        };
      case 'thisMonth':
        const monthStart = new Date(Date.UTC(today.getUTCFullYear(), today.getUTCMonth(), 1));
        return {
          start_date: formatDate(monthStart),
          end_date: formatDate(today)
        };
      case 'thisYear':
        const yearStart = new Date(Date.UTC(today.getUTCFullYear(), 0, 1));
        return {
          start_date: formatDate(yearStart),
          end_date: formatDate(today)
        };
      case 'custom':
        return customStartDate && customEndDate
          ? { start_date: customStartDate, end_date: customEndDate }
          : {};
      default:
        return {};
    }
  };

  const dateParams = getDateRangeParams();

  const { data: reportData, isLoading: reportLoading, isError: reportError, error: reportRequestError, refetch } = useQuery({
    queryKey: ['reports', selectedReport, dateRange, customStartDate, customEndDate],
    enabled: dateRange !== 'custom' || Boolean(customStartDate && customEndDate),
    queryFn: async () => {
      switch (selectedReport) {
        case 'sales':
          return reportsApi.sales(dateParams);
        case 'net-sales':
          return reportsApi.netSales(dateParams);
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
        case 'used-items': {
          const [inventoryResponse, productsResponse] = await Promise.all([
            inventoryApi.list({ condition: 'USED', page: 1, per_page: 100 }),
            productsApi.list({ page: 1, per_page: 100 }),
          ]);
          const products = (productsResponse.data?.products || []) as Array<{ id: string; name: string }>;
          const productNames = new Map(products.map(product => [product.id, product.name]));
          const items = inventoryResponse.data?.items || [];
          return {
            ...inventoryResponse,
            data: {
              ...inventoryResponse.data,
              items: items.map(item => ({
                ...item,
                product_name: productNames.get(item.product_id) || 'منتج غير معروف',
              })),
            },
          };
        }
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
