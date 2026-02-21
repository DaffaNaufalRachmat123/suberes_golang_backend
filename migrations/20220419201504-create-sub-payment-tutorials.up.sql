CREATE TABLE sub_payment_tutorials (
  id INT NOT NULL AUTO_INCREMENT,
  payment_id INT,
  sub_payment_id INT,
  title VARCHAR(255),
  description TEXT,
  "createdAt" DATETIME NOT NULL,
  "updatedAt" DATETIME NOT NULL,
  PRIMARY KEY (id)
);