import { ChevronDown, Plus, Search, UserRound, X } from 'lucide-react';

interface PosCustomerOption {
  id: string;
  name: string;
  phone?: string;
}

interface PosCustomerSelectorProps {
  customers: PosCustomerOption[];
  selectedCustomer: string;
  selectedCustomerOption?: Pick<PosCustomerOption, 'id' | 'name'>;
  searchQuery: string;
  isLoading: boolean;
  isOpen: boolean;
  onSearchChange: (value: string) => void;
  onSelect: (customerId: string) => void;
  onToggle: () => void;
  onCreate: () => void;
}

export function PosCustomerSelector({
  customers,
  selectedCustomer,
  selectedCustomerOption,
  searchQuery,
  isLoading,
  isOpen,
  onSearchChange,
  onSelect,
  onToggle,
  onCreate,
}: PosCustomerSelectorProps) {
  const visibleCustomers = [
    ...(selectedCustomerOption && !customers.some((customer) => String(customer.id) === selectedCustomerOption.id)
      ? [selectedCustomerOption]
      : []),
    ...customers,
  ];

  return (
    <div className="pos-customer-bar">
      <div className="customer-select-wrapper">
        <button
          type="button"
          className={`customer-select-trigger ${isOpen ? 'open' : ''}`}
          onClick={onToggle}
          aria-haspopup="listbox"
          aria-expanded={isOpen}
        >
          <span className="customer-select-copy">
            <span className="customer-select-label">العميل</span>
            <span className="customer-select-value">{selectedCustomerOption?.name || 'عميل عام'}</span>
          </span>
          <ChevronDown className="customer-select-chevron" aria-hidden="true" />
        </button>
        {isOpen && (
          <div className="customer-select-menu" role="listbox" aria-label="اختيار العميل">
            <div className="customer-select-search">
              <Search className="h-4 w-4" aria-hidden="true" />
              <input
                type="search"
                value={searchQuery}
                onChange={(event) => onSearchChange(event.target.value)}
                placeholder="ابحث بالاسم أو الهاتف..."
                aria-label="بحث عن عميل بالاسم أو الهاتف"
                autoFocus
              />
              {searchQuery && (
                <button
                  type="button"
                  className="customer-select-search-clear"
                  onClick={() => onSearchChange('')}
                  aria-label="مسح بحث العملاء"
                >
                  <X className="h-3.5 w-3.5" aria-hidden="true" />
                </button>
              )}
            </div>
            <div className="customer-select-results-hint" aria-live="polite">
              {isLoading ? 'جارٍ البحث...' : searchQuery.trim().length >= 2 ? `${customers.length} نتيجة مطابقة` : 'اكتب حرفين على الأقل للبحث'}
            </div>
            <button
              type="button"
              className={`customer-option ${!selectedCustomer ? 'selected' : ''}`}
              onClick={() => onSelect('')}
              role="option"
              aria-selected={!selectedCustomer}
            >
              <span className="customer-option-avatar guest" aria-hidden="true"><UserRound className="h-4 w-4" /></span>
              <span className="customer-option-copy"><strong>عميل عام</strong><small>بيع مباشر بدون حساب عميل</small></span>
            </button>
            {visibleCustomers.map((customer) => (
              <button
                key={customer.id}
                type="button"
                className={`customer-option ${String(customer.id) === selectedCustomer ? 'selected' : ''}`}
                onClick={() => onSelect(String(customer.id))}
                role="option"
                aria-selected={String(customer.id) === selectedCustomer}
              >
                <span className="customer-option-avatar" aria-hidden="true"><UserRound className="h-4 w-4" /></span>
                <span className="customer-option-copy"><strong>{customer.name}</strong><small>حساب عميل</small></span>
              </button>
            ))}
            {!isLoading && customers.length === 0 && searchQuery.trim().length >= 2 && (
              <p className="customer-select-empty">لا يوجد عميل مطابق للبحث</p>
            )}
          </div>
        )}
      </div>
      <button type="button" className="quick-customer-btn" onClick={onCreate} aria-label="إضافة عميل">
        <Plus className="w-4 h-4" aria-hidden="true" />
      </button>
    </div>
  );
}
