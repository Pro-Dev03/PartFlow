import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { acquisitionsApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { SearchInput } from '../../../components/ui/search-input';
import { Select } from '../../../components/ui/select';
import { PageHeader } from '../../../components/ui/page-header';
import { Badge } from '../../../components/ui/badge';
import { Modal } from '../../../components/ui/modal';
import {
  CheckCircle,
  XCircle,
  AlertTriangle,
  Package,
  Search,
  Filter,
  Layers
} from 'lucide-react';

interface InspectionItem {
  id: string;
  product_name: string;
  serial_number?: string;
  condition: string;
  grade: string;
  inspection_status: string;
  acquisition_date: string;
  customer_name?: string;
  cost: number;
}

export function InspectionsPage() {
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [selectedItem, setSelectedItem] = useState<InspectionItem | null>(null);
  const [isInspectionModalOpen, setIsInspectionModalOpen] = useState(false);
  const [inspectionChecks, setInspectionChecks] = useState<string[]>([]);
  const [inspectionNotes, setInspectionNotes] = useState('');

  const { data: acquisitionsData, isLoading } = useQuery({
    queryKey: ['acquisitions'],
    queryFn: () => acquisitionsApi.list({ type: 'CUSTOMER' }),
  });

  const acquisitions = Array.isArray(acquisitionsData?.data) ? acquisitionsData.data : [];
  
  // Extract items from acquisitions
  const inspectionItems: InspectionItem[] = acquisitions.flatMap((acq: any) => 
    acq.items?.map((item: any) => ({
      ...item,
      customer_name: acq.customer_name,
      acquisition_date: acq.acquisition_date,
    })) || []
  );

  // Filter items
  const filteredItems = inspectionItems.filter((item: InspectionItem) => {
    const matchesSearch = !searchQuery || 
      item.product_name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      item.serial_number?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      item.customer_name?.toLowerCase().includes(searchQuery.toLowerCase());
    
    const matchesStatus = !statusFilter || item.inspection_status === statusFilter;
    
    return matchesSearch && matchesStatus;
  });

  const handleStartInspection = (item: InspectionItem) => {
    setSelectedItem(item);
    setInspectionChecks([]);
    setInspectionNotes('');
    setIsInspectionModalOpen(true);
  };

  const handleCompleteInspection = async (passed: boolean) => {
    if (!selectedItem) return;

    try {
      await acquisitionsApi.updateStatus(selectedItem.id, passed ? 'PASSED' : 'FAILED');
      queryClient.invalidateQueries({ queryKey: ['acquisitions'] });
      queryClient.invalidateQueries({ queryKey: ['inspections'] });
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      alert(passed ? 'اجتازت القطعة الفحص بنجاح!' : 'فشلت القطعة في الفحص');
      setIsInspectionModalOpen(false);
      setSelectedItem(null);
    } catch (error) {
      console.error('Inspection failed:', error);
      alert('فشل تحديث حالة الفحص');
    }
  };

  const toggleInspectionCheck = (checkName: string) => {
    setInspectionChecks(prev =>
      prev.includes(checkName)
        ? prev.filter(c => c !== checkName)
        : [...prev, checkName]
    );
  };

  const allRequiredChecksPassed = () => {
    return inspectionChecks.length >= 3;
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'PASSED':
        return <Badge variant="success">اجتاز الفحص</Badge>;
      case 'FAILED':
        return <Badge variant="danger">فشل الفحص</Badge>;
      case 'PENDING':
        return <Badge variant="warning">قيد الفحص</Badge>;
      default:
        return <Badge variant="secondary">{status}</Badge>;
    }
  };

  const getConditionLabel = (condition: string) => {
    const labels: Record<string, string> = {
      'new': 'جديد',
      'used': 'مستعمل',
      'refurbished': 'مجدّد',
    };
    return labels[condition] || condition;
  };

  const getGradeLabel = (grade: string) => {
    const labels: Record<string, string> = {
      'excellent': 'ممتاز',
      'very_good': 'جيد جداً',
      'good': 'جيد',
      'fair': 'متوسط',
      'poor': 'ضعيف',
    };
    return labels[grade] || grade;
  };

  return (
    <div>
      <PageHeader
        eyebrow="Inspections Management"
        title="إدارة الفحص"
        description="فحص القطع المستعملة قبل إضافتها للمخزون"
      />

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-3 mb-4">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-cyan/10">
                <Layers className="w-5 h-5 text-cyan" />
              </div>
              <div>
                <p className="text-sm text-gray-400">إجمالي القطع</p>
                <p className="text-2xl font-bold">{inspectionItems.length}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-yellow/10">
                <AlertTriangle className="w-5 h-5 text-yellow" />
              </div>
              <div>
                <p className="text-sm text-gray-400">قيد الفحص</p>
                <p className="text-2xl font-bold">
                  {inspectionItems.filter((i: any) => i.inspection_status === 'PENDING').length}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-green/10">
                <CheckCircle className="w-5 h-5 text-green" />
              </div>
              <div>
                <p className="text-sm text-gray-400">اجتازت الفحص</p>
                <p className="text-2xl font-bold">
                  {inspectionItems.filter((i: any) => i.inspection_status === 'PASSED').length}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-red/10">
                <XCircle className="w-5 h-5 text-red" />
              </div>
              <div>
                <p className="text-sm text-gray-400">فشلت الفحص</p>
                <p className="text-2xl font-bold">
                  {inspectionItems.filter((i: any) => i.inspection_status === 'FAILED').length}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Search and Filters */}
      <Card className="mb-4 border border-[var(--border-default)] bg-[var(--card-bg)] shadow-sm">
        <CardContent className="p-4">
          <div className="flex flex-col md:flex-row gap-3 items-stretch">
            <div className="flex-1">
              <SearchInput
                placeholder="بحث عن قطعة..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onClear={() => setSearchQuery('')}
                size="sm"
                className="w-full"
              />
            </div>
            <div className="flex gap-2">
              <Select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                options={[
                  { value: '', label: 'جميع الحالات' },
                  { value: 'PENDING', label: 'قيد الفحص' },
                  { value: 'PASSED', label: 'اجتازت الفحص' },
                  { value: 'FAILED', label: 'فشلت الفحص' },
                ]}
                size="sm"
                className="w-48"
              />
              <Button
                variant="secondary"
                size="sm"
                onClick={() => {
                  setSearchQuery('');
                  setStatusFilter('');
                }}
                className="h-10 px-4"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="mr-1.5">
                  <path d="M3 6h18"></path>
                  <path d="M7 12h10"></path>
                  <path d="M10 18h4"></path>
                </svg>
                <span>مسح</span>
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Inspections Grid */}
      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      ) : filteredItems.length === 0 ? (
        <Card className="border border-[var(--border-default)] bg-[var(--card-bg)] shadow-sm">
          <CardContent className="p-12 text-center">
            <AlertTriangle className="w-12 h-12 mx-auto mb-4 text-[var(--text-muted)]" />
            <p className="text-[var(--text-secondary)]">لا توجد قطع للفحص</p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filteredItems.map((item: InspectionItem) => (
            <Card key={item.id} className="border border-[var(--border-default)] bg-[var(--card-bg)] shadow-sm hover:shadow-md transition-shadow">
              <CardContent className="p-4">
                <div className="flex justify-between items-start mb-3">
                  <div>
                    <h3 className="font-semibold text-[var(--text-primary)] mb-1">
                      {item.product_name}
                    </h3>
                    {item.serial_number && (
                      <p className="text-xs text-[var(--text-muted)]">SN: {item.serial_number}</p>
                    )}
                  </div>
                  {getStatusBadge(item.inspection_status)}
                </div>

                <div className="space-y-2 mb-3">
                  <div className="flex justify-between text-sm">
                    <span className="text-[var(--text-secondary)]">الحالة:</span>
                    <span className="text-[var(--text-primary)]">{getConditionLabel(item.condition)}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-[var(--text-secondary)]">التقييم:</span>
                    <span className="text-[var(--text-primary)]">{getGradeLabel(item.grade)}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-[var(--text-secondary)]">التكلفة:</span>
                    <span className="text-[var(--text-primary)]">₪{(item.cost / 100).toFixed(2)}</span>
                  </div>
                  {item.customer_name && (
                    <div className="flex justify-between text-sm">
                      <span className="text-[var(--text-secondary)]">البائع:</span>
                      <span className="text-[var(--text-primary)]">{item.customer_name}</span>
                    </div>
                  )}
                </div>

                {item.inspection_status === 'PENDING' && (
                  <Button
                    variant="primary"
                    size="sm"
                    className="w-full"
                    onClick={() => handleStartInspection(item)}
                  >
                    <AlertTriangle className="w-4 h-4 mr-1" />
                    بدء الفحص
                  </Button>
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Inspection Modal */}
      <Modal
        isOpen={isInspectionModalOpen}
        onClose={() => setIsInspectionModalOpen(false)}
        title="فحص القطعة"
        variant="modern"
        size="lg"
        style={{
          borderRadius: '24px',
          overflow: 'hidden',
          background: 'var(--bg-surface)',
          border: '1px solid var(--border-primary)',
          boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25), 0 0 0 1px rgba(255, 255, 255, 0.05) inset, 0 0 40px rgba(99, 102, 241, 0.1)'
        }}
      >
        <div className="space-y-6">
          {/* Item Info */}
          {selectedItem && (
            <div style={{
              padding: '20px',
              background: 'linear-gradient(135deg, rgba(99, 102, 241, 0.05) 0%, rgba(34, 211, 238, 0.05) 100%)',
              borderRadius: '16px',
              border: '1px solid rgba(99, 102, 241, 0.2)'
            }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                <div style={{ width: '36px', height: '36px', borderRadius: '10px', background: 'rgba(99, 102, 241, 0.1)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                  <Package className="w-5 h-5" style={{ color: 'var(--color-primary)' }} />
                </div>
                <div>
                  <p style={{ fontSize: '16px', fontWeight: '600', color: 'var(--text-primary)', marginBottom: '2px' }}>
                    {selectedItem.product_name}
                  </p>
                  {selectedItem.serial_number && (
                    <p style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
                      SN: {selectedItem.serial_number}
                    </p>
                  )}
                </div>
              </div>
            </div>
          )}

          {/* Inspection Checks */}
          <div style={{
            padding: '20px',
            background: 'var(--bg-surface-elevated)',
            borderRadius: '16px',
            border: '1px solid var(--border-subtle)'
          }}>
            <label style={{
              fontSize: '14px',
              fontWeight: '600',
              color: 'var(--text-primary)',
              marginBottom: '16px',
              display: 'block',
              letterSpacing: '0.2px'
            }}>
              فحص القطعة
            </label>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
              {[
                'التشغيل يعمل',
                'الحالة الخارجية جيدة',
                'الحرارة طبيعية',
                'لا توجد أضرار واضحة',
                'المكونات سليمة',
                'الأداء جيد'
              ].map((check) => (
                <button
                  key={check}
                  onClick={() => toggleInspectionCheck(check)}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '12px',
                    padding: '12px 16px',
                    borderRadius: '12px',
                    background: inspectionChecks.includes(check)
                      ? 'rgba(34, 197, 94, 0.1)'
                      : 'var(--bg-surface)',
                    border: inspectionChecks.includes(check)
                      ? '1px solid rgba(34, 197, 94, 0.3)'
                      : '1px solid var(--border-subtle)',
                    cursor: 'pointer',
                    transition: 'all 0.2s ease',
                    textAlign: 'right'
                  }}
                >
                  <div style={{
                    width: '20px',
                    height: '20px',
                    borderRadius: '6px',
                    background: inspectionChecks.includes(check)
                      ? 'rgba(34, 197, 94, 0.2)'
                      : 'var(--bg-surface-elevated)',
                    border: inspectionChecks.includes(check)
                      ? '2px solid rgba(34, 197, 94, 0.5)'
                      : '2px solid var(--border-default)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center'
                  }}>
                    {inspectionChecks.includes(check) && (
                      <CheckCircle className="w-3 h-3" style={{ color: 'rgba(34, 197, 94, 0.8)' }} />
                    )}
                  </div>
                  <span style={{
                    fontSize: '13px',
                    fontWeight: '500',
                    color: inspectionChecks.includes(check)
                      ? 'rgba(34, 197, 94, 0.9)'
                      : 'var(--text-primary)'
                  }}>
                    {check}
                  </span>
                </button>
              ))}
            </div>
          </div>

          {/* Notes */}
          <div style={{
            padding: '20px',
            background: 'var(--bg-surface-elevated)',
            borderRadius: '16px',
            border: '1px solid var(--border-subtle)'
          }}>
            <label style={{
              fontSize: '14px',
              fontWeight: '600',
              color: 'var(--text-primary)',
              marginBottom: '12px',
              display: 'block',
              letterSpacing: '0.2px'
            }}>
              ملاحظات الفحص
            </label>
            <textarea
              value={inspectionNotes}
              onChange={(e) => setInspectionNotes(e.target.value)}
              placeholder="أدخل ملاحظات الفحص..."
              rows={3}
              style={{
                width: '100%',
                padding: '12px 16px',
                borderRadius: '12px',
                border: '1px solid var(--border-default)',
                background: 'var(--bg-surface)',
                color: 'var(--text-primary)',
                fontSize: '14px',
                resize: 'vertical',
                outline: 'none',
                transition: 'all 0.2s ease'
              }}
            />
          </div>

          {/* Actions */}
          <div style={{
            display: 'flex',
            gap: '12px',
            justifyContent: 'flex-end',
            paddingTop: '32px',
            borderTop: '1px solid var(--border-subtle)',
            marginTop: '24px'
          }}>
            <Button
              variant="secondary"
              onClick={() => setIsInspectionModalOpen(false)}
            >
              إلغاء
            </Button>
            <Button
              variant="danger"
              onClick={() => handleCompleteInspection(false)}
            >
              <XCircle className="w-4 h-4 mr-1" />
              فشل الفحص
            </Button>
            <Button
              variant="success"
              onClick={() => handleCompleteInspection(true)}
              disabled={!allRequiredChecksPassed()}
            >
              <CheckCircle className="w-4 h-4 mr-1" />
              اجتاز الفحص
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}