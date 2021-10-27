package main

import (
	"database/sql"
	"fmt"
	"time"
	"reflect"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

const (
	// Constant for table operations.
	TABLE_USERS = "users"
	SELECT_ROW  = "SELECT %s FROM %s WHERE %s = ?"
	INSERT_CRED = "INSERT INTO %s (%s, %s, %s, %s) VALUES (?, ?, ?, ?)"
	INSERT_FULL = "INSERT INTO %s (%s, %s, %s, %s, %s, %s, %s) VALUES (?, ?, ?, ?, ?, ?, ?)"
	UPDATE_ROWS = "UPDATE %s SET %s = ? WHERE %s = ?"
	DELETE_ALL_ROWS = "DELETE FROM %s"

	// Constants for usertype
	USERTYPE_NIL 				= 0
	USERTYPE_PUBLISHER 			= 1
	USERTYPE_REVIEWER 			= 2
	USERTYPE_REVIEWER_PUBLISHER = 3
	USERTYPE_USER 				= 4
)
// Structure for user table.
type Credentials struct {
	// Mandatory credentials on signup.
	Pw       		string `json:"password" db:"password" validate:"min=8,max=64,validpw"`
	Fname    		string `json:"firstname" db:"firstname" validate:"nonzero,max=32"`
	Lname    		string `json:"lastname" db:"lastname" validate:"nonzero,max=32"`
	Email    		string `json:"email" db:"email" validate:"nonzero,max=100"`

	// Optional variables for credentials passthrough.
	Id 				int `json:"userId" db:"id"` 						// Auto-generated ID
	Usertype 		int `json:"usertype" db:"usertype"`				// User role in journal.
	PhoneNumber 	string `json:"phonenumber" db:"phonenumber" validate:"max=11"`
	Organization 	string `json:"organization" db:"organization" validate:"max=32"`
}

// Get the tag in a struct.
func getTag(v interface {}, structVar string, tag string) string {
	field, ok := reflect.TypeOf(v).Elem().FieldByName(structVar)
	if !ok {
		return ""
	} else {
		return field.Tag.Get(tag)
	}
}

// Get the database tag for a struct.
func getDbTag(v interface{}, structVar string) string {
	return getTag(v, structVar , "db")
}

// Get the database tag for a struct.
func getJsonTag(v interface{}, structVar string) string {
	return getTag(v, structVar, "json")
}

// Initialise connection to the database.
func dbInit(user string, pw string, protocol string, h string, port int, dbname string) error {
	var err error

	// Set MySQL info in DSN format according to Go MySQL Drive -
	// user:password@protocol(host:port)/dbname?[param1=val...]
	mysqlInfo := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?%s=%s", user, pw, protocol, h, port, dbname,
		"interpolateParams", "true") // Setting this to allow prepared statements.
	db, err = sql.Open("mysql", mysqlInfo)
	if err != nil {
		return err
	}

	// Set connection sanity options for database.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	return nil
}

func dbCloseConnection() {
	db.Close()
}
