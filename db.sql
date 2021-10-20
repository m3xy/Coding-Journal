CREATE TABLE IF NOT EXISTS users (
  id int NOT NULL AUTO_INCREMENT,
  name varchar(32) NOT NULL,
  PRIMARY KEY (id)
);

INSERT INTO users (name) VALUES
('Alexandre'),
('David'),
('Eric'),
('Manuel'),
('Marcus');