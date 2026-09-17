# Purchases, Supplier Balance, Payments, and Inventory Audit

Date: 2026-09-16
Scope: Purchases -> Supplier Balance/Payments/Returns -> Inventory
Barcode logic: not changed.

## Status Rules

`PASS` requires an executable test or live evidence for the complete claim. Static code inspection alone is not PASS. `PARTIAL` means some layer is tested but DB -> Backend -> API -> Frontend -> E2E continuity is incomplete. `FAIL` means an executable path contradicts the requirement. `NOT PROVEN` means no sufficient evidence was found.

## Evidence Matrix

| Entity | Official Business Date | DB | Backend | API | Frontend | E2E | Status |
|---|---|---|---|---|---|---|---|
| Purchase creation | `purchases.purchase_date` | SQLite create/ledger test exists | `CreatePurchase` is transactional and persists items | `POST /purchases` exists | `purchasesApi.create` and purchase form exist | No live purchase E2E | PARTIAL |
| Supplier link on Purchase | Purchase `supplier_id` | Schema and joins exist | Purchase validation requires supplier | API request carries `supplier_id` | Supplier selector is wired in purchase flow | No live proof | PARTIAL |
| Purchase Items | Purchase business date inherited from Purchase | `purchase_items` persistence test exists; barcode continuity is separately proven | Quantity/cost validation exists | Items returned by `GET /purchases/:id` | Purchase details/hooks consume item data | No live proof for quantity/cost reconciliation | PARTIAL |
| Receive | Purchase `purchase_date`; receipt event has no separate official date identified | Inventory rows are created by receive tests | `ReceivePurchase` is idempotent for existing item codes | `POST /purchases/:id/receive` exists | Receive mutation exists | No live UI receive proof | PARTIAL |
| Received vs remaining quantity | Purchase/receipt state, not `created_at` | Repository computes counts from inventory and supplier return rows | `ReceivedQuantity`, `ReturnedQuantity`, `AvailableForReturn` are computed | Returned in purchase response | Displayed by purchase UI | No E2E reconciliation | PARTIAL |
| Purchase Cost | `purchase_date` for purchase reporting | `purchase_items.unit_price/item_total`, inventory `purchase_cost` exist | Receive copies unit cost into inventory | Purchase and inventory APIs expose costs | Hooks calculate/display totals | No E2E reconciliation | PARTIAL |
| Supplier balance | Ledger transaction date / purchase business date | SQLite lifecycle test seeds purchase debit and asserts balance transitions | Supplier payment service updates balance; return service creates credit | Supplier ledger/payment APIs exist | Supplier ledger/payment UI is wired | No live E2E proof | PARTIAL |
| Supplier Payments | Payment date (`payments.payment_date` where present) | SQLite lifecycle test records partial/full payments and counts ledger rows | Partial/full/overpayment and duplicate reference protection are executable | `POST /purchases/:id/payment`, `POST /suppliers/:id/payments` exist | API wrappers exist | No live UI/E2E | PARTIAL |
| Supplier Returns | Return created/updated date; original Purchase remains authoritative | SQLite lifecycle test verifies `RETURNED` item, refund, movement, and one credit | Create/AddItem/Complete path links Purchase/Supplier and removes inventory availability | Supplier return routes and API wrappers exist | Supplier return page/API exists | No live UI/E2E | PARTIAL |
| Quantity Products | Inventory business state, no single item date | Aggregate `inventory` and `inventory_items` schemas exist | Inventory service handles quantity creation | Inventory item APIs exist | Inventory hooks load aggregate and item views | No live quantity flow | PARTIAL |
| Individual Products | `purchase_date` for acquired item; sale/return have their own dates | `inventory_items` stores item identity/status/cost/supplier | Sale service selects an available individual item | Inventory and sales APIs expose `inventory_item_id` | POS can submit item ID | No live individual-item E2E in this audit | PARTIAL |
| Used Items | `purchase_date`/acquisition date depending source | Used condition is stored on inventory/acquisition rows | Used/trade-in path exists | Trade-in endpoint exists | Used Parts UI exists | No live used-item report proof | PARTIAL |
| Items without Supplier | Source event date | SQLite test proves nullable supplier and persisted item identity | Used item creation accepts nil SupplierID | Inventory API accepts optional supplier_id | UI path exists indirectly | No live UI/E2E for sale/return | PARTIAL |
| Opening Stock | `inventory_movements.business_date` supplied explicitly | `source_type=OPENING_STOCK` movement plus aggregate `inventory` or nullable-supplier `inventory_items` | `POST /inventory/opening-stock` records Quantity or Individual/Used stock without Purchase | Opening Stock modal and API client are present | Passing `opening-stock-lifecycle.spec.ts` covers UI/API/identity flow | DB/Backend/API/UI/E2E proof; screenshots and trace produced | PASS |
| Stock Movements | Movement event `created_at` unless a domain date exists | `inventory_movements` table is used | Movement creation/reversal paths exist | Inventory history endpoint exists | Inventory ledger consumes history | No live movement E2E | PARTIAL |
| Negative inventory quantities | Inventory adjustment business event date | SQLite test proves `-1` and `0` create no rows; `1` creates one; `10001` is rejected | `CreateInventoryItem` rejects `quantity <= 0` and retains the upper limit | HTTP test proves `-1`, `0`, and `10001` return 400 | UI validation not proven | No negative-quantity E2E | PARTIAL |
| Purchase/Sale/Return inventory linkage | Each source event's official date | Foreign-key-like IDs and item-code conventions are used | Purchase receive, sale, and return services link records | APIs expose the references | Frontend consumes linked records | No full live chain | PARTIAL |

## Concrete Findings

1. **Negative quantity behavior is now explicitly rejected.** `CreateInventoryItem` no longer normalizes invalid input. `TestCreateInventoryItemQuantityValidationSQLite` covers service/SQLite behavior for `-1`, `0`, `1`, and `10001`; `TestCreateInventoryItemRejectsInvalidQuantityAPI` covers HTTP rejection. The final status remains PARTIAL because no live UI/E2E evidence exists.
2. **Purchase business date now has an executable regression.** `TestSQLitePurchaseReadsUseOfficialBusinessDate` sets `purchase_date` and `created_at` to different values and proves `GetByID` and `ListSummaries` use `purchase_date`. SQLite payment loading and legacy schema fallback were also routed through the official column when present. API/UI/E2E evidence is still missing.
3. **Supplier Ledger lifecycle now has executable SQLite evidence.** `TestSupplierLedgerPaymentLifecycleSQLite` proves partial/full payment, overpayment rejection, duplicate reference rejection, and no duplicate payment ledger row. `TestSupplierReturnCreditsLedgerAndRemovesInventorySQLite` proves supplier-return credit and inventory removal. UI/E2E evidence remains open.
4. **Opening Stock is now proven end to end.** `source_type=OPENING_STOCK` and `business_date` are stored on inventory movements; Quantity updates aggregate inventory, while Individual/Used creates an item with optional Supplier and no Purchase. The passing E2E preserves identity through Inventory, POS, Sale, and Customer Return.
5. **No live Frontend/E2E evidence exists for this audit scope.** Existing Playwright tests cover navigation/POS hardening, not the complete purchases and inventory financial lifecycle.

## Remediation Status

- Negative quantity normalization was removed. Invalid `-1` and `0` inputs are rejected with `inventory quantity must be greater than zero`; `10001` remains rejected by the explicit maximum; trade-in now sends `quantity: 1` explicitly.
- SQLite purchase reads and payment loading now select `purchase_date` when the column exists. Legacy fallback to `created_at` is limited to schemas that do not have `purchase_date`.
- Regression coverage now proves the official purchase date differs from `created_at` and is used by `GetByID` and `ListSummaries`.
- The Purchases/Inventory audit gate remains **OPEN / PARTIAL**. Supplier Ledger and Opening Stock backend/API evidence now exists; live UI/E2E evidence still requires focused tests before this scope can close.

## Existing Executable Evidence

- Purchase package: `TestReceivePurchaseSkipsDuplicateInventoryItemsOnSQLite`, `TestCreatePurchaseAndReceivePreserveBarcodeContinuityOnSQLite`, and atomic ledger/audit test.
- Inventory package: condition/grade/status/request helper tests, SQLite timestamp repository test, and trade-in handler test.
- Inventory quantity package: `TestCreateInventoryItemQuantityValidationSQLite` and `TestCreateInventoryItemRejectsInvalidQuantityAPI`.
- Purchase business-date package: `TestSQLitePurchaseReadsUseOfficialBusinessDate`.
- Supplier ledger package: `TestSupplierLedgerPaymentLifecycleSQLite`.
- Supplier return package: `TestSupplierReturnCreditsLedgerAndRemovesInventorySQLite` plus SQLite text-timestamp list test.
- Supplier-less package: `TestCreateUsedInventoryItemWithoutSupplierPreservesIdentitySQLite`.
- Opening Stock package: `TestCreateOpeningStockSQLiteSupportsQuantityAndIndividualWithoutPurchase` and `TestCreateOpeningStockAPI`.
- Opening Stock E2E: `frontend/e2e/opening-stock-lifecycle.spec.ts` covers UI creation, central barcode resolution, POS, Sale, and Customer Return identity evidence; it requires `E2E_EMAIL`, `E2E_PASSWORD`, and `E2E_RUN_OPENING_STOCK_LIFECYCLE=true`.
- The full DB -> Backend -> API -> Frontend -> E2E matrix remains open until focused tests and live E2E are added.
