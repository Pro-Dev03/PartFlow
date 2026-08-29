import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { acquisitionsApi } from '../../../services/api/endpoints';
import { Card, CardContent } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { SearchInput } from '../../../components/ui/search-input';
import { Select } from '../../../components/ui/select';
import { PageHeader } from '../../../components/ui/page-header';
import { Badge } from '../../../components/ui/badge';
import {
  AlertTriangle,
  Clock,
  TrendingDown,
  Layers,
  ArrowRight
} from 'lucide-react';

interface AgingItem {
  id: string;
  product_name: string;
  serial_number?: string;
  condition: string;
  grade: string;
  acquisition_date: string;
  cost: number;
  days_in_stock: number;
  alert_level: string;
  status: string;
}

export function AgingPage() {
  const navigate = useNavigate();
  const [searchQuery, setSearchQuery] = useState('');
  const [alertLevelFilter, setAlertLevelFilter] = useState<string>('');

  const { data: agingData, isLoading } = useQuery({
    queryKey: ['acquisitions-aging'],
    queryFn: () => acquisitionsApi.getAging(),
  });

  const agingItems = Array.isArray(agingData?.data) ? agingData.data : [];
  const longAgingItems = agingItems.filter((item: AgingItem) => item.days_in_stock >= 60);

  // Filter items
  const filteredItems = longAgingItems.filter((item: AgingItem) => {
    const matchesSearch = !searchQuery || 
      item.product_name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      item.serial_number?.toLowerCase().includes(searchQuery.toLowerCase());
    
    const matchesAlert = !alertLevelFilter || item.alert_level === alertLevelFilter;
    
    return matchesSearch && matchesAlert;
  });

  const getAlertBadge = (level: string) => {
    switch (level) {
      case 'CRITICAL':
        return <Badge variant="danger">حرج</Badge>;
      case 'WARNING':
        return <Badge variant="warning">تحذير</Badge>;
      case 'INFO':
        return <Badge variant="info">معلومة</Badge>;
      default:
        return <Badge variant="secondary">{level}</Badge>;
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

  const getAgingColor = (days: number) => {
    if (days >= 90) return 'rgba(239, 68, 68, 0.1)'; // red
    if (days >= 60) return 'rgba(245, 158, 11, 0.1)'; // orange
    if (days >= 30) return 'rgba(59, 130, 246, 0.1)'; // blue
    return 'rgba(34, 197, 94, 0.1)'; // green
  };

  const getAgingBorderColor = (days: number) => {
    if (days >= 90) return 'rgba(239, 68, 68, 0.3)';
    if (days >= 60) return 'rgba(245, 158, 11, 0.3)';
    if (days >= 30) return 'rgba(59, 130, 246, 0.3)';
    return 'rgba(34, 197, 94, 0.3)';
  };

  return (
    <div>
      <PageHeader
        eyebrow="Inventory Aging"
        title="تقادم القطع المستعملة"
        description="القطع المستعملة التي بقيت في المخزون مدة طويلة وتحتاج إلى متابعة"
        actions={
          <Button variant="secondary" onClick={() => navigate('/app/usedparts')}>
            <ArrowRight className="w-4 h-4" />
            رجوع
          </Button>
        }
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
                <p className="text-2xl font-bold">{agingItems.length}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-red/10">
                <AlertTriangle className="w-5 h-5 text-red" />
              </div>
              <div>
                <p className="text-sm text-gray-400">حرج (90+ يوم)</p>
                <p className="text-2xl font-bold">
                  {longAgingItems.filter((i: any) => i.days_in_stock >= 90).length}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-orange/10">
                <Clock className="w-5 h-5 text-orange" />
              </div>
              <div>
                <p className="text-sm text-gray-400">تحذير (60-89 يوم)</p>
                <p className="text-2xl font-bold">
                  {longAgingItems.filter((i: any) => i.days_in_stock >= 60 && i.days_in_stock < 90).length}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-blue/10">
                <TrendingDown className="w-5 h-5 text-blue" />
              </div>
              <div>
                <p className="text-sm text-gray-400">متوسط الأيام</p>
                <p className="text-2xl font-bold">
                  {agingItems.length > 0 
                    ? Math.round(longAgingItems.reduce((sum: number, i: any) => sum + i.days_in_stock, 0) / longAgingItems.length)
                    : 0}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Search and Filters */}
      <Card className="mb-4 border border-[var(--border-default)] bg-[var(--card-bg)] shadow-sm">
        <CardContent className="p-4">
          <div className="pf-search-row flex flex-col md:flex-row gap-3 items-stretch">
            <div className="min-w-0 flex-1">
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
                value={alertLevelFilter}
                onChange={(e) => setAlertLevelFilter(e.target.value)}
                options={[
                  { value: '', label: 'جميع المستويات' },
                  { value: 'CRITICAL', label: 'حرج' },
                  { value: 'WARNING', label: 'تحذير' },
                  { value: 'INFO', label: 'معلومة' },
                  { value: 'OK', label: 'جيد' },
                ]}
                size="sm"
                className="w-48"
              />
              <Button
                variant="secondary"
                size="sm"
                onClick={() => {
                  setSearchQuery('');
                  setAlertLevelFilter('');
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

      {/* Aging Grid */}
      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      ) : filteredItems.length === 0 ? (
        <Card className="border border-[var(--border-default)] bg-[var(--card-bg)] shadow-sm">
          <CardContent className="p-12 text-center">
            <AlertTriangle className="w-12 h-12 mx-auto mb-4 text-[var(--text-muted)]" />
            <p className="text-[var(--text-secondary)]">لا توجد قطع في المخزون</p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filteredItems.map((item: AgingItem) => (
            <Card 
              key={item.id} 
              className="border border-[var(--border-default)] bg-[var(--card-bg)] shadow-sm hover:shadow-md transition-shadow"
              style={{
                background: getAgingColor(item.days_in_stock),
                borderColor: getAgingBorderColor(item.days_in_stock)
              }}
            >
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
                  {getAlertBadge(item.alert_level)}
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
                  <div className="flex justify-between text-sm">
                    <span className="text-[var(--text-secondary)]">تاريخ الشراء:</span>
                    <span className="text-[var(--text-primary)]">{item.acquisition_date}</span>
                  </div>
                </div>

                <div className="pt-3 border-t border-[var(--border-subtle)]">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Clock className="w-4 h-4 text-[var(--text-secondary)]" />
                      <span className="text-sm text-[var(--text-secondary)]">أيام في المخزون:</span>
                    </div>
                    <span className="text-lg font-bold" style={{
                      color: item.days_in_stock >= 90 ? 'var(--color-danger)' : 
                             item.days_in_stock >= 60 ? 'var(--color-warning)' : 
                             'var(--color-success)'
                    }}>
                      {item.days_in_stock}
                    </span>
                  </div>
                </div>

                {item.days_in_stock >= 60 && (
                  <div className="mt-3 p-2 rounded-lg" style={{
                    background: item.days_in_stock >= 90 ? 'rgba(239, 68, 68, 0.1)' : 'rgba(245, 158, 11, 0.1)',
                    border: `1px solid ${item.days_in_stock >= 90 ? 'rgba(239, 68, 68, 0.2)' : 'rgba(245, 158, 11, 0.2)'}`
                  }}>
                    <p className="text-xs text-center" style={{
                      color: item.days_in_stock >= 90 ? 'var(--color-danger)' : 'var(--color-warning)'
                    }}>
                      {item.days_in_stock >= 90 
                        ? 'اقتراح: فكر في تخفيض السعر أو بيعها بالجملة' 
                        : 'اقتراح: فكر في تخفيض السعر بنسبة 5-10%'}
                    </p>
                  </div>
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
