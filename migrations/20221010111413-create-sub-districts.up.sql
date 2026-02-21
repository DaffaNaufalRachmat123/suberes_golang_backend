CREATE TABLE sub_districts (
  id INT NOT NULL AUTO_INCREMENT,
  country_id INT,
  region_id INT,
  district_id INT,
  sub_district_name VARCHAR(255),
  "createdAt" DATETIME NOT NULL,
  "updatedAt" DATETIME NOT NULL,
  PRIMARY KEY (id)
);