import { useState, useEffect } from 'react';
import { searchApi } from '../../services/api/endpoints';
import { Card, CardContent } from '../ui/card';
import { Input } from '../ui/input';
import { Button } from '../ui/button';
import { 
  Search, 
  X, 
  Package, 
  ShoppingCart, 
  Users, 
  Truck,
  FileText,
  Barcode
} from 'lucide-react';

interface SearchResult {
  type: 'product' | 'item' | 'customer' | 'supplier' | 'sale' | 'invoice';
  id: string;
  title: string;
  subtitle?: string;
  url: string;
}

export function GlobalSearch() {
  const [isOpen, setIsOpen] = useState(false);
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<SearchResult[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setIsOpen(true);
      }
      if (e.key === 'Escape') {
        setIsOpen(false);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  useEffect(() => {
    const search = async () => {
      if (query.length < 2) {
        setResults([]);
        return;
      }

      setIsLoading(true);
      try {
        const response = await searchApi.global(query);
        setResults((response.data as SearchResult[]) || []);
      } catch (error) {
        console.error('Search failed:', error);
        setResults([]);
      } finally {
        setIsLoading(false);
      }
    };

    const debounceTimer = setTimeout(search, 300);
    return () => clearTimeout(debounceTimer);
  }, [query]);

  const getResultIcon = (type: string) => {
    switch (type) {
      case 'product':
      case 'item':
        return Package;
      case 'sale':
      case 'invoice':
        return ShoppingCart;
      case 'customer':
        return Users;
      case 'supplier':
        return Truck;
      default:
        return FileText;
    }
  };

  const getResultTypeLabel = (type: string) => {
    switch (type) {
      case 'product':
        return 'منتج';
      case 'item':
        return 'قطعة';
      case 'sale':
        return 'بيع';
      case 'invoice':
        return 'فاتورة';
      case 'customer':
        return 'عميل';
      case 'supplier':
        return 'مورد';
      default:
        return type;
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-24 px-4">
      <div className="absolute inset-0 bg-black/50" onClick={() => setIsOpen(false)} />
      <Card className="relative w-full max-w-2xl max-h-[600px] overflow-hidden">
        <CardContent className="p-0">
          {/* Search Input */}
          <div className="flex items-center border-b border-gray-200 dark:border-gray-700 p-4">
            <Search className="w-5 h-5 text-gray-400 me-3" />
            <Input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="بحث عن منتجات، عملاء، موردين، فواتير..."
              className="border-0 focus-visible:ring-0 px-0 text-lg"
              autoFocus
            />
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setIsOpen(false)}
              className="ms-2"
            >
              <X className="w-4 h-4" />
            </Button>
          </div>

          {/* Search Results */}
          <div className="overflow-y-auto max-h-[500px] p-4">
            {isLoading && (
              <div className="flex items-center justify-center py-8">
                <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary-600" />
              </div>
            )}

            {!isLoading && query.length < 2 && (
              <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                <Search className="w-12 h-12 mx-auto mb-3 opacity-50" />
                <p className="text-lg font-medium">ابدأ البحث</p>
                <p className="text-sm mt-1">
                  اضغط <kbd className="px-2 py-1 bg-gray-100 dark:bg-gray-800 rounded text-xs">Ctrl</kbd> + <kbd className="px-2 py-1 bg-gray-100 dark:bg-gray-800 rounded text-xs">K</kbd> لفتح البحث
                </p>
              </div>
            )}

            {!isLoading && query.length >= 2 && results.length === 0 && (
              <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                <p className="text-lg font-medium">لا توجد نتائج</p>
                <p className="text-sm mt-1">جرب كلمات مفتاحية مختلفة</p>
              </div>
            )}

            {!isLoading && results.length > 0 && (
              <div className="space-y-2">
                {results.map((result) => {
                  const Icon = getResultIcon(result.type);
                  return (
                    <button
                      key={result.id}
                      className="w-full flex items-center gap-3 p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors text-right"
                      onClick={() => {
                        window.location.href = result.url;
                        setIsOpen(false);
                      }}
                    >
                      <div className="p-2 bg-gray-100 dark:bg-gray-700 rounded-lg">
                        <Icon className="w-4 h-4 text-gray-600 dark:text-gray-400" />
                      </div>
                      <div className="flex-1 text-right">
                        <p className="font-medium text-gray-900 dark:text-gray-100">
                          {result.title}
                        </p>
                        {result.subtitle && (
                          <p className="text-sm text-gray-500 dark:text-gray-400">
                            {result.subtitle}
                          </p>
                        )}
                      </div>
                      <Badge variant="outline" className="text-xs">
                        {getResultTypeLabel(result.type)}
                      </Badge>
                    </button>
                  );
                })}
              </div>
            )}
          </div>

          {/* Footer */}
          <div className="border-t border-gray-200 dark:border-gray-700 p-3 bg-gray-50 dark:bg-gray-800">
            <div className="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
              <div className="flex items-center gap-4">
                <span className="flex items-center gap-1">
                  <kbd className="px-1.5 py-0.5 bg-gray-100 dark:bg-gray-700 rounded">↑↓</kbd>
                  للتنقل
                </span>
                <span className="flex items-center gap-1">
                  <kbd className="px-1.5 py-0.5 bg-gray-100 dark:bg-gray-700 rounded">Enter</kbd>
                  للاختيار
                </span>
                <span className="flex items-center gap-1">
                  <kbd className="px-1.5 py-0.5 bg-gray-100 dark:bg-gray-700 rounded">Esc</kbd>
                  للإغلاق
                </span>
              </div>
              <div className="flex items-center gap-2">
                <Button variant="ghost" size="sm" className="gap-1 text-xs">
                  <Barcode className="w-3 h-3" />
                  مسح باركود
                </Button>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function Badge({ children, variant, className }: { children: React.ReactNode; variant?: string; className?: string }) {
  return (
    <span className={`px-2 py-1 rounded-full text-xs font-medium ${
      variant === 'outline' 
        ? 'border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300' 
        : 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300'
    } ${className || ''}`}>
      {children}
    </span>
  );
}