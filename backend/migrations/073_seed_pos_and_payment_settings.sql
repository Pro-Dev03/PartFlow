INSERT INTO settings (id, key, value, value_type, category, description, is_public, created_at, updated_at)
VALUES
    ('00000000-0000-0000-0000-000000000073', 'pos_products_per_page', '12', 'number', 'appearance', 'عدد منتجات نقطة البيع في الصفحة', false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('00000000-0000-0000-0000-000000000074', 'pos_product_view_mode', 'cards', 'string', 'appearance', 'طريقة عرض منتجات نقطة البيع', false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('00000000-0000-0000-0000-000000000075', 'electronic_payments_enabled', 'false', 'boolean', 'payments', 'تفعيل الدفع الإلكتروني', false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('00000000-0000-0000-0000-000000000076', 'payment_provider', 'manual', 'string', 'payments', 'مزود الدفع الإلكتروني', false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('00000000-0000-0000-0000-000000000077', 'payment_environment', 'test', 'string', 'payments', 'بيئة الدفع الإلكتروني', false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('00000000-0000-0000-0000-000000000078', 'payment_public_key', '', 'string', 'payments', 'المفتاح العام لمزود الدفع', false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('00000000-0000-0000-0000-000000000079', 'payment_secret_key', '', 'string', 'payments', 'المفتاح السري لمزود الدفع', false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('00000000-0000-0000-0000-000000000080', 'payment_merchant_id', '', 'string', 'payments', 'معرف التاجر لدى مزود الدفع', false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('00000000-0000-0000-0000-000000000081', 'payment_terminal_id', '', 'string', 'payments', 'معرف جهاز الدفع', false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('00000000-0000-0000-0000-000000000082', 'payment_webhook_url', '', 'string', 'payments', 'عنوان Webhook للدفع', false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('00000000-0000-0000-0000-000000000083', 'payment_webhook_secret', '', 'string', 'payments', 'سر توقيع Webhook للدفع', false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('00000000-0000-0000-0000-000000000084', 'payment_methods', '["card"]', 'json', 'payments', 'طرق الدفع الإلكتروني المفعلة', false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (key) DO NOTHING;
