CREATE TABLE IF NOT EXISTS project_images (
  id INT UNSIGNED NOT NULL AUTO_INCREMENT,
  projectId INT UNSIGNED NOT NULL,

  url VARCHAR(500) NOT NULL,
  type ENUM('cover', 'screenshot', 'demo') NOT NULL,
  position INT UNSIGNED DEFAULT NULL,

  createdAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updatedAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (id),

  FOREIGN KEY (projectId) REFERENCES projects(id) ON DELETE CASCADE,

  INDEX idx_project_images_project (projectId),
  INDEX idx_project_images_type (type),
  INDEX idx_project_images_position (projectId, position)
);