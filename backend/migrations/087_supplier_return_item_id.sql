BEGIN;

-- Keep Cloud inserts safe when a legacy or third party writer omits id. The
-- application still creates one UUID and reuses it for Local, Queue, and Cloud.
-- This only changes the default for future rows; existing data and posted
-- inventory/financial effects are untouched.
ALTER TABLE supplier_return_items
    ALTER COLUMN id SET DEFAULT gen_random_uuid();

COMMIT;
