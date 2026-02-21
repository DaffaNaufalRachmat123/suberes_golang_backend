-- +migrate Up
CREATE TABLE payment_accounts (
  id INT AUTO_INCREMENT PRIMARY KEY,
  mitra_id VARCHAR(36),
  beneficiary_bank VARCHAR(255),
  beneficiary_account VARCHAR(255),
  beneficiary_name VARCHAR(255),
  beneficiary_type ENUM('simulator', 'real'),
  beneficiary_status ENUM('not active', 'active'),
  createdAt DATETIME NOT NULL,
  updatedAt DATETIME NOT NULL
);

-- +migrate Down
DROP TABLE payment_accounts;
