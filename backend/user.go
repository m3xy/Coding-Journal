package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

const (
	SUBROUTE_USERS = "/users"
	SUBROUTE_USER  = "/user"
	ENDPOINT_GET   = "/get"
)

func getUserSubroutes(r *mux.Router) {
	user := r.PathPrefix(SUBROUTE_USER).Subrouter()

	// User routes:
	// + GET /user/{id} - Get given user profile.
	user.HandleFunc("/{id}", getUserProfile).Methods(http.MethodGet)
}

func getUserOutFromUser(tx *gorm.DB) *gorm.DB {
	return tx.Select("GlobalUserID", "Email", "FirstName", "LastName", "PhoneNumber", "Organization", "CreatedAt")
}

/*
	Get user profile info for a user.
	Content type: application/json
	Success: 200, Credentials can be passed down.
	Failure: 404, User not found.
*/
func getUserProfile(w http.ResponseWriter, r *http.Request) {
	// Get user details from user ID.
	vars := mux.Vars(r)
	user := &GlobalUser{ID: vars["id"]}
	if res := gormDb.Preload("AuthoredSubmissions").Preload("User", getUserOutFromUser).Limit(1).Find(&user); res.Error != nil {
		log.Printf("[ERROR] SQL query error: %v", res.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if res.RowsAffected == 0 {
		log.Printf("[WARN] No user linked to %s", vars["id"])
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Encode user and send.
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("[ERROR] User data JSON encoding failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
