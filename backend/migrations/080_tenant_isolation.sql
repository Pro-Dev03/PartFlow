-- Restore per-store isolation before business traffic is moved from desktop
-- SQLite into the shared Supabase database.
--
-- Set partflow.legacy_tenant_owner_email in the SQL session when running the migration. Existing
-- global business rows are assigned to exactly that account's tenant. Other
-- existing accounts receive empty tenants so historic data is never copied
-- across subscriptions.
--
-- In Supabase SQL Editor, run this in the same session as the migration:
-- SELECT set_config('partflow.legacy_tenant_owner_email', 'owner@example.com', false);
-- SELECT set_config('partflow.runtime_login_role', 'partflow_app_login', false);
-- Replace the email with the account that owns the existing shared business data.
-- Create a dedicated NOINHERIT, non-owner LOGIN role for Render and use it in
-- DATABASE_URL. Never use the Supabase postgres/service-owner credentials.
-- Stop all Render processes connected to this database before applying it.

BEGIN;

CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tenant_memberships (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    membership_role TEXT NOT NULL DEFAULT 'owner' CHECK (membership_role = 'owner'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- An applied marker prevents a second run from moving every tenant's data back
-- to the legacy owner's tenant.
CREATE TABLE IF NOT EXISTS partflow_tenant_schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tenant_memberships_tenant_id
    ON tenant_memberships (tenant_id);

-- The application uses a dedicated non-owner role. It must never be a
-- superuser or have BYPASSRLS, otherwise PostgreSQL would ignore the policies.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'partflow_runtime') THEN
        CREATE ROLE partflow_runtime NOLOGIN NOSUPERUSER NOBYPASSRLS;
    ELSIF EXISTS (
        SELECT 1 FROM pg_roles
        WHERE rolname = 'partflow_runtime' AND (rolsuper OR rolbypassrls)
    ) THEN
        RAISE EXCEPTION 'partflow_runtime must not be SUPERUSER or BYPASSRLS';
    END IF;
    ALTER ROLE partflow_runtime NOLOGIN NOINHERIT NOSUPERUSER NOBYPASSRLS;
    IF EXISTS (
        SELECT 1 FROM pg_auth_members
        WHERE member = (SELECT oid FROM pg_roles WHERE rolname = 'partflow_runtime')
    ) THEN
        RAISE EXCEPTION 'partflow_runtime must not be a member of any other database role';
    END IF;
    EXECUTE format('GRANT CONNECT ON DATABASE %I TO partflow_runtime', current_database());
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'anon') THEN
        REVOKE ALL PRIVILEGES ON SCHEMA public FROM anon;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'authenticated') THEN
        REVOKE ALL PRIVILEGES ON SCHEMA public FROM authenticated;
    END IF;
END $$;

REVOKE ALL PRIVILEGES ON SCHEMA public FROM PUBLIC, partflow_runtime;
GRANT USAGE ON SCHEMA public TO partflow_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO partflow_runtime;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO partflow_runtime;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO partflow_runtime;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT USAGE, SELECT ON SEQUENCES TO partflow_runtime;

-- The application only needs to resolve the authenticated user's membership.
-- Memberships are created by the SECURITY DEFINER registration trigger, so
-- runtime SQL must never be able to create, move, or remove a user's tenant.
REVOKE ALL PRIVILEGES ON TABLE public.tenants FROM PUBLIC, partflow_runtime;
REVOKE ALL PRIVILEGES ON TABLE public.tenant_memberships FROM PUBLIC, partflow_runtime;
GRANT SELECT ON TABLE public.tenant_memberships TO partflow_runtime;
REVOKE ALL PRIVILEGES ON TABLE public.partflow_tenant_schema_migrations FROM PUBLIC, partflow_runtime;
GRANT SELECT ON TABLE public.partflow_tenant_schema_migrations TO partflow_runtime;

CREATE OR REPLACE FUNCTION public.partflow_assign_tenant_id()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    request_tenant_id UUID;
BEGIN
    request_tenant_id := NULLIF(current_setting('partflow.tenant_id', true), '')::uuid;
    IF request_tenant_id IS NULL THEN
        RAISE EXCEPTION 'tenant context is required';
    END IF;
    -- Never trust a tenant_id supplied by the browser or a sync payload.
    NEW.tenant_id := request_tenant_id;
    RETURN NEW;
END;
$$;

REVOKE ALL ON FUNCTION public.partflow_assign_tenant_id() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.partflow_assign_tenant_id() TO partflow_runtime;

DO $$
DECLARE
    legacy_owner_email TEXT := NULLIF(current_setting('partflow.legacy_tenant_owner_email', true), '');
    runtime_login_role TEXT := NULLIF(current_setting('partflow.runtime_login_role', true), '');
    runtime_login_oid OID;
    runtime_login_is_superuser BOOLEAN;
    runtime_login_bypasses_rls BOOLEAN;
    runtime_login_can_login BOOLEAN;
    runtime_login_inherits BOOLEAN;
    runtime_login_owns_database BOOLEAN;
    runtime_login_owns_public_objects BOOLEAN;
    legacy_owner_id UUID;
    legacy_tenant_id CONSTANT UUID := '08000000-0000-4000-8000-000000000001';
    existing_user RECORD;
    new_tenant_id UUID;
    business_table RECORD;
    business_view RECORD;
    business_sequence RECORD;
    unsupported_materialized_view TEXT;
    existing_policy RECORD;
BEGIN
    IF legacy_owner_email IS NULL THEN
        RAISE EXCEPTION 'partflow.legacy_tenant_owner_email is required for migration 080';
    END IF;
    IF runtime_login_role IS NULL THEN
        RAISE EXCEPTION 'partflow.runtime_login_role is required; use a dedicated restricted LOGIN role for Render';
    END IF;
    SELECT oid, rolsuper, rolbypassrls, rolcanlogin, rolinherit
    INTO runtime_login_oid, runtime_login_is_superuser, runtime_login_bypasses_rls,
         runtime_login_can_login, runtime_login_inherits
    FROM pg_roles WHERE rolname = runtime_login_role;
    IF runtime_login_oid IS NULL THEN
        RAISE EXCEPTION 'runtime login role % was not found; set partflow.runtime_login_role to the Render database login role', runtime_login_role;
    END IF;
    IF runtime_login_is_superuser OR runtime_login_bypasses_rls OR NOT runtime_login_can_login OR runtime_login_inherits THEN
        RAISE EXCEPTION 'runtime login role % must be LOGIN, NOINHERIT, NOSUPERUSER, and NOBYPASSRLS', runtime_login_role;
    END IF;
    SELECT EXISTS (
        SELECT 1 FROM pg_database WHERE datname = current_database() AND datdba = runtime_login_oid
    ) INTO runtime_login_owns_database;
    IF runtime_login_owns_database THEN
        RAISE EXCEPTION 'runtime login role % must not own the application database', runtime_login_role;
    END IF;
    SELECT EXISTS (
        SELECT 1 FROM pg_namespace WHERE nspname = 'public' AND nspowner = runtime_login_oid
        UNION ALL
        SELECT 1 FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public' AND c.relowner = runtime_login_oid
        UNION ALL
        SELECT 1 FROM pg_proc p JOIN pg_namespace n ON n.oid = p.pronamespace
        WHERE n.nspname = 'public' AND p.proowner = runtime_login_oid
    ) INTO runtime_login_owns_public_objects;
    IF runtime_login_owns_public_objects THEN
        RAISE EXCEPTION 'runtime login role % must not own public schema objects', runtime_login_role;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_auth_members WHERE member = runtime_login_oid) THEN
        RAISE EXCEPTION 'runtime login role % must not be a member of any database role before migration 080', runtime_login_role;
    END IF;
    EXECUTE format('ALTER ROLE %I NOINHERIT NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOREPLICATION', runtime_login_role);
    EXECUTE format('REVOKE CREATE, TEMPORARY ON DATABASE %I FROM %I', current_database(), runtime_login_role);
    EXECUTE format('REVOKE ALL PRIVILEGES ON SCHEMA public FROM %I', runtime_login_role);
    EXECUTE format('REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM %I', runtime_login_role);
    EXECUTE format('REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public FROM %I', runtime_login_role);
    EXECUTE format('GRANT CONNECT ON DATABASE %I TO %I', current_database(), runtime_login_role);
    EXECUTE format('GRANT partflow_runtime TO %I', runtime_login_role);
    IF EXISTS (SELECT 1 FROM public.partflow_tenant_schema_migrations WHERE version = 80) THEN
        RAISE EXCEPTION 'tenant isolation migration 080 is already applied; refusing to remap tenant data';
    END IF;

    IF current_setting('server_version_num')::INTEGER < 150000 THEN
        RAISE EXCEPTION 'migration 080 requires PostgreSQL 15 or newer for security-invoker views';
    END IF;
    SELECT c.relname INTO unsupported_materialized_view
    FROM pg_class c
    JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = 'public'
      AND c.relkind = 'm'
      AND NOT EXISTS (
          SELECT 1 FROM pg_depend d
          WHERE d.classid = 'pg_class'::regclass AND d.objid = c.oid AND d.deptype = 'e'
      )
    LIMIT 1;
    IF unsupported_materialized_view IS NOT NULL THEN
        RAISE EXCEPTION 'public materialized view % must be removed or redesigned for tenant isolation', unsupported_materialized_view;
    END IF;

    SELECT id INTO legacy_owner_id
    FROM users
    WHERE lower(email) = lower(legacy_owner_email)
    LIMIT 1;

    IF legacy_owner_id IS NULL THEN
        RAISE EXCEPTION 'legacy tenant owner % was not found in users', legacy_owner_email;
    END IF;

    INSERT INTO tenants (id, name)
    SELECT legacy_tenant_id, COALESCE(NULLIF(first_name, ''), 'Legacy') || ' Store'
    FROM users WHERE id = legacy_owner_id
    ON CONFLICT (id) DO NOTHING;

    INSERT INTO tenant_memberships (user_id, tenant_id, membership_role)
    VALUES (legacy_owner_id, legacy_tenant_id, 'owner')
    ON CONFLICT (user_id) DO UPDATE SET tenant_id = EXCLUDED.tenant_id;

    FOR business_view IN
        SELECT c.relname
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public'
          AND c.relkind = 'v'
          AND NOT EXISTS (
              SELECT 1 FROM pg_depend d
              WHERE d.classid = 'pg_class'::regclass AND d.objid = c.oid AND d.deptype = 'e'
          )
    LOOP
        -- PostgreSQL 15+ security-invoker views evaluate underlying table
        -- privileges and RLS policies as partflow_runtime instead of as owner.
        EXECUTE format('ALTER VIEW public.%I SET (security_invoker = true)', business_view.relname);
        EXECUTE format('REVOKE ALL PRIVILEGES ON TABLE public.%I FROM PUBLIC', business_view.relname);
        IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'anon') THEN
            EXECUTE format('REVOKE ALL PRIVILEGES ON TABLE public.%I FROM anon', business_view.relname);
        END IF;
        IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'authenticated') THEN
            EXECUTE format('REVOKE ALL PRIVILEGES ON TABLE public.%I FROM authenticated', business_view.relname);
        END IF;
    END LOOP;

    FOR existing_user IN
        SELECT id, first_name
        FROM users
        WHERE id <> legacy_owner_id
          AND NOT EXISTS (SELECT 1 FROM tenant_memberships m WHERE m.user_id = users.id)
    LOOP
        INSERT INTO tenants (name)
        VALUES (COALESCE(NULLIF(existing_user.first_name, ''), 'PartFlow') || ' Store')
        RETURNING id INTO new_tenant_id;

        INSERT INTO tenant_memberships (user_id, tenant_id, membership_role)
        VALUES (existing_user.id, new_tenant_id, 'owner');
    END LOOP;

    FOR business_table IN
        SELECT t.tablename
        FROM pg_tables t
        JOIN pg_class c ON c.relname = t.tablename
        JOIN pg_namespace n ON n.oid = c.relnamespace AND n.nspname = t.schemaname
        WHERE t.schemaname = 'public'
          AND t.tablename NOT IN (
              'users', 'refresh_tokens', 'password_reset_tokens',
              'schema_migrations', 'tenants', 'tenant_memberships',
              'partflow_tenant_schema_migrations'
          )
          -- Public extension tables (for example PostGIS metadata) are not
          -- PartFlow business data and must not be rewritten or tenant-scoped.
          AND NOT EXISTS (
              SELECT 1 FROM pg_depend d
              WHERE d.classid = 'pg_class'::regclass
                AND d.objid = c.oid
                AND d.deptype = 'e'
          )
    LOOP
        EXECUTE format('ALTER TABLE public.%I ADD COLUMN IF NOT EXISTS tenant_id UUID', business_table.tablename);
        IF NOT EXISTS (
            SELECT 1
            FROM pg_attribute a
            JOIN pg_class c ON c.oid = a.attrelid
            JOIN pg_namespace n ON n.oid = c.relnamespace
            WHERE n.nspname = 'public'
              AND c.relname = business_table.tablename
              AND a.attname = 'tenant_id'
              AND a.atttypid = 'uuid'::regtype
              AND NOT a.attisdropped
        ) THEN
            RAISE EXCEPTION 'public.% must have a UUID tenant_id column', business_table.tablename;
        END IF;
        EXECUTE format(
            'UPDATE public.%I SET tenant_id = $1',
            business_table.tablename
        ) USING legacy_tenant_id;
        EXECUTE format('ALTER TABLE public.%I ALTER COLUMN tenant_id SET NOT NULL', business_table.tablename);
        EXECUTE format(
            'CREATE INDEX IF NOT EXISTS %I ON public.%I (tenant_id)',
            left('idx_' || business_table.tablename || '_tenant_id', 63), business_table.tablename
        );
        EXECUTE format('ALTER TABLE public.%I ENABLE ROW LEVEL SECURITY', business_table.tablename);

        -- Policies are permissive by default and PostgreSQL ORs them together.
        -- Remove older policies so none can widen access beyond tenant_id.
        FOR existing_policy IN
            SELECT policyname FROM pg_policies
            WHERE schemaname = 'public' AND tablename = business_table.tablename
        LOOP
            EXECUTE format('DROP POLICY %I ON public.%I', existing_policy.policyname, business_table.tablename);
        END LOOP;
        EXECUTE format(
            'CREATE POLICY partflow_tenant_isolation ON public.%I USING (tenant_id = NULLIF(current_setting(''partflow.tenant_id'', true), '''')::uuid) WITH CHECK (tenant_id = NULLIF(current_setting(''partflow.tenant_id'', true), '''')::uuid)',
            business_table.tablename
        );

        EXECUTE format('ALTER TABLE public.%I FORCE ROW LEVEL SECURITY', business_table.tablename);
        EXECUTE format(
            'CREATE OR REPLACE TRIGGER partflow_assign_tenant_id BEFORE INSERT ON public.%I FOR EACH ROW EXECUTE FUNCTION public.partflow_assign_tenant_id()',
            business_table.tablename
        );

        IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'anon') THEN
            EXECUTE format('REVOKE ALL PRIVILEGES ON TABLE public.%I FROM anon', business_table.tablename);
        END IF;
        IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'authenticated') THEN
            EXECUTE format('REVOKE ALL PRIVILEGES ON TABLE public.%I FROM authenticated', business_table.tablename);
        END IF;
        EXECUTE format('REVOKE ALL PRIVILEGES ON TABLE public.%I FROM PUBLIC', business_table.tablename);

        FOR business_sequence IN
            SELECT seq.relname
            FROM pg_class seq
            JOIN pg_depend d ON d.classid = 'pg_class'::regclass
                AND d.objid = seq.oid
                AND d.refclassid = 'pg_class'::regclass
                AND d.refobjid = to_regclass(format('public.%I', business_table.tablename))
                AND d.deptype = 'a'
            WHERE seq.relnamespace = 'public'::regnamespace AND seq.relkind = 'S'
        LOOP
            EXECUTE format('REVOKE ALL PRIVILEGES ON SEQUENCE public.%I FROM PUBLIC', business_sequence.relname);
            IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'anon') THEN
                EXECUTE format('REVOKE ALL PRIVILEGES ON SEQUENCE public.%I FROM anon', business_sequence.relname);
            END IF;
            IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'authenticated') THEN
                EXECUTE format('REVOKE ALL PRIVILEGES ON SEQUENCE public.%I FROM authenticated', business_sequence.relname);
            END IF;
        END LOOP;
    END LOOP;

    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'anon') THEN
        REVOKE ALL PRIVILEGES ON TABLE public.tenants, public.tenant_memberships FROM anon;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'authenticated') THEN
        REVOKE ALL PRIVILEGES ON TABLE public.tenants, public.tenant_memberships FROM authenticated;
        REVOKE ALL PRIVILEGES ON TABLE public.partflow_tenant_schema_migrations FROM authenticated;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'anon') THEN
        REVOKE ALL PRIVILEGES ON TABLE public.partflow_tenant_schema_migrations FROM anon;
    END IF;
END $$;

-- Scope membership lookups to the account authenticated by the API. Place this
-- after legacy membership creation so migration-time owner operations cannot
-- be blocked by the policy.
ALTER TABLE public.tenant_memberships ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS partflow_membership_self ON public.tenant_memberships;
CREATE POLICY partflow_membership_self ON public.tenant_memberships
    FOR SELECT
    USING (user_id = NULLIF(current_setting('partflow.user_id', true), '')::uuid);

CREATE OR REPLACE FUNCTION public.partflow_create_tenant_for_user()
RETURNS TRIGGER
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
DECLARE
    new_tenant_id UUID;
BEGIN
    INSERT INTO public.tenants (name)
    VALUES (COALESCE(NULLIF(NEW.first_name, ''), 'PartFlow') || ' Store')
    RETURNING id INTO new_tenant_id;

    INSERT INTO public.tenant_memberships (user_id, tenant_id, membership_role)
    VALUES (NEW.id, new_tenant_id, 'owner');
    RETURN NEW;
END;
$$;

REVOKE ALL ON FUNCTION public.partflow_create_tenant_for_user() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.partflow_create_tenant_for_user() TO partflow_runtime;

DROP TRIGGER IF EXISTS partflow_create_tenant_for_user ON users;
CREATE TRIGGER partflow_create_tenant_for_user
AFTER INSERT ON users
FOR EACH ROW EXECUTE FUNCTION public.partflow_create_tenant_for_user();

INSERT INTO public.partflow_tenant_schema_migrations (version) VALUES (80);

COMMIT;
