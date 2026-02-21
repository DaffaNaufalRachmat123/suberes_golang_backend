CREATE TABLE sub_service_addeds (
  id INT NOT NULL AUTO_INCREMENT,
  order_id CHAR(36) NOT NULL,
  sub_service_add_id INT,
  title VARCHAR(255) DEFAULT '',
  base_amount INT DEFAULT 0,
  amount INT DEFAULT 0,
  additional_type ENUM('choice', 'cashback', 'discount', 'free') DEFAULT NULL,
  "createdAt" DATETIME NOT NULL,
  "updatedAt" DATETIME NOT NULL,
  PRIMARY KEY (id)
);