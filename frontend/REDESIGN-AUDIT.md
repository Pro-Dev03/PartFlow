# PartFlow interface review and design direction

Baseline commit: `d1d31ac`.

## Design direction

- Use one calm visual language: neutral surfaces, a single blue action color, semantic success/warning/error colors, readable Arabic type, and consistent 8/12/16/24 px spacing.
- Keep page titles, supporting text, primary actions, filters, content, and feedback in the same order. Use a distinct cashier layout for the point of sale without changing its transaction flow.
- Keep cards informational. Put actions in a small, predictable action area. Use the same focus, disabled, loading, empty, and error treatment in both themes.
- On narrow screens, give navigation readable labels and let tables scroll inside their own surface. Stack forms and action rows without hiding operations.
- Preserve routes, API calls, permissions, and business calculations. Review rendering and requests before changing any data flow.

## Page inventory

| Area | Pages reviewed | Main interface work |
| --- | --- | --- |
| Access | Login, SubscriptionExpired | Form hierarchy, status explanation, responsive access surface |
| Dashboard | Dashboard, Activity | Metrics, activity density, loading and empty states |
| Sales | POS | Cashier layout, search, cart, payment feedback, invoice dialog |
| Stock | Inventory, Categories, PartTypes, Archive | Product list, filters, stock details, categories, history |
| People | Customers, CustomerPurchases, Suppliers, SellerBalances | Record cards, balance summaries, actions, detail tables |
| Purchasing | Purchases, CreatePurchase, EditPurchase, PurchaseDetails | List, staged entry dialog, edit form, invoice details |
| Finance | Debts, Expenses, Reports | Dense tables, amounts, filters, charts, exports |
| Returns | Returns, CreateReturn, ReturnDetails, return-details/ReturnDetails, SupplierReturns | List, forms, detail sections, separate supplier flow |
| System | Settings | Account overview, navigation, all settings sections |

This is 26 page modules. `PartTypes`, `SellerBalances`, and the legacy `return-details/ReturnDetails` module are not direct routes in the current router; they are still included in the interface review. No removed feature is reinstated.

## Findings before implementation

- The shared visual stylesheet already covers many pages, but page headers and local action rows still vary. Several components use fixed light colors that do not follow theme tokens.
- The header search accepts text without performing a search. The mobile sidebar renders as an icon rail with no readable labels. Several active routes are missing from its menu.
- The current app already loads page modules lazily. Keep this split and avoid adding page wide dependencies to the entry bundle.
- Several pages contain large lists and tables. Their existing pagination and API requests should be preserved while shared table surfaces and responsive overflow are improved.
- Authentication, cashier payments, purchase creation, returns, and settings have distinct forms. Their data and validation rules must remain intact while field labels and feedback are made consistent.

## Verification targets

1. Production TypeScript and Vite build after the shared layer and after page work.
2. Navigation and visible content for every routed page at desktop, tablet, and mobile widths, in both themes.
3. Read only browser checks for authentication, search, filters, dialogs, cards, and tables; transactional checks only with safe test records or mocks.
4. Check console and failed network requests, then report any limits honestly.
