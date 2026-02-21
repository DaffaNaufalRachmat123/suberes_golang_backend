CREATE TABLE order_chat_messages (
  id INT NOT NULL AUTO_INCREMENT,
  order_chat_id VARCHAR(255),
  message TEXT,
  type_message ENUM('text', 'image', 'video', 'audio'),
  message_file_path TEXT,
  blur_message_file_path TEXT,
  message_file_size VARCHAR(255),
  is_message_sent ENUM('0', '1'),
  is_message_read ENUM('0', '1'),
  is_message_listened ENUM('0', '1'),
  "createdAt" DATETIME NOT NULL,
  "updatedAt" DATETIME NOT NULL,
  PRIMARY KEY (id)
);