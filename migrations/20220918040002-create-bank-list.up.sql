CREATE TABLE bank_lists (
  id INT NOT NULL AUTO_INCREMENT,
  bank_image VARCHAR(255),
  name VARCHAR(255),
  code VARCHAR(255),
  disbursement_code VARCHAR(255),
  method_type ENUM('bank', 'ewallet'),
  can_topup ENUM('0', '1'),
  can_disbursement ENUM('0', '1'),
  min_topup INT,
  min_disbursement INT,
  topup_fee DOUBLE,
  disbursement_fee DOUBLE,
  is_percentage ENUM('0', '1'),
  "createdAt" DATETIME NOT NULL,
  "updatedAt" DATETIME NOT NULL,
  PRIMARY KEY (id)
);