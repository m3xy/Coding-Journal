package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"testing"
)

// Test user log in.
func TestLogIn(t *testing.T) {
	// Set up test
	testInit()
	srv := testingServerSetup()

	// Start server.
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v\n", err)
		}
	}()

	// Populate database with valid users.
	trialUsers := getGlobalCopies(testUsers)
	for i := range trialUsers {
		trialUsers[i].ID, _ = registerUser(trialUsers[i].User)
	}

	// Test valid logins
	t.Run("Valid logins", func(t *testing.T) {
		for i := range testUsers {
			// Create a request for user login.
			loginMap := make(map[string]string)
			loginMap[getJsonTag(&User{}, "Email")] = testUsers[i].Email
			loginMap[JSON_TAG_PW] = testUsers[i].Password
			resp, err := sendJsonRequest(ENDPOINT_LOGIN, http.MethodPost, loginMap)
			assert.Nil(t, err, "Request should not error.")
			assert.Equalf(t, http.StatusOK, resp.StatusCode, "Response status should be %d", http.StatusOK)

			// Get ID from user response.
			respMap := make(map[string]string)
			err = json.NewDecoder(resp.Body).Decode(&respMap)
			assert.Nil(t, err, "Body unparsing should succeed")
			storedId, exists := respMap[getJsonTag(&User{}, "ID")]
			assert.True(t, exists, "ID should exist in response.")

			// Check if gotten
			assert.Equal(t, trialUsers[i].ID, storedId, "ID must equal registration's ID.")
		}
	})

	// Test invalid password login.
	t.Run("Invalid password logins", func(t *testing.T) {
		for i := 0; i < len(testUsers); i++ {
			loginMap := make(map[string]string)
			loginMap[getJsonTag(&User{}, "Email")] = testUsers[i].Email
			loginMap[JSON_TAG_PW] = VALID_PW // Ensure this pw is different from all test users.

			resp, err := sendJsonRequest(ENDPOINT_LOGIN, http.MethodPost, loginMap)
			assert.Nil(t, err, "Request should not error.")
			assert.Equalf(t, http.StatusUnauthorized, resp.StatusCode, "Response should have status %d", http.StatusUnauthorized)
		}
	})

	// Test invalid email login.
	t.Run("Invalid email logins", func(t *testing.T) {
		for i := 1; i < len(testUsers); i++ {
			loginMap := make(map[string]string)
			loginMap[getJsonTag(&User{}, "Email")] = testUsers[0].Email
			loginMap[JSON_TAG_PW] = testUsers[i].Password

			resp, err := sendJsonRequest(ENDPOINT_LOGIN, http.MethodPost, loginMap)
			assert.Nil(t, err, "Request should not error.")
			assert.Equalf(t, http.StatusUnauthorized, resp.StatusCode, "Response should have status %d", http.StatusUnauthorized)
		}
	})

	// Close server.
	if err := srv.Shutdown(context.Background()); err != nil {
		fmt.Printf("HTTP server shutdown: %v", err)
	}
	testEnd()
}
