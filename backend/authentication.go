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
	"encoding/json"
	"errors"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"
	"log"
	"net/http"
)

const (
	SPECIAL_CHARS = "//!//@//#//$//%//^//&//*//,//.//;//://_//-//+//-//=//\"//'"
	A_NUMS        = "a-zA-Z0-9"
)

// ----
// User log in
// ----
/*
	Log in to website with any server's database.
	Content type: application/json
	Input: {"email": string, "password": string, "groupNumber": int}
	Success: 200, Credentials are correct.
	Failure: 401, Unauthorized
	Returns: userId
*/
func logInGlobal(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Received global login request from %s.", r.RemoteAddr)
	propsMap := make(map[string]string)
	err := json.NewDecoder(r.Body).Decode(&propsMap)
	if err != nil {
		log.Printf("[WARN] Invalid security token received from %s.", r.RemoteAddr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Query path from team ID.
	var retServer Server
	res := gormDb.Limit(1).Find(&retServer, propsMap[getJsonTag(&Server{}, "GroupNumber")])
	if res.RowsAffected == 0 {
		log.Printf("[WARN] Group number %s doesn't exist in database.", propsMap[getJsonTag(&Server{}, "GroupNumber")])
		w.WriteHeader(http.StatusUnauthorized)
		return
	} else if res.Error != nil {
		log.Printf("[ERROR] SQL query error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
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
	client := &http.Client{}
	globalReq.Header.Set(SECURITY_TOKEN_KEY, retServer.Token)
	foreignRes, err := client.Do(globalReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("[ERROR] HTTP Request error: %v", err)
		return
	} else if foreignRes.StatusCode != http.StatusOK {
		log.Printf("[WARN] Foreign server login request failed, mirroring...")
		w.WriteHeader(foreignRes.StatusCode)
		return
	}

	mirrorProps := make(map[string]string)
	err = json.NewDecoder(foreignRes.Body).Decode(&mirrorProps)
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
	log.Printf("[INFO] Received sign up request from %s.", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")

	// Get credentials from JSON request and validate them.
	user := &User{}
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		log.Printf("[ERROR] JSON decoding failed: %v", err)
		goto ERROR
	}
	if validator.SetValidationFunc("validpw", validpw); validator.Validate(*user) != nil {
		log.Printf("[WARN] Invalid password format received.")
		goto ERROR
	}

	if _, err := registerUser(*user); err != nil {
		log.Printf("[ERROR] User registration failed: %v", err)
		goto ERROR
	}
	log.Printf("[INFO] User signup from %s successful.", r.RemoteAddr)
	w.WriteHeader(http.StatusOK)
	return

ERROR:
	w.WriteHeader(http.StatusBadRequest)
	return
}

// Register a user to the database. Returns user global ID.
func registerUser(user User) (string, error) {
	// Hash password and store new credentials to database.
	user.Password = string(hashPw(user.Password))

	registeredUser := GlobalUser{
		FullName: user.FirstName+" "+user.LastName,
		User: user,
	}
	if err := gormDb.Transaction(func(tx *gorm.DB) error {
		// Check constraints on user
		if !isUnique(tx, User{}, "Email", user.Email) {
			return errors.New("Email already taken!")
		}

		// Make credentials insert transaction.
		if err := gormDb.Create(&registeredUser).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return "", err
	}

	// Return user's primary key (the UUID)
	return registeredUser.ID, nil
}
