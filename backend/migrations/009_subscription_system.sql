-- =====================================================
-- Migration 009: Subscription System (Admin Only)
-- Adds subscription support to users table based on worktrack system
-- Simplified for admin login only (no employee features)
-- =====================================================

-- Add subscription columns to users table
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS subscription_status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (subscription_status IN ('active', 'expired', 'canceled'));

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS subscription_expires_at TIMESTAMPTZ NULL;

-- Update existing users to have active subscriptions
UPDATE users 
SET subscription_status = 'active',
    subscription_expires_at = NOW() + INTERVAL '1 year'
WHERE subscription_expires_at IS NULL AND subscription_status = 'active';

-- Create indexes for subscription management
CREATE INDEX IF NOT EXISTS idx_users_subscription_status ON users(subscription_status);
CREATE INDEX IF NOT EXISTS idx_users_subscription_expires_at ON users(subscription_expires_at);

-- =====================================================
-- Insert default admin users with subscriptions
-- =====================================================
-- Note: These are default admin users for testing purposes
-- Passwords should be changed in production
INSERT INTO users (
  id,
  organization_id,
  full_name,
  email,
  password_hash,
  role_id,
  is_active,
  subscription_status,
  subscription_expires_at,
  created_at,
  updated_at
)
SELECT
  gen_random_uuid(),
  (SELECT id FROM organizations LIMIT 1),
  'System Admin',
  'admin@partflow.com',
  crypt('admin123', gen_salt('bf', 12)),  -- Password: admin123
  (SELECT id FROM roles WHERE name = 'admin' LIMIT 1),
  TRUE,
  'active',
  NOW() + INTERVAL '1 year',  -- 1 year subscription
  NOW(),
  NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM users WHERE email = 'admin@partflow.com'
);

-- Insert another admin with lifetime subscription
INSERT INTO users (
  id,
  organization_id,
  full_name,
  email,
  password_hash,
  role_id,
  is_active,
  subscription_status,
  subscription_expires_at,
  created_at,
  updated_at
)
SELECT
  gen_random_uuid(),
  (SELECT id FROM organizations LIMIT 1),
  'Super Admin',
  'superadmin@partflow.com',
  crypt('superadmin123', gen_salt('bf', 12)),  -- Password: superadmin123
  (SELECT id FROM roles WHERE name = 'admin' LIMIT 1),
  TRUE,
  'active',
  NULL,  -- Lifetime subscription
  NOW(),
  NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM users WHERE email = 'superadmin@partflow.com'
);

-- Verify the users
SELECT 
  id,
  full_name,
  email,
  subscription_status,
  subscription_expires_at,
  CASE
    WHEN subscription_expires_at IS NULL THEN 'Lifetime'
    WHEN subscription_expires_at > NOW() THEN 'Active'
    ELSE 'Expired'
  END as current_status
FROM users
WHERE email IN ('admin@partflow.com', 'superadmin@partflow.com');