CREATE TABLE privacy_policies (
  id INT NOT NULL AUTO_INCREMENT,
  admin_id VARCHAR(255),
  policy_title VARCHAR(255),
  policy_description VARCHAR(255),
  is_valid ENUM('0', '1'),
  user_type ENUM('customer', 'mitra'),
  "createdAt" DATETIME NOT NULL,
  "updatedAt" DATETIME NOT NULL,
  PRIMARY KEY (id)
);