package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// Set of all supergroup-appliant controllers and routes
// Authors: 190014935

/*
 Log in to website, check credentials correctness.
 Content type: application/json
 Success: 200, Credentials are correct
 Failure: 401, Unauthorized
 Returns: userId
*/
func logIn(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Received log in request from %v", r.RemoteAddr)

	// Set up writer response.
	w.Header().Set("Content-Type", "application/json")

	// Get credentials from log in request.
	user := &User{}
	err := json.NewDecoder(r.Body).Decode(user)
	if err != nil {
		log.Printf("[ERROR] JSON decoder failed on log in.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get credentials at given email, and assign it.
	var globalUser GlobalID
	res := gormDb.Joins("User").Where("User.Email = ?", user.Email).Limit(1).Find(&globalUser)
	if res.RowsAffected == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("[INFO] Incorrect email: %s", user.Email)
		return
	} else if err := res.Error; err != nil {
		log.Printf("[ERROR] SQL query failure on login: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Compare password to hash in database, and conclude status.
	if !comparePw(user.Password, globalUser.User.Password) {
		log.Printf("[INFO] Given password and password registered on %s do not match.", user.Email)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Marshal JSON and insert it into the response.
	jsonResp, _ := json.Marshal(map[string]string{getJsonTag(&User{}, "ID"): globalUser.ID})
	w.Write(jsonResp)
	log.Printf("[INFO] log in from %s at email %s successful.", r.RemoteAddr, user.Email)
}
