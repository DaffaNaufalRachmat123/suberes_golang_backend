-- +migrate Up
CREATE TABLE news_lists (
  id INT AUTO_INCREMENT PRIMARY KEY,
  creator_id INT,
  creator_name VARCHAR(255),
  news_title VARCHAR(255),
  news_body TEXT,
  news_type ENUM('Suberes Update', 'News'),
  news_image TEXT,
  news_image_size VARCHAR(255),
  news_image_dimension VARCHAR(255),
  is_revision ENUM('0','1'),
  read_count INT,
  like_count INT,
  comment_count INT,
  share_count INT,
  narasumber VARCHAR(255),
  is_broadcast ENUM('0','1'),
  timezone_code VARCHAR(50),
  createdAt DATETIME NOT NULL,
  updatedAt DATETIME NOT NULL
);

-- +migrate Down
DROP TABLE news_lists;
