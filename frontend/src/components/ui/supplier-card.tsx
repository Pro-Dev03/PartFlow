import { type HTMLAttributes } from 'react';
import { cn } from '../../utils';
import { Phone, Mail, Package, Eye, Edit, Trash2, ChevronDown, ChevronUp, CircleDollarSign, RotateCcw } from 'lucide-react';
import { Button } from './button';
import { Badge } from './badge';
import { getButtonSize } from '../../config/button-sizes';
import { DataTable, Column } from '../tables/data-table';

export interface SupplierCardProps extends Omit<HTMLAttributes<HTMLDivElement>, 'onToggle'> {
  supplier: {
    id: string | number;
    name: string;
    phone?: string;
    email?: string;
    totalPurchases?: number;
    paidAmount?: number;
    outstanding?: number;
    lastPurchase?: string | null;
  };
  expanded?: boolean;
  inventory?: any;
  inventoryLoading?: boolean;
  onToggle?: (id: string | number) => void;
  onView?: (supplier: SupplierCardProps['supplier']) => void;
  onEdit?: (supplier: SupplierCardProps['supplier']) => void;
  onDelete?: (supplier: SupplierCardProps['supplier']) => void;
  onRestore?: (supplier: SupplierCardProps['supplier']) => void;
}

const AVATAR_PALETTE = [
  { bg: 'var(--color-primary-10)', color: 'var(--color-primary)' },
  { bg: 'var(--color-success-10)', color: 'var(--color-success)' },
  { bg: 'var(--color-warning-10)', color: 'var(--color-warning)' },
  { bg: 'var(--color-info-10)', color: 'var(--color-info)' },
  { bg: 'var(--color-danger-10)', color: 'var(--color-danger)' },
];

const getAvatarStyle = (name: string) => {
  let hash = 0;
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash);
  }
  return AVATAR_PALETTE[Math.abs(hash) % AVATAR_PALETTE.length];
};

const formatCurrency = (value?: number) => `₪${(value || 0).toLocaleString()}`;

function SupplierCard({
  supplier,
  expanded = false,
  inventory,
  inventoryLoading = false,
  onToggle,
  onView,
  onEdit,
  onDelete,
  onRestore,
  className,
  ...props
}: SupplierCardProps) {
  const avatar = getAvatarStyle(supplier.name || '؟');
  const initial = (supplier.name || '؟').trim().charAt(0);
  const outstanding = supplier.outstanding || 0;
  const hasOutstanding = outstanding > 0;

  const inventoryColumns: Column<any>[] = [
    { key: 'product_name', title: 'المنتج', sortable: true, render: (item) => <span className="font-medium">{item.product_name}</span> },
    { key: 'total_received', title: 'المستلمة', sortable: true },
    { key: 'available', title: 'المتاحة', sortable: true, render: (item) => <span className="text-green-600">{item.available}</span> },
    { key: 'sold', title: 'المباعة', sortable: true, render: (item) => <span className="text-red-600">{item.sold}</span> },
    { key: 'reserved', title: 'المحجوزة', sortable: true, render: (item) => <span className="text-yellow-600">{item.reserved}</span> },
    { key: 'damaged', title: 'التالفة', sortable: true, render: (item) => <span className="text-orange-600">{item.damaged}</span> },
    { key: 'avg_cost', title: 'متوسط التكلفة', sortable: true, render: (item) => `₪${(item.avg_cost || 0).toFixed(2)}` },
    { key: 'avg_price', title: 'متوسط السعر', sortable: true, render: (item) => `₪${(item.avg_price || 0).toFixed(2)}` },
  ];

  return (
    <div
      className={cn(
        'group relative rounded-xl border border-border bg-surface',
        'transition-all duration-200 ease-out',
        'hover:border-primary hover:shadow-[0_4px_12px_rgba(37,99,235,0.12)]',
        'active:scale-[0.99]',
        className
      )}
      {...props}
    >
      <div className="p-4">
        {/* Header: avatar + name + outstanding */}
        <div className="flex items-start gap-3">
          <div
            className="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg border border-border text-base font-bold"
            style={{ background: avatar.bg, color: avatar.color }}
            aria-hidden="true"
          >
            {initial}
          </div>
          <div className="min-w-0 flex-1">
            <div className="flex items-start justify-between gap-2">
              <div className="min-w-0">
                <span className="block truncate text-sm font-semibold text-text-primary">
                  {supplier.name}
                </span>
                <span className="mt-0.5 block text-tiny text-text-muted">
                  رمز المورد: {('code' in supplier && supplier.code) || '-'}
                </span>
              </div>
              {hasOutstanding ? (
                <Badge variant="danger" size="sm">
                  {formatCurrency(outstanding)}
                </Badge>
              ) : (
                <Badge variant="success" size="sm">
                  مسدد
                </Badge>
              )}
            </div>

            {/* Contact */}
            <div className="mt-2 space-y-1.5">
              <div className="flex items-center gap-2 text-small text-text-secondary">
                <Phone className="h-3.5 w-3.5 text-text-muted" />
                <span className="truncate" dir="ltr">{supplier.phone || '-'}</span>
              </div>
              {supplier.email ? (
                <div className="flex items-center gap-2 text-small text-text-secondary">
                  <Mail className="h-3.5 w-3.5 text-text-muted" />
                  <span className="truncate">{supplier.email}</span>
                </div>
              ) : null}
            </div>
          </div>
        </div>

        {/* Finance summary */}
        <div className="mt-4 grid grid-cols-3 gap-2 rounded-lg border border-border bg-surface-elevated p-3">
          <div className="text-center">
            <div className="text-tiny text-text-muted">المشتريات</div>
            <div className="mt-0.5 text-small font-semibold text-text-primary">
              {formatCurrency(supplier.totalPurchases)}
            </div>
          </div>
          <div className="text-center border-x border-border">
            <div className="text-tiny text-text-muted">المدفوع</div>
            <div className="mt-0.5 text-small font-semibold text-green">
              {formatCurrency(supplier.paidAmount)}
            </div>
          </div>
          <div className="text-center">
            <div className="text-tiny text-text-muted">المستحق</div>
            <div className={`mt-0.5 flex items-center justify-center gap-1 text-small font-semibold ${hasOutstanding ? 'text-red-500' : 'text-text-secondary'}`}>
              {hasOutstanding && <CircleDollarSign className="h-3.5 w-3.5" />}
              {formatCurrency(outstanding)}
            </div>
          </div>
        </div>

        {/* Last purchase + actions */}
        <div className="mt-3 flex items-center justify-between border-t border-border pt-3">
          <div className="flex items-center gap-1.5 text-tiny text-text-tertiary">
            <span>آخر شراء:</span>
            <span>
              {supplier.lastPurchase
                ? new Date(supplier.lastPurchase).toLocaleDateString('ar-SA')
                : '-'}
            </span>
          </div>
          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size={getButtonSize('suppliers', 'iconAction')}
              onClick={() => onToggle?.(supplier.id)}
              aria-label={expanded ? 'إخفاء البضاعة' : 'عرض البضاعة'}
            >
              {expanded ? <ChevronUp className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}
            </Button>
            <Button
              variant="ghost"
              size={getButtonSize('suppliers', 'iconAction')}
              onClick={() => onView?.(supplier)}
              aria-label="عرض"
            >
              <Eye className="w-3.5 h-3.5" />
            </Button>
            <Button
              variant="ghost"
              size={getButtonSize('suppliers', 'iconAction')}
              onClick={() => onEdit?.(supplier)}
              aria-label="تعديل"
            >
              <Edit className="w-3.5 h-3.5" />
            </Button>
            {onRestore ? (
              <Button
                variant="ghost"
                size={getButtonSize('suppliers', 'iconAction')}
                onClick={() => onRestore(supplier)}
                aria-label="إعادة تفعيل"
                className="text-green-600 hover:text-green-700 hover:bg-green-50"
              >
                <RotateCcw className="w-3.5 h-3.5" />
              </Button>
            ) : (
              <Button
                variant="ghost"
                size={getButtonSize('suppliers', 'iconAction')}
                onClick={() => onDelete?.(supplier)}
                aria-label="إيقاف"
                className="text-red-500 hover:text-red-600 hover:bg-red-50"
              >
                <Trash2 className="w-3.5 h-3.5" />
              </Button>
            )}
          </div>
        </div>
      </div>

      {/* Expandable inventory */}
      {expanded && (
        <div className="border-t border-border p-4 bg-surface-elevated">
          <div className="mb-4 flex items-center gap-2">
            <Package className="w-4 h-4 text-primary" />
            <h4 className="font-semibold text-text">بضاعة المورد - {supplier.name}</h4>
          </div>
          {inventoryLoading ? (
            <div className="py-8 text-center text-text-muted">جاري التحميل...</div>
          ) : inventory && inventory.data && inventory.data.length > 0 ? (
            <DataTable data={inventory.data} columns={inventoryColumns} expandable={false} />
          ) : (
            <div className="py-8 text-center text-text-muted">لا توجد بضاعة من هذا المورد</div>
          )}
        </div>
      )}
    </div>
  );
}

export { SupplierCard };
