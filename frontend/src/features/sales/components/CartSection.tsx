import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Badge } from '../../../components/ui/badge';
import { ShoppingCart, Plus, Minus, Trash2 } from 'lucide-react';
import { CartItem } from '../types/pos.types';
import { cn } from '../../../utils';
import type { InventoryItem } from '../../../types/models';

interface PartTypeOption {
  id: string;
  name_ar?: string;
  color?: string;
}

interface CartSectionProps {
  cart: CartItem[];
  inventoryItems: InventoryItem[];
  partTypes: PartTypeOption[];
  onUpdateQuantity: (barcode: string, quantity: number) => void;
  onRemoveFromCart: (barcode: string) => void;
}

export function CartSection({
  cart,
  inventoryItems,
  partTypes,
  onUpdateQuantity,
  onRemoveFromCart,
}: CartSectionProps) {
  return (
    <Card className="cart-section-card">
      <CardHeader>
        <CardTitle className="cart-section-title">
          <div className="cart-section-icon">
            <ShoppingCart className="w-5 h-5 text-emerald-400" />
          </div>
          سلة المشتريات
          <Badge variant="default" className="cart-section-badge">
            {cart.length}
          </Badge>
        </CardTitle>
      </CardHeader>
      <CardContent>
        {cart.length === 0 ? (
          <div className="cart-section-empty">
            <ShoppingCart className="w-12 h-12 text-secondary mx-auto mb-4" />
            <p className="cart-section-empty-text">
              السلة فارغة
            </p>
            <p className="cart-section-empty-subtext">
              امسح الباركود أو اختر منتجات للبدء
            </p>
          </div>
        ) : (
          <div className="scrollable-card-sm cart-section-items">
            {cart.map((item) => {
              const inventoryItem = inventoryItems.find((inv) => inv.id === item.id);
              const partType = inventoryItem ? partTypes.find((pt) => pt.id === inventoryItem.part_type_id) : null;
              return (
                <div
                  key={item.barcode}
                  className={cn(
                    'cart-section-item',
                    partType && 'cart-section-item-with-type'
                  )}
                  style={
                    partType 
                      ? { background: `linear-gradient(135deg, ${partType.color}15 0%, transparent 100%)` }
                      : undefined
                  }
                >
                  <div className="cart-section-item-info">
                    <div className="cart-section-item-name">
                      {item.name}
                    </div>
                    {partType && (
                      <div 
                        className="cart-section-item-type"
                        style={{ color: partType.color }}
                      >
                        {partType.name_ar}
                      </div>
                    )}
                    <div className="cart-section-item-price">
                      ₪{item.price} × {item.quantity}
                    </div>
                    {item.serialNumber && (
                      <div className="cart-section-item-cost">S/N: {item.serialNumber}</div>
                    )}
                    {item.purchaseCost && (
                      <div className="cart-section-item-cost">
                        ت: ₪{item.purchaseCost}
                      </div>
                    )}
                  </div>
                  <div className="cart-section-item-total">
                    <div className="cart-section-item-total-text">
                      ₪{item.total}
                    </div>
                  </div>
                  <div className="cart-section-item-actions">
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => onUpdateQuantity(item.barcode, item.quantity - 1)}
                      className="cart-section-btn cart-section-btn-minus"
                    >
                      <Minus className="w-3 h-3" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => onUpdateQuantity(item.barcode, item.quantity + 1)}
                      className="cart-section-btn cart-section-btn-plus"
                    >
                      <Plus className="w-3 h-3" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => onRemoveFromCart(item.barcode)}
                      className="cart-section-btn cart-section-btn-delete"
                    >
                      <Trash2 className="w-3 h-3" />
                    </Button>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </CardContent>
    </Card>
  );
}