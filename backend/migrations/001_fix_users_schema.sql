-- Fix users table schema to match auth model
-- This migration ensures the users table has the correct structure

-- Add role column if it doesn't exist (text-based role for simpler auth)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'role'
    ) THEN
        ALTER TABLE users ADD COLUMN role VARCHAR(50) DEFAULT 'owner';
    END IF;
END $$;

-- Update existing users to have role based on their role_id
UPDATE users u
SET role = COALESCE(
    (SELECT name FROM roles r WHERE r.id = u.role_id),
    'owner'
)
WHERE role IS NULL OR role = '';

-- Make sure subscription fields have proper defaults
ALTER TABLE users 
ALTER COLUMN subscription_status SET DEFAULT 'active',
ALTER COLUMN is_active SET DEFAULT true;

-- Add indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_subscription ON users(subscription_status, subscription_expires_at);

-- Add comment for documentation
COMMENT ON COLUMN users.role IS 'User role: owner, admin, staff, etc. (simpler than role_id for auth)';