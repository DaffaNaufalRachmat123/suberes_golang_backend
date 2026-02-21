CREATE TABLE service_guarantees (
  id INT NOT NULL AUTO_INCREMENT,
  service_id INT NOT NULL UNIQUE,
  user_id INT,
  guarantee_name VARCHAR(255),
  guarantee_description VARCHAR(255),
  is_guarantee_enabled ENUM('0', '1'),
  "createdAt" DATETIME NOT NULL,
  "updatedAt" DATETIME NOT NULL,
  PRIMARY KEY (id)
);