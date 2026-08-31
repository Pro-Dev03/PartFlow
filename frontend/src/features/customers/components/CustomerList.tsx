import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Badge } from '../../../components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
import { TableCard, TableCardItem, TableCardActions, ResponsiveTable } from '../../../components/ui/table-card';
import { getButtonSize } from '../../../config/button-sizes';
import { Users, Phone, Mail, Eye, Edit, Trash2 } from 'lucide-react';
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
    <Card>
      <CardHeader>
        <CardTitle style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <Users style={{ width: '20px', height: '20px', color: 'var(--primary)' }} />
          قائمة العملاء
        </CardTitle>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '256px' }}>
            <div style={{ animation: 'spin 1s linear infinite', borderRadius: '50%', height: '32px', width: '32px', borderBottom: '2px solid var(--primary)' }} />
          </div>
        ) : filteredCustomers.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '40px 20px' }}>
            <Users style={{ width: '48px', height: '48px', color: 'var(--text-secondary)', margin: '0 auto 16px' }} />
            <p style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
              لا يوجد عملاء
            </p>
            <p style={{ fontSize: '11px', color: 'var(--text-tertiary)', marginTop: '4px' }}>
              ابدأ بإضافة عملاء جدد
            </p>
          </div>
        ) : (
          <>
            {/* Desktop Table */}
            <div className="hidden md:block">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>الكود</TableHead>
                    <TableHead>الاسم</TableHead>
                    <TableHead>الهاتف</TableHead>
                    <TableHead>البريد</TableHead>
                    <TableHead>إجمالي المشتريات</TableHead>
                    <TableHead>الديون</TableHead>
                    <TableHead>الحالة</TableHead>
                    <TableHead style={{ textAlign: 'right' }}>الإجراءات</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredCustomers.map((customer: Customer) => (
                    <TableRow key={customer.id}>
                      <TableCell style={{ fontWeight: '500' }}>{customer.code}</TableCell>
                      <TableCell style={{ fontWeight: '500' }}>{customer.name}</TableCell>
                      <TableCell>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                          <Phone style={{ width: '14px', height: '14px', color: 'var(--text-secondary)' }} />
                          {customer.phone}
                        </div>
                      </TableCell>
                      <TableCell>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                          <Mail style={{ width: '14px', height: '14px', color: 'var(--text-secondary)' }} />
                          {customer.email || '-'}
                        </div>
                      </TableCell>
                      <TableCell><span className="numeric-price">₪{customer.totalPurchases?.toLocaleString() || 0}</span></TableCell>
                      <TableCell>
                        <span style={{ color: customer.outstanding > 0 ? '#ef4444' : '#10b981', fontWeight: '500' }}>
                          <span className="numeric-price">₪{customer.outstanding?.toLocaleString() || 0}</span>
                        </span>
                      </TableCell>
                      <TableCell>
                        <Badge variant={customer.is_active === false ? 'danger' : customer.outstanding > 0 ? 'warning' : 'success'}>
                          {customer.is_active === false ? 'غير نشط' : 'نشط'}
                        </Badge>
                      </TableCell>
                      <TableCell style={{ textAlign: 'right' }}>
                        <div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
                          <Button
                            variant="ghost"
                            size={getButtonSize('customers', 'iconAction')}
                            onClick={() => onViewCustomer(customer)}
                          >
                            <Eye className="w-3.5 h-3.5" />
                          </Button>
                          <Button variant="ghost" size={getButtonSize('customers', 'iconAction')} onClick={() => onEditCustomer(customer)}>
                            <Edit className="w-3.5 h-3.5" />
                          </Button>
                          <Button
                            variant="ghost"
                            size={getButtonSize('customers', 'iconAction')}
                            onClick={() => onDeleteCustomer(customer.id)}
                          >
                            <Trash2 className="w-3.5 h-3.5" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>

            {/* Mobile Cards */}
            <ResponsiveTable>
              <div className="grid gap-3">
                {filteredCustomers.map((customer: Customer) => (
                  <TableCard key={customer.id}>
                    <TableCardItem label="الكود" value={customer.code} variant="highlight" />
                    <TableCardItem label="الاسم" value={customer.name} />
                    <TableCardItem label="الهاتف" value={customer.phone} />
                    <TableCardItem label="البريد" value={customer.email || '-'} />
                    <TableCardItem label="المشتريات" value={<span className="numeric-price">₪{customer.totalPurchases?.toLocaleString() || 0}</span>} variant="success" />
                    <TableCardItem label="الديون" value={<span className="numeric-price">₪{customer.outstanding?.toLocaleString() || 0}</span>} variant={customer.outstanding > 0 ? 'danger' : 'success'} />
                    <TableCardItem label="الحالة" value={<Badge variant={customer.is_active === false ? 'danger' : customer.outstanding > 0 ? 'warning' : 'success'}>{customer.is_active === false ? 'غير نشط' : 'نشط'}</Badge>} />
                    <TableCardActions>
                      <Button
                        variant="ghost"
                        onClick={() => onViewCustomer(customer)}
                        className="flex-1"
                      >
                        <Eye className="w-3.5 h-3.5 me-1" />
                        عرض
                      </Button>
                      <Button
                        variant="ghost"
                        onClick={() => onEditCustomer(customer)}
                        className="flex-1"
                      >
                        <Edit className="w-3.5 h-3.5 me-1" />
                        تعديل
                      </Button>
                      <Button
                        variant="ghost"
                        onClick={() => onDeleteCustomer(customer.id)}
                        className="flex-1"
                      >
                        <Trash2 className="w-3.5 h-3.5 me-1" />
                        حذف
                      </Button>
                    </TableCardActions>
                  </TableCard>
                ))}
              </div>
            </ResponsiveTable>
          </>
        )}
      </CardContent>
    </Card>
  );
}
