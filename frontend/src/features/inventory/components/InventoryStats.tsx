import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { StatCard } from '../../../components/ui/stat-card';
import { Button } from '../../../components/ui/button';
import { getButtonSize } from '../../../config/button-sizes';
import { normalizeCurrencyValue } from '../../../utils';
import { 
  Package, 
  Sparkles, 
  TrendingUp, 
  AlertTriangle 
} from 'lucide-react';
import { InventoryItem, Product } from '../types/inventory.types';

interface InventoryStatsProps {
  products: Product[];
  inventoryItems: InventoryItem[];
  onRecommendationClick: (action: string) => void;
  isMobile: boolean;
}

export function InventoryStats({ products, inventoryItems, isMobile }: InventoryStatsProps) {
  const inactiveStatuses = new Set(['SOLD', 'RETURNED', 'REVERSED', 'CANCELLED', 'DELETED', 'VOID']);

  const normalizedItems = inventoryItems.reduce((acc: Map<string, { stock: number; unitPrice: number; condition: string }>, item: InventoryItem) => {
    const status = String((item as any).status || '').trim().toUpperCase();
    const productId = String((item as any).product_id || (item as any).product?.id || '').trim();

    if (inactiveStatuses.has(status) || status !== 'AVAILABLE' || !productId) {
      return acc;
    }

    if (acc.has(productId)) {
      return acc;
    }

    const explicitStock = Number((item as any).available_quantity ?? (item as any).current_quantity ?? (item as any).stock ?? (item as any).quantity ?? 0);
    const fallbackStock = status === 'AVAILABLE' ? 1 : 0;
    const stock = Number.isFinite(explicitStock) && explicitStock > 0 ? explicitStock : fallbackStock;

    if (stock <= 0) {
      return acc;
    }

    const nextCondition = String(item.condition || '').toUpperCase();
    const unitPrice = normalizeCurrencyValue((item as any).selling_price ?? item.price ?? (item as any).purchase_cost ?? 0);

    acc.set(productId, {
      stock,
      unitPrice,
      condition: nextCondition,
    });

    return acc;
  }, new Map());

  const visibleProductIds = new Set(products.map((product) => product.id));
  const summaryItems = Array.from(normalizedItems.entries())
    .filter(([productId, item]) => visibleProductIds.has(productId) && item.condition !== 'USED')
    .map(([, item]) => item);

  const lowStockItems = summaryItems.filter((item) => item.stock > 0 && item.stock < 10).length;

  const totalInventoryValue = summaryItems.reduce((total, item) => total + (item.stock * item.unitPrice), 0);
  const formattedValue = `₪${Math.round(totalInventoryValue).toLocaleString('en-US')}`;

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
          value={products.length}
          icon={Package}
          subtitle="إجمالي العناصر"
          variant="featured"
        />
        <StatCard 
          title="إجمالي قيمة البيع للمخزون المتاح"
          value={formattedValue} 
          icon={Package}
          subtitle="المنتجات الجديدة الجاهزة للبيع"
          variant="default"
        />
        <StatCard 
          title="يحتاج طلب" 
          value={lowStockItems} 
          icon={AlertTriangle}
          subtitle="منخفض المخزون"
          variant={lowStockItems > 0 ? 'warning' : 'success'}
        />
      </div>
    </>
  );
}