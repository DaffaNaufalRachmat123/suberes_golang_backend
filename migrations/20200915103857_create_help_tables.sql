-- +migrate Up
CREATE TABLE help_tables (
  id INT AUTO_INCREMENT PRIMARY KEY,
  title VARCHAR(255),
  description TEXT,
  help_type ENUM('customer', 'mitra'),
  watching_count INT DEFAULT 0,
  createdAt DATETIME NOT NULL,
  updatedAt DATETIME NOT NULL
);

-- +migrate Down
DROP TABLE help_tables;
