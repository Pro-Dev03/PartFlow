export interface SalesCleanupProgress {
  completed: number;
  total: number;
}

export interface SalesCleanupSummary {
  deleted: number;
  blocked: number;
  total: number;
}

export class SalesCleanupError extends Error {
  readonly summary: SalesCleanupSummary;

  constructor(summary: SalesCleanupSummary, cause?: unknown) {
    super('Sales cleanup stopped before all records were processed.', { cause });
    this.name = 'SalesCleanupError';
    this.summary = summary;
  }
}

type SaleListRequest = { page: number; per_page: number };
type ListSales = (params: SaleListRequest) => Promise<unknown>;
type DeleteSale = (saleId: string) => Promise<unknown>;

function salesPage(response: unknown) {
  let payload: any = (response as any)?.data ?? response;
  if (!Array.isArray(payload)
    && !Array.isArray(payload?.sales)
    && !Array.isArray(payload?.items)
    && payload?.data) {
    payload = payload.data;
  }

  const rows = Array.isArray(payload)
    ? payload
    : payload?.sales ?? payload?.items ?? payload?.data?.sales ?? payload?.data?.items ?? [];
  const total = Number(payload?.total ?? payload?.meta?.total);
  return { rows: rows as Array<Record<string, unknown>>, total };
}

function deleteResult(response: unknown) {
  const payload: any = (response as any)?.data ?? response;
  return payload?.data ?? payload;
}

function responseStatus(error: unknown): number {
  return Number((error as any)?.status ?? (error as any)?.response?.status);
}

export async function cleanSalesHistoryRecords(
  listSales: ListSales,
  deleteSale: DeleteSale,
  onProgress?: (progress: SalesCleanupProgress) => void,
): Promise<SalesCleanupSummary> {
  const pageSize = 100;
  const allSales: Array<Record<string, unknown>> = [];
  const firstPage = salesPage(await listSales({ page: 1, per_page: pageSize }));
  allSales.push(...firstPage.rows);
  const pageCount = Number.isFinite(firstPage.total) && firstPage.total > 0
    ? Math.ceil(firstPage.total / pageSize)
    : undefined;

  for (let page = 2; pageCount ? page <= pageCount : firstPage.rows.length === pageSize; page += 1) {
    const nextPage = salesPage(await listSales({ page, per_page: pageSize }));
    if (nextPage.rows.length === 0) break;
    allSales.push(...nextPage.rows);
    if (!pageCount && nextPage.rows.length < pageSize) break;
  }

  const saleIds = [...new Set(allSales.map((sale) => String(sale.id ?? '').trim()).filter(Boolean))];
  const summary: SalesCleanupSummary = { deleted: 0, blocked: 0, total: saleIds.length };
  onProgress?.({ completed: 0, total: summary.total });

  for (let index = 0; index < saleIds.length; index += 1) {
    try {
      const result = deleteResult(await deleteSale(saleIds[index]));
      if (result?.action === 'blocked' || result?.action === 'not_found') summary.blocked += 1;
      else summary.deleted += 1;
    } catch (error) {
      const status = responseStatus(error);
      if (status === 409 || status === 404) {
        summary.blocked += 1;
      } else {
        throw new SalesCleanupError(summary, error);
      }
    }
    onProgress?.({ completed: index + 1, total: summary.total });
  }

  return summary;
}
