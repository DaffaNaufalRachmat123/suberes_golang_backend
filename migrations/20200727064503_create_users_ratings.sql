-- +migrate Up
CREATE TABLE users_ratings (
  id INT AUTO_INCREMENT PRIMARY KEY,
  order_id VARCHAR(36) NOT NULL,
  customer_id INT,
  mitra_id INT,
  layanan_id INT,
  service_id INT,
  sub_service_id INT,
  rating DOUBLE,
  comment VARCHAR(255),
  rating_type ENUM('customer to mitra', 'mitra to customer'),
  createdAt DATETIME NOT NULL,
  updatedAt DATETIME NOT NULL
);

-- +migrate Down
DROP TABLE users_ratings;
