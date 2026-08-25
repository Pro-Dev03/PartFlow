import { useState } from 'react';
import { useTranslation } from '../../../hooks/useTranslation';
import { PageHeader } from '../../../components/ui/page-header';
import { Button } from '../../../components/ui/button';
import { getButtonSize } from '../../../config/button-sizes';
import { exportToCSV, printTable } from '../../../lib/export-utils';
import { Plus, Download, Printer } from 'lucide-react';
import { FinancialTimeline, LedgerEntry } from '../../../components/ui/financial-timeline';
import { customersApi } from '../../../services/api/endpoints';

// Custom hooks
import { useCustomers } from '../hooks/useCustomers';

// Components
import { CustomerStats } from '../components/CustomerStats';
import { CustomerFilters } from '../components/CustomerFilters';
import { CustomerList } from '../components/CustomerList';
import { CustomerModals } from '../components/CustomerModals';

// Types
import { Customer, CustomerFormData } from '../types/customers.types';

export function CustomersPage() {
  const { t } = useTranslation();
  
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingCustomer, setEditingCustomer] = useState<Customer | null>(null);
  const [selectedCustomer, setSelectedCustomer] = useState<Customer | null>(null);
  const [isViewModalOpen, setIsViewModalOpen] = useState(false);
  const [ledgerEntries, setLedgerEntries] = useState<LedgerEntry[]>([]);
  const [loadingLedger, setLoadingLedger] = useState(false);

  // Custom hook
  const {
    customers,
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
    // Load ledger entries for this customer
    loadCustomerLedger(customer.id);
  };

  const loadCustomerLedger = async (customerId: string) => {
    setLoadingLedger(true);
    try {
      const response = await customersApi.ledger(customerId, { page: 1, per_page: 50 });
      if (response && response.data) {
        const formattedEntries = response.data.map((entry: any) => ({
          id: entry.id,
          transaction_type: entry.transaction_type,
          amount: entry.amount,
          balance: entry.balance,
          previous_balance: entry.previous_balance,
          description: entry.description,
          created_at: entry.created_at,
          reference_id: entry.reference_id,
          reference_type: entry.reference_type,
        }));
        setLedgerEntries(formattedEntries);
      } else {
        setLedgerEntries([]);
      }
    } catch (error) {
      console.error('Failed to load ledger entries:', error);
      setLedgerEntries([]);
    } finally {
      setLoadingLedger(false);
    }
  };

  const handleDeleteCustomer = (customerId: string) => {
    if (window.confirm('هل أنت متأكد من حذف هذا العميل؟')) {
      deleteMutation.mutate(customerId);
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

  const handleRecommendationClick = () => {
    // Filter for inactive customers (simple implementation)
    setSearchQuery('');
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
      <CustomerStats 
        stats={stats}
        onRecommendationClick={handleRecommendationClick}
      />

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
        ledgerEntries={ledgerEntries}
        loadingLedger={loadingLedger}
      />
    </div>
  );
}