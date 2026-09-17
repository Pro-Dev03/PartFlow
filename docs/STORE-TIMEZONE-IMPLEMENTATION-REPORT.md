# Store Timezone Implementation Report

Date: 2026-09-16

## Implemented

- Added persisted `store_timezone` as an IANA timezone setting.
- Added migration `063_store_timezone.sql`.
- Added backend `LoadStoreTimezone`, `StoreNow`, and `StoreMonthBounds` primitives.
- Loads the persisted timezone before dashboard and other services are constructed.
- Added one-time `POST /settings/regional/initialize` for device timezone detection.
- Restricted `PUT /settings/regional` to authenticated active subscribers; other administrative settings remain administrator-only.
- Validates timezone values with `time.LoadLocation`.
- Frontend detects the device IANA timezone once with `Intl.DateTimeFormat().resolvedOptions().timeZone`.
- Frontend now stores and uses the server-provided regional profile.
- Manual timezone changes are behind an explicit warning in Settings.
- Centralized frontend parsing and store-date helpers support ISO, offset, legacy Go timestamps, and Go `m=+...` suffixes.
- Reports, dashboard date calculations, debts display, and returns display now use store-time helpers.

## Verification

- Backend: `go test ./internal/accounting ./internal/settings ./internal/api ./internal/dashboard` PASS.
- Frontend: `npx vitest run src/__tests__/utils/store-time.test.ts` PASS.
- Frontend: `npm run build:check` PASS.
- Boundary tests for store `23:59`, `00:00`, and `00:01` PASS.
- Store timezone independence test using persisted `America/New_York` PASS.
- Legacy timestamp parsing tests PASS.

## Acceptance Status

| Area | Status | Evidence / gap |
|---|---|---|
| Auto detection | PASS | One-time frontend initialization endpoint and device IANA detection implemented. |
| Persistence | PASS | `store_timezone` setting and migration implemented. |
| Central source | PASS | Accounting loads persisted value before services are built. |
| Browser independence | PASS | Store-date utilities and backend tests use persisted timezone. |
| Frontend parsing | PASS | Central parser handles ISO, offsets, legacy Go strings, and monotonic suffixes. |
| Dashboard cards | PASS | Existing dashboard metrics use store date; live payment regression fixed and tested. |
| Debts display/aging | PASS | Display, customer overdue summary, dashboard overdue list, and debt summary now use store-date parameters. Additional legacy debt queries still need audit. |
| Returns display | PASS | Customer return display uses store formatting. Backend-wide return query audit remains. |
| Reports date ranges | PARTIAL | Frontend ranges use store timezone; several backend report queries still use database current-date expressions. |
| Profit / expenses | PARTIAL | Dashboard accounting period is store-bound; recurring/legacy report queries need migration. |
| Cash flow | PASS | Dashboard payment metrics use store date and legacy timestamp normalization. |
| Purchases | NOT PROVEN | Existing purchase/report query audit still has date functions requiring migration. |
| Inventory stale/aging | NOT PROVEN | Report queries still contain `CURRENT_DATE` windows. |
| Legacy data | PARTIAL | Frontend parser and SQLite payment/expense paths support legacy strings; every backend repository is not yet migrated. |
| Device timezone change scenarios | PASS | Store timezone is persisted and is not replaced by later device timezone reads. |

## Remaining Migration Work

The remaining backend date queries should be migrated from database-clock expressions to explicit `StoreDayBounds`/`StoreMonthBounds` parameters, especially:

- customer and debt overdue queries;
- dashboard overdue/aging queries;
- expenses monthly comparisons;
- inventory stagnant/aging report windows;
- supplier return creation timestamps and purchase report filters.

Until those queries are migrated and covered by boundary tests, the overall architecture is intentionally classified as `PARTIAL`, not fully closed.
