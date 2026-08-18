import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { customersApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
import { Badge } from '../../../components/ui/badge';
import { Modal } from '../../../components/ui/modal';
import { CustomerForm, type CustomerFormData } from '../../../components/forms/CustomerForm';
import { AdvancedSearch } from '../../../components/ui/advanced-search';
import {
  Users,
  Plus,
  Eye,
  Edit,
  Phone,
  Mail,
  DollarSign,
  TrendingUp
} from 'lucide-react';

export function CustomersPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingCustomer, setEditingCustomer] = useState<any>(null);

  const { data: customersData, isLoading } = useQuery({
    queryKey: ['customers'],
    queryFn: () => customersApi.list(),
  });

  const createMutation = useMutation({
    mutationFn: (data: CustomerFormData) => customersApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['customers'] });
      setIsModalOpen(false);
      setEditingCustomer(null);
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: CustomerFormData }) =>
      customersApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['customers'] });
      setIsModalOpen(false);
      setEditingCustomer(null);
    },
  });

  const handleAddCustomer = () => {
    setEditingCustomer(null);
    setIsModalOpen(true);
  };

  const handleEditCustomer = (customer: any) => {
    setEditingCustomer(customer);
    setIsModalOpen(true);
  };

  const handleSubmit = (data: CustomerFormData) => {
    if (editingCustomer) {
      updateMutation.mutate({ id: editingCustomer.id, data });
    } else {
      createMutation.mutate(data);
    }
  };

  const customers = (customersData?.data as any[]) || [];

  const filteredCustomers = customers.filter((customer: any) =>
    customer.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    customer.phone.includes(searchQuery)
  );

  const totalCustomers = customers.length;
  const activeCustomers = customers.filter((c: any) => c.totalPurchases > 0).length;
  const customersWithDebt = customers.filter((c: any) => c.outstanding > 0).length;
  const totalOutstanding = customers.reduce((sum: number, c: any) => sum + c.outstanding, 0);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            {t('customers.title')}
          </h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">
            إدارة العملاء والديون
          </p>
        </div>
        <Button className="gap-2" onClick={handleAddCustomer}>
          <Plus className="w-4 h-4" />
          {t('customers.addCustomer')}
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <StatCard title="إجمالي العملاء" value={totalCustomers} icon={Users} />
        <StatCard title="العملاء النشطين" value={activeCustomers} icon={TrendingUp} />
        <StatCard title="عملاء لديهم ديون" value={customersWithDebt} icon={DollarSign} />
        <StatCard title="إجمالي المستحق" value={`₪${totalOutstanding.toLocaleString()}`} icon={DollarSign} />
      </div>

      {/* Search and Filters */}
      <Card>
        <CardContent className="p-4">
          <AdvancedSearch
            onSearch={(query, filters) => {
              setSearchQuery(query);
              // Handle filters if needed
            }}
            filters={[
              {
                key: 'hasDebt',
                label: 'الحالة',
                type: 'select',
                options: [
                  { value: 'all', label: 'الكل' },
                  { value: 'hasDebt', label: 'لديهم ديون' },
                  { value: 'noDebt', label: 'بدون ديون' },
                ],
              },
              {
                key: 'sortBy',
                label: 'الترتيب',
                type: 'select',
                options: [
                  { value: 'name', label: 'الاسم' },
                  { value: 'totalPurchases', label: 'إجمالي المشتريات' },
                  { value: 'lastPurchase', label: 'آخر شراء' },
                ],
              },
            ]}
            placeholder={t('common.search')}
          />
        </CardContent>
      </Card>

      {/* Customers Table */}
      <Card>
        <CardHeader>
          <CardTitle>قائمة العملاء</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center h-64">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600" />
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>الاسم</TableHead>
                  <TableHead>الهاتف</TableHead>
                  <TableHead>البريد الإلكتروني</TableHead>
                  <TableHead>إجمالي المشتريات</TableHead>
                  <TableHead>المستحق</TableHead>
                  <TableHead>آخر شراء</TableHead>
                  <TableHead className="text-left">الإجراءات</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredCustomers.map((customer: any) => (
                  <TableRow key={customer.id}>
                    <TableCell className="font-medium">{customer.name}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Phone className="w-4 h-4 text-gray-400" />
                        {customer.phone}
                      </div>
                    </TableCell>
                    <TableCell>
                      {customer.email ? (
                        <div className="flex items-center gap-2">
                          <Mail className="w-4 h-4 text-gray-400" />
                          {customer.email}
                        </div>
                      ) : (
                        '-'
                      )}
                    </TableCell>
                    <TableCell>₪{customer.totalPurchases?.toLocaleString() || 0}</TableCell>
                    <TableCell>
                      <Badge variant={customer.outstanding > 0 ? 'destructive' : 'default'}>
                        ₪{customer.outstanding?.toLocaleString() || 0}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {customer.lastPurchase 
                        ? new Date(customer.lastPurchase).toLocaleDateString('ar-SA')
                        : '-'
                      }
                    </TableCell>
                    <TableCell className="text-left">
                      <div className="flex gap-2">
                        <Button variant="ghost" size="sm">
                          <Eye className="w-4 h-4" />
                        </Button>
                        <Button variant="ghost" size="sm" onClick={() => handleEditCustomer(customer)}>
                          <Edit className="w-4 h-4" />
                        </Button>
                        {customer.outstanding > 0 && (
                          <Button variant="ghost" size="sm" className="text-green-600">
                            <DollarSign className="w-4 h-4" />
                          </Button>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Customer Form Modal */}
      <Modal
        isOpen={isModalOpen}
        onClose={() => {
          setIsModalOpen(false);
          setEditingCustomer(null);
        }}
        title={editingCustomer ? 'تعديل العميل' : 'إضافة عميل جديد'}
      >
        <CustomerForm
          onSubmit={handleSubmit}
          onCancel={() => {
            setIsModalOpen(false);
            setEditingCustomer(null);
          }}
          initialData={editingCustomer}
        />
      </Modal>
    </div>
  );
}

interface StatCardProps {
  title: string;
  value: number | string;
  icon: any;
}

function StatCard({ title, value, icon: Icon }: StatCardProps) {
  return (
    <Card>
      <CardContent className="p-6">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm text-gray-500 dark:text-gray-400">{title}</p>
            <p className="text-2xl font-bold text-gray-900 dark:text-gray-100 mt-1">
              {value}
            </p>
          </div>
          <Icon className="w-5 h-5 text-gray-400" />
        </div>
      </CardContent>
    </Card>
  );
}