-- +migrate Up
CREATE TABLE messages (
  id INT AUTO_INCREMENT PRIMARY KEY,
  sender_id INT,
  sender_name VARCHAR(255),
  recipient_ids VARCHAR(255),
  is_blast_message ENUM('yes','no'),
  is_multiple_message ENUM('yes','no'),
  image_message VARCHAR(255),
  title VARCHAR(255),
  caption_text VARCHAR(255),
  body TEXT,
  type_message ENUM('ORDER','CAMPAIGN','EXPIRED_PAYMENT','FAILURE_PAYMENT','SUCCESS_PAYMENT','PENDING_PAYMENT','ORDER_CANCELED_BY_ADMIN' , 'NOTIFICATION' , 'MESSAGE'),
  is_read ENUM('0','1'),
  createdAt DATETIME NOT NULL,
  updatedAt DATETIME NOT NULL
);

-- +migrate Down
DROP TABLE messages;
