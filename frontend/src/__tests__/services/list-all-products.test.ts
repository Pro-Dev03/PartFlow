import { afterEach, describe, expect, it, vi } from 'vitest';
import { listAllProducts, productsApi } from '../../services/api/endpoints';

describe('listAllProducts', () => {
  afterEach(() => vi.restoreAllMocks());

  it('fetches every capped page and preserves the product list response shape', async () => {
    const list = vi.spyOn(productsApi, 'list').mockImplementation(async (params) => {
      const page = params?.page ?? 1;
      const start = (page - 1) * 100;
      const products = Array.from({ length: Math.max(0, Math.min(100, 205 - start)) }, (_, index) => ({
        id: `product-${start + index + 1}`,
        name: `Product ${start + index + 1}`,
      }));
      return {
        success: true,
        data: { products, total: 205, page, per_page: 100 },
      } as any;
    });

    const response = await listAllProducts({ search: 'brake' });

    expect(list).toHaveBeenCalledTimes(3);
    expect(list).toHaveBeenNthCalledWith(1, { search: 'brake', page: 1, per_page: 100 });
    expect(list).toHaveBeenNthCalledWith(2, { search: 'brake', page: 2, per_page: 100 });
    expect(list).toHaveBeenNthCalledWith(3, { search: 'brake', page: 3, per_page: 100 });
    expect(response.data.products).toHaveLength(205);
    expect(response.data.products[204]).toMatchObject({ id: 'product-205' });
    expect(response.data.total).toBe(205);
    expect(response.data.per_page).toBe(205);
  });
});
