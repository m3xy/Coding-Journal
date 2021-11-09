// === === === === === === === === === === === === ===
// authentication.go
// Set of all functions relating to user authentication,
// registration, and migration.
//
// Authors: 190014935
// Creation Date: 19/10/2021
// Last Modified: 04/11/2021
// === === === === === === === === === === === === ===

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
	HASH_COST     = 8
)

// ----
// User log in
// ----

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
	stmt := fmt.Sprintf(SELECT_ROW, "*", VIEW_LOGIN, getDbTag(&Credentials{}, "Email"))
	res := db.QueryRow(stmt, creds.Email)
	storedCreds := &Credentials{}
	err = res.Scan(&storedCreds.Id, &storedCreds.Email, &storedCreds.Pw)
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
		respMap[getJsonTag(&Credentials{}, "Id")] = strconv.Itoa(storedCreds.Id)
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

// ----
// User signup
// ----

/*
  Router function to sign up to website.
  Content type: application/json
  Success: 200, OK
  Failure: 400, bad request
*/
func signUp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get credentials from JSON request and validate them.
	creds := newUser()
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	validator.SetValidationFunc("validpw", validpw)
	if validator.Validate(*creds) != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = registerUser(creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Register a user to the database.
// Returns user global ID.
func registerUser(creds *Credentials) (int, error) {
	// Check email uniqueness.
	unique := checkUnique(TABLE_USERS, getDbTag(&Credentials{}, "Email"), creds.Email)
	if !unique {
		return -1, errors.New(getDbTag(&Credentials{}, "Email") + " is not unique!")
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
	res, err := db.Exec(stmt,
		hash, creds.Fname, creds.Lname, creds.Email,
		creds.Usertype, creds.PhoneNumber, creds.Organization)
	if err != nil {
		return -1, err
	}
	id, err := res.LastInsertId()
	return mapUserToGlobal(int(id))

}

// Register local user to global ID mappings.
func mapUserToGlobal(userId int) (int, error) {
	// Check if ID exists in users table.
	if checkUnique(TABLE_USERS, getDbTag(&Credentials{}, "Id"), strconv.Itoa(userId)) {
		return -1, errors.New("No user with this ID!")
	}

	// Check if ID is unique in idMappings table.
	if !checkUnique(TABLE_IDMAPPINGS, getDbTag(&IdMappings{}, "Id"), strconv.Itoa(userId)) {
		return -1, errors.New("ID already exists in ID mappings!")
	}

	// Set new global ID for user.
	globalId, _ := strconv.Atoi(TEAM_ID + strconv.Itoa(userId))

	// Insert new mapping to ID Mappings.
	stmt := fmt.Sprintf(INSERT_DOUBLE, TABLE_IDMAPPINGS, getDbTag(&IdMappings{}, "GlobalId"), getDbTag(&IdMappings{}, "Id"))
	_, err := db.Exec(stmt, globalId, userId)
	if err != nil {
		return -1, err
	} else {
		return globalId, nil
	}
}

// Set new user credentials
func newUser() *Credentials {
	return &Credentials{Usertype: USERTYPE_USER, PhoneNumber: "", Organization: ""}
}

// ----
// User exportation/importation
// ----

// Export user credentials. Exports all available details.
// TODO Testing for after MVP.
//
// Content type: application/json
// Success: 200, OK
// Failure: 401, Unauthorized
// Parameters: email, password
// Returns: {
// 	email, password, first name, last name, phone number, organisation, id (global)
// }
func exportUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")

	// Decode input into credentials and check necessary parameters.
	inputCreds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(inputCreds)
	if err != nil || inputCreds.Pw == "" || inputCreds.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// SELECT - FROM idMappings INNER JOIN users WHERE Email = ?
	stmt := fmt.Sprintf(SELECT_ROW,
		fmt.Sprintf("(%s, %s, %s, %s, %s, %s, %s)",
			getDbTag(&IdMappings{}, "GlobalId"),
			getDbTag(&Credentials{}, "Email"),
			getDbTag(&Credentials{}, "Pw"),
			getDbTag(&Credentials{}, "Fname"),
			getDbTag(&Credentials{}, "Lname"),
			getDbTag(&Credentials{}, "PhoneNumer"),
			getDbTag(&Credentials{}, "Organization")),
		fmt.Sprintf(INNER_JOIN, TABLE_IDMAPPINGS, TABLE_USERS),
		getDbTag(&Credentials{}, "Email"))
	query := db.QueryRow(stmt, inputCreds.Email)

	storedCreds := newUser()
	err = query.Scan(getCols(storedCreds)...)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println()
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			fmt.Printf("Error occured! %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Check password hash and proceed to export.
	if !comparePw(inputCreds.Pw, storedCreds.Pw) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	} else {
		storedCreds.Pw = inputCreds.Pw // Make sure not to send the hash to server.
		err = json.NewEncoder(w).Encode(storedCreds)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

// ----
// Password control
// ----

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
			regexp.MustCompile("[A-Z]"),                                      // Must contain uppercase.
			regexp.MustCompile("[0-9]"),                                      // Must contain numerics.
			regexp.MustCompile("[" + SPECIAL_CHARS + "]")}                    // Must contain special characters.
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
