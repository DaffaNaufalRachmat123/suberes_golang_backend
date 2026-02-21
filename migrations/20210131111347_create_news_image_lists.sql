-- +migrate Up
CREATE TABLE news_image_lists (
  id INT AUTO_INCREMENT PRIMARY KEY,
  news_id INT,
  news_image TEXT,
  news_image_size VARCHAR(255),
  news_image_dimension VARCHAR(255),
  news_image_source VARCHAR(255),
  createdAt DATETIME NOT NULL,
  updatedAt DATETIME NOT NULL
);

-- +migrate Down
DROP TABLE news_image_lists;
