package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// Test user info getter.
func TestGetUserProfile(t *testing.T) {
	testInit()
	defer testEnd()

	// Start mux router
	router := mux.NewRouter()
	router.HandleFunc(SUBROUTE_USER+"/{id}", getUserProfile)

	// Populate database for testing and test valid user.
	globalUsers := make([]GlobalUser, len(testUsers))
	for i := range testUsers {
		globalUsers[i].ID, _ = registerUser(testUsers[i], USERTYPE_USER)
	}

	t.Run("Valid user profiles", func(t *testing.T) {
		for i, u := range globalUsers {
			r, w := httptest.NewRequest(http.MethodGet, SUBROUTE_USER+"/"+u.ID, nil), httptest.NewRecorder()
			router.ServeHTTP(w, r)
			res := w.Result()

			assert.Equal(t, http.StatusOK, res.StatusCode, "Status should be OK.")

			resCreds := GlobalUser{}
			if err := json.NewDecoder(res.Body).Decode(&resCreds); !assert.Nil(t, err, "JSON decoding must not error.") {
				return
			}

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
		r, w := httptest.NewRequest(http.MethodGet, SUBROUTE_USER+"/"+INVALID_ID, nil), httptest.NewRecorder()
		router.ServeHTTP(w, r)
		res := w.Result()

		assert.Equalf(t, http.StatusNotFound, res.StatusCode, "Request should return status %d", http.StatusNotFound)
	})
}
