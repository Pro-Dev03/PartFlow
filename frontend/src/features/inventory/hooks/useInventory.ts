import { useState, useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { productsApi, inventoryApi, barcodeApi } from '../../../services/api/endpoints';
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

  const { data: inventoryData, isLoading: inventoryLoading } = useQuery({
    queryKey: ['inventory'],
    queryFn: () => inventoryApi.list({ page: 1, per_page: 100 }),
  });

  const products = (productsData?.data?.products as Product[]) || [];
  const inventoryItems = (inventoryData?.data as InventoryItem[]) || [];

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
    onError: (error) => {
      console.error('Delete product failed:', error);
      toast.error('فشل حذف المنتج');
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
  }, [safeProducts, searchQuery, filters, sortConfig]);

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
    createProductMutation,
    updateProductMutation,
    
    // Helpers
    lookupProduct,
  };
}