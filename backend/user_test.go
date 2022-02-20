package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const (
	TEST_PORT_USERS = ":59219"
)

// Set up server used for user testing.
func userServerSetup() *http.Server {
	router := mux.NewRouter()
	getUserSubroutes(router)

	return &http.Server{
		Addr:    TEST_PORT_USERS,
		Handler: router,
	}
}

// Test user info getter.
func TestGetUserProfile(t *testing.T) {
	testInit()
	srv := userServerSetup()

	// Start server.
	go srv.ListenAndServe()

	// Populate database for testing and test valid user.
	globalUsers := make([]GlobalUser, len(testUsers))
	for i := range testUsers {
		globalUsers[i].ID, _ = registerUser(testUsers[i], USERTYPE_USER)
	}

	t.Run("Valid user profiles", func(t *testing.T) {
		for i, u := range globalUsers {
			res, err := sendJsonRequest(SUBROUTE_USER+"/"+u.ID, http.MethodGet, nil, TEST_PORT_USERS)
			assert.Nil(t, err, "Request should not error.")
			assert.Equal(t, http.StatusOK, res.StatusCode, "Status should be OK.")

			resCreds := GlobalUser{}
			err = json.NewDecoder(res.Body).Decode(&resCreds)
			assert.Nil(t, err, "JSON decoding must not error.")

			// Check equality for all user info.
			equal := (testUsers[i].Email == resCreds.User.Email) &&
				(testUsers[i].FirstName == resCreds.User.FirstName) &&
				(testUsers[i].LastName == resCreds.User.LastName) &&
				(testUsers[i].PhoneNumber == resCreds.User.PhoneNumber) &&
				(testUsers[i].Organization == resCreds.User.Organization)
			assert.Equal(t, true, equal, "Users should be equal.")
		}
	})

	// Test invalid users.
	t.Run("Invalid user profile", func(t *testing.T) {
		res, err := sendJsonRequest(SUBROUTE_USER+"/"+INVALID_ID, http.MethodGet, nil, TEST_PORT_USERS)
		assert.Nil(t, err, "Request should not error.")
		assert.Equalf(t, http.StatusNotFound, res.StatusCode, "Request should return status %d", http.StatusNotFound)
	})

	// Close server.
	if err := srv.Shutdown(context.Background()); err != nil {
		fmt.Printf("HTTP server shutdown: %v", err)
	}
	testEnd()
}
