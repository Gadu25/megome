CREATE TABLE IF NOT EXISTS technologies (
  id INT UNSIGNED NOT NULL AUTO_INCREMENT,
  createdByUserId INT UNSIGNED NULL,
  name VARCHAR(100) NOT NULL,
  slug VARCHAR(100) NOT NULL,

  category ENUM(
    'frontend',
    'backend',
    'devops',
    'database',
    'mobile',
    'tool',
    'language'
  ) NOT NULL,

  isVerified BOOLEAN NOT NULL DEFAULT FALSE,
  createdAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updatedAt TIMESTAMP NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (id),
  UNIQUE KEY unique_slug (slug),
  INDEX idx_slug (slug),
  INDEX idx_category (category),

  FOREIGN KEY (createdByUserId) REFERENCES users(id) ON DELETE SET NULL
);