-- +migrate Up
CREATE TABLE services (
  id INT AUTO_INCREMENT PRIMARY KEY,
  parent_id INT,
  service_name VARCHAR(255),
  service_description TEXT,
  service_image_thumbnail TEXT,
  service_count INT,
  service_type ENUM('Durasi', 'Luas Ruangan', 'Single'),
  service_category ENUM('Cleaning','Disinfectant','Fogging','Borongan','Lainnya'),
  is_residental ENUM('true', 'false'),
  service_status ENUM('Regular','Premium','Pro Premium'),
  is_active ENUM('0', '1'),
  max_order_count INT,
  payment_timeout INT,
  createdAt DATETIME NOT NULL,
  updatedAt DATETIME NOT NULL
);

-- +migrate Down
DROP TABLE services;
