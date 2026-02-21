-- +migrate Up
CREATE TABLE category_services (
  id INT AUTO_INCREMENT PRIMARY KEY,
  layanan_id INT,
  creator_id INT,
  category_service TEXT,
  createdAt DATETIME NOT NULL,
  updatedAt DATETIME NOT NULL
);

-- +migrate Down
DROP TABLE category_services;
