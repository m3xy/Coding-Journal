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
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"
)

const (
	SPECIAL_CHARS = "//!//@//#//$//%//^//&//*//,//.//;//://_//-//+//-//=//\"//'"
	A_NUMS        = "a-zA-Z0-9"
)

// ----
// User log in
// ----

/*
	Get user profile info for a user.
	Content type: application/json
	Success: 200, Credentials can be passed down.
	Failure: 404, User not found.
*/
func getUserProfile(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Received user credential request from %s", r.RemoteAddr)

	// Get user from parameters.
	vars := mux.Vars(r)
	/* if checkUnique(TABLE_IDMAPPINGS,
		getDbTag(&User{}, "GlobalId"), vars[getJsonTag(&User{}, "ID")]) {
		log.Printf("[WARN] User (%s) not found.", vars[getJsonTag(&User{}, "ID")])
		w.WriteHeader(http.StatusNotFound)
		return
	} */
	if isUnique(gormDb, &GlobalID{}, "ID", vars[getJsonTag(&User{}, "ID")]) {
		log.Printf("[WARN] User (%s) not found.", vars[getJsonTag(&User{}, "ID")])
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Get user details from user ID.
	var info GlobalID
	if res := gormDb.Joins("User").Where("global_ids.id = ?", vars[getJsonTag(&User{}, "ID")]).Find(&info); res.Error != nil {
		log.Printf("[ERROR] SQL query error: %v", res.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if res.RowsAffected == 0 {
		log.Printf("[WARN] No user found with that ID.")
		return
	}

	// Get map of submission IDs to submission names.
	/* submissionsMap, err := getUserSubmissions(vars[getJsonTag(&Credentials{}, "Id")])
	if err != nil {
		log.Printf("[ERROR] Failed to retrieve user (%s)'s projects: %v'", vars[getJsonTag(&Credentials{}, "Id")], err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} */
	info.User.Password = "" // Remove password - ensure it isn't passed around.
	err := json.NewEncoder(w).Encode(info.User)
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
	r.Header.Set(SECURITY_TOKEN_KEY, retServer.Token)
	foreignRes, err := client.Do(r)
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
  Content type: application/json Success: 200, OK Failure: 400, bad request
*/
func signUp(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Received sign up request from %s.", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")

	// Get credentials from JSON request and validate them.
	user := &User{}
	err := json.NewDecoder(r.Body).Decode(user)
	if err != nil {
		log.Printf("[ERROR] JSON decoding failed: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	validator.SetValidationFunc("validpw", validpw)
	if validator.Validate(*user) != nil {
		log.Printf("[WARN] Invalid password format received.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = registerUser(*user)
	if err != nil {
		log.Printf("[ERROR] User registration failed: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("[INFO] User signup from %s successful.", r.RemoteAddr)
	w.WriteHeader(http.StatusOK)
}

// Register a user to the database. Returns user global ID.
func registerUser(user User) (string, error) {
	// Hash password and store new credentials to database.
	user.Password = string(hashPw(user.Password))

	// Make credentials insert transaction.
	if err := gormDb.Transaction(func(tx *gorm.DB) error {
		user.ID = uuid.NewV4().String()
		if err := tx.Create(&user).Error; err != nil {
			return err
		} else if err := tx.Create(&GlobalID{ID: (strconv.Itoa(TEAM_ID) + user.ID), UserID: user.ID}).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return "", err
	}

	// Return user's primary key (the UUID)
	return strconv.Itoa(TEAM_ID) + user.ID, nil
}
