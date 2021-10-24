CREATE TABLE IF NOT EXISTS users (
  id int NOT NULL AUTO_INCREMENT, -- id which is auto-generated during user registration
  username varchar(64) NOT NULL, -- username of the user
  password varchar(64) NOT NULL, -- encrypted user password
  firstname varchar(32) NOT NULL, -- first name of the user
  lastname varchar(32) NOT NULL, -- last name of the user
  email varchar(100), -- user email, can be null
  usertype ENUM('publisher', 'reviewer', 'publisher-reviewer', 'user') NOT NULL DEFAULT 'user', -- "type" of user. 
  phonenumber varchar(11), -- user phone number, is optional, 11 chars to allow for + and 10 digits
  organization varchar(32), -- organization the user is associated with (research org or company)

  PRIMARY KEY (id) -- makes the ID the primary key as we know it will be unique
);

-- add code here to initialise other tables
