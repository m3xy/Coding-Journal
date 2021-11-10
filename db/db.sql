-- Create and use database (for testing)
CREATE DATABASE IF NOT EXISTS testdb;
USE testdb;

-- table for storing basic user information and credentials. Other info like user description, articles published
-- etc. can all be kept in other tables.
CREATE TABLE IF NOT EXISTS users (
  id int NOT NULL AUTO_INCREMENT, -- id which is auto-generated during user registration

  -- actual login credentials
  email varchar(100) NOT NULL UNIQUE, -- user email, functions as unique username
  password varchar(64) NOT NULL, -- encrypted user password

  -- necessary user info
  firstName varchar(32) NOT NULL, -- first name of the user
  lastName varchar(32) NOT NULL, -- last name of the user
  userType int NOT NULL DEFAULT 4, -- user type as an integer (mapped to constants in db.go)

  -- extra/optional user info
  phonenumber varchar(11), -- user phone number, is optional, 11 chars to allow for + and 10 digits
  organization varchar(32), -- organization the user is associated with (research org or company)

  CONSTRAINT userTypeCheck CHECK (userType IN (0, 1, 2, 3, 4)), -- makes userType into an integer backed enum
  PRIMARY KEY (id) -- makes the ID the primary key as we know it will be unique
);

/*
	Servers table
	Stores server group number, security token, and URL for server.
*/
CREATE TABLE IF NOT EXISTS servers (
	groupNumber int NOT NULL,    -- Group number.
	token varchar(1028) NOT NULL, -- Token corresponding to the group.
	url varchar(512) NOT NULL, -- URL for server access.

	PRIMARY KEY(groupNumber)
);

/* 
ID mappings table

table to store local to global ID mappings. These are kept distinct
from one another so that the user IDs for our users table can be auto
generated by SQL, keeping ID generation atomic, and simple. This also
allows migrated users to be treated exactly the same as local users
internal to our application.
*/
CREATE TABLE IF NOT EXISTS idMappings (
  localId int NOT NULL UNIQUE, -- ID stored locally in users table
  globalId int NOT NULL UNIQUE, -- ID which gets sent to other Journals in the Federation

  PRIMARY KEY (globalId),
  FOREIGN KEY (localId) REFERENCES users(id) -- makes local id track to the users table
);

/* 
Projects Table

table to store project ID -> name mappings. Note that in the filesystem,
projects are stored inside a wrapper directory named for their ID, to ensure
uniqueness of IDs
*/
CREATE TABLE IF NOT EXISTS projects (
  id int NOT NULL AUTO_INCREMENT, -- autogenerated project ID
  projectName varchar(64) NOT NULL, -- project name as submitted by the user
  creationDate date DEFAULT CURDATE(), -- Project creation date.
  license varchar(64), -- License the project uses for its code.

  PRIMARY KEY (id)
);

/*
Files Table

table to store individual files with their paths in the filesystem. Note that
this table is for locating files in the filesystem and does NOT store any of
the actual file data
*/
CREATE TABLE IF NOT EXISTS files (
  id int NOT NULL AUTO_INCREMENT, -- unique file id number
  projectId int NOT NULL, -- id of the project which this file is a part of 
  filePath varchar(128) NOT NULL, -- relative path from the project root directory to the file

  PRIMARY KEY (id),
  FOREIGN KEY (projectId) REFERENCES projects(id)
);

/*
Authors Table

Table mapping project IDs to their author's user IDs. Allows multiple authors to
exist for a given file/project
*/
CREATE TABLE IF NOT EXISTS authors (
  projectId int NOT NULL,
  userId int NOT NULL,

  PRIMARY KEY (projectId, userId),
  FOREIGN KEY (projectId) REFERENCES projects(id),
  FOREIGN KEY (userId) REFERENCES users(id)
);


/*
Categories table

Many to many relationship table for project corresponding categories.
*/
CREATE TABLE IF NOT EXISTS categories (
  projectId int NOT NULL,
  tag varchar(32) NOT NULL,

  PRIMARY KEY (projectId, tag),
  FOREIGN KEY (projectId) REFERENCES projects(id)
);

/*
Reviewers Table

Table mapping project ID to reviewer's IDs. Allows multiple reviewers to exist
for a project, and be held separate from the authors.
*/
CREATE TABLE IF NOT EXISTS reviewers (
  projectId int NOT NULL,
  userId int NOT NULL,

  PRIMARY KEY (projectId, userId),
  FOREIGN KEY (projectId) REFERENCES projects(id),
  FOREIGN KEY (userId) REFERENCES users(id)
);

/* Set view for users with global ID as ID. */
CREATE VIEW IF NOT EXISTS globalLogins AS
	SELECT globalId, email, password
	FROM idMappings INNER JOIN users
	ON idMappings.localId = users.id;
