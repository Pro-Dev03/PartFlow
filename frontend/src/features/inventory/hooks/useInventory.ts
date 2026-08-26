import { useState, useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { productsApi, inventoryApi, barcodeApi, categoriesApi } from '../../../services/api/endpoints';
import { Product, InventoryItem, FilterConfig, SortConfig } from '../types/inventory.types';

export function useInventory() {
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [sortConfig, setSortConfig] = useState<SortConfig>({ key: '', direction: null });
  const [filters, setFilters] = useState<FilterConfig[]>([]);

  // Fetch data
  const { data: productsData, isLoading: productsLoading } = useQuery({
    queryKey: ['products'],
    queryFn: () => productsApi.list({ page: 1, per_page: 100 }),
  });

  const { data: categoriesData } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoriesApi.list(),
  });

  const { data: inventoryData, isLoading: inventoryLoading } = useQuery({
    queryKey: ['inventory', filters],
    queryFn: () => {
      const params: any = { page: 1, per_page: 100 };
      
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

  const products = (productsData?.data?.products as Product[]) || [];
  const inventoryItems = (inventoryData?.data?.items as InventoryItem[]) || [];
  const categories = (categoriesData?.data as any[]) || [];

  // Create category map for easy lookup
  const categoryMap = useMemo(() => {
    const map = new Map<string, string>();
    categories.forEach((cat: any) => {
      map.set(cat.id, cat.name);
    });
    return map;
  }, [categories]);

  // Safe arrays
  const safeProducts = Array.isArray(products) ? products : [];
  const safeInventoryItems = Array.isArray(inventoryItems) ? inventoryItems : [];

  // Mutations
  const deleteProductMutation = useMutation({
    mutationFn: (productId: string) => productsApi.delete(productId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      toast.success('تم حذف المنتج بنجاح');
    },
    onError: (error: any) => {
      console.error('Delete product failed:', error);
      toast.error('فشل حذف المنتج');
    },
  });

  const archiveProductMutation = useMutation({
    mutationFn: (productId: string) => productsApi.archive(productId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      toast.success('تم أرشفة المنتج بنجاح');
    },
    onError: (error) => {
      console.error('Archive product failed:', error);
      toast.error('فشل أرشفة المنتج');
    },
  });

  const createProductMutation = useMutation({
    mutationFn: (data: any) => productsApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      toast.success('تم إضافة المنتج بنجاح');
    },
    onError: (error) => {
      console.error('Create product failed:', error);
      toast.error('فشل إضافة المنتج');
    },
  });

  const updateProductMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) => productsApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      toast.success('تم تحديث المنتج بنجاح');
    },
    onError: (error) => {
      console.error('Update product failed:', error);
      toast.error('فشل تحديث المنتج');
    },
  });

  // Filter and sort logic
  const processedProducts = useMemo(() => {
    let result = [...safeProducts];

    // Add category name to each product
    result = result.map((product: Product) => ({
      ...product,
      category_name: categoryMap.get(product.category_id) || product.category || product.category_name || '-'
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
  }, [safeProducts, searchQuery, filters, sortConfig, categoryMap]);

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
    return safeInventoryItems.filter((item: InventoryItem) => {
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
      const product = response as Product;
      
      if (product && product.id) {
        return product;
      }
      return null;
    } catch (error) {
      console.error('Barcode lookup failed:', error);
      // Fallback to local search
      const product = safeProducts.find((p: Product) => 
        p.sku === barcode || 
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
    
    // State
    searchQuery,
    setSearchQuery,
    sortConfig,
    setSortConfig,
    filters,
    setFilters,
    
    // Mutations
    deleteProductMutation,
    archiveProductMutation,
    createProductMutation,
    updateProductMutation,
    
    // Helpers
    lookupProduct,
  };
}