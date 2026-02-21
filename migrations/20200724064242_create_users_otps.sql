-- +migrate Up
CREATE TABLE users_otps (
  id INT AUTO_INCREMENT PRIMARY KEY,
  users_id VARCHAR(36) NOT NULL UNIQUE,
  otp_code TEXT,
  otp_type ENUM('login_code', 'email_verification_code', 'change_data', 'change_pin', 'change_phone_number', 'forgot_password'),
  session_time TIME NOT NULL DEFAULT CURRENT_TIME,
  createdAt DATETIME NOT NULL,
  updatedAt DATETIME NOT NULL,
  FOREIGN KEY (users_id) REFERENCES users(id)
);

-- +migrate Down
DROP TABLE users_otps;
