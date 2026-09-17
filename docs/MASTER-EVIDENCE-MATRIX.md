# Master Evidence Matrix

## Evidence classification standard

This matrix uses the same evidence rules that define the current review baseline:

- **PASS** = proven by an executable test or live evidence for the complete claim, using the actual source chain (`DB -> Query -> Backend -> API` or equivalent service boundary for the relevant scope).
- **PARTIAL** = some layers are proven, but not all required layers are evidenced yet.
- **NOT PROVEN** = no sufficient evidence exists. This includes any case where a value is forced to match by fallback logic, aggregation rewriting, or `Math.max`-style masking without a documented source-of-truth contract.
- **Browser E2E** = not considered evidential unless a live browser test is actually executed and checked against the rendered values.

### Current frozen baseline (2026-09-16)

| Monthly value | Status | Official source | Evidence |
|---|---|---|---|
| Net Sales | PASS | `sales.sale_date`, sale totals adjusted for completed returns and tax treatment | Monthly SQLite reconciliation regression |
| Tax | PASS | `sales.tax_amount` within the same monthly sale set | Monthly SQLite reconciliation regression |
| Gross Profit | PASS | `sales.sale_date` + returned cost + completed return adjustments | Monthly SQLite reconciliation regression |
| Net Profit | PASS | same as above plus approved expenses | Monthly SQLite reconciliation regression |
| Monthly Returns | PARTIAL | `returns.return_date` + `return_items.total_refund_amount` | indirect effect on profit; direct monthly proof not yet recorded |
| Monthly Purchases | NOT PROVEN | `purchases.purchase_date` + `purchase_items` | no same-source monthly proof across ledger/query/backend/API |
| Monthly Supplier Balance | NOT PROVEN | `supplier_ledger` signed movements + `suppliers.current_balance` | no direct monthly parity proof |
| Monthly Customer Debt | NOT PROVEN | `debts.remaining_amount` / `customer_ledger` signed balances | no direct monthly parity proof |
| Dashboard/Reports UI parity | PARTIAL | same sources above, rendered values omitted | backend/API contract proven; rendered UI not proven |
| Browser E2E | NOT PROVEN | rendered browser output of dashboard and reports | no live browser evidence |

## Temporal contract

`Store Timezone` is the configured business timezone (`Asia/Jerusalem` by default). A **Business Date** is the date that defines which store-day, financial report period, or operational bucket an entity belongs to. A **Timestamp** records when the event was actually created or processed and must not silently replace the Business Date.

For sales, the approved contract is:

- `sales.sale_date` is the official commercial Business Date.
- `sales.created_at` is the actual record-creation timestamp.
- `created_at` may be used only for event ordering/display or as an explicit legacy fallback when `sale_date` is unavailable.
- Existing historical `sale_date` values must not be rewritten merely to make a chart agree with another field.

## Sales evidence

| Entity / surface | Official Business Date | Timestamp | Store Timezone | UI | API | DB | Status |
|---|---|---|---|---|---|---|---|
| Dashboard Sales | `sales.sale_date` | `sales.created_at` for event timestamp only | Yes | Dashboard metric `todaySales` | Dashboard stats endpoint | `sales.sale_date`, legacy fallback only when absent | PARTIAL: backend/API reconciliation PASS; displayed UI comparison not proven |
| Dashboard Profit | `sales.sale_date` | `sales.created_at` is not the financial day | Yes | Dashboard metric `todayProfit` | Dashboard stats endpoint | Sale-day filter plus accounting period bounds | PARTIAL: backend/API contract PASS; displayed UI comparison not proven |
| Activity | `sales.sale_date` for sale-day display | `sales.created_at` for ordering/actual event time | Yes | Dashboard and Activity page | Activity item `sale_date` and `time` | `sales.sale_date`, `sales.created_at` | PASS |
| Sales Chart | `sales.sale_date` | `sales.created_at` only for legacy fallback | Yes | Sales chart points | `salesChart[].name` | SQLite/PostgreSQL group by `sale_date` | PASS |
| Sales Report | `sales.sale_date` | `sales.created_at` is metadata only | Yes | Sales report | Sales report endpoint | Report filters/grouping use `sale_date` | PARTIAL: backend/API contract PASS; displayed UI comparison not proven |
| Net Sales | `sales.sale_date` | `sales.created_at` is metadata only | Yes | Net sales values | Dashboard/report response | Sale totals minus returns, filtered by sale date | PASS |
| Profit | `sales.sale_date` | `sales.created_at` is metadata only | Yes | Profit report | Profit report endpoint | Profit aggregation uses sale date | PARTIAL: backend/API contract PASS; displayed UI comparison not proven |
| Tax | `sales.sale_date` | `sales.created_at` is metadata only | Yes | Tax report | Tax report endpoint | Tax aggregation uses sale date | PARTIAL: backend/API contract PASS; displayed UI comparison not proven |
| Daily aggregations | `sales.sale_date` | `created_at` only for legacy rows without sale date | Yes | Daily dashboard/report values | Aggregation endpoints | SQLite/PostgreSQL daily summary queries | PARTIAL: backend contract PASS; cross-consumer display comparison not proven |
| Monthly aggregations | `sales.sale_date` | `created_at` only for legacy rows without sale date | Yes | Monthly dashboard/report values | Aggregation endpoints | SQLite/PostgreSQL monthly summary queries | PARTIAL: backend contract PASS; cross-consumer display comparison not proven |
| Customer Returns comparison | Sale's `sale_date`; return's official date is `returns.return_date` | Return `created_at` is creation metadata/fallback only | Yes | Returns/report comparisons | Returns endpoints | Sale and return dates are compared explicitly | PASS |

## Other entity contracts

These definitions are the required contract for the next verification pass. `UNVERIFIED` means the definition is recorded but the complete UI/API/DB path still needs a focused boundary test.

| Entity | Official Business Date | Timestamp | Store Timezone | UI | API | DB | Status |
|---|---|---|---|---|---|---|---|
| Purchase | `purchases.purchase_date` | `purchases.created_at` for creation time; legacy repository aliases it as `purchase_date` when the official column is absent | Yes | Purchase list/details | Purchase endpoints | Some legacy read paths use `created_at AS purchase_date` | PARTIAL |
| Customer Return | `returns.return_date` | `returns.created_at` for record creation; fallback only for legacy rows without return date | Yes | Returns list/details | Returns endpoints | `returns.return_date` | UNVERIFIED |
| Supplier Return | `returns.return_date` for supplier return records | `returns.created_at` for record creation | Yes | Supplier returns UI | Supplier return endpoints | `returns.return_date` where present | PARTIAL |
| Customer Payment | `payments.payment_date` | `payments.created_at` for insertion time; fallback only when payment date is absent | Yes | Payment/debt history | Payment endpoints | `payments.payment_date` | UNVERIFIED |
| Supplier Payment | `payments.payment_date` with supplier reference/type | `payments.created_at` for insertion time; supplier-ledger aging currently uses `created_at` | Yes | Supplier payment/history UI | Payment/supplier endpoints | `payments.payment_date`; supplier-ledger aging uses `created_at` | PARTIAL |
| Expense | `expenses.expense_date` | `expenses.created_at` for insertion time | Yes | Expense list/report | Expense endpoints | `expenses.expense_date` | UNVERIFIED |
| Debt Creation / Business Date | No separate `debt_date`; current creation day is derived from `debts.created_at` | `debts.created_at` is the creation timestamp, not `debts.due_date` | Yes | Debt creation/history UI | Debt endpoints | `debts.created_at` | PARTIAL |
| Debt Due Date | `debts.due_date` for aging/overdue rules only | `debts.created_at` remains creation time | Yes | Debt aging/overdue UI | Debt endpoints | `debts.due_date` | UNVERIFIED |
| Inventory Event | Event-specific field where available (`sold_at`, `event_date`, `purchase_date`) | `created_at` for event-record creation; legacy fallback only where explicitly documented | Yes | Inventory/history UI | Inventory endpoints | Event table fields plus legacy fallback | UNVERIFIED |
| Opening Stock | `inventory_movements.business_date` with `source_type=OPENING_STOCK` | `inventory_movements.created_at` records insertion time | Yes, date is supplied explicitly | Inventory page has dedicated `Opening Stock / الرصيد الافتتاحي` modal | `POST /inventory/opening-stock` plus frontend API client | Aggregate `inventory` and optional `inventory_items`; no Purchase/Supplier Ledger link | PASS |

## Lifecycle evidence

| Lifecycle | Executable evidence | Status |
|---|---|---|
| Purchase lifecycle | `go test ./internal/purchases -run TestPurchaseLifecycleSupplierBalanceAndReturnLedgerSQLite -count=1` proves `Product -> Purchase -> Receive -> Inventory -> Supplier Balance -> Supplier Payment -> Supplier Return`, including supplier-return credit and final balance reconciliation | PASS |
| Supplier Ledger payment lifecycle | `go test ./internal/suppliers -run TestSupplierLedgerPaymentLifecycleSQLite -count=1` proves Purchase debit, Partial, Full, Overpayment rejection, duplicate-reference rejection, Supplier Return credit, inventory return, final balance, and one ledger entry per accepted movement | PASS |
| Supplier Ledger adjustment reconciliation | `go test ./internal/suppliers -run TestSupplierLedgerAdjustmentReconcilesProjectionSQLite -count=1` proves a referenced supplier adjustment debit is recorded once and matches `suppliers.current_balance` | PASS |
| Customer Debt lifecycle | `go test ./internal/customers -run TestCustomerDebtLifecyclePaymentAndReturnCreditSQLite -count=1` proves Credit Sale -> Debt -> Partial Payment -> Remaining Balance -> Full Payment -> Debt 0 -> Customer Return -> Customer Credit, including overpayment, duplicate payment, duplicate adjustment, non-negative debt, ledger/projection parity, and no payment row for credit | PASS |
| Cash payment and change | `go test ./internal/sales -run TestCreateCashSaleStoresAppliedPaymentAndChangeSeparatelySQLite` proves a 125 sale receiving 150 stores `paid_amount=125`, `cash_received=150`, `change_amount=25`, payment amount 125, and profit based on 125 | PASS |
| Supplier Return lifecycle | `go test ./internal/supplierreturns -run TestSupplierReturnCreditsLedgerAndRemovesInventorySQLite` proves `RETURNED` inventory status, one supplier-return credit, and refund amount equal to cost; UI/API boundary remains unverified | PARTIAL |
| Supplier-less used inventory identity | `go test ./internal/inventory -run TestCreateUsedInventoryItemWithoutSupplierPreservesIdentitySQLite` proves `supplier_id IS NULL` while barcode, serial, condition, and product identity are preserved | PARTIAL |
| Opening Stock lifecycle | Backend SQLite/API tests plus `E2E_EMAIL=... E2E_RUN_OPENING_STOCK_LIFECYCLE=true npm run test:e2e -- opening-stock-lifecycle.spec.ts --project=chromium` prove Quantity and Individual/Used flows, explicit business date/source, no Purchase creation, nullable Supplier, aggregate quantity, item identity, and UI access | PASS |
| Supplier-less Opening Stock -> POS -> Customer Return | `TestSupplierlessOpeningStockSaleAndCustomerReturnPreserveIdentitySQLite` plus the passing `opening-stock-lifecycle.spec.ts` prove the Individual/Used item keeps the same item ID, barcode, serial, condition, and nullable Supplier through Inventory, POS, Sale, and Customer Return; screenshots, trace, and resolver/API network evidence were produced | PASS |

## Dashboard / Reports reconciliation

The following map is the authoritative review of the current consumer chain. `PASS` below means the DB/query/backend/API contract is reconciled by executable evidence; a displayed Card or Report remains `PARTIAL` until the live UI value is captured and compared.

| Card / Report value | Source of Truth | Repository / Service | API | Frontend hook / state | Display status |
|---|---|---|---|---|---|
| Dashboard today net sales | `sales.sale_date`, `sales.total_amount - tax`, completed sales, completed customer refunds | `dashboard.fetchTodayMetrics` | `/dashboard/stats` -> `todaySales` | `dashboardApi.getStats` -> `DashboardMetrics.todaySales` | PARTIAL: backend/report reconciliation PASS; rendered UI comparison not proven |
| Dashboard today profit | sales cost, approved expenses, completed refunds and returned cost | `dashboard.fetchTodayMetrics` | `/dashboard/stats` -> `todayProfit` | `dashboardApi.getStats` -> `DashboardMetrics.todayProfit` | PARTIAL |
| Dashboard outstanding debt | `debts.remaining_amount > 0` | `dashboard.CachedService.fetchFromDatabase` | `/dashboard/stats` -> `outstandingDebts` | `dashboardApi.getStats` -> `DashboardMetrics.outstandingDebts` | PARTIAL: reconciled with debt report backend; rendered UI not proven |
| Dashboard low stock | product minimum and available `inventory_items` | `dashboard.CachedService` | `/dashboard/stats` -> `lowStockCount` | `dashboardApi.getStats` -> `DashboardMetrics.lowStockCount` | PARTIAL |
| Dashboard supplier returns today | completed `supplier_returns.refund_amount` / return business date | `dashboard.fetchTodayMetrics` | `/dashboard/stats` -> `todaySupplierReturns` | `dashboardApi.getStats` -> `DashboardMetrics.todaySupplierReturns` | PARTIAL |
| Sales / Net Sales report | `sales.sale_date`, sale totals and return adjustments | `reports.Repository.GetSalesData` and report service | `/reports/sales`, `/reports/net-sales` | `useReports` -> `ReportsPage` / `ReportStats` | PARTIAL |
| Profit / Tax report | sales date, sale cost/tax, accounting expenses and returns | report repository/service | `/reports/profit`, `/reports/tax` | `useReports` -> `ReportsPage` | PARTIAL |
| Purchases report | purchases and supplier-return credits | `reports.GeneratePurchasesReport` | `/reports/purchases` | `useReports` -> `ReportsPage` | PARTIAL |
| Supplier report | purchases, supplier payments, supplier-return credits, including negative supplier credit | `reports.GenerateSuppliersReport` | `/reports/suppliers` | `useReports` -> `ReportsPage` / `ReportCharts` | PARTIAL: negative-credit clipping fixed and backend tested; UI not proven |
| Debt report | `debts.amount`, `paid_amount`, `remaining_amount`, due date | `reports.Repository.GetDebtsData` | `/reports/debts` | `useReports` -> `ReportsPage` | PARTIAL |
| Inventory report | `inventory` / `inventory_items` current quantities and cost | `reports.Repository.GetInventoryData` | `/reports/inventory` | `useReports` -> `ReportsPage` | PARTIAL |

**Reconciliation regressions executed**

- `TestDashboardAndReportsReconcileDailySalesAndDebtSQLite` compares a decimal/large daily sale and outstanding debt between Dashboard backend stats and Reports backend data.
- `TestSuppliersReportPreservesSupplierCreditSQLite` proves a negative supplier credit remains `-100.25` in the supplier report instead of being clipped to zero.
- The daily sales mismatch was fixed in SQLite by using net revenue after tax, matching Reports.
- Dashboard active-customer and low-stock counts no longer use cross-source `Math.max` fallbacks; they use the Dashboard stats source directly.
- Reports purchase tax differences no longer use `Math.max` to hide inconsistent API values.

**Current reconciliation status:** DB/Repository/Service/API contract `PASS` for the tested daily cases and monthly Net Sales/Tax/Gross Profit/Net Profit cases. Full Card/Report UI display parity, monthly Returns/Purchases/Supplier Balance/Customer Debt parity, complete zero/negative/large/decimal coverage, and Browser E2E remain `PARTIAL` or `NOT PROVEN`.

## Monthly reconciliation

| Monthly metric | Monthly aggregation source | Reports source | Evidence | Status |
|---|---|---|---|---|
| Net Sales | `monthly_sales_summary.total_revenue` from `sales.sale_date` after tax | `GetSalesData.TotalRevenue` for the same month | `TestMonthlyAggregationMatchesProfitReportAfterTaxAndReturnsSQLite` | PASS for DB/Repository/API contract |
| Tax | `sales.tax_amount` in the same monthly sale set | `GetTaxData` for the same month | Same monthly regression uses decimal tax `0.25` | PASS for Backend report contract |
| Gross Profit | `monthly_profit_summary.gross_profit` after return refund and returned cost | `GetProfitsData.GrossProfit` after the same adjustments | Monthly reconciliation regression | PASS for DB/Repository/API contract |
| Net Profit | `monthly_profit_summary.net_profit` after approved expense | `GetProfitsData.NetProfit` after the same adjustments | Monthly reconciliation regression with expense `7.25` | PASS for DB/Repository/API contract |
| Returns | Monthly profit adjustment uses completed return refund and returned cost | Returns report has monthly rows, but no direct cross-endpoint assertion yet | Included indirectly in monthly profit regression | PARTIAL |
| Purchases | No monthly Dashboard aggregation endpoint for supplier purchases | `GetPurchasesData` has monthly report rows | No same-source monthly Dashboard/Report fixture | NOT PROVEN |
| Supplier Balance | Current supplier ledger/projection, no monthly aggregation contract | Supplier report is current-state, not monthly | No monthly balance boundary contract | NOT PROVEN |
| Customer Debt | `monthly_debt_summary` uses opening/new/payment/overdue buckets | Debt report is current-state totals | Semantics are not the same monthly measure; no parity fixture | NOT PROVEN |

The monthly fix aligns monthly profit revenue with the established net-revenue contract by excluding sales tax before return and cost adjustments. Legacy report fixtures without a `returns` table continue to return zero return adjustments without changing current-schema behavior.

Monthly Backend result: `PASS` for Net Sales, Tax, Gross Profit, and Net Profit in the tested SQLite path. Monthly Returns remains `PARTIAL`; monthly Purchases, Supplier Balance, and Customer Debt remain `NOT PROVEN`. Frontend monthly display and Browser E2E remain `NOT PROVEN`.

## Exception classification

| Allowed `created_at` use | Classification | Rule |
|---|---|---|
| Event ordering | Allowed exception | May order Activity or history, but may not assign a financial/business day when an official date exists. |
| Actual creation timestamp | Allowed exception | May be displayed as the event timestamp and retained for audit. |
| Legacy fallback | Allowed exception | Must be conditional on the official date being absent and must convert the instant through Store Timezone. |
| Financial/business-day filter | PARTIAL unless official date is used | Any path that groups or filters by `created_at` despite an available official date is not PASS. |

## Final `created_at` scan

| Finding | Classification | Status |
|---|---|---|
| Sales Activity ordering and displayed event time | Ordering / Event Timestamp | PASS |
| Sales Chart and daily sales fallback when `sale_date` is absent | Legacy Fallback | PASS |
| Purchase repository aliases and filters `created_at AS purchase_date` in legacy paths | Business Date | PARTIAL |
| Supplier-ledger overdue/aging calculations based on `supplier_ledger.created_at` | Business Date | PARTIAL |
| Debt creation day based on `debts.created_at` because no dedicated `debt_date` exists | Event Timestamp currently used as business-day proxy | PARTIAL |
| Payment, expense, return, inventory, and report metadata fields | Event Timestamp / Ordering | PASS where no financial grouping uses the field; otherwise PARTIAL |

No `created_at` use is treated as an official Business Date without an explicit `PASS` or `PARTIAL` classification above.

## Phase 2 execution status

This section records the final review evidence for the live Phase 2 scope only: Purchases, Inventory, Supplier Ledger, and Customer Debt. The Phase 2 review is closed with residual `PARTIAL`/`NOT PROVEN` boundaries; no Phase 4 or Phase 5 work started.

| Phase 2 module | Required lifecycle | Status | Evidence |
|---|---|---|---|
| Purchases | Purchase -> Receive -> Inventory -> Supplier Balance -> Supplier Payment -> Supplier Return | PARTIAL | Backend lifecycle PASS; API/UI/E2E consumer proof remains open |
| Inventory | Quantity / Individual / Used / Opening Stock / Supplier-less / Movements / Negative Quantity / Barcode identity | PARTIAL | Backend and focused lifecycle evidence PASS; complete API/UI/report consumer proof remains open |
| Supplier Ledger | Purchase -> Payable -> Partial Payment -> Full Payment -> Supplier Return Credit -> Final Balance | PARTIAL | DB/Repository/Service lifecycle, adjustment reconciliation, API package, and frontend consumer checks pass; browser E2E and generic-credit coverage remain unproven |
| Customer Debt | Credit Sale -> Debt -> Partial Payment -> Full Payment -> Customer Return -> Debt Adjustment/Credit | PARTIAL | DB/Repository/Service lifecycle, API package, and frontend build pass; live API/UI/browser E2E and report/dashboard fixture remain open |
| Dashboard / Reports | Same source-of-truth values across cards and reports | PARTIAL | Contract exists; consumer-by-consumer proof is still partial |

**Executable evidence used for this phase**

```bash
cd "c:/Users/Administrator/Desktop/PartFlow/backend"; go test ./... -count=1
cd "c:/Users/Administrator/Desktop/PartFlow/frontend"; npm run test:run
cd "c:/Users/Administrator/Desktop/PartFlow/frontend"; npm run build:check
cd "c:/Users/Administrator/Desktop/PartFlow"; git diff --check
```

The current full verification returned success for all Go packages, all frontend tests, frontend typecheck/build, and diff check. No FAIL case is proven by this verification. Live API/UI/browser E2E and dashboard/report cross-fixtures remain `PARTIAL` or `NOT PROVEN` as recorded above.

## Sale boundary evidence

`TestSaleDayContractUsesSaleDateAcrossChartAndActivity` verifies four ₪880 sales around the Store Timezone midnight boundary: before midnight, 23:59, 00:00, and after midnight. Chart grouping and Activity API output use the same `sale_date` values. The focused backend suites for dashboard, reports, aggregations, and accounting, plus the frontend TypeScript check, pass.

The remaining required evidence is an end-to-end fixture that exposes the same ₪880 sale through Dashboard, Sales Report, Net Sales, and Profit responses in one test. Until that fixture is added, those rows retain the PASS status based on their verified SQL contracts, while the cross-endpoint proof remains an explicit follow-up gap.