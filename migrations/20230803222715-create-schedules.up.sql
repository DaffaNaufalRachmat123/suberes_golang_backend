CREATE TABLE schedules (
  id INT NOT NULL AUTO_INCREMENT,
  creator_id INT,
  creator_name VARCHAR(255),
  schedule_name VARCHAR(255),
  schedule_level ENUM('executive_level', 'c_level', 'employee_level', 'mitra_level', 'customer_level', 'all_level'),
  schedule_title VARCHAR(255),
  schedule_place VARCHAR(255),
  schedule_date_time VARCHAR(255),
  schedule_message TEXT,
  schedule_template TEXT,
  schedule_is_active ENUM('0', '1'),
  timezone_code VARCHAR(255),
  "createdAt" DATETIME NOT NULL,
  "updatedAt" DATETIME NOT NULL,
  PRIMARY KEY (id)
);