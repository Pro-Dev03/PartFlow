-- Revoke short-lived access JWTs immediately on logout or password change.
BEGIN;

ALTER TABLE public.users
    ADD COLUMN IF NOT EXISTS session_version BIGINT NOT NULL DEFAULT 0;

COMMIT;
