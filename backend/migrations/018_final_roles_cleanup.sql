-- ============================================
-- Final Cleanup: Remove All Role References
-- ============================================
-- This migration ensures complete removal of role/permission system
-- Some references might still exist in user table structure

-- Check and remove any remaining role columns from users table
DO $$
BEGIN
    -- Remove role column if exists
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'role'
    ) THEN
        ALTER TABLE users DROP COLUMN role;
        RAISE NOTICE 'Dropped role column from users table';
    END IF;

    -- Remove role_id column if exists
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'role_id'
    ) THEN
        ALTER TABLE users DROP COLUMN role_id;
        RAISE NOTICE 'Dropped role_id column from users table';
    END IF;

    -- Remove permissions_jsonb column if exists
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'permissions_jsonb'
    ) THEN
        ALTER TABLE users DROP COLUMN permissions_jsonb;
        RAISE NOTICE 'Dropped permissions_jsonb column from users table';
    END IF;
END $$;

-- Drop any remaining role-related indexes
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_users_role_id;

-- Drop any remaining role-related constraints
ALTER TABLE users DROP CONSTRAINT IF EXISTS check_role_is_owner;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;

-- Remove any remaining RLS policies that reference roles
DROP POLICY IF EXISTS "Owners can view organization users" ON users;
DROP POLICY IF EXISTS "Owners can manage organization users" ON users;
DROP POLICY IF EXISTS "Users can view organization users" ON users;
DROP POLICY IF EXISTS "Users can manage organization users" ON users;

-- Create simple policies without role checks (all authenticated users have full access)
DROP POLICY IF EXISTS "users_select_policy" ON users;
DROP POLICY IF EXISTS "users_insert_policy" ON users;
DROP POLICY IF EXISTS "users_update_policy" ON users;
DROP POLICY IF EXISTS "users_delete_policy" ON users;

-- Note: RLS policies should be recreated in a separate migration if needed
-- For now, we're removing role-based restrictions

-- Verify cleanup
SELECT 
    'Final roles and permissions cleanup completed' as status,
    (SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'users' AND column_name LIKE '%role%') as remaining_role_columns,
    (SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'users' AND column_name LIKE '%permission%') as remaining_permission_columns;