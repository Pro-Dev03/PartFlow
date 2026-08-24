-- ============================================
-- Remove Warranties Feature
-- ============================================
-- This migration removes all warranty-related tables and columns

-- Drop warranty_claims table
DROP TABLE IF EXISTS warranty_claims CASCADE;

-- Drop warranties table
DROP TABLE IF EXISTS warranties CASCADE;

-- Remove warranty policy column from products table
ALTER TABLE products DROP COLUMN IF EXISTS warranty_policy;

-- Remove warranty_period column from sales table (if exists)
ALTER TABLE sales DROP COLUMN IF EXISTS warranty_period;

-- Remove expires_at column from sales table (if exists)
ALTER TABLE sales DROP COLUMN IF EXISTS expires_at;

-- Remove warranty-related permissions from permissions table
DELETE FROM permissions WHERE resource = 'warranties';

-- Remove warranty-related notification preferences columns
ALTER TABLE notification_preferences DROP COLUMN IF EXISTS warranty_expiring;

-- Update audit log to remove warranty entity types
DELETE FROM audit_log WHERE entity_type = 'warranty';
DELETE FROM audit_log WHERE entity_type = 'warranty_claim';
