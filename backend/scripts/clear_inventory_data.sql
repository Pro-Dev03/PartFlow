-- Script to clear inventory data (especially used parts/trade-ins)
-- This will remove all inventory items and related data
-- Use this if you want to start with a clean inventory

-- WARNING: This will delete ALL inventory data!
-- Make sure to backup your database before running this script

BEGIN;

-- 1. Delete inventory movements first (they reference items)
DELETE FROM inventory_movements;

-- 2. Delete reservations (they reference items)
DELETE FROM reservations;

-- 3. Delete item specification values
DELETE FROM item_specification_values;

-- 4. Delete trade-ins records
DELETE FROM trade_ins;

-- 5. Delete all inventory items
DELETE FROM inventory_items;

-- 6. Delete the old inventory table records (if exists)
DELETE FROM inventory;

COMMIT;

-- Verify the cleanup
SELECT 
    'inventory_items' as table_name,
    COUNT(*) as remaining_count
FROM inventory_items
UNION ALL
SELECT 
    'inventory_movements' as table_name,
    COUNT(*) as remaining_count
FROM inventory_movements
UNION ALL
SELECT 
    'reservations' as table_name,
    COUNT(*) as remaining_count
FROM reservations
UNION ALL
SELECT 
    'trade_ins' as table_name,
    COUNT(*) as remaining_count
FROM trade_ins
UNION ALL
SELECT 
    'inventory' as table_name,
    COUNT(*) as remaining_count
FROM inventory;

-- Display completion message
DO $$
BEGIN
    RAISE NOTICE '✅ Inventory data cleared successfully!';
    RAISE NOTICE 'All inventory items, movements, reservations, and trade-ins have been deleted.';
    RAISE NOTICE 'You can now start fresh with your inventory system.';
END $$;