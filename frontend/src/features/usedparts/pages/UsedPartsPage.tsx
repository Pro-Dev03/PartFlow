import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { inventoryApi, partTypesApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { SearchInput } from '../../../components/ui/search-input';
import { Select } from '../../../components/ui/select';
import { PageHeader } from '../../../components/ui/page-header';
import { Badge } from '../../../components/ui/badge';
import { getButtonSize } from '../../../config/button-sizes';
import { 
  Plus, 
  Edit, 
  Trash2, 
  Search,
  Filter,
  Layers,
  Package,
  Cpu
} from 'lucide-react';

export function UsedPartsPage() {
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedPartType, setSelectedPartType] = useState<string>('');

  const handleClearSearch = () => {
    setSearchQuery('');
  };

  const { data: inventoryData, isLoading } = useQuery({
    queryKey: ['inventory'],
    queryFn: () => inventoryApi.list({ page: 1, per_page: 100 }),
  });

  const { data: partTypesData, error: partTypesError, isLoading: partTypesLoading } = useQuery({
    queryKey: ['part-types'],
    queryFn: () => partTypesApi.list(),
    retry: 1,
  });

  const inventoryItems = (inventoryData?.data as any[]) || [];
  const partTypes = (partTypesData?.data as any[]) || [];

  // Handle part types error gracefully
  if (partTypesError) {
    console.warn('Failed to load part types:', partTypesError);
  }

  // Filter used parts only
  const usedParts = inventoryItems.filter((item: any) => 
    item.condition === 'USED' && item.status === 'AVAILABLE'
  );

  // Filter by search and part type
  const filteredParts = usedParts.filter((item: any) => {
    const matchesSearch = !searchQuery || 
      (item.product_name && item.product_name.toLowerCase().includes(searchQuery.toLowerCase())) ||
      (item.notes && item.notes.toLowerCase().includes(searchQuery.toLowerCase()));
    
    const matchesType = !selectedPartType || item.part_type_id === selectedPartType;
    
    return matchesSearch && matchesType;
  });

  const getIconComponent = (iconName: string) => {
    return <Layers className="w-4 h-4" />;
  };

  const getPartType = (partTypeId: string) => {
    if (!partTypes || partTypes.length === 0) return null;
    return partTypes.find((pt: any) => pt.id === partTypeId);
  };

  return (
    <div>
      <PageHeader
        eyebrow="Used Parts Inventory"
        title="القطع المستعملة"
        description="إدارة القطع المستعملة ومواصفاتها"
        actions={
          <Button
            variant="primary"
            size={getButtonSize('usedparts', 'headerActions')}
            onClick={() => {/* Will open trade-in modal */}}
          >
            <Plus className="w-3.5 h-3.5 me-1.5" />
            شراء قطعة مستعملة
          </Button>
        }
      />

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-cyan/10">
                <Layers className="w-5 h-5 text-cyan" />
              </div>
              <div>
                <p className="text-sm text-gray-400">إجمالي القطع</p>
                <p className="text-2xl font-bold">{usedParts.length}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-green/10">
                <Package className="w-5 h-5 text-green" />
              </div>
              <div>
                <p className="text-sm text-gray-400">قيمة المخزون</p>
                <p className="text-2xl font-bold">
                  ₪{usedParts.reduce((sum: number, item: any) => sum + (item.selling_price / 100), 0).toFixed(0)}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-purple/10">
                <Cpu className="w-5 h-5 text-purple" />
              </div>
              <div>
                <p className="text-sm text-gray-400">أنواع القطع</p>
                <p className="text-2xl font-bold">{partTypes.length}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Search and Filters */}
      <Card className="mb-6" style={{
        background: 'linear-gradient(135deg, rgba(17, 24, 39, 0.6) 0%, rgba(17, 24, 39, 0.4) 100%)',
        border: '1px solid rgba(99, 102, 241, 0.15)',
        backdropFilter: 'blur(10px)'
      }}>
        <CardContent className="p-4">
          <div className="flex flex-col md:flex-row gap-3">
            <div className="flex-1">
              <SearchInput
                placeholder="بحث عن قطعة..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onClear={handleClearSearch}
                size="sm"
                className="w-full"
              />
            </div>
            <div className="flex gap-2">
              <Select
                value={selectedPartType}
                onChange={(e) => setSelectedPartType(e.target.value)}
                loading={partTypesLoading}
                options={[
                  { value: '', label: 'جميع الأنواع' },
                  ...partTypes.map((pt: any) => ({ value: pt.id, label: pt.name_ar })),
                ]}
                emptyMessage="لا يوجد أنواع"
              />
              <Button
                variant="secondary"
                size="sm"
                onClick={() => {
                  setSearchQuery('');
                  setSelectedPartType('');
                }}
                style={{
                  height: '48px',
                  fontSize: '14px',
                  fontWeight: '600',
                  letterSpacing: '0.3px',
                  borderRadius: '8px',
                  background: 'linear-gradient(135deg, rgba(99, 102, 241, 0.2) 0%, rgba(34, 211, 238, 0.2) 100%)',
                  border: '1px solid rgba(99, 102, 241, 0.3)',
                  color: 'var(--text-primary)',
                  transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.background = 'linear-gradient(135deg, rgba(99, 102, 241, 0.3) 0%, rgba(34, 211, 238, 0.3) 100%)';
                  e.currentTarget.style.borderColor = 'rgba(99, 102, 241, 0.5)';
                  e.currentTarget.style.transform = 'translateY(-1px)';
                  e.currentTarget.style.boxShadow = '0 4px 12px rgba(99, 102, 241, 0.2)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.background = 'linear-gradient(135deg, rgba(99, 102, 241, 0.2) 0%, rgba(34, 211, 238, 0.2) 100%)';
                  e.currentTarget.style.borderColor = 'rgba(99, 102, 241, 0.3)';
                  e.currentTarget.style.transform = 'translateY(0)';
                  e.currentTarget.style.boxShadow = 'none';
                }}
              >
                مسح
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Parts Grid */}
      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      ) : filteredParts.length === 0 ? (
        <Card>
          <CardContent className="p-12 text-center">
            <Package className="w-12 h-12 mx-auto mb-4 text-gray-400" />
            <p className="text-gray-400 mb-4">لا توجد قطع مستعملة متاحة</p>
            <Button
              variant="primary"
              onClick={() => {/* Will open trade-in modal */}}
            >
              شراء قطعة مستعملة
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {filteredParts.map((item: any) => {
            const partType = getPartType(item.part_type_id);
            return (
              <Card 
                key={item.id}
                className="hover:border-cyan-500 transition-colors cursor-pointer"
                style={{
                  background: partType ? `linear-gradient(135deg, ${partType.color}15 0%, transparent 100%)` : undefined
                }}
              >
                <CardContent className="p-4">
                  <div className="flex items-start justify-between mb-3">
                    {partType && (
                      <div
                        className="p-2 rounded-lg"
                        style={{ background: `${partType.color}30`, color: partType.color }}
                      >
                        {getIconComponent(partType.icon)}
                      </div>
                    )}
                    <Badge variant="success">متاح</Badge>
                  </div>
                  
                  <h3 className="font-semibold mb-1">
                    {item.product_name || item.product?.name || 'قطعة بدون اسم'}
                  </h3>
                  
                  {partType && (
                    <p className="text-sm text-gray-400 mb-2">{partType.name_ar}</p>
                  )}
                  
                  <div className="space-y-2 mb-3">
                    <div className="flex justify-between text-sm">
                      <span className="text-gray-400">اشتريت بـ:</span>
                      <span>₪{(item.purchase_cost / 100).toFixed(2)}</span>
                    </div>
                    <div className="flex justify-between text-sm font-semibold">
                      <span className="text-gray-400">سعر البيع:</span>
                      <span className="text-cyan">₪{(item.selling_price / 100).toFixed(2)}</span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span className="text-gray-400">الربح:</span>
                      <span className={item.selling_price > item.purchase_cost ? 'text-green' : 'text-red'}>
                        ₪{((item.selling_price - item.purchase_cost) / 100).toFixed(2)}
                      </span>
                    </div>
                  </div>
                  
                  {item.notes && (
                    <p className="text-xs text-gray-500 mt-2 line-clamp-2">{item.notes}</p>
                  )}
                  
                  <div className="flex gap-2 mt-4">
                    <Button variant="secondary" size="sm" className="flex-1">
                      <Edit className="w-3 h-3 mr-1" />
                      تعديل
                    </Button>
                    <Button variant="primary" size="sm" className="flex-1">
                      بيع
                    </Button>
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
}