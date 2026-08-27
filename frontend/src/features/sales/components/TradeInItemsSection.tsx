import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Layers } from 'lucide-react';
import type { InventoryItem } from '../../../types/models';

interface PartTypeOption {
  id: string;
  name_ar?: string;
  color?: string;
}
interface TradeInItemsSectionProps {
  inventoryItems: InventoryItem[];
  partTypes: PartTypeOption[];
  onTradeInClick: (item: InventoryItem) => void;
}

export function TradeInItemsSection({
  inventoryItems,
  partTypes,
  onTradeInClick,
}: TradeInItemsSectionProps) {
  const availableItems = inventoryItems?.filter((item) =>
    item.condition === 'USED' && item.status === 'AVAILABLE'
  ) || [];

  return (
    <Card style={{
      background: 'linear-gradient(135deg, rgba(236, 72, 153, 0.05) 0%, var(--color-warning-05) 100%)',
      border: '1px solid rgba(236, 72, 153, 0.1)',
      backdropFilter: 'blur(10px)',
      transition: 'all 0.3s ease'
    }}
    onMouseEnter={(e) => {
      e.currentTarget.style.background = 'linear-gradient(135deg, rgba(236, 72, 153, 0.08) 0%, var(--color-warning-08) 100%)';
      e.currentTarget.style.borderColor = 'rgba(236, 72, 153, 0.2)';
      e.currentTarget.style.transform = 'translateY(-2px)';
      e.currentTarget.style.boxShadow = '0 8px 25px rgba(236, 72, 153, 0.1)';
    }}
    onMouseLeave={(e) => {
      e.currentTarget.style.background = 'linear-gradient(135deg, rgba(236, 72, 153, 0.05) 0%, var(--color-warning-05) 100%)';
      e.currentTarget.style.borderColor = 'rgba(236, 72, 153, 0.1)';
      e.currentTarget.style.transform = 'translateY(0)';
      e.currentTarget.style.boxShadow = 'none';
    }}>
      <CardHeader>
        <CardTitle style={{ 
          display: 'flex', 
          alignItems: 'center', 
          gap: '10px',
          fontSize: '15px',
          fontWeight: '600',
          color: 'var(--text-primary)'
        }}>
          <div style={{
            width: '36px',
            height: '36px',
            borderRadius: '10px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            background: 'linear-gradient(135deg, rgba(236, 72, 153, 0.15) 0%, var(--color-warning-15) 100%)',
            border: '1px solid rgba(236, 72, 153, 0.25)'
          }}>
            <Layers className="w-5 h-5 text-pink-400" />
          </div>
          قطع مستعملة متاحة
        </CardTitle>
      </CardHeader>
      <CardContent>
        {availableItems.length > 0 ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
            {availableItems.slice(0, 5).map((item) => {
              const partType = partTypes.find((pt) => pt.id === item.part_type_id);
              return (
                <div
                  key={item.id}
                  onClick={() => onTradeInClick(item)}
                  style={{
                    padding: '12px',
                    borderRadius: '10px',
                    border: '1px solid rgba(236, 72, 153, 0.2)',
                    background: partType ? `linear-gradient(135deg, ${partType.color}20 0%, transparent 100%)` : 'rgba(236, 72, 153, 0.08)',
                    cursor: 'pointer',
                    transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                    backdropFilter: 'blur(10px)'
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.background = partType ? `linear-gradient(135deg, ${partType.color}30 0%, transparent 100%)` : 'rgba(236, 72, 153, 0.15)';
                    e.currentTarget.style.borderColor = partType ? partType.color : 'rgba(236, 72, 153, 0.4)';
                    e.currentTarget.style.transform = 'translateY(-2px) scale(1.01)';
                    e.currentTarget.style.boxShadow = '0 8px 20px rgba(236, 72, 153, 0.15)';
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.background = partType ? `linear-gradient(135deg, ${partType.color}20 0%, transparent 100%)` : 'rgba(236, 72, 153, 0.08)';
                    e.currentTarget.style.borderColor = partType ? partType.color : 'rgba(236, 72, 153, 0.2)';
                    e.currentTarget.style.transform = 'translateY(0) scale(1)';
                    e.currentTarget.style.boxShadow = 'none';
                  }}
                >
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <div>
                      <div style={{
                        fontSize: '14px',
                        fontWeight: '600',
                        color: 'var(--text-primary)',
                        letterSpacing: '0.2px'
                      }}>
                        {item.product_name || item.product?.name}
                      </div>
                      {partType && (
                        <div style={{ 
                          fontSize: '11px', 
                          color: partType.color, 
                          marginTop: '3px',
                          fontWeight: '500'
                        }}>
                          {partType.name_ar}
                        </div>
                      )}
                      <div style={{
                        fontSize: '11px',
                        color: 'var(--text-secondary)',
                        marginTop: '2px',
                        fontWeight: '500'
                      }}>
                         اشتريت بـ: ₪{item.purchase_cost.toFixed(2)}
                      </div>
                    </div>
                    <div style={{
                      fontSize: '15px',
                      fontWeight: '700',
                      color: 'var(--color-info)',
                      textShadow: '0 0 20px var(--color-info-30)'
                    }}>
                       ₪{item.selling_price.toFixed(2)}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        ) : (
          <div style={{ textAlign: 'center', padding: '20px', color: 'var(--text-secondary)', fontSize: '13px' }}>
            لا توجد قطع مستعملة متاحة
          </div>
        )}
      </CardContent>
    </Card>
  );
}