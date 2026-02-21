-- +migrate Up
CREATE TABLE payments (
  id INT AUTO_INCREMENT PRIMARY KEY,
  icon VARCHAR(255),
  is_active ENUM('0', '1'),
  title VARCHAR(255),
  type ENUM('tunai', 'virtual account', 'transfer'),
  `desc` VARCHAR(255),
  createdAt DATETIME NOT NULL,
  updatedAt DATETIME NOT NULL
);

-- +migrate Down
DROP TABLE payments;
