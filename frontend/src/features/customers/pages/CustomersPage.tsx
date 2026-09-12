import { useState, lazy, Suspense } from 'react';
import { useTranslation } from '../../../hooks/useTranslation';
import { PageHeader } from '../../../components/ui/page-header';
import { Button } from '../../../components/ui/button';
import { getButtonSize } from '../../../config/button-sizes';
import { exportToCSV, printTable } from '../../../lib/export-utils';
import { Plus, Download, Printer } from 'lucide-react';

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

  const handleExport = () => {
    const dataToExport = filteredCustomers.map((customer: Customer) => ({
      'الاسم': customer.name,
      'الهاتف': customer.phone,
      'البريد': customer.email,
      'المشتريات': customer.totalPurchases,
      'الديون': customer.outstanding
    }));
    exportToCSV(dataToExport, `customers-${new Date().toISOString().split('T')[0]}`);
  };

  const handlePrint = () => {
    const dataToPrint = filteredCustomers.map((customer: Customer) => ({
      'الاسم': customer.name,
      'الهاتف': customer.phone,
      'البريد': customer.email,
      'المشتريات': customer.totalPurchases,
      'الديون': customer.outstanding
    }));
    printTable(dataToPrint, ['الاسم', 'الهاتف', 'البريد', 'المشتريات', 'الديون'], 'تقرير العملاء');
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
            <Button variant="secondary" size={getButtonSize('customers', 'headerActions')} onClick={handleExport}>
              <Download style={{ width: '16px', height: '16px', marginRight: '8px' }} />
              تصدير
            </Button>
            <Button variant="secondary" size={getButtonSize('customers', 'headerActions')} onClick={handlePrint}>
              <Printer style={{ width: '16px', height: '16px', marginRight: '8px' }} />
              طباعة
            </Button>
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
        message="هل أنت متأكد من حذف هذا العميل؟ هذا الإجراء لا يمكن التراجع عنه."
        confirmText="حذف العميل"
        cancelText="إلغاء"
        variant="danger"
        isLoading={deleteMutation.isPending}
      />
    </div>
  );
}