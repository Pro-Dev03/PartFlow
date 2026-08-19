import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { customersApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { PageHeader } from '../../../components/ui/page-header';
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
  TrendingUp,
  TrendingDown,
  Sparkles,
  UserPlus,
  Heart,
  Shield
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
    <div className="space-y-md">
      {/* Page Header - Futuristic + Clean */}
      <PageHeader
        eyebrow="Customer Hub"
        title={t('customers.title')}
        description="إدارة العملاء والديون مع تجربة نظيفة ومبتكرة"
        actions={
          <div className="flex items-center gap-sm">
            <Button variant="primary" className="gap-2" onClick={handleAddCustomer}>
              <UserPlus className="w-4 h-4" />
              {t('customers.addCustomer')}
            </Button>
            <Button variant="secondary" className="gap-2">
              <Heart className="w-4 h-4" />
              وفاء
            </Button>
          </div>
        }
      />

      {/* AI Customer Insight - Futuristic + Clean */}
      <Card variant="ai">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Sparkles className="w-5 h-5 text-cyan" />
            AI Customer Insight
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-md">
            <div className="w-8 h-8 rounded-sm bg-cyan/10 flex items-center justify-center flex-shrink-0">
              <Heart className="w-4 h-4 text-cyan" />
            </div>
            <div>
              <p className="text-small font-semibold text-text">العملاء الأكثر وفاءً</p>
              <p className="text-tiny text-text-muted mt-1">
                لديك 5 عملاء حققوا 80% من إجمالي المبيعات هذا الشهر. يمكنهم الاستفادة من برنامج الولاء.
              </p>
              <Button variant="ghost" size="sm" className="mt-2 text-cyan">
                عرض التوصية ←
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Stats Cards - Futuristic + Clean */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-md">
        <StatCard 
          title="إجمالي العملاء" 
          value={totalCustomers} 
          icon={Users}
          subtitle="جميع العملاء المسجلين"
          variant="featured"
          trend="+12%"
          trendUp={true}
        />
        <StatCard 
          title="العملاء النشطين" 
          value={activeCustomers} 
          icon={TrendingUp}
          subtitle="لديهم مشتريات"
          variant="default"
          trend="+8%"
          trendUp={true}
        />
        <StatCard 
          title="عملاء لديهم ديون" 
          value={customersWithDebt} 
          icon={DollarSign}
          subtitle="ديون مستحقة"
          variant={customersWithDebt > 0 ? 'warning' : 'success'}
          trend={customersWithDebt > 0 ? '+2' : null}
          trendUp={false}
        />
        <StatCard 
          title="إجمالي المستحق" 
          value={`₪${totalOutstanding.toLocaleString()}`} 
          icon={Shield}
          subtitle="مجموع الديون"
          variant={totalOutstanding > 0 ? 'warning' : 'success'}
          trend={totalOutstanding > 0 ? '+5%' : null}
          trendUp={false}
        />
      </div>

      {/* Search and Filters - Futuristic + Clean */}
      <Card>
        <CardContent className="p-lg">
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

      {/* Customers Table - Futuristic + Clean */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Users className="w-5 h-5 text-cyan" />
            قائمة العملاء
          </CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center h-64">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-cyan" />
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
                  <TableHead className="text-start">الإجراءات</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredCustomers.map((customer: any) => (
                  <TableRow key={customer.id}>
                    <TableCell className="font-medium text-text">{customer.name}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-sm">
                        <Phone className="w-4 h-4 text-cyan" />
                        <span className="text-small text-text-muted">{customer.phone}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      {customer.email ? (
                        <div className="flex items-center gap-sm">
                          <Mail className="w-4 h-4 text-cyan" />
                          <span className="text-small text-text-muted">{customer.email}</span>
                        </div>
                      ) : (
                        <span className="text-tiny text-text-muted">-</span>
                      )}
                    </TableCell>
                    <TableCell className="text-small text-text">₪{customer.totalPurchases?.toLocaleString() || 0}</TableCell>
                    <TableCell>
                      <Badge variant={customer.outstanding > 0 ? 'danger' : 'success'} size="sm">
                        ₪{customer.outstanding?.toLocaleString() || 0}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-small text-text-muted">
                      {customer.lastPurchase 
                        ? new Date(customer.lastPurchase).toLocaleDateString('ar-SA')
                        : '-'
                      }
                    </TableCell>
                    <TableCell className="text-start">
                      <div className="flex gap-sm">
                        <Button variant="ghost" size="sm">
                          <Eye className="w-4 h-4" />
                        </Button>
                        <Button variant="ghost" size="sm" onClick={() => handleEditCustomer(customer)}>
                          <Edit className="w-4 h-4" />
                        </Button>
                        {customer.outstanding > 0 && (
                          <Button variant="ghost" size="sm" className="text-green">
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
  subtitle?: string;
  variant?: 'default' | 'featured' | 'warning' | 'ai' | 'danger' | 'success' | 'info';
  trend?: string | null;
  trendUp?: boolean | null;
}

function StatCard({ title, value, icon: Icon, subtitle, variant = 'default', trend, trendUp }: StatCardProps) {
  return (
    <Card 
      variant={variant} 
      className="hover:border-border/22 hover:-translate-y-1 cursor-pointer"
      hoverable
    >
      <CardContent className="p-lg">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <p className="text-small text-text-muted">{title}</p>
            <p className="text-metric font-bold text-text mt-1">
              {value}
            </p>
            {subtitle && (
              <p className="text-tiny text-text-muted mt-1">
                {subtitle}
              </p>
            )}
            {trend && (
              <div className="flex items-center gap-1 mt-2">
                {trendUp ? (
                  <TrendingUp className="w-3 h-3 text-green" />
                ) : (
                  <TrendingDown className="w-3 h-3 text-red" />
                )}
                <span className={`text-tiny ${trendUp ? 'text-green' : 'text-red'}`}>
                  {trend}
                </span>
              </div>
            )}
          </div>
          <div className={`w-10 h-10 rounded-sm flex items-center justify-center ${
            variant === 'featured' ? 'bg-cyan/10' :
            variant === 'warning' ? 'bg-yellow/10' :
            variant === 'danger' ? 'bg-red/10' :
            variant === 'success' ? 'bg-green/10' :
            variant === 'info' ? 'bg-cyan/10' :
            variant === 'ai' ? 'bg-cyan/10' :
            'bg-cyan/10'
          }`}>
            <Icon className={`w-5 h-5 ${
              variant === 'featured' ? 'text-cyan' :
              variant === 'warning' ? 'text-yellow' :
              variant === 'danger' ? 'text-red' :
              variant === 'success' ? 'text-green' :
              variant === 'info' ? 'text-cyan' :
              variant === 'ai' ? 'text-cyan' :
              'text-cyan'
            }`} />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}