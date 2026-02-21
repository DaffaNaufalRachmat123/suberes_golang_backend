CREATE TABLE banner_lists (
  id INT NOT NULL AUTO_INCREMENT,
  creator_id INT,
  creator_name VARCHAR(255),
  banner_title VARCHAR(255),
  banner_body TEXT,
  banner_image TEXT,
  banner_image_size VARCHAR(255),
  banner_image_dimension VARCHAR(255),
  banner_type ENUM('promo', 'coupon', 'visi misi', 'other'),
  is_revision ENUM('0', '1'),
  is_broadcast ENUM('0', '1'),
  "createdAt" DATETIME NOT NULL,
  "updatedAt" DATETIME NOT NULL,
  PRIMARY KEY (id)
);