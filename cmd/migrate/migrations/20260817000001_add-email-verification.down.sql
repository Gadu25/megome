DROP TABLE IF EXISTS email_verification_otps;

ALTER TABLE users DROP COLUMN emailVerifiedAt;
