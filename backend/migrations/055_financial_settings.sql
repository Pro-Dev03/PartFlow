INSERT INTO settings (key, value, value_type, category, description, is_public)
VALUES
    ('max_discount_rate', '15', 'number', 'financial', 'الحد الأقصى للخصم المئوي', false)
ON CONFLICT (key) DO NOTHING;
