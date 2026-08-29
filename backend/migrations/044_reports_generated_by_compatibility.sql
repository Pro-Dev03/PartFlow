BEGIN;

ALTER TABLE reports DROP CONSTRAINT IF EXISTS reports_generated_by_fkey;

COMMIT;
