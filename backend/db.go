package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"reflect"
	"time"
)

var db *sql.DB

const (
	// Constant for table operations.
	TABLE_USERS      = "users"
	TABLE_IDMAPPINGS = "idMappings"
	TEAM_ID          = "11"
	SELECT_ROW       = "SELECT %s FROM %s WHERE %s = ?"
	INSERT_CRED      = "INSERT INTO %s (%s, %s, %s, %s) VALUES (?, ?, ?, ?)"
	INSERT_FULL      = "INSERT INTO %s (%s, %s, %s, %s, %s, %s, %s) VALUES (?, ?, ?, ?, ?, ?, ?)"
	INSERT_DOUBLE    = "INSERT INTO %s (%s, %s) VALUES (?, ?)"
	UPDATE_ROWS      = "UPDATE %s SET %s = ? WHERE %s = ?"
	DELETE_ALL_ROWS  = "DELETE FROM %s"

	USERTYPE_NIL                = 0
	USERTYPE_PUBLISHER          = 1
	USERTYPE_REVIEWER           = 2
	USERTYPE_REVIEWER_PUBLISHER = 3
	USERTYPE_USER               = 4
)

var DB_PARAMS map[string]string = map[string]string{
	"interpolateParams": "true",
}

// Structure for user table.
type Credentials struct {
	// Password - given as plaintext by front end, and as hash by the database.
	Pw string `json:"password" db:"password" validate:"min=8,max=64,validpw"`
	// First Name.
	Fname string `json:"firstname" db:"firstname" validate:"nonzero,max=32"`
	// Last Name.
	Lname string `json:"lastname" db:"lastname" validate:"nonzero,max=32"`
	// Email Address.
	Email string `json:"email" db:"email" validate:"nonzero,max=100"`

	// User auto incremented ID.
	Id int `json:"userId" db:"id"`
	// User role.
	Usertype int `json:"usertype" db:"usertype"`
	// User phone number.
	PhoneNumber string `json:"phonenumber" db:"phonenumber" validate:"max=11"`
	// Organization name.
	Organization string `json:"organization" db:"organization" validate:"max=32"`
}

// Structure for ID mappings.
type IdMappings struct {
	GlobalId int `json:"globalId" db:"globalId"`
	Id       int `json:"userId" db:"id"`
}

// Get the tag in a struct.
func getTag(v interface{}, structVar string, tag string) string {
	field, ok := reflect.TypeOf(v).Elem().FieldByName(structVar)
	if !ok {
		return ""
	} else {
		return field.Tag.Get(tag)
	}
}

// Check if a value is unique in a given table.
func checkUnique(table string, varName string, val string) bool {
	// Query prepared and formatted statement.
	stmt := fmt.Sprintf(SELECT_ROW, varName, table, varName)
	query := db.QueryRow(stmt, val)

	// Scan query and check for existing rows.
	var res interface{}
	err := query.Scan(&res)
	if err != sql.ErrNoRows {
		// Table isn't empty or error occured, return false.
		return false
	} else {
		return true
	}
}

// Get the database tag for a struct.
func getDbTag(v interface{}, structVar string) string {
	return getTag(v, structVar, "db")
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
