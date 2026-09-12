INSERT INTO settings (key, value, value_type, category, description, is_public)
VALUES
    ('discounts_enabled', 'true', 'boolean', 'financial', 'السماح بالخصومات', false)
ON CONFLICT (key) DO NOTHING;
