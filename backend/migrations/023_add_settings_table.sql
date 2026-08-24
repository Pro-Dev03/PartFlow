-- Settings Table for System Configuration
CREATE TABLE IF NOT EXISTS settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT,
    value_type VARCHAR(20) DEFAULT 'string' CHECK (value_type IN ('string', 'number', 'boolean', 'json')),
    category VARCHAR(50) DEFAULT 'general',
    description TEXT,
    is_public BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_settings_key ON settings(key);
CREATE INDEX IF NOT EXISTS idx_settings_category ON settings(category);

-- Insert default tax setting (0% by default)
INSERT INTO settings (key, value, value_type, category, description, is_public) VALUES
('tax_rate', '0', 'number', 'financial', 'نسبة الضريبة المئوية', true)
ON CONFLICT (key) DO NOTHING;

-- Insert other default settings
INSERT INTO settings (key, value, value_type, category, description, is_public) VALUES
('store_name', 'PartFlow Store', 'string', 'general', 'اسم المتجر', true),
('currency', 'ILS', 'string', 'general', 'العملة الافتراضية', true),
('currency_symbol', '₪', 'string', 'general', 'رمز العملة', true),
('low_stock_threshold', '5', 'number', 'inventory', 'حد المخزون المنخفض', false),
('debt_default_days', '30', 'number', 'financial', 'أيام السداد الافتراضية للديون', false)
ON CONFLICT (key) DO NOTHING;

COMMENT ON TABLE settings IS 'إعدادات النظام والتكوينات';
COMMENT ON COLUMN settings.key IS 'مفتاح الإعداد (يجب أن يكون فريداً)';
COMMENT ON COLUMN settings.value IS 'قيمة الإعداد';
COMMENT ON COLUMN settings.value_type IS 'نوع القيمة (string, number, boolean, json)';
COMMENT ON COLUMN settings.category IS 'فئة الإعداد';
COMMENT ON COLUMN settings.is_public IS 'هل الإعداد متاح للواجهة الأمامية';
