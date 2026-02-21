-- +migrate Up
CREATE TABLE service_promos (
  id INT AUTO_INCREMENT PRIMARY KEY,
  service_id INT,
  promo_name VARCHAR(255),
  promo_count INT,
  promo_category ENUM('Discount','Cashback','Free Service'),
  promo_price INT,
  promo_description TEXT,
  promo_start_date DATETIME,
  promo_end_date DATETIME,
  promo_image TEXT,
  createdAt DATETIME NOT NULL,
  updatedAt DATETIME NOT NULL
);

-- +migrate Down
DROP TABLE service_promos;
