import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Badge } from '../../../components/ui/badge';
import { ShoppingCart, Plus, Minus, Trash2 } from 'lucide-react';
import { CartItem } from '../types/pos.types';

interface CartSectionProps {
  cart: CartItem[];
  inventoryItems: any[];
  partTypes: any[];
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
    <Card style={{
      background: 'linear-gradient(135deg, var(--color-success-05) 0%, var(--color-info-05) 100%)',
      border: '1px solid var(--color-success-10)',
      backdropFilter: 'blur(10px)',
      transition: 'all 0.3s ease'
    }}
    onMouseEnter={(e) => {
      e.currentTarget.style.background = 'linear-gradient(135deg, var(--color-success-08) 0%, var(--color-info-08) 100%)';
      e.currentTarget.style.borderColor = 'var(--color-success-20)';
      e.currentTarget.style.transform = 'translateY(-2px)';
      e.currentTarget.style.boxShadow = '0 8px 25px var(--color-success-10)';
    }}
    onMouseLeave={(e) => {
      e.currentTarget.style.background = 'linear-gradient(135deg, var(--color-success-05) 0%, var(--color-info-05) 100%)';
      e.currentTarget.style.borderColor = 'var(--color-success-10)';
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
            background: 'linear-gradient(135deg, var(--color-success-15) 0%, var(--color-info-15) 100%)',
            border: '1px solid var(--color-success-25)'
          }}>
            <ShoppingCart className="w-5 h-5 text-emerald-400" />
          </div>
          سلة المشتريات
          <Badge
            variant="default"
            style={{
              fontSize: '11px',
              fontWeight: '600',
              padding: '4px 8px',
              borderRadius: '6px',
              background: 'linear-gradient(135deg, var(--color-success-20) 0%, var(--color-info-20) 100%)',
              border: '1px solid var(--color-success-30)',
              boxShadow: '0 2px 8px var(--color-success-15)'
            }}
          >
            {cart.length}
          </Badge>
        </CardTitle>
      </CardHeader>
      <CardContent>
        {cart.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '40px 20px' }}>
            <ShoppingCart className="w-12 h-12 text-secondary mx-auto mb-4" />
            <p style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
              السلة فارغة
            </p>
            <p style={{ fontSize: '11px', color: 'var(--text-tertiary)', marginTop: '4px' }}>
              امسح الباركود أو اختر منتجات للبدء
            </p>
          </div>
        ) : (
          <div className="scrollable-card-sm" style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
            {cart.map((item) => {
              const inventoryItem = inventoryItems.find((inv: any) => inv.id === item.id);
              const partType = inventoryItem ? partTypes.find((pt: any) => pt.id === inventoryItem.part_type_id) : null;
              return (
                <div
                  key={item.barcode}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '12px',
                    padding: '14px',
                    borderRadius: '10px',
                    border: '1px solid var(--color-success-15)',
                    background: partType ? `linear-gradient(135deg, ${partType.color}15 0%, transparent 100%)` : 'linear-gradient(135deg, var(--bg-surface-elevated) 0%, var(--bg-surface) 100%)',
                    backdropFilter: 'blur(10px)',
                    transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)'
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.borderColor = 'var(--color-success-30)';
                    e.currentTarget.style.transform = 'translateX(4px)';
                    e.currentTarget.style.boxShadow = '0 4px 15px var(--color-success-10)';
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.borderColor = 'var(--color-success-15)';
                    e.currentTarget.style.transform = 'translateX(0)';
                    e.currentTarget.style.boxShadow = 'none';
                  }}
                >
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{
                      fontSize: '14px',
                      fontWeight: '600',
                      color: 'var(--text-primary)',
                      whiteSpace: 'nowrap',
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      letterSpacing: '0.2px'
                    }}>
                      {item.name}
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
                    <div style={{ fontSize: '11px', color: 'var(--text-secondary)' }}>
                      ₪{item.price} × {item.quantity}
                    </div>
                  </div>
                  <div style={{ textAlign: 'right' }}>
                    <div style={{
                      fontSize: '15px',
                      fontWeight: '700',
                      color: 'var(--color-info)',
                      textShadow: '0 0 20px var(--color-info-30)'
                    }}>
                      ₪{item.total}
                    </div>
                  </div>
                  <div style={{ display: 'flex', gap: '6px' }}>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => onUpdateQuantity(item.barcode, item.quantity - 1)}
                      style={{
                        width: '32px',
                        height: '32px',
                        borderRadius: '6px',
                        background: 'var(--color-danger-10)',
                        border: '1px solid var(--color-danger-20)',
                        transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)'
                      }}
                      onMouseEnter={(e) => {
                        e.currentTarget.style.background = 'var(--color-danger-20)';
                        e.currentTarget.style.borderColor = 'var(--color-danger-30)';
                      }}
                      onMouseLeave={(e) => {
                        e.currentTarget.style.background = 'var(--color-danger-10)';
                        e.currentTarget.style.borderColor = 'var(--color-danger-20)';
                      }}
                    >
                      <Minus className="w-3 h-3" style={{ color: 'var(--color-danger)' }} />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => onUpdateQuantity(item.barcode, item.quantity + 1)}
                      style={{
                        width: '32px',
                        height: '32px',
                        borderRadius: '6px',
                        background: 'var(--color-success-10)',
                        border: '1px solid var(--color-success-20)',
                        transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)'
                      }}
                      onMouseEnter={(e) => {
                        e.currentTarget.style.background = 'var(--color-success-20)';
                        e.currentTarget.style.borderColor = 'var(--color-success-30)';
                      }}
                      onMouseLeave={(e) => {
                        e.currentTarget.style.background = 'var(--color-success-10)';
                        e.currentTarget.style.borderColor = 'var(--color-success-20)';
                      }}
                    >
                      <Plus className="w-3 h-3" style={{ color: 'var(--color-success)' }} />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => onRemoveFromCart(item.barcode)}
                      style={{
                        width: '32px',
                        height: '32px',
                        borderRadius: '6px',
                        background: 'var(--color-warning-10)',
                        border: '1px solid var(--color-warning-20)',
                        transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)'
                      }}
                      onMouseEnter={(e) => {
                        e.currentTarget.style.background = 'var(--color-warning-20)';
                        e.currentTarget.style.borderColor = 'var(--color-warning-30)';
                      }}
                      onMouseLeave={(e) => {
                        e.currentTarget.style.background = 'var(--color-warning-10)';
                        e.currentTarget.style.borderColor = 'var(--color-warning-20)';
                      }}
                    >
                      <Trash2 className="w-3 h-3" style={{ color: 'var(--color-warning)' }} />
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