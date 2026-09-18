import { useState, lazy, Suspense } from 'react';
import { useTranslation } from '../../../hooks/useTranslation';
import { PageHeader } from '../../../components/ui/page-header';
import { Button } from '../../../components/ui/button';
import { getButtonSize } from '../../../config/button-sizes';
import { exportToCSV, printTable } from '../../../lib/export-utils';
import { ReportActions } from '../../../components/ui/report-actions';
import { Plus } from 'lucide-react';

// Custom hooks
import { useCustomers } from '../hooks/useCustomers';

// Components
import { CustomerFilters } from '../components/CustomerFilters';
import { CustomerList } from '../components/CustomerList';
import { CustomerModals } from '../components/CustomerModals';
import { ConfirmDialog } from '../../../components/ui/confirm-dialog';

// Lazy load heavy components
const CustomerStats = lazy(() => import('../components/CustomerStats').then(m => ({ default: m.CustomerStats })));

// Types
import { Customer, CustomerFormData } from '../types/customers.types';
import { customersApi } from '../../../services/api/endpoints';

export function CustomersPage() {
  const { t } = useTranslation();
  
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingCustomer, setEditingCustomer] = useState<Customer | null>(null);
  const [selectedCustomer, setSelectedCustomer] = useState<Customer | null>(null);
  const [isViewModalOpen, setIsViewModalOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [customerToDelete, setCustomerToDelete] = useState<string | null>(null);

  // Custom hook
  const {
    filteredCustomers,
    isLoading,
    stats,
    searchQuery,
    setSearchQuery,
    sortConfig,
    setSortConfig,
    createMutation,
    updateMutation,
    deleteMutation,
    handleSort,
    page,
    pageSize,
    total,
    setPage,
  } = useCustomers();

  const handleClearSearch = () => {
    setSearchQuery('');
  };

  const handleAddCustomer = () => {
    setEditingCustomer(null);
    setIsModalOpen(true);
  };

  const handleEditCustomer = (customer: Customer) => {
    setEditingCustomer(customer);
    setIsModalOpen(true);
  };

  const handleViewCustomer = (customer: Customer) => {
    setSelectedCustomer(customer);
    setIsViewModalOpen(true);
  };

  const handleDeleteCustomer = (customerId: string) => {
    setCustomerToDelete(customerId);
    setDeleteDialogOpen(true);
  };

  const handleConfirmDelete = () => {
    if (customerToDelete) {
      deleteMutation.mutate(customerToDelete);
      setDeleteDialogOpen(false);
      setCustomerToDelete(null);
    }
  };

  const handleSubmit = (data: CustomerFormData) => {
    if (editingCustomer) {
      updateMutation.mutate({ id: editingCustomer.id, data });
      setIsModalOpen(false);
      setEditingCustomer(null);
    } else {
      createMutation.mutate(data);
      setIsModalOpen(false);
    }
  };

  const getCustomerReportRows = (customers: Customer[]) => customers.map((customer: Customer) => ({
      'الاسم': customer.name,
      'الهاتف': customer.phone,
      'البريد': customer.email,
      'المشتريات': customer.totalPurchases,
      'الديون': customer.outstanding
    }));

  const handleExport = () => {
    exportToCSV(getCustomerReportRows(filteredCustomers), `customers-${new Date().toISOString().split('T')[0]}`);
  };

  const handlePrint = () => {
    printTable(getCustomerReportRows(filteredCustomers), ['الاسم', 'الهاتف', 'البريد', 'المشتريات', 'الديون'], 'تقرير العملاء');
  };

  const loadAllCustomers = async () => {
    const response = await customersApi.list({ page: 1, per_page: 1000, ...(searchQuery ? { search: searchQuery } : {}) });
    return (((response as any)?.data ?? []) as any[]).map((row) => ({
      ...row,
      totalPurchases: Number(row.totalPurchases ?? row.total_purchases ?? 0),
      outstanding: Number(row.outstanding ?? row.current_balance ?? 0),
    })) as Customer[];
  };

  const handleExportAll = async () => {
    exportToCSV(getCustomerReportRows(await loadAllCustomers()), `customers-all-${new Date().toISOString().split('T')[0]}`);
  };

  const handlePrintAll = async () => {
    printTable(getCustomerReportRows(await loadAllCustomers()), ['الاسم', 'الهاتف', 'البريد', 'المشتريات', 'الديون'], 'تقرير كل العملاء');
  };

  const handleSortName = () => {
    handleSort('name');
  };

  const handleSortPurchases = () => {
    handleSort('totalPurchases');
  };

  const handleClearSort = () => {
    setSortConfig({ key: '', direction: null });
  };

  return (
    <div>
      {/* Page Header */}
      <PageHeader
        eyebrow="Customer Hub"
        title={t('customers.title')}
        description="إدارة العملاء والديون مع تجربة نظيفة ومبتكرة"
        actions={
          <div style={{ display: 'flex', gap: '10px' }}>
            <Button variant="secondary" size={getButtonSize('customers', 'headerActions')} onClick={handleAddCustomer}>
              <Plus style={{ width: '16px', height: '16px', marginRight: '8px' }} />
              إضافة عميل
            </Button>
            <ReportActions onExportCurrent={handleExport} onPrintCurrent={handlePrint} onExportAll={() => { void handleExportAll(); }} onPrintAll={() => { void handlePrintAll(); }} />
          </div>
        }
      />

      {/* Customer Stats */}
      <Suspense fallback={
        <div style={{ 
          display: 'flex', 
          justifyContent: 'center', 
          alignItems: 'center', 
          padding: '40px',
          minHeight: '120px'
        }}>
          <div className="animate-spin rounded-full border-2 border-text-muted/20 border-t-primary" 
               style={{ width: '32px', height: '32px' }} />
        </div>
      }>
        <CustomerStats 
          stats={stats}
        />
      </Suspense>

      {/* Customer Filters */}
      <CustomerFilters
        searchQuery={searchQuery}
        setSearchQuery={setSearchQuery}
        onClearSearch={handleClearSearch}
        sortConfig={sortConfig}
        onSortName={handleSortName}
        onSortPurchases={handleSortPurchases}
        onClearSort={handleClearSort}
      />

      {/* Customer List */}
      <CustomerList
        filteredCustomers={filteredCustomers}
        isLoading={isLoading}
        onViewCustomer={handleViewCustomer}
        onEditCustomer={handleEditCustomer}
        onDeleteCustomer={handleDeleteCustomer}
        pagination={{ page, pageSize, total, onPageChange: setPage }}
      />

      {/* Customer Modals */}
      <CustomerModals
        isModalOpen={isModalOpen}
        setIsModalOpen={setIsModalOpen}
        isViewModalOpen={isViewModalOpen}
        setIsViewModalOpen={setIsViewModalOpen}
        editingCustomer={editingCustomer}
        setEditingCustomer={setEditingCustomer}
        selectedCustomer={selectedCustomer}
        setSelectedCustomer={setSelectedCustomer}
        onSubmit={handleSubmit}
      />

      {/* Delete Confirmation Dialog */}
      <ConfirmDialog
        isOpen={deleteDialogOpen}
        onClose={() => {
          setDeleteDialogOpen(false);
          setCustomerToDelete(null);
        }}
        onConfirm={handleConfirmDelete}
        title="حذف العميل"
        message="سيتم حذف هذا العميل مع جميع البيانات المالية المرتبطة به، مثل الديون والدفعات والمبيعات والمرتجعات والقيود المرتبطة. هذا الإجراء لا يمكن التراجع عنه."
        confirmText="حذف العميل مع البيانات"
        cancelText="إلغاء"
        variant="danger"
        isLoading={deleteMutation.isPending}
      />
    </div>
  );
}