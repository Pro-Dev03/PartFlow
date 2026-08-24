-- PartFlow Simplified Permissions Schema
-- Use simple "owner" role only - no complex RBAC

-- ============================================
-- Simplify users table (if not already done)
-- ============================================

-- Add simple role column if not exists
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) DEFAULT 'owner' CHECK (role = 'owner');

-- Update existing users to have owner role
UPDATE users SET role = 'owner' WHERE role IS NULL;

-- Remove any old role_id column if it exists
ALTER TABLE users DROP COLUMN IF EXISTS role_id;
ALTER TABLE users DROP COLUMN IF EXISTS permissions_jsonb;

-- ============================================
-- Create index on role for performance
-- ============================================
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- ============================================
-- Note: Complex RBAC tables (roles, permissions, role_permissions) 
-- will be dropped in migration 015 if they still exist
-- ============================================
