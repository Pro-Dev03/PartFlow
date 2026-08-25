import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { StatCard } from '../../../components/ui/stat-card';
import { Button } from '../../../components/ui/button';
import { getButtonSize } from '../../../config/button-sizes';
import { 
  Package, 
  Sparkles, 
  TrendingUp, 
  Clock, 
  AlertTriangle, 
  Layers 
} from 'lucide-react';
import { InventoryItem } from '../types/inventory.types';

interface InventoryStatsProps {
  inventoryItems: InventoryItem[];
  onRecommendationClick: (action: string) => void;
  isMobile: boolean;
}

export function InventoryStats({ inventoryItems, onRecommendationClick, isMobile }: InventoryStatsProps) {
  // Calculate real statistics from actual data
  const usedItemsCount = inventoryItems.filter((item: InventoryItem) => item.condition === 'USED').length;
  const newItemsCount = inventoryItems.filter((item: InventoryItem) => item.condition === 'NEW').length;
  const lowStockItems = inventoryItems.filter((item: InventoryItem) => item.stock < 10).length;
  
  // Calculate total inventory value
  const totalInventoryValue = inventoryItems.reduce((total, item) => {
    const price = (item.selling_price || item.price || 0) / 100;
    const stock = item.stock || 1;
    return total + (price * stock);
  }, 0);
  
  // Format the value
  const formattedValue = `₪${totalInventoryValue.toLocaleString('en-US')}`;

  return (
    <>
      {/* AI Inventory Insight */}
      <Card variant="ai">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Sparkles className="w-5 h-5 text-primary" />
            AI Inventory Insight
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-3.5">
            <div className="w-5 h-5 rounded-lg flex items-center justify-center bg-cyan/10 flex-shrink-0">
              <TrendingUp className="w-3 h-3 text-primary" />
            </div>
            <div>
              <p style={{ fontSize: '13px', fontWeight: '600', color: 'var(--text-primary)' }}>قيد التطوير</p>
              <p style={{ fontSize: '11px', color: 'var(--text-secondary)', marginTop: '4px' }}>
                هذه الميزة قيد التطوير حالياً. ستوفر تحليلات ذكية للمخزون وتوصيات لتحسين إدارة القطع والطلبات.
              </p>
              <Button 
                variant="secondary" 
                size={getButtonSize('inventory', 'recommendation')} 
                disabled
              >
                قيد التطوير
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Stats Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : 'repeat(auto-fit, minmax(250px, 1fr))', gap: '14px' }}
           className="grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard 
          title="إجمالي العناصر" 
          value={inventoryItems.length} 
          icon={Package}
          subtitle="إجمالي العناصر"
          variant="featured"
        />
        <StatCard 
          title="قيمة المخزون" 
          value={formattedValue} 
          icon={Package}
          subtitle="قيمة المخزون"
          variant="default"
        />
        <StatCard 
          title="يحتاج طلب" 
          value={lowStockItems} 
          icon={AlertTriangle}
          subtitle="منخفض المخزون"
          variant={lowStockItems > 0 ? 'warning' : 'success'}
        />
        <StatCard 
          title="قطع مستعملة" 
          value={usedItemsCount} 
          icon={Layers}
          subtitle="متاحة للبيع"
          variant="info"
        />
      </div>
    </>
  );
}