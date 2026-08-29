import { StatCard } from '../../../components/ui/stat-card';
import { ShoppingCart, Calendar, Package, RotateCcw, Truck, CreditCard } from 'lucide-react';
import { PurchaseStats } from '../types/purchases.types';

interface PurchaseStatsProps {
  stats: PurchaseStats;
}

export function PurchaseStats({ stats }: PurchaseStatsProps) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 xl:grid-cols-7 gap-md">
      <StatCard
        title="إجمالي المشتريات"
        value={stats.totalPurchases}
        icon={ShoppingCart}
        variant="featured"
      />
      <StatCard
        title="قيد الانتظار"
        value={stats.pendingCount}
        icon={Calendar}
        variant="warning"
      />
      <StatCard
        title="تم الاستلام"
        value={stats.receivedCount}
        icon={Package}
        variant="success"
      />
      <StatCard
        title="تم العكس"
        value={stats.reversedCount}
        icon={RotateCcw}
        variant="destructive"
      />
      <StatCard
        title="تكلفة غير المستلمة"
        value={`₪${stats.pendingCost.toLocaleString()}`}
        icon={Truck}
        variant="default"
      />
      <StatCard
        title="تكلفة المستلمة"
        value={`₪${stats.receivedCost.toLocaleString()}`}
        icon={Package}
        variant="success"
      />
      <StatCard
        title="المتبقي غير المدفوع"
        value={`₪${stats.outstandingAmount.toLocaleString()}`}
        icon={CreditCard}
        variant="warning"
      />
    </div>
  );
}