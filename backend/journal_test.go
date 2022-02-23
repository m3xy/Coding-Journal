package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// Test user log in.
func TestJournalLogIn(t *testing.T) {
	// Set up test
	testInit()
	defer testEnd()

	// Populate database with valid users.
	trialUsers := make([]GlobalUser, len(testUsers))
	for i, u := range testUsers {
		trialUsers[i] = GlobalUser{User: u.getCopy(), UserType: USERTYPE_REVIEWER_PUBLISHER}
		trialUsers[i].ID, _ = registerUser(trialUsers[i].User, USERTYPE_REVIEWER_PUBLISHER)
	}

	router := mux.NewRouter()
	router.HandleFunc(SUBROUTE_JOURNAL+ENDPOINT_LOGIN, logIn)

	// Test valid logins
	t.Run("Valid logins", func(t *testing.T) {
		for i := range testUsers {
			// Create a request for user login.
			loginBody := JournalLoginPostBody{Email: testUsers[i].Email, Password: testUsers[i].Password}
			reqBody, _ := json.Marshal(loginBody)
			req, w := httptest.NewRequest("POST", SUBROUTE_JOURNAL+ENDPOINT_LOGIN, bytes.NewBuffer(reqBody)), httptest.NewRecorder()
			router.ServeHTTP(w, req)
			resp := w.Result()

			assert.Equalf(t, http.StatusOK, resp.StatusCode, "Response status should be %d", http.StatusOK)

			// Get ID from user response.
			var respMap JournalLogInResponse
			if err := json.NewDecoder(resp.Body).Decode(&respMap); !assert.Nil(t, err, "Body unparsing should succeed") {
				return
			}
			// Check if gotten
			assert.Equal(t, trialUsers[i].ID, respMap.ID, "ID must equal registration's ID.")
		}
	})

	// Test invalid password login.
	t.Run("Invalid password logins", func(t *testing.T) {
		for i := 0; i < len(testUsers); i++ {
			loginMap := JournalLoginPostBody{Email: testUsers[i].Email, Password: VALID_PW}

			reqBody, _ := json.Marshal(loginMap)
			req, w := httptest.NewRequest("POST", SUBROUTE_JOURNAL+ENDPOINT_LOGIN, bytes.NewBuffer(reqBody)), httptest.NewRecorder()
			router.ServeHTTP(w, req)
			resp := w.Result()

			assert.Equalf(t, http.StatusUnauthorized, resp.StatusCode, "Response should have status %d", http.StatusUnauthorized)
		}
	})

	// Test invalid email login.
	t.Run("Invalid email logins", func(t *testing.T) {
		for i := 1; i < len(testUsers); i++ {
			loginMap := JournalLoginPostBody{Email: testUsers[0].Email, Password: testUsers[i].Password}

			reqBody, _ := json.Marshal(loginMap)
			req, w := httptest.NewRequest("POST", SUBROUTE_JOURNAL+ENDPOINT_LOGIN, bytes.NewBuffer(reqBody)), httptest.NewRecorder()
			router.ServeHTTP(w, req)
			resp := w.Result()

			assert.Equalf(t, http.StatusUnauthorized, resp.StatusCode, "Response should have status %d", http.StatusUnauthorized)
		}
	})
}
