# PartFlow Architecture Baseline

**Status:** Official baseline  
**Effective date:** 2026-09-16  
**Authority:** This document defines the current ownership and source-of-truth contract. It is read together with [ENGINEERING-RULES.md](ENGINEERING-RULES.md), [ARCHITECTURE-PRINCIPLES.md](ARCHITECTURE-PRINCIPLES.md), and [docs/MASTER-EVIDENCE-MATRIX.md](docs/MASTER-EVIDENCE-MATRIX.md).

## 1. Purpose

This baseline prevents a local fix from creating a second definition of the same entity, amount, balance, or date. For every entity it defines:

```text
Entity -> Source of Truth -> Current State -> Immutable History
       -> Business Date -> Owner Service -> API -> Main Consumers
```

`CANONICAL` means the value is authoritative. `PROJECTION` means a derived or denormalized value that must be rebuildable from the canonical source. `LEGACY` means compatibility data that must not become a new business source. `NOT PROVEN` means the contract is defined here but the full DB -> Backend -> API -> UI evidence is not yet complete.

## 2. Global Source-of-Truth Rules

1. A business fact has one canonical source.
2. A current-state projection may exist for speed, but it is not an independent authority.
3. An immutable event may be stored separately from current state; it records what happened, while current state records what is true now.
4. A cached or aggregated value must carry a rebuild/reconciliation path and must never silently replace the transaction source.
5. `created_at` is an event timestamp. It is not a Business Date when an official date exists.
6. Frontend, dashboard, reports, and aggregations consume the same Service/API contract. They do not create alternative financial formulas.
7. Existing duplicate columns are classified as canonical, projection, or legacy below. No new duplicate field may be added without a Change Impact record.

## 3. Entity Baseline

### 3.1 Product

| Field | Contract |
|---|---|
| Source of Truth | `products` (`id`, `sku`, `name`, prices, tracking flags, active/archive state) |
| Current State | The active product definition and catalog attributes in `products` |
| Immutable History | Product changes through audit events; sales/purchases/inventory retain their own transaction snapshots and references |
| Business Date | No financial Business Date; catalog changes use `created_at`/`updated_at` as timestamps |
| Owner Service | `internal/products` Products Service |
| API | `/api/v1/products`, `/api/v1/categories`, `/api/v1/brands` |
| Main Consumers | Inventory, Purchases, POS/Sales, Barcode Resolver, Reports, Categories UI |

Prices on `products` are catalog defaults. A completed transaction must use its own sale/purchase line price and cost snapshot, not silently reread the current product price.

### 3.2 Category

| Field | Contract |
|---|---|
| Source of Truth | `categories` |
| Current State | Name, hierarchy, active state, icon/color, and description |
| Immutable History | Audit events and references from products/inventory; category deletion must preserve references or archive the category |
| Business Date | None; `created_at`/`updated_at` are timestamps |
| Owner Service | Products Service / category operations |
| API | `/api/v1/categories` |
| Main Consumers | Product catalog, Inventory, Reports, Categories UI |

### 3.3 Inventory

| Field | Contract |
|---|---|
| Source of Truth | `inventory` for aggregate/quantity products only |
| Current State | `quantity`, `reserved_quantity`, product reference, location, and aggregate updated timestamp |
| Immutable History | `inventory_movements` for every quantity-changing event |
| Business Date | Event-specific; opening stock uses `inventory_movements.business_date`, purchase uses `purchase_date`, sale uses `sale_date` |
| Owner Service | Inventory Service; Purchase, Sales, and Returns may request changes only through their service boundaries |
| API | `/api/v1/inventory`, `/api/v1/inventory/products/:id/quantity`, inventory item endpoints |
| Main Consumers | Inventory UI, POS, Purchases/Receive, Returns, Dashboard, Reports |

`inventory` is not a second source for individual serialized pieces. For an individual item, `inventory_items` is canonical and aggregate `inventory` is updated only when the product is also tracked as quantity stock.

### 3.4 Inventory Item

| Field | Contract |
|---|---|
| Source of Truth | `inventory_items` for each physical/individually tracked piece |
| Current State | `status`, product, barcode, serial, condition, grade, purchase cost, selling price, supplier, location, and sale state |
| Immutable History | `inventory_movements` plus audit/business references; item identity is never recreated for sale or return |
| Business Date | Source event date (`purchase_date`, opening `business_date`, or return/sale event date); `sold_at` records the sale event timestamp/date where available |
| Owner Service | Inventory Service |
| API | `/api/v1/inventory/items`, `/api/v1/inventory/items/:id`, `/api/v1/barcodes/resolve/:code` |
| Main Consumers | Receive, Inventory, Used Items, POS, Customer Returns, Supplier Returns, Reports |

`inventory_items` is canonical for individual identity. A sale changes the existing item to `SOLD`; a customer return changes that same item according to return resolution. Neither flow creates a replacement item.

### 3.5 Inventory Movement

| Field | Contract |
|---|---|
| Source of Truth | `inventory_movements` |
| Current State | Not applicable; current state is maintained by Inventory/Inventory Item |
| Immutable History | One business event per movement with `movement_type`, `reason`, `reference_type`, `reference_id`, `before_quantity`, `quantity`, `after_quantity`, `created_by`, and `created_at` |
| Business Date | `business_date` when the event supplies one; otherwise the owning transaction date; `created_at` remains the event timestamp |
| Owner Service | Inventory Service; source Services request movements inside their transaction boundary |
| API | Inventory history endpoints and source transaction responses |
| Main Consumers | Inventory history, Smart Delete/Reverse, Audit, Reports, reconciliation jobs |

Every movement must answer: why, which reference, who, when, before, change, and after. A movement is never deleted to hide a correction; reversal creates a compensating movement.

### 3.6 Opening Stock

| Field | Contract |
|---|---|
| Source of Truth | `inventory` or `inventory_items` current state plus an `inventory_movements` event with `source_type=OPENING_STOCK` |
| Current State | Quantity stock or an individual/used item, with optional supplier and identity attributes |
| Immutable History | The opening-stock movement and audit record |
| Business Date | Required `inventory_movements.business_date` |
| Owner Service | Inventory Service / Opening Stock operation |
| API | `POST /api/v1/inventory/opening-stock` |
| Main Consumers | Inventory UI, POS, Barcode Resolver, Reports, inventory history |

Opening Stock is not a fake purchase. It creates no Purchase, Supplier Ledger Entry, or Supplier Payment unless a future explicit commercial workflow requires it.

### 3.7 Purchase

| Field | Contract |
|---|---|
| Source of Truth | `purchases` |
| Current State | Supplier, invoice number, totals, tax, paid/remaining amounts, status, and notes |
| Immutable History | Purchase status transitions, receive/reversal movements, supplier ledger references, and audit events |
| Business Date | `purchases.purchase_date` |
| Owner Service | Purchases Service |
| API | `/api/v1/purchases`, receive/cancel/reverse/payment subroutes |
| Main Consumers | Purchases UI, Inventory Receive, Supplier Ledger, Reports, Dashboard |

`created_at` is creation time. A legacy fallback to `created_at` is allowed only when the old schema has no `purchase_date`, and must be visible in the repository compatibility path.

### 3.8 Purchase Item

| Field | Contract |
|---|---|
| Source of Truth | `purchase_items` |
| Current State | Product, quantity, unit cost, line total, barcode/serial inputs, and purchase reference |
| Immutable History | Receive movement, supplier return item, and purchase audit references |
| Business Date | Inherits the owning Purchase `purchase_date`; receive is a separate event timestamp/date |
| Owner Service | Purchases Service |
| API | Purchase detail/items endpoints and receive endpoint |
| Main Consumers | Purchase UI, Receive, Inventory Items, Supplier Returns, Supplier Ledger, Reports |

Receiving creates or updates inventory identity from the purchase item. It must not create a second unrelated item for the same physical barcode/serial.

### 3.9 Sale

| Field | Contract |
|---|---|
| Source of Truth | `sales` transaction row |
| Current State | Sale status, totals, tax, discount, payment status, paid amount, remaining amount, cash received, change, and customer reference |
| Immutable History | `sales` row plus sale items, payment records, inventory movements, ledger/debt references, and audit events; confirmed sales are reversed, not deleted |
| Business Date | `sales.sale_date` |
| Owner Service | Sales Service |
| API | `/api/v1/sales`, payment/cancel/held/summary/profit subroutes |
| Main Consumers | POS, Invoice, Customer/Debt, Inventory, Payments, Returns, Dashboard, Reports |

### 3.10 Sale Item

| Field | Contract |
|---|---|
| Source of Truth | `sale_items` |
| Current State | Product, `inventory_item_id` where applicable, quantity, unit price, unit cost, tax, discount, and line total |
| Immutable History | Sale item row and linked inventory movement/return references |
| Business Date | Inherits Sale `sale_date` |
| Owner Service | Sales Service |
| API | Sale detail and sale creation contracts |
| Main Consumers | POS, COGS/Profit, Inventory, Returns, Reports, Barcode Resolver |

For serialized items, `inventory_item_id` is the identity link. For quantity products, quantity and product reference are authoritative.

### 3.11 Customer

| Field | Contract |
|---|---|
| Source of Truth | `customers` for profile, credit limit, and active state |
| Current State | Name, contact data, `credit_limit`, and a balance field only as a rebuildable projection |
| Immutable History | `customer_ledger`, sales, payments, debts, returns, and audit events |
| Business Date | No single transaction date; consuming entities provide the financial date |
| Owner Service | Customers Service |
| API | `/api/v1/customers` and customer ledger/payment/debt routes |
| Main Consumers | POS, Debts, Customer History, Payments, Returns, Dashboard, Reports |

`customers.current_balance` is a projection/cache, not an independent authority. It must reconcile to the customer debt/ledger contract.

### 3.12 Customer Debt

| Field | Contract |
|---|---|
| Source of Truth | `debts.remaining_amount` for outstanding debt by debt record; `debts` is the canonical debt state |
| Current State | `amount`, `paid_amount`, `remaining_amount`, `status`, customer, sale, and due date |
| Immutable History | Credit Sale, payment allocations, return adjustments, and `customer_ledger` entries |
| Business Date | Debt creation inherits Sale `sale_date`; `due_date` is the official aging/due date |
| Owner Service | Debts/Customers Service, with Sales and Returns invoking explicit debt operations |
| API | `/api/v1/debts`, customer debt summary, sale payment, customer payment routes |
| Main Consumers | POS, Customers, Debts UI, Dashboard, Reports, Returns |

`customer_ledger` is the statement/history projection. It does not replace `debts.remaining_amount` for the open-debt number. Customer credit is not a payment: a return may reduce debt or create customer credit according to the return policy.

### 3.13 Customer Payment

| Field | Contract |
|---|---|
| Source of Truth | `payments` payment event with customer reference/type |
| Current State | Amount, method, status, reference, customer, and payment date |
| Immutable History | Payment row, allocation to `debts`, `customer_ledger` credit, and audit event |
| Business Date | `payments.payment_date` |
| Owner Service | Payments Service, with Customers/Debts Service applying allocations |
| API | `/api/v1/payments` and `/api/v1/customers/:id/payments` |
| Main Consumers | Customers, Debts, Sales, Cash Flow, Reports, Receipts |

Duplicate references, overpayment, and double allocation must be rejected or explicitly represented as a separate approved credit policy. Negative outstanding debt is invalid.

### 3.14 Supplier

| Field | Contract |
|---|---|
| Source of Truth | `suppliers` for profile, credit limit, and active state |
| Current State | Supplier identity plus `current_balance` as a rebuildable projection |
| Immutable History | `supplier_ledger`, purchases, supplier payments, supplier returns, and audit events |
| Business Date | No single transaction date; consuming entities provide the financial date |
| Owner Service | Suppliers/Supplier Ledger Service |
| API | `/api/v1/suppliers` and supplier ledger/payment routes |
| Main Consumers | Purchases, Inventory, Supplier Returns, Payments, Reports, Supplier UI |

`suppliers.current_balance` is not allowed to diverge from the canonical supplier balance calculation.

### 3.15 Supplier Payment

| Field | Contract |
|---|---|
| Source of Truth | `payments` supplier payment event; its effect on payable balance is represented in `supplier_ledger` |
| Current State | Payment amount, method, reference, supplier, status, and payment date |
| Immutable History | Payment event plus one supplier ledger credit and audit event |
| Business Date | `payments.payment_date` |
| Owner Service | Payments/Suppliers Service |
| API | `/api/v1/payments`, `/api/v1/suppliers/:id/payments`, purchase payment route |
| Main Consumers | Supplier UI, Purchases, Supplier Ledger, Cash Flow, Reports |

The payment event and ledger entry are two representations of one operation. `payments` is canonical for the payment fact; `supplier_ledger` is canonical for the supplier balance/history projection. They must share a stable reference/idempotency key and must never both be counted as two payments.

### 3.16 Supplier Return

| Field | Contract |
|---|---|
| Source of Truth | `supplier_returns` and `supplier_return_items` |
| Current State | Return status, purchase/supplier references, refund/credit amount, and returned quantity |
| Immutable History | Supplier return rows, inventory reversal/movement, supplier ledger credit, and audit |
| Business Date | Supplier return date field when present; otherwise explicit return event date, never silently purchase `created_at` |
| Owner Service | Supplier Returns Service |
| API | `/api/v1/supplier-returns` |
| Main Consumers | Purchases, Inventory, Supplier Ledger, Supplier UI, Reports |

### 3.17 Customer Return

| Field | Contract |
|---|---|
| Source of Truth | `returns` and `return_items` |
| Current State | Status, refund status/method, sale/customer reference, total refund, resolution, and returned condition |
| Immutable History | Return rows, inventory movement/status change, debt/customer ledger adjustment, refund/payment reference, and audit |
| Business Date | `returns.return_date` |
| Owner Service | Returns Service |
| API | `/api/v1/returns` |
| Main Consumers | POS/Sales, Inventory, Customers/Debts, Payments, Dashboard, Reports |

Customer Return is not a payment. It may refund cash, reduce an open debt, or create customer credit according to the selected policy. The same sale item and inventory item identity must be preserved.

### 3.18 Expense

| Field | Contract |
|---|---|
| Source of Truth | `expenses` approved expense row |
| Current State | Amount, category, status, recurrence, approval, and notes |
| Immutable History | Expense row status transitions, approval/rejection audit, and payment/cash event when present |
| Business Date | `expenses.expense_date` |
| Owner Service | Expenses Service; recurring creation is invoked by Worker |
| API | `/api/v1/expenses` |
| Main Consumers | Expenses UI, Net Profit, Cash Flow, Dashboard, Reports, Worker |

Only approved expenses enter Net Profit. A recurring definition is not itself a realized expense until the Worker creates the dated expense record.

### 3.19 Barcode

| Field | Contract |
|---|---|
| Source of Truth | The identity-bearing `products.barcode` or `inventory_items.barcode`, with `barcodes` as managed/generated registry where applicable |
| Current State | Code, active state, product/item link, serial context, and lifecycle status |
| Immutable History | Purchase, receive, sale, return, supplier return, and inventory movement references; barcode changes require audit |
| Business Date | Inherits the owning product/item event; resolver itself has no Business Date |
| Owner Service | Barcode Service and Central Barcode Resolver |
| API | `/api/v1/barcodes/resolve/:code` and barcode management routes |
| Main Consumers | Products, Purchases, Receive, Inventory, Used Items, POS, Customer Returns, Supplier Returns, Reports |

No screen may implement a competing barcode lookup. Duplicate code, sold-item reuse, and unrelated fallback matches are rejected by the central resolver contract.

## 4. Official Financial Values

### 4.1 Customer Outstanding Debt

**Canonical number:** sum of `debts.remaining_amount` for active open debt statuses (`pending`, `partial`, `overdue`) for the customer.

```text
Customer Outstanding Debt = SUM(open debts.remaining_amount)
```

- `customer_ledger` is the statement/history and reconciliation evidence.
- `payments` is the payment event source and reduces debt through an allocation transaction.
- `customers.current_balance` is a projection and must not be used as an independent calculation.
- Returns with `DEBT_ADJUSTMENT` reduce debt through the Returns/Debt Service, not by inserting a fake payment.

**Current status:** `PARTIAL` until all customer screens and reports consume the same service contract.

### 4.2 Supplier Balance

**Canonical number:** the latest rebuildable balance in `supplier_ledger`, with debit purchases and credit payments/returns.

```text
Supplier Balance = Purchase Debits - Supplier Payments - Supplier Return Credits
```

- `payments` is canonical for the payment fact.
- `supplier_ledger` is canonical for supplier payable balance/history.
- `suppliers.current_balance` is a projection/cache and must reconcile to the ledger.
- `ledger_entries` is a general ledger surface and is not a second supplier-balance authority unless a future migration explicitly promotes it.

**Current status:** `PARTIAL`; repository queries still contain legacy/current-balance paths that require reconciliation before PASS.

### 4.3 Inventory Quantity

Two explicit, non-conflicting modes exist:

```text
Quantity-tracked product: inventory.quantity - inventory.reserved_quantity
Individual item: COUNT(inventory_items with current eligible status)
```

- `inventory` is canonical for aggregate quantity products.
- `inventory_items` is canonical for individually tracked pieces.
- `inventory_movements` is immutable history and reconciliation source.
- Never add a `CurrentQuantity` field that duplicates either table.

**Current status:** `PARTIAL`; the contract exists, but every dashboard/report path must be audited for consistent mode selection.

### 4.4 COGS

**Canonical formula for a completed sale:**

```text
COGS = SUM(sale_items.quantity * sale_items.unit_cost)
```

For individual items, `unit_cost` is the cost snapshot taken from the inventory item at sale time. `products.cost_price` is only a catalog/default fallback for legacy rows and must not replace a transaction snapshot when one exists.

`sales.cost_amount` is a transaction snapshot/cache that must equal the line-level formula and is not an independent source.

**Current status:** `PARTIAL`; dashboard and reports contain fallback SQL for legacy schemas and require cross-endpoint reconciliation.

### 4.5 Net Sales

**Canonical formula:**

```text
Net Sales = Completed Sale Gross Amount - Tax - Customer Return Refund Net of Tax
```

The exact refund tax treatment follows the established tax-adjusted return rule. Gross sale, tax, and return values must use their official Business Dates.

**Current status:** `PARTIAL`; SQL evidence exists, but one end-to-end response must prove the same value in Dashboard, Reports, and Net Sales consumers.

### 4.6 Gross Profit

```text
Gross Profit = Net Sales Before Operating Expenses - COGS
```

`gross_profit` on `sales` is a transaction snapshot. Period reports use the central Sales/Accounting calculation, not a separate UI formula.

**Current status:** `PARTIAL`.

### 4.7 Net Profit

```text
Net Profit = Gross Profit - Approved Expenses
```

Customer return adjustments and tax treatment must occur before the period gross profit is finalized. Only approved realized expenses enter the formula.

**Current status:** `PARTIAL`; a shared cross-endpoint reconciliation test is required.

### 4.8 Cash Flow

Cash Flow has no single fully proven contract in the current repository. The target contract is:

```text
Cash In = Customer Payments + Cash Sales + Other Approved Cash Receipts
Cash Out = Supplier Payments + Approved Cash Expenses + Cash Refunds
Net Cash Flow = Cash In - Cash Out
```

`payments.payment_date` and `expenses.expense_date` are the relevant official dates. Credit sales are revenue but not cash until payment.

**Current status:** `NOT PROVEN`; do not label a dashboard number PASS until one service owns this formula.

### 4.9 Tax

Tax is sourced from transaction tax fields (`sales.tax_amount`, purchase tax fields, and line tax where present) as calculated and persisted by the owning Service. Store `tax_rate` is configuration, not historical tax for an already completed transaction.

```text
Tax Report = SUM(transaction tax_amount) - applicable tax-adjusted return tax
```

**Current status:** `PARTIAL`; historical rows with different schemas require the documented legacy fallback only.

### 4.10 Customer Credit

```text
Credit Limit = customers.credit_limit
Available Credit = max(0, Credit Limit - Customer Outstanding Debt)
```

Customer credit is not a payment and must not be added to cash inflow. A return may create a credit adjustment according to the return policy.

**Current status:** `PARTIAL`.

### 4.11 Supplier Credit

Supplier credit means the payable amount owed by the store:

```text
Supplier Credit = Supplier Balance
```

It is reduced by a supplier payment or supplier return credit, never by merely changing a product or inventory row.

**Current status:** `PARTIAL`.

## 5. Current State and Immutable History

The system must keep these layers separate without duplicating the same fact:

| Layer | Purpose | Examples |
|---|---|---|
| Current state | Fast operational decisions | `products`, `inventory`, `inventory_items`, active `debts`, `suppliers.current_balance` projection |
| Immutable business history | What actually happened | sales, purchases, payments, returns, expenses, ledger entries, inventory movements, audit events |
| Projection/cache | Rebuildable speed optimization | `sales.cost_amount`, `customers.current_balance`, `suppliers.current_balance`, summary tables |

Before adding a field, answer:

1. Does the value already exist?
2. Which table/service owns it?
3. Which operation updates it?
4. Can it disagree with another field?
5. What reconciliation rebuilds it?

If those answers are not documented, the field must not be added.

## 6. Inventory Lifecycle Contract

```text
Opening Stock / Purchase Receive
        -> Inventory or Inventory Item
        -> Sale
        -> Customer Return or Supplier Return
```

Rules:

- Purchase Receive creates the item once and records a purchase movement/reference.
- Opening Stock creates a stock/item record and an `OPENING_STOCK` movement, not a Purchase.
- Sale changes quantity or item status and records a sale movement.
- Customer Return preserves the original sale item and inventory item identity.
- Supplier Return references the original purchase item and supplier.
- Every movement records reason, reference, before, delta, after, date, and user.
- Smart Delete/Reversal creates compensating history; it does not erase confirmed business history.

## 7. Barcode Contract

The Central Barcode Resolver is the only operational lookup authority:

```text
Products -> Purchases -> Receive -> Inventory -> Used Items
         -> POS -> Customer Returns -> Supplier Returns -> Reports
```

The resolver must return the same product/item identity for the same code and reject:

- duplicate active barcode;
- reuse of a `SOLD` individual item;
- unrelated fallback sale/return matches;
- creation of a new item merely because a lookup occurred.

No Barcode implementation changes are part of this baseline; this section freezes the current contract.

## 8. Database Contract Baseline

The PostgreSQL and SQLite contracts must converge on the following minimum shape:

| Contract area | Required authority |
|---|---|
| Product catalog | `products`, `categories`, `brands` |
| Aggregate stock | `inventory` |
| Individual stock | `inventory_items` |
| Stock history | `inventory_movements` |
| Purchases | `purchases`, `purchase_items` |
| Sales | `sales`, `sale_items` |
| Customer debt | `debts`, `customer_ledger` history |
| Supplier payable | `supplier_ledger` |
| Payment events | `payments` |
| Returns | `returns`, `return_items`, supplier return tables |
| Expenses | `expenses`, expense categories |
| Sync | `sync_queue`, `sync_conflicts`, snapshot contract |

Every migration must specify tables, columns, types, nullability, defaults, foreign keys, unique constraints, indexes, and both-engine behavior. Required tests are:

1. Fresh PostgreSQL bootstrap.
2. Fresh SQLite bootstrap.
3. Legacy SQLite upgrade.
4. PostgreSQL migration upgrade.
5. Snapshot/import into a fresh SQLite database.
6. Reconciliation of projections after import.

Compatibility code is valid only when the old schema is named, detected, and tested. It must never turn a missing official field into a silently incorrect value.

## 9. API and Frontend Contract

Every endpoint must define Request DTO, Response DTO, validation, stable error codes, status transitions, and idempotency behavior. A field change requires a consumer search across:

```text
Frontend -> Dashboard -> Reports -> Worker -> Sync -> Unit/Service/API/E2E tests
```

The frontend contract is:

```text
Fetch -> Parse -> Display
```

Allowed UI-only calculations include formatting, sorting, labels, and visible cash change when the backend has not yet persisted it. Financial totals, debt, COGS, profit, tax, and balance must come from the API contract.

## 10. Smart Delete and Audit

| State | User action | Backend behavior |
|---|---|---|
| Draft | Delete | Physical delete is allowed if no business history exists |
| Confirmed/Received | Delete | Reverse if safe, otherwise block with an understandable message |
| Confirmed Sale | Delete | Reverse only; preserve sale and inventory history |
| Payment | Delete | Cancel/reverse with audit; no silent delete |
| Return | Delete | Reverse/cancel according to return state; preserve history |
| Inventory Movement | Delete | Never directly delete; create compensating movement |

Audit business events include purchase creation/receive/reverse, sale, payment, return, debt change, inventory adjustment, and financial data change. UI debug logging remains separate.

## 11. Worker and Aggregation Rules

Every Worker must document its job, schedule, last successful run, last failure, logging, and duplicate-execution protection. Until heartbeat/metrics exist, Worker health is `PARTIAL`.

Aggregations are projections, not sources of truth:

```text
Source Transaction -> Aggregation Update -> Dashboard/Report
                         ^
                         |
                 Reconciliation Job
```

No new large aggregation should be added until its required dashboard data, update trigger, rebuild path, and reconciliation query are defined. Archive is future-ready only; it is not implemented by deleting old data.

## 12. Required Change Impact

Every implementation change must start with:

```text
What changes:
Which entities:
Which source-of-truth values:
Which services/repositories:
Which APIs/DTOs:
Which screens/components:
Which reports/aggregations:
Which worker/sync paths:
Which migrations/contracts:
Which tests and E2E:
How old data is preserved:
Expected status: PASS / PARTIAL / FAIL / NOT PROVEN
```

## 13. PASS Gate

`PASS` requires reproducible evidence for the claimed scope. The required gate is:

```text
go test ./...
go vet ./...                  # Backend change
npm run test:run              # Frontend change
npm run build:check           # Typecheck + build
npm run lint                  # Frontend change
git diff --check
Focused DB/API/service tests
Focused Browser E2E            # When DB -> Backend -> API -> UI is claimed
```

`PARTIAL` means some layers pass. `FAIL` means an executable test contradicts the contract. `NOT PROVEN` means evidence is insufficient. Static code reading never upgrades a status to `PASS`.

## 14. Baseline Gaps to Close

The following are intentionally recorded rather than hidden:

| Gap | Status | Required closure |
|---|---|---|
| Customer debt cross-screen source reconciliation | PARTIAL | One Customer/Debt Service response consumed by dashboard, customer, debt, and reports |
| Supplier ledger vs `current_balance` reconciliation | PARTIAL | Rebuild/query comparison and consumer migration |
| COGS/profit cross-endpoint reconciliation | PARTIAL | Same sale fixture asserted in sale, dashboard, and reports |
| Cash Flow formula and owner service | NOT PROVEN | Implement one service contract and tests |
| Fresh PostgreSQL/SQLite/legacy bootstrap matrix | PARTIAL | Add reproducible migration fixtures |
| Worker heartbeat and duplicate protection evidence | PARTIAL | Add run status/metrics and test |
| Full Purchase -> Receive -> Inventory -> POS -> Payment/Debt -> Return UI E2E | PARTIAL | Live DB/API/UI test with evidence artifacts |

These gaps are work items for later phases; they are not permission to create parallel calculations in the meantime.
