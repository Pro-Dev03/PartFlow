# Tenant isolation rollout

Migration `080_tenant_isolation.sql` restores one-store-per-tenant isolation in the shared PostgreSQL database. It must be applied before moving business traffic from SQLite to Supabase.

## Before applying

1. Stop every Render API and worker process that uses this Supabase database. The previous runtime connection may use the database owner role, which bypasses RLS even when policies exist.
2. Identify the existing account that owns the current shared business rows. The migration assigns those rows to that account's new tenant. Other existing accounts receive empty tenants; their old data is not copied into their stores.
3. In one Supabase SQL Editor execution, prepend these statements to the complete migration script. Replace the email. The runtime login role defaults to `postgres`; set it explicitly if Render's PostgreSQL login uses another database role:

   ```sql
   SELECT set_config('partflow.legacy_tenant_owner_email', 'owner@example.com', false);
   SELECT set_config('partflow.runtime_login_role', 'postgres', false);
   ```

   Both settings and the migration must run in the same database session. `partflow.runtime_login_role` must name the database role used by Render's `DATABASE_URL`; migration 080 grants that role permission to assume `partflow_runtime`. Do not put database URLs, passwords, or Supabase keys in the repository.

## Enabling the API

After the migration succeeds, deploy the updated API with `PARTFLOW_TENANT_RLS_ENABLED=true`. The API switches each database session to the non-login `partflow_runtime` role, checks that this role cannot bypass RLS, verifies that every business table has forced RLS and the PartFlow policy, and rejects startup if any check fails. It also verifies that memberships are visible only to their own account and cannot be modified by runtime queries. Keep the Supabase owner URL only in Render's server-side environment.

Then redeploy the hosted frontend at `https://partflow-hpv7.onrender.com/`. Its business API requests go to `https://partflow-api.onrender.com/api/v1` regardless of the device's local/cloud selector. If Render has `CORS_ALLOWED_ORIGINS` configured, include `https://partflow-hpv7.onrender.com`. Until both API migration and API deployment are complete, the frontend must not be switched to the new cloud-only traffic path.

The current repository layer shares one `*sqlx.DB` across handlers. To avoid tenant context crossing between pooled sessions, this rollout pins the runtime pool to one connection and serializes tenant-scoped HTTP requests. This is secure but limits throughput; replace it with request-bound transactions before scaling.

The worker service deliberately refuses to start while tenant RLS is enabled. Its scheduled jobs still use global queries and must be converted to explicit per-tenant scopes first. Do not leave an older worker running with owner credentials after applying the migration.

## Data and schema checks

- All pre-existing business rows are assigned to the selected legacy owner's tenant.
- New account registration creates a fresh tenant and owner membership in the database trigger.
- Inserts overwrite any client-supplied `tenant_id` with the authenticated tenant. Reads, updates, and deletes are filtered by RLS.
- Membership changes are created only by the database registration trigger; normal API role can read only its authenticated account's membership.
- Supabase `anon` and `authenticated` roles and `PUBLIC` lose access to the `public` schema and direct privileges on business tables and their owned sequences. Extension-owned public tables are excluded from tenant migration.
- Any future business table added after migration 080 will cause API startup to fail until it has a `tenant_id`, forced RLS, and the PartFlow policy.
- Local SQLite files are not moved by this migration. They remain local until the separate cloud business-data migration is implemented and verified.
- The current generic payment webhook does not identify a tenant. It returns `503 PAYMENT_WEBHOOK_TENANT_ROUTING_REQUIRED` while RLS is enabled; add a signed, tenant-aware callback route before enabling provider webhooks in production.
