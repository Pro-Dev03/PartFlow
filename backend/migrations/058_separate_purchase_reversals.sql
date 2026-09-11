-- Keep purchase reversals distinct from supplier returns in inventory status.
ALTER TABLE inventory_items DROP CONSTRAINT IF EXISTS inventory_items_status_check;
ALTER TABLE inventory_items
    ADD CONSTRAINT inventory_items_status_check CHECK (status IN (
        'PURCHASED', 'RECEIVED', 'INSPECTION', 'AVAILABLE', 'RESERVED', 'SOLD',
        'DAMAGED', 'IN_REPAIR', 'RETURNED', 'REVERSED', 'FOR_PARTS', 'ARCHIVED'
    ));

UPDATE inventory_items ii
SET status = 'REVERSED', updated_at = NOW()
WHERE ii.status = 'RETURNED'
  AND EXISTS (
      SELECT 1
      FROM inventory_movements im
      WHERE im.item_id = ii.id
        AND im.movement_type = 'PURCHASE_REVERSAL'
  );