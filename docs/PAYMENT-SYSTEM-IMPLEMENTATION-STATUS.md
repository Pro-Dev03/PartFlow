# PartFlow Payment System Status

Date: 2026-09-17

## Scope

PartFlow now has a provider-independent payment transaction layer. Sales and the existing `payments` table remain the financial source of truth. Provider-specific behavior is isolated behind `PaymentProvider`.

## Architecture

- `internal/paymentproviders`: shared contract, lifecycle states, and Cardcom adapter.
- `internal/paymentproviders/registry.go`: provider catalog and adapter factory. Cardcom is registered; other named providers are catalogued as not implemented.
- `internal/paymenttransactions`: transaction orchestration, idempotency, verification, cancellation, refunds, webhook event persistence, and financial finalization.
- `internal/secrets`: AES-GCM encryption for payment secrets using `PARTFLOW_PAYMENT_ENCRYPTION_KEY`.
- `internal/settings`: provider/environment/merchant/terminal/webhook settings with secret redaction.
- `frontend/src/features/settings/components/ElectronicPaymentSettings.tsx`: payment settings UI and real connection-test action.

Supported lifecycle states:

`pending -> processing -> paid|failed|cancelled -> partially_refunded -> refunded`

Invalid transitions are rejected. A webhook is treated as a notification; `VerifyPayment` performs the authoritative provider lookup before a transaction becomes paid.

## Database

Migration `070_payment_provider_transactions.sql` creates:

- `payment_transactions`
- `payment_refunds`
- `payment_webhook_events`

Migration `071_payment_transaction_order_id.sql` adds `order_id` for pre-sale checkout flows. Migration `072_return_payment_refunds.sql` adds durable return-to-payment refund state. Unique keys protect provider/idempotency requests, provider webhook event IDs, and one refund workflow per return. Refund idempotency and remaining-balance checks happen before the external provider is called.

## API

Protected routes:

- `POST /api/v1/payment-transactions`
- `GET /api/v1/payment-transactions/:id`
- `POST /api/v1/payment-transactions/:id/verify`
- `POST /api/v1/payment-transactions/:id/cancel`
- `POST /api/v1/payment-transactions/:id/refund`
- `POST /api/v1/payment-providers/test-connection`
- `GET /api/v1/payment-providers`

Provider webhook route:

- `POST /api/v1/payment-webhooks/:provider`

The webhook is recorded with a unique provider event ID and then triggers server-side verification. It does not directly mark a sale paid.

## Settings

Payment settings are stored through the existing settings system:

- `electronic_payments_enabled`
- `payment_provider`
- `payment_environment`
- `payment_public_key`
- `payment_secret_key`
- `payment_merchant_id`
- `payment_terminal_id`
- `payment_webhook_url`
- `payment_webhook_secret`
- `payment_methods`

Secrets are encrypted at rest and returned as `********`. The encryption key must be configured in the backend environment and must never be sent to React or localStorage.

Provider choices in Settings are loaded from `GET /payment-providers`. Planned adapters remain visibly unavailable until their factory is registered. Cardcom-specific fields are shown in the UI only when Cardcom is selected. The connection test calls the backend adapter and reports the real response.

## POS and financial flow

The intended electronic flow is:

`POS -> create pending transaction -> provider checkout -> backend verify -> finalize payment and sale`

`FinalizePaid` writes the existing `payments` projection and updates the sale in one database transaction. It does not create a second financial ledger.

A pre-sale payment can be linked to a sale using `payment_transaction_id`; the sale endpoint accepts it only after the transaction is paid and amount-validated.

Customer return completion now checks for a paid electronic transaction on the sale. If present, it creates/updates `return_payment_refunds`, invokes the configured payment transaction service with `return-refund:<return_id>` idempotency, and only then commits return inventory and customer-ledger effects. Manual/cash sales skip the provider. A provider failure records `failed` refund state and leaves the return uncompleted for retry.

## Provider evidence

| Provider | Adapter | External sandbox/credentials tested here | Status |
|---|---:|---:|---|
| Cardcom | Implemented | No credentials supplied in this workspace | PARTIAL |
| PayMe | Contract name reserved | Not implemented | NOT PROVEN |
| Grow | Contract name reserved | Not implemented | NOT PROVEN |
| PalPay | Contract name reserved | Public API contract was not sufficient for implementation | NOT PROVEN |
| Stripe | Contract name reserved | Not implemented | NOT PROVEN |
| PayPal | Contract name reserved | Not implemented | NOT PROVEN |

Cardcom integration uses the documented v11 endpoints for low-profile creation, result lookup, connection testing, and transaction refund. A real provider `PASS` requires a merchant sandbox account, configured credentials, a reachable webhook URL, and an end-to-end transaction run.

Terminal hardware is not assumed. Cardcom hosted checkout/API is separate from physical reader integration; a reader model and its official SDK/API are required before claiming Terminal support.

## Security position

- Secrets are backend-only and encrypted at rest.
- Provider status is decided by backend verification.
- Payment and refund requests have idempotency keys.
- Duplicate webhook events are ignored by database uniqueness.
- Invalid payment state transitions are rejected.
- Offline mode must not call these provider routes or bypass verification.

Cardcom's webhook payload is treated as an untrusted trigger. The adapter does not claim a Cardcom signature scheme without an official account-specific signing contract; verification is always performed against Cardcom before local finalization.

## Verification report

| Area | Result | Evidence |
|---|---|---|
| Database | PASS | migrations 070, 071, and 072; SQLite return/refund lifecycle tests |
| Backend provider contract | PASS | provider package tests |
| Provider registry | PASS locally | registry descriptor and factory tests |
| Cardcom adapter behavior | PASS locally / NOT PROVEN externally | httptest contract tests; no merchant sandbox credentials |
| Payment API | PASS locally | API package tests and route registration |
| Settings UI | PASS locally | TypeScript check; real test-connection call |
| POS provider checkout | PARTIAL | transaction orchestration and cancellation API exist; live provider flow is not proven |
| Webhooks | PASS locally / NOT PROVEN externally | duplicate events stop before Verify; new events are re-verified; external delivery not proven |
| Refunds | PASS locally / NOT PROVEN externally | full/partial path, remaining-balance guard, idempotency, failure state, and return orchestration tests pass; external provider verification is not proven |
| Returns | PASS locally / NOT PROVEN externally | electronic full return, partial return, manual return, duplicate completion, and provider failure tests pass |
| Inventory | PASS locally / NOT PROVEN externally | completion updates inventory once; provider failure occurs before inventory mutation |
| Financial ledger | PASS locally / NOT PROVEN externally | existing sale/payment/customer-ledger paths remain transactional; provider-backed report reconciliation is not proven |
| Customer financial effect | PASS locally / NOT PROVEN externally | existing debt adjustment idempotency remains covered; external refund settlement is not proven |
| Reports/dashboard | NOT PROVEN | no live provider transaction was available for report evidence |
| Security | PASS locally / PARTIAL operationally | encryption and redaction tests pass; production key management remains deployment responsibility |
| Documentation | PASS | this document and provider contract comments |
| E2E | NOT PROVEN | no provider-backed browser run was available |
| Production readiness | NOT PROVEN | requires sandbox credentials, webhook reachability, and end-to-end evidence |

## Required production setup

Set `PARTFLOW_PAYMENT_ENCRYPTION_KEY` in the backend environment, configure Cardcom credentials only through protected Settings, expose the webhook URL over HTTPS, run migrations, and execute a real sandbox payment followed by verify, duplicate webhook, partial refund, and full refund checks. Never place provider secrets in frontend environment variables.

## Adding a provider

1. Implement `PaymentProvider` in `internal/paymentproviders`.
2. Register its adapter in the handler/provider registry.
3. Add provider-specific settings metadata and UI fields.
4. Add contract tests for create, verify, cancel, refund, webhook, timeout, and idempotency.
5. Add sandbox evidence before changing the provider status from `NOT PROVEN`.
