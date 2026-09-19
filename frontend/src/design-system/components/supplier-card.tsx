import { type HTMLAttributes } from 'react';
import { cn } from '../../utils';
import { Phone, Mail, Package, Eye, Edit, Trash2, ChevronDown, ChevronUp, CircleDollarSign, RotateCcw, UserRound } from 'lucide-react';
import { Button } from './button';
import { Badge } from './badge';
import { DataTable, Column } from './tables/data-table';
import { ActionMenu } from './action-menu';

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

const formatCurrency = (value?: number) => `₪${Math.round(value || 0).toLocaleString('en-US')}`;

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
  const outstanding = supplier.outstanding || 0;
  const hasOutstanding = outstanding > 0;

  const inventoryColumns: Column<any>[] = [
    { key: 'product_name', title: 'المنتج', sortable: true, render: (item) => <span className="font-medium">{item.product_name}</span> },
    { key: 'total_received', title: 'المستلمة', sortable: true },
    { key: 'available', title: 'المتاحة', sortable: true, render: (item) => <span className="text-green-600">{item.available}</span> },
    { key: 'sold', title: 'المباعة', sortable: true, render: (item) => <span className="text-red-600">{item.sold}</span> },
    { key: 'reserved', title: 'المحجوزة', sortable: true, render: (item) => <span className="text-yellow-600">{item.reserved}</span> },
    { key: 'damaged', title: 'التالفة', sortable: true, render: (item) => <span className="text-orange-600">{item.damaged}</span> },
    { key: 'avg_cost', title: 'متوسط التكلفة', sortable: true, render: (item) => formatCurrency(item.avg_cost) },
    { key: 'avg_price', title: 'متوسط السعر', sortable: true, render: (item) => formatCurrency(item.avg_price) },
  ];

  return (
    <div
      className={cn(
        'group relative overflow-hidden rounded-[18px] border border-[var(--border-default)] border-t-4 border-t-[var(--primary)] bg-[var(--bg-surface)]',
        'shadow-[0_8px_22px_rgba(15,23,42,0.04)] transition-all duration-200 ease-out',
        'hover:-translate-y-0.5 hover:border-[var(--primary)] hover:shadow-[0_12px_30px_rgba(37,99,235,0.12)]',
        'active:scale-[0.99]',
        className
      )}
      {...props}
    >
      <div className="bg-[var(--color-primary-05)]/35 p-5">
        {/* Header: avatar + name + outstanding */}
        <div className="flex items-start gap-3">
          <div
            className="flex h-14 w-14 shrink-0 items-center justify-center rounded-2xl border border-[var(--border-subtle)] text-xl font-black shadow-sm"
            style={{ background: avatar.bg, color: avatar.color }}
            aria-hidden="true"
          >
            <UserRound className="h-6 w-6" aria-hidden="true" />
          </div>
          <div className="min-w-0 flex-1">
            <div className="flex items-start justify-between gap-2">
              <div className="min-w-0">
                <span className="block truncate text-lg font-black text-[var(--text-primary)]">
                  {supplier.name}
                </span>
                <span className="mt-1 block truncate text-[11px] font-medium text-[var(--text-muted)]">
                  رمز المورد: {('code' in supplier && supplier.code) || '-'}
                </span>
              </div>
              {hasOutstanding ? (
                <Badge variant="danger" size="sm" className="rounded-full">
                  {formatCurrency(outstanding)}
                </Badge>
              ) : (
                <Badge variant="success" size="sm" className="rounded-full">
                  مسدد
                </Badge>
              )}
            </div>

            {/* Contact */}
            <div className="mt-4 grid gap-2 sm:grid-cols-2">
              <div className="flex min-w-0 items-center gap-2 rounded-lg border border-[var(--border-subtle)] bg-[var(--bg-surface)]/70 px-2.5 py-2 text-small font-medium text-[var(--text-secondary)]">
                <Phone className="h-3.5 w-3.5 text-[var(--text-muted)]" />
                <span className="truncate" dir="ltr">{supplier.phone || '-'}</span>
              </div>
              {supplier.email ? (
                <div className="flex min-w-0 items-center gap-2 rounded-lg border border-[var(--border-subtle)] bg-[var(--bg-surface)]/70 px-2.5 py-2 text-small font-medium text-[var(--text-secondary)]">
                  <Mail className="h-3.5 w-3.5 text-[var(--text-muted)]" />
                  <span className="truncate">{supplier.email}</span>
                </div>
              ) : null}
            </div>
          </div>
        </div>

        {/* Finance summary */}
        <div className="mt-5 grid grid-cols-3 gap-2 rounded-2xl border border-[var(--border-subtle)] bg-[var(--bg-surface)] p-3.5 shadow-sm">
          <div className="text-center">
            <div className="text-[11px] font-medium text-[var(--text-muted)]">المشتريات</div>
            <div className="mt-1 text-small font-black text-[var(--text-primary)]">
              {formatCurrency(supplier.totalPurchases)}
            </div>
          </div>
          <div className="border-x border-[var(--border-subtle)] text-center">
            <div className="text-[11px] font-medium text-[var(--text-muted)]">المدفوع</div>
            <div className="mt-1 text-small font-black text-[var(--color-success)]">
              {formatCurrency(supplier.paidAmount)}
            </div>
          </div>
          <div className="text-center">
            <div className="text-[11px] font-medium text-[var(--text-muted)]">المستحق</div>
            <div className={`mt-1 flex items-center justify-center gap-1 text-small font-black ${hasOutstanding ? 'text-[var(--color-danger)]' : 'text-[var(--text-secondary)]'}`}>
              {hasOutstanding && <CircleDollarSign className="h-3.5 w-3.5" />}
              {formatCurrency(outstanding)}
            </div>
          </div>
        </div>

        {/* Last purchase + actions */}
        <div className="mt-4 flex flex-wrap items-center justify-between gap-3 border-t border-[var(--border-subtle)] pt-4">
          <div className="flex items-center gap-1.5 text-[11px] font-medium text-[var(--text-muted)]">
            <span>آخر شراء:</span>
            <span>
              {supplier.lastPurchase
                ? new Date(supplier.lastPurchase).toLocaleDateString('ar-SA')
                : '-'}
            </span>
          </div>
          <div className="flex items-center gap-1">
            <Button
              variant="secondary"
              size="sm"
              className="gap-1.5 rounded-lg"
              onClick={() => onToggle?.(supplier.id)}
              aria-label={expanded ? 'إخفاء البضاعة' : 'عرض البضاعة'}
            >
              {expanded ? <ChevronUp className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}
              <span>{expanded ? 'إخفاء البضاعة' : 'عرض البضاعة'}</span>
            </Button>
            <ActionMenu
              label="خيارات المورد"
              widthClassName="w-48"
              items={[
                { label: 'عرض', icon: Eye, onClick: () => onView?.(supplier) },
                { label: 'تعديل', icon: Edit, onClick: () => onEdit?.(supplier) },
                ...(onRestore ? [{ label: 'إعادة تفعيل', icon: RotateCcw, onClick: () => onRestore(supplier), danger: false }] : []),
                { label: onRestore ? 'إيقاف المورد' : 'حذف', icon: onRestore ? RotateCcw : Trash2, onClick: () => (onRestore ? onRestore(supplier) : onDelete?.(supplier)), danger: true },
              ]}
            />
          </div>
        </div>
      </div>

      {/* Expandable inventory */}
      {expanded && (
        <div className="border-t border-[var(--border-subtle)] bg-[var(--bg-surface-elevated)]/60 p-4">
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
