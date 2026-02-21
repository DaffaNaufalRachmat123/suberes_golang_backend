-- +migrate Up
CREATE TABLE guide_tables (
  id INT AUTO_INCREMENT PRIMARY KEY,
  guide_title VARCHAR(255),
  guide_description TEXT,
  guide_type ENUM('customer', 'mitra'),
  watching_count INT DEFAULT 0,
  createdAt DATETIME NOT NULL,
  updatedAt DATETIME NOT NULL
);

-- +migrate Down
DROP TABLE guide_tables;
