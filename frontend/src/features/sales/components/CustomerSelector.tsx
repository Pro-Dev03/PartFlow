import { Select } from '../../../components/ui/select';
import { SearchInput } from '../../../components/ui/search-input';
import { User, Users } from 'lucide-react';

interface CustomerOption {
  id: string;
  name: string;
}

interface CustomerSelectorProps {
  selectedCustomer: string;
  setSelectedCustomer: (value: string) => void;
  customers: CustomerOption[];
  customersLoading: boolean;
  customerSearchQuery?: string;
  setCustomerSearchQuery?: (value: string) => void;
  selectedCustomerOption?: CustomerOption;
}

export function CustomerSelector({
  selectedCustomer,
  setSelectedCustomer,
  customers,
  customersLoading,
  customerSearchQuery = '',
  setCustomerSearchQuery,
  selectedCustomerOption,
}: CustomerSelectorProps) {
  const customerOptions = [
    ...(selectedCustomerOption && !customers.some((c) => String(c.id) === String(selectedCustomerOption.id))
      ? [selectedCustomerOption]
      : []),
    ...customers,
  ];
  return (
    <div className="pf-customer-selector">
      <div className="pf-customer-selector-header">
        <div className="flex min-w-0 items-center gap-2">
          <span className="pf-customer-selector-icon"><User className="h-4 w-4" /></span>
          <div className="min-w-0">
            <p className="pf-customer-selector-title">العميل</p>
            <p className="pf-customer-selector-subtitle">اربط البيع بعميل أو اتركه عاماً</p>
          </div>
        </div>
        <span className="pf-customer-selector-count"><Users className="h-3.5 w-3.5" />{customers.length}</span>
      </div>
      
      {/* Search input for customers */}
      {setCustomerSearchQuery && (
        <div className="pf-customer-search">
          <SearchInput
            placeholder="بحث عن عميل بالاسم أو الهاتف..."
            value={customerSearchQuery}
            onChange={(e) => setCustomerSearchQuery(e.target.value)}
            aria-label="بحث عن عميل"
            onClear={() => setCustomerSearchQuery('')}
            size="sm"
          />
        </div>
      )}
      
      <Select
        value={selectedCustomer}
        onChange={(e) => setSelectedCustomer(e.target.value)}
        loading={customersLoading}
        options={customerOptions.map((c) => ({ value: String(c.id), label: c.name }))}
        emptyMessage="لا يوجد عملاء"
        aria-label="اختيار العميل"
        size="sm"
        className="pf-customer-select"
      />
      <p className="pf-customer-selector-hint">
        {customersLoading ? 'جارٍ تحميل العملاء...' : selectedCustomer ? 'تم اختيار العميل' : 'اختياري - للعمليات الآجلة وتسجيل المبيعات'}
      </p>
    </div>
  );
}