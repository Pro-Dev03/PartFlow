-- ============================================
-- Remove Roles and Permissions System Completely
-- ============================================
-- This migration removes all role-based access control
-- System becomes a simple single-user system

-- Remove RLS policies that use role checks
DROP POLICY IF EXISTS "Owners can view organization users" ON users;
DROP POLICY IF EXISTS "Owners can manage organization users" ON users;

-- Remove role column from users table
ALTER TABLE users DROP COLUMN IF EXISTS role;

-- Remove role index
DROP INDEX IF EXISTS idx_users_role;

-- Drop any remaining role-related constraints
ALTER TABLE users DROP CONSTRAINT IF EXISTS check_role_is_owner;

-- ============================================
-- Update authentication to work without roles
-- ============================================
-- All users will have full access (admin by default)
-- No role-based restrictions

-- Verify cleanup
SELECT 'Roles and permissions system completely removed' as status;