-- +migrate Up
CREATE TABLE tools (
  id INT AUTO_INCREMENT PRIMARY KEY,
  tool_name VARCHAR(255),
  tool_count INT,
  tool_price INT,
  company_price_additional INT,
  tool_type VARCHAR(255),
  debt_per_week INT,
  installment_period INT,
  tool_image TEXT,
  is_out_of_stock ENUM('0', '1'),
  createdAt DATETIME NOT NULL,
  updatedAt DATETIME NOT NULL
);

-- +migrate Down
DROP TABLE tools;
