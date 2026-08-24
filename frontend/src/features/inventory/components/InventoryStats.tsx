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
  const usedItemsCount = inventoryItems.filter((item: InventoryItem) => item.condition === 'USED').length;

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
              <p style={{ fontSize: '13px', fontWeight: '600', color: 'var(--text-primary)' }}>فرصة شراء معالجات Intel</p>
              <p style={{ fontSize: '11px', color: 'var(--text-secondary)', marginTop: '4px' }}>
                الأسعار الحالية أقل من المتوسط بنسبة 15%. هناك طلب متزايد من 3 عملاء رئيسيين.
              </p>
              <Button 
                variant="secondary" 
                size={getButtonSize('inventory', 'recommendation')} 
                onClick={() => onRecommendationClick('search_intel')}
              >
                عرض التوصية ←
              </Button>
            </div>
          </div>
          <div className="flex gap-3.5">
            <div className="w-5 h-5 rounded-lg flex items-center justify-center bg-warning/10 flex-shrink-0">
              <Clock className="w-3 h-3 text-warning" />
            </div>
            <div>
              <p style={{ fontSize: '13px', fontWeight: '600', color: 'var(--text-primary)' }}>تنبيه انخفاض المخزون</p>
              <p style={{ fontSize: '11px', color: 'var(--text-secondary)', marginTop: '4px' }}>
                5 منتجات وصلت للحد الأدنى. كروت الشاشة Samsung 27" الأكثر طلباً تبقى 3 قطع فقط.
              </p>
              <Button 
                variant="secondary" 
                size={getButtonSize('inventory', 'recommendation')} 
                onClick={() => onRecommendationClick('filter_used')}
              >
                عرض المنتجات ←
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
          trend="+8.3%"
          trendUp={true}
        />
        <StatCard 
          title="قيمة المخزون" 
          value="₪185,400" 
          icon={Package}
          subtitle="قيمة المخزون"
          variant="default"
          trend="+12.1%"
          trendUp={true}
        />
        <StatCard 
          title="يحتاج طلب" 
          value="12" 
          icon={AlertTriangle}
          subtitle="يحتاج طلب"
          variant="warning"
          trend="-2"
          trendUp={false}
        />
        <StatCard 
          title="قطع مستعملة" 
          value={usedItemsCount} 
          icon={Layers}
          subtitle="متاحة للبيع"
          variant="info"
          trend={usedItemsCount > 0 ? "+" + usedItemsCount.toString() : "0"}
          trendUp={true}
        />
      </div>
    </>
  );
}