package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"gopkg.in/validator.v2"
	"gorm.io/gorm/logger"
)

const (
	TEST_LOG_PATH  = "./test.log"
	TEST_PORT      = ":59213"
	LOCALHOST      = "http://localhost"
	VALID_PW       = "aB12345$"
	PW_NO_UC       = "a123456$"
	PW_NO_LC       = "B123456$"
	PW_NO_NUM      = "aBcdefg$"
	PW_NO_SC       = "aB123456"
	PW_WRONG_CHARS = "asbd/\\s@!"
	TEST_DB        = "testdb"
	JSON_TAG_PW    = "password"
	INVALID_ID     = "invalid-always"
)

var testUsers []User = []User{
	{Email: "test.test@test.test", Password: "123456aB$", FirstName: "test",
		LastName: "test", PhoneNumber: "0574349206", UserType: USERTYPE_USER},
	{Email: "john.doe@test.com", Password: "dlbjDs2!", FirstName: "John",
		LastName: "Doe", Organization: "TestOrg", UserType: USERTYPE_USER},
	{Email: "jane.doe@test.net", Password: "dlbjDs2!", FirstName: "Jane",
		LastName: "Doe", UserType: USERTYPE_PUBLISHER},
}

var wrongCredsUsers []User = []User{
	{Email: "test.nospec@test.com", Password: "badN0Special", FirstName: "test", LastName: "nospec"},
	{Email: "test.nonum@test.com", Password: "testNoNum!", FirstName: "test", LastName: "nonum"},
	{Email: "test.toosmall@test.com", Password: "g0.Ku", FirstName: "test", LastName: "toosmall"},
	{Email: "test.wrongchars@test.com", Password: "Tho/se]chars|ille\"gal", FirstName: "test", LastName: "wrongchars"},
	{Email: "test.nolowercase@test.com", Password: "ALLCAP5!", FirstName: "test", LastName: "nolower"},
	{Email: "test.nouppercase@test.com", Password: "nocap5!!", FirstName: "test", LastName: "noupper"},
}

var testLogger logger.Interface = logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
	SlowThreshold:             time.Second,
	LogLevel:                  logger.Silent,
	IgnoreRecordNotFoundError: true,
	Colorful:                  true,
})

// Initialise the database for testing.
func testInit() {
	gormDb, _ = gormInit(TEST_DB, testLogger)
	setup(gormDb, TEST_LOG_PATH)
	if err := gormDb.Transaction(gormClear); err != nil {
		fmt.Printf("Error occurred while clearing Db: %v", err)
	}
}

// Middleware for the test authentication server.
func testingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !validateToken(gormDb, r.Header.Get(SECURITY_TOKEN_KEY)) {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			if r.Header.Get("user") != "" {
				ctx := context.WithValue(r.Context(), "user", r.Header.Get("user"))
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				next.ServeHTTP(w, r)
			}
		}
	})
}

func testingServerSetup() *http.Server {
	router := mux.NewRouter()
	router.Use(testingMiddleware)

	// Call authentication endpoints.
	router.HandleFunc(ENDPOINT_LOGIN, logIn).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc(ENDPOINT_LOGIN_GLOBAL, logInGlobal).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc(ENDPOINT_SIGNUP, signUp).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc(ENDPOINT_VALIDATE, tokenValidation).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/users/{"+getJsonTag(&User{}, "ID")+"}", getUserProfile).Methods(http.MethodGet, http.MethodOptions)

	// Setup testing HTTP server
	return &http.Server{
		Addr:    TEST_PORT,
		Handler: router,
	}
}

// Close database at the end of test.
func testEnd() {
	if err := gormDb.Transaction(gormClear); err != nil {
		fmt.Printf("Error occurred while clearing Db: %v", err)
	}
	getDB, _ := gormDb.DB()
	getDB.Close()
}

func sendJsonRequest(endpoint string, method string, data interface{}) (*http.Response, error) {
	var req *http.Request
	if data != nil {
		jsonDat, _ := json.Marshal(data)
		req, _ = http.NewRequest(method, LOCALHOST+TEST_PORT+endpoint, bytes.NewBuffer(jsonDat))
	} else {
		req, _ = http.NewRequest(method, LOCALHOST+TEST_PORT+endpoint, nil)
	}
	return sendSecureRequest(gormDb, req, TEAM_ID)
}

func testAuth(t *testing.T) {
	// Set up database.
	var err error
	gormDb, err = gormInit(TEST_DB, testLogger)
	if err != nil {
		fmt.Printf("Error opening database :%v", err)
	}
	if err := gormDb.Transaction(gormClear); err != nil {
		fmt.Printf("Error occured while clearing Db: %v", err)
	}

	// Set up logging to a local testing file.
	file, err := os.OpenFile(TEST_LOG_PATH, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("Log file creation failure: %v! Exiting...", err)
	}
	log.SetOutput(file)
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
			_, err := registerUser(testUsers[i])
			if err != nil {
				t.Errorf("User registration error: %v\n", err.Error())
				return
			}
		}
	})

	// Test reregistering those users
	t.Run("Repeat registrations", func(t *testing.T) {
		for i := range testUsers {
			_, err := registerUser(testUsers[i])
			if err == nil {
				t.Error("Already registered account cannot be reregistered.")
				return
			}
		}
	})
	testEnd()
}

// Test credential uniqueness with test database.
func TestIsUnique(t *testing.T) {
	testInit()
	// Test uniqueness in empty table
	t.Run("Unique elements", func(t *testing.T) {
		for i := 0; i < len(testUsers); i++ {
			assert.Truef(t, isUnique(gormDb, &User{}, "email", testUsers[i].Email),
				"Email %s Shouldn't exist in database!", testUsers[i].Email)
		}
	})

	// Add an element to table
	// Add test users to database
	if err := gormDb.Create(&testUsers).Error; err != nil {
		t.Errorf("Batch user creation error: %v", err)
		return
	}

	// Test uniquenes if element already exists in table.
	t.Run("Non-unique elements", func(t *testing.T) {
		for i := 0; i < len(testUsers); i++ {
			assert.Falsef(t, isUnique(gormDb, &User{}, "email", testUsers[i].Email),
				"Email %s should already be in database!", testUsers[i].Email)
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
			resp, err := sendJsonRequest(ENDPOINT_SIGNUP, http.MethodPost, testUsers[i])
			if err != nil {
				t.Errorf("Error sending request: %v", err)
				return
			}
			defer resp.Body.Close()

			// Check if response OK and user registered.
			assert.Equalf(t, http.StatusOK, resp.StatusCode, "Expected %d but got %d status code!", http.StatusOK, resp.StatusCode)

			assert.NotEqualf(t, false,
				isUnique(gormDb, &User{}, "email", testUsers[i].Email), "User should be in database!")

			var exists bool
			if err := gormDb.Model(&GlobalID{}).Select("count(*) > 0").
				Where(&GlobalID{User: testUsers[i]}).Find(&exists).Error; err != nil {
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

// Test user log in.
func TestLogIn(t *testing.T) {
	// Set up test
	testInit()
	srv := testingServerSetup()

	// Start server.
	go srv.ListenAndServe()

	// Populate database with valid users.
	gormDb.Create(&testUsers)

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
			assert.Equal(t, testUsers[i].ID, storedId, "ID must equal registration's ID.")
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

// Test user info getter.
func TestGetUserProfile(t *testing.T) {
	testInit()
	srv := testingServerSetup()

	// Start server.
	go srv.ListenAndServe()

	// Populate database for testing and test valid user.
	globalUsers := make([]GlobalID, len(testUsers))
	for i := range testUsers {
		globalUsers[i].ID, _ = registerUser(testUsers[i])
	}

	t.Run("Valid user profiles", func(t *testing.T) {
		for i := range testUsers {
			res, err := sendJsonRequest(ENDPOINT_USERINFO+"/"+globalUsers[i].ID, http.MethodGet, nil)
			assert.Nil(t, err, "Request should not error.")
			assert.Equal(t, http.StatusOK, res.StatusCode, "Status should be OK.")

			resCreds := User{}
			err = json.NewDecoder(res.Body).Decode(&resCreds)
			assert.Nil(t, err, "JSON decoding must not error.")

			// Check equality for all user info.
			equal := reflect.DeepEqual(testUsers[i], resCreds)
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
