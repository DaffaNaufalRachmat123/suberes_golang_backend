CREATE TABLE rewards (
  id INT NOT NULL AUTO_INCREMENT,
  user_id INT,
  service_id INT,
  sub_service_id INT,
  reward_title VARCHAR(255),
  reward_description TEXT,
  reward_type ENUM('no level', 'silver', 'gold', 'platinum'),
  reward_start_date DATETIME,
  reward_end_date DATETIME,
  "createdAt" DATETIME NOT NULL,
  "updatedAt" DATETIME NOT NULL,
  PRIMARY KEY (id)
);