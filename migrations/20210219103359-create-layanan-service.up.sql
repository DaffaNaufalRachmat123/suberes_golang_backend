CREATE TABLE layanan_services (
  id INT NOT NULL AUTO_INCREMENT,
  creator_id VARCHAR(255),
  layanan_title VARCHAR(255),
  layanan_description TEXT,
  layanan_image TEXT,
  layanan_image_size VARCHAR(255),
  layanan_image_dimension VARCHAR(255),
  is_active ENUM('0', '1'),
  "createdAt" DATETIME NOT NULL,
  "updatedAt" DATETIME NOT NULL,
  PRIMARY KEY (id)
);