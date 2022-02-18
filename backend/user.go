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
	user := r.PathPrefix(SUBROUTE_USER + "/{id}").Subrouter()

	user.HandleFunc("/", getUserProfile).Methods(http.MethodGet)
	user.HandleFunc(ENDPOINT_SUBMISSIONS, getAllAuthoredSubmissions).Methods(http.MethodGet)
}

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
	user := &GlobalUser{ID: vars["id"]}
	if res := gormDb. /*Preload("AuthoredSubmissions").Preload("ReviewedSubmissions").*/ Preload("User", getUserOutFromUser).Limit(1).Find(&user); res.Error != nil {
		log.Printf("[ERROR] SQL query error: %v", res.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if res.RowsAffected == 0 {
		log.Printf("[WARN] No user linked to %s", vars["id"])
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Encode user and send.
	err := json.NewEncoder(w).Encode(user)
	if err != nil {
		log.Printf("[ERROR] User data JSON encoding failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("[INFO] User credential request from %s successful.", r.RemoteAddr)
}
