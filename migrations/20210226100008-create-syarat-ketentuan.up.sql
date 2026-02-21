CREATE TABLE syarat_ketentuans (
  id INT NOT NULL AUTO_INCREMENT,
  creator_id INT,
  title VARCHAR(255),
  body TEXT,
  image TEXT,
  is_pendaftaran_mitra ENUM('0', '1'),
  is_active ENUM('0', '1'),
  expired_date DATETIME,
  "createdAt" DATETIME NOT NULL,
  "updatedAt" DATETIME NOT NULL,
  PRIMARY KEY (id)
);