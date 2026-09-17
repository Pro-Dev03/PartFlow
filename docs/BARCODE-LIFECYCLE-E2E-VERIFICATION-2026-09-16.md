# Barcode Lifecycle E2E Verification

Date: 2026-09-16
Barcode under test: `FNX-GPU-000421`

## Acceptance Status

The barcode integration is **not closed as PASS yet**. Backend continuity and frontend resolver wiring are passing, but the requested live Browser E2E could not be executed in this workspace because no E2E credentials were available and no local listeners were running on ports 5174 or 8080.

The accepted current result is:

- **Purchase -> Receive -> Inventory: PASS**
- **Browser E2E: PARTIAL**
- **Barcode Lifecycle harness: Ready for Live E2E**
- Barcode integration remains open until Sale, Customer Return, Supplier Return, and Reports are proven from the real UI.

## Evidence Matrix

| Stage | Status | UI evidence | API evidence | DB evidence |
|---|---|---|---|---|
| Category -> Product | PARTIAL | Existing product/category screens compile; live browser run unavailable | Product model supports barcode | Existing barcode resolver tests use stable product IDs |
| Product -> Purchase | PASS (backend) / PARTIAL (UI) | Purchase UI exists; live browser run unavailable | `PurchaseItemRequest.Barcode` is persisted | `TestCreatePurchaseAndReceivePreserveBarcodeContinuityOnSQLite` verifies purchase item barcode |
| Purchase -> Receive | PASS (backend) / PARTIAL (UI) | Receive screen not exercised live | Receive preserves the original barcode for a single item | Same test verifies purchase and inventory barcode equality |
| Receive -> Inventory | PASS (backend) / PARTIAL (UI) | Inventory UI was not exercised live | `/barcodes/resolve/:code` returns product and inventory item | SQLite test verifies one inventory row and stable barcode |
| Inventory -> POS scan | PARTIAL | POS scanner exists and now consumes the central resolver response | `barcodeApi.lookupProduct` now calls `/barcodes/resolve/:code` | Resolver tests verify Product ID and Item ID continuity |
| POS -> Sale | PARTIAL | POS sale was not executed in a browser | POS now includes `inventory_item_id` when resolver returns an individual item | Sale service requires the referenced item to be AVAILABLE and does not create an item |
| Sale -> Customer Return | PARTIAL | Customer return UI was not executed in a browser | Return request supports `sale_item_id` and `inventory_item_id` | Existing resolver lifecycle tests verify return references |
| Customer Return -> Supplier Return | PARTIAL | Supplier-return UI was not executed in a browser | Supplier return uses `purchase_id` and `purchase_item_id` | Backend service links supplier return to the original purchase and supplier |
| Reports | PARTIAL | Inventory, purchases, returns, and used-item report screens were not exercised live | Resolver exposes purchase, sale, and return IDs | Existing lifecycle fixture verifies linked purchase, sale, and return rows |

## Negative Scenarios

| Scenario | Status | Evidence |
|---|---|---|
| Duplicate barcode rejected | PARTIAL | Inventory service returns `ErrDuplicateBarcode`; live UI/API assertion not run |
| SOLD barcode cannot be sold again | PARTIAL | Sale service only accepts `AVAILABLE` inventory items; live double-sale attempt not run |
| Supplier-less item remains searchable | PARTIAL | Resolver does not require supplier ID; live fixture was not run |
| Sale does not create a new item | PASS (service rule) / PARTIAL (E2E) | Sale uses `inventory_item_id` and updates the existing item; browser evidence unavailable |
| Customer return preserves `item_id` | PASS (fixture) / PARTIAL (E2E) | Existing SQLite resolver fixture links `return_items.inventory_item_id`; live UI evidence unavailable |
| Used item keeps barcode in Used Items report | PARTIAL | Used item UI/report was not run against a live database |

## Automated Validation Completed

- `go test ./...`: PASS
- `go test ./internal/barcodes ./internal/purchases`: PASS
- `npm run test:run`: PASS (14 files, 46 tests)
- `npm run lint:types`: PASS
- `npm run build`: PASS
- `npm run lint`: PASS with pre-existing warnings

## Browser E2E Blocker

The Playwright configuration targets `http://localhost:5174` and the application targets the local API at `http://localhost:8080/api/v1`. At verification time:

- `E2E_EMAIL` was empty.
- `E2E_PASSWORD` was empty.
- No listeners were available on ports 5174 or 8080.

Therefore no UI/API/DB claim for the full Purchase -> Receive -> Inventory -> POS -> Sale -> Customer Return -> Supplier Return path is marked PASS. The acceptance gate remains open until the live Playwright run produces trace, network, UI, and database evidence for the same barcode and IDs.

## Prepared Live Harness

The live Playwright harness is available at [barcode-lifecycle.spec.ts](../frontend/e2e/barcode-lifecycle.spec.ts). It is gated by `E2E_RUN_BARCODE_LIFECYCLE=true` plus `E2E_EMAIL` and `E2E_PASSWORD`; without those values it reports `skipped`, never PASS. It creates its fixture through the real API, scans the real UI, captures resolver network responses and screenshots, executes sale and return mutations, and asserts duplicate and double-sale rejection. Its identity assertions are API evidence; they are intentionally not labeled DB evidence. DB evidence must be collected from the actual local/cloud database during the live run before updating this report.

## Resolver Safety Fix

The resolver no longer selects an unrelated sale item when a lifecycle row contains the barcode but `inventory_items.barcode` is empty. Fallback resolution is now constrained to the Product linked by the matching purchase or return barcode.

This was treated as an identity-safety defect, not only a test failure: an unrelated `sale_item` could make the UI appear to show the right product while returning the wrong backend `item_id`. The fix is therefore a prerequisite for the live E2E, which must record DB -> Backend -> API -> UI evidence for Product ID, Item ID, Barcode, Serial, Purchase ID, and Supplier ID at every stage.
