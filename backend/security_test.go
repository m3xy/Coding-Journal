package main

import (
	"testing"
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
