package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Set of all supergroup-appliant controllers and routes
// Authors: 190014935

func getJournalSubroute(r *mux.Router) {
	r.Use(journalMiddleWare)
	r.HandleFunc(ENDPOINT_LOGIN, logIn).Methods(http.MethodPost, http.MethodOptions)
}

// Validate if given security token works.
// Params:
// 	Header: securityToken
// Return:
//  200: Success - security token valid.
//  401: Failure - security token invalid.
func tokenValidation(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Token validation from %v successful.", r.RemoteAddr)
}

/*
 Log in to website, check credentials correctness.
 Content type: application/json
 Success: 200, Credentials are correct
 Failure: 401, Unauthorized
 Returns: userId
*/
func logIn(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Received log in request from %v", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")

	// Get credentials from log in request.
	user := JournalLoginPostBody{}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Printf("[ERROR] JSON decoder failed on log in.")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Get User ID from local credentials check.
	uuid, status := GetLocalUserID(user)
	if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}

	// Marshal JSON and insert it into the response.
	if err := json.NewEncoder(w).Encode(JournalLogInResponse{ID: uuid}); err != nil {
		log.Printf("[ERROR] JSON Response Encoding failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	log.Printf("[INFO] log in from %s at email %s successful.", r.RemoteAddr, user.Email)
}
