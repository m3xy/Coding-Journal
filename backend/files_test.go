// ====================================================
// files_test.go
// Authors: 190010425
// Created: November 23, 2021
//
// This file depends heavily on submissions_test.go
// ====================================================

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const (
	// BE VERY CAREFUL WITH THIS PATH!! IT GETS RECURSIVELY REMOVED!!
	TEST_FILES_DIR = "../filesystem_test/" // environment variable set to this value
)

var testFiles []File = []File{
	{SubmissionID: 0, Path: "testFile1.txt", Base64Value: "hello world"},
	{SubmissionID: 0, Path: "testFile2.txt", Base64Value: "hello world"},
}

// -----------
// Router Function Tests
// -----------

// Tests the basic ability of the files.go code to load the data from a
// valid file path passed to it via HTTP request
func TestGetFile(t *testing.T) {
	testInit()
	defer testEnd()
	testFile := testFiles[0]             // the test file to be added to the db and filesystem (saved here so it can be easily changed)
	testSubmission := testSubmissions[0] // the test submission to be added to the db and filesystem (saved here so it can be easily changed)

	router := mux.NewRouter()
	router.HandleFunc(SUBROUTE_FILE+"/{id}", GetFile)

	// adds a submission to the database and filesystem
	authorID, err := registerUser(testAuthors[0], USERTYPE_PUBLISHER)
	if !assert.NoErrorf(t, err, "Error occurred while registering user: %v", err) {
		return
	}
	testSubmission.Authors = []GlobalUser{{ID: authorID}}

	reviewerID, err := registerUser(testReviewers[0], USERTYPE_REVIEWER)
	if !assert.NoErrorf(t, err, "Error occurred while registering user: %v", err) {
		return
	}
	testSubmission.Reviewers = []GlobalUser{{ID: reviewerID}}


	testSubmission.Files = []File{testFile}
	submissionID, err := addSubmission(&testSubmission)
	if !assert.NoErrorf(t, err, "Error adding submission %s: %v", testSubmission.Name, err) {
		return
	}
	if !assert.NoError(t, gormDb.Model(&File{}).Find(&testFile).Error, "error occurred while getting file ID") { return }
	fileID := testFile.ID

	// tests getting a single valid file without comments
	t.Run("Get One File no comments", func(t *testing.T) {
		// builds the request url inserting query parameters
		urlString := fmt.Sprintf("%s/%d", SUBROUTE_FILE, fileID)
		req, w := httptest.NewRequest("GET", urlString, nil), httptest.NewRecorder()
		router.ServeHTTP(w, req)
		resp := w.Result()
		defer resp.Body.Close()
		if !assert.Equalf(t, resp.StatusCode, http.StatusOK, "Error: %d", resp.StatusCode) {
			return
		}
		// marshals the json response into a file struct
		respData := &GetFileResponse{}
		if !assert.NoError(t, json.NewDecoder(resp.Body).Decode(&respData), "Error decoding JSON in server response") {
			return
		}
		// tests that the file was retrieved with the correct information
		switch {
		case !assert.Equal(t, testFile.Path, respData.File.Path, "file paths do not match"),
			!assert.Equal(t, submissionID, respData.File.SubmissionID, "Submission IDs do not match"),
			!assert.Equal(t, testFile.Base64Value, respData.File.Base64Value, "File Content does not match"):
			return
		}
	})

	t.Run("non-existant file ID", func(t *testing.T) {
		// builds the request url inserting query parameters
		urlString := fmt.Sprintf("%s/%d", SUBROUTE_FILE, fileID+1)
		req, w := httptest.NewRequest("GET", urlString, nil), httptest.NewRecorder()
		router.ServeHTTP(w, req)
		resp := w.Result()
		defer resp.Body.Close()
		if !assert.Equalf(t, http.StatusBadRequest, resp.StatusCode, "returned incorrect status code") {
			return
		}
	})
}

// -------------
// Helper Function Tests
// -------------

// tests the functionality to upload files from the backend (no use of HTTP requests in this test)
func TestAddFile(t *testing.T) {
	testInit()
	defer testEnd()

	testSubmission := testSubmissions[0]
	testFiles := testFiles[0:2]

	authorID, err := registerUser(testAuthors[0], USERTYPE_PUBLISHER)
	if !assert.NoErrorf(t, err, "Error occurred while registering user: %v", err) {
		return
	}
	testSubmission.Authors = []GlobalUser{{ID: authorID}}

	reviewerID, err := registerUser(testReviewers[0], USERTYPE_REVIEWER)
	if !assert.NoErrorf(t, err, "Error occurred while registering user: %v", err) {
		return
	}
	testSubmission.Reviewers = []GlobalUser{{ID: reviewerID}}

	submissionID, err := addSubmission(&testSubmission) // adds a submission for the file to be uploaded to
	if !assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err) {
		return
	}

	// tests that multiple files can be successfully added to one code submission
	t.Run("Upload Valid Files to One Submission", func(t *testing.T) {
		for _, testFile := range testFiles {
			// adds file to the already instantiated submission
			fileID, err := addFileTo(&testFile, submissionID)
			if !assert.NoErrorf(t, err, "failed to add file to the given submission: %v", err) {
				return
			}

			// gets the submission name from the db
			submission := &Submission{}
			if !assert.NoError(t, gormDb.Select("Name, created_at, ID").First(submission, submissionID).Error,
				"Error retrieving submission name") {
				return
			}

			// gets the file data from the db
			queriedFile := &File{}
			queriedFile.ID = fileID
			if !assert.NoError(t, gormDb.Model(queriedFile).Select("files.submission_id", "files.path").
				First(queriedFile).Error, "Error retrieving added file") {
				return
			}

			// gets the file content from the filesystem
			filePath := filepath.Join(getSubmissionDirectoryPath(*submission), fmt.Sprint(queriedFile.ID))
			fileBytes, err := ioutil.ReadFile(filePath)
			if !assert.NoErrorf(t, err, "File read failure after added to filesystem: %v", err) {
				return
			}

			queriedFileContent := string(fileBytes)

			// asserts the file was added properly
			switch {
			case !assert.Equalf(t, submissionID, testFile.SubmissionID,
				"Submission ID mismatch: %d vs %d", submissionID, testFile.SubmissionID):
			case !assert.Equalf(t, testFile.Path, queriedFile.Path,
				"File path mismatch:  %s vs %s", testFile.Path, queriedFile.Path):
			case !assert.Equal(t, testFile.Base64Value, queriedFileContent,
				"file content not written to filesystem properly"):
				return

			}
		}
	})

	// tests that a file cannot be added to a non-existant submission
	t.Run("Non-existant submission", func(t *testing.T) {
		_, err := addFileTo(&testFiles[0], submissionID+1) // invalid submission ID
		assert.Error(t, err, "No error occurred when attempting to add a file to an invalid submission")
	})
}

// Tests the ability of the backend helper functions to retrieve a file's data
// Test Depends On:
// 	- TestAddSubmission (in submissions_test.go)
func TestGetFileData(t *testing.T) {
	testInit()
	defer testEnd()

	testFile := testFiles[0]             // the test file to be added to the db and filesystem.
	testSubmission := testSubmissions[0] // test submission to be added to db and filesystem.
	testAuthor := testAuthors[0]
	testReviewer := testReviewers[0]

	// configures the test submission fields
	testSubmission.Files = []File{testFile}

	authorID, err := registerUser(testAuthor, USERTYPE_PUBLISHER)
	if !assert.NoErrorf(t, err, "Error occurred while registering user: %v", err) {
		return
	}
	testSubmission.Authors = []GlobalUser{{ID: authorID}}

	reviewerID, err := registerUser(testReviewer, USERTYPE_REVIEWER)
	if !assert.NoErrorf(t, err, "Error occurred while registering user: %v", err) {
		return
	}
	testSubmission.Reviewers = []GlobalUser{{ID: reviewerID}}

	// adds a submission to the database and filesystem
	submissionID, err := addSubmission(&testSubmission)
	if !assert.NoErrorf(t, err, "Error adding submission %s: %v", testSubmission.Name, err) {
		return
	}

	// getting single valid file TODO: add comments here
	t.Run("Single Valid File", func(t *testing.T) {

		// queries the file's data from it's ID
		queriedFile, err := getFileData(testSubmission.Files[0].ID)
		if !assert.NoErrorf(t, err, "Should not error but got: %v", err) {
			return
		}

		// tests the returned data for equality with the sent data (files do not have comments here)
		switch {
		case !assert.Equal(t, submissionID, queriedFile.SubmissionID, "Submission IDs do not match"):
		case !assert.Equal(t, testFile.Path, queriedFile.Path, "File paths do not match"):
		case !assert.Equal(t, testFile.Base64Value, queriedFile.Base64Value, "File Content does not match"):
			return
		}
	})
}
