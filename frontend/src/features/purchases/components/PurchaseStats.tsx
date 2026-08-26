import { StatCard } from '../../../components/ui/stat-card';
import { ShoppingCart, Calendar, Package, RotateCcw, Truck } from 'lucide-react';
import { PurchaseStats } from '../types/purchases.types';

interface PurchaseStatsProps {
  stats: PurchaseStats;
}

export function PurchaseStats({ stats }: PurchaseStatsProps) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-md">
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
        title="إجمالي التكلفة"
        value={`₪${stats.totalCost.toLocaleString()}`}
        icon={Truck}
        variant="default"
      />
    </div>
  );
}