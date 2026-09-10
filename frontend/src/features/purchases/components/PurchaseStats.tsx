import { StatCard } from '../../../components/ui/stat-card';
import { ShoppingCart, Calendar, Package, RotateCcw, Truck, CreditCard, Receipt } from 'lucide-react';
import { PurchaseStats } from '../types/purchases.types';

interface PurchaseStatsProps {
  stats: PurchaseStats;
}

export function PurchaseStats({ stats }: PurchaseStatsProps) {
  return (
    <div
      className="grid gap-2"
      style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(150px, 1fr))' }}
    >
      <StatCard
        title="إجمالي المشتريات"
        value={stats.totalPurchases}
        icon={ShoppingCart}
        variant="featured"
        compact
      />
      <StatCard
        title="قيد الانتظار"
        value={stats.pendingCount}
        icon={Calendar}
        variant="warning"
        compact
      />
      <StatCard
        title="تم الاستلام"
        value={stats.receivedCount}
        icon={Package}
        variant="success"
        compact
      />
      <StatCard
        title="تم العكس"
        value={stats.reversedCount}
        icon={RotateCcw}
        variant="destructive"
        compact
      />
      <StatCard
        title="تكلفة غير المستلمة"
        value={`₪${stats.pendingCost.toLocaleString()}`}
        icon={Truck}
        variant="default"
        compact
      />
      <StatCard
        title="تكلفة المستلمة"
        value={`₪${stats.receivedCost.toLocaleString()}`}
        icon={Package}
        variant="success"
        compact
      />
      <StatCard
        title="المتبقي غير المدفوع"
        value={`₪${stats.outstandingAmount.toLocaleString()}`}
        icon={CreditCard}
        variant="warning"
        compact
      />
      <StatCard
        title="فواتير بلا ضريبة"
        value={`₪${stats.untaxedCost.toLocaleString()}`}
        icon={Receipt}
        subtitle={`${stats.untaxedCount.toLocaleString()} فواتير معفاة`}
        variant="default"
        compact
      />
      <StatCard
        title="فواتير خاضعة للضريبة"
        value={`₪${stats.taxedCost.toLocaleString()}`}
        icon={Receipt}
        subtitle={`${stats.taxedCount.toLocaleString()} فواتير، شامل الضريبة`}
        variant="info"
        compact
      />
    </div>
  );
}