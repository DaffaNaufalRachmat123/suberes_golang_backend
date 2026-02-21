CREATE TABLE regions (
  id INT NOT NULL AUTO_INCREMENT,
  country_id INT,
  region_name VARCHAR(255),
  "createdAt" DATETIME NOT NULL,
  "updatedAt" DATETIME NOT NULL,
  PRIMARY KEY (id)
);