# Debt Metrics Evidence Matrix

## Source-of-truth contract

- **Debt Ledger:** `debts` is authoritative for original debt, current remaining debt, and debtor count.
- **Payment Ledger:** `payments` is authoritative for financial payment events and daily collections.
- **Customer Credit Ledger:** `customer_ledger` records customer account movements; credit from returns must not be treated as a payment.
- **Legacy compatibility:** `customer_payments` may contain historical records, but it is not included in financial payment aggregates. New payment writes use `payments`.

| Metric | Definition | Source | API/UI location | Status |
|---|---|---|---|---|
| Original Debt | `SUM(debts.amount)` | Debt Ledger | `/debts` `meta.summary.total_debt` | Canonical |
| Paid From Debts | `SUM(max(amount - remaining_amount, 0))` | Debt Ledger snapshot | `/debts` `meta.summary.paid_amount` | Canonical summary |
| Outstanding Debt | `SUM(debts.remaining_amount)` | Debt Ledger | Dashboard `outstandingDebts`; `/debts` `meta.summary.remaining_amount` | Canonical |
| Debtor Count | Distinct customers with `remaining_amount > 0` | Debt Ledger | Dashboard `outstandingDebtorCount`; `/debts` `meta.summary.customer_count` | Canonical |
| Today's Debt Collections | Sum of qualifying `payments` dated to the store day | Payment Ledger | Dashboard `todayDebtCollected` | Canonical |
| Historical Payment Total | Sum of payment events in `payments` | Payment Ledger | Payment history/report queries | Canonical |
| Customer Credit | Credit movements in `customer_ledger`, including return adjustments where applicable | Customer Credit Ledger | Customer account/report queries | Must not enter payment totals |

## Linking evidence

- Direct debt payment: `payments.id` is the payment identifier and `payments.reference` contains the `debt_id` in the direct debt endpoint.
- Sale-linked debt payment: `payments.sale_id` links the payment to the sale and the matching `debts.sale_id` identifies the debt origin.
- Customer-level multi-debt payment: `payments.id` is linked from `customer_ledger.reference_id`; the payment is allocated to open debts oldest first, so one payment can settle multiple debt rows.
- `customer_payments` is retained only for backward compatibility and is deliberately excluded from dashboard payment sums to prevent double counting.

## Acceptance checks

- A `250` debt paid as `100` then `150` reports outstanding `150`, then `0`.
- The same lifecycle reports paid `100`, then historical paid `250`.
- A matching legacy `customer_payments` row does not change today's collection total.
- Debt Summary is calculated by the backend across all debt rows and is independent of pagination.
- Closing a debt changes its remaining balance/status but does not delete payment history.

## Sale date contract

- **Official commercial sale date:** `sales.sale_date` (a calendar date in the configured Store Timezone).
- **Event timestamp:** `sales.created_at` (the instant at which the sale record was created; not a separate commercial day).
- New sales write `sale_date` from `accounting.StoreDate(eventTime)` and write `created_at` as the event instant.
- Local database startup normalizes legacy `sale_date` values from `created_at` using the configured Store Timezone.

| Surface | Date used | Contract status |
|---|---|---|
| Activity | `sale_date` for the displayed commercial day; `created_at` only for event ordering/time | Official Store Day is explicit in the API; gross invoice amount remains tax-inclusive |
| Dashboard Sales Chart | `sale_date`; legacy rows without it fall back to `created_at` converted with Store Timezone; revenue excludes tax | Same Store Day as reports and dashboard metrics |
| Dashboard metrics (Today Sales/Profit) | `sale_date` with legacy fallback to `created_at` | Same Store Day as Chart and reports; `created_at` remains an event timestamp only |
| Sales and Net Sales reports | `sale_date`; revenue excludes tax where defined | Official commercial day |
| Profit and Tax reports | `sale_date` | Official commercial day |
| Daily/Monthly sales aggregations | `sale_date` when present, legacy fallback to `created_at` | Official day with legacy compatibility |
| Returns comparison | `sale_date` against the return's commercial date | Official commercial day |

Boundary evidence: `TestSaleDayContractUsesSaleDateAcrossChartAndActivity` covers a sale before midnight, at 23:59, at 00:00, and after midnight. A sale recorded at `2026-09-11T22:33:09Z` in `Asia/Jerusalem` is Store Day `2026-09-12`; its normalized `sale_date`, Activity entry, chart point, and reports must all use `2026-09-12`.

Re-scan result: remaining sale-related `created_at` references are limited to event ordering/display time, legacy-row fallback, or non-sale event domains such as payments, expenses, inventory, and returns. No remaining sales-day report path uses `created_at` as a competing commercial-day definition.

The canonical cross-entity temporal contract is maintained in [MASTER-EVIDENCE-MATRIX.md](MASTER-EVIDENCE-MATRIX.md); this document retains the debt-specific evidence and links to the shared sale-date decision.
