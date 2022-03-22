package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
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
	if validate == nil {
		validate = validator.New()
		validate.RegisterValidation("ispw", ispw)
	}

	t.Run("Passwords valid", func(t *testing.T) {
		for i := 0; i < len(testUsers); i++ {
			if err := validate.Struct(testUsers[i]); err != nil {
				fmt.Printf("%v\n", err)
				assert.Nilf(t, err, "%s Should be valid!", testUsers[i].Password)
			}
		}
	})
	t.Run("Passwords invalid", func(t *testing.T) {
		for i := 0; i < len(wrongCredsUsers); i++ {
			assert.NotNilf(t, validate.Struct(wrongCredsUsers[i]), "%s should be illegal!", wrongCredsUsers[i].Password)
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
			_, err := registerUser(*trialUser, USERTYPE_NIL)
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
			_, err := registerUser(*trialUser, USERTYPE_NIL)
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
	defer testEnd()

	router := mux.NewRouter()
	router.HandleFunc(ENDPOINT_SIGNUP, signUp)

	// Test not yet registered users.
	t.Run("Valid signup requests", func(t *testing.T) {
		for i := range testUsers {
			// Create JSON body for sign up request based on test user.
			trialUser := testUsers[i].getCopy()
			reqBody, _ := json.Marshal(trialUser)
			req, w := httptest.NewRequest(http.MethodPost, ENDPOINT_SIGNUP, bytes.NewBuffer(reqBody)), httptest.NewRecorder()
			router.ServeHTTP(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			// Check if response OK and user registered.
			assert.Equalf(t, http.StatusOK, resp.StatusCode, "Expected %d but got %d status code!", http.StatusOK, resp.StatusCode)

			assert.NotEqualf(t, true,
				isUnique(gormDb, &User{}, "email", testUsers[i].Email), "User should be in database!")

			var exists bool
			if err := gormDb.Model(&GlobalUser{}).Select("count(*) > 0").
				Where(&GlobalUser{User: &testUsers[i]}).Find(&exists).Error; err != nil {
				t.Errorf("Global ID test query error: %v", err)
			}
			assert.NotEqual(t, false, exists, "ID should be in database!")
		}
	})

	// Test bad request response for an already registered user.
	t.Run("Repeat user signups", func(t *testing.T) {
		for i := 0; i < len(testUsers); i++ {
			trialUser := testUsers[i].getCopy()
			reqBody, _ := json.Marshal(trialUser)
			req, w := httptest.NewRequest(http.MethodPost, ENDPOINT_SIGNUP, bytes.NewBuffer(reqBody)), httptest.NewRecorder()
			router.ServeHTTP(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			// Check if response is indeed unsuccessful.
			assert.Equalf(t, http.StatusBadRequest, resp.StatusCode, "Request should output %d", http.StatusBadRequest)
		}
	})

	// Test bad request response for invalid credentials.
	t.Run("Invalid signups", func(t *testing.T) {
		for _, user := range wrongCredsUsers {
			trialUser := user.getCopy()
			reqBody, _ := json.Marshal(trialUser)
			req, w := httptest.NewRequest(http.MethodPost, ENDPOINT_SIGNUP, bytes.NewBuffer(reqBody)), httptest.NewRecorder()
			router.ServeHTTP(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			// Check if response is indeed unsuccessful.
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("Status incorrect, should be %d, got %d\n", http.StatusBadRequest, resp.StatusCode)
				return
			}
		}
	})
}

// Test user client login
func TestAuthLogIn(t *testing.T) {
	// Set up test and start server.
	testInit()
	defer testEnd()

	router := mux.NewRouter()
	router.HandleFunc(ENDPOINT_LOGIN, PostAuthLogIn)

	// Populate database.
	for _, u := range testUsers {
		registerUser(u, USERTYPE_NIL)
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
			body := AuthLoginPostBody{Email: user.Email, Password: user.Password, GroupNumber: TEAM_ID}
			bodyBuf, _ := json.Marshal(body)

			// Send request
			req, w := httptest.NewRequest("POST", ENDPOINT_LOGIN, bytes.NewBuffer(bodyBuf)), httptest.NewRecorder()
			router.ServeHTTP(w, req)
			resp := w.Result()
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
			var actualUser User
			if err := gormDb.Where("Email = ?", user.Email).Find(&actualUser).Error; err != nil {
				t.Errorf("SQL query error: %v\n", err)
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
				assert.Equal(t, actualUser.GlobalUserID, claims.ID, "User should be inside token.")
			}
		}
	})

	t.Run("Invalid logins", func(t *testing.T) {
		for _, user := range wrongCredsUsers {
			body := AuthLoginPostBody{Email: user.Email, Password: user.Password, GroupNumber: TEAM_ID}
			bodyBuf, _ := json.Marshal(body)

			// Send request
			req, w := httptest.NewRequest("POST", ENDPOINT_LOGIN, bytes.NewBuffer(bodyBuf)), httptest.NewRecorder()
			router.ServeHTTP(w, req)
			resp := w.Result()
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
}
