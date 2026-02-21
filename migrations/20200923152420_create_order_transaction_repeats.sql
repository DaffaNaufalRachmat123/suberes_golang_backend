-- +migrate Up
CREATE TABLE order_transaction_repeats (
  id VARCHAR(36) PRIMARY KEY,
  order_id VARCHAR(36) NOT NULL,
  customer_id VARCHAR(36),
  mitra_id VARCHAR(36),
  service_id INT,
  sub_service_id INT,
  customer_name VARCHAR(255),
  address VARCHAR(255),
  order_time DATETIME,
  order_timestamp VARCHAR(255),
  canceled_reason TEXT,
  canceled_user ENUM('customer', 'mitra'),
  order_note TEXT,
  payment_type ENUM('cash', 'debit', 'credit card'),
  order_status ENUM('WAITING_PAYMENT','WAIT_SCHEDULE','OTW','ON_PROGRESS','FINISH','CANCELED','FINDING_MITRA'),
  id_transaction VARCHAR(255),
  gross_amount INT,
  gross_amount_mitra INT,
  gross_amount_company INT,
  customer_latitude DOUBLE,
  customer_longitude DOUBLE,
  mitra_latitude DOUBLE,
  mitra_longitude DOUBLE,
  createdAt DATETIME NOT NULL,
  updatedAt DATETIME NOT NULL
);

-- +migrate Down
DROP TABLE order_transaction_repeats;
