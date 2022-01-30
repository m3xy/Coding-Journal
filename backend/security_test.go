package main

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	WRONG_SECURITY_TOKEN = "testwrongToken"
)

func TestRandString(t *testing.T) {
	// Test if given new rand string is not null.
	t.Run("Valid size=128", func(t *testing.T) {
		assert.NotEqual(t, "", randStringBase64(1, 128), "Random string must have been created!")
	})
	t.Run("Valid size=8", func(t *testing.T) {
		assert.NotEqual(t, "", randStringBase64(2, 8), "Random string must have been created!")
	})
	t.Run("Invalid size=0", func(t *testing.T) {
		assert.Equal(t, "", randStringBase64(3, 0), "Random string must be empty!")
	})
}

func TestSecurityCheck(t *testing.T) {
	testInit()
	// Test security check
	t.Run("First run", func(t *testing.T) {
		err := securityCheck(gormDb)
		if err != nil {
			t.Errorf("Security check failure: %v\n", err)
		}
	})

	// Test security check if non-empty token server
	t.Run("Subsequent run", func(t *testing.T) {
		err := securityCheck(gormDb)
		if err != nil {
			t.Errorf("Security check failure: %v\n", err)
		}
	})
	testEnd()
}

func TestValidateToken(t *testing.T) {
	testInit()

	// Test valid security token.
	t.Run("Valid security token", func(t *testing.T) {
		var storedToken string
		err := gormDb.Model(&Server{}).Select("token").Find(&storedToken, TEAM_ID)
		assert.Nil(t, err, "Database query should not error!")
		assert.True(t, validateToken(gormDb, storedToken), "Token should be valid!")
	})

	// Test invalid security token.
	t.Run("Invalid security token", func(t *testing.T) {
		storedToken := WRONG_SECURITY_TOKEN
		assert.False(t, validateToken(gormDb, storedToken), "Token should NOT be valid!")
	})
	testEnd()
}

func TestTokenValidation(t *testing.T) {
	testInit()
	srv := setupCORSsrv()

	// Start server
	go srv.ListenAndServe()

	// Write valid security token response
	t.Run("Valid token validation", func(t *testing.T) {
		validReq, _ := http.NewRequest("GET", BACKEND_ADDRESS+ENDPOINT_VALIDATE, nil)
		res, err := sendSecureRequest(gormDb, validReq, TEAM_ID)
		if err != nil {
			t.Errorf("HTTP request error: %v\n", err)
		} else if res.StatusCode != http.StatusOK {
			t.Errorf("Response Status code should be OK, but is %d", res.StatusCode)
		}
	})

	// Write invalid security token response
	t.Run("Invalid token validation", func(t *testing.T) {
		client := http.Client{}
		invalidReq, _ := http.NewRequest("GET", BACKEND_ADDRESS+ENDPOINT_VALIDATE, nil)
		invalidReq.Header.Set(SECURITY_TOKEN_KEY, WRONG_SECURITY_TOKEN)
		res, err := client.Do(invalidReq)
		if err != nil {
			t.Errorf("HTTP request error: %v\n", err)
		} else if res.StatusCode != http.StatusUnauthorized {
			t.Errorf("Response Status code should be 401, but is %d", res.StatusCode)
		}
	})
}
