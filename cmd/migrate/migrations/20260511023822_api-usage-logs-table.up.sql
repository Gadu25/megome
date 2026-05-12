CREATE TABLE IF NOT EXISTS api_usage_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  userId INT UNSIGNED NOT NULL,
  tokenId INT UNSIGNED NOT NULL,
  endpoint VARCHAR(255) NOT NULL,
  method VARCHAR(10) NOT NULL,
  statusCode SMALLINT UNSIGNED NOT NULL,
  ipAddress VARCHAR(45) NULL,
  userAgent VARCHAR(500) NULL,
  responseTimeMs INT UNSIGNED NULL,
  createdAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  
  PRIMARY KEY (id),

  INDEX idx_user_created (userId, createdAt),
  INDEX idx_token_created (tokenId, createdAt),
  INDEX idx_status_code (statusCode),

  CONSTRAINT fk_api_logs_user
    FOREIGN KEY (userId)
    REFERENCES users(id)
    ON DELETE CASCADE,

  CONSTRAINT fk_api_logs_token
    FOREIGN KEY (tokenId)
    REFERENCES personal_access_tokens(id)
    ON DELETE CASCADE
);