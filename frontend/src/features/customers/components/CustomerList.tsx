import { Button } from '../../../components/ui/button';
import { EmptyState } from '../../../components/ui/empty-state';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
import { Badge } from '../../../components/ui/badge';
import { Users, Eye, Edit, Trash2, Phone, UserRound, Inbox } from 'lucide-react';
import { ActionMenu } from '../../../components/ui/action-menu';
import { Customer } from '../types/customers.types';

interface CustomerListProps {
  filteredCustomers: Customer[];
  isLoading: boolean;
  onViewCustomer: (customer: Customer) => void;
  onEditCustomer: (customer: Customer) => void;
  onDeleteCustomer: (customerId: string) => void;
}

export function CustomerList({
  filteredCustomers,
  isLoading,
  onViewCustomer,
  onEditCustomer,
  onDeleteCustomer,
}: CustomerListProps) {
  return (
    <div className="rounded-[16px] border border-[var(--border-default)] bg-[var(--bg-surface)] shadow-[0_10px_28px_rgba(15,23,42,0.05)]">
      <div className="flex items-center justify-between gap-2 border-b border-[var(--border-subtle)] px-5 py-4">
        <div className="flex items-center gap-2">
          <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-[var(--color-primary-10)] text-[var(--primary)]">
            <Users className="h-4 w-4" />
          </span>
          <div>
            <h3 className="text-sm font-extrabold text-[var(--text-primary)]">قائمة العملاء</h3>
            <span className="text-[11px] font-medium text-[var(--text-muted)]">{filteredCustomers.length} عميل</span>
          </div>
        </div>
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
        <div className="overflow-x-auto">
          <Table className="min-w-[760px]">
            <TableHeader>
              <TableRow>
                <TableHead className="w-[13%]">الكود</TableHead>
                <TableHead className="w-[25%]">العميل</TableHead>
                <TableHead className="w-[20%]">الهاتف</TableHead>
                <TableHead className="w-[16%] text-center">المشتريات</TableHead>
                <TableHead className="w-[14%] text-center">المستحق</TableHead>
                <TableHead className="w-[12%] text-end">إجراءات</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody className="divide-y divide-[var(--border-subtle)]" dir="rtl">
              {filteredCustomers.map((customer: Customer) => (
                <TableRow
                  key={customer.id}
                  dir="rtl"
                  className="cursor-pointer transition-all duration-200 hover:bg-[var(--color-primary-05)] [&>td]:h-[68px]"
                  onClick={() => onViewCustomer(customer)}
                  title="فتح تفاصيل العميل"
                >
                  <TableCell className="pf-customer-code font-mono text-xs font-semibold text-[var(--text-secondary)]">
                    <span className="inline-flex items-center gap-1 rounded-full border border-[var(--border-subtle)] bg-[var(--color-primary-10)] px-2.5 py-1 text-[11px] font-extrabold text-[var(--primary)]">
                      <span className="h-1.5 w-1.5 rounded-full bg-[var(--primary)]" />
                      {customer.code}
                    </span>
                  </TableCell>
                  <TableCell>
                    <div className="flex min-w-0 items-center gap-3">
                      <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl border border-[var(--border-subtle)] bg-[var(--color-primary-10)] text-[var(--primary)] shadow-sm">
                        <UserRound className="h-4 w-4" />
                      </div>
                      <div className="min-w-0">
                        <div className="pf-customer-name truncate font-black text-[var(--text-primary)]">{customer.name}</div>
                        <div className="mt-0.5 truncate text-[11px] font-medium text-[var(--text-tertiary)]">عميل مسجل</div>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell>
                    <div dir="ltr" className="pf-customer-phone flex items-center justify-end gap-2 text-sm font-semibold text-[var(--text-secondary)]">
                      <Phone className="h-3.5 w-3.5 text-[var(--text-muted)]" />
                      <span>{customer.phone}</span>
                    </div>
                  </TableCell>
                  <TableCell className="pf-customer-total text-center font-black text-[var(--primary)]">₪{customer.totalPurchases?.toLocaleString() || 0}</TableCell>
                  <TableCell className="text-center">
                    <Badge variant={customer.outstanding > 0 ? 'danger' : 'success'} size="sm" className="min-w-[62px] justify-center rounded-full">
                      {customer.outstanding > 0 ? `₪${customer.outstanding.toLocaleString()}` : 'سليم'}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-end">
                    <div className="flex items-center justify-end gap-1">
                      <Button type="button" variant="primary" size="sm" onClick={(event) => { event.stopPropagation(); onViewCustomer(customer); }} className="h-8 px-2.5 text-[11px]" aria-label="عرض العميل" title="عرض العميل">
                        <Eye className="h-3.5 w-3.5" />
                      </Button>
                      <ActionMenu
                        label="خيارات العميل"
                        widthClassName="w-48"
                        items={[
                          { label: 'عرض', icon: Eye, onClick: () => onViewCustomer(customer) },
                          { label: 'تعديل', icon: Edit, onClick: () => onEditCustomer(customer) },
                          { label: 'حذف', icon: Trash2, onClick: () => onDeleteCustomer(customer.id), danger: true },
                        ]}
                      />
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  );
}
