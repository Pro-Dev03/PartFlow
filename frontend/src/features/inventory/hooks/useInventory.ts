import { useState, useMemo, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { productsApi, inventoryApi, barcodeApi, categoriesApi, suppliersApi } from '../../../services/api/endpoints';
import { Product, InventoryItem, FilterConfig, SortConfig } from '../types/inventory.types';
import { useDebounce } from '../../../hooks/useDebounce';
import { Category } from '../../../types/models';
import { getLocalProductImage } from '../../../services/localProductImages';
import { getCategoryImage } from '../../../services/localCategoryImages';

export function useInventory() {
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const debouncedSearchQuery = useDebounce(searchQuery, 300);
  const [sortConfig, setSortConfig] = useState<SortConfig>({ key: '', direction: null });
  const [filters, setFilters] = useState<FilterConfig[]>([]);
  const [productPage, setProductPage] = useState(1);
  const [inventoryPage, setInventoryPage] = useState(1);
  const pageSize = 10;
  const lowStockOnly = filters.some((filter) => filter.key === 'low_stock');
  const manualOnly = filters.some((filter) => filter.key === 'manual_only' && String(filter.value).toLowerCase() === 'true');
  const supplierOnly = filters.some((filter) => filter.key === 'supplier_only' && String(filter.value).toLowerCase() === 'true');

  // Fetch data with debounce search for scalability
  const { data: productsData, isLoading: productsLoading, refetch: refetchProducts } = useQuery({
    queryKey: ['products', lowStockOnly ? 'low-stock' : productPage, lowStockOnly ? 1000 : pageSize, debouncedSearchQuery],
    queryFn: () => {
      if (debouncedSearchQuery) {
        // Search mode - use API search when query exists
        return productsApi.list({
          page: lowStockOnly ? 1 : productPage,
          per_page: lowStockOnly ? 1000 : pageSize,
          search: debouncedSearchQuery
        });
      } else {
        // Initial load - fetch limited results for performance
        return productsApi.list({
          page: lowStockOnly ? 1 : productPage,
          per_page: lowStockOnly ? 1000 : pageSize,
        });
      }
    },
    enabled: true, // Always enabled, but will refetch when search changes
  });

  const { data: categoriesData } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoriesApi.list(),
  });

  const { data: suppliersData } = useQuery({
    queryKey: ['suppliers', 'inventory-products'],
    queryFn: () => suppliersApi.list({ page: 1, per_page: 100, is_active: true }),
  });

  const { data: inventoryData, isLoading: inventoryLoading, refetch: refetchInventory } = useQuery({
    queryKey: ['inventory', inventoryPage, pageSize, debouncedSearchQuery, filters],
    queryFn: () => {
      const supplierOnlyFilter = filters.find(f => f.key === 'supplier_only');
      const manualOnlyFilter = filters.find(f => f.key === 'manual_only');
      const params: any = { page: inventoryPage, per_page: pageSize };
      if (supplierOnlyFilter?.value === 'true') {
        params.supplier_only = 'true';
      } else if (manualOnlyFilter?.value === 'true') {
        params.manual_only = 'true';
      } else {
        params.exclude_condition = 'USED';
      }
      
      // Apply search
      if (debouncedSearchQuery) {
        params.search = debouncedSearchQuery;
      }

      const categoryFilter = filters.find(f => f.key === 'category_id');
      if (categoryFilter && categoryFilter.value) {
        params.category_id = categoryFilter.value;
      }
      
      // Apply supplier filter
      const supplierFilter = filters.find(f => f.key === 'supplier_id');
      if (supplierFilter && supplierFilter.value) {
        params.supplier_id = supplierFilter.value;
      }
      
      // Apply purchase date filters
      const purchaseDateFrom = filters.find(f => f.key === 'purchase_date_from');
      if (purchaseDateFrom && purchaseDateFrom.value) {
        params.purchase_date_from = purchaseDateFrom.value;
      }
      
      const purchaseDateTo = filters.find(f => f.key === 'purchase_date_to');
      if (purchaseDateTo && purchaseDateTo.value) {
        params.purchase_date_to = purchaseDateTo.value;
      }
      
      // Apply purchase cost filters
      const minPurchaseCost = filters.find(f => f.key === 'min_purchase_cost');
      if (minPurchaseCost && minPurchaseCost.value) {
        params.min_purchase_cost = parseFloat(minPurchaseCost.value);
      }
      
      const maxPurchaseCost = filters.find(f => f.key === 'max_purchase_cost');
      if (maxPurchaseCost && maxPurchaseCost.value) {
        params.max_purchase_cost = parseFloat(maxPurchaseCost.value);
      }
      
      return inventoryApi.listWithSupplier(params);
    },
  });

  const { data: completeInventoryData } = useQuery({
    queryKey: ['inventory', 'complete-general-stock', manualOnly ? 'manual' : 'all'],
    queryFn: () => inventoryApi.listWithSupplier({ page: 1, per_page: 1000, exclude_condition: 'USED', ...(manualOnly ? { manual_only: 'true' } : {}) }),
    staleTime: 60000,
  });

  useEffect(() => {
    setProductPage(1);
    setInventoryPage(1);
  }, [debouncedSearchQuery, filters]);

  const products = (productsData?.data?.products as Product[]) || [];
  const inventoryItems = (inventoryData?.data?.items as InventoryItem[]) || [];
  const categories = (categoriesData?.data as Category[]) || [];
  const suppliers = Array.isArray(suppliersData?.data)
    ? suppliersData.data as Array<{ id: string; name?: string; supplier_name?: string }>
    : ((suppliersData?.data as any)?.suppliers || (suppliersData as any)?.suppliers || []);

  // Create category map for easy lookup
  const categoryMap = useMemo(() => {
    const map = new Map<string, string>();
    categories.forEach((cat: Category) => {
      map.set(cat.id, cat.name);
    });
    return map;
  }, [categories]);

  const supplierMap = useMemo(() => {
    const map = new Map<string, string>();
    suppliers.forEach((supplier) => {
      const name = supplier.name || supplier.supplier_name || '';
      if (supplier.id && name) map.set(String(supplier.id), name);
    });
    return map;
  }, [suppliers]);

  // Safe arrays
  const safeProducts = Array.isArray(products) ? products : [];
  const safeInventoryItems = Array.isArray(inventoryItems) ? inventoryItems : [];
  const completeInventoryItems = Array.isArray(completeInventoryData?.data?.items)
    ? completeInventoryData.data.items as InventoryItem[]
    : safeInventoryItems;
  const completeInventoryProductCount = useMemo(() => new Set(
    completeInventoryItems
      .map((item: any) => String(item.product_id || item.product?.id || item.productId || '').trim())
      .filter(Boolean)
  ).size, [completeInventoryItems]);

  const productsWithInventoryFallback = useMemo(() => {
    const productsById = new Map(safeProducts.map((product) => [String(product.id), product]));

    [...completeInventoryItems, ...safeInventoryItems].forEach((item: any) => {
      const productId = String(item.product_id || item.product?.id || item.productId || '').trim();
      if (!productId) return;

      const supplierId = String(item.supplier_id || item.supplier?.id || '').trim();
      const supplierName = String(item.supplier_name || item.supplier?.name || '').trim();
      const itemCondition = String(item.condition || '').trim();
      const existingProduct = productsById.get(productId);
      if (existingProduct) {
        productsById.set(productId, {
          ...existingProduct,
          supplier_id: existingProduct.supplier_id || (existingProduct as any).preferred_supplier_id || supplierId,
          supplier_name: existingProduct.supplier_name || supplierName || supplierMap.get(String(existingProduct.supplier_id || (existingProduct as any).preferred_supplier_id || supplierId)) || '',
          condition: existingProduct.condition || itemCondition,
          category_id: existingProduct.category_id || item.category_id,
          category_name: existingProduct.category_name || item.category_name,
        });
        return;
      }

      // The products endpoint is paginated. Inventory rows fetched for totals
      // must not append every product to the current page and render the full
      // catalogue. Use inventory as a product fallback only when that endpoint
      // has no products to display.
      if (safeProducts.length > 0) return;

      const productName = String(item.product_name || item.product?.name || '').trim();
      if (!productName) return;

      const stock = Number(item.available_quantity ?? item.current_quantity ?? item.stock ?? item.quantity ?? 0);
      productsById.set(productId, {
        id: productId,
        name: productName,
        sku: String(item.item_code || productId),
        sellingPrice: Number(item.product_selling_price ?? item.selling_price ?? item.price ?? 0),
        costPrice: Number(item.purchase_cost ?? 0),
        stock: Number.isFinite(stock) ? stock : 0,
        condition: String(item.condition || ''),
        category: item.category_name,
        category_id: item.category_id,
        barcode: item.barcode,
        supplier_id: supplierId,
        supplier_name: supplierName || supplierMap.get(supplierId) || '',
      });
    });

    return Array.from(productsById.values()).map((product: any) => ({
      ...product,
      supplier_name: product.supplier_name || supplierMap.get(String(product.supplier_id || product.preferred_supplier_id || '')) || '',
    }));
  }, [completeInventoryItems, safeInventoryItems, safeProducts, supplierMap]);

  const isUsedItemCondition = (value: unknown) => {
    const condition = String(value ?? '').trim().toUpperCase();
    return condition === 'USED' || condition.includes('USED');
  };

  const regularProductIds = useMemo(() => {
    const inactiveStatuses = new Set(['SOLD', 'RETURNED', 'REVERSED', 'CANCELLED', 'DELETED', 'VOID']);
    return new Set(
      safeInventoryItems
        .filter((item: any) => {
          const status = String(item.status || '').trim().toUpperCase();
          const quantity = Number(item.available_quantity ?? item.current_quantity ?? item.stock ?? item.quantity ?? 0);
          return !isUsedItemCondition(item.condition) &&
            !inactiveStatuses.has(status) &&
            (quantity > 0 || status === 'AVAILABLE');
        })
        .map((item: any) => String(item.product_id || item.product?.id || item.productId || '').trim())
        .filter(Boolean)
    );
  }, [safeInventoryItems]);

  const usedProductIds = useMemo(() => new Set(
    safeInventoryItems
      .filter((item: any) => isUsedItemCondition(item.condition))
      .map((item: any) => String(item.product_id || item.product?.id || item.productId || '').trim())
      .filter(Boolean)
  ), [safeInventoryItems]);

  const regularProducts = useMemo(
    () => productsWithInventoryFallback.filter((product: Product) => {
      if (manualOnly || supplierOnly) {
        const hasMatchingInventory = [...completeInventoryItems, ...safeInventoryItems].some((item: any) =>
          String(item.product_id || item.product?.id || item.productId || '') === String(product.id) &&
          (manualOnly
            ? !String(item.supplier_id || item.supplier?.id || '').trim()
            : Boolean(String(item.supplier_id || item.supplier?.id || '').trim()))
        );
        if (!hasMatchingInventory) return false;
      }
      const sku = String((product as any).sku || '').trim().toUpperCase();
      const isUsedProduct = usedProductIds.has(product.id) || sku.startsWith('USED-');
      return !isUsedProduct || regularProductIds.has(product.id);
    }),
    [productsWithInventoryFallback, regularProductIds, usedProductIds, manualOnly, supplierOnly, completeInventoryItems, safeInventoryItems]
  );

  const inventoryStockMap = useMemo(() => {
    const map = new Map<string, number>();
    const inactiveStatuses = new Set(['SOLD', 'RETURNED', 'REVERSED', 'CANCELLED', 'DELETED', 'VOID']);

    completeInventoryItems.forEach((item: any) => {
      const productId = String(item.product_id || item.product?.id || item.productId || '').trim();
      if (!productId) return;
      if (isUsedItemCondition(item.condition)) return;

      const status = String(item.status || '').trim().toUpperCase();
      if (inactiveStatuses.has(status)) return;

      const explicitStock = Number(
        item.available_quantity ??
        item.current_quantity ??
        item.stock ??
        item.quantity ??
        0
      );

      const fallbackStock = status === 'AVAILABLE' ? 1 : 0;
      const stock = Number.isFinite(explicitStock) && explicitStock > 0 ? explicitStock : fallbackStock;

      if (stock <= 0) return;

      // available_quantity/current_quantity are product-level totals repeated on
      // every item row, so do not add them once per row.
      const hasProductTotal = item.available_quantity !== undefined || item.current_quantity !== undefined;
      const current = map.get(productId) || 0;
      map.set(productId, hasProductTotal ? Math.max(current, stock) : current + stock);
    });

    return map;
  }, [completeInventoryItems]);

  const inventoryConditionMap = useMemo(() => {
    const map = new Map<string, string>();
    safeInventoryItems.forEach((item: any) => {
      if (isUsedItemCondition(item.condition)) return;
      const productId = String(item.product_id || item.product?.id || item.productId || '').trim();
      const condition = String(item.condition || '').trim();
      if (productId && condition && !map.has(productId)) {
        map.set(productId, condition);
      }
    });
    return map;
  }, [safeInventoryItems]);

  // Mutations
  const deleteProductMutation = useMutation({
    mutationFn: (productId: string) => productsApi.delete(productId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
      queryClient.invalidateQueries({ queryKey: ['reports'] });
      toast.success('تم حذف المنتج بنجاح');
    },
    onError: (error: any) => {
      console.error('Delete product failed:', error);
      const message = String(error?.arabicMessage || error?.message || '');
      toast.error(message || 'فشل حذف المنتج');
    },
  });

  const deleteInventoryItemMutation = useMutation({
    mutationFn: (itemId: string) => inventoryApi.delete(itemId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
      queryClient.invalidateQueries({ queryKey: ['reports'] });
      toast.success('تمت إزالة عنصر المخزون من القائمة وحفظه في السجل');
    },
    onError: (error: any) => {
      console.error('Delete inventory item failed:', error);
      toast.error(error?.message || 'تعذر حذف عنصر المخزون');
    },
  });

  const createProductMutation = useMutation({
    mutationFn: (data: any) => productsApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      toast.success('تم إضافة المنتج بنجاح');
    },
    onError: (error: any) => {
      console.error('Create product failed:', error);
      toast.error(error?.arabicMessage || error?.message || 'فشل إضافة المنتج');
    },
  });

  const updateProductMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) => productsApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
      queryClient.invalidateQueries({ queryKey: ['low-stock-items'] });
      queryClient.invalidateQueries({ queryKey: ['reports'] });
      toast.success('تم تحديث المنتج بنجاح');
    },
    onError: (error) => {
      console.error('Update product failed:', error);
      toast.error('فشل تحديث المنتج');
    },
  });

  const updateMinimumStockMutation = useMutation({
    mutationFn: ({ id, minStockLevel }: { id: string; minStockLevel: number }) => productsApi.updateMinimumStock(id, minStockLevel),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
      queryClient.invalidateQueries({ queryKey: ['low-stock-items'] });
      queryClient.invalidateQueries({ queryKey: ['reports'] });
      toast.success('تم تحديث الحد الأدنى للمخزون بنجاح');
    },
    onError: (error) => {
      console.error('Update minimum stock failed:', error);
      toast.error('فشل تحديث الحد الأدنى للمخزون');
    },
  });

  // Filter and sort logic
  const processedProducts = useMemo(() => {
    let result = [...regularProducts];

    // Add category name to each product
    result = result.map((product: Product) => ({
      ...product,
      category_name: categoryMap.get(product.category_id) || product.category || product.category_name || '-',
      condition: product.condition || inventoryConditionMap.get(product.id) || '',
       stock: Number(
         inventoryStockMap.get(product.id) ??
         product.current_quantity ??
         product.stock ??
         product.quantity ??
         0
       ),
      image_url: product.image_url || getLocalProductImage(product.id) || (product.category_id ? getCategoryImage(product.category_id) : undefined),
    }));

    // Search filter
    if (searchQuery) {
      result = result.filter((product: Product) =>
        product.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        product.sku.toLowerCase().includes(searchQuery.toLowerCase())
      );
    }

    // Additional filters
    if (filters.length > 0) {
      result = result.filter((product: Product) => {
        return filters.every((filter) => {
          // Handle category filter specially
          if (filter.key === 'category_id') {
            return product.category_id === filter.value;
          }

          if (filter.key === 'low_stock') {
            const stock = Number(product.stock ?? 0);
            const minimumStock = Math.max(1, Number(product.min_stock_level) || 3);
            return stock > 0 && stock <= minimumStock;
          }

          if (filter.key === 'out_of_stock') {
            return Number(product.stock ?? 0) <= 0;
          }

          if (filter.key === 'manual_only' || filter.key === 'supplier_only') {
            return true;
          }

          const value = product[filter.key as keyof Product];
          const filterValue = filter.value.toLowerCase();

          if (value === undefined || value === null) return false;

          const itemValue = String(value).toLowerCase();
          return itemValue.includes(filterValue);
        });
      });
    }

    // Sorting
    if (sortConfig.key && sortConfig.direction) {
      result.sort((a: Product, b: Product) => {
        const aValue = a[sortConfig.key as keyof Product];
        const bValue = b[sortConfig.key as keyof Product];

        if (aValue === bValue) return 0;

        const comparison = aValue < bValue ? -1 : 1;
        return sortConfig.direction === 'asc' ? comparison : -comparison;
      });
    }

    return result;
  }, [regularProducts, searchQuery, filters, sortConfig, categoryMap, inventoryStockMap, inventoryConditionMap]);

  const filteredProducts = useMemo(() => {
    return processedProducts.filter((product: Product) => {
      if (filters.length === 0) return true;
      
      return filters.every((filter) => {
        if (filter.key === 'condition') {
          return product.condition === filter.value;
        }
        return true;
      });
    });
  }, [processedProducts, filters]);

  const filteredInventoryItems = useMemo(() => {
    return safeInventoryItems.map((item: InventoryItem) => {
      const product = safeProducts.find((candidate) => candidate.id === item.product_id);
      return {
        ...item,
        category_id: item.category_id || product?.category_id,
        category_name: item.category_name || (product?.category_id ? categoryMap.get(product.category_id) : undefined),
      };
    }).filter((item: InventoryItem) => {
      if (filters.length === 0) return true;
      
      return filters.every((filter) => {
        if (filter.key === 'condition') {
          return item.condition === filter.value;
        }
        return true;
      });
    });
  }, [safeInventoryItems, filters]);

  // Barcode lookup
  const lookupProduct = async (barcode: string): Promise<Product | null> => {
    try {
      const response = await barcodeApi.lookupProduct(barcode);
      const product = response.data as Product;
      
      if (product && product.id) {
        return product;
      }
      return null;
    } catch (error) {
      console.error('Barcode lookup failed:', error);
      // Fallback to local search
      const product = safeProducts.find((p: Product) => 
        p.barcode === barcode
      );
      return product || null;
    }
  };

  return {
    // Data
    products: safeProducts,
    inventoryItems: safeInventoryItems,
    filteredProducts,
    filteredInventoryItems,
    productsLoading,
    inventoryLoading,
    inventoryStockMap,
    
    // State
    searchQuery,
    setSearchQuery,
    sortConfig,
    setSortConfig,
    filters,
    setFilters,
    refetch: async () => {
      await Promise.all([refetchProducts(), refetchInventory()]);
    },
    
    // Mutations
    deleteProductMutation,
    deleteInventoryItemMutation,
    createProductMutation,
    updateProductMutation,
    updateMinimumStockMutation,
    
    // Helpers
    lookupProduct,
    productPage,
    inventoryPage,
    pageSize,
    productTotal: Math.max(Number(productsData?.meta?.total || 0), completeInventoryProductCount, productsWithInventoryFallback.length),
    inventoryTotal: Number(inventoryData?.meta?.total || safeInventoryItems.length),
    setProductPage,
    setInventoryPage,
  };
}
