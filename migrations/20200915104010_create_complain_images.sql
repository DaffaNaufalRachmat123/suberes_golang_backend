-- +migrate Up
CREATE TABLE complain_images (
  id INT AUTO_INCREMENT PRIMARY KEY,
  complain_id VARCHAR(36),
  image_name TEXT,
  image_size VARCHAR(255),
  image_size_dimension VARCHAR(255),
  createdAt DATETIME NOT NULL,
  updatedAt DATETIME NOT NULL
);

-- +migrate Down
DROP TABLE complain_images;
