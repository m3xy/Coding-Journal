package main

import (
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
		err := gormDb.Model(&Server{}).Select("token").Find(&storedToken, TEAM_ID).Error
		assert.Nil(t, err, "Database query should not error!")
		assert.True(t, validateSecurityKey(gormDb, storedToken), "Token should be valid!")
	})

	// Test invalid security token.
	t.Run("Invalid security token", func(t *testing.T) {
		assert.False(t, validateSecurityKey(gormDb, WRONG_SECURITY_TOKEN), "Token should NOT be valid!")
	})
	testEnd()
}
