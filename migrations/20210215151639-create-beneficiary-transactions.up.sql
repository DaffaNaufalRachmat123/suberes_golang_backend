CREATE TABLE beneficiary_transactions (
  id INT NOT NULL AUTO_INCREMENT,
  user_id INT,
  beneficiary_id INT,
  external_id VARCHAR(255),
  transaction_amount INT,
  transaction_status VARCHAR(255),
  "createdAt" DATETIME NOT NULL,
  "updatedAt" DATETIME NOT NULL,
  PRIMARY KEY (id)
);