import { describe, expect, it, vi } from 'vitest';
import { cleanSalesHistoryRecords, SalesCleanupError } from '../../../features/sales/utils/cleanSalesHistory';

describe('cleanSalesHistoryRecords', () => {
  it('loads every page before deleting and reports completed progress', async () => {
    const listSales = vi.fn(async ({ page }: { page: number; per_page: number }) => {
      if (page === 1) {
        return { data: { sales: Array.from({ length: 100 }, (_, index) => ({ id: `sale-${index}` })), total: 101 } };
      }
      return { data: { sales: [{ id: 'sale-100' }], total: 101 } };
    });
    const deleteSale = vi.fn(async (id: string) => ({ data: { action: id === 'sale-1' ? 'blocked' : 'deleted' } }));
    const progress: Array<{ completed: number; total: number }> = [];

    const result = await cleanSalesHistoryRecords(listSales, deleteSale, (value) => progress.push(value));

    expect(result).toEqual({ deleted: 100, blocked: 1, total: 101 });
    expect(listSales).toHaveBeenCalledTimes(2);
    expect(deleteSale).toHaveBeenCalledTimes(101);
    expect(progress[0]).toEqual({ completed: 0, total: 101 });
    expect(progress.at(-1)).toEqual({ completed: 101, total: 101 });
  });

  it('continues past sales blocked by financial dependencies', async () => {
    const listSales = vi.fn().mockResolvedValue({ sales: [{ id: 'blocked' }, { id: 'deletable' }], total: 2 });
    const deleteSale = vi.fn(async (id: string) => {
      if (id === 'blocked') throw Object.assign(new Error('sale has payments'), { status: 409 });
      return { action: 'deleted' };
    });

    await expect(cleanSalesHistoryRecords(listSales, deleteSale)).resolves.toEqual({
      deleted: 1,
      blocked: 1,
      total: 2,
    });
    expect(deleteSale).toHaveBeenCalledTimes(2);
  });

  it('stops on a temporary failure and retains an accurate partial result', async () => {
    const listSales = vi.fn().mockResolvedValue({ sales: [{ id: 'deleted' }, { id: 'pending' }, { id: 'untouched' }], total: 3 });
    const deleteSale = vi.fn(async (id: string) => {
      if (id === 'pending') throw Object.assign(new Error('service unavailable'), { status: 503 });
      return { action: 'deleted' };
    });

    const cleanup = cleanSalesHistoryRecords(listSales, deleteSale);

    await expect(cleanup).rejects.toMatchObject({
      name: 'SalesCleanupError',
      summary: { deleted: 1, blocked: 0, total: 3 },
    } satisfies Partial<SalesCleanupError>);
    expect(deleteSale.mock.calls.map(([id]) => id)).toEqual(['deleted', 'pending']);
  });

  it('deduplicates repeated sale ids from paged data', async () => {
    const listSales = vi.fn()
      .mockResolvedValueOnce({ data: { data: { sales: Array.from({ length: 100 }, (_, index) => ({ id: index === 99 ? 'duplicate' : `sale-${index}` })), total: 101 } } })
      .mockResolvedValueOnce({ data: { sales: [{ id: 'duplicate' }], total: 101 } });
    const deleteSale = vi.fn().mockResolvedValue({ action: 'deleted' });

    const result = await cleanSalesHistoryRecords(listSales, deleteSale);

    expect(result.total).toBe(100);
    expect(deleteSale).toHaveBeenCalledTimes(100);
  });
});
