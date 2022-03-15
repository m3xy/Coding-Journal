package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	WRONG_SECURITY_TOKEN = "testwrongToken"
)

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
