package main

import (
	"database/sql"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
	"os"
	"reflect"
	"regexp"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

const (
	TEAM_ID = 11

	// Constant for table operations.
	VIEW_PERMISSIONS    = "globalPermissions"
	TABLE_SERVERS       = "servers"
	VIEW_LOGIN          = "globalLogins"
	TABLE_USERS         = "users"
	TABLE_SUBMISSIONS   = "submissions"
	TABLE_FILES         = "files"
	TABLE_AUTHORS       = "authors"
	TABLE_REVIEWERS     = "reviewers"
	TABLE_CATEGORIES    = "categories"
	TABLE_IDMAPPINGS    = "idMappings"
	VIEW_USER_INFO      = "globalUserInfo"
	VIEW_SUBMISSIONLIST = "submissionList"

	// TEMP: reconcile these
	INNER_JOIN    = "%s INNER JOIN %s"
	INSERT_DOUBLE = "INSERT INTO %s (%s, %s) VALUES (?, ?)"

	SELECT_ROW               = "SELECT %s FROM %s WHERE %s = ?"
	SELECT_EXISTS            = "SELECT EXISTS (SELECT %s FROM %s WHERE %s = ?)"
	SELECT_ROW_TWO_CONDITION = "SELECT %s FROM %s WHERE %s = ? AND %s = ?"
	SELECT_ALL_ORDER_BY      = "SELECT %s FROM %s ORDER BY ?"
	SELECT_ROW_INNER_JOIN    = "SELECT %s FROM %s INNER JOIN %s ON %s = %s WHERE %s = ?"
	SELECT_ROW_ORDER_BY      = "SELECT %s FROM %s ORDER BY ? WHERE %s = ?"
	INSERT_CRED              = "INSERT INTO %s (%s, %s, %s, %s) VALUES (?, ?, ?, ?)"
	INSERT_PROJ              = "INSERT INTO %s (%s) VALUES (?) RETURNING id"
	INSERT_FILE              = "INSERT INTO %s (%s, %s) VALUES (?, ?) RETURNING id"
	INSERT_AUTHOR            = "INSERT INTO %s VALUES (?, ?)"
	INSERT_REVIEWER          = "INSERT INTO %s VALUES (?, ?)"
	INSERT_FULL              = "INSERT INTO %s (%s, %s, %s, %s, %s, %s, %s) VALUES (?, ?, ?, ?, ?, ?, ?)"
	UPDATE_ROWS              = "UPDATE %s SET %s = ? WHERE %s = ?"
	DELETE_ALL_ROWS          = "DELETE FROM %s"

	USERTYPE_NIL                = 0
	USERTYPE_PUBLISHER          = 1
	USERTYPE_REVIEWER           = 2
	USERTYPE_REVIEWER_PUBLISHER = 3
	USERTYPE_USER               = 4

	// Password related
	HASH_COST = 8
)

var DB_PARAMS map[string]string = map[string]string{
	"interpolateParams": "true",
	"parseTime":         "true",
}

// structure to hold supergroup compliant submission metadata
type SupergroupSubmissionMetaData struct {
	// creation date of the submission
	CreationDate string `json:"creationDate"`
	// author name
	AuthorName string `json:"authorName"`
}

type SuperGroupFile struct {
	Name    string `json:"filename"`
	Content string `json:"base64Value"`
}

// structure to hold formatted supergroup submissions
type SupergroupSubmission struct {
	// name of the submission
	Name string `json:"name"`
	// metadata about the submission
	Metadata *SupergroupSubmissionMetaData `json:"metadata"`
	// files array
	Files []*SuperGroupFile `json:"files"`
}

// Structure for servers.

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
		if err != nil {
			log.Printf("Scan error on checkUnique: %v\n", err)
		}
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
func dbInit(dbname string) error {
	var err error

	// Set MySQL info in DSN format according to Go MySQL Drive -
	// user:password@protocol(host:port)/dbname?[param1=val...]
	mysqlInfo := fmt.Sprintf("%s/%s?%s", os.Getenv("DATABASE_URL"), dbname,
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

// Close a database connection.
// WARNING: this function clears all data from the database, setting it
// back to the state it'd be in
func dbClear() error {
	// db tables to clear ORDER MATTERS HERE
	tablesToClear := []string{
		TABLE_CATEGORIES,
		TABLE_AUTHORS,
		TABLE_REVIEWERS,
		TABLE_FILES,
		TABLE_SUBMISSIONS,
		TABLE_IDMAPPINGS,
		TABLE_USERS,
	}
	// formats and executes a delete command for each table
	for _, table := range tablesToClear {
		stmt := fmt.Sprintf(DELETE_ALL_ROWS, table)
		_, err := db.Exec(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

// Close database connection.
func dbCloseConnection() {
	db.Close()
}

// Get all columns from an interface.
func getCols(v interface{}) []interface{} {
	s := reflect.ValueOf(v).Elem()
	numCols := s.NumField()
	cols := make([]interface{}, numCols)
	for i := 0; i < numCols; i++ {
		cols[i] = s.Field(i).Addr().Interface()
	}
	return cols
}

// -----
// Password Control
// -----

// Checks if a password contains upper case, lower case, numbers, and special characters.
func validpw(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	if st.Kind() != reflect.String {
		return errors.New("Value must be string!")
	} else {
		// Set password and character number.
		pw := st.String()
		restrictions := []string{"[a-z]", // Must contain lowercase.
			"^[" + A_NUMS + SPECIAL_CHARS + "]*$", // Must contain only some characters.
			"[A-Z]",                               // Must contain uppercase.
			"[0-9]",                               // Must contain numerics.
			"[" + SPECIAL_CHARS + "]"}             // Must contain special characters.
		for _, restriction := range restrictions {
			matcher := regexp.MustCompile(restriction)
			if !matcher.MatchString(pw) {
				return errors.New("Restriction not matched!")
			}
		}
		return nil
	}
}

// Hash a password
func hashPw(pw string) []byte {
	hash, _ := bcrypt.GenerateFromPassword([]byte(pw), HASH_COST)
	return hash
}

// Compare password and hash for validity.
func comparePw(pw string, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}
