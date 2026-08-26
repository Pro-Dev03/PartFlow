-- ============================================
-- Fix Critical Issues for PartFlow Testing
-- ============================================

-- 1. Fix trade_ins table column name mismatch
-- The code uses "item_id" but the schema has "inventory_item_id"
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'trade_ins' AND column_name = 'inventory_item_id'
    ) THEN
        ALTER TABLE trade_ins RENAME COLUMN inventory_item_id TO item_id;
        RAISE NOTICE 'Renamed inventory_item_id to item_id in trade_ins table';
    END IF;
END $$;

-- 2. Make sales user_id nullable for API compatibility
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'sales' AND column_name = 'user_id' AND is_nullable = 'NO'
    ) THEN
        ALTER TABLE sales ALTER COLUMN user_id DROP NOT NULL;
        RAISE NOTICE 'Made sales.user_id nullable';
    END IF;
END $$;

-- 3. Fix inventory_items location_id constraint
-- First check if locations table has proper data
DO $$
BEGIN
    -- Ensure main warehouse location exists
    IF NOT EXISTS (SELECT 1 FROM locations WHERE id = '9b561ebf-0382-42c5-82e6-61d9627e918d') THEN
        INSERT INTO locations (id, name, type, is_active, created_at, updated_at)
        VALUES ('9b561ebf-0382-42c5-82e6-61d9627e918d', 'المخزن الرئيسي', 'warehouse', true, NOW(), NOW())
        ON CONFLICT (id) DO NOTHING;
        RAISE NOTICE 'Ensured main warehouse location exists';
    END IF;
END $$;

-- 4. Add proper barcode uniqueness constraint
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE table_name = 'inventory_items' AND constraint_name = 'inventory_items_barcode_key'
    ) THEN
        -- Drop existing barcode constraint if exists
        ALTER TABLE inventory_items DROP CONSTRAINT IF EXISTS inventory_items_barcode_key;
        
        -- Add proper unique constraint
        ALTER TABLE inventory_items ADD CONSTRAINT inventory_items_barcode_key UNIQUE (barcode);
        RAISE NOTICE 'Added proper barcode uniqueness constraint';
    END IF;
END $$;

-- 5. Add purchase_cost to trade_ins if missing
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'trade_ins' AND column_name = 'purchase_cost'
    ) THEN
        ALTER TABLE trade_ins ADD COLUMN purchase_cost INT;
        RAISE NOTICE 'Added purchase_cost column to trade_ins';
    END IF;
END $$;

SELECT 'All critical fixes applied successfully' as status;