package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/validator.v2"
)

const (
	TEST_LOG_PATH  = "test.log"
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

var testUsers []*Credentials = []*Credentials{
	{Email: "test.test@test.test", Pw: "123456aB$", Fname: "test",
		Lname: "test", PhoneNumber: "0574349206", Usertype: USERTYPE_USER},
	{Email: "john.doe@test.com", Pw: "dlbjDs2!", Fname: "John",
		Lname: "Doe", Organization: "TestOrg", Usertype: USERTYPE_USER},
	{Email: "jane.doe@test.net", Pw: "dlbjDs2!", Fname: "Jane",
		Lname: "Doe", Usertype: USERTYPE_PUBLISHER},
}

var wrongCredsUsers []*Credentials = []*Credentials{
	{Email: "test.nospec@test.com", Pw: "badN0Special", Fname: "test", Lname: "nospec"},
	{Email: "test.nonum@test.com", Pw: "testNoNum!", Fname: "test", Lname: "nonum"},
	{Email: "test.toosmall@test.com", Pw: "g0.Ku", Fname: "test", Lname: "toosmall"},
	{Email: "test.wrongchars@test.com", Pw: "Tho/se]chars|ille\"gal", Fname: "test", Lname: "wrongchars"},
	{Email: "test.nolowercase@test.com", Pw: "ALLCAP5!", Fname: "test", Lname: "nolower"},
	{Email: "test.nouppercase@test.com", Pw: "nocap5!!", Fname: "test", Lname: "noupper"},
}

// Initialise the database for testing.
func testInit() {
	dbInit(user, password, protocol, host, port, TEST_DB)
	setup()
	if err := dbClear(); err != nil {
		fmt.Printf("Error occurred while clearing Db: %v", err)
	}
}

// Close database at the end of test.
func testEnd() {
	if err := dbClear(); err != nil {
		fmt.Printf("Error occurred while clearing Db: %v", err)
	}
	db.Close()
}

func sendJsonRequest(endpoint string, method string, data interface{}) (*http.Response, error) {
	var req *http.Request
	if data != nil {
		jsonDat, _ := json.Marshal(data)
		req, _ = http.NewRequest(method, BACKEND_ADDRESS+endpoint, bytes.NewBuffer(jsonDat))
	} else {
		req, _ = http.NewRequest(method, BACKEND_ADDRESS+endpoint, nil)
	}
	return sendSecureRequest(req, TEAM_ID)
}

func testAuth(t *testing.T) {
	// Set up database.
	dbInit(user, password, protocol, host, port, TEST_DB)
	setup()
	if err := dbClear(); err != nil {
		fmt.Printf("Error occured while clearing Db: %v", err)
	}

	// Set up logging to a local testing file.
	file, err := os.OpenFile(TEST_LOG_PATH, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
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
			assert.Nilf(t, validator.Validate(*testUsers[i]), "%s Should be valid!", testUsers[i].Pw)
		}
	})
	t.Run("Passwords invalid", func(t *testing.T) {
		for i := 0; i < len(wrongCredsUsers); i++ {
			assert.NotNilf(t, validator.Validate(*wrongCredsUsers[i]), "%s should be illegal!", wrongCredsUsers[i].Pw)
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
func TestCheckUnique(t *testing.T) {
	testInit()
	// Test uniqueness in empty table
	t.Run("Unique elements", func(t *testing.T) {
		for i := 0; i < len(testUsers); i++ {
			assert.Truef(t, checkUnique(TABLE_USERS, getDbTag(&Credentials{}, "Email"),
				testUsers[i].Email), "Email %s Shouldn't exist in database!", testUsers[i].Email)
		}
	})

	// Add an element to table
	stmt := fmt.Sprintf(INSERT_CRED,
		TABLE_USERS,
		getDbTag(&Credentials{}, "Pw"),
		getDbTag(&Credentials{}, "Fname"),
		getDbTag(&Credentials{}, "Lname"),
		getDbTag(&Credentials{}, "Email"))

	// Test uniquenes if element already exists in table.
	t.Run("Non-unique elements", func(t *testing.T) {
		for i := 0; i < len(testUsers); i++ {
			_, err := db.Query(stmt, testUsers[i].Pw, testUsers[i].Fname, testUsers[i].Lname, testUsers[i].Email)
			assert.Nilf(t, err, "Database query should work, but error gotten: %v", err)
			assert.Falsef(t, checkUnique(TABLE_USERS, getDbTag(&Credentials{}, "Email"), testUsers[0].Email),
				"Email %s should already be in database!", testUsers[i].Email)
		}
	})
	testEnd()
}

// Test user sign-up using test database.
func TestSignUp(t *testing.T) {
	// Set up test
	testInit()
	srv := setupCORSsrv()

	// Start server.
	go srv.ListenAndServe()

	// Test not yet registered users.
	t.Run("Valid signup requests", func(t *testing.T) {
		UserStmt := fmt.Sprintf(SELECT_ROW, getDbTag(&Credentials{}, "Id"), TABLE_USERS, getDbTag(&Credentials{}, "Email"))
		IdsStmt := fmt.Sprintf(SELECT_ROW, getDbTag(&IdMappings{}, "GlobalId"), TABLE_IDMAPPINGS, getDbTag(&IdMappings{}, "Id"))
		for i := range testUsers {
			// Create JSON body for sign up request based on test user.
			resp, _ := sendJsonRequest(ENDPOINT_SIGNUP, http.MethodPost, testUsers[i])
			defer resp.Body.Close()

			// Check if response OK and user registered.
			assert.Equalf(t, http.StatusOK, resp.StatusCode, "Expected %d but got %d status code!", http.StatusOK, resp.StatusCode)

			res := db.QueryRow(UserStmt, testUsers[i].Email)
			storedCreds := &Credentials{}
			err := res.Scan(&storedCreds.Id)
			assert.NotEqualf(t, sql.ErrNoRows, err, "User %s %s should be in database!", testUsers[i].Fname, testUsers[i].Lname)

			// Check if global ID exists for user.
			res = db.QueryRow(IdsStmt, storedCreds.Id)
			storedMapping := &IdMappings{Id: storedCreds.Id}
			err = res.Scan(storedMapping.GlobalId)
			assert.NotEqualf(t, sql.ErrNoRows, err, "ID %s should be in database!", storedMapping.GlobalId)
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
	srv := setupCORSsrv()

	// Start server.
	go srv.ListenAndServe()

	// Populate database with valid users.
	for i := range testUsers {
		id, err := registerUser(testUsers[i])
		if err != nil {
			t.Errorf("User registration error: %v\n", err)
			return
		} else {
			// Set user ID for ID checking.
			testUsers[i].Id = id
		}
	}

	// Test valid logins
	t.Run("Valid logins", func(t *testing.T) {
		for i := range testUsers {
			// Create a request for user login.
			loginMap := make(map[string]string)
			loginMap[getJsonTag(&Credentials{}, "Email")] = testUsers[i].Email
			loginMap[JSON_TAG_PW] = testUsers[i].Pw
			resp, err := sendJsonRequest(ENDPOINT_LOGIN, http.MethodPost, loginMap)
			assert.Nil(t, err, "Request should not error.")
			assert.Equalf(t, http.StatusOK, resp.StatusCode, "Response status should be %d", http.StatusOK)

			// Get ID from user response.
			respMap := make(map[string]string)
			err = json.NewDecoder(resp.Body).Decode(&respMap)
			assert.Nil(t, err, "Body unparsing should succeed")
			storedId, exists := respMap[getJsonTag(&Credentials{}, "Id")]
			assert.True(t, exists, "ID should exist in response.")

			// Check if gotten
			assert.Equal(t, testUsers[i].Id, storedId, "ID must equal registration's ID.")
		}
	})

	// Test invalid password login.
	t.Run("Invalid password logins", func(t *testing.T) {
		for i := 0; i < len(testUsers); i++ {
			loginMap := make(map[string]string)
			loginMap[getJsonTag(&Credentials{}, "Email")] = testUsers[i].Email
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
			loginMap[getJsonTag(&Credentials{}, "Email")] = testUsers[0].Email
			loginMap[JSON_TAG_PW] = testUsers[i].Pw

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
	srv := setupCORSsrv()

	// Start server.
	go srv.ListenAndServe()

	// Populate database for testing and test valid user.
	for i := range testUsers {
		testUsers[i].Id, _ = registerUser(testUsers[i])
	}

	t.Run("Valid user profiles", func(t *testing.T) {
		for i := range testUsers {
			res, err := sendJsonRequest(ENDPOINT_USERINFO+"/"+testUsers[i].Id, http.MethodGet, nil)
			assert.Nil(t, err, "Request should not error.")
			assert.Equal(t, http.StatusOK, res.StatusCode, "Status should be OK.")

			resCreds := Credentials{}
			err = json.NewDecoder(res.Body).Decode(&resCreds)
			assert.Nil(t, err, "JSON decoding must not error.")

			// Check equality for all user info.
			assert.Equal(t, testUsers[i].Email, resCreds.Email, "Email should be equal.")
			assert.Equal(t, testUsers[i].Fname, resCreds.Fname, "First name should be equal.")
			assert.Equal(t, testUsers[i].Lname, resCreds.Lname, "Last name should be equal.")
			assert.Equal(t, testUsers[i].Usertype, resCreds.Usertype, "Usertype should be equal.")
			assert.Equal(t, testUsers[i].PhoneNumber, resCreds.PhoneNumber, "Phone number should be equal.")
			assert.Equal(t, testUsers[i].Organization, resCreds.Organization, "Organization should be equal.")
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
func TestExport(t *testing.T) {

}
