INSERT INTO settings (key, value, value_type, category, description, is_public)
VALUES (
    'default_profit_margin',
    '30',
    'number',
    'financial',
    'نسبة الربح المقترحة عند إضافة منتج',
    false
)
ON CONFLICT (key) DO NOTHING;
