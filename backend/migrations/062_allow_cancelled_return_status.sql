-- Keep return cancellation aligned with the UI and return reversal workflow.
ALTER TABLE returns DROP CONSTRAINT IF EXISTS returns_status_check;

ALTER TABLE returns
    ADD CONSTRAINT returns_status_check
    CHECK (status IN ('PENDING', 'APPROVED', 'PROCESSING', 'COMPLETED', 'REJECTED', 'CANCELLED'));
