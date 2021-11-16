package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gopkg.in/validator.v2"
	"net/http"
	"testing"
	"time"
)

const (
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

// Test password hashing
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

// Test if valid password tests password well.
func TestValidPw(t *testing.T) {
	// Initialise test credentials.
	testCreds0 := &Credentials{}
	testCreds0.Fname = "test"
	testCreds0.Lname = "test"
	testCreds0.Email = "test.test@test.test"
	validator.SetValidationFunc("validpw", validpw)

	// Get valid password and test validation.
	testCreds0.Pw = VALID_PW
	if err := validator.Validate(*testCreds0); err != nil {
		t.Error("Valid password error!")
	}

	// Password without lowercase invalid.
	testCreds0.Pw = PW_NO_LC
	if validator.Validate(*testCreds0) == nil {
		t.Error("No lowercase error!")
	}

	// Password without uppercase invalid.
	testCreds0.Pw = PW_NO_UC
	if validator.Validate(*testCreds0) == nil {
		t.Error("No uppercase error!")
	}

	// Password without number invalid.
	testCreds0.Pw = PW_NO_NUM
	if validator.Validate(*testCreds0) == nil {
		t.Error("No number error!")
	}

	// Password without special characters invalid.
	testCreds0.Pw = PW_NO_SC
	if validator.Validate(*testCreds0) == nil {
		t.Error("No special charactacter error!")
	}

	// Password with wrong characters invalid.
	testCreds0.Pw = PW_WRONG_CHARS
	if validator.Validate(*testCreds0) == nil {
		t.Error("Wrong special charactacters error!")
	}
}

// test user registration.
func TestRegisterUser(t *testing.T) {
	testInit()
	// Test registering new users with default credentials.
	for i := range testUsers {
		_, err := registerUser(testUsers[i])
		if err != nil {
			t.Errorf("User registration error: %v\n", err.Error())
			return
		}
	}

	// Test reregistering those users
	for i := range testUsers {
		_, err := registerUser(testUsers[i])
		if err == nil {
			t.Error("Already registered account cannot be reregistered.")
			return
		}
	}
	testEnd()
}

// Test credential uniqueness with test database.
func TestCheckUnique(t *testing.T) {
	testInit()
	// Test uniqueness in empty table
	unique := checkUnique(TABLE_USERS, getDbTag(testUsers[0], "Email"), testUsers[0].Email)
	if !unique {
		t.Error(getDbTag(&Credentials{}, "Email") + "is not unique!")
	}

	// Add an element to table
	stmt := fmt.Sprintf(INSERT_CRED,
		TABLE_USERS,
		getDbTag(testUsers[0], "Pw"),
		getDbTag(testUsers[0], "Fname"),
		getDbTag(testUsers[0], "Lname"),
		getDbTag(testUsers[0], "Email"))
	_, err := db.Query(stmt, testUsers[0].Pw, testUsers[0].Fname, testUsers[0].Lname, testUsers[0].Email)
	if err != nil {
		t.Errorf("Testing function error: %v\n", err.Error())
		return
	}

	// Test uniquenes if element already exists in table.
	unique = checkUnique(TABLE_USERS, getDbTag(&Credentials{}, "Email"), testUsers[0].Email)
	if unique {
		t.Error("User should not be unique here!")
	}
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
	for i := range testUsers {
		// Create JSON body for sign up request based on test user.
		buffer, err := json.Marshal(testUsers[i])
		if err != nil {
			t.Errorf("Error marshalling user: %v/n", err)
			return
		}
		req, err := http.NewRequest("POST", "http://localhost:3333/register", bytes.NewBuffer(buffer))
		if err != nil {
			t.Errorf("Error in request: %v/n", err)
			return
		}
		resp, err := sendSecureRequest(req, TEAM_ID)
		if err != nil {
			t.Errorf("Error in response: %v/n", err)
			return
		}
		defer resp.Body.Close()

		// Check if response OK and user registered.
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Error occured: %d\n", resp.StatusCode)
			return
		}

		stmt := fmt.Sprintf(SELECT_ROW, getDbTag(&Credentials{}, "Id"), TABLE_USERS, getDbTag(&Credentials{}, "Email"))
		res := db.QueryRow(stmt, testUsers[i].Email)

		storedCreds := &Credentials{}
		err = res.Scan(&storedCreds.Id)
		if err == sql.ErrNoRows {
			t.Errorf("No rows despire register: %v\n", err)
			return
		}

		// Check if global ID exists for user.
		stmt = fmt.Sprintf(SELECT_ROW, getDbTag(&IdMappings{}, "GlobalId"), TABLE_IDMAPPINGS, getDbTag(&IdMappings{}, "Id"))
		res = db.QueryRow(stmt, storedCreds.Id)

		storedMapping := &IdMappings{Id: storedCreds.Id}
		err = res.Scan(storedMapping.GlobalId)
		if err == sql.ErrNoRows {
			t.Errorf("No rows despite register: %v\n", err)
		}
	}

	// Test bad request response for an already registered user.
	func() {
		buffer, err := json.Marshal(testUsers[0])
		if err != nil {
			t.Errorf("Error marshalling user: %v/n", err)
			return
		}
		req, err := http.NewRequest("POST", "http://localhost:3333/register", bytes.NewBuffer(buffer))
		if err != nil {
			t.Errorf("Request error in already registered user: %v\n", err)
			return
		}
		resp, err := sendSecureRequest(req, TEAM_ID)
		if err != nil {
			t.Errorf("Request error in already registered user: %v\n", err)
			return
		}
		defer resp.Body.Close()

		// Check if response is indeed unsuccessful.
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Status should be %d, got %d\n", http.StatusBadRequest, resp.StatusCode)
			return
		}
	}()

	// Test bad request response for invalid password.
	for i := range wrongCredsUsers {
		buffer, err := json.Marshal(wrongCredsUsers[i])
		if err != nil {
			t.Errorf("Error marshalling user: %v/n", err)
			return
		}
		req, err := http.NewRequest("POST", "http://localhost:3333/register", bytes.NewBuffer(buffer))
		if err != nil {
			t.Errorf("Request error: %v\n", err.Error())
			return
		}
		resp, err := sendSecureRequest(req, TEAM_ID)
		if err != nil {
			t.Errorf("Response error: %v\n", err.Error())
			return
		}
		defer resp.Body.Close()
		// Check if response is indeed unsuccessful.
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Status incorrect, should be %d, got %d\n", http.StatusBadRequest, resp.StatusCode)
			return
		}
	}

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
	for i := range testUsers {
		// Create a request for user login.
		loginMap := make(map[string]string)
		loginMap[getJsonTag(&Credentials{}, "Email")] = testUsers[i].Email
		loginMap[JSON_TAG_PW] = testUsers[i].Pw
		buffer, err := json.Marshal(loginMap)
		if err != nil {
			t.Errorf("JSON Marshal Error: %v\n", err)
			return
		}
		req, err := http.NewRequest("POST", "http://localhost:3333/login", bytes.NewBuffer(buffer))
		if err != nil {
			t.Errorf("Request error on correct login: %v\n", err)
			return
		}
		resp, err := sendSecureRequest(req, TEAM_ID)
		if err != nil {
			t.Errorf("Response error on correct login: %v\n", err)
			return
		} else if resp.StatusCode != http.StatusOK {
			t.Errorf("Response status should be %d, got %d\n", http.StatusOK, resp.StatusCode)
			return
		}
		// Get ID from user response.
		respMap := make(map[string]string)
		err = json.NewDecoder(resp.Body).Decode(&respMap)
		if err != nil {
			t.Errorf("JSON Decode error: %v\n", err)
			return
		} else if _, exists := respMap[getJsonTag(&Credentials{}, "Id")]; !exists {
			t.Error("ID not in http response!")
			return
		}

		// Check if gotten
		storedId := respMap[getJsonTag(&Credentials{}, "Id")]
		if storedId != testUsers[i].Id {
			t.Errorf("IDs don't correspond! %s vs %s", storedId, testUsers[i].Id)
			return
		}
	}

	// Test invalid password login.
	func() {
		loginMap := make(map[string]string)
		loginMap[getJsonTag(&Credentials{}, "Email")] = testUsers[0].Email
		loginMap[JSON_TAG_PW] = testUsers[1].Pw
		buffer, err := json.Marshal(loginMap)
		if err != nil {
			t.Errorf("JSON Marshal Error: %v\n", err)
			return
		}
		req, err := http.NewRequest("POST", "http://localhost:3333/login", bytes.NewBuffer(buffer))
		if err != nil {
			t.Errorf("Request error on correct login: %s\n", err)
			return
		}
		resp, err := sendSecureRequest(req, TEAM_ID)
		if err != nil {
			t.Errorf("Request error on correct login: %v\n", err)
			return
		} else if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Response status should be %d, got %d\n", http.StatusUnauthorized, resp.StatusCode)
			return
		}

	}()

	// Test invalid email login.
	func() {
		loginMap := make(map[string]string)
		loginMap[getJsonTag(&Credentials{}, "Email")] = wrongCredsUsers[0].Email
		loginMap[JSON_TAG_PW] = testUsers[0].Pw
		buffer, err := json.Marshal(loginMap)
		if err != nil {
			t.Errorf("JSON Marshal Error: %v\n", err)
			return
		}
		req, err := http.NewRequest("POST", "http://localhost:3333/login", bytes.NewBuffer(buffer))
		if err != nil {
			t.Errorf("Request error on correct login: %v\n", err)
		}
		resp, err := sendSecureRequest(req, TEAM_ID)
		if err != nil {
			t.Errorf("Request error on correct login: %v\n", err)
			return
		} else if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Response status should be %d, got %d\n", http.StatusUnauthorized, resp.StatusCode)
			return
		}

	}()

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
		validReq, err := http.NewRequest("GET", "http://localhost:3333/users/"+
			testUsers[i].Id, nil)
		if err != nil {
			t.Errorf("Request making error: %v\n", err)
			return
		}
		res, err := sendSecureRequest(validReq, TEAM_ID)
		if err != nil {
			t.Errorf("Error in request sending: %v\n", err)
			return
		}
		assert.Equal(t, http.StatusOK, res.StatusCode, "Status should be OK.")
		resCreds := Credentials{}
		err = json.NewDecoder(res.Body).Decode(&resCreds)
		if err != nil {
			t.Errorf("JSON decoding error: %v\n", err)
			return
		}
		// Check equality for all user info.
		assert.Equal(t, testUsers[i].Email, resCreds.Email,
			"Email should be equal.")
		assert.Equal(t, testUsers[i].Fname, resCreds.Fname,
			"First name should be equal.")
		assert.Equal(t, testUsers[i].Lname, resCreds.Lname,
			"Last name should be equal.")
		assert.Equal(t, testUsers[i].Usertype, resCreds.Usertype,
			"Usertype should be equal.")
		assert.Equal(t, testUsers[i].PhoneNumber, resCreds.PhoneNumber,
			"Phone number should be equal.")
		assert.Equal(t, testUsers[i].Organization, resCreds.Organization,
			"Organization should be equal.")
	}

	// Test invalid users.
	invalidReq, err := http.NewRequest("GET", "http://localhost:3333/users/"+
		INVALID_ID, nil)
	if err != nil {
		t.Errorf("Request making error: %v\n", err)
		return
	}
	res, err := sendSecureRequest(invalidReq, TEAM_ID)
	if err != nil {
		t.Errorf("Response error: %v\n", err)
		return
	}
	assert.Equal(t, res.StatusCode, http.StatusNotFound, "Status should be 404.")

	// Close server.
	if err := srv.Shutdown(context.Background()); err != nil {
		fmt.Printf("HTTP server shutdown: %v", err)
	}
	testEnd()
}

// Test user import.
func TestExport(t *testing.T) {

}
