import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from '../../../hooks/useTranslation';
import { inventoryApi, productsApi } from '../../../services/api/endpoints';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table';
import { Badge } from '../../../components/ui/badge';
import { 
  Package, 
  Search, 
  Plus, 
  Filter,
  ArrowUpDown,
  Eye,
  Edit,
  Trash2,
  PackageOpen
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
    const variants: Record<string, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }> = {
      new: { label: t('products.new'), variant: 'default' },
      used: { label: t('products.used'), variant: 'secondary' },
      refurbished: { label: t('products.refurbished'), variant: 'outline' },
      parts_only: { label: t('products.partsOnly'), variant: 'destructive' },
    };
    return variants[condition] || { label: condition, variant: 'default' };
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            {t('inventory.title')}
          </h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">
            إدارة المخزون والقطع
          </p>
        </div>
        <Button className="gap-2">
          <Plus className="w-4 h-4" />
          {t('inventory.addItem')}
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <StatCard title={t('inventory.totalItems')} value={inventoryItems.length} icon={Package} />
        <StatCard title={t('inventory.totalValue')} value="₪185,400" icon={Package} />
        <StatCard title={t('inventory.lowStock')} value="12" icon={Package} />
        <StatCard title="مستعمل" value="84" icon={Package} />
      </div>

      {/* Search and Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex flex-col md:flex-row gap-4">
            <div className="flex-1 relative">
              <Search className="absolute inset-y-0 end-3 w-4 h-4 text-gray-400" />
              <Input
                placeholder={t('common.search')}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pe-10"
              />
            </div>
            <div className="flex gap-2">
              <Button variant="outline" className="gap-2">
                <Filter className="w-4 h-4" />
                {t('common.filter')}
              </Button>
              <Button variant="outline" className="gap-2">
                <ArrowUpDown className="w-4 h-4" />
                ترتيب
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* View Toggle */}
      <div className="flex gap-2">
        <Button
          variant={viewMode === 'products' ? 'primary' : 'outline'}
          onClick={() => setViewMode('products')}
          className="gap-2"
        >
          <Package className="w-4 h-4" />
          {t('products.title')}
        </Button>
        <Button
          variant={viewMode === 'items' ? 'primary' : 'outline'}
          onClick={() => setViewMode('items')}
          className="gap-2"
        >
          <PackageOpen className="w-4 h-4" />
          {t('inventory.items')}
        </Button>
      </div>

      {/* Products Table */}
      {viewMode === 'products' && (
        <Card>
          <CardHeader>
            <CardTitle>المنتجات</CardTitle>
          </CardHeader>
          <CardContent>
            {productsLoading ? (
              <div className="flex items-center justify-center h-64">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600" />
              </div>
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
                        <Badge variant={product.status === 'active' ? 'default' : 'secondary'}>
                          {product.status === 'active' ? 'نشط' : 'غير نشط'}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-end">
                        <div className="flex gap-2 justify-end">
                          <Button variant="ghost" size="sm">
                            <Eye className="w-4 h-4" />
                          </Button>
                          <Button variant="ghost" size="sm">
                            <Edit className="w-4 h-4" />
                          </Button>
                          <Button variant="ghost" size="sm">
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

      {/* Items Table */}
      {viewMode === 'items' && (
        <Card>
          <CardHeader>
            <CardTitle>القطع</CardTitle>
          </CardHeader>
          <CardContent>
            {inventoryLoading ? (
              <div className="flex items-center justify-center h-64">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600" />
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
                        <Badge variant={item.status === 'available' ? 'default' : 'secondary'}>
                          {item.status === 'available' ? 'متاح' : item.status}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-end">
                        <div className="flex gap-2 justify-end">
                          <Button variant="ghost" size="sm">
                            <Eye className="w-4 h-4" />
                          </Button>
                          <Button variant="ghost" size="sm">
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
}

function StatCard({ title, value, icon: Icon }: StatCardProps) {
  return (
    <Card>
      <CardContent className="p-6">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm text-gray-500 dark:text-gray-400">{title}</p>
            <p className="text-2xl font-bold text-gray-900 dark:text-gray-100 mt-1">
              {value}
            </p>
          </div>
          <Icon className="w-5 h-5 text-gray-400" />
        </div>
      </CardContent>
    </Card>
  );
}