package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/validator.v2"
)

const (
	SPECIAL_CHARS = "//!//@//#//$//%//^//&//*//,//.//;//:"
	ALPHANUMERICS = "a-zA-Z0-9"
	HASH_COST = 8
)

/*
 Router function to log in to website.
 Content type: application/json
 Sucess: 200, Credentials are correct
 Failure: 401, Unauthorized
 Returns: userId
*/
func logIn(w http.ResponseWriter, r *http.Request) {
	// Set up writer response.
	w.Header().Set("Content-Type", "application/json")
	respMap := make(map[string]string)


	// Get credentials from log in request.
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get credentials at given email, and assign it.
	vars := fmt.Sprintf("%s, %s", getDbTag(&Credentials{}, "Pw"), getDbTag(&Credentials{}, "Id"))
	stmt := fmt.Sprintf(SELECT_ROW, vars, TABLE_USERS, getDbTag(&Credentials{}, "Email"))
	res := db.QueryRow(stmt, creds.Email)
	storedCreds := &Credentials{}
	err = res.Scan(&storedCreds.Pw, &storedCreds.Id)
	if err != nil {
		// Error in scan.
		if err == sql.ErrNoRows {
			// User doesn't exist
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			// Other database related error.
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Compare password to hash in database, and conclude status.
	if comparePw(creds.Pw, storedCreds.Pw) {
		// Password incorrect.
		// Write JSON body for successful response return.
		w.WriteHeader(http.StatusOK)
		respMap[getJsonTag(&Credentials{}, "Id")] = strconv.FormatInt(int64(storedCreds.Id), 10)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Marshal JSON and insert it into the response.
	jsonResp, err := json.Marshal(respMap)
	if err != nil {
	} else {
		w.Write(jsonResp)
	}
}

/*
  Router function to sign up to website.
  Content type: application/json
  Success: 200, OK
  Failure: 400, bad request
*/
func signUp(w http.ResponseWriter, r *http.Request) {
	// Set up writer response.
	w.Header().Set("Content-Type", "application/json")

	// Get credentials from JSON request and validate them.
	creds := newUser()
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		// Bad request
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	validator.SetValidationFunc("validpw", validpw)
	if validator.Validate(*creds) != nil {
		// Bad credential semantics
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Register user to database.
	err = registerUser(creds)
	if err != nil {	
		// User registration error.
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Return code OK
		w.WriteHeader(http.StatusOK)
	}
}

// Register a user to the database.
func registerUser(creds *Credentials) error {
	// Check email uniqueness.
	err := checkUnique(TABLE_USERS, getDbTag(&Credentials{}, "Email"), creds.Email)
	if err != nil {
		return err
	}

	// Hash password and store new credentials to database.
	hash := hashPw(creds.Pw)

	// Make credentials insert statement for query.
	stmt := fmt.Sprintf(INSERT_FULL,
		TABLE_USERS,
		getDbTag(&Credentials{}, "Pw"),
		getDbTag(&Credentials{}, "Fname"),
		getDbTag(&Credentials{}, "Lname"),
		getDbTag(&Credentials{}, "Email"),
		getDbTag(&Credentials{}, "Usertype"),
		getDbTag(&Credentials{}, "PhoneNumber"),
		getDbTag(&Credentials{}, "Organization"))

	// Query full insert statement.
	_, err = db.Query(stmt,
		hash, creds.Fname, creds.Lname, creds.Email,
		creds.Usertype, creds.PhoneNumber, creds.Organization)
	if err != nil {
		return err
	} else {
		return nil
	}
}

// Set new user credentials
func newUser() *Credentials {
	return &Credentials{Usertype: USERTYPE_USER, PhoneNumber: "", Organization: ""}
}

// Check if a credential is unique in the database.
func checkUnique(table string, credType string, credVal string) error {
	// Prepare and query selection.
	stmt := fmt.Sprintf(SELECT_ROW, getDbTag(&Credentials{}, "Id"), table, credType)
	res := db.QueryRow(stmt, credVal)

	// Scan query and check for existing rows.
	resScan := &Credentials{}
	// Make columns interface for scan
	err := res.Scan(&resScan.Id)
	if err != sql.ErrNoRows {
		// Table isn't empty, error will be thrown.
		if err != nil {
			return err
		} else {
			return errors.New(credType + " already exists!")
		}
	} else {
		return nil
	}
}

// Checks if a password contains upper case, lower case, numbers, and special characters.
func validpw(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	if st.Kind() != reflect.String {
		return errors.New("Value must be string!")
	} else {
		// Set password and character number.
		pw := st.String()
		restrictions := []*regexp.Regexp{regexp.MustCompile("[a-z]"), // Must contain lowercase.
			regexp.MustCompile("^[" + ALPHANUMERICS + SPECIAL_CHARS + "]*$"), // Must contain only some characters.
			regexp.MustCompile("[A-Z]"), // Must contain uppercase.
			regexp.MustCompile("[0-9]"), // Must contain numerics.
			regexp.MustCompile("[" + SPECIAL_CHARS + "]")} // Must contain special characters.
		for _, restriction := range restrictions {
			if !restriction.MatchString(pw) {
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
