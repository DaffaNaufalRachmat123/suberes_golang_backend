-- +migrate Up
CREATE TABLE complains (
  id INT AUTO_INCREMENT PRIMARY KEY,
  complain_code VARCHAR(10),
  problem VARCHAR(255),
  status ENUM('SENT','ON_REVIEW','SOLVED','CLOSED'),
  createdAt DATETIME NOT NULL,
  updatedAt DATETIME NOT NULL
);

-- +migrate Down
DROP TABLE complains;
