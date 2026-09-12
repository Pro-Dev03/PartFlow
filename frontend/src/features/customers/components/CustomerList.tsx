import { useState } from 'react';
import { Button } from '../../../components/ui/button';
import { EmptyState } from '../../../components/ui/empty-state';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
import { Badge } from '../../../components/ui/badge';
import { Users, Eye, Edit, Trash2, MoreHorizontal, Phone, UserRound, Inbox } from 'lucide-react';
import { Customer } from '../types/customers.types';
import { cn } from '../../../utils';

interface CustomerListProps {
  filteredCustomers: Customer[];
  isLoading: boolean;
  onViewCustomer: (customer: Customer) => void;
  onEditCustomer: (customer: Customer) => void;
  onDeleteCustomer: (customerId: string) => void;
}

function RowActionMenu({
  customer,
  onViewCustomer,
  onEditCustomer,
  onDeleteCustomer,
}: {
  customer: Customer;
  onViewCustomer: (customer: Customer) => void;
  onEditCustomer: (customer: Customer) => void;
  onDeleteCustomer: (customerId: string) => void;
}) {
  const [open, setOpen] = useState(false);

  const actions = [
    { label: 'عرض', icon: Eye, onClick: () => onViewCustomer(customer) },
    { label: 'تعديل', icon: Edit, onClick: () => onEditCustomer(customer) },
    { label: 'حذف', icon: Trash2, onClick: () => onDeleteCustomer(customer.id), danger: true },
  ];

  return (
    <div className="relative">
      <Button
        type="button"
        variant="ghost"
        size="icon"
        onClick={() => setOpen((prev) => !prev)}
        className="text-text-secondary hover:text-text-primary"
        aria-label="خيارات العميل"
        title="خيارات العميل"
      >
        <MoreHorizontal className="h-4 w-4" />
      </Button>

      {open && (
        <div className="absolute left-0 top-full z-20 mt-2 w-40 rounded-xl border border-border bg-surface-elevated p-1 shadow-[0_16px_36px_rgba(15,23,42,0.22)]">
          {actions.map(({ label, icon: Icon, onClick, danger }) => (
            <button
              key={label}
              type="button"
              onClick={() => {
                onClick();
                setOpen(false);
              }}
              className={cn(
                'flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-sm transition-colors',
                danger ? 'text-danger hover:bg-danger/8' : 'text-text-primary hover:bg-surface'
              )}
            >
              <Icon className="h-4 w-4" />
              <span>{label}</span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

export function CustomerList({
  filteredCustomers,
  isLoading,
  onViewCustomer,
  onEditCustomer,
  onDeleteCustomer,
}: CustomerListProps) {
  return (
    <div className="rounded-[12px] border border-border bg-surface shadow-[0_8px_18px_rgba(15,23,42,0.04)]">
      <div className="flex items-center gap-2 border-b border-border px-5 py-4">
        <Users className="h-4 w-4 text-primary" />
        <h3 className="text-sm font-semibold text-text-primary">قائمة العملاء</h3>
      </div>

      {isLoading ? (
        <div className="flex h-64 items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
        </div>
      ) : filteredCustomers.length === 0 ? (
        <EmptyState
          icon={<Inbox className="h-5 w-5" />}
          title="لا يوجد عملاء"
          description="ابدأ بإضافة عملاء جدد"
        />
      ) : (
        <>
          <div className="hidden md:block">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[12%]">الكود</TableHead>
                  <TableHead className="w-[20%]">الاسم</TableHead>
                  <TableHead className="w-[18%]">الهاتف</TableHead>
                  <TableHead className="w-[18%] text-center">مشتريات العميل (تراكمي)</TableHead>
                  <TableHead className="w-[18%] text-center">الديون</TableHead>
                  <TableHead className="w-[14%] text-end">الإجراءات</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredCustomers.map((customer: Customer) => (
                  <TableRow key={customer.id}>
                    <TableCell className="font-semibold text-text-primary">{customer.code}</TableCell>
                    <TableCell>
                      <div className="flex min-w-0 items-center gap-2">
                        <div className="flex h-8 w-8 items-center justify-center rounded-lg border border-border bg-surface-elevated text-text-secondary">
                          <UserRound className="h-4 w-4" />
                        </div>
                        <div className="min-w-0">
                          <div className="truncate font-semibold text-text-primary">{customer.name}</div>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2 text-text-secondary">
                        <Phone className="h-3.5 w-3.5" />
                        <span>{customer.phone}</span>
                      </div>
                    </TableCell>
                    <TableCell className="text-center font-semibold text-primary">₪{customer.totalPurchases?.toLocaleString() || 0}</TableCell>
                    <TableCell className="text-center">
                      <Badge variant={customer.outstanding > 0 ? 'danger' : 'success'} size="sm">
                        ₪{customer.outstanding?.toLocaleString() || 0}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-end">
                      <div className="flex items-center justify-end gap-2">
                        <Button type="button" variant="primary" size="sm" onClick={() => onViewCustomer(customer)} className="h-8 px-2.5 text-[11px]" aria-label="عرض العميل" title="عرض العميل">
                          <Eye className="h-3.5 w-3.5" />
                        </Button>
                        <RowActionMenu
                          customer={customer}
                          onViewCustomer={onViewCustomer}
                          onEditCustomer={onEditCustomer}
                          onDeleteCustomer={onDeleteCustomer}
                        />
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

          <div className="block md:hidden">
            <div className="grid gap-3 p-4">
              {filteredCustomers.map((customer: Customer) => (
                <div key={customer.id} className="rounded-xl border border-border bg-surface-elevated/25 p-4">
                  <div className="mb-3 flex items-start justify-between gap-3">
                    <div className="flex min-w-0 items-center gap-2">
                      <div className="flex h-9 w-9 items-center justify-center rounded-lg border border-border bg-surface text-text-secondary">
                        <UserRound className="h-4 w-4" />
                      </div>
                      <div className="min-w-0">
                        <div className="truncate font-semibold text-text-primary">{customer.name}</div>
                        <div className="text-[11px] text-text-tertiary">{customer.code}</div>
                      </div>
                    </div>
                    <Badge variant={customer.outstanding > 0 ? 'danger' : 'success'} size="sm">
                      {customer.outstanding > 0 ? 'مدين' : 'سليم'}
                    </Badge>
                  </div>

                  <div className="space-y-2 text-sm">
                    <div className="flex items-center justify-between gap-2">
                      <span className="text-text-tertiary">الهاتف</span>
                      <span className="font-medium text-text-secondary">{customer.phone}</span>
                    </div>
                    <div className="flex items-center justify-between gap-2">
                      <span className="text-text-tertiary">مشتريات العميل (تراكمي)</span>
                      <span className="font-semibold text-primary">₪{customer.totalPurchases?.toLocaleString() || 0}</span>
                    </div>
                    <div className="flex items-center justify-between gap-2">
                      <span className="text-text-tertiary">الديون</span>
                      <span className="font-semibold text-text-primary">₪{customer.outstanding?.toLocaleString() || 0}</span>
                    </div>
                  </div>

                  <div className="mt-3 flex items-center justify-between gap-2 border-t border-border pt-3">
                    <Button type="button" variant="primary" size="sm" onClick={() => onViewCustomer(customer)} className="h-8 px-3 text-[11px]">
                      <Eye className="h-3.5 w-3.5" />
                      عرض
                    </Button>
                    <RowActionMenu
                      customer={customer}
                      onViewCustomer={onViewCustomer}
                      onEditCustomer={onEditCustomer}
                      onDeleteCustomer={onDeleteCustomer}
                    />
                  </div>
                </div>
              ))}
            </div>
          </div>
        </>
      )}
    </div>
  );
}
