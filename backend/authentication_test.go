package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/validator.v2"
)

const (
	VALID_PW       = "aB12345$"
	PW_NO_UC       = "a123456$"
	PW_NO_LC       = "B123456$"
	PW_NO_NUM      = "aBcdefg$"
	PW_NO_SC       = "aB123456"
	PW_WRONG_CHARS = "asbd/\\s@!"
	INVALID_ID     = "invalid-always"
)

// Test successful password hashing
func TestPwHash(t *testing.T) {
	// Generate a password
	t_random := time.Microsecond.Microseconds()
	se := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", t_random)))

	// Get password hash
	hash := hashPw(se)
	if string(hash) == se {
		t.Error("Hash unsuccessful!")
	}
}

// Test password hash comparison
func TestPwComp(t *testing.T) {
	// Generate a password
	t_random := time.Microsecond.Microseconds()
	se := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", t_random)))

	// Get password hash
	hash := hashPw(se)
	if !comparePw(se, string(hash)) {
		t.Error("Hash comparison false!")
	}
}

func TestPw(t *testing.T) {
	validator.SetValidationFunc("validpw", validpw)
	t.Run("Passwords valid", func(t *testing.T) {
		for i := 0; i < len(testUsers); i++ {
			assert.Nilf(t, validator.Validate(testUsers[i]), "%s Should be valid!", testUsers[i].Password)
		}
	})
	t.Run("Passwords invalid", func(t *testing.T) {
		for i := 0; i < len(wrongCredsUsers); i++ {
			assert.NotNilf(t, validator.Validate(wrongCredsUsers[i]), "%s should be illegal!", wrongCredsUsers[i].Password)
		}
	})
}

// test user registration.
func TestRegisterUser(t *testing.T) {
	testInit()
	// Test registering new users with default credentials.
	t.Run("Valid registrations", func(t *testing.T) {
		for i := range testUsers {
			trialUser := testUsers[i].getCopy()
			_, err := registerUser(trialUser)
			if err != nil {
				t.Errorf("User registration error: %v\n", err.Error())
				return
			}
		}
	})

	// Test reregistering those users
	t.Run("Repeat registrations", func(t *testing.T) {
		for i := range testUsers {
			trialUser := testUsers[i].getCopy()
			_, err := registerUser(trialUser)
			if err == nil {
				t.Error("Already registered account cannot be reregistered.")
				return
			}
		}
	})
	testEnd()
}

// Test user sign-up using test database.
func TestSignUp(t *testing.T) {
	// Set up test
	testInit()
	srv := testingServerSetup()

	// Start server.
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v\n", err)
		}
	}()

	// Test not yet registered users.
	t.Run("Valid signup requests", func(t *testing.T) {
		for i := range testUsers {
			// Create JSON body for sign up request based on test user.
			trialUser := testUsers[i].getCopy()
			resp, err := sendJsonRequest(ENDPOINT_SIGNUP, http.MethodPost, trialUser)
			if err != nil {
				t.Errorf("Error sending request: %v", err)
				return
			}
			defer resp.Body.Close()

			// Check if response OK and user registered.
			assert.Equalf(t, http.StatusOK, resp.StatusCode, "Expected %d but got %d status code!", http.StatusOK, resp.StatusCode)

			assert.NotEqualf(t, true,
				isUnique(gormDb, &User{}, "email", testUsers[i].Email), "User should be in database!")

			var exists bool
			if err := gormDb.Model(&GlobalUser{}).Select("count(*) > 0").
				Where(&GlobalUser{User: testUsers[i]}).Find(&exists).Error; err != nil {
				t.Errorf("Global ID test query error: %v", err)
			}
			assert.NotEqual(t, false, exists, "ID should be in database!")
		}
	})

	// Test bad request response for an already registered user.
	t.Run("Repeat user signups", func(t *testing.T) {
		for i := 0; i < len(testUsers); i++ {
			resp, _ := sendJsonRequest(ENDPOINT_SIGNUP, http.MethodPost, testUsers[i])
			defer resp.Body.Close()

			// Check if response is indeed unsuccessful.
			assert.Equalf(t, http.StatusBadRequest, resp.StatusCode, "Request should output %d", http.StatusBadRequest)
		}
	})

	// Test bad request response for invalid credentials.
	t.Run("Invalid signups", func(t *testing.T) {
		for i := range wrongCredsUsers {
			resp, _ := sendJsonRequest(ENDPOINT_SIGNUP, http.MethodPost, wrongCredsUsers[i])
			defer resp.Body.Close()
			// Check if response is indeed unsuccessful.
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("Status incorrect, should be %d, got %d\n", http.StatusBadRequest, resp.StatusCode)
				return
			}
		}
	})

	// Close server.
	if err := srv.Shutdown(context.Background()); err != nil {
		fmt.Printf("HTTP server shutdown: %v", err)
	}
	testEnd()
}

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

// Test user import.
func testExport(t *testing.T) {

}
