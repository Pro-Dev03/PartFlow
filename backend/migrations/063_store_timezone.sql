-- Persist the store's IANA timezone independently from country presentation settings.
INSERT INTO settings (key, value, value_type, category, description, is_public)
VALUES ('store_timezone', 'Asia/Jerusalem', 'string', 'regional', 'المنطقة الزمنية الثابتة للمتجر', true)
ON CONFLICT (key) DO NOTHING;