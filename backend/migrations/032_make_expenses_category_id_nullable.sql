-- ============================================
-- Make expenses.category_id nullable to handle expenses without categories
-- ============================================
-- This ensures that expenses can exist without being assigned to a category

-- Drop the constraint if it exists (NOT NULL constraint)
ALTER TABLE expenses 
ALTER COLUMN category_id DROP NOT NULL;

-- Update any NULL category_id values to uuid_nil (00000000-0000-0000-0000-000000000000)
-- This is a safe default that our Go code can handle
UPDATE expenses 
SET category_id = '00000000-0000-0000-0000-000000000000'::uuid 
WHERE category_id IS NULL;

SELECT 'Expenses category_id made nullable' as status;
