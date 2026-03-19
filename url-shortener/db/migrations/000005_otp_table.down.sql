BEGIN;

SET search_path TO urlshortener;

ALTER TABLE IF EXISTS otp DROP CONSTRAINT IF EXISTS otp_purpose_check;
ALTER TABLE IF EXISTS otp DROP CONSTRAINT IF EXISTS otp_user_id_fkey;
DROP INDEX IF EXISTS idx_otp_user_code_purpose;
DROP INDEX IF EXISTS otp_code_key;

DROP TABLE IF EXISTS otp;

COMMIT;
