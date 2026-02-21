-- +migrate Up
CREATE TABLE order_rejecteds (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  order_id VARCHAR(36) NOT NULL,
  customer_id VARCHAR(36),
  mitra_id VARCHAR(36),
  service_id INT,
  sub_service_id INT,
  createdAt DATETIME NOT NULL,
  updatedAt DATETIME NOT NULL
);

-- +migrate Down
DROP TABLE order_rejecteds;
