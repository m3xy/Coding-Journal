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

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
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
	TEST_PORT_AUTH = ":59215"
)

func authServerSetup() *http.Server {
	router := mux.NewRouter()
	router.Use(testingMiddleware)

	auth := router.PathPrefix(SUBROUTE_AUTH).Subrouter()
	auth.HandleFunc(ENDPOINT_SIGNUP, signUp).Methods(http.MethodPost, http.MethodOptions)
	auth.HandleFunc(ENDPOINT_LOGIN, authLogIn).Methods(http.MethodPost, http.MethodOptions)

	return &http.Server{
		Addr:    TEST_PORT_AUTH,
		Handler: router,
	}
}

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
	validator.SetValidationFunc("ispw", ispw)
	validator.SetValidationFunc("isemail", isemail)
	t.Run("Passwords valid", func(t *testing.T) {
		for i := 0; i < len(testUsers); i++ {
			if err := validator.Validate(testUsers[i]); err != nil {
				fmt.Printf("%v\n", err)
				assert.Nilf(t, err, "%s Should be valid!", testUsers[i].Password)
			}
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
	srv := authServerSetup()

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
			resp, err := sendJsonRequest(SUBROUTE_AUTH+ENDPOINT_SIGNUP, http.MethodPost, trialUser, TEST_PORT_AUTH)
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
			resp, _ := sendJsonRequest(SUBROUTE_AUTH+ENDPOINT_SIGNUP, http.MethodPost, testUsers[i], TEST_PORT_AUTH)
			defer resp.Body.Close()

			// Check if response is indeed unsuccessful.
			assert.Equalf(t, http.StatusBadRequest, resp.StatusCode, "Request should output %d", http.StatusBadRequest)
		}
	})

	// Test bad request response for invalid credentials.
	t.Run("Invalid signups", func(t *testing.T) {
		for i := range wrongCredsUsers {
			resp, _ := sendJsonRequest(SUBROUTE_AUTH+ENDPOINT_SIGNUP, http.MethodPost, wrongCredsUsers[i], TEST_PORT_AUTH)
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

// Test user client login
func TestAuthLogIn(t *testing.T) {
	// Set up test and start server.
	testInit()
	srv := authServerSetup()
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v\n", err)
		}
	}()

	// Populate database.
	for _, u := range testUsers {
		registerUser(u)
	}

	// Set JWT Secret.
	if myEnv, err := godotenv.Read("secrets.env"); err != nil {
		t.Errorf("Dotenv file reading error: %v", err)
		return
	} else {
		JwtSecret = myEnv["JWT_SECRET"]
	}

	t.Run("Valid logins", func(t *testing.T) {
		for _, user := range testUsers {
			// Send request
			resp, err := sendJsonRequest(SUBROUTE_AUTH+ENDPOINT_LOGIN, http.MethodPost, user, TEST_PORT_AUTH)
			if err != nil {
				t.Errorf("Request error: %v\n", err)
				return
			}
			defer resp.Body.Close()
			if !assert.Equalf(t, 200, resp.StatusCode, "Request should be successful!") {
				return
			}

			// Validate body content
			var validBody AuthLogInResponse
			if err := json.NewDecoder(resp.Body).Decode(&validBody); err != nil {
				t.Errorf("Login response decoding error: %v\n", err)
				return
			}

			if token, err := jwt.ParseWithClaims(validBody.AccessToken, &JwtClaims{}, func(t *jwt.Token) (interface{}, error) {
				return []byte(JwtSecret), nil
			}); err != nil {
				t.Errorf("Error parsing JWT token: %v\n", err)
				return
			} else if claims, ok := token.Claims.(*JwtClaims); !ok {
				t.Errorf("Error getting JWT token claims: %v\n", err)
				return
			} else {
				assert.Equal(t, user.Email, claims.Email, "User should be inside token.")
			}
		}
	})

	t.Run("Invalid logins", func(t *testing.T) {
		for _, user := range wrongCredsUsers {
			// Send request
			resp, err := sendJsonRequest(SUBROUTE_AUTH+ENDPOINT_LOGIN, http.MethodPost, user, TEST_PORT_AUTH)
			if err != nil {
				t.Errorf("Request error: %v\n", err)
				return
			}
			defer resp.Body.Close()
			if !assert.Equal(t, 401, resp.StatusCode, "Status code should be 401 - unauthorized") {
				return
			}

			// Verify body format
			var validBody StandardResponse
			if err := json.NewDecoder(resp.Body).Decode(&validBody); err != nil {
				t.Errorf("Response failed to decode to correct format: %v\n", err)
			}
		}
	})

	testEnd()
}
