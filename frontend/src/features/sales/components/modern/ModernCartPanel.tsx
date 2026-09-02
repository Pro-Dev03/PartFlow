import { ShoppingCart, Minus, Plus, Trash2, ShoppingBag } from 'lucide-react';
import { CartItem } from '../../types/pos.types';

interface ModernCartPanelProps {
  cart: CartItem[];
  total: number;
  onUpdateQuantity: (barcode: string, quantity: number) => void;
  onRemoveFromCart: (barcode: string) => void;
  onClearCart: () => void;
}

export function ModernCartPanel({
  cart,
  total,
  onUpdateQuantity,
  onRemoveFromCart,
  onClearCart,
}: ModernCartPanelProps) {
  return (
    <div className="pos-modern-cart-panel">
      {/* Cart Header */}
      <div className="cart-panel-header">
        <div className="cart-panel-title-group">
          <div className="cart-panel-icon">
            <ShoppingCart className="w-5 h-5" />
          </div>
          <span className="cart-panel-title">السلة</span>
          <span className="cart-panel-count">{cart.length}</span>
        </div>
        {cart.length > 0 && (
          <button className="cart-panel-clear" onClick={onClearCart}>
            <Trash2 className="w-4 h-4" />
            <span>مسح</span>
          </button>
        )}
      </div>

      {/* Cart Items */}
      <div className="cart-panel-items">
        {cart.length === 0 ? (
          <div className="cart-panel-empty">
            <ShoppingBag className="empty-icon" />
            <p className="empty-text">السلة فارغة</p>
            <p className="empty-hint">امسح الباركود أو اختر منتجاً</p>
          </div>
        ) : (
          <div className="cart-items-list">
            {cart.map((item) => (
              <ModernCartItem
                key={item.barcode}
                item={item}
                onUpdateQuantity={onUpdateQuantity}
                onRemove={onRemoveFromCart}
              />
            ))}
          </div>
        )}
      </div>

      {/* Cart Total */}
      {cart.length > 0 && (
        <div className="cart-panel-footer">
          <div className="cart-panel-subtotal">
            <span>المجموع الفرعي</span>
            <span>₪{total.toLocaleString()}</span>
          </div>
          <div className="cart-panel-total">
            <span>الإجمالي</span>
            <span className="total-value">₪{total.toLocaleString()}</span>
          </div>
        </div>
      )}
    </div>
  );
}

interface ModernCartItemProps {
  item: CartItem;
  onUpdateQuantity: (barcode: string, quantity: number) => void;
  onRemove: (barcode: string) => void;
}

function ModernCartItem({
  item,
  onUpdateQuantity,
  onRemove,
}: ModernCartItemProps) {
  return (
    <div className="cart-item">
      <div className="cart-item-main">
        <div className="cart-item-info">
          <span className="cart-item-name">{item.name}</span>
          {item.partType && (
            <span
              className="cart-item-type"
              style={{ color: item.partTypeColor }}
            >
              {item.partType}
            </span>
          )}
          {item.serialNumber && (
            <span className="cart-item-serial">S/N: {item.serialNumber}</span>
          )}
        </div>
        <button
          className="cart-item-remove"
          onClick={() => onRemove(item.barcode)}
          aria-label="إزالة المنتج"
        >
          <Trash2 className="w-3.5 h-3.5" />
        </button>
      </div>
      <div className="cart-item-footer">
        <div className="cart-item-quantity">
          <button
            className="qty-btn minus"
            onClick={() => onUpdateQuantity(item.barcode, item.quantity - 1)}
            aria-label="تقليل"
          >
            <Minus className="w-3.5 h-3.5" />
          </button>
          <span className="qty-value">{item.quantity}</span>
          <button
            className="qty-btn plus"
            onClick={() => onUpdateQuantity(item.barcode, item.quantity + 1)}
            aria-label="زيادة"
          >
            <Plus className="w-3.5 h-3.5" />
          </button>
        </div>
        <div className="cart-item-pricing">
          {item.quantity > 1 && (
            <span className="cart-item-unit-price">₪{item.price.toLocaleString()}</span>
          )}
          <span className="cart-item-total">₪{item.total.toLocaleString()}</span>
        </div>
      </div>
    </div>
  );
}
