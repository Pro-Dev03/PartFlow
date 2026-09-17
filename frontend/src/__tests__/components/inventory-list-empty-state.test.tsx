import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '../../test/utils';
import { InventoryList } from '../../features/inventory/components/InventoryList';

describe('InventoryList empty state', () => {
  it('renders a single empty-state message when no products match', () => {
    render(
      <InventoryList
        viewMode="products"
        filteredProducts={[]}
        filteredInventoryItems={[]}
        productsLoading={false}
        inventoryLoading={false}
        inventoryStockMap={new Map()}
        searchQuery="abc"
        onViewProduct={vi.fn()}
        onAddPurchase={vi.fn()}
        onEditProduct={vi.fn()}
        onEditMinimumStock={vi.fn()}
        onDeleteProduct={vi.fn()}
        onClearSearch={vi.fn()}
        layoutMode="table"
      />
    );

    expect(screen.getAllByText('لا توجد منتجات')).toHaveLength(1);
    expect(screen.getAllByText('لم يتم العثور على منتجات تطابق بحثك')).toHaveLength(1);
  });
});
