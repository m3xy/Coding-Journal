package main

import (
	"testing"
)

func TestDbInit (t *testing.T) {
	err := dbInit(user, password, protocol, host, port, dbname)
	if err != nil {
		t.Error(err.Error())
	}
}
