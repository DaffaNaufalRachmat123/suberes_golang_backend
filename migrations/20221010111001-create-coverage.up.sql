CREATE TABLE coverages (
  id INT NOT NULL AUTO_INCREMENT,
  country_id INT,
  region_id INT,
  district_id INT,
  sub_district_id INT,
  "createdAt" DATETIME NOT NULL,
  "updatedAt" DATETIME NOT NULL,
  PRIMARY KEY (id)
);