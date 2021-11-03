-- Create and use database (for production)
CREATE DATABASE IF NOT EXISTS mydb;
USE mydb;

-- table for storing basic user information and credentials. Other info like user description, articles published
-- etc. can all be kept in other tables.
CREATE TABLE IF NOT EXISTS users (
  id int NOT NULL AUTO_INCREMENT, -- id which is auto-generated during user registration

  -- actual login credentials
  email varchar(100) NOT NULL UNIQUE, -- user email, functions as unique username
  password varchar(64) NOT NULL, -- encrypted user password

  -- necessary user info
  firstname varchar(32) NOT NULL, -- first name of the user
  lastname varchar(32) NOT NULL, -- last name of the user
  usertype ENUM('publisher', 'reviewer', 'publisher-reviewer', 'user') NOT NULL DEFAULT 'user', -- role of the user in the organization

  -- extra/optional user info
  phonenumber varchar(11), -- user phone number, is optional, 11 chars to allow for + and 10 digits
  organization varchar(32), -- organization the user is associated with (research org or company)

  PRIMARY KEY (id) -- makes the ID the primary key as we know it will be unique
);

CREATE TABLE IF NOT EXISTS idMappings (
  globalId int NOT NULL UNIQUE, -- Global user ID
  id int NOT NULL UNIQUE,       -- Local user ID linking to user on users table.

  PRIMARY KEY (globalId)
  FOREIGN KEY (id)
)

-- add code here to initialise other tables
