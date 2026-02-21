-- +migrate Up
CREATE TABLE sub_payments (
  id INT AUTO_INCREMENT PRIMARY KEY,
  payment_id INT,
  icon VARCHAR(255),
  title VARCHAR(255),
  title_payment VARCHAR(255),
  enabled ENUM('0', '1'),
  `desc` VARCHAR(255),
  createdAt DATETIME NOT NULL,
  updatedAt DATETIME NOT NULL
);

-- +migrate Down
DROP TABLE sub_payments;
