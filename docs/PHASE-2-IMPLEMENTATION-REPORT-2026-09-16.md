# Phase 2 Implementation Report

**Status:** Phase 2 review closed with recorded residual PARTIAL/NOT PROVEN boundaries; no Phase 4 or Phase 5 work started  
**Effective date:** 2026-09-16  
**Scope:** Purchases, Inventory, Supplier Ledger, Customer Debt  
**Rule:** No PASS claim without executable evidence. No Phase 4 or Phase 5 work started.

## Baseline freeze: current review status

تم تثبيت وضعية المراجعة الحالية كـ baseline رسمي قبل أي تطوير إضافي. لا تُفتح أي Phase جديد، ولا تُعدل Barcode، ولا تُعاد Inspection/Testing، ولا يُسمح برفع أي حالة من `PARTIAL` أو `NOT PROVEN` إلى `PASS` بدون دليل قابل لإعادة التشغيل.

### Definitions used in this review

- **PASS** = مثبت باختبار قابل لإعادة التشغيل أو دليل حي يثبت الادعاء الكامل عبر طبقات الحقيقة المستهدفة.
- **PARTIAL** = بعض الطبقات مثبتة، لكن هناك طبقة أو أكثر مفتوحة أو غير مكتملة.
- **NOT PROVEN** = لا يوجد دليل كافٍ لإثبات الادعاء. لا يُسمح باختلاق Aggregation أو fallback لإخفاء الاختلاف.
- **Browser E2E** = لا يُعد جزءًا من PASS إلا إذا كان موجودًا فعليًا في متصفح حقيقي ومؤكد.

### Freeze status for the current review

| Area | Status | Notes |
|---|---|---|
| Net Sales / Tax / Gross Profit / Net Profit | PASS | Validated by monthly reconciliation regression in backend |
| Monthly Returns | PARTIAL | Included indirectly in profit logic, but not directly proven as a separate official monthly value |
| Monthly Purchases | NOT PROVEN | No direct DB → Query → Backend → API proof for the official monthly purchase value |
| Monthly Supplier Balance | NOT PROVEN | No direct balance-month proof across source and API |
| Monthly Customer Debt | NOT PROVEN | No direct debt-month proof across source and API |
| Dashboard/Reports UI parity | PARTIAL | Backend/API contract is tested; rendered UI parity remains unproven |
| Browser E2E | NOT PROVEN | No live browser proof for the monthly dashboard/report values |

**This review boundary is frozen as of 2026-09-16.** No Phase 4 or Phase 5 work starts from this baseline.

## 1. Change Impact

### Purchases

**Affected entities**
- `purchases`
- `purchase_items`
- `suppliers`
- `inventory`
- `inventory_items`
- `payments`
- `supplier_ledger` / supplier balance projection

**Affected APIs**
- Purchase list/detail/create/receive endpoints
- Supplier payment endpoints
- Supplier return endpoints
- Inventory receive endpoints

**Affected screens**
- Purchase screen
- Receive screen
- Supplier payment screen
- Supplier ledger screen
- Inventory screen

**Affected reports**
- Purchase totals
- Inventory movement report
- Supplier balance / payable report
- Dashboard financial cards depending on purchase and payable values

### Inventory

**Affected entities**
- `inventory`
- `inventory_items`
- `inventory_movements`
- `products`
- `purchases`
- `returns`
- `barcodes`

**Affected APIs**
- Inventory list/detail endpoints
- Opening stock endpoint
- Barcode resolution endpoints
- Inventory movement/history endpoints

**Affected screens**
- Inventory dashboard
- Opening stock modal
- POS and used-item flows
- Barcode lookup UI
- Returns UI

**Affected reports**
- Inventory stock report
- Store movement report
- Sales/profit report using inventory state

### Supplier Ledger

**Affected entities**
- `suppliers`
- `payments`
- `purchases`
- `supplier_returns`
- `supplier_ledger`

**Affected APIs**
- Supplier ledger endpoints
- Supplier payment endpoints
- Purchase and supplier-return APIs

**Affected screens**
- Supplier balance page
- Payment screen
- Supplier ledger screen

**Affected reports**
- Supplier payable report
- Payment report
- Dashboard supply/payable cards

### Customer Debt

**Affected entities**
- `sales`
- `payments`
- `debts`
- `customer_ledger`
- `returns`
- `customers`

**Affected APIs**
- Customer debt summary endpoints
- Customer payment endpoints
- Sale creation/payment endpoints
- Return and credit endpoints

**Affected screens**
- POS debt flow
- Customer payment screen
- Debt aging screen
- Returns and credit screen

**Affected reports**
- Debt aging
- Customer financial summary
- Dashboard outstanding debt and collection cards

## 2. Evidence status by module

### 2.1 Purchases

**Boundary:** DB -> Repository -> Service -> API -> Frontend -> Regression -> E2E  
**Status:** PASS for the explicit purchase lifecycle boundary; UI/API/E2E remains PARTIAL

**What is proven**
- Purchase business-date contract and the exact lifecycle proof are covered in [docs/MASTER-EVIDENCE-MATRIX.md](MASTER-EVIDENCE-MATRIX.md).
- `go test ./internal/purchases -run TestPurchaseLifecycleSupplierBalanceAndReturnLedgerSQLite -count=1` proves `Product -> Purchase -> Receive -> Inventory -> Supplier Balance -> Supplier Payment -> Supplier Return`, including supplier-return credit and final balance reconciliation.
- The focused regression command `go test ./internal/purchases ./internal/supplierreturns ./internal/suppliers -count=1` passes.

**What is still not proven**
- Live UI flow from Purchase to Supplier Balance to payment screens is not captured by a focused end-to-end fixture.
- Full API contract coverage for purchase receive + supplier payment + return credit across all consumers is still open in the live UI/API chain.

**Evidence summary**
- DB/Repository/Service: PASS by executable lifecycle proof and focused package regression
- API/UI: PARTIAL / NOT PROVEN
- E2E: NOT PROVEN

### 2.2 Inventory

**Boundary:** DB -> Repository -> Service -> API -> Frontend -> Regression -> E2E  
**Status:** PASS for the validated backend inventory contract, PARTIAL for full UI/API coverage

**What is proven**
- Inventory backend lifecycle tests pass, including quantity/individual/used and opening-stock behavior.
- The following scenarios are explicitly supported by project evidence:
  - Quantity
  - Individual
  - Used
  - Opening Stock
  - Supplier-less identity preservation
  - Movements
  - Negative quantity guardrails where applicable
  - Barcode identity continuity
- The explicit evidence matrix includes the opening-stock and supplier-less inventory identity tests.

**What is still not proven**
- Full UI, API contract, and front-end display continuity across the full inventory lifecycle remain partially covered.
- E2E chain from opening-stock to dashboard/report display is still not complete for every inventory branch.

**Evidence summary**
- DB/Repository/Service: PASS
- API/UI: PARTIAL
- E2E: NOT PROVEN (for the complete cross-entity flow)

### 2.3 Supplier Ledger

**Boundary:** DB -> Repository -> Service -> API -> Frontend -> Regression -> E2E  
**Status:** PARTIAL overall; PASS for DB/Repository/Service and focused API/Frontend contract checks, with browser E2E and generic credit adjustments still open

**What is proven**
- `TestSupplierLedgerPaymentLifecycleSQLite` covers Purchase debit, partial payment, full payment, overpayment rejection, duplicate payment rejection, Supplier Return credit, returned inventory, final balance, and one ledger entry per accepted movement.
- `TestSupplierLedgerAdjustmentReconcilesProjectionSQLite` covers a referenced adjustment debit and verifies `supplier_ledger.balance` equals `suppliers.current_balance` without double counting.
- `go test ./internal/api -count=1` passes for the API package.
- `npm run test:run -- src/features/suppliers/utils/supplier-normalization.test.ts` and `npm run build:check` pass for the frontend supplier consumers.
- The production fix makes payment and supplier-return ledger balances derive from signed movement totals rather than timestamp ordering.

**What is still not proven**
- The full chain `Purchase -> Payable -> Partial Payment -> Full Payment -> Supplier Return Credit -> Final Balance` is not captured as a single end-to-end proof from API to UI.
- Browser E2E through the live supplier screens and dashboard/report display parity remains unproven in one complete fixture.
- A generic credit adjustment API separate from Supplier Return is not present as an independently evidenced operation.

**Evidence summary**
- DB/Repository/Service: PASS
- API: PASS for package regression; route behavior remains PARTIAL without live API fixture
- Frontend: PASS for focused consumer test and build; live UI flow remains PARTIAL
- E2E: NOT PROVEN

### 2.4 Customer Debt

**Boundary:** DB -> Repository -> Service -> API -> Frontend -> Regression -> E2E  
**Status:** PARTIAL overall; PASS for DB/Repository/Service and focused API/Frontend contract checks

**What is proven**
- `TestCustomerDebtLifecyclePaymentAndReturnCreditSQLite` proves `Credit Sale -> Debt -> Partial Payment -> Remaining Balance -> Full Payment -> Debt 0 -> Customer Return -> Debt Adjustment / Customer Credit`.
- The lifecycle proves overpayment rejection, duplicate payment rejection, duplicate adjustment idempotency, `remaining_amount >= 0`, ledger/projection parity, and that Customer Credit does not create a payment row.
- Customer payment references are now checked by Customer Service/API before debt or balance mutation.
- Payment persistence no longer mutates debts a second time; debt application is owned by the service flow, preventing double deduction.
- `go test ./internal/customers ./internal/returns ./internal/sales ./internal/api -count=1` passes.
- Frontend `npm run build:check` passes for debt, customer, returns, dashboard, and reports consumers.

**What is still not proven**
- Live API route execution through an authenticated environment is not captured in a dedicated fixture.
- Browser E2E through debt payment and customer return screens is not proven.
- Dashboard and Debts Report reconciliation is covered by package/build checks but not by one cross-endpoint fixture.
- Generic customer credit adjustment outside `DEBT_ADJUSTMENT` returns remains `NOT PROVEN`.

**Evidence summary**
- DB/Repository/Service: PASS
- API: PASS for package regression; live route evidence PARTIAL
- Frontend: PASS for build/type contract; live UI evidence PARTIAL
- E2E: NOT PROVEN

## 3. Executable evidence collected in this phase

The Purchases lifecycle closure proof passed:

```bash
cd "c:/Users/Administrator/Desktop/PartFlow/backend"; go test ./internal/purchases -run TestPurchaseLifecycleSupplierBalanceAndReturnLedgerSQLite -count=1
cd "c:/Users/Administrator/Desktop/PartFlow/backend"; go test ./internal/purchases ./internal/supplierreturns ./internal/suppliers -count=1
cd "c:/Users/Administrator/Desktop/PartFlow/backend"; go test ./internal/suppliers -run 'TestSupplierLedger(PaymentLifecycleSQLite|AdjustmentReconcilesProjectionSQLite)$' -count=1
cd "c:/Users/Administrator/Desktop/PartFlow/backend"; go test ./internal/customers -run TestCustomerDebtLifecyclePaymentAndReturnCreditSQLite -count=1
cd "c:/Users/Administrator/Desktop/PartFlow/backend"; go test ./internal/returns -run 'TestSQLiteDebtAdjustmentCapsDebtAndCreatesCustomerCredit|TestServiceCreateReturnLinksActiveDebtForDebtAdjustment' -count=1
cd "c:/Users/Administrator/Desktop/PartFlow/backend"; go test ./internal/api -count=1
cd "c:/Users/Administrator/Desktop/PartFlow/backend"; go test ./internal/dashboard -run TestDashboardAndReportsReconcileDailySalesAndDebtSQLite -count=1
cd "c:/Users/Administrator/Desktop/PartFlow/backend"; go test ./internal/reports -run TestSuppliersReportPreservesSupplierCreditSQLite -count=1
cd "c:/Users/Administrator/Desktop/PartFlow/backend"; go test ./internal/api -run TestMonthlyAggregationMatchesProfitReportAfterTaxAndReturnsSQLite -count=1
cd "c:/Users/Administrator/Desktop/PartFlow/frontend"; npm run test:run -- src/features/suppliers/utils/supplier-normalization.test.ts
cd "c:/Users/Administrator/Desktop/PartFlow/frontend"; npm run build:check
```

The lifecycle test verifies the purchase date and supplier association, purchase totals and item quantity, receiving into inventory, partial supplier payment, supplier balance reduction, supplier return inventory removal, return refund amount, supplier-return ledger credit, and final supplier balance.

The Supplier Ledger tests additionally verify partial/full supplier payment, overpayment and duplicate rejection, supplier-return credit, adjustment debit, single movement recording, and equality between the signed ledger balance and `suppliers.current_balance`.

The Customer Debt lifecycle test additionally verifies credit-sale debt creation, partial/full payment, overpayment and duplicate payment rejection, debt adjustment idempotency, customer credit separation from payments, non-negative debt, and equality between the signed `customer_ledger` balance and `customers.current_balance`.

Dashboard/Reports reconciliation additionally proves that daily Dashboard net sales matches the Sales Report after tax, Dashboard outstanding debt matches the Debt Report, and supplier credit remains negative in the Supplier Report. The reconciliation pass also removed frontend `Math.max` fallbacks that could conceal source mismatches.

Monthly reconciliation proves monthly Net Sales, Tax, Gross Profit, and Net Profit parity between aggregation summaries and Reports after decimal tax, completed returns, returned cost, and approved expense adjustments. Monthly Purchases, Supplier Balance, Customer Debt, direct monthly Returns parity, and browser-rendered monthly values remain unproven.

The current full repository verification also returned success:

```bash
cd "c:/Users/Administrator/Desktop/PartFlow/backend"; go test ./... -count=1
cd "c:/Users/Administrator/Desktop/PartFlow/frontend"; npm run test:run
cd "c:/Users/Administrator/Desktop/PartFlow/frontend"; npm run build:check
cd "c:/Users/Administrator/Desktop/PartFlow"; git diff --check
```

Result: Backend all packages PASS; Frontend 15 test files / 48 tests PASS; typecheck and production build PASS; diff check PASS. These results do not upgrade live API/UI/E2E or dashboard/report cross-fixture statuses.

## 4. Phase 2 matrix

| Module | Requirement | Status | Evidence |
|---|---|---|---|
| Purchases | Purchase -> Receive -> Inventory -> Supplier Balance -> Supplier Payment -> Supplier Return | PARTIAL | Backend purchase suite passes; API/UI/E2E gap remains |
| Inventory | Quantity / Individual / Used / Opening Stock / Supplier-less / Movements / Negative Quantity / Barcode identity | PARTIAL overall | Backend contract and focused E2E branches pass; complete API/UI/report coverage remains open |
| Supplier Ledger | Purchase -> Payable -> Partial Payment -> Full Payment -> Supplier Return Credit -> Final Balance | PARTIAL | Backend lifecycle, adjustment, API package, focused frontend test, and build pass; live browser E2E and generic credit adjustment remain open |
| Customer Debt | Credit Sale -> Debt -> Partial Payment -> Full Payment -> Customer Return -> Debt Adjustment/Credit | PARTIAL | Backend lifecycle, API package, and frontend build pass; live API/UI/browser E2E and dashboard/report cross-fixture remain open |
| Dashboard / Reports | Same source-of-truth data without different formulas | PARTIAL | Daily reconciliation and monthly Net Sales/Tax/Gross Profit/Net Profit backend checks pass; monthly Returns/Purchases/Supplier Balance/Customer Debt and full Card/Report UI remain open |

## 5. Remaining gaps

**PARTIAL / NOT PROVEN**
- Full API contract coverage for purchase and supplier ledger consumer chain
- Complete Frontend parity for purchase/inventory/debt screens
- Dashboard and reports reconciliation across all phase-2 entities, including monthly and full Store Timezone boundaries
- Single end-to-end proof for all required business flows

**FAIL**
- None identified in the current backend execution sweep for the targeted Phase 2 packages.
- No current repository-level FAIL is shown in this execution window.

## 6. Execution gate for this phase

The Phase 2 review remains closed; Dashboard/Reports reconciliation is recorded as a separate follow-up boundary and is not a reason to start Phase 4 or Phase 5.

**Current status:** PARTIAL, not PASS.

The Dashboard/Reports boundary will be eligible for PASS only when the following are proven in a single reproducible flow:
- DB -> Repository -> Service -> API -> Frontend -> E2E
- same-source financial calculations everywhere
- dashboard/report values reconciled to the same business record

No Phase 4 or Phase 5 work was started.
