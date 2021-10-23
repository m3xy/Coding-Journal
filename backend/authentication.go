package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/validator.v2"
)

const (
	SELECT_ROW  = "SELECT $1 FROM $2 WHERE $3 = $4"
	INSERT_CRED = "INSERT INTO $1 ($2, $3, $4, $5, $6) VALUES ($7, $8, $9, $10)"
)

// Credentials - for sign up.
type Credentials struct {
	Username string `json:"username" db:"username" validate:"nonzero"`
	Pw       string `json:"password" db:"password" validate:"min=8,max=64,validpw"`
	Fname    string `json:"firstname" db:"firstname" validate:"nonzero"`
	Lname    string `json:"lastname" db:"lastname" validate:"nonzero"`
	Email    string `json:"email" db:"email" validate:"nonzero"`
}

// Router function to log in to website.
func logIn(w http.ResponseWriter, r *http.Request) {
	// Get credentials from log in request.
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	// Get password at given username, and assign it.
	res := db.QueryRow(SELECT_ROW, "password", "users", "username", creds.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	storedCreds := &Credentials{}
	err = res.Scan(storedCreds)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	// Compare password to hash in database, and conclude status.
	if !comparePw(creds.Pw, storedCreds.Pw) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Successfully logged in!")
	}
}

// Router function to sign up to website.
func signUp(w http.ResponseWriter, r *http.Request) {
	// Get credentials from JSON request and validate them.
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil { // Bad request
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	validator.SetValidationFunc("validpw", validpw)
	if validator.Validate(*creds) != nil { // Bad credential semantics
		w.WriteHeader(http.StatusBadRequest)
	}

	// Register user to database.
	err = registerUser(creds)
	if err != nil { // User registration error.
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err.Error())
	} else {
		fmt.Fprintln(w, "Sign-up successful!")
		w.WriteHeader(http.StatusOK)
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
			regexp.MustCompile("[A-Z]"), // Must contain uppercase.
			regexp.MustCompile("[0-9]"), // Must contain numerics.
			regexp.MustCompile("[$-!]")} // Must contain special characters.
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
	hash, _ := bcrypt.GenerateFromPassword([]byte(pw), 8)
	return hash
}

func comparePw(pw string, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

// Register a user to the database.
func registerUser(creds *Credentials) error {
	// Check username and email uniqueness.
	err := checkUnique("users", creds.Username, "username")
	if err != nil {
		return err
	}
	err = checkUnique("users", creds.Email, "email")
	if err != nil {
		return err
	}

	// Hash password and store new credentials to database.
	hash := hashPw(creds.Pw)

	_, err = db.Query(INSERT_CRED, "users", "username", "password", "firstname", "lastname", "email",
		creds.Username, hash, creds.Fname, creds.Lname, creds.Email)
	if err != nil {
		return err
	} else {
		return nil
	}
}

// Check if a credential is unique in the database.
func checkUnique(table string, cred string, credtype string) error {
	res := db.QueryRow(SELECT_ROW, "password", table, credtype, cred)
	resScan := &Credentials{}
	err := res.Scan(resScan)
	if err != sql.ErrNoRows {
		if err != nil {
			return err
		} else {
			return errors.New(credtype + " already exists!")
		}
	} else {
		return nil
	}
}
