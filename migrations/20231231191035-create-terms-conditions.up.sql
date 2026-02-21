CREATE TABLE terms_conditions (
  id INT NOT NULL AUTO_INCREMENT,
  creator_id VARCHAR(255),
  title VARCHAR(255),
  body TEXT,
  is_active ENUM('0', '1'),
  can_select ENUM('0', '1') DEFAULT '1',
  toc_type ENUM('terms_of_condition', 'terms_of_service', 'privacy_policy'),
  toc_user_type ENUM('customer', 'mitra'),
  "createdAt" DATETIME NOT NULL,
  "updatedAt" DATETIME NOT NULL,
  PRIMARY KEY (id)
);