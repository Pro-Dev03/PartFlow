import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { inventoryApi, productsApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { PageHeader } from '../../../components/ui/page-header';
import { Badge } from '../../../components/ui/badge';
import { EmptyState } from '../../../components/ui/empty-state';
import { ErrorState } from '../../../components/ui/error-state';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
import { 
  Package, 
  Search, 
  Plus, 
  Filter,
  ArrowUpDown,
  Eye,
  Edit,
  Trash2,
  PackageOpen,
  Inbox,
  AlertCircle,
  Sparkles,
  TrendingUp,
  TrendingDown,
  AlertTriangle,
  Layers,
  Zap
} from 'lucide-react';

export function InventoryPage() {
  const { t } = useTranslation();
  const [searchQuery, setSearchQuery] = useState('');
  const [viewMode, setViewMode] = useState<'products' | 'items'>('products');

  const { data: productsData, isLoading: productsLoading } = useQuery({
    queryKey: ['products'],
    queryFn: () => productsApi.list(),
  });

  const { data: inventoryData, isLoading: inventoryLoading } = useQuery({
    queryKey: ['inventory'],
    queryFn: () => inventoryApi.list(),
  });

  const products = (productsData?.data as any[]) || [];
  const inventoryItems = (inventoryData?.data as any[]) || [];

  const filteredProducts = products.filter((product: any) =>
    product.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    product.sku.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const getConditionBadge = (condition: string) => {
    const variants: Record<string, { label: string; variant: 'default' | 'success' | 'warning' | 'danger' | 'info' | 'secondary' | 'outline' }> = {
      new: { label: t('products.new'), variant: 'success' },
      used: { label: t('products.used'), variant: 'secondary' },
      refurbished: { label: t('products.refurbished'), variant: 'info' },
      parts_only: { label: t('products.partsOnly'), variant: 'danger' },
    };
    return variants[condition] || { label: condition, variant: 'default' };
  };

  return (
    <div className="space-y-md">
      {/* Page Header - Futuristic + Data Dense */}
      <PageHeader
        eyebrow="Inventory Intelligence"
        title={t('inventory.title')}
        description="إدارة المخزون والقطع مع تحليلات فورية"
        actions={
          <div className="flex items-center gap-sm">
            <Button variant="primary" className="gap-2">
              <Plus className="w-4 h-4" />
              {t('inventory.addItem')}
            </Button>
            <Button variant="secondary" className="gap-2">
              <Zap className="w-4 h-4" />
              تصدير
            </Button>
          </div>
        }
      />

      {/* AI Inventory Insight - Futuristic + Data Dense */}
      <Card variant="ai">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Sparkles className="w-5 h-5 text-cyan" />
            AI Inventory Insight
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-md">
            <div className="w-8 h-8 rounded-sm bg-cyan/10 flex items-center justify-center flex-shrink-0">
              <TrendingUp className="w-4 h-4 text-cyan" />
            </div>
            <div>
              <p className="text-small font-semibold text-text">فرصة شراء معالجات Intel</p>
              <p className="text-tiny text-text-muted mt-1">
                الأسعار الحالية أقل من المتوسط بنسبة 15%. هناك طلب متزايد من 3 عملاء رئيسيين.
              </p>
              <Button variant="ghost" size="sm" className="mt-2 text-cyan">
                عرض التوصية ←
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Stats Cards - Futuristic + Data Dense */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-md">
        <StatCard 
          title={t('inventory.totalItems')} 
          value={inventoryItems.length} 
          icon={Package}
          subtitle="إجمالي العناصر"
          variant="featured"
          trend="+8.3%"
          trendUp={true}
        />
        <StatCard 
          title={t('inventory.totalValue')} 
          value="₪185,400" 
          icon={Package}
          subtitle="قيمة المخزون"
          variant="default"
          trend="+12.1%"
          trendUp={true}
        />
        <StatCard 
          title={t('inventory.lowStock')} 
          value="12" 
          icon={AlertTriangle}
          subtitle="يحتاج طلب"
          variant="warning"
          trend="-2"
          trendUp={false}
        />
        <StatCard 
          title="مستعمل" 
          value="84" 
          icon={Layers}
          subtitle="حالة البضاعة"
          variant="info"
          trend="+5"
          trendUp={true}
        />
      </div>

      {/* Search and Filters - Futuristic + Data Dense */}
      <Card>
        <CardContent className="p-lg">
          <div className="flex flex-col md:flex-row gap-md">
            <div className="flex-1 relative">
              <Search className="absolute inset-y-0 end-3 w-4 h-4 text-text-muted" />
              <Input
                placeholder={t('common.search')}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pe-10"
                leftIcon={<Search className="w-4 h-4" />}
              />
            </div>
            <div className="flex gap-sm">
              <Button variant="secondary" className="gap-2">
                <Filter className="w-4 h-4" />
                {t('common.filter')}
              </Button>
              <Button variant="secondary" className="gap-2">
                <ArrowUpDown className="w-4 h-4" />
                ترتيب
              </Button>
              <Button variant="secondary" className="gap-2">
                <Zap className="w-4 h-4" />
                تحديث
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* View Toggle - Futuristic + Data Dense */}
      <div className="flex gap-sm">
        <Button
          variant={viewMode === 'products' ? 'primary' : 'secondary'}
          onClick={() => setViewMode('products')}
          className="gap-2"
        >
          <Package className="w-4 h-4" />
          {t('products.title')}
        </Button>
        <Button
          variant={viewMode === 'items' ? 'primary' : 'secondary'}
          onClick={() => setViewMode('items')}
          className="gap-2"
        >
          <PackageOpen className="w-4 h-4" />
          {t('inventory.items')}
        </Button>
      </div>

      {/* Products Table - Futuristic + Data Dense */}
      {viewMode === 'products' && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Package className="w-5 h-5 text-cyan" />
              المنتجات
            </CardTitle>
          </CardHeader>
          <CardContent>
            {productsLoading ? (
              <div className="flex items-center justify-center h-64" role="status" aria-label="Loading">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-cyan" />
              </div>
            ) : filteredProducts.length === 0 ? (
              <EmptyState
                icon={<Inbox className="w-8 h-8" />}
                title="لا توجد منتجات"
                description="لم يتم العثور على منتجات تطابق بحثك"
                action={
                  <Button variant="primary" onClick={() => setSearchQuery('')}>
                    مسح البحث
                  </Button>
                }
              />
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>الاسم</TableHead>
                    <TableHead>SKU</TableHead>
                    <TableHead>الفئة</TableHead>
                    <TableHead>الحالة</TableHead>
                    <TableHead>المخزون</TableHead>
                    <TableHead>السعر</TableHead>
                    <TableHead>الحالة</TableHead>
                    <TableHead className="text-end">الإجراءات</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredProducts.map((product: any) => (
                    <TableRow key={product.id}>
                      <TableCell className="font-medium">{product.name}</TableCell>
                      <TableCell>{product.sku}</TableCell>
                      <TableCell>{product.category}</TableCell>
                      <TableCell>
                        <Badge variant={getConditionBadge(product.condition).variant}>
                          {getConditionBadge(product.condition).label}
                        </Badge>
                      </TableCell>
                      <TableCell>{product.stock}</TableCell>
                      <TableCell>₪{product.sellingPrice?.toLocaleString()}</TableCell>
                      <TableCell>
                        <Badge variant={product.status === 'active' ? 'success' : 'secondary'}>
                          {product.status === 'active' ? 'نشط' : 'غير نشط'}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-end">
                        <div className="flex gap-sm justify-end">
                          <Button variant="ghost" size="sm" aria-label="View product">
                            <Eye className="w-4 h-4" />
                          </Button>
                          <Button variant="ghost" size="sm" aria-label="Edit product">
                            <Edit className="w-4 h-4" />
                          </Button>
                          <Button variant="ghost" size="sm" aria-label="Delete product">
                            <Trash2 className="w-4 h-4" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      )}

      {/* Items Table - Futuristic + Data Dense */}
      {viewMode === 'items' && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <PackageOpen className="w-5 h-5 text-cyan" />
              القطع
            </CardTitle>
          </CardHeader>
          <CardContent>
            {inventoryLoading ? (
              <div className="flex items-center justify-center h-64" role="status" aria-label="Loading">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-cyan" />
              </div>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>الباركود</TableHead>
                    <TableHead>المنتج</TableHead>
                    <TableHead>الرقم التسلسلي</TableHead>
                    <TableHead>الحالة</TableHead>
                    <TableHead>الموقع</TableHead>
                    <TableHead>سعر الشراء</TableHead>
                    <TableHead>سعر البيع</TableHead>
                    <TableHead>الحالة</TableHead>
                    <TableHead className="text-end">الإجراءات</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {(inventoryItems as any[]).map((item: any) => (
                    <TableRow key={item.id}>
                      <TableCell className="font-medium">{item.barcode}</TableCell>
                      <TableCell>{item.product?.name}</TableCell>
                      <TableCell>{item.serial || '-'}</TableCell>
                      <TableCell>
                        <Badge variant={getConditionBadge(item.condition).variant}>
                          {getConditionBadge(item.condition).label}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        {item.location?.warehouse}/{item.location?.shelf}/{item.location?.box}
                      </TableCell>
                      <TableCell>₪{item.purchaseCost?.toLocaleString()}</TableCell>
                      <TableCell>₪{item.sellingPrice?.toLocaleString()}</TableCell>
                      <TableCell>
                        <Badge variant={item.status === 'available' ? 'success' : 'secondary'}>
                          {item.status === 'available' ? 'متاح' : item.status}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-end">
                        <div className="flex gap-sm justify-end">
                          <Button variant="ghost" size="sm" aria-label="View item">
                            <Eye className="w-4 h-4" />
                          </Button>
                          <Button variant="ghost" size="sm" aria-label="Edit item">
                            <Edit className="w-4 h-4" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
}

interface StatCardProps {
  title: string;
  value: number | string;
  icon: any;
  subtitle?: string;
  variant?: 'default' | 'featured' | 'warning' | 'ai' | 'danger' | 'success' | 'info';
  trend?: string | null;
  trendUp?: boolean | null;
}

function StatCard({ title, value, icon: Icon, subtitle, variant = 'default', trend, trendUp }: StatCardProps) {
  return (
    <Card 
      variant={variant} 
      className="hover:border-border/22 hover:-translate-y-1 cursor-pointer"
      hoverable
    >
      <CardContent className="p-lg">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <p className="text-small text-text-muted">{title}</p>
            <p className="text-metric font-bold text-text mt-1">
              {value}
            </p>
            {subtitle && (
              <p className="text-tiny text-text-muted mt-1">
                {subtitle}
              </p>
            )}
            {trend && (
              <div className="flex items-center gap-1 mt-2">
                {trendUp ? (
                  <TrendingUp className="w-3 h-3 text-green" />
                ) : (
                  <TrendingDown className="w-3 h-3 text-red" />
                )}
                <span className={`text-tiny ${trendUp ? 'text-green' : 'text-red'}`}>
                  {trend}
                </span>
              </div>
            )}
          </div>
          <div className={`w-10 h-10 rounded-sm flex items-center justify-center ${
            variant === 'featured' ? 'bg-cyan/10' :
            variant === 'warning' ? 'bg-yellow/10' :
            variant === 'danger' ? 'bg-red/10' :
            variant === 'success' ? 'bg-green/10' :
            variant === 'info' ? 'bg-cyan/10' :
            variant === 'ai' ? 'bg-cyan/10' :
            'bg-cyan/10'
          }`}>
            <Icon className={`w-5 h-5 ${
              variant === 'featured' ? 'text-cyan' :
              variant === 'warning' ? 'text-yellow' :
              variant === 'danger' ? 'text-red' :
              variant === 'success' ? 'text-green' :
              variant === 'info' ? 'text-cyan' :
              variant === 'ai' ? 'text-cyan' :
              'text-cyan'
            }`} />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}