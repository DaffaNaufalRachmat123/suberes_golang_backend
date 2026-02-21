CREATE TABLE tools_credits (
  id INT NOT NULL AUTO_INCREMENT,
  tool_id INT,
  mitra_id VARCHAR(36),
  "createdAt" DATETIME NOT NULL,
  "updatedAt" DATETIME NOT NULL,
  PRIMARY KEY (id)
);