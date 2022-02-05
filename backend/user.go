package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func getUserOutFromUser(tx *gorm.DB) *gorm.DB {
	return tx.Select("GlobalUserID", "Email", "FirstName", "LastName", "UserType", "PhoneNumber", "Organization", "CreatedAt")
}

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
	user := &GlobalUser{ID: vars[getJsonTag(&GlobalUser{}, "ID")]}
	if res := gormDb.Preload("User", getUserOutFromUser).Limit(1).Find(&user); res.Error != nil {
		log.Printf("[ERROR] SQL query error: %v", res.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if res.RowsAffected == 0 {
		log.Printf("[WARN] No user linked to %s", vars[getJsonTag(&GlobalUser{}, "ID")])
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

	// Encode user and send.
	err := json.NewEncoder(w).Encode(user)
	if err != nil {
		log.Printf("[ERROR] User data JSON encoding failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("[INFO] User credential request from %s successful.", r.RemoteAddr)
}
