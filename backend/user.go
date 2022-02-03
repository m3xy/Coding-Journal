package main

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
)

/*
	Get user profile info for a user.
	Content type: application/json
	Success: 200, Credentials can be passed down.
	Failure: 404, User not found.
*/
func getUserProfile(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Received user credential request from %s", r.RemoteAddr)

	// Get user details from user ID.
	vars := mux.Vars(r)
	var user User
	if err := gormDb.Model(&GlobalUser{ID: vars[getJsonTag(&User{}, "ID")]}).Association("User").Find(&user); err != nil {
		log.Printf("[ERROR] SQL query error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if reflect.DeepEqual(user, User{}) {
		log.Printf("[WARN] No user linked to %s", vars[getJsonTag(&User{}, "ID")])
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// Get map of submission IDs to submission names.
	/* submissionsMap, err := getUserSubmissions(vars[getJsonTag(&Credentials{}, "Id")])
	if err != nil {
		log.Printf("[ERROR] Failed to retrieve user (%s)'s projects: %v'", vars[getJsonTag(&Credentials{}, "Id")], err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} */

	// Remove private data and encode user.
	user.Password = ""
	user.ID = vars[getJsonTag(&User{}, "ID")]
	user.GlobalUserID = ""
	err := json.NewEncoder(w).Encode(user)
	if err != nil {
		log.Printf("[ERROR] User data JSON encoding failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("[INFO] User credential request from %s successful.", r.RemoteAddr)
}
