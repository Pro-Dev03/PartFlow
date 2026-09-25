import { useState, lazy, Suspense } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from '../../../hooks/useTranslation';
import { PageHeader } from '../../../design-system/components/page-header';
import { Button } from '../../../design-system/components/button';
import { getButtonSize } from '../../../config/button-sizes';
import { exportToCSV, printTable } from '../../../lib/export-utils';
import { ReportActions } from '../../../design-system/components/report-actions';
import { CreditCard, Plus } from 'lucide-react';
import { getStoreToday } from '../../../utils/store-time';
import { toast } from 'sonner';

// Custom hooks
import { useCustomers } from '../hooks/useCustomers';

// Components
import { CustomerFilters } from '../components/CustomerFilters';
import { CustomerList } from '../components/CustomerList';
import { ConfirmDialog } from '../../../design-system/components/confirm-dialog';

// Lazy load heavy components
const CustomerStats = lazy(() => import('../components/CustomerStats').then(m => ({ default: m.CustomerStats })));
const CustomerModals = lazy(() => import('../components/CustomerModals').then(m => ({ default: m.CustomerModals })));

// Types
import { Customer, CustomerFormData } from '../types/customers.types';
import { customersApi } from '../../../services/api/endpoints';

function getCustomerRows(response: any): any[] {
  if (Array.isArray(response)) return response;
  if (Array.isArray(response?.data)) return response.data;
  if (Array.isArray(response?.data?.customers)) return response.data.customers;
  if (Array.isArray(response?.data?.data)) return response.data.data;
  return [];
}

function getCustomerTotal(response: any): number {
  const total = Number(response?.meta?.total ?? response?.total ?? response?.data?.meta?.total);
  return Number.isFinite(total) ? total : getCustomerRows(response).length;
}

export function CustomersPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  
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
    isError,
    refetch,
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
      deleteMutation.mutate(customerToDelete, {
        onSuccess: () => {
          setDeleteDialogOpen(false);
          setCustomerToDelete(null);
        },
      });
    }
  };

  const handleSubmit = (data: CustomerFormData) => {
    if (editingCustomer) {
      updateMutation.mutate({ id: editingCustomer.id, data }, {
        onSuccess: () => {
          setIsModalOpen(false);
          setEditingCustomer(null);
        },
      });
    } else {
      createMutation.mutate(data, {
        onSuccess: () => setIsModalOpen(false),
      });
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
    exportToCSV(getCustomerReportRows(filteredCustomers), `customers-${getStoreToday()}`);
  };

  const handlePrint = () => {
    printTable(getCustomerReportRows(filteredCustomers), ['الاسم', 'الهاتف', 'البريد', 'المشتريات', 'الديون'], 'تقرير العملاء');
  };

  const loadAllCustomers = async () => {
    const perPage = 100;
    const filters = { is_active: true, ...(searchQuery.trim() ? { search: searchQuery.trim() } : {}) };
    const firstPage = await customersApi.list({ page: 1, per_page: perPage, ...filters });
    const totalPages = Math.ceil(getCustomerTotal(firstPage) / perPage);
    const rows = getCustomerRows(firstPage);

    for (let firstPageNumber = 2; firstPageNumber <= totalPages; firstPageNumber += 5) {
      const pageNumbers = Array.from(
        { length: Math.min(5, totalPages - firstPageNumber + 1) },
        (_, index) => firstPageNumber + index,
      );
      const responses = await Promise.all(pageNumbers.map((pageNumber) => customersApi.list({
        page: pageNumber,
        per_page: perPage,
        ...filters,
      })));
      responses.forEach((response) => rows.push(...getCustomerRows(response)));
    }

    return rows.map((row) => ({
      ...row,
      phone: String(row.phone ?? ''),
      totalPurchases: Number(row.totalPurchases ?? row.total_purchases ?? 0),
      outstanding: Number(row.outstanding ?? row.current_balance ?? 0),
    })) as Customer[];
  };

  const handleExportAll = async () => {
    try {
      exportToCSV(getCustomerReportRows(await loadAllCustomers()), `customers-all-${getStoreToday()}`);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'تعذر تصدير العملاء');
    }
  };

  const handlePrintAll = async () => {
    try {
      printTable(getCustomerReportRows(await loadAllCustomers()), ['الاسم', 'الهاتف', 'البريد', 'المشتريات', 'الديون'], 'تقرير كل العملاء');
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'تعذرت طباعة العملاء');
    }
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
        title={t('customers.title')}
        description="إدارة الزبائن والديون مع تجربة نظيفة ومبتكرة"
        actions={
          <div style={{ display: 'flex', gap: '10px' }}>
            <Button variant="secondary" size={getButtonSize('customers', 'headerActions')} onClick={handleAddCustomer}>
              <Plus style={{ width: '16px', height: '16px', marginRight: '8px' }} />
              إضافة زبون
            </Button>
            <Button variant="secondary" size={getButtonSize('customers', 'headerActions')} onClick={() => navigate('/app/debts')}>
              <CreditCard style={{ width: '16px', height: '16px', marginRight: '8px' }} />
              الديون
            </Button>
            <ReportActions onExportCurrent={handleExport} onPrintCurrent={handlePrint} onExportAll={() => { void handleExportAll(); }} onPrintAll={() => { void handlePrintAll(); }} />
          </div>
        }
      />

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
        <CustomerStats stats={stats} />
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
        isError={isError}
        onRetry={() => { void refetch(); }}
        onViewCustomer={handleViewCustomer}
        onEditCustomer={handleEditCustomer}
        onDeleteCustomer={handleDeleteCustomer}
        pagination={{ page, pageSize, total, onPageChange: setPage }}
      />

      {/* Customer Modals */}
      {(isModalOpen || isViewModalOpen) && (
        <Suspense fallback={null}>
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
            isSubmitting={createMutation.isPending || updateMutation.isPending}
          />
        </Suspense>
      )}

      {/* Delete Confirmation Dialog */}
      <ConfirmDialog
        isOpen={deleteDialogOpen}
        onClose={() => {
          setDeleteDialogOpen(false);
          setCustomerToDelete(null);
        }}
        onConfirm={handleConfirmDelete}
        title="أرشفة العميل"
        message="سيتم إخفاء العميل من القوائم النشطة مع الاحتفاظ بجميع المبيعات والدفعات والديون والمرتجعات والسجلات المالية المرتبطة به."
        confirmText="أرشفة العميل"
        cancelText="إلغاء"
        variant="danger"
        isLoading={deleteMutation.isPending}
      />
    </div>
  );
}
