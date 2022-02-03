package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

// Test user info getter.
func TestGetUserProfile(t *testing.T) {
	testInit()
	srv := testingServerSetup()

	// Start server.
	go srv.ListenAndServe()

	// Populate database for testing and test valid user.
	globalUsers := make([]GlobalUser, len(testUsers))
	for i := range testUsers {
		globalUsers[i].ID, _ = registerUser(testUsers[i])
	}

	t.Run("Valid user profiles", func(t *testing.T) {
		for i, u := range globalUsers {
			res, err := sendJsonRequest(ENDPOINT_USERINFO+"/"+u.ID, http.MethodGet, nil)
			assert.Nil(t, err, "Request should not error.")
			assert.Equal(t, http.StatusOK, res.StatusCode, "Status should be OK.")

			resCreds := User{}
			err = json.NewDecoder(res.Body).Decode(&resCreds)
			assert.Nil(t, err, "JSON decoding must not error.")

			// Check equality for all user info.
			equal := (testUsers[i].Email == resCreds.Email) &&
				(testUsers[i].FirstName == resCreds.FirstName) &&
				(testUsers[i].LastName == resCreds.LastName) &&
				(testUsers[i].PhoneNumber == resCreds.PhoneNumber) &&
				(testUsers[i].Organization == resCreds.Organization)

			assert.Equal(t, true, equal, "Users should be equal.")
		}
	})

	// Test invalid users.
	t.Run("Invalid user profile", func(t *testing.T) {
		res, err := sendJsonRequest(ENDPOINT_USERINFO+"/"+INVALID_ID, http.MethodGet, nil)
		assert.Nil(t, err, "Request should not error.")
		assert.Equalf(t, http.StatusNotFound, res.StatusCode, "Request should return status %d", http.StatusNotFound)
	})

	// Close server.
	if err := srv.Shutdown(context.Background()); err != nil {
		fmt.Printf("HTTP server shutdown: %v", err)
	}
	testEnd()
}
