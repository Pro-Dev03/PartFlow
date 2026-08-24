-- ============================================
-- Fix expenses table schema to match code expectations
-- ============================================
-- The Go code expects different column names than the current schema

-- Add missing columns to expenses table
ALTER TABLE expenses 
ADD COLUMN IF NOT EXISTS category_id UUID,
ADD COLUMN IF NOT EXISTS title VARCHAR(255),
ADD COLUMN IF NOT EXISTS currency VARCHAR(10) DEFAULT 'USD',
ADD COLUMN IF NOT EXISTS reference VARCHAR(100),
ADD COLUMN IF NOT EXISTS is_recurring BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS recurring_period VARCHAR(50),
ADD COLUMN IF NOT EXISTS approved_by UUID REFERENCES users(id),
ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'pending';

-- Create index for category_id if it doesn't exist
CREATE INDEX IF NOT EXISTS idx_expenses_category_id ON expenses(category_id);
CREATE INDEX IF NOT EXISTS idx_expenses_status ON expenses(status);

-- Migrate existing data from category to category_id if needed
-- This is a simplified migration - in production you'd need proper category mapping
UPDATE expenses 
SET title = COALESCE(title, description),
    reference = COALESCE(reference, reference_number)
WHERE title IS NULL OR reference IS NULL;

SELECT 'Expenses schema fixed' as status;
