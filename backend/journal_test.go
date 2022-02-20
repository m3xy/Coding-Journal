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
	TEST_PORT_JOURNAL = ":59214"
)

// Set up server used for journal testing.
func journalServerSetup() *http.Server {
	router := mux.NewRouter()
	router.Use(journalMiddleWare)

	journal := router.PathPrefix(SUBROUTE_JOURNAL).Subrouter()
	journal.HandleFunc(ENDPOINT_LOGIN, logIn).Methods(http.MethodPost, http.MethodOptions)
	journal.HandleFunc(ENDPOINT_VALIDATE, tokenValidation).Methods(http.MethodGet)

	// Setup testing HTTP server
	return &http.Server{
		Addr:    TEST_PORT_JOURNAL,
		Handler: router,
	}
}

// Test security key validation.
func TestTokenValidation(t *testing.T) {
	testInit()
	srv := journalServerSetup()

	// Start server.
	go srv.ListenAndServe()

	// Write valid security token response
	t.Run("Valid token validation", func(t *testing.T) {
		validReq, _ := http.NewRequest("GET", LOCALHOST+TEST_PORT_JOURNAL+SUBROUTE_JOURNAL+ENDPOINT_VALIDATE, nil)
		res, err := sendSecureRequest(gormDb, validReq, TEAM_ID)
		if err != nil {
			t.Errorf("HTTP request error: %v\n", err)
		} else if res.StatusCode != http.StatusOK {
			t.Errorf("Response Status code should be OK, but is %d", res.StatusCode)
		}
	})

	// Write invalid security token response
	t.Run("Invalid token validation", func(t *testing.T) {
		client := http.Client{}
		invalidReq, _ := http.NewRequest("GET", LOCALHOST+TEST_PORT_JOURNAL+SUBROUTE_JOURNAL+ENDPOINT_VALIDATE, nil)
		invalidReq.Header.Set(SECURITY_TOKEN_KEY, WRONG_SECURITY_TOKEN)
		res, err := client.Do(invalidReq)
		if err != nil {
			t.Errorf("HTTP request error: %v\n", err)
		} else if res.StatusCode != http.StatusUnauthorized {
			t.Errorf("Response Status code should be 401, but is %d", res.StatusCode)
		}
	})
}

// Test user log in.
func TestJournalLogIn(t *testing.T) {
	// Set up test
	testInit()
	srv := journalServerSetup()

	// Start server.
	go srv.ListenAndServe()

	// Populate database with valid users.
	trialUsers := getGlobalCopies(testUsers)
	for i := range trialUsers {
		trialUsers[i].ID, _ = registerUser(trialUsers[i].User, USERTYPE_REVIEWER_PUBLISHER)
	}

	// Test valid logins
	t.Run("Valid logins", func(t *testing.T) {
		for i := range testUsers {
			// Create a request for user login.
			loginMap := make(map[string]string)
			loginMap[getJsonTag(&User{}, "Email")] = testUsers[i].Email
			loginMap[JSON_TAG_PW] = testUsers[i].Password
			resp, err := sendJsonRequest(SUBROUTE_JOURNAL+ENDPOINT_LOGIN, http.MethodPost, loginMap, TEST_PORT_JOURNAL)
			assert.Nil(t, err, "Request should not error.")
			assert.Equalf(t, http.StatusOK, resp.StatusCode, "Response status should be %d", http.StatusOK)

			// Get ID from user response.
			respMap := make(map[string]string)
			err = json.NewDecoder(resp.Body).Decode(&respMap)
			assert.Nil(t, err, "Body unparsing should succeed")
			storedId, exists := respMap[getJsonTag(&JournalLogInResponse{}, "ID")]
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

			resp, err := sendJsonRequest(SUBROUTE_JOURNAL+ENDPOINT_LOGIN, http.MethodPost, loginMap, TEST_PORT_JOURNAL)
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

			resp, err := sendJsonRequest(SUBROUTE_JOURNAL+ENDPOINT_LOGIN, http.MethodPost, loginMap, TEST_PORT_JOURNAL)
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
