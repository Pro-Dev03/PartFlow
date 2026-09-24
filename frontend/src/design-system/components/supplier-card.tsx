import { type HTMLAttributes, useState } from 'react';
import { cn } from '../../utils';
import { formatStoreDate } from '../../utils/store-time';
import { Phone, Mail, Package, Eye, Edit, Trash2, ChevronDown, ChevronUp, CircleDollarSign, RotateCcw, UserRound, ScanSearch, CircleX } from 'lucide-react';
import { Button } from './button';
import { Badge } from './badge';
import { Input } from './input';
import { Table, TableHeader, TableBody, TableHead, TableRow, TableCell } from './table';

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
  showActions?: boolean;
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
  showActions = true,
  onToggle,
  onView,
  onEdit,
  onDelete,
  onRestore,
  className,
  ...props
}: SupplierCardProps) {
  const [productSearch, setProductSearch] = useState('');
  const avatar = getAvatarStyle(supplier.name || '؟');
  const outstanding = supplier.outstanding || 0;
  const hasOutstanding = outstanding > 0;

  const inventoryRows = Array.isArray(inventory?.data) ? inventory.data : [];
  const filteredInventoryRows = inventoryRows.filter((item: any) =>
    String(item.product_name || '').toLowerCase().includes(productSearch.trim().toLowerCase())
  );
  const inventoryTotals = inventoryRows.reduce((totals: { received: number; available: number; sold: number; returned: number }, item: any) => ({
    received: totals.received + Number(item.total_received || 0),
    available: totals.available + Number(item.available || 0),
    sold: totals.sold + Number(item.sold || 0),
    returned: totals.returned + Number(item.returned || 0),
  }), { received: 0, available: 0, sold: 0, returned: 0 });

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
      <div className="bg-[var(--color-primary-05)]/35 p-[14px]">
        {/* Header: avatar + name + outstanding */}
        <div className="flex items-start gap-[12px]">
          <div
            className="flex h-14 w-14 shrink-0 items-center justify-center rounded-2xl border border-[var(--border-subtle)] text-xl font-black shadow-sm"
            style={{ background: avatar.bg, color: avatar.color }}
            aria-hidden="true"
          >
            <UserRound className="h-6 w-6" aria-hidden="true" />
          </div>
          <div className="min-w-0 flex-1">
            <div className="flex items-start justify-between gap-[8px]">
              <div className="min-w-0">
                <span className="block truncate text-lg font-black text-[var(--text-primary)]">
                  {supplier.name}
                </span>
                <span className="mt-[4px] block truncate text-[11px] font-medium text-[var(--text-muted)]">
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
            <div className="mt-[10px] flex flex-wrap items-center justify-between gap-[8px]">
              <div className="flex min-w-0 items-center gap-[8px] rounded-lg border border-[var(--border-subtle)] bg-[var(--bg-surface)]/70 px-[10px] py-[8px] text-small font-medium text-[var(--text-secondary)]">
                <Phone className="h-3.5 w-3.5 text-[var(--text-muted)]" />
                <span className="truncate" dir="ltr">{supplier.phone || '-'}</span>
              </div>
              <div className="flex items-center gap-[6px] text-[11px] font-medium text-[var(--text-muted)]">
                <span>آخر شراء:</span>
                <span className="text-[var(--text-secondary)]">
                  {supplier.lastPurchase
                    ? formatStoreDate(supplier.lastPurchase, 'ar-SA')
                    : '-'}
                </span>
              </div>
              {supplier.email ? (
                <div className="flex min-w-0 basis-full items-center gap-[8px] rounded-lg border border-[var(--border-subtle)] bg-[var(--bg-surface)]/70 px-[10px] py-[8px] text-small font-medium text-[var(--text-secondary)]">
                  <Mail className="h-3.5 w-3.5 text-[var(--text-muted)]" />
                  <span className="truncate">{supplier.email}</span>
                </div>
              ) : null}
            </div>
          </div>
        </div>

        {/* Finance summary */}
        <div className="mt-[12px] grid grid-cols-3 gap-[8px] rounded-2xl border border-[var(--border-subtle)] bg-[var(--bg-surface)] p-[10px] shadow-sm">
          <div className="text-center">
            <div className="text-[11px] font-medium text-[var(--text-muted)]">المشتريات</div>
            <div className="mt-[4px] text-small font-black text-[var(--text-primary)]">
              {formatCurrency(supplier.totalPurchases)}
            </div>
          </div>
          <div className="border-x border-[var(--border-subtle)] text-center">
            <div className="text-[11px] font-medium text-[var(--text-muted)]">المدفوع</div>
            <div className="mt-[4px] text-small font-black text-[var(--color-success)]">
              {formatCurrency(supplier.paidAmount)}
            </div>
          </div>
          <div className="text-center">
            <div className="text-[11px] font-medium text-[var(--text-muted)]">المستحق</div>
            <div className={`mt-[4px] flex items-center justify-center gap-[4px] text-small font-black ${hasOutstanding ? 'text-[var(--color-danger)]' : 'text-[var(--text-secondary)]'}`}>
              {hasOutstanding && <CircleDollarSign className="h-3.5 w-3.5" />}
              {formatCurrency(outstanding)}
            </div>
          </div>
        </div>

        <div className="mt-[10px] flex flex-wrap items-center justify-end gap-[12px] border-t border-[var(--border-subtle)] pt-[10px]">
          {showActions && (
            <div className="flex items-center gap-[4px]">
              <Button
                variant="secondary"
                size="sm"
                className="gap-[6px] rounded-lg"
                onClick={() => onToggle?.(supplier.id)}
                aria-label={expanded ? 'إخفاء البضاعة' : 'عرض البضاعة'}
              >
                {expanded ? <ChevronUp className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}
                <span>{expanded ? 'إخفاء البضاعة' : 'عرض البضاعة'}</span>
              </Button>
              <Button
                variant="ghost"
                size="icon"
                className="rounded-lg"
                onClick={() => onView?.(supplier)}
                aria-label="عرض بيانات المورد"
                title="عرض بيانات المورد"
              >
                <Eye className="h-4 w-4" aria-hidden="true" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                className="rounded-lg"
                onClick={() => onEdit?.(supplier)}
                aria-label="تعديل المورد"
                title="تعديل المورد"
              >
                <Edit className="h-4 w-4" aria-hidden="true" />
              </Button>
              {onRestore ? (
                <Button
                  variant="success"
                  size="icon"
                  className="rounded-lg"
                  onClick={() => onRestore(supplier)}
                  aria-label="إعادة تفعيل المورد"
                  title="إعادة تفعيل المورد"
                >
                  <RotateCcw className="h-4 w-4" aria-hidden="true" />
                </Button>
              ) : (
                <Button
                  variant="danger"
                  size="icon"
                  className="rounded-lg"
                  onClick={() => onDelete?.(supplier)}
                  aria-label="إيقاف المورد"
                  title="إيقاف المورد"
                >
                  <Trash2 className="h-4 w-4" aria-hidden="true" />
                </Button>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Expandable inventory */}
      {expanded && (
        <div className="border-t border-[var(--border-subtle)] bg-[var(--bg-surface-elevated)]/60 p-[10px] sm:p-[12px]">
          <div className="mb-[8px] flex flex-wrap items-center justify-between gap-[8px]">
            <div className="flex items-center gap-[10px]">
              <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-[var(--color-primary-10)] text-primary">
                <Package className="h-4.5 w-4.5" />
              </span>
              <div>
                <h4 className="font-bold leading-tight text-text">بضاعة المورد - {supplier.name}</h4>
                <p className="mt-0.5 text-[11px] text-text-muted">حركة المخزون حسب المنتج</p>
              </div>
            </div>
            <Badge variant="info" size="sm">{inventoryRows.length} منتجات</Badge>
          </div>
          <div className="mb-[8px] flex flex-col gap-[8px] sm:flex-row sm:items-center sm:justify-between">
            <div
              className="group relative w-full overflow-hidden border border-border bg-surface-muted/60 shadow-sm transition-all focus-within:border-primary/50 focus-within:bg-surface focus-within:shadow-[0_0_0_3px_var(--color-primary-10)]"
              style={{ maxWidth: '420px', borderRadius: '14px', padding: '6px' }}
            >
              <div
                className="absolute top-1/2 z-10 flex items-center gap-1"
                style={{ left: '8px', transform: 'translateY(-50%)' }}
              >
                <span
                  className="pointer-events-none flex items-center justify-center bg-primary text-white"
                  style={{ width: '28px', height: '28px', borderRadius: '999px' }}
                  aria-hidden="true"
                >
                  <ScanSearch className="h-3.5 w-3.5" />
                </span>
                {productSearch && (
                  <button
                    type="button"
                    onClick={() => setProductSearch('')}
                    className="flex items-center justify-center text-text-muted transition-colors hover:bg-danger/10 hover:text-danger"
                    style={{ width: '28px', height: '28px', padding: 0, border: 'none', borderRadius: '999px', background: 'transparent' }}
                    aria-label="مسح البحث"
                    title="مسح البحث"
                  >
                    <CircleX className="h-4 w-4" strokeWidth={2} aria-hidden="true" />
                  </button>
                )}
              </div>
              <Input
                value={productSearch}
                onChange={(event) => setProductSearch(event.target.value)}
                placeholder="ابحث باسم المنتج"
                className="h-9 border-0 bg-transparent text-sm shadow-none transition-colors placeholder:text-text-muted focus:border-0 focus:bg-transparent focus:ring-0"
                style={{ borderRadius: '10px', paddingLeft: productSearch ? '78px' : '45px', paddingRight: '12px' }}
                dir="rtl"
                aria-label={`البحث في منتجات ${supplier.name}`}
              />
            </div>
            <span className="text-xs text-text-muted">عرض {filteredInventoryRows.length} من {inventoryRows.length} منتجات</span>
          </div>
          {inventoryLoading ? (
            <div className="py-8 text-center text-text-muted">جاري التحميل...</div>
          ) : inventoryRows.length > 0 ? (
            <>
              <div className="mb-[8px] grid grid-cols-2 gap-1.5 sm:grid-cols-4">
                {[
                  { label: 'إجمالي المستلم', value: inventoryTotals.received, tone: 'text-text-primary', bg: 'bg-surface' },
                  { label: 'المتاح للبيع', value: inventoryTotals.available, tone: 'text-success', bg: 'bg-success/10' },
                  { label: 'المباع', value: inventoryTotals.sold, tone: 'text-info', bg: 'bg-info/10' },
                  { label: 'مرتجعات المورد', value: inventoryTotals.returned, tone: 'text-warning', bg: 'bg-warning/10' },
                ].map((summary) => (
                  <div
                    key={summary.label}
                    className={`flex items-center justify-between gap-2 border border-border ${summary.bg}`}
                    style={{ borderRadius: '12px', padding: '9px 12px' }}
                  >
                    <div className="text-[11px] font-medium text-text-muted">{summary.label}</div>
                    <div className={`text-base font-black ${summary.tone}`}>{summary.value}</div>
                  </div>
                ))}
              </div>

              <div className="max-h-[360px] overflow-auto rounded-xl">
                <Table className="min-w-[760px] text-xs">
                  <TableHeader className="sticky top-0 z-10">
                    <TableRow className="hover:bg-[var(--bg-surface-muted)]">
                      <TableHead
                        className="bg-[var(--bg-surface-muted)]"
                        style={{ padding: '9px 12px', backgroundColor: 'var(--bg-surface-muted)', color: 'var(--primary)', fontWeight: 800, borderBottom: '2px solid var(--primary)' }}
                      >
                        المنتج
                      </TableHead>
                      {['المستلمة', 'المتاحة', 'المباعة', 'المحجوزة', 'المرتجعة', 'متوسط التكلفة', 'متوسط السعر'].map((title) => (
                        <TableHead
                          key={title}
                          className="bg-[var(--bg-surface-muted)] text-center"
                          style={{ padding: '9px 12px', backgroundColor: 'var(--bg-surface-muted)', color: 'var(--text-secondary)', fontWeight: 700, textAlign: 'center', borderBottom: '2px solid var(--border-default)' }}
                        >
                          {title}
                        </TableHead>
                      ))}
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredInventoryRows.map((item: any) => (
                      <TableRow key={item.product_id || item.product_name}>
                        <TableCell className="max-w-[180px] truncate py-2 font-semibold" title={item.product_name}>
                          {item.product_name}
                        </TableCell>
                        <TableCell className="py-2 text-center">{Number(item.total_received || 0)}</TableCell>
                        <TableCell className="py-2 text-center font-bold text-success">{Number(item.available || 0)}</TableCell>
                        <TableCell className="py-2 text-center text-info">{Number(item.sold || 0)}</TableCell>
                        <TableCell className="py-2 text-center text-warning">{Number(item.reserved || 0)}</TableCell>
                        <TableCell className="py-2 text-center text-danger">{Number(item.returned || 0)}</TableCell>
                        <TableCell className="py-2 text-center">{formatCurrency(item.avg_cost)}</TableCell>
                        <TableCell className="py-2 text-center font-semibold text-success">{formatCurrency(item.avg_price)}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
              {filteredInventoryRows.length === 0 && (
                <div className="mt-2 rounded-xl border border-dashed border-border bg-surface-muted px-4 py-6 text-center text-text-muted">
                  {productSearch ? 'لا توجد منتجات مطابقة للبحث' : 'لا توجد بضاعة من هذا المورد'}
                </div>
              )}
            </>
          ) : (
            <div className="rounded-xl border border-dashed border-border bg-surface-muted px-4 py-8 text-center text-text-muted">لا توجد بضاعة من هذا المورد</div>
          )}
        </div>
      )}
    </div>
  );
}

export { SupplierCard };
