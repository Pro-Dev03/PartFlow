import { useQuery } from '@tanstack/react-query';
import { reportsApi, inventoryApi, productsApi, acquisitionsApi, customersApi } from '../../../services/api/endpoints';
import { addStoreDays, getStoreDateKey } from '../../../utils/store-time';

export function useReports(selectedReport: string, dateRange: string, customStartDate?: string, customEndDate?: string) {
  const storeDate = (date: Date) => getStoreDateKey(date) || '';

  // Calculate date range based on selection
  const getDateRangeParams = () => {
    const now = new Date();
    const today = storeDate(now);
    const shiftStoreDate = (date: string, days: number) => {
      return addStoreDays(date, days);
    };
    
    switch (dateRange) {
      case 'today':
        return {
          start_date: today,
          end_date: shiftStoreDate(today, 1)
        };
      case 'thisWeek':
        const todayDate = new Date(`${today}T12:00:00Z`);
        const weekStart = shiftStoreDate(today, -todayDate.getUTCDay());
        return {
          start_date: weekStart,
          end_date: shiftStoreDate(today, 1)
        };
      case 'thisMonth':
        const monthStart = `${today.slice(0, 7)}-01`;
        return {
          start_date: monthStart,
          end_date: shiftStoreDate(today, 1)
        };
      case 'thisYear':
        const yearStart = `${today.slice(0, 4)}-01-01`;
        return {
          start_date: yearStart,
          end_date: shiftStoreDate(today, 1)
        };
      case 'custom':
        return customStartDate && customEndDate
          ? { start_date: customStartDate, end_date: shiftStoreDate(customEndDate, 1) }
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
          const collectPages = async (
            fetchPage: (page: number) => Promise<any>,
            collectionKey: string,
          ) => {
            const rows: any[] = [];
            let firstResponse: any;
            for (let page = 1; page <= 100; page += 1) {
              const response = await fetchPage(page);
              firstResponse ??= response;
              const payload = response?.data;
              const pageRows = Array.isArray(payload) ? payload : (payload?.[collectionKey] || []);
              rows.push(...pageRows);
              const metadata = response?.meta || payload?.meta;
              if (pageRows.length === 0 || pageRows.length < 100 || (metadata?.total_pages && page >= metadata.total_pages)) {
                break;
              }
            }
            return { response: firstResponse, rows };
          };
          const [inventoryResult, productsResult, acquisitionsResult, customersResult] = await Promise.all([
            collectPages((page) => inventoryApi.list({ condition: 'USED', page, per_page: 100 }), 'items'),
            collectPages((page) => productsApi.list({ page, per_page: 100 }), 'products'),
            collectPages((page) => acquisitionsApi.list({ type: 'CUSTOMER', page, per_page: 100 }), 'acquisitions'),
            collectPages((page) => customersApi.list({ page, per_page: 100 }), 'customers'),
          ]);
          const inventoryResponse = inventoryResult.response;
          const products = Array.from(new Map(
            productsResult.rows.map((product: { id: string; name: string }) => [product.id, product]),
          ).values()) as Array<{ id: string; name: string }>;
          const productNames = new Map(products.map(product => [product.id, product.name]));
          const customers = Array.from(new Map(
            customersResult.rows.map((customer: any) => [customer.id, customer]),
          ).values());
          const customerNames = new Map(customers.map((customer: any) => [customer.id, customer.name]));
          const sellerByInventoryItemId = new Map<string, string>();
          const acquisitions = acquisitionsResult.rows;
          acquisitions.forEach((acquisition: any) => {
            const sellerName = customerNames.get(acquisition.customer_id) || 'بائع غير معروف';
            (acquisition.items || []).forEach((acquisitionItem: any) => {
              if (acquisitionItem.inventory_item_id) {
                sellerByInventoryItemId.set(acquisitionItem.inventory_item_id, sellerName);
              }
            });
          });
          const items = Array.from(new Map(
            inventoryResult.rows
              .filter(item => String(item.status || '').toUpperCase() !== 'ARCHIVED')
              .map(item => [item.id, item]),
          ).values());
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
