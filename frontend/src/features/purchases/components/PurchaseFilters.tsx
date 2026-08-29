import { Card, CardContent } from '../../../components/ui/card';
import { Input } from '../../../components/ui/input';
import { Select } from '../../../components/ui/select';
import { Button } from '../../../components/ui/button';
import { Search, Filter, X } from 'lucide-react';

interface PurchaseFiltersProps {
  searchQuery: string;
  setSearchQuery: (value: string) => void;
  statusFilter: string;
  setStatusFilter: (value: string) => void;
}

export function PurchaseFilters({
  searchQuery,
  setSearchQuery,
  statusFilter,
  setStatusFilter,
}: PurchaseFiltersProps) {
  return (
    <Card>
      <CardContent className="p-lg">
        <div className="flex flex-col md:flex-row gap-md">
          <div className="flex-1 relative">
            <Search className={`absolute inset-y-0 ${searchQuery ? 'end-10' : 'end-3'} w-4 h-4 text-cyan`} />
            <Input
              placeholder="ابحث برقم الفاتورة أو المورد أو اسم القطعة..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pe-10"
              aria-label="البحث في المشتريات"
            />
            {searchQuery && (
              <button
                type="button"
                onClick={() => setSearchQuery('')}
                className="absolute inset-y-0 end-3 flex items-center text-text-muted hover:text-text-primary"
                aria-label="مسح البحث"
              >
                <X className="w-4 h-4" />
              </button>
            )}
          </div>
          <div className="flex flex-col sm:flex-row gap-sm">
            <Select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              options={[
                { value: '', label: 'كل الحالات' },
                { value: 'draft', label: 'مسودة' },
                { value: 'pending', label: 'قيد الانتظار' },
                { value: 'ordered', label: 'تم الطلب' },
                { value: 'received', label: 'تم الاستلام' },
                { value: 'partially_received', label: 'استلام جزئي' },
                { value: 'cancelled', label: 'ملغي' },
                { value: 'reversed', label: 'تم العكس' },
              ]}
              emptyMessage="لا توجد حالات"
            />
            <Button variant="secondary" className="gap-2">
              <Filter className="w-4 h-4" />
              تصفية
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}