-- +migrate Up
CREATE TABLE suberes_logs (
  id INT AUTO_INCREMENT PRIMARY KEY,
  log_name VARCHAR(255),
  log_type VARCHAR(255),
  log_url VARCHAR(255),
  log_body TEXT,
  log_time VARCHAR(255),
  createdAt DATETIME NOT NULL,
  updatedAt DATETIME NOT NULL
);

-- +migrate Down
DROP TABLE suberes_logs;
