import { useEffect, useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { dashboardApi } from '../../services/api/endpoints';
import { useAuthStore } from '../../stores/authStore';
import { customersApi } from '../../services/api/endpoints';
import FloatingAIButton from './floating-ai-button';
import AIChatModern from './ai-chat-modern';

const STORAGE_KEY = 'partflow-assistant-cache';

function readOfflineSnapshot() {
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

function writeOfflineSnapshot(data: Record<string, unknown>) {
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify({
      updatedAt: new Date().toISOString(),
      ...data,
    }));
  } catch {
    // ignore storage failures in private mode or restricted browser environments
  }
}

export default function AIAssistantWrapper() {
  const { isAuthenticated, token } = useAuthStore();
  const isAuthReady = Boolean(isAuthenticated && token);
  const [isChatOpen, setIsChatOpen] = useState(false);
  const [chatAnchor, setChatAnchor] = useState({ x: 20, y: 20 });
  const [isOnline, setIsOnline] = useState(typeof navigator !== 'undefined' ? navigator.onLine : true);

  useEffect(() => {
    if (!isAuthReady) {
      return;
    }

    const handleOnline = () => setIsOnline(true);
    const handleOffline = () => setIsOnline(false);

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, [isAuthReady]);

  const { data: dashboardData } = useQuery({
    queryKey: ['assistant-dashboard'],
    queryFn: () => dashboardApi.getStats(),
    staleTime: 60000,
    enabled: isAuthReady && isOnline,
    retry: 1,
    onSuccess: (response) => {
      const stats = response?.data as {
        todaySales?: number;
        todayProfit?: number;
        overdueDebts?: number;
        lowStock?: number;
      } | undefined;

      writeOfflineSnapshot({
        todaySales: Number(stats?.todaySales ?? 0),
        todayProfit: Number(stats?.todayProfit ?? 0),
        overdueDebts: Number(stats?.overdueDebts ?? 0),
        lowStock: Number(stats?.lowStock ?? 0),
      });
    },
  });

  const { data: lowStockData } = useQuery({
    queryKey: ['assistant-low-stock'],
    queryFn: () => dashboardApi.getLowStockItems(),
    staleTime: 60000,
    enabled: isAuthReady && isOnline,
    retry: 1,
    onSuccess: (response) => {
      writeOfflineSnapshot({
        lowStockItems: Array.isArray(response?.data) ? response.data : [],
      });
    },
  });

  const { data: overdueDebtsData } = useQuery({
    queryKey: ['assistant-overdue-debts'],
    queryFn: () => dashboardApi.getOverdueDebts(),
    staleTime: 60000,
    enabled: isAuthReady && isOnline,
    retry: 1,
    onSuccess: (response) => {
      writeOfflineSnapshot({
        overdueDebtsList: Array.isArray(response?.data) ? response.data : [],
      });
    },
  });

  const { data: customersData } = useQuery({
    queryKey: ['assistant-customers-count'],
    queryFn: () => customersApi.list({ page: 1, per_page: 1 }),
    staleTime: 120000,
    enabled: isAuthReady && isOnline,
    retry: 1,
  });

  const { data: dailySalesData } = useQuery({
    queryKey: ['assistant-daily-sales'],
    queryFn: () => dashboardApi.getDailySalesSummary(),
    staleTime: 60000,
    enabled: isAuthReady && isOnline,
    retry: 1,
  });

  const assistantContext = useMemo(() => {
    const offlineSnapshot = readOfflineSnapshot();

    const stats = (dashboardData?.data as {
      todaySales?: number;
      todayProfit?: number;
      overdueDebts?: number;
      lowStock?: number;
    } | undefined) ?? {
      todaySales: offlineSnapshot?.todaySales ?? 0,
      todayProfit: offlineSnapshot?.todayProfit ?? 0,
      overdueDebts: offlineSnapshot?.overdueDebts ?? 0,
      lowStock: offlineSnapshot?.lowStock ?? 0,
    };

    const lowStockItems = Array.isArray(lowStockData?.data)
      ? lowStockData.data
      : Array.isArray(offlineSnapshot?.lowStockItems)
        ? offlineSnapshot.lowStockItems
        : [];

    const overdueDebts = Array.isArray(overdueDebtsData?.data)
      ? overdueDebtsData.data
      : Array.isArray(offlineSnapshot?.overdueDebtsList)
        ? offlineSnapshot.overdueDebtsList
        : [];

    const todaySales = Number(stats?.todaySales ?? 0);
    const yesterdaySales =
      typeof dailySalesData?.data === 'object' && dailySalesData.data !== null
        ? Number((dailySalesData.data as Record<string, unknown>).total_sales ?? (dailySalesData.data as Record<string, unknown>).sales ?? todaySales)
        : Number(stats?.todaySales ?? todaySales);

    const totalCustomers =
      (customersData as { data?: unknown[]; total?: number } | undefined)?.total ??
      (Array.isArray((customersData as { data?: unknown[] } | undefined)?.data)
        ? (customersData as { data?: unknown[] }).data!.length
        : offlineSnapshot?.totalCustomers ?? 0);

    const topSellingProducts = Array.isArray(
      (dailySalesData?.data as Record<string, unknown> | undefined)?.top_products
    )
      ? ((dailySalesData?.data as Record<string, unknown>).top_products as Array<Record<string, unknown>>).map((product) => ({
          name: String(product.name ?? product.product_name ?? product.product?.name ?? 'منتج'),
          quantity: Number(product.quantity ?? product.sold_quantity ?? product.count ?? 0),
          revenue: Number(product.revenue ?? product.total_revenue ?? product.sales ?? 0),
        }))
      : Array.isArray(offlineSnapshot?.topSellingProducts)
        ? offlineSnapshot.topSellingProducts
        : [];

    return {
      lowStockCount: Number(stats?.lowStock ?? lowStockItems.length ?? 0),
      overdueDebtsCount: Number(stats?.overdueDebts ?? overdueDebts.length ?? 0),
      salesToday: todaySales,
      salesYesterday: yesterdaySales > 0 ? yesterdaySales : todaySales,
      todayProfit: Number(stats?.todayProfit ?? 0),
      totalCustomers,
      totalProducts: Number(stats?.lowStock ?? 0),
      lowStockItems: lowStockItems.slice(0, 5).map((item: any) => ({
        name: item?.name || item?.productName || item?.product?.name || 'منتج',
        quantity: item?.quantity ?? item?.stock ?? item?.available_quantity ?? 0,
        minQuantity: item?.minQuantity ?? item?.minimum_quantity ?? 0,
        category: item?.categoryName || item?.category?.name || item?.partType?.name,
      })),
      overdueDebts: overdueDebts.slice(0, 5).map((item: any) => ({
        customerName: item?.customerName || item?.customer?.name || 'عميل',
        amount: item?.amount ?? item?.balance ?? item?.total ?? 0,
        days: item?.days ?? item?.daysOverdue ?? item?.overdueDays ?? 0,
        phone: item?.phone || item?.customer?.phone || item?.customer_phone,
      })),
      topSellingProducts: topSellingProducts.slice(0, 5),
    };
  }, [dashboardData, lowStockData, overdueDebtsData, dailySalesData, customersData]);

  const handleButtonClick = (position: { x: number; y: number }) => {
    setChatAnchor(position);
    setIsChatOpen((prev) => !prev);
  };

  const handleClose = () => {
    setIsChatOpen(false);
  };

  if (!isAuthReady) {
    return null;
  }

  return (
    <>
      <FloatingAIButton onClick={handleButtonClick} />
      {isChatOpen && (
        <AIChatModern
          onClose={handleClose}
          position={chatAnchor}
          embedded={false}
          isOffline={!isOnline}
          assistantContext={assistantContext}
        />
      )}
    </>
  );
}
