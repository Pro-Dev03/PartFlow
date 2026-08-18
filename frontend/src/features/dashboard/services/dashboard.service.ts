import { DashboardStats, Alert, QuickAction, RecentActivity } from '../types/dashboard.types'

// Mock service - replace with actual API calls
export const dashboardService = {
  async getStats(): Promise<DashboardStats> {
    // Replace with actual API call
    return {
      todaySales: 7450,
      todayProfit: 1850,
      inventoryValue: 185400,
      outstandingDebts: 24800,
      lowStockCount: 4,
      overdueDebtsCount: 3,
      todaySalesCount: 18,
      newCustomersCount: 3,
    }
  },

  async getAlerts(): Promise<Alert[]> {
    // Replace with actual API call
    return [
      {
        id: '1',
        type: 'low-stock',
        severity: 'high',
        title: 'RTX 3060 وصل إلى الحد الأدنى',
        description: 'المخزون الحالي: 2 قطع',
        actionLabel: 'عرض القطعة',
        actionLink: '/inventory/rtx-3060',
        createdAt: new Date(),
      },
      {
        id: '2',
        type: 'overdue-debt',
        severity: 'critical',
        title: 'أحمد لديه دين متأخر',
        description: '₪1,250 متأخر 14 يوم',
        actionLabel: 'فتح العميل',
        actionLink: '/customers/ahmed',
        createdAt: new Date(),
      },
      {
        id: '3',
        type: 'inspection-needed',
        severity: 'medium',
        title: 'قطعة مستعملة تحتاج فحص',
        description: 'RTX 3070 - تم الشراء قبل 3 أيام',
        actionLabel: 'بدء الفحص',
        actionLink: '/inventory/rtx-3070/inspection',
        createdAt: new Date(),
      },
    ]
  },

  async getQuickActions(): Promise<QuickAction[]> {
    return [
      { id: '1', label: 'بيع جديد', icon: '💰', path: '/sales/new', color: 'primary' },
      { id: '2', label: 'مسح', icon: '📱', path: '/barcode', color: 'success' },
      { id: '3', label: 'إضافة قطعة', icon: '📦', path: '/inventory/new', color: 'primary' },
      { id: '4', label: 'إضافة عميل', icon: '👤', path: '/customers/new', color: 'primary' },
      { id: '5', label: 'تسجيل دفعة', icon: '💳', path: '/debts/payment', color: 'warning' },
    ]
  },

  async getRecentActivity(): Promise<RecentActivity[]> {
    return [
      {
        id: '1',
        type: 'sale',
        title: 'بيع RTX 4070',
        description: 'تم البيع لأحمد',
        amount: 2350,
        timestamp: new Date(Date.now() - 1000 * 60 * 30), // 30 min ago
        user: 'محمد',
      },
      {
        id: '2',
        type: 'payment',
        title: 'دفعة من سارة',
        description: 'دفعة على الدين',
        amount: 500,
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 2), // 2 hours ago
        user: 'علي',
      },
      {
        id: '3',
        type: 'purchase',
        title: 'شراء RAM 16GB',
        description: 'تم الشراء من المورد A',
        amount: 1800,
        timestamp: new Date(Date.now() - 1000 * 60 * 60 * 5), // 5 hours ago
        user: 'خالد',
      },
    ]
  },
}