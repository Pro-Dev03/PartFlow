-- Serialize return quantity checks and align database validation with the API.
BEGIN;

CREATE OR REPLACE FUNCTION public.validate_return_quantity()
RETURNS TRIGGER
LANGUAGE plpgsql
SET search_path = pg_catalog, public, pg_temp
AS $$
DECLARE
    original_qty INTEGER;
    total_returned INTEGER;
BEGIN
    IF NEW.sale_item_id IS NULL THEN
        RETURN NEW;
    END IF;

    -- Lock the sold line first so concurrent returns cannot both pass using
    -- the same pre-return quantity.
    SELECT si.quantity
    INTO original_qty
    FROM public.sale_items AS si
    WHERE si.id = NEW.sale_item_id
    FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Sale item % does not exist', NEW.sale_item_id
            USING ERRCODE = '23503';
    END IF;

    SELECT COALESCE(SUM(ri.quantity_returned), 0)
    INTO total_returned
    FROM public.return_items AS ri
    JOIN public.returns AS r ON r.id = ri.return_id
    WHERE ri.sale_item_id = NEW.sale_item_id
      AND ri.id IS DISTINCT FROM NEW.id
      AND UPPER(COALESCE(r.status, '')) NOT IN ('REJECTED', 'CANCELLED');

    total_returned := total_returned + NEW.quantity_returned;
    IF total_returned > original_qty THEN
        RAISE EXCEPTION 'Cannot return more items than originally sold. Original: %, Already returned: %, Current return: %',
            original_qty, total_returned - NEW.quantity_returned, NEW.quantity_returned
            USING ERRCODE = '23514';
    END IF;

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS validate_return_quantity_trigger ON public.return_items;
CREATE TRIGGER validate_return_quantity_trigger
    BEFORE INSERT OR UPDATE OF quantity_returned ON public.return_items
    FOR EACH ROW
    EXECUTE FUNCTION public.validate_return_quantity();

COMMIT;
