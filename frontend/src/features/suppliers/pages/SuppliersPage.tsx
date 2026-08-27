import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { useToast } from '../../../hooks/useToast';
import { useDebounce } from '../../../hooks/useDebounce';
import { suppliersApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { PageHeader } from '../../../components/ui/page-header';
import { LoadingSpinner } from '../../../components/ui/loading-spinner';
import { SupplierCard } from '../../../components/ui/supplier-card';
import { SupplierModals } from '../components/SupplierModals';
import { ConfirmDialog } from '../../../components/ui/confirm-dialog';
import type { SupplierFormData } from '../../../components/forms/SupplierForm';
import { Supplier } from '../../../types/models';
import { exportToCSV, printTable } from '../../../lib/export-utils';
import { getButtonSize } from '../../../config/button-sizes';
import {
  Truck,
  Search,
  Plus,
  Filter,
  DollarSign,
  Sparkles,
  Target,
  TrendingUp,
  AlertTriangle,
  Download,
  Printer,
  RefreshCw
} from 'lucide-react';

export function SuppliersPage() {
  const { t } = useTranslation();
  const { success: showSuccess, error: showError } = useToast();
  const [searchQuery, setSearchQuery] = useState('');
  const debouncedSearchQuery = useDebounce(searchQuery, 300);
  const [expandedSupplierId, setExpandedSupplierId] = useState<string | null>(null);
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [editingSupplier, setEditingSupplier] = useState<any | null>(null);
  const [isViewModalOpen, setIsViewModalOpen] = useState(false);
  const [viewingSupplier, setViewingSupplier] = useState<any | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [supplierToDelete, setSupplierToDelete] = useState<any | null>(null);

  const { data: suppliersData, isLoading, refetch } = useQuery({
    queryKey: ['suppliers', debouncedSearchQuery],
    queryFn: () => {
      if (debouncedSearchQuery) {
        return suppliersApi.list({
          page: 1,
          per_page: 50,
          search: debouncedSearchQuery
        });
      } else {
        return suppliersApi.list({ page: 1, per_page: 50 });
      }
    },
    enabled: true,
  });

  const { data: supplierInventory, isLoading: inventoryLoading } = useQuery({
    queryKey: ['supplier-inventory', expandedSupplierId],
    queryFn: () => suppliersApi.getSupplierInventory(expandedSupplierId!),
    enabled: !!expandedSupplierId,
  });

  const suppliers = (suppliersData?.data as Supplier[]) || [];

  const filteredSuppliers = suppliers.filter((supplier: Supplier) =>
    supplier.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    supplier.phone.includes(searchQuery) ||
    supplier.code.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const totalSuppliers = suppliers.length;
  const totalPurchases = suppliers.reduce((sum: number, s: Supplier) => sum + ('totalPurchases' in s ? (s as Record<string, unknown>).totalPurchases as number : 0), 0);
  const totalPaid = suppliers.reduce((sum: number, s: Supplier) => sum + ('paidAmount' in s ? (s as Record<string, unknown>).paidAmount as number : 0), 0);
  const totalOutstanding = suppliers.reduce((sum: number, s: Supplier) => sum + ('outstanding' in s ? (s as Record<string, unknown>).outstanding as number : 0), 0);

  const handleExport = () => {
    const dataToExport = filteredSuppliers.map((supplier: any) => ({
      'الاسم': supplier.name,
      'الهاتف': supplier.phone,
      'البريد': supplier.email || '-',
      'المشتريات': supplier.totalPurchases,
      'المدفوع': supplier.paidAmount,
      'المستحق': supplier.outstanding
    }));
    exportToCSV(dataToExport, `suppliers-${new Date().toISOString().split('T')[0]}`);
  };

  const handlePrint = () => {
    const dataToPrint = filteredSuppliers.map((supplier: any) => ({
      'الاسم': supplier.name,
      'الهاتف': supplier.phone,
      'البريد': supplier.email || '-',
      'المشتريات': supplier.totalPurchases,
      'المدفوع': supplier.paidAmount,
      'المستحق': supplier.outstanding
    }));
    printTable(dataToPrint, ['الاسم', 'الهاتف', 'البريد', 'المشتريات', 'المدفوع', 'المستحق'], 'تقرير الموردين');
  };

  const handleSubmitSupplier = async (data: SupplierFormData) => {
    try {
      if (editingSupplier) {
        await suppliersApi.update(editingSupplier.id, data);
        showSuccess('تم تحديث المورد بنجاح');
      } else {
        await suppliersApi.create(data);
        showSuccess('تمت إضافة المورد بنجاح');
      }
      setIsAddModalOpen(false);
      setEditingSupplier(null);
      refetch();
    } catch (err) {
      showError('حدث خطأ أثناء حفظ المورد');
    }
  };

  const handleDeleteSupplier = async (supplier: any) => {
    setSupplierToDelete(supplier);
    setDeleteDialogOpen(true);
  };

  const handleConfirmDelete = async () => {
    if (supplierToDelete) {
      try {
        await suppliersApi.delete(supplierToDelete.id);
        showSuccess('تم حذف المورد بنجاح');
        setDeleteDialogOpen(false);
        setSupplierToDelete(null);
        refetch();
      } catch (err) {
        showError('حدث خطأ أثناء حذف المورد');
      }
    }
  };

  return (
    <div>
      {/* Page Header */}
      <PageHeader
        eyebrow="Supplier Hub"
        title={t('suppliers.title')}
        description="إدارة الموردين والمشتريات مع رؤى ذكية"
        actions={
          <div style={{ display: 'flex', gap: '10px' }}>
            <Button variant="primary" size={getButtonSize('suppliers', 'headerActions')} onClick={() => setIsAddModalOpen(true)} className="gap-2">
              <Plus className="w-4 h-4" />
              {t('suppliers.addSupplier')}
            </Button>
            <Button variant="secondary" size={getButtonSize('suppliers', 'headerActions')} onClick={handleExport} className="gap-2">
              <Download className="w-4 h-4" />
              تصدير
            </Button>
            <Button variant="secondary" size={getButtonSize('suppliers', 'headerActions')} onClick={handlePrint} className="gap-2">
              <Printer className="w-4 h-4" />
              طباعة
            </Button>
          </div>
        }
      />

      {/* Stats Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '14px' }}
           className="grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-small text-text-muted">إجمالي الموردين</p>
                <p className="text-h2 font-bold text-text mt-1">
                  {totalSuppliers}
                </p>
              </div>
              <Truck className="w-5 h-5 text-text-muted" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-small text-text-muted">إجمالي المشتريات</p>
                <p className="text-h2 font-bold text-text mt-1">
                  ₪{totalPurchases.toLocaleString()}
                </p>
              </div>
              <DollarSign className="w-5 h-5 text-text-muted" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-small text-text-muted">المدفوع</p>
                <p className="text-h2 font-bold text-text mt-1">
                  ₪{totalPaid.toLocaleString()}
                </p>
              </div>
              <DollarSign className="w-5 h-5 text-text-muted" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-small text-text-muted">المستحق</p>
                <p className="text-h2 font-bold text-text mt-1">
                  ₪{totalOutstanding.toLocaleString()}
                </p>
              </div>
              <DollarSign className="w-5 h-5 text-text-muted" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* AI Insights */}
      <Card variant="ai">
        <CardHeader>
          <CardTitle style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <Sparkles style={{ width: '20px', height: '20px', color: 'var(--color-primary)' }} />
            AI Insights - الموردين
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
            <div style={{ display: 'flex', gap: '14px' }}>
              <div style={{
                width: '32px',
                height: '32px',
                borderRadius: '8px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                background: 'rgba(34, 211, 238, 0.1)',
                flexShrink: 0
              }}>
                <Target style={{ width: '16px', height: '16px', color: 'var(--color-primary)' }} />
              </div>
              <div>
                <p style={{ fontSize: '13px', fontWeight: '600', color: 'var(--text-primary)' }}>فرصة تحسين التوريد</p>
                <p style={{ fontSize: '11px', color: 'var(--text-secondary)', marginTop: '4px' }}>
                  المورد "إلكترونيات المتقدمة" يقدم أسعاراً أقل 15% من المنافسين مع جودة مماثلة. يُنصح بزيادة حجم التعامل.
                </p>
              </div>
            </div>
            <div style={{ display: 'flex', gap: '14px' }}>
              <div style={{
                width: '32px',
                height: '32px',
                borderRadius: '8px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                background: 'rgba(251, 191, 36, 0.1)',
                flexShrink: 0
              }}>
                <AlertTriangle style={{ width: '16px', height: '16px', color: 'var(--color-warning)' }} />
              </div>
              <div>
                <p style={{ fontSize: '13px', fontWeight: '600', color: 'var(--text-primary)' }}>تنبيه تأخير التوريد</p>
                <p style={{ fontSize: '11px', color: 'var(--text-secondary)', marginTop: '4px' }}>
                  المورد "شاشات المستقبل" تأخر 3 مرات هذا الشهر. يُنصح بالبحث عن بدائل أو تقييم العقد.
                </p>
              </div>
            </div>
            <div style={{ display: 'flex', gap: '14px' }}>
              <div style={{
                width: '32px',
                height: '32px',
                borderRadius: '8px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                background: 'rgba(52, 211, 153, 0.1)',
                flexShrink: 0
              }}>
                <TrendingUp style={{ width: '16px', height: '16px', color: 'var(--color-success)' }} />
              </div>
              <div>
                <p style={{ fontSize: '13px', fontWeight: '600', color: 'var(--text-primary)' }}>أداء ممتاز</p>
                <p style={{ fontSize: '11px', color: 'var(--text-secondary)', marginTop: '4px' }}>
                  المورد "بطاريات القوة" حقق 100% من مواعيد التسليم هذا الربع. يُنصح بتجديد العقد تلقائياً.
                </p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Search and Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex flex-col md:flex-row gap-4">
            <div className="flex-1 relative">
              <Search className="absolute inset-y-0 right-3 w-4 h-4 text-text-muted" />
              <Input
                placeholder={t('common.search')}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pr-10"
              />
            </div>
            <div className="flex gap-2">
              <Button variant="outline" size={getButtonSize('suppliers', 'headerActions')} className="gap-2">
                <Filter className="w-4 h-4" />
                {t('common.filter')}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Suppliers Grid */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between gap-3">
            <CardTitle>قائمة الموردين ({filteredSuppliers.length})</CardTitle>
            <Button
              variant="outline"
              size={getButtonSize('suppliers', 'headerActions')}
              onClick={() => refetch()}
              className="gap-2"
            >
              <RefreshCw className="w-4 h-4" />
              تحديث
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center h-64">
              <LoadingSpinner size="md" color="cyan" />
            </div>
          ) : filteredSuppliers.length === 0 ? (
            <div className="text-center py-8 text-text-muted">
              لا يوجد موردين
            </div>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-4">
              {filteredSuppliers.map((supplier: any) => (
                <SupplierCard
                  key={supplier.id}
                  supplier={supplier}
                  expanded={expandedSupplierId === supplier.id}
                  inventory={expandedSupplierId === supplier.id ? supplierInventory : undefined}
                  inventoryLoading={inventoryLoading && expandedSupplierId === supplier.id}
                  onToggle={(id) =>
                    setExpandedSupplierId(expandedSupplierId === id ? null : (id as string))
                  }
                  onEdit={(s) => {
                    setEditingSupplier(s);
                    setIsAddModalOpen(true);
                  }}
                  onView={(s) => {
                    setViewingSupplier(s);
                    setIsViewModalOpen(true);
                  }}
                  onDelete={handleDeleteSupplier}
                />
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <SupplierModals
        isOpen={isAddModalOpen}
        setIsOpen={setIsAddModalOpen}
        editingSupplier={editingSupplier}
        setEditingSupplier={setEditingSupplier}
        onSubmit={handleSubmitSupplier}
        isViewModalOpen={isViewModalOpen}
        setIsViewModalOpen={setIsViewModalOpen}
        viewingSupplier={viewingSupplier}
      />

      {/* Delete Confirmation Dialog */}
      <ConfirmDialog
        isOpen={deleteDialogOpen}
        onClose={() => {
          setDeleteDialogOpen(false);
          setSupplierToDelete(null);
        }}
        onConfirm={handleConfirmDelete}
        title="حذف المورد"
        message="هل أنت متأكد من حذف هذا المورد؟ هذا الإجراء لا يمكن التراجع عنه."
        confirmText="حذف المورد"
        cancelText="إلغاء"
        variant="danger"
      />
    </div>
  );
}
