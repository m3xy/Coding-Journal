package main

import (
	"testing"
)

func TestDbInit(t *testing.T) {
	err := dbInit(dbname)
	if err != nil {
		t.Error(err.Error())
	}
}
