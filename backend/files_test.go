// ===================================================================
// files_test.go
// Authors: 190010425
// Created: November 23, 2021
//
// test file for files.go
// Note that the tests are written dependency wise from top to bottom.
// Hence if a test breaks, fix the top one first and then re-run.
//
// This file depends heavily on submissions_test.go
// ===================================================================

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	// "gorm.io/gorm"
)

const (
	// BE VERY CAREFUL WITH THIS PATH!! IT GETS RECURSIVELY REMOVED!!
	TEST_FILES_DIR     = "../filesystem/" // environment variable set to this value
	TEST_PORT_FILES    = ":59216"
	TEST_FILES_ADDRESS = "http://localhost:59216"
)

// Set up server used for files testing.
func resourceServerSetup() *http.Server {
	router := mux.NewRouter()
	router.HandleFunc(ENDPOINT_FILE, getFile).Methods(http.MethodGet)
	router.HandleFunc(ENDPOINT_NEWCOMMENT, uploadUserComment).Methods(http.MethodPost, http.MethodOptions)

	return &http.Server{
		Addr:    TEST_PORT_FILES,
		Handler: router,
	}
}

// NOTE: ID gets set upon file insertion, so these should not be used as pointers in tests
// as to prevent adding a file with the same SubmissionID twice
var testFiles []File = []File{
	{SubmissionID: 0, Path: "testFile1.txt",
		Name: "testFile1.txt", Base64Value: "hello world", MetaData: nil},
	{SubmissionID: 0, Path: "testFile2.txt",
		Name: "testFile2.txt", Base64Value: "hello world", MetaData: nil},
}

var testComments []*Comment = []*Comment{
	{
		AuthorID:    "",
		CreatedAt:        fmt.Sprint(time.Now()),
		Base64Value: "Hello World",
		Replies:     []*Comment{},
	},
	{
		AuthorID:    "",
		CreatedAt:        fmt.Sprint(time.Now()),
		Base64Value: "Goodbye World",
		Replies:     []*Comment{},
	},
}
var testFileData []*FileData = []*FileData{
	{Comments: testComments},
}

// -----------
// Router Function Tests
// -----------

// Tests the basic ability of the files.go code to load the data from a
// valid file path passed to it via HTTP request
//
// TODO : test whether having / in file path query param breaks the function
//
// Test Depends On:
// 	- TestAddSubmission (in submissions_test.go)
// 	- TestAddFile
func TestGetFile(t *testing.T) {
	// tests getting a single valid file without comments
	t.Run("Get One File no comments", func(t *testing.T) {
		testFile := testFiles[0]             // the test file to be added to the db and filesystem (saved here so it can be easily changed)
		testSubmission := testSubmissions[0] // the test submission to be added to the db and filesystem (saved here so it can be easily changed)

		// Set up server and configures filesystem/db
		testInit()
		srv := resourceServerSetup()
		go srv.ListenAndServe()

		// adds a submission to the database and filesystem
		authorID, err := registerUser(testAuthors[0])
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission.Authors = []GlobalUser{{ID: authorID}}

		reviewerID, err := registerUser(testReviewers[0])
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission.Reviewers = []GlobalUser{{ID: reviewerID}}

		submissionID, err := addSubmission(&testSubmission)
		assert.NoErrorf(t, err, "Error adding submission %s: %v", testSubmission.Name, err)

		// adds a file to the database and filesystem
		fileID, err := addFileTo(&testFile, submissionID)
		assert.NoErrorf(t, err, "Error adding file %s: %v", testFile.Name, err)

		// builds the request url inserting query parameters
		urlString := fmt.Sprintf("%s%s?%s=%d", TEST_FILES_ADDRESS, ENDPOINT_FILE, "id", fileID)
		req, err := http.NewRequest("GET", urlString, nil)

		// send GET request
		resp, err := sendSecureRequest(gormDb, req, TEAM_ID)
		assert.NoErrorf(t, err, "Error occurred in request: %v", err)
		defer resp.Body.Close()
		assert.Equalf(t, resp.StatusCode, http.StatusOK, "Error: %d", resp.StatusCode) // fails if status code is not 200

		// marshals the json response into a file struct
		file := &File{}
		assert.NoError(t, json.NewDecoder(resp.Body).Decode(&file), "Error decoding JSON in server response")

		// tests that the file was retrieved with the correct information (no comments here so metadata is not checked)
		assert.Equal(t, testFile.Path, file.Path, "file paths do not match")
		assert.Equal(t, testFile.Name, file.Name, "file names do not match")
		assert.Equal(t, submissionID, file.SubmissionID, "Submission IDs do not match")
		assert.Equal(t, testFile.Base64Value, file.Base64Value, "File Content does not match")

		// clears environment
		assert.NoError(t, srv.Shutdown(context.Background()), "HTTP server shutdown error")
		testEnd()
	})
}

// Tests the basic ability of the CodeFiles module to add a comment to a file
// given file path and submission id
//
// Test Depends On:
// 	- TestAddComment()
// 	- TestCreateSubmission()
// 	- TestAddFiles()
func TestUploadUserComment(t *testing.T) {
	// upload a single user comment to a valid file in a valid submission
	t.Run("Upload Single User Comment", func(t *testing.T) {
		// the test values added to the db and filesystem (saved here so it can be easily changed)
		testFile := testFiles[0]
		testSubmission := testSubmissions[0]
		testAuthor := testAuthors[1] // author of the comment
		testComment := testComments[0]

		// Set up server and configures filesystem/db
		testInit()
		srv := resourceServerSetup()
		go srv.ListenAndServe()

		// adds test values to the db and filesystem
		subAuthorID, err := registerUser(testAuthors[0])
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission.Authors = []GlobalUser{{ID: subAuthorID}}

		submissionID, err := addSubmission(&testSubmission)
		assert.NoErrorf(t, err, "error occurred while adding testSubmission: %v", err)

		// adds the file to the submission
		fileID, err := addFileTo(&testFile, submissionID)
		assert.NoErrorf(t, err, "error occurred while adding testSubmission: %v", err)

		// registers the test author for the comment
		authorID, err := registerUser(testAuthor)
		assert.NoErrorf(t, err, "error occurred while adding testAuthor: %v", err)
		testComment.AuthorID = authorID // sets test comment author

		// formats the request body to send to the server to add a comment
		reqBody, err := json.Marshal(map[string]interface{}{
			// getJsonTag(&File{}, "SubmissionID"):	submissionID,
			// getJsonTag(&File{}, "Path"):			testFile.Path,
			// getJsonTag(&Comment{}, "AuthorID"):		testAuthor.ID,
			getJsonTag(&Comment{}, "Base64Value"): testComment.Base64Value,
		})
		assert.NoErrorf(t, err, "Error formatting request body: %v", err)

		// formats and executes the request
		req, err := http.NewRequest("POST", fmt.Sprintf("%s%s?%s=%s&%s=%d", TEST_FILES_ADDRESS, ENDPOINT_NEWCOMMENT,
			"authorID", authorID, "fileID", fileID), bytes.NewBuffer(reqBody))
		assert.NoErrorf(t, err, "Error creating request: %v", err)

		// sends a request to the server to post a user comment
		resp, err := sendSecureRequest(gormDb, req, TEAM_ID)
		assert.NoErrorf(t, err, "Error executing request: %v", err)
		defer resp.Body.Close()
		assert.Equalf(t, resp.StatusCode, http.StatusOK, "HTTP request error: %d", resp.StatusCode)

		// tests that the comment was added properly
		fileDataPath := filepath.Join(
			TEST_FILES_DIR,
			fmt.Sprint(testSubmission.ID),
			DATA_DIR_NAME,
			testSubmission.Name,
			strings.TrimSuffix(testFile.Path, filepath.Ext(testFile.Path))+".json",
		)
		codeData, err := getFileMetaData(fileDataPath)
		assert.NoError(t, err, "Error occurred while retrieving file metadata")

		// extracts the last comment (most recently added) from the comments and checks for equality with
		// the passed in comment
		addedComment := codeData.Comments[len(codeData.Comments)-1]
		assert.Equal(t, testComment.AuthorID, addedComment.AuthorID, "Comment author ID is incorrect")
		assert.Equal(t, testComment.Base64Value, addedComment.Base64Value, "Comment content does not match")

		// clears environment
		assert.NoError(t, srv.Shutdown(context.Background()), "HTTP server shutdown error")
		testEnd()
	})
}

// -------------
// Helper Function Tests
// -------------

// tests the functionality to upload files from the backend (no use of HTTP requests in this test)
func TestAddFile(t *testing.T) {
	// utility function to add a single file
	testAddFile := func(file *File, submissionID uint) {
		// adds file to the already instantiated submission
		fileID, err := addFileTo(file, submissionID)
		assert.NoErrorf(t, err, "failed to add file to the given submission: %v", err)

		// gets the submission name from the db
		submission := &Submission{}
		submission.ID = submissionID
		assert.NoError(t, gormDb.Model(submission).Select("submissions.name").First(submission).Error, "Error retrieving submission name")

		// gets the file data from the db
		queriedFile := &File{}
		queriedFile.ID = fileID
		assert.NoError(t, gormDb.Model(queriedFile).Select("files.submission_id", "files.path").First(queriedFile).Error, "Error retrieving added file")

		// gets the file content from the filesystem
		filePath := filepath.Join(TEST_FILES_DIR, fmt.Sprint(submissionID), submission.Name, queriedFile.Path)
		fileBytes, err := ioutil.ReadFile(filePath)
		assert.NoErrorf(t, err, "File read failure after added to filesystem: %v", err)
		queriedFileContent := string(fileBytes)

		// checks that a data file has been generated for the uploaded file
		fileDataPath := filepath.Join(
			TEST_FILES_DIR,
			fmt.Sprint(submissionID),
			DATA_DIR_NAME,
			submission.Name,
			strings.TrimSuffix(queriedFile.Path, filepath.Ext(queriedFile.Path))+".json",
		)
		// gets data about the file, and tests it for equality against the added file
		_, err = os.Stat(fileDataPath)
		assert.NotErrorIs(t, err, os.ErrNotExist, "Data file not generated during file upload")
		assert.Equalf(t, submissionID, file.SubmissionID, "Submission ID mismatch: %d vs %d", submissionID, file.SubmissionID)
		assert.Equalf(t, file.Path, queriedFile.Path, "File path mismatch:  %s vs %s", file.Path, queriedFile.Path)
		assert.Equal(t, file.Base64Value, queriedFileContent, "file content not written to filesystem properly")
	}

	// tests that a single given valid file will be uploaded to the db and filesystem properly
	t.Run("Upload One File", func(t *testing.T) {
		testSubmission := testSubmissions[0]
		testFile := testFiles[0]

		testInit()
		authorID, err := registerUser(testAuthors[0])
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission.Authors = []GlobalUser{{ID: authorID}}

		reviewerID, err := registerUser(testReviewers[0])
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission.Reviewers = []GlobalUser{{ID: reviewerID}}

		submissionID, err := addSubmission(&testSubmission) // adds a submission for the file to be uploaded to
		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)
		testAddFile(&testFile, submissionID)
		testEnd()
	})

	// tests that multiple files can be successfully added to one code submission
	t.Run("Upload Multiple Files to One Submission", func(t *testing.T) {
		testSubmission := testSubmissions[0]
		testFiles := testFiles[0:2]

		testInit()
		authorID, err := registerUser(testAuthors[0])
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission.Authors = []GlobalUser{{ID: authorID}}

		reviewerID, err := registerUser(testReviewers[0])
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission.Reviewers = []GlobalUser{{ID: reviewerID}}

		submissionID, err := addSubmission(&testSubmission) // adds a submission for the file to be uploaded to
		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)
		for _, testFile := range testFiles {
			testAddFile(&testFile, submissionID)
		}
		testEnd()
	})

	// tests that a file cannot be added to a non-existant submission
	t.Run("Non-existant submission", func(t *testing.T) {
		testInit()
		_, err := addFileTo(&testFiles[0], 1000) // invalid submission ID
		assert.Error(t, err, "No error occurred when attempting to add a file to an invalid submission")
		testEnd()
	})
}

// tests the ability of the backend to add comments to a given file.
// Test Depends On:
// 	- TestAddSubmission (in submissions_test.go)
// 	- TestAddFile
/*
func TestAddComment(t *testing.T) {
	// tests adding one valid comment. Uses the testAddComment() utility method
	t.Run("Add One Comment", func(t *testing.T) {
		testSubmission := testSubmissions[0] // test submission to add testFile to
		testFile := testFiles[0]             // test file to add comments to
		testAuthor := testAuthors[1]         // test author of comment
		testComment := testComments[0]

		testInit()

		// adds a submission for the test file to be added to
		authorID, err := registerUser(testAuthors[0])
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission.Authors = []GlobalUser{{ID: authorID}}

		reviewerID, err := registerUser(testReviewers[0])
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission.Reviewers = []GlobalUser{{ID: reviewerID}}

		testSubmission.Files = []File{testFile}
		testSubmissionID, err := addSubmission(&testSubmission)
		assert.NoErrorf(t, err, "failed to add submission: %v", err)

		// adds a test user to author a comment
		commentAuthorID, err := registerUser(testAuthor)
		assert.NoErrorf(t, err, "failed to add user to the database: %v", err)
		testComment.AuthorID = commentAuthorID

		// adds a comment to the file
		assert.NoError(t, addComment(testComment, testFile.ID), "failed to add comment to the submission")

		// reads the data file into a CodeDataFile struct
		fileDataPath := filepath.Join(
			TEST_FILES_DIR,
			fmt.Sprint(testSubmissionID),
			DATA_DIR_NAME,
			testSubmission.Name,
			strings.TrimSuffix(testFile.Path, filepath.Ext(testFile.Path))+".json",
		)
		fileBytes, err := ioutil.ReadFile(fileDataPath)
		assert.NoErrorf(t, err, "failed to read data file: %v", err)

		// unmarshalls the file's meta-data to extract the added comment
		codeData := &FileData{}
		assert.NoError(t, json.Unmarshal(fileBytes, codeData), "failed to unmarshal code file data")

		// extracts the last comment (most recently added) from the comments and checks for equality with
		// the passed in comment
		addedComment := codeData.Comments[len(codeData.Comments)-1]
		assert.Equalf(t, testComment.AuthorID, addedComment.AuthorID,
			"Comment author ID mismatch: %s vs %s", testComment.AuthorID, addedComment.AuthorID)
		assert.Equal(t, testComment.Base64Value, addedComment.Base64Value, "Comment content does not match")

		testEnd()
	})
}
*/

// Tests the ability of the backend helper functions to retrieve a file's data
// Test Depends On:
// 	- TestAddSubmission (in submissions_test.go)
func TestGetFileData(t *testing.T) {
	// getting single valid file TODO: add comments here
	t.Run("Single Valid File", func(t *testing.T) {
		testFile := testFiles[0]             // the test file to be added to the db and filesystem.
		testSubmission := testSubmissions[0] // test submission to be added to db and filesystem.
		testAuthor := testAuthors[0]
		testReviewer := testReviewers[0]

		testInit()

		// configures the test submission fields
		testSubmission.Files = []File{testFile}

		authorID, err := registerUser(testAuthor)
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission.Authors = []GlobalUser{{ID: authorID}}

		reviewerID, err := registerUser(testReviewer)
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission.Reviewers = []GlobalUser{{ID: reviewerID}}

		// adds a submission to the database and filesystem
		submissionID, err := addSubmission(&testSubmission)
		assert.NoErrorf(t, err, "Error adding submission %s: %v", testSubmission.Name, err)

		// queries the file's data from it's ID
		queriedFile, err := getFileData(testSubmission.Files[0].ID)

		// tests the returned data for equality with the sent data (files do not have comments here)
		assert.Equal(t, submissionID, queriedFile.SubmissionID, "Submission IDs do not match")
		assert.Equal(t, testFile.Name, queriedFile.Name, "File names do not match")
		assert.Equal(t, testFile.Path, queriedFile.Path, "File paths do not match")
		assert.Equal(t, testFile.Base64Value, queriedFile.Base64Value, "File Content does not match")
	})
}
