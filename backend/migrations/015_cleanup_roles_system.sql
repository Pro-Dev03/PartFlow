-- ============================================
-- Cleanup Roles and Permissions System
-- ============================================
-- This migration removes the complex RBAC system completely
-- Since we now use only the simple "owner" role in the users table

-- Drop triggers related to roles
DROP TRIGGER IF EXISTS sync_role_permissions_insert ON role_permissions;
DROP TRIGGER IF EXISTS sync_role_permissions_delete ON role_permissions;

-- Drop function
DROP FUNCTION IF EXISTS sync_role_permissions();

-- Drop tables (in correct order due to dependencies)
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;

-- Remove any remaining role-related columns from users table
-- (This should already be done in migration 010, but ensuring it's complete)
ALTER TABLE users DROP COLUMN IF EXISTS role_id;
ALTER TABLE users DROP COLUMN IF EXISTS permissions_jsonb;

-- Add simple role column if not exists (for safety)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'role'
    ) THEN
        ALTER TABLE users ADD COLUMN role VARCHAR(20) DEFAULT 'owner' CHECK (role = 'owner');
    END IF;
END $$;

-- Update any users without a role to be owners
UPDATE users SET role = 'owner' WHERE role IS NULL;

-- Add constraint to ensure only owner role is allowed
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE table_name = 'users' AND constraint_name = 'check_role_is_owner'
    ) THEN
        ALTER TABLE users ADD CONSTRAINT check_role_is_owner CHECK (role = 'owner');
    END IF;
END $$;

-- Create index on role for performance
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- ============================================
-- RLS Policies removed - system no longer uses RLS
-- ============================================

-- Verify cleanup
SELECT 'Roles and permissions system cleanup completed' as status;
