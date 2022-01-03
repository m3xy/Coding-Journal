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
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"regexp"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/validator.v2"
)

const (
	SPECIAL_CHARS = "//!//@//#//$//%//^//&//*//,//.//;//://_//-//+//-//=//\"//'"
	A_NUMS        = "a-zA-Z0-9"
	HASH_COST     = 8
)

// ----
// User log in
// ----

/*
 Log in to website, check credentials correctness.
 Content type: application/json
 Success: 200, Credentials are correct
 Failure: 401, Unauthorized
 Returns: userId
*/
func logIn(w http.ResponseWriter, r *http.Request) {
	if useCORSresponse(&w, r); r.Method == http.MethodOptions {
		return
	}
	if !validateToken(r.Header.Get(SECURITY_TOKEN_KEY)) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	log.Printf("[INFO] Received log in request from %v", r.RemoteAddr)

	// Set up writer response.
	w.Header().Set("Content-Type", "application/json")

	// Get credentials from log in request.
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		log.Printf("[ERROR] JSON decoder failed on log in.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get credentials at given email, and assign it.
	storedCreds := &Credentials{}
	stmt := fmt.Sprintf(SELECT_ROW, "*", VIEW_LOGIN, getDbTag(&Credentials{}, "Email"))
	err = db.QueryRow(stmt, creds.Email).Scan(&storedCreds.Id, &storedCreds.Email, &storedCreds.Pw)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
			log.Printf("[INFO] Incorrect email: %s", creds.Email)
		} else {
			fmt.Println(err)
			log.Printf("[ERROR] SQL query failure on login: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Compare password to hash in database, and conclude status.
	if !comparePw(creds.Pw, storedCreds.Pw) {
		log.Printf("[INFO] Given password and password registered on %s do not match.", creds.Email)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Marshal JSON and insert it into the response.
	jsonResp, err := json.Marshal(map[string]string{getJsonTag(&Credentials{}, "Id"): storedCreds.Id})
	if err != nil {
	} else {
		w.Write(jsonResp)
	}
	log.Printf("[INFO] log in from %s at email %s successful.", r.RemoteAddr, creds.Email)
}

/*
	Get user profile info for a user.
	Content type: application/json
	Success: 200, Credentials can be passed down.
	Failure: 404, User not found.
*/
func getUserProfile(w http.ResponseWriter, r *http.Request) {
	if useCORSresponse(&w, r); r.Method == http.MethodOptions {
		return
	}
	// Check security token.
	if !validateToken(r.Header.Get(SECURITY_TOKEN_KEY)) {
		log.Printf("[WARN] Invalid security token received from %s.", r.RemoteAddr)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	log.Printf("[INFO] Received user credential request from %s", r.RemoteAddr)

	// Get user from parameters.
	vars := mux.Vars(r)
	if checkUnique(TABLE_IDMAPPINGS,
		getDbTag(&IdMappings{}, "GlobalId"), vars[getJsonTag(&Credentials{}, "Id")]) {
		log.Printf("[WARN] User (%s) not found.", vars[getJsonTag(&Credentials{}, "Id")])
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Get user details from user ID.
	info := &Credentials{Pw: ""}
	err := db.QueryRow(fmt.Sprintf(SELECT_ROW, "*", VIEW_USER_INFO,
		getDbTag(&IdMappings{}, "GlobalId")), vars[getJsonTag(&Credentials{}, "Id")]).
		Scan(&info.Id, &info.Email, &info.Fname, &info.Lname, &info.Usertype,
			&info.PhoneNumber, &info.Organization)
	if err != nil {
		log.Printf("[ERROR] SQL query error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get map of submission IDs to submission names.
	submissionsMap, err := getUserSubmissions(vars[getJsonTag(&Credentials{}, "Id")])
	if err != nil {
		log.Printf("[ERROR] Failed to retrieve user (%s)'s projects: %v'", vars[getJsonTag(&Credentials{}, "Id")], err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	buffMap := map[string]interface{}{
		getJsonTag(&Credentials{}, "Email"):        info.Email,
		getJsonTag(&Credentials{}, "Fname"):        info.Fname,
		getJsonTag(&Credentials{}, "Lname"):        info.Lname,
		getJsonTag(&Credentials{}, "Usertype"):     info.Usertype,
		getJsonTag(&Credentials{}, "PhoneNumber"):  info.PhoneNumber,
		getJsonTag(&Credentials{}, "Organization"): info.Organization,
		"submissions": submissionsMap,
	}
	err = json.NewEncoder(w).Encode(buffMap)
	if err != nil {
		log.Printf("[ERROR] User data JSON encoding failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("[INFO] User credential request from %s successful.", r.RemoteAddr)
}

/*
	Log in to website with any server's database.
	Content type: application/json
	Input: {"email": string, "password": string, "groupNumber": int}
	Success: 200, Credentials are correct.
	Failure: 401, Unauthorized
	Returns: userId
*/
func logInGlobal(w http.ResponseWriter, r *http.Request) {
	if useCORSresponse(&w, r); r.Method == http.MethodOptions {
		return
	}
	// Check validity
	log.Println("Global login request sent!")
	if !validateToken(r.Header.Get(SECURITY_TOKEN_KEY)) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	log.Printf("[INFO] Received global login request from %s.", r.RemoteAddr)
	propsMap := make(map[string]string)
	err := json.NewDecoder(r.Body).Decode(&propsMap)
	if err != nil {
		log.Printf("[WARN] Invalid security token received from %s.", r.RemoteAddr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Query path from team ID.
	retServer := &Servers{}
	stmt := fmt.Sprintf(SELECT_ROW, "*", TABLE_SERVERS, getDbTag(&Servers{}, "GroupNb"))
	err = db.QueryRow(stmt, propsMap[getJsonTag(&Servers{}, "GroupNb")]).Scan(getCols(retServer)...)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[WARN] Group number %s doesn't exist in database.", propsMap[getJsonTag(&Servers{}, "GroupNumber")])
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			log.Printf("[ERROR] SQL query error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Make request from given URL and security token
	jsonBody, err := json.Marshal(propsMap)
	if err != nil {
		log.Printf("[ERROR] JSON body encoding failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	globalReq, _ := http.NewRequest(
		"POST", retServer.Url+"/login", bytes.NewBuffer(jsonBody))
	globalReq.Header.Set(SECURITY_TOKEN_KEY, retServer.Token)

	// Get response from login request.
	res, err := sendSecureRequest(globalReq, propsMap[getJsonTag(&Servers{}, "GroupNb")])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("[ERROR] HTTP Request error: %v", err)
		return
	} else if res.StatusCode != http.StatusOK {
		log.Printf("[WARN] Foreign server login request failed, mirroring...")
		w.WriteHeader(res.StatusCode)
		return
	}
	err = json.NewDecoder(res.Body).Decode(&propsMap)
	if err != nil {
		log.Printf("[ERROR] JSON decoding error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(&propsMap)
	if err != nil {
		log.Printf("[ERROR] JSON encoding error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
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
	if useCORSresponse(&w, r); r.Method == http.MethodOptions {
		return
	}
	log.Println("Sign up request sent!")
	if !validateToken(r.Header.Get(SECURITY_TOKEN_KEY)) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	log.Printf("[INFO] Received sign up request from %s.", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")

	// Get credentials from JSON request and validate them.
	creds := newUser()
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		log.Printf("[ERROR] JSON decoding failed: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	validator.SetValidationFunc("validpw", validpw)
	if validator.Validate(*creds) != nil {
		log.Printf("[WARN] Invalid password format received.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = registerUser(creds)
	if err != nil {
		log.Printf("[ERROR] User registration failed: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("[INFO] User signup from %s successful.", r.RemoteAddr)
	w.WriteHeader(http.StatusOK)
}

// Register a user to the database.
// Returns user global ID.
func registerUser(creds *Credentials) (string, error) {
	// Check email uniqueness.
	unique := checkUnique(TABLE_USERS, getDbTag(&Credentials{}, "Email"), creds.Email)
	if !unique {
		return "", errors.New(getDbTag(&Credentials{}, "Email") + " is not unique!")
	}

	// Hash password and store new credentials to database.
	hash := hashPw(creds.Pw)

	// Make credentials insert statement for query.
	stmt := fmt.Sprintf(INSERT_FULL, TABLE_USERS,
		getDbTag(&Credentials{}, "Email"),
		getDbTag(&Credentials{}, "Pw"),
		getDbTag(&Credentials{}, "Fname"),
		getDbTag(&Credentials{}, "Lname"),
		getDbTag(&Credentials{}, "Usertype"),
		getDbTag(&Credentials{}, "PhoneNumber"),
		getDbTag(&Credentials{}, "Organization"))

	// Query full insert statement.
	_, err := db.Exec(stmt, creds.Email, hash, creds.Fname, creds.Lname,
		creds.Usertype, creds.PhoneNumber, creds.Organization)
	if err != nil {
		return "", err
	}

	// Get new UUID from query
	err = db.QueryRow(
		fmt.Sprintf(SELECT_ROW, getDbTag(&Credentials{}, "Id"),
			TABLE_USERS, getDbTag(&Credentials{}, "Email")), creds.Email).
		Scan(&creds.Id)
	// Query UUID
	query := db.QueryRow(fmt.Sprintf(SELECT_ROW, getDbTag(&Credentials{}, "Id"),
		TABLE_USERS, getDbTag(&Credentials{}, "Email")), creds.Email)
	err = query.Scan(&creds.Id)
	if err != nil {
		return "", err
	}
	return mapUserToGlobal(creds.Id)

}

// Register local user to global ID mappings.
func mapUserToGlobal(userId string) (string, error) {
	// Check if ID exists in users, and is unique in idMappings.
	if checkUnique(TABLE_USERS, getDbTag(&Credentials{}, "Id"), userId) {
		return "", errors.New("No user with this ID!")
	} else if !checkUnique(TABLE_IDMAPPINGS, getDbTag(&IdMappings{}, "Id"), userId) {
		return "", errors.New("ID already exists in ID mappings!")
	}

	// Insert new mapping to ID Mappings.
	idMap := &IdMappings{Id: userId, GlobalId: TEAM_ID + userId}

	// Insert new mapping to ID Mappings.
	stmt := fmt.Sprintf(INSERT_DOUBLE, TABLE_IDMAPPINGS, getDbTag(&IdMappings{}, "GlobalId"), getDbTag(&IdMappings{}, "Id"))
	_, err := db.Query(stmt, idMap.GlobalId, idMap.Id)
	if err != nil {
		return "", err
	} else {
		return idMap.GlobalId, nil
	}
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
	if useCORSresponse(&w, r); r.Method == http.MethodOptions {
		return
	}
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
	}
	w.WriteHeader(http.StatusOK)
}

// ----
// Password control
// ----

// Set new user credentials
func newUser() *Credentials {
	// return &Credentials{Usertype: USERTYPE_USER, PhoneNumber: "", Organization: ""}
	// TODO: fix permissions later
	return &Credentials{Usertype: USERTYPE_REVIEWER_PUBLISHER, PhoneNumber: "", Organization: ""}
}

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
