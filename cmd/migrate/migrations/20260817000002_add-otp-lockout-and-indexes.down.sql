ALTER TABLE email_verification_otps
  DROP INDEX idx_email_hash,
  DROP INDEX idx_user,
  DROP COLUMN failedAttempts;
