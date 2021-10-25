package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

const (
	// Constants for database connection.
	host     = "127.0.0.1"
	port     = 3600
	user     = "mysql"
	protocol = "tcp"
	password = "secret"
	dbname   = "my-db"

	// Constant for table operations.
	TABLE_USERS = "users"
	SELECT_ROW  = "SELECT $1 FROM $2 WHERE $3 = $4"
	INSERT_CRED = "INSERT INTO $1 ($2, $3, $4, $5, $6) VALUES ($7, $8, $9, $10)"
	UPDATE_ROWS = "UPDATE $1 SET $2 = $3 WHERE $4 = $5"

	// Constants for usertype
	USERTYPE_PUBLISHER 			= 1
	USERTYPE_REVIEWER 			= 2
	USERTYPE_REVIEWER_PUBLISHER = 3
	USERTYPE_USER 				= 4
)

// Structure for user table.
type Credentials struct {
	// Mandatory credentials on signup.
	Username 		string `json:"username" db:"username" validate:"nonzero"`
	Pw       		string `json:"password" db:"password" validate:"min=8,max=64,validpw"`
	Fname    		string `json:"firstname" db:"firstname" validate:"nonzero"`
	Lname    		string `json:"lastname" db:"lastname" validate:"nonzero"`
	Email    		string `json:"email" db:"email" validate:"nonzero"`

	// Optional variables for credentials passthrough.
	Id 				int `json:"id" db:"id"` 						// Auto-generated ID
	Usertype 		int `json:"usertype" db:"usertype"`				// User role in journal.
	PhoneNumber 	string `json:"phonenumber" db:"phonenumber"`
	Organization 	string `json:"organization" db:"organization"`
}


// Initialise connection to the database.
func dbInit() error {
	var err error

	// Set MySQL info in DSN format according to Go MySQL Drive -
	// user:password@protocol(host:port)/dbname?[param1=val...]
	mysqlInfo := fmt.Sprintf("%s:%s@%s(%s:%d)/%s", user, password, protocol, host, port, dbname)
	db, err = sql.Open("mariadb", mysqlInfo)
	if err != nil {
		return err
	}

	// Set connection sanity options for database.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	return nil
}
