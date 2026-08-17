ALTER TABLE email_verification_otps
  ADD COLUMN failedAttempts INT UNSIGNED NOT NULL DEFAULT 0,
  ADD KEY idx_email_hash (email, otpHash),
  ADD KEY idx_user (userId);
