package main

import (
	"encoding/base64"
	"testing"
	"time"
	"fmt"
)

// Test password hashing
func testPwHash(t *testing.T) {
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
func testPwComp(t *testing.T) {

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
