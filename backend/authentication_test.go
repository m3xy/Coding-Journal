package main

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"gopkg.in/validator.v2"
)

const (
	VALID_PW  = "aB12345$"
	PW_NO_UC  = "a123456$"
	PW_NO_LC  = "B123456$"
	PW_NO_NUM = "aBcdefg$"
	PW_NO_SC  = "aB123456"
)

// Test password hashing
func TestPwHash(t *testing.T) {
	// Generate a password
	t_random := time.Microsecond.Microseconds()
	se := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", t_random)))

	// Get password hash
	hash := hashPw(se)
	if string(hash) == se  {
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
	if !comparePw(se, string(hash))  {
		t.Error("Hash comparison false!")
	}
}

// Test if valid password tests password well.
func TestValidPw(t *testing.T) {
	// Initialise test credentials.
	testCreds0 := &Credentials{}
	testCreds0.Username = "test"
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
}

// test user registration.
func testRegisterUser(t *testing.T) {

}

// Test credential uniqueness with test database.
func testCheckUnique(t *testing.T) {

}

// Test user sign-up using test database.
func testSignUp(t *testing.T) {

}

// Test user log in.
func testLogIn(t *testing.T) {

}
