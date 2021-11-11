package main

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

const (
	WRONG_SECURITY_TOKEN = "testwrongToken"
)

func TestRandString (t *testing.T) {
	// Test if given new rand string is not null.
	if randStringBase64(1, 128) == "" {
		t.Error("Random string empty!")
		return
	}
	if randStringBase64(2, 8) == "" {
		t.Error("Random string empty!")
		return
	}
	if randStringBase64(2, 0) != "" {
		t.Error("Null size random string not empty!")
		return
	}
}

func TestSecurityCheck (t *testing.T) {
	testInit()
	// Test security check
	err := securityCheck()
	if err != nil {
		t.Errorf("Security check failure: %v\n", err)
	}

	// Test security check if non-empty token server
	err = securityCheck()
	if err != nil {
		t.Errorf("Security check failure: %v\n", err)
	}
	testEnd()
}

func TestValidateToken (t *testing.T) {
	testInit()
	securityCheck() // Generate the security token.

	// Test valid security token.
	storedToken := ""
	err := db.QueryRow(fmt.Sprintf(
		SELECT_ROW, getDbTag(&Servers{}, "Token"), TABLE_SERVERS,
		getDbTag(&Servers{}, "GroupNb")), TEAM_ID).
		Scan(&storedToken)
	if err != nil {
		t.Errorf("Token query error: %v\n", err)
	} else if !validateToken(storedToken) {
		t.Errorf("Token validation error: token should be valid.\n")
	}

	// Test invalid security token.
	storedToken = WRONG_SECURITY_TOKEN
	if validateToken(storedToken) {
		t.Errorf("Token validation error: test token should be invalid.\n")
	}

	testEnd()
}

// Send a given request using needed authentication.
func sendSecureRequest(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, errors.New("Request nil!")
	}
	// Fetch valid security token from database.
	storedToken := ""
	err := db.QueryRow(fmt.Sprintf(
		SELECT_ROW, getDbTag(&Servers{}, "Token"), TABLE_SERVERS,
		getDbTag(&Servers{}, "GroupNb")), TEAM_ID).
		Scan(&storedToken)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	req.Header.Set(SECURITY_TOKEN_KEY, storedToken)
	return client.Do(req)
}

func TestTokenValidation(t *testing.T) {
	testInit()
	securityCheck()
	srv := setupCORSsrv()

	// Start server
	go srv.ListenAndServe()

	// Write valid security token response
	validReq, err := http.NewRequest("GET", "http://localhost:3333/validate", nil)
	res, err := sendSecureRequest(validReq)
	if err != nil {
		t.Errorf("HTTP request error: %v\n", err)
	} else if res.StatusCode != http.StatusOK  {
		t.Errorf("Response Status code should be OK, but is %d", res.StatusCode)
	}

	// Write invalid security token response
	client := http.Client{}
	validReq, err = http.NewRequest("GET", "http://localhost:3333/validate", nil)
	if err != nil {
		t.Errorf("Request creation error: %v\n", err)
	}
	validReq.Header.Set(SECURITY_TOKEN_KEY, WRONG_SECURITY_TOKEN)
	res, err = client.Do(validReq)
	if err != nil {
		t.Errorf("HTTP request error: %v\n", err)
	} else if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("Response Status code should be 401, but is %d", res.StatusCode)
	}
}
