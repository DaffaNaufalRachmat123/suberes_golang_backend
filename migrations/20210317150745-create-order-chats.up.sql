CREATE TABLE order_chats (
  id CHAR(36) NOT NULL,
  customer_id VARCHAR(36),
  mitra_id VARCHAR(36),
  order_id CHAR(36) NOT NULL,
  service_id INT,
  sub_service_id INT,
  order_chat_count INT,
  "createdAt" DATETIME NOT NULL,
  "updatedAt" DATETIME NOT NULL,
  PRIMARY KEY (id)
);