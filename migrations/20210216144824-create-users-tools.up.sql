CREATE TABLE users_tools (
  id INT NOT NULL AUTO_INCREMENT,
  user_id INT,
  tool_id INT,
  tool_name VARCHAR(255),
  tool_count INT,
  tool_status ENUM('Company', 'Mitra'),
  "createdAt" DATETIME NOT NULL,
  "updatedAt" DATETIME NOT NULL,
  PRIMARY KEY (id)
);