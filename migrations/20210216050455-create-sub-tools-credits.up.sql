CREATE TABLE sub_tools_credits (
  id INT NOT NULL AUTO_INCREMENT,
  tool_id INT NOT NULL,
  mitra_id VARCHAR(36) NOT NULL,
  tools_credits_id INT,
  amount_paid INT,
  paid_status ENUM('0', '1'),
  installment_deadline DATETIME,
  "createdAt" DATETIME NOT NULL,
  "updatedAt" DATETIME NOT NULL,
  PRIMARY KEY (id)
);