BEGIN;

ALTER TABLE reports
    ALTER COLUMN parameters TYPE TEXT USING parameters::text,
    ALTER COLUMN data TYPE TEXT USING data::text;

COMMIT;
