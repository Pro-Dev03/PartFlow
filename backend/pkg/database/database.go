package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/pkg/config"
	_ "modernc.org/sqlite"
)

var DB *sqlx.DB

const tenantRuntimeRole = "partflow_runtime"

// TenantIsolationEnabled is deliberately explicit: the migration and runtime
// role must be installed before a deployment can turn on tenant-scoped RLS.
func TenantIsolationEnabled() bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv("PARTFLOW_TENANT_RLS_ENABLED")))
	return value == "1" || value == "true" || value == "yes"
}

// Initialize initializes the database connection
func Initialize() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	databaseURL := cfg.DatabaseURL
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	_ = os.Setenv("DATABASE_URL", databaseURL)

	var driver string
	var connectURL string
	if strings.HasPrefix(databaseURL, "sqlite://") {
		driver = "sqlite"
		connectURL = strings.TrimPrefix(databaseURL, "sqlite://")
		if connectURL == "" {
			return fmt.Errorf("sqlite database path is empty")
		}
		if !strings.Contains(connectURL, "?_pragma=") {
			connectURL += "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)"
		}
		log.Println("Using SQLite database for local mode:", connectURL)
	} else {
		driver = "pgx"
		connectURL = databaseURL
		if !strings.Contains(connectURL, "sslmode") {
			connectURL += "?sslmode=require"
		}
		log.Println("Using PostgreSQL database for cloud mode")
	}

	if driver == "pgx" {
		pgxConfig, parseErr := pgx.ParseConfig(connectURL)
		if parseErr != nil {
			return fmt.Errorf("failed to parse database URL: %w", parseErr)
		}
		if pgxConfig.RuntimeParams == nil {
			pgxConfig.RuntimeParams = make(map[string]string)
		}
		pgxConfig.RuntimeParams["timezone"] = "UTC"
		pgxConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
		if TenantIsolationEnabled() {
			// Keep the database owner credentials server-side but execute all
			// application SQL as a non-owner role that is subject to RLS.
			pgxConfig.RuntimeParams["role"] = tenantRuntimeRole
			// Never resolve an unqualified application table from a role-named or
			// attacker-created schema before the trusted public schema.
			pgxConfig.RuntimeParams["search_path"] = "pg_catalog,public"
		}
		DB = sqlx.NewDb(stdlib.OpenDB(*pgxConfig), driver)
		err = DB.Ping()
	} else {
		DB, err = sqlx.Connect(driver, connectURL)
	}
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	if TenantIsolationEnabled() {
		// Existing handlers share a database handle and do not all accept a
		// transaction. TenantScope serializes protected requests and pins the
		// session tenant through this single connection until every handler has
		// been migrated to transaction-aware repositories.
		DB.SetMaxOpenConns(1)
		DB.SetMaxIdleConns(1)
	} else {
		DB.SetMaxOpenConns(25)
		DB.SetMaxIdleConns(10)
	}
	DB.SetConnMaxLifetime(5 * time.Minute)
	DB.SetConnMaxIdleTime(1 * time.Minute)

	// Test connection
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	if driver == "pgx" && !TenantIsolationEnabled() {
		var tenantRuntimeRoleExists bool
		if err := DB.Get(&tenantRuntimeRoleExists,
			`SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = $1)`, tenantRuntimeRole); err != nil {
			return fmt.Errorf("verify tenant runtime configuration: %w", err)
		}
		if tenantRuntimeRoleExists {
			return fmt.Errorf("tenant RLS role exists; set PARTFLOW_TENANT_RLS_ENABLED=true on every service using this database")
		}
	}

	if TenantIsolationEnabled() {
		if driver != "pgx" {
			return fmt.Errorf("tenant RLS requires PostgreSQL; refusing to start with %s", driver)
		}
		if err := validateTenantIsolationSchema(DB); err != nil {
			return fmt.Errorf("tenant RLS is enabled but the database is not ready: %w", err)
		}
	} else if err := ensureRequiredSchema(DB); err != nil {
		return fmt.Errorf("failed to ensure required schema: %w", err)
	}

	log.Println("Database connection established successfully using", driver)
	return nil
}

func validateTenantIsolationSchema(db *sqlx.DB) error {
	var roleName string
	var isSuperuser, bypassRLS, canLogin, canInherit, hasRoleMembership bool
	var runtimeSessionSafe bool
	if err := db.QueryRowx(`
		SELECT current_user, runtime.rolsuper, runtime.rolbypassrls, runtime.rolcanlogin, runtime.rolinherit,
		       EXISTS (SELECT 1 FROM pg_auth_members m WHERE m.member = runtime.oid),
	       session_login.rolcanlogin
	           AND NOT session_login.rolsuper
	           AND NOT session_login.rolbypassrls
	           AND NOT session_login.rolinherit
	           AND NOT session_login.rolcreatedb
			AND NOT session_login.rolcreaterole
			AND NOT session_login.rolreplication
			AND db_info.datdba <> session_login.oid
			AND NOT has_database_privilege(session_login.oid, current_database(), 'CREATE')
			AND NOT EXISTS (
			    SELECT 1 FROM pg_namespace owned_schema
			    WHERE owned_schema.nspname = 'public' AND owned_schema.nspowner = session_login.oid
			)
			AND NOT EXISTS (
			    SELECT 1 FROM pg_class owned_object
			    JOIN pg_namespace owned_schema ON owned_schema.oid = owned_object.relnamespace
			    WHERE owned_schema.nspname = 'public' AND owned_object.relowner = session_login.oid
			)
			AND NOT EXISTS (
			    SELECT 1 FROM pg_proc owned_function
			    JOIN pg_namespace owned_schema ON owned_schema.oid = owned_function.pronamespace
			    WHERE owned_schema.nspname = 'public' AND owned_function.proowner = session_login.oid
			)
			AND EXISTS (
	               SELECT 1 FROM pg_auth_members m
	               WHERE m.member = session_login.oid AND m.roleid = runtime.oid AND NOT m.admin_option
	           )
	           AND NOT EXISTS (
	               SELECT 1 FROM pg_auth_members m
	               WHERE m.member = session_login.oid AND m.roleid <> runtime.oid
	           )
		FROM pg_roles runtime
		JOIN pg_roles session_login ON session_login.rolname = session_user
		JOIN pg_database db_info ON db_info.datname = current_database()
		WHERE runtime.rolname = current_user
	`).Scan(&roleName, &isSuperuser, &bypassRLS, &canLogin, &canInherit, &hasRoleMembership, &runtimeSessionSafe); err != nil {
		return fmt.Errorf("verify runtime database role: %w", err)
	}
	if roleName != tenantRuntimeRole || isSuperuser || bypassRLS || canLogin || canInherit || hasRoleMembership {
		return fmt.Errorf("runtime database role must be %s with NOLOGIN, NOINHERIT, NOSUPERUSER, NOBYPASSRLS, and no role memberships (current=%s, login=%t, inherit=%t, superuser=%t, bypassrls=%t, memberships=%t)", tenantRuntimeRole, roleName, canLogin, canInherit, isSuperuser, bypassRLS, hasRoleMembership)
	}
	if !runtimeSessionSafe {
		return fmt.Errorf("database session login must be a dedicated NOINHERIT, non-owner, non-superuser login whose only role membership is %s", tenantRuntimeRole)
	}
	var tenantMigrationApplied bool
	if err := db.Get(&tenantMigrationApplied, `
		SELECT EXISTS (
			SELECT 1 FROM public.partflow_tenant_schema_migrations WHERE version = 80
		)
	`); err != nil {
		return fmt.Errorf("verify tenant isolation migration marker: %w", err)
	}
	if !tenantMigrationApplied {
		return fmt.Errorf("tenant isolation migration 080 is not recorded as applied")
	}
	var membershipPolicyReady bool
	if err := db.Get(&membershipPolicyReady, `
		SELECT c.relrowsecurity
		   AND EXISTS (
		       SELECT 1 FROM pg_policies p
		       WHERE p.schemaname = 'public'
		         AND p.tablename = 'tenant_memberships'
		         AND p.policyname = 'partflow_membership_self'
		         AND p.cmd = 'SELECT'
		         AND position('partflow.user_id' in p.qual) > 0
		         AND position('user_id' in p.qual) > 0
		   )
		   AND NOT EXISTS (
		       SELECT 1 FROM pg_policies p
		       WHERE p.schemaname = 'public'
		         AND p.tablename = 'tenant_memberships'
		         AND p.policyname <> 'partflow_membership_self'
	   )
		   AND NOT has_table_privilege(current_user, 'public.tenant_memberships', 'INSERT')
		   AND NOT has_table_privilege(current_user, 'public.tenant_memberships', 'UPDATE')
		   AND NOT has_table_privilege(current_user, 'public.tenant_memberships', 'DELETE')
		   AND NOT has_table_privilege(current_user, 'public.tenants', 'SELECT')
			   AND NOT has_table_privilege(current_user, 'public.tenants', 'INSERT')
			   AND NOT has_table_privilege(current_user, 'public.tenants', 'UPDATE')
		   AND NOT has_table_privilege(current_user, 'public.tenants', 'DELETE')
		   AND has_schema_privilege(current_user, 'public', 'USAGE')
		   AND NOT has_schema_privilege(current_user, 'public', 'CREATE')
		   AND NOT EXISTS (
		       SELECT 1
		       FROM pg_namespace protected_schema
		       CROSS JOIN LATERAL aclexplode(coalesce(protected_schema.nspacl, acldefault('n', protected_schema.nspowner))) acl
		       LEFT JOIN pg_roles grantee ON grantee.oid = acl.grantee
		       WHERE protected_schema.nspname = 'public'
		         AND (acl.grantee = 0 OR grantee.rolname IN ('anon','authenticated'))
		         AND acl.privilege_type IN ('USAGE','CREATE')
		   )
		   AND NOT EXISTS (
		       SELECT 1
		       FROM pg_class protected_table
		       JOIN pg_namespace protected_schema ON protected_schema.oid = protected_table.relnamespace
		       CROSS JOIN LATERAL aclexplode(coalesce(protected_table.relacl, acldefault('r', protected_table.relowner))) acl
		       LEFT JOIN pg_roles grantee ON grantee.oid = acl.grantee
		       WHERE protected_schema.nspname = 'public'
		         AND protected_table.relname IN ('tenants','tenant_memberships')
		         AND (acl.grantee = 0 OR grantee.rolname IN ('anon','authenticated'))
		         AND acl.privilege_type IN ('SELECT','INSERT','UPDATE','DELETE','TRUNCATE','REFERENCES','TRIGGER')
		   )
		   AND NOT EXISTS (
		       SELECT 1 FROM aclexplode(coalesce(c.relacl, acldefault('r', c.relowner))) acl
		       WHERE acl.grantee = 0
		         AND acl.privilege_type IN ('SELECT','INSERT','UPDATE','DELETE','TRUNCATE','REFERENCES','TRIGGER')
		   )
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = 'public' AND c.relname = 'tenant_memberships'
	`); err != nil {
		return fmt.Errorf("verify tenant membership policy: %w", err)
	}
	if !membershipPolicyReady {
		return fmt.Errorf("tenant_memberships must be read-only to the runtime role and protected by the self-membership policy")
	}

	var unsafeTable string
	err := db.Get(&unsafeTable, `
		SELECT c.relname
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = 'public'
		  AND c.relkind IN ('r', 'p')
		  AND NOT EXISTS (
		      SELECT 1 FROM pg_depend d
		      WHERE d.classid = 'pg_class'::regclass
		        AND d.objid = c.oid
		        AND d.deptype = 'e'
		  )
		  AND c.relname NOT IN (
		      'users', 'refresh_tokens', 'password_reset_tokens',
		      'schema_migrations', 'tenants', 'tenant_memberships',
		      'partflow_tenant_schema_migrations'
		  )
		  AND (
		      NOT EXISTS (
		          SELECT 1 FROM pg_attribute a
		          WHERE a.attrelid = c.oid AND a.attname = 'tenant_id' AND NOT a.attisdropped
		      )
		      OR NOT c.relrowsecurity
		      OR NOT c.relforcerowsecurity
		      OR NOT EXISTS (
		          SELECT 1 FROM pg_policies p
		          WHERE p.schemaname = 'public'
		            AND p.tablename = c.relname
		            AND p.policyname = 'partflow_tenant_isolation'
		            AND p.cmd = 'ALL'
		            AND position('tenant_id' in p.qual) > 0
		            AND position('partflow.tenant_id' in p.qual) > 0
		            AND position('tenant_id' in p.with_check) > 0
		            AND position('partflow.tenant_id' in p.with_check) > 0
		      )
		      OR EXISTS (
		          SELECT 1 FROM pg_policies p
		          WHERE p.schemaname = 'public'
		            AND p.tablename = c.relname
		            AND p.policyname <> 'partflow_tenant_isolation'
		      )
		      OR EXISTS (
		          SELECT 1 FROM aclexplode(coalesce(c.relacl, acldefault('r', c.relowner))) acl
		          JOIN pg_roles grantee ON grantee.oid = acl.grantee
		          WHERE grantee.rolname IN ('anon','authenticated')
		            AND acl.privilege_type IN ('SELECT','INSERT','UPDATE','DELETE','TRUNCATE','REFERENCES','TRIGGER')
		      )
		      OR EXISTS (
		          SELECT 1 FROM aclexplode(coalesce(c.relacl, acldefault('r', c.relowner))) acl
		          WHERE acl.grantee = 0
		            AND acl.privilege_type IN ('SELECT','INSERT','UPDATE','DELETE','TRUNCATE','REFERENCES','TRIGGER')
		      )
		      OR EXISTS (
		          SELECT 1 FROM pg_roles r
		          WHERE r.rolname IN ('anon','authenticated')
		            AND (
		                has_table_privilege(r.oid, c.oid, 'SELECT')
		                OR has_table_privilege(r.oid, c.oid, 'INSERT')
		                OR has_table_privilege(r.oid, c.oid, 'UPDATE')
		                OR has_table_privilege(r.oid, c.oid, 'DELETE')
		                OR has_table_privilege(r.oid, c.oid, 'TRUNCATE')
		                OR has_table_privilege(r.oid, c.oid, 'REFERENCES')
		                OR has_table_privilege(r.oid, c.oid, 'TRIGGER')
		            )
		      )
		  )
		ORDER BY c.relname
		LIMIT 1
	`)
	if err == nil {
		return fmt.Errorf("table public.%s is missing tenant_id, forced RLS, or the PartFlow tenant policy", unsafeTable)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("verify tenant policies: %w", err)
	}

	var unsafeView string
	err = db.Get(&unsafeView, `
		SELECT c.relname
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = 'public'
		  AND c.relkind IN ('v', 'm')
		  AND NOT EXISTS (
		      SELECT 1 FROM pg_depend d
		      WHERE d.classid = 'pg_class'::regclass
		        AND d.objid = c.oid
		        AND d.deptype = 'e'
		  )
		  AND (
		      c.relkind = 'm'
		      OR NOT coalesce(c.reloptions @> ARRAY['security_invoker=true'], false)
		      OR NOT has_table_privilege(current_user, c.oid, 'SELECT')
		      OR EXISTS (
		          SELECT 1 FROM aclexplode(coalesce(c.relacl, acldefault('r', c.relowner))) acl
		          JOIN pg_roles grantee ON grantee.oid = acl.grantee
		          WHERE grantee.rolname IN ('anon','authenticated')
		            AND acl.privilege_type IN ('SELECT','INSERT','UPDATE','DELETE','TRUNCATE','REFERENCES','TRIGGER')
		      )
		      OR EXISTS (
		          SELECT 1 FROM aclexplode(coalesce(c.relacl, acldefault('r', c.relowner))) acl
		          WHERE acl.grantee = 0
		            AND acl.privilege_type IN ('SELECT','INSERT','UPDATE','DELETE','TRUNCATE','REFERENCES','TRIGGER')
		      )
		  )
		ORDER BY c.relname
		LIMIT 1
	`)
	if err == nil {
		return fmt.Errorf("view public.%s is not safe for tenant-scoped access", unsafeView)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("verify tenant views: %w", err)
	}
	return nil
}

func ensureRequiredSchema(db *sqlx.DB) error {
	if db == nil {
		return fmt.Errorf("database is nil")
	}

	if !strings.EqualFold(db.DriverName(), "sqlite") {
		if _, err := db.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto;`); err != nil {
			return err
		}
		if _, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS part_types (
				id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
				name_ar VARCHAR(255) NOT NULL,
				name_en VARCHAR(255) NOT NULL,
				icon VARCHAR(100),
				color VARCHAR(7),
				is_active BOOLEAN DEFAULT true,
				sort_order INTEGER DEFAULT 0,
				created_at TIMESTAMPTZ DEFAULT NOW(),
				updated_at TIMESTAMPTZ DEFAULT NOW()
			);
		`); err != nil {
			return err
		}

		if _, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS held_sales (
				id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
				user_id UUID NOT NULL,
				items JSONB NOT NULL,
				created_at TIMESTAMPTZ DEFAULT NOW()
			);
		`); err != nil {
			return err
		}
		if ok, err := tableExists(db, "inventory_items"); err != nil {
			return err
		} else if ok {
			if _, err := db.Exec(`ALTER TABLE inventory_items ADD COLUMN IF NOT EXISTS part_type_id UUID REFERENCES part_types(id) ON DELETE SET NULL;`); err != nil {
				return err
			}
		}
		return nil
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS part_types (
			id TEXT PRIMARY KEY,
			name_ar TEXT NOT NULL,
			name_en TEXT NOT NULL,
			icon TEXT,
			color TEXT,
			is_active INTEGER DEFAULT 1,
			sort_order INTEGER DEFAULT 0,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT DEFAULT CURRENT_TIMESTAMP
		);
	`); err != nil {
		return err
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS settings (
			id TEXT PRIMARY KEY,
			key TEXT NOT NULL UNIQUE,
			value TEXT,
			value_type TEXT DEFAULT 'string',
			category TEXT DEFAULT 'general',
			description TEXT,
			is_public INTEGER DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`); err != nil {
		return err
	}

	if _, err := db.Exec(`
		INSERT OR IGNORE INTO settings (id, key, value, value_type, category, description, is_public, created_at, updated_at)
		VALUES
			('setting-tax-rate', 'tax_rate', '0', 'number', 'financial', 'Tax rate', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
			('setting-store-name', 'store_name', 'PartFlow Store', 'string', 'general', 'Store name', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
			('setting-discounts-enabled', 'discounts_enabled', 'true', 'boolean', 'financial', 'Allow discounts', 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
			('setting-currency', 'currency', 'ILS', 'string', 'general', 'Currency', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
			('setting-default-profit-margin', 'default_profit_margin', '30', 'number', 'financial', 'Default profit margin percentage', 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`); err != nil {
		return err
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS held_sales (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			items TEXT NOT NULL,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		);
	`); err != nil {
		return err
	}

	// localdb.Open applies the complete local schema later, but this table is
	// required by startup checks that run before localdb.Open.
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS inventory_items (
			id TEXT PRIMARY KEY,
			product_id TEXT NOT NULL,
			item_code TEXT NOT NULL UNIQUE,
			barcode TEXT,
			serial_number TEXT,
			condition TEXT,
			grade TEXT,
			purchase_cost REAL DEFAULT 0,
			selling_price REAL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'AVAILABLE',
			location_id TEXT,
			supplier_id TEXT,
			customer_id TEXT,
			purchase_date TEXT,
			sold_at TEXT,
			notes TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
	`); err != nil {
		return err
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS audit_logs (
			id TEXT PRIMARY KEY,
			user_id TEXT,
			action TEXT NOT NULL,
			entity_type TEXT NOT NULL,
			entity_id TEXT NOT NULL,
			ip_address TEXT,
			user_agent TEXT,
			request_id TEXT,
			changes TEXT,
			description TEXT,
			status TEXT DEFAULT 'success',
			error_message TEXT,
			metadata TEXT DEFAULT '{}',
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`); err != nil {
		return err
	}
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS item_history (
			id TEXT PRIMARY KEY,
			inventory_item_id TEXT NOT NULL,
			event_type TEXT NOT NULL,
			event_date TEXT NOT NULL,
			reference_type TEXT,
			reference_id TEXT,
			description TEXT,
			metadata TEXT DEFAULT '{}',
			created_by TEXT,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`); err != nil {
		return err
	}
	if _, err := db.Exec(`
		UPDATE audit_logs
		SET created_at = substr(created_at, 1, instr(created_at, ' m=') - 1)
		WHERE instr(created_at, ' m=') > 0
	`); err != nil {
		return err
	}

	if ok, err := columnExists(db, "inventory_items", "part_type_id"); err != nil {
		return err
	} else if !ok {
		if _, err := db.Exec(`ALTER TABLE inventory_items ADD COLUMN part_type_id TEXT;`); err != nil {
			return err
		}
	}

	return nil
}

func columnExists(db *sqlx.DB, tableName, columnName string) (bool, error) {
	if strings.EqualFold(db.DriverName(), "postgres") || strings.EqualFold(db.DriverName(), "pgx") {
		var exists bool
		query := `
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_name = $1 AND column_name = $2
			)
		`
		if err := db.Get(&exists, query, tableName, columnName); err != nil {
			return false, err
		}
		return exists, nil
	}

	var exists bool
	query := `SELECT EXISTS (
		SELECT 1 FROM PRAGMA_TABLE_INFO(?) WHERE name = ?
	)`
	if err := db.Get(&exists, query, tableName, columnName); err != nil {
		return false, err
	}
	return exists, nil
}

func tableExists(db *sqlx.DB, tableName string) (bool, error) {
	if strings.EqualFold(db.DriverName(), "postgres") || strings.EqualFold(db.DriverName(), "pgx") {
		var exists bool
		query := `SELECT EXISTS (
			SELECT 1 FROM information_schema.tables WHERE table_name = $1
		)`
		if err := db.Get(&exists, query, tableName); err != nil {
			return false, err
		}
		return exists, nil
	}

	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = ?)`
	if err := db.Get(&exists, query, tableName); err != nil {
		return false, err
	}
	return exists, nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// GetDB returns the database instance
func GetDB() *sqlx.DB {
	return DB
}

// Health checks the database connection
func Health() error {
	if DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}
	return DB.Ping()
}
