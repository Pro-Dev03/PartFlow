# E2E Environment Setup

## Status

Current status: ENVIRONMENT BLOCKED / NOT PROVEN.

This is an environment requirement issue, not a Product/Business-Logic fix. No CloudGuard bypass, no auth bypass, no barcode logic changes, and no workflow workaround are allowed.

## 1. Cloud Environment

The local API is configured to validate protected requests against the cloud authority, not only against the local SQLite users table.

Evidence:
- [backend/.env](backend/.env#L1-L7) sets `PARTFLOW_CLOUD_API_URL=https://partflow-api.onrender.com/api/v1`
- [.env.example](.env.example#L27-L34) documents the same cloud authority requirement
- [backend/internal/auth/cloud_guard.go](backend/internal/auth/cloud_guard.go#L11-L52) rejects protected local requests unless a valid `X-PartFlow-Cloud-Token` is supplied
- [backend/internal/api/router.go](backend/internal/api/router.go#L109-L125) applies `auth.CloudGuard(authService)` to protected routes

Operational result:
- Local API is not a stand-alone auth environment for E2E.
- Local SQLite and Cloud auth are two different trust boundaries.
- Protected endpoints require both a valid local session and a valid cloud validation result.

## 2. Test Account

The supplied E2E account is:
- Email: `test@test.com`
- Password: `85265400`

Verification result from the active cloud environment:
- The cloud login call to `https://partflow-api.onrender.com/api/v1/auth/login` succeeded and returned an access token and user payload for `test@test.com`.
- The response included `subscription_status: active` and a valid `user.id`.

This proves that the test account exists in the cloud environment used by the local API.

## 3. Why `invalid credentials` is returned by the Local API

The local backend login flow is implemented in [backend/internal/auth/service.go](backend/internal/auth/service.go#L261-L338). It checks the local database table `users` directly:
- `SELECT ... FROM users WHERE email = $1 AND is_active = TRUE`
- then verifies the password hash against that row

The local server start log shows that the local backend is running against SQLite, not the cloud auth database:
- `Using SQLite database for local mode: C:\Users\Administrator\AppData\Roaming\PartFlow\data\partflow.db`

So the real operational issue is:
- the cloud account exists and can log in to the cloud auth service,
- but the local SQLite instance does not necessarily contain the same user row or is not synchronized with the cloud account data,
- therefore the local `/api/v1/auth/login` endpoint rejects the same credential pair with `invalid credentials`.

This is a data-source mismatch between Cloud auth and Local SQLite auth, not a login logic bug.

## 4. Is this account present in the Cloud/Auth environment used by the Local API?

Yes, in the cloud environment referenced by the app config, the login succeeds.

Verified evidence:
- Cloud login returned a valid response for `test@test.com` from `https://partflow-api.onrender.com/api/v1/auth/login`
- The response included:
  - `access_token`
  - `refresh_token`
  - `user.email = test@test.com`
  - `subscription_status = active`

This means the account exists in the cloud environment used by the local app configuration.

## 5. Is the Local API connected to the correct Cloud environment?

Yes, the local backend is configured to use the same cloud base URL:
- [backend/.env](backend/.env#L1-L7) sets `PARTFLOW_CLOUD_API_URL=https://partflow-api.onrender.com/api/v1`

This matches the cloud login endpoint that succeeded.

However, there is still a separate mismatch in the local auth layer:
- the app is validating protected routes against the cloud authority,
- but the public login endpoint used by the browser still authenticates against local SQLite users,
- so the same account can exist in Cloud while the local SQLite user table is missing or stale.

This is why E2E is blocked by environment mismatch, not by business logic.

## 6. Required flow to obtain a valid Cloud session/token for E2E

The correct sequence must be:

### Cloud Environment → Test Account → Login → Cloud Token/Session → Local API → Frontend → Playwright

1. Cloud Environment
   - Use the same cloud authority configured in [backend/.env](backend/.env#L1-L7)
   - Required base: `https://partflow-api.onrender.com/api/v1`

2. Test Account
   - The test account must exist and be active in that cloud identity environment
   - It must have a valid `subscription_status` and non-expired subscription

3. Login
   - Authenticate against the cloud login endpoint:
   - `POST /api/v1/auth/login`
   - Body must contain the real test user email and password for that cloud account

4. Cloud Token / Session
   - The cloud login response must produce a valid `access_token` / `refresh_token`
   - The UI and Playwright flow must persist the cloud token in local storage and send it as:
     - `Authorization: Bearer <token>`
     - `X-PartFlow-Cloud-Token: <token>`

5. Local API
   - The local backend validates protected routes using `CloudGuard`
   - The local API expects a valid cloud validation response before accepting protected operations

6. Frontend
   - The browser must maintain the valid cloud token/session and call protected endpoints with the proper headers

7. Playwright
   - Playwright must start only after the cloud session is valid and the frontend is authenticated in the same environment

## 7. Prerequisites before any Browser E2E can be considered valid

A real E2E run is only valid when all of the following are true:

- The cloud auth environment is the exact one configured in the app
- The test account exists in that same cloud environment
- The account is active and subscription-valid
- The login call returns a valid cloud token
- Local API protected routes receive the cloud token and validate successfully
- The browser obtains a working local session and can call protected APIs without `CLOUD_AUTH_REQUIRED`
- Playwright runs against that exact environment and not against a stale local SQLite table

## 8. Current classification

- Barcode Lifecycle = NOT PROVEN
- Opening Stock Lifecycle = NOT PROVEN
- Financial Lifecycle = NOT PROVEN
- Browser E2E = BLOCKED BY ENVIRONMENT

No bypass or workaround is allowed. Until the same cloud environment, test account, valid session, and local API chain are all active, no Browser E2E result may be classified as PASS.
