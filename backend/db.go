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

	USERTYPE_NIL 				= 0
	USERTYPE_PUBLISHER 			= 1
	USERTYPE_REVIEWER 			= 2
	USERTYPE_REVIEWER_PUBLISHER = 3
	USERTYPE_USER 				= 4

)
var DB_PARAMS map[string]string = map[string]string {
	"interpolateParams" : "true",
}

// Structure for user table.
type Credentials struct {
	// Password - given as plaintext by front end, and as hash by the database.
	Pw       		string `json:"password" db:"password" validate:"min=8,max=64,validpw"`
	// First Name.
	Fname    		string `json:"firstname" db:"firstName" validate:"nonzero,max=32"`
	// Last Name.
	Lname    		string `json:"lastname" db:"lastName" validate:"nonzero,max=32"`
	// Email Address.
	Email    		string `json:"email" db:"email" validate:"nonzero,max=100"`

	// User auto incremented ID.
	Id 				int `json:"userId" db:"id"`
	// User role.
	Usertype 		int `json:"usertype" db:"userType"`
	// User phone number.
	PhoneNumber 	string `json:"phonenumber" db:"phoneNumber" validate:"max=11"`
	// Organization name.
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

// Get database parameters string to place into DSN from a map.
func getDbParams(paramMap map[string]string) string {
	params := ""
	i := 0
	for key, val := range paramMap {
		if i > 0 {
			params += "&"
		}
		params += key + "=" + val
		i++
	}
	return params
}

// Initialise connection to the database.
func dbInit(user string, pw string, protocol string, h string, port int, dbname string) error {
	var err error

	// Set MySQL info in DSN format according to Go MySQL Drive -
	// user:password@protocol(host:port)/dbname?[param1=val...]
	mysqlInfo := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?%s", user, pw, protocol, h, port, dbname,
		getDbParams(DB_PARAMS)) // Setting this to allow prepared statements.
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
