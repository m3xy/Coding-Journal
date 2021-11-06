/*
CodeFiles_test.go
author: 190010425
created: November 2, 2021

Test file for the CodeFiles module.
*/

package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/gorilla/mux"
)

const (
	// constants for filesystem
	// TEST_DB = "testdb" // TEMP: declared in authentication_test.go

	// BE VERY CAREFUL WITH THIS PATH!!
	TEST_FILES_DIR = "/home/ewp3/Documents/CS3099/project-code/testProjects/" // environment variable set to this value

	TEST_PROJECT_NAME = "testProject" // valid project name for testing queries of the filesystem
	TEST_FILE_NAME    = "testfile.txt"

	// File Mode Constants
	DIR_PERMISSIONS  = 0755 // permissions for code files
	FILE_PERMISSIONS = 0644 // permissions for project files
)

// function to test querying files
func TestGetFile(t *testing.T) {
	// config db
	testInit()

	// variables for all tests
	projectId := 1 // project id is auto-incremented, so it should be one here
	fileId := 1    // file id is auto-incremented, so it should be one, as we only have 1 id
	projectRoot := filepath.Join(TEST_FILES_DIR, fmt.Sprint(projectId))
	projectPath := filepath.Join(projectRoot, TEST_PROJECT_NAME)
	filePath := filepath.Join(projectPath, TEST_FILE_NAME)
	// for data about the stored file (i.e. comments) note that this dir structure does not contain a data file about the project
	projectDataPath := filepath.Join(projectRoot, "data", TEST_PROJECT_NAME)
	fileDataPath := filepath.Join(projectDataPath, TEST_FILE_NAME)

	// test data for the test project file and its accompanying data file
	fileContent := "Hello World"
	dataFileContent := "{}" // TEMP: not implemented yet

	// configures the file system by creating a test project
	var err error
	if err = os.MkdirAll(projectPath, 0755); err != nil {
		t.Error(err)
	}
	if err = os.MkdirAll(projectDataPath, 0755); err != nil {
		t.Error(err)
	}
	if err != nil {
		return
	}

	// populates the filesystem with a test file and data about said test file
	testFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Error(err)
		return
	}
	testDataFile, err := os.OpenFile(fileDataPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Error(err)
		return
	}

	// writes data to the file
	if _, err = testFile.Write([]byte(fileContent)); err != nil {
		t.Error(err)
	}
	if _, err = testDataFile.Write([]byte(dataFileContent)); err != nil {
		t.Error(err)
	}
	testFile.Close()
	testDataFile.Close()

	/*
		Tests the basic ability of the CodeFiles module to load the data from a
		file id passed to it
	*/
	func() {
		resp, err := http.Post("http://localhost:8080/", "application/json", bytes.NewBuffer(buffer))
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

	// Set up server to listen with the getFile() function.
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/project/file", getFile) // TEMP: this path could change
	srv := &http.Server{Addr: ":3333", Handler: muxRouter}

	// Start server.
	go srv.ListenAndServe()

	// Close server.
	if err = srv.Shutdown(context.Background()); err != nil {
		fmt.Printf("HTTP server shutdown: %v", err)
	}

	// deletes the test project
	if err = os.RemoveAll(TEST_FILES_DIR); err != nil {
		t.Error(err)
	}

	// close DB
	testEnd()
}
