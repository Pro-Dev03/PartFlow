import { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from '../../../hooks/useTranslation';
import { useToast } from '../../../hooks/useToast';
import { useDebounce } from '../../../hooks/useDebounce';
import { suppliersApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../design-system/components/card';
import { Button } from '../../../design-system/components/button';
import { Input } from '../../../design-system/components/input';
import { PageHeader } from '../../../design-system/components/page-header';
import { EmptyState } from '../../../design-system/components/empty-state';
import { LoadingSpinner } from '../../../design-system/components/loading-spinner';
import { SupplierCard } from '../../../design-system/components/supplier-card';
import { SupplierModals } from '../components/SupplierModals';
import { ConfirmDialog } from '../../../design-system/components/confirm-dialog';
import { PaginationControls } from '../../../design-system/components/pagination-controls';
import { SortButton } from '../../../design-system/components/sort-button';
import type { SupplierFormData } from '../../../components/forms/SupplierForm';
import { Supplier } from '../../../types/models';
import { exportToCSV, printTable } from '../../../lib/export-utils';
import { ReportActions } from '../../../design-system/components/report-actions';
import { StatCard } from '../../../design-system/components/stat-card';
import { getButtonSize } from '../../../config/button-sizes';
import { normalizeSupplier } from '../utils/supplier-normalization';
import {
  Truck,
  Search,
  Plus,
  ShoppingCart,
  Filter,
  DollarSign,
  RefreshCw
  , RotateCcw
} from 'lucide-react';

export function SuppliersPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
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
  const [showOutstandingOnly, setShowOutstandingOnly] = useState(false);
  const [sortBy, setSortBy] = useState<'recent' | 'outstanding' | 'purchases'>('recent');
  const [showInactive, setShowInactive] = useState(false);
  const [page, setPage] = useState(1);
  const pageSize = 10;

  const { data: suppliersData, isLoading, refetch } = useQuery({
    queryKey: ['suppliers', page, pageSize, debouncedSearchQuery, showInactive],
    queryFn: () => suppliersApi.list({
      page,
      per_page: pageSize,
      ...(debouncedSearchQuery ? { search: debouncedSearchQuery } : {}),
      is_active: !showInactive,
    }),
    enabled: true,
  });

  useEffect(() => {
    setPage(1);
  }, [debouncedSearchQuery, showInactive]);

  const { data: supplierInventory, isLoading: inventoryLoading } = useQuery({
    queryKey: ['supplier-inventory', expandedSupplierId],
    queryFn: () => suppliersApi.getSupplierInventory(expandedSupplierId!),
    enabled: !!expandedSupplierId,
  });

  const suppliers = Array.isArray(suppliersData?.data)
    ? (suppliersData.data as any[]).map((supplier) => normalizeSupplier(supplier))
    : [];

  const filteredSuppliers = suppliers.filter((supplier: any) =>
    (supplier.name || '').toLowerCase().includes(searchQuery.toLowerCase()) ||
    (supplier.phone || '').includes(searchQuery) ||
    (supplier.code || '').toLowerCase().includes(searchQuery.toLowerCase())
  ).filter((supplier: any) => !showOutstandingOnly || Number(supplier.outstanding || 0) > 0)
    .sort((a: any, b: any) => {
      if (sortBy === 'outstanding') return Number(b.outstanding || 0) - Number(a.outstanding || 0);
      if (sortBy === 'purchases') return Number(b.totalPurchases || 0) - Number(a.totalPurchases || 0);
      return 0;
    });

  const totalSuppliers = Number(suppliersData?.meta?.total || suppliers.length);
  const totalPurchases = suppliers.reduce((sum: number, s: any) => sum + Number(s.totalPurchases || 0), 0);
  const totalPaid = suppliers.reduce((sum: number, s: any) => sum + Number(s.paidAmount || 0), 0);
  const totalOutstanding = suppliers.reduce((sum: number, s: any) => sum + Number(s.outstanding || 0), 0);

  const getSupplierReportRows = (supplierRows: any[]) => supplierRows.map((supplier: any) => ({
      'الاسم': supplier.name,
      'الهاتف': supplier.phone,
      'البريد': supplier.email || '-',
      'المشتريات': supplier.totalPurchases,
      'المدفوع': supplier.paidAmount,
      'صافي المستحق': supplier.outstanding
    }));

  const handleExport = () => {
    exportToCSV(getSupplierReportRows(filteredSuppliers), `suppliers-${new Date().toISOString().split('T')[0]}`);
  };

  const handlePrint = () => {
    printTable(getSupplierReportRows(filteredSuppliers), ['الاسم', 'الهاتف', 'البريد', 'المشتريات', 'المدفوع', 'صافي المستحق'], 'تقرير التجار');
  };

  const loadAllSuppliers = async () => {
    const response = await suppliersApi.list({ page: 1, per_page: 1000, ...(debouncedSearchQuery ? { search: debouncedSearchQuery } : {}), is_active: !showInactive });
    const allSuppliers = (((response as any)?.data ?? []) as any[]).map((supplier) => normalizeSupplier(supplier));
    return allSuppliers
      .filter((supplier) => !showOutstandingOnly || Number(supplier.outstanding || 0) > 0)
      .sort((a, b) => {
        if (sortBy === 'outstanding') return Number(b.outstanding || 0) - Number(a.outstanding || 0);
        if (sortBy === 'purchases') return Number(b.totalPurchases || 0) - Number(a.totalPurchases || 0);
        return 0;
      });
  };

  const handleExportAll = async () => {
    exportToCSV(getSupplierReportRows(await loadAllSuppliers()), `suppliers-all-${new Date().toISOString().split('T')[0]}`);
  };

  const handlePrintAll = async () => {
    printTable(getSupplierReportRows(await loadAllSuppliers()), ['الاسم', 'الهاتف', 'البريد', 'المشتريات', 'المدفوع', 'صافي المستحق'], 'تقرير كل التجار');
  };

  const handleSubmitSupplier = async (data: SupplierFormData) => {
    try {
      if (editingSupplier) {
        await suppliersApi.update(editingSupplier.id, data);
        showSuccess('تم تحديث التاجر بنجاح');
      } else {
        await suppliersApi.create(data);
        showSuccess('تمت إضافة التاجر بنجاح');
      }
      setIsAddModalOpen(false);
      setEditingSupplier(null);
      refetch();
    } catch (err) {
      showError('حدث خطأ أثناء حفظ التاجر');
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
        showSuccess('تم إيقاف التاجر وإخفاؤه من التعاملات اليومية');
        setDeleteDialogOpen(false);
        setSupplierToDelete(null);
        refetch();
      } catch (err) {
        showError('تعذر إيقاف التاجر');
      }
    }
  };

  const handleRestoreSupplier = async (supplier: any) => {
    try {
      await suppliersApi.update(supplier.id, { ...supplier, is_active: true });
      showSuccess('تمت إعادة تفعيل التاجر بنجاح');
      refetch();
    } catch (err) {
      showError('تعذر إعادة تفعيل التاجر');
    }
  };

  return (
    <div>
      {/* Page Header */}
      <PageHeader
        title={t('suppliers.title')}
        description="إدارة التجار والمشتريات مع رؤى ذكية"
        actions={
          <div style={{ display: 'flex', gap: '10px' }}>
            <Button variant="primary" size={getButtonSize('suppliers', 'headerActions')} onClick={() => setIsAddModalOpen(true)} className="gap-2">
              <Plus className="w-4 h-4" />
              {t('suppliers.addSupplier')}
            </Button>
            <Button variant="secondary" size={getButtonSize('suppliers', 'headerActions')} onClick={() => navigate('/app/purchases')} className="gap-2">
              <ShoppingCart className="w-4 h-4" />
              المشتريات
            </Button>
            <ReportActions onExportCurrent={handleExport} onPrintCurrent={handlePrint} onExportAll={() => { void handleExportAll(); }} onPrintAll={() => { void handlePrintAll(); }} />
          </div>
        }
      />

      {/* Stats Cards */}
      <div className="unified-stats-grid supplier-stats grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="إجمالي التجار"
          value={totalSuppliers}
          icon={Truck}
          subtitle="تاجر نشط"
          variant="featured"
          size="sm"
        />
        <StatCard
          title="إجمالي المشتريات"
          value={`₪${totalPurchases.toLocaleString()}`}
          icon={DollarSign}
          subtitle="قيمة المشتريات"
          variant="default"
          size="sm"
        />
        <StatCard
          title="المدفوع"
          value={`₪${totalPaid.toLocaleString()}`}
          icon={DollarSign}
          subtitle="تم الدفع"
          variant="success"
          size="sm"
        />
        <StatCard
          title="صافي المستحق"
          value={`₪${totalOutstanding.toLocaleString()}`}
          icon={DollarSign}
          subtitle="المتبقي"
          variant="warning"
          size="sm"
        />
      </div>

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
              <Button
                variant={showOutstandingOnly ? 'primary' : 'outline'}
                size={getButtonSize('suppliers', 'headerActions')}
                className="gap-2"
                onClick={() => setShowOutstandingOnly((current) => !current)}
              >
                <Filter className="w-4 h-4" />
                {showOutstandingOnly ? 'عرض كل التجار' : 'التجار المستحقون'}
              </Button>
              <Button
                variant={showInactive ? 'primary' : 'outline'}
                size={getButtonSize('suppliers', 'headerActions')}
                className="gap-2"
                onClick={() => {
                  setShowInactive((current) => !current);
                  setShowOutstandingOnly(false);
                }}
              >
                {showInactive ? <RotateCcw className="w-4 h-4" /> : <RefreshCw className="w-4 h-4" />}
                {showInactive ? 'عرض التجار النشطين' : 'التجار المعطلون'}
              </Button>
              <div className="flex flex-wrap gap-2" role="group" aria-label="ترتيب التجار">
                <SortButton
                  label="الأحدث إضافة"
                  active={sortBy === 'recent'}
                  direction={sortBy === 'recent' ? 'desc' : null}
                  onClick={() => setSortBy('recent')}
                  aria-label="ترتيب التجار حسب الأحدث إضافة"
                  className="min-w-0 flex-1"
                />
                <SortButton
                  label="الأعلى استحقاقًا"
                  active={sortBy === 'outstanding'}
                  direction={sortBy === 'outstanding' ? 'desc' : null}
                  onClick={() => setSortBy('outstanding')}
                  aria-label="ترتيب التجار حسب الأعلى استحقاقًا"
                  className="min-w-0 flex-1"
                />
                <SortButton
                  label="الأعلى مشتريات"
                  active={sortBy === 'purchases'}
                  direction={sortBy === 'purchases' ? 'desc' : null}
                  onClick={() => setSortBy('purchases')}
                  aria-label="ترتيب التجار حسب الأعلى مشتريات"
                  className="min-w-0 flex-1"
                />
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Suppliers Grid */}
      <Card className="rounded-[16px] border-[var(--border-default)] shadow-[0_10px_28px_rgba(15,23,42,0.05)]">
        <CardHeader className="border-b border-[var(--border-subtle)] px-5 py-4">
          <div className="flex items-center justify-between gap-3">
            <div className="flex items-center gap-3">
              <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-[var(--color-primary-10)] text-[var(--primary)]">
                <Truck className="h-4 w-4" />
              </span>
              <div>
                <CardTitle className="text-sm font-extrabold text-[var(--text-primary)]">{showInactive ? 'التجار المعطلون' : 'التجار النشطون'} ({filteredSuppliers.length})</CardTitle>
                  <p className="mt-0.5 text-[11px] font-medium text-[var(--text-muted)]">إدارة بيانات التجار وحساباتهم المالية</p>
              </div>
            </div>
            <Button
              variant="secondary"
              size={getButtonSize('suppliers', 'headerActions')}
              onClick={() => refetch()}
              className="gap-2"
            >
              <RefreshCw className="w-4 h-4" />
              تحديث
            </Button>
          </div>
        </CardHeader>
        <CardContent className="p-5">
          {isLoading ? (
            <div className="flex items-center justify-center h-64">
              <LoadingSpinner size="md" color="cyan" />
            </div>
          ) : filteredSuppliers.length === 0 ? (
            <EmptyState
              icon={<Truck className="h-5 w-5" />}
                title={showInactive ? 'لا يوجد تجار معطلون' : 'لا يوجد تجار نشطون'}
                  description={searchQuery ? 'جرّب تعديل عبارة البحث أو إزالة الفلاتر' : 'أضف أول تاجر لبدء إدارة المشتريات'}
              size="sm"
            />
          ) : (
            <div className="supplier-cards-grid gap-4">
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
                  onRestore={showInactive ? handleRestoreSupplier : undefined}
                />
              ))}
            </div>
          )}
          {filteredSuppliers.length > 0 && (
            <PaginationControls
              page={page}
              pageSize={pageSize}
              total={totalSuppliers}
              onPageChange={setPage}
              isLoading={isLoading}
            />
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
        title="إيقاف التاجر"
        message="سيتم إيقاف التاجر وإخفاؤه من القوائم اليومية مع الاحتفاظ بسجلاته المالية."
        confirmText="إيقاف التاجر"
        cancelText="إلغاء"
        variant="danger"
      />
    </div>
  );
}
