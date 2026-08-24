import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { SearchInput } from '../../../components/ui/search-input';
import { Badge } from '../../../components/ui/badge';
import { Package } from 'lucide-react';
import { useTranslation } from '../../../hooks/useTranslation';

interface ProductSearchProps {
  products: any[];
  searchQuery: string;
  setSearchQuery: (value: string) => void;
  onClearSearch: () => void;
  quickAddMode: boolean;
  onProductClick: (product: any) => void;
}

export function ProductSearch({
  products,
  searchQuery,
  setSearchQuery,
  onClearSearch,
  quickAddMode,
  onProductClick,
}: ProductSearchProps) {
  const { t } = useTranslation();

  const filteredProducts = (products || []).filter((product: any) =>
    product.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    product.sku.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <Card style={{
      background: 'linear-gradient(135deg, var(--color-info-05) 0%, rgba(168, 85, 247, 0.05) 100%)',
      border: '1px solid var(--color-info-10)',
      backdropFilter: 'blur(10px)',
      transition: 'all 0.3s ease'
    }}
    onMouseEnter={(e) => {
      e.currentTarget.style.background = 'linear-gradient(135deg, var(--color-info-08) 0%, rgba(168, 85, 247, 0.08) 100%)';
      e.currentTarget.style.borderColor = 'var(--color-info-20)';
      e.currentTarget.style.transform = 'translateY(-2px)';
      e.currentTarget.style.boxShadow = '0 8px 25px var(--color-info-10)';
    }}
    onMouseLeave={(e) => {
      e.currentTarget.style.background = 'linear-gradient(135deg, var(--color-info-05) 0%, rgba(168, 85, 247, 0.05) 100%)';
      e.currentTarget.style.borderColor = 'var(--color-info-10)';
      e.currentTarget.style.transform = 'translateY(0)';
      e.currentTarget.style.boxShadow = 'none';
    }}>
      <CardHeader>
        <CardTitle style={{ 
          display: 'flex', 
          alignItems: 'center', 
          gap: '10px',
          fontSize: '15px',
          fontWeight: '600',
          color: 'var(--text-primary)'
        }}>
          {t('sales.recentProducts')}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="mb-sm">
          <SearchInput
            placeholder={t('common.search')}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            onClear={onClearSearch}
            size="sm"
            className="w-full md:w-[500px] lg:w-[600px]"
          />
        </div>
        
        <div style={{
          display: 'grid',
          gridTemplateColumns: quickAddMode ? 'repeat(auto-fill, minmax(120px, 1fr))' : 'repeat(auto-fill, minmax(180px, 1fr))',
          gap: '14px'
        }}>
          {filteredProducts.map((product: any) => (
            <div
              key={product.id}
              onClick={() => onProductClick({
                id: product.id,
                name: product.name,
                barcode: product.barcode || product.sku,
                price: product.sellingPrice,
                stock: product.stock
              })}
              style={{
                padding: quickAddMode ? '12px' : '16px',
                borderRadius: '12px',
                border: '1px solid var(--color-info-15)',
                background: 'linear-gradient(135deg, var(--bg-surface-elevated) 0%, var(--bg-surface) 100%)',
                cursor: 'pointer',
                transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                backdropFilter: 'blur(10px)'
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.borderColor = 'var(--color-info-40)';
                e.currentTarget.style.transform = 'translateY(-4px) scale(1.02)';
                e.currentTarget.style.boxShadow = '0 12px 30px var(--color-info-15)';
                e.currentTarget.style.background = 'linear-gradient(135deg, var(--color-info-10) 0%, rgba(168, 85, 247, 0.1) 100%)';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.borderColor = 'var(--color-info-15)';
                e.currentTarget.style.transform = 'translateY(0) scale(1)';
                e.currentTarget.style.boxShadow = 'none';
                e.currentTarget.style.background = 'linear-gradient(135deg, var(--bg-surface-elevated) 0%, var(--bg-surface) 100%)';
              }}
            >
              {quickAddMode ? (
                <div style={{ textAlign: 'center' }}>
                  <div style={{
                    fontSize: '12px',
                    fontWeight: '600',
                    color: 'var(--text-primary)',
                    whiteSpace: 'nowrap',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    marginBottom: '6px',
                    letterSpacing: '0.2px'
                  }}>
                    {product.name}
                  </div>
                  <div style={{
                    fontSize: '14px',
                    fontWeight: '700',
                    color: 'var(--color-info)',
                    textShadow: '0 0 20px var(--color-info-30)'
                  }}>
                    ₪{product.sellingPrice}
                  </div>
                </div>
              ) : (
                <>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '10px' }}>
                    <div style={{
                      width: '44px',
                      height: '44px',
                      borderRadius: '10px',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      background: 'linear-gradient(135deg, var(--color-info-15) 0%, rgba(168, 85, 247, 0.15) 100%)',
                      border: '1px solid var(--color-info-25)',
                      boxShadow: '0 4px 15px var(--color-info-10)'
                    }}>
                      <Package className="w-5 h-5 text-cyan-400" />
                    </div>
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <div style={{
                        fontSize: '14px',
                        fontWeight: '600',
                        color: 'var(--text-primary)',
                        whiteSpace: 'nowrap',
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        letterSpacing: '0.2px'
                      }}>
                        {product.name}
                      </div>
                      <div style={{
                        fontSize: '11px',
                        color: 'var(--text-secondary)',
                        fontWeight: '500'
                      }}>
                        {product.sku}
                      </div>
                    </div>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <div style={{
                      fontSize: '15px',
                      fontWeight: '700',
                      color: 'var(--color-info)',
                      textShadow: '0 0 20px var(--color-info-30)'
                    }}>
                      ₪{product.sellingPrice}
                    </div>
                    <Badge
                      variant={product.stock > 10 ? 'success' : 'warning'}
                      style={{
                        fontSize: '10px',
                        fontWeight: '600',
                        padding: '4px 8px',
                        borderRadius: '6px',
                        boxShadow: product.stock > 10
                          ? '0 2px 10px var(--color-success-20)'
                          : '0 2px 10px var(--color-warning-20)'
                      }}
                    >
                      {product.stock > 10 ? 'متوفر' : `${product.stock}`}
                    </Badge>
                  </div>
                </>
              )}
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}