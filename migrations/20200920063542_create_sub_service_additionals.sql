-- +migrate Up
CREATE TABLE sub_service_additionals (
  id INT AUTO_INCREMENT PRIMARY KEY,
  sub_service_id INT NOT NULL,
  title VARCHAR(255),
  base_amount DOUBLE,
  amount DOUBLE,
  additional_type ENUM('choice','cashback','discount','free'),
  createdAt DATETIME NOT NULL,
  updatedAt DATETIME NOT NULL
);

-- +migrate Down
DROP TABLE sub_service_additionals;
