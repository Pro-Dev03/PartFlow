import { useQuery } from '@tanstack/react-query';
import { reportsApi, inventoryApi, productsApi, acquisitionsApi, customersApi } from '../../../services/api/endpoints';

export function useReports(selectedReport: string, dateRange: string, customStartDate?: string, customEndDate?: string) {
  // Calculate date range based on selection
  const getDateRangeParams = () => {
    const now = new Date();
    const today = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate()));
    const formatDate = (date: Date) => date.toISOString().split('T')[0];
    const formatExclusiveEndDate = (date: Date) => {
      const exclusiveEnd = new Date(date);
      exclusiveEnd.setUTCDate(exclusiveEnd.getUTCDate() + 1);
      return formatDate(exclusiveEnd);
    };
    
    switch (dateRange) {
      case 'today':
        return {
          start_date: formatDate(today),
          end_date: formatExclusiveEndDate(today)
        };
      case 'thisWeek':
        const weekStart = new Date(today);
        weekStart.setUTCDate(today.getUTCDate() - today.getUTCDay());
        return {
          start_date: formatDate(weekStart),
          end_date: formatExclusiveEndDate(today)
        };
      case 'thisMonth':
        const monthStart = new Date(Date.UTC(today.getUTCFullYear(), today.getUTCMonth(), 1));
        return {
          start_date: formatDate(monthStart),
          end_date: formatExclusiveEndDate(today)
        };
      case 'thisYear':
        const yearStart = new Date(Date.UTC(today.getUTCFullYear(), 0, 1));
        return {
          start_date: formatDate(yearStart),
          end_date: formatExclusiveEndDate(today)
        };
      case 'custom':
        return customStartDate && customEndDate
          ? { start_date: customStartDate, end_date: formatExclusiveEndDate(new Date(`${customEndDate}T00:00:00Z`)) }
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
        case 'used-items': {
          const [inventoryResponse, productsResponse, acquisitionsResponse, customersResponse] = await Promise.all([
            inventoryApi.list({ condition: 'USED', page: 1, per_page: 100 }),
            productsApi.list({ page: 1, per_page: 100 }),
            acquisitionsApi.list({ type: 'CUSTOMER', page: 1, per_page: 100 }),
            customersApi.list({ page: 1, per_page: 100 }),
          ]);
          const products = (productsResponse.data?.products || []) as Array<{ id: string; name: string }>;
          const productNames = new Map(products.map(product => [product.id, product.name]));
          const customers = Array.isArray(customersResponse.data) ? customersResponse.data : [];
          const customerNames = new Map(customers.map((customer: any) => [customer.id, customer.name]));
          const sellerByInventoryItemId = new Map<string, string>();
          const acquisitions = Array.isArray(acquisitionsResponse.data) ? acquisitionsResponse.data : [];
          acquisitions.forEach((acquisition: any) => {
            const sellerName = customerNames.get(acquisition.customer_id) || 'بائع غير معروف';
            (acquisition.items || []).forEach((acquisitionItem: any) => {
              if (acquisitionItem.inventory_item_id) {
                sellerByInventoryItemId.set(acquisitionItem.inventory_item_id, sellerName);
              }
            });
          });
          const items = (inventoryResponse.data?.items || []).filter(
            item => String(item.status || '').toUpperCase() !== 'ARCHIVED'
          );
          return {
            ...inventoryResponse,
            data: {
              ...inventoryResponse.data,
              items: items.map(item => ({
                ...item,
                product_name: productNames.get(item.product_id) || 'منتج غير معروف',
                seller_name: sellerByInventoryItemId.get(item.id) || 'غير محدد',
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
