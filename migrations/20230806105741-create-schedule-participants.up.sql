CREATE TABLE schedule_participants (
  id INT NOT NULL AUTO_INCREMENT,
  schedule_id INT,
  user_id VARCHAR(255),
  participant_type ENUM('executive_level', 'c_level', 'employee_level', 'mitra_level', 'customer_level', 'all_level'),
  participant_complete_name VARCHAR(255),
  participant_email VARCHAR(255),
  "createdAt" DATETIME NOT NULL,
  "updatedAt" DATETIME NOT NULL,
  PRIMARY KEY (id)
);