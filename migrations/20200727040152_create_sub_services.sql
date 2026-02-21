-- +migrate Up
CREATE TABLE sub_services (
  id INT AUTO_INCREMENT PRIMARY KEY,
  service_id INT NOT NULL,
  sub_price_service_title VARCHAR(255),
  sub_price_service INT,
  sub_service_description VARCHAR(255),
  company_percentage DOUBLE,
  minutes_sub_services INT(11),
  criteria VARCHAR(255),
  is_recommended ENUM('0', '1'),
  createdAt DATETIME NOT NULL,
  updatedAt DATETIME NOT NULL,
  FOREIGN KEY (service_id) REFERENCES services(id)
);

-- +migrate Down
DROP TABLE sub_services;
