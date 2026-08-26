-- ============================================
-- Fix purchases table to make user_id nullable
-- ============================================
-- This allows purchases to be created without requiring a valid user_id
-- Useful for testing and API compatibility

ALTER TABLE purchases 
ALTER COLUMN user_id DROP NOT NULL;

SELECT 'Purchases user_id made nullable' as status;