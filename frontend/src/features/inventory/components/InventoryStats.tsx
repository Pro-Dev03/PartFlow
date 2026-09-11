import { StatCard } from '../../../components/ui/stat-card';
import { Button } from '../../../components/ui/button';
import { useQuery } from '@tanstack/react-query';
import { settingsApi } from '../../../services/api/endpoints';
import { getButtonSize } from '../../../config/button-sizes';
import { normalizeCurrencyValue } from '../../../utils';
import { 
  Package, 
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
  const activeStatuses = new Set(['AVAILABLE']);
  const inactiveStatuses = new Set(['SOLD', 'REVERSED', 'CANCELLED', 'DELETED', 'VOID']);
  const { data: taxSetting } = useQuery({
    queryKey: ['settings', 'tax_rate'],
    queryFn: () => settingsApi.getSetting('tax_rate'),
    retry: false,
  });

  const normalizedItems = inventoryItems.reduce((acc: Map<string, { stock: number; unitPrice: number; condition: string }>, item: InventoryItem) => {
    const status = String((item as any).status || '').trim().toUpperCase();
    const productId = String((item as any).product_id || (item as any).product?.id || '').trim();

    if (inactiveStatuses.has(status) || !activeStatuses.has(status) || !productId) {
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
    .map(([productId, item]) => ({
      ...item,
      minimumStockLevel: Math.max(1, Number(products.find((product) => product.id === productId)?.min_stock_level) || 3),
    }));

  const lowStockItems = summaryItems.filter((item) => item.stock > 0 && item.stock <= item.minimumStockLevel).length;

  const totalInventoryValue = summaryItems.reduce((total, item) => total + (item.stock * item.unitPrice), 0);
  const formattedValue = `₪${Math.round(totalInventoryValue).toLocaleString('en-US')}`;
  const configuredTaxRate = Number(taxSetting?.data?.value);
  const taxRate = Number.isFinite(configuredTaxRate) && configuredTaxRate >= 0 ? configuredTaxRate : 0;
  const totalInventoryValueWithTax = totalInventoryValue * (1 + taxRate / 100);
  const formattedValueWithTax = `₪${totalInventoryValueWithTax.toFixed(2)}`;

  return (
    <>
      {/* AI Inventory Insight */}
      <div className="premium-insight">
        <div className="premium-insight-icon"><TrendingUp className="h-3.5 w-3.5" /></div>
        <div className="premium-insight-copy">
          <p className="premium-insight-title">AI Inventory Insight · قيد التطوير</p>
          <p className="premium-insight-text">ستوفر تحليلات ذكية للمخزون وتوصيات لتحسين إدارة القطع والطلبات.</p>
        </div>
        <Button variant="secondary" size={getButtonSize('inventory', 'recommendation')} disabled className="premium-insight-action">قيد التطوير</Button>
      </div>

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
          title="إجمالي قيمة البيع قبل الضريبة"
          value={formattedValue} 
          icon={Package}
          subtitle={`شامل الضريبة ${taxRate}%: ${formattedValueWithTax}`}
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