-- ============================================
-- Create expense_categories table
-- ============================================
-- This table stores categories for expenses

CREATE TABLE IF NOT EXISTS expense_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    color VARCHAR(7),
    icon VARCHAR(50),
    budget DECIMAL(10,2) DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(name)
);

CREATE INDEX IF NOT EXISTS idx_expense_categories_name ON expense_categories(name);
CREATE INDEX IF NOT EXISTS idx_expense_categories_active ON expense_categories(is_active);

-- Add trigger for updated_at
CREATE TRIGGER update_expense_categories_updated_at BEFORE UPDATE ON expense_categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert some default expense categories
INSERT INTO expense_categories (name, description, color, icon, budget, is_active) VALUES
('رواتب', 'رواتب الموظفين', '#3B82F6', 'users', 0, true),
('إيجار', 'إيجار المحل', '#10B981', 'home', 0, true),
('كهرباء', 'فواتير الكهرباء', '#F59E0B', 'zap', 0, true),
('مياه', 'فواتير المياه', '#06B6D4', 'droplet', 0, true),
('إنترنت', 'فواتير الإنترنت', '#8B5CF6', 'wifi', 0, true),
('تسويق', 'مصاريف التسويق', '#EC4899', 'megaphone', 0, true),
('صيانة', 'صيانة المعدات', '#EF4444', 'wrench', 0, true),
('نقل', 'مصاريف النقل', '#6B7280', 'truck', 0, true),
('أخرى', 'مصاريف أخرى', '#14B8A6', 'more-horizontal', 0, true)
ON CONFLICT (name) DO NOTHING;

SELECT 'Expense categories table created' as status;
