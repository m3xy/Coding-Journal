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
// 	"bytes"
// 	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
// 	"net/http"
// 	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	// "gorm.io/gorm"
)

const (
	// 	// BE VERY CAREFUL WITH THIS PATH!! IT GETS RECURSIVELY REMOVED!!
	TEST_FILES_DIR = "../filesystem/" // environment variable set to this value

// 	TEST_URL         = "http://localhost"
// 	TEST_SERVER_PORT = "3333"
)

// NOTE: ID gets set upon file insertion, so these should not be used as pointers in tests
// as to prevent adding a file with the same SubmissionID twice
var testFiles []File = []File{
	{SubmissionID: 0, SubmissionName: "testSubmission1", Path: "testFile1.txt",
		Name: "testFile1.txt", Base64Value: "hello world", MetaData: nil},
	{SubmissionID: 0, SubmissionName: "testSubmission1", Path: "testFile2.txt",
		Name: "testFile2.txt", Base64Value: "hello world", MetaData: nil},
}

var testComments []*Comment = []*Comment{
	{
		AuthorId:    "",
		Time:        fmt.Sprint(time.Now()),
		Base64Value: "Hello World",
		Replies:     []*Comment{},
	},
	{
		AuthorId:    "",
		Time:        fmt.Sprint(time.Now()),
		Base64Value: "Goodbye World",
		Replies:     []*Comment{},
	},
}
var testFileData []*FileData = []*FileData{
	{Comments: testComments},
}

// // TODO: move these to some common place as they are used here and in submissions_test
// // Initialise and clear filesystem and database.
// func initTestEnvironment() error {
// 	err := dbInit(TEST_DB)
// 	if err != nil {
// 		return err
// 	}
// 	err = dbClear()
// 	if err != nil {
// 		return err
// 	}
// 	err = setup(TEST_LOG_PATH)
// 	if err != nil {
// 		return err
// 	}
// 	if _, err = os.Stat(TEST_FILES_DIR); err == nil {
// 		os.RemoveAll(TEST_FILES_DIR)
// 	}
// 	if err := os.Mkdir(TEST_FILES_DIR, DIR_PERMISSIONS); err != nil {
// 		return err
// 	}
// 	return nil
// }

// // Clear filesystem and database before closing connections.
// func clearTestEnvironment() error {
// 	if err := os.RemoveAll(TEST_FILES_DIR); err != nil {
// 		return err
// 	}
// 	err := dbClear()
// 	if err != nil {
// 		return err
// 	}
// 	db.Close()
// 	return nil
// }

// // tests the functionality to upload files from the backend (no use of HTTP requests in this test)
// func TestAddFile(t *testing.T) {
// 	// utility function to add a single file
// 	testAddFile := func(file *File, submissionId int) {
// 		var submissionName string     // name of the submission as queried from the SQL db
// 		var fileId int                // id of the file as returned from addFileTo()
// 		var queriedFileContent string // the content of the file
// 		var queriedSubmissionId int   // the id of the submission as gotten from the files table
// 		var queriedFilePath string    // the file path as queried from the files table

// 		// adds file to the already instantiated submission
// 		fileId, err := addFileTo(file, submissionId)
// 		assert.NoErrorf(t, err, "failed to add file to the given submission: %v", err)

// 		// gets the submission name from the db
// 		querySubmissionName := fmt.Sprintf(
// 			SELECT_ROW,
// 			getDbTag(&Submission{}, "Name"),
// 			TABLE_SUBMISSIONS,
// 			getDbTag(&Submission{}, "Id"),
// 		)
// 		// executes the query
// 		row := db.QueryRow(querySubmissionName, submissionId)
// 		assert.NoError(t, row.Scan(&submissionName), "Query failure on submission name")

// 		// gets the file data from the db
// 		queryFileData := fmt.Sprintf(
// 			SELECT_ROW,
// 			fmt.Sprintf("%s, %s", getDbTag(&File{}, "SubmissionId"), getDbTag(&File{}, "Path")),
// 			TABLE_FILES,
// 			getDbTag(&File{}, "Id"),
// 		)
// 		// executes query
// 		row = db.QueryRow(queryFileData, fileId)
// 		assert.NoError(t, row.Scan(&queriedSubmissionId, &queriedFilePath), "Failed to query submission name after db")

// 		// gets the file content from the filesystem
// 		filePath := filepath.Join(TEST_FILES_DIR, fmt.Sprint(submissionId), submissionName, queriedFilePath)
// 		fileBytes, err := ioutil.ReadFile(filePath)
// 		assert.NoErrorf(t, err, "File read failure after added to filesystem: %v", err)
// 		queriedFileContent = string(fileBytes)

// 		// checks that a data file has been generated for the uploaded file
// 		fileDataPath := filepath.Join(
// 			TEST_FILES_DIR,
// 			fmt.Sprint(submissionId),
// 			DATA_DIR_NAME,
// 			submissionName,
// 			strings.TrimSuffix(queriedFilePath, filepath.Ext(queriedFilePath))+".json",
// 		)
// 		// gets data about the file, and tests it for equality against the added file
// 		_, err = os.Stat(fileDataPath)
// 		assert.NotErrorIs(t, err, os.ErrNotExist, "Data file not generated during file upload")
// 		assert.Equalf(t, submissionId, queriedSubmissionId, "Submission ID mismatch: %d vs %d", submissionId, queriedSubmissionId)
// 		assert.Equalf(t, file.Path, queriedFilePath, "File path mismatch:  %s vs %s", file.Path, queriedFilePath)
// 		assert.Equal(t, file.Base64Value, queriedFileContent, "file content not written to filesystem properly")
// 	}

// 	// tests that a single given valid file will be uploaded to the db and filesystem properly
// 	t.Run("Upload One File", func(t *testing.T) {
// 		testSubmission := testSubmissions[0]
// 		testFile := testFiles[0]

// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
// 		testSubmission.Authors = registerUsers(t, testAuthors[:1])
// 		testSubmission.Reviewers = registerUsers(t, testReviewers[:1])
// 		submissionId, err := addSubmission(testSubmission) // adds a submission for the file to be uploaded to
// 		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)
// 		testAddFile(testFile, submissionId)
// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 	})

// 	// tests that multiple files can be successfully added to one code submission
// 	t.Run("Upload Multiple Files to One Submission", func(t *testing.T) {
// 		testSubmission := testSubmissions[0]
// 		testFiles := testFiles[0:2]

// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
// 		testSubmission.Authors = registerUsers(t, testAuthors[:1])
// 		testSubmission.Reviewers = registerUsers(t, testReviewers[:1])
// 		submissionId, err := addSubmission(testSubmission) // adds a submission for the file to be uploaded to
// 		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)
// 		for _, testFile := range testFiles {
// 			testAddFile(testFile, submissionId)
// 		}
// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 	})
// }

// tests the ability of the backend to add comments to a given file.
// Test Depends On:
// 	- TestAddSubmission (in submissions_test.go)
// 	- TestAddFile
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
		authorId, err := registerUser(testAuthor)
		assert.NoErrorf(t, err, "failed to add user to the database: %v", err)
		testComment.AuthorId = authorId


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
		assert.Equalf(t, testComment.AuthorId, addedComment.AuthorId,
			"Comment author ID mismatch: %s vs %s", testComment.AuthorId, addedComment.AuthorId)
		assert.Equal(t, testComment.Base64Value, addedComment.Base64Value, "Comment content does not match")



		testEnd()
	})
}

// Tests the ability of the backend helper functions to retrieve a file's data
//
// Test Depends On:
// 	- TestAddSubmission (in submissions_test.go)
func TestGetFileData(t *testing.T) {
	// getting single valid file TODO: add comments here
	t.Run("Single Valid File", func(t *testing.T) {
		testFile := testFiles[0]             // the test file to be added to the db and filesystem (saved here so it can be easily changed)
		testSubmission := testSubmissions[0] // the test submission to be added to the db and filesystem (saved here so it can be easily changed)
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
		submissionId, err := addSubmission(&testSubmission)
		assert.NoErrorf(t, err, "Error adding submission %s: %v", testSubmission.Name, err)

		// queries the file's data from it's ID
		queriedFile, err := getFileData(testSubmission.Files[0].ID)

		// tests the returned data for equality with the sent data (files do not have comments here)
		assert.Equal(t, submissionId, queriedFile.SubmissionID, "Submission IDs do not match")
		assert.Equal(t, testFile.Name, queriedFile.Name, "File names do not match")
		assert.Equal(t, testFile.Path, queriedFile.Path, "File paths do not match")
		assert.Equal(t, testFile.Base64Value, queriedFile.Base64Value, "File Content does not match")
	})
}

// // Tests the basic ability of the files.go code to load the data from a
// // valid file path passed to it via HTTP request
// //
// // TODO : test whether having / in file path query param breaks the function
// //
// // Test Depends On:
// // 	- TestAddSubmission (in submissions_test.go)
// // 	- TestAddFile
// func TestGetFile(t *testing.T) {
// 	// tests getting a single valid file
// 	t.Run("Get One File", func(t *testing.T) {
// 		var submissionId int                 // stores submission id returned by addSubmission()
// 		testFile := testFiles[0]             // the test file to be added to the db and filesystem (saved here so it can be easily changed)
// 		testSubmission := testSubmissions[0] // the test submission to be added to the db and filesystem (saved here so it can be easily changed)

// 		// Set up server and configures filesystem/db
// 		srv := setupCORSsrv()
// 		go srv.ListenAndServe()
// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")

// 		// adds a submission to the database and filesystem
// 		testSubmission.Authors = registerUsers(t, testAuthors[:1])
// 		testSubmission.Reviewers = registerUsers(t, testReviewers[:1])
// 		submissionId, err := addSubmission(testSubmission)
// 		assert.NoErrorf(t, err, "Error adding submission %s: %v", testSubmission.Name, err)

// 		// adds a file to the database and filesystem
// 		_, err = addFileTo(testFile, submissionId)
// 		assert.NoErrorf(t, err, "Error adding file %s: %v", testFile.Name, err)

// 		// builds the request url inserting query parameters
// 		urlString := fmt.Sprintf("%s:%s%s?%s=%d&%s=%s", TEST_URL, TEST_SERVER_PORT,
// 			ENDPOINT_FILE, getJsonTag(&File{}, "SubmissionId"), testFile.SubmissionId,
// 			getJsonTag(&File{}, "Path"), testFile.Path)
// 		req, err := http.NewRequest("GET", urlString, nil)

// 		// send GET request
// 		resp, err := sendSecureRequest(req, TEAM_ID)
// 		assert.NoErrorf(t, err, "Error occurred in request: %v", err)
// 		defer resp.Body.Close()
// 		assert.Equalf(t, resp.StatusCode, http.StatusOK, "Error: %d", resp.StatusCode) // fails if status code is not 200

// 		// marshals the json response into a file struct
// 		file := &File{}
// 		assert.NoError(t, json.NewDecoder(resp.Body).Decode(&file), "Error decoding JSON in server response")

// 		// tests that the file was retrieved with the correct information
// 		assert.Equalf(t, testFile.Path, file.Path, "Incorrect file path %d != %d", file.Id, testFile.Id)
// 		assert.Equalf(t, testFile.SubmissionId, file.SubmissionId,
// 			"Incorrect file submission Id %d != %d", file.SubmissionId, testFile.SubmissionId)
// 		assert.Equal(t, testFile.Base64Value, file.Base64Value, "File Content does not match")

// 		// clears environment
// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 		assert.NoError(t, srv.Shutdown(context.Background()), "HTTP server shutdown error")
// 	})
// }

// // Tests the ability of the files.go code to upload files to the backend
// //
// // Test Depends on:
// // 	- TestCreateSubmission()
// // 	- TestAddFile()
// func TestUploadFile(t *testing.T) {
// 	// uploads a single valid file
// 	t.Run("Upload Single File", func(t *testing.T) {
// 		// the test values added to the db and filesystem (saved here so it can be easily changed)
// 		testFile := testFiles[0]
// 		testAuthor := testAuthors[0]

// 		// Set up server and configures filesystem/db
// 		srv := setupCORSsrv()
// 		go srv.ListenAndServe()
// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")

// 		// registers test author
// 		var err error
// 		testAuthor.Id, err = registerUser(testAuthor)
// 		assert.NoErrorf(t, err, "failed to add test author: %v", err)

// 		// formats the request body to send to the server to add a comment
// 		reqBody, err := json.Marshal(map[string]string{
// 			"author":                       testAuthor.Id,
// 			getJsonTag(&File{}, "Name"):    testFile.Name,
// 			getJsonTag(&File{}, "Base64Value"): testFile.Base64Value,
// 		})
// 		assert.NoErrorf(t, err, "Error formatting request body: %v", err)

// 		// formats and executes the request
// 		req, err := http.NewRequest("POST", fmt.Sprintf("%s:%s%s",
// 			TEST_URL, TEST_SERVER_PORT, ENDPOINT_NEWFILE), bytes.NewBuffer(reqBody))
// 		assert.NoErrorf(t, err, "Error creating request: %v", err)
// 		resp, err := sendSecureRequest(req, TEAM_ID)
// 		assert.NoErrorf(t, err, "Error executing request: %v", err)
// 		defer resp.Body.Close()

// 		// TODO : maybe add more checks for correctness here??
// 		// tests that the result is as desired
// 		assert.Equalf(t, resp.StatusCode, http.StatusOK, "Error: %d", resp.StatusCode)

// 		// clears environment
// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 		assert.NoError(t, srv.Shutdown(context.Background()), "HTTP server shutdown error")
// 	})
// }

// // Tests the basic ability of the CodeFiles module to add a comment to a file
// // given file path and submission id
// //
// // Test Depends On:
// // 	- TestAddComment()
// // 	- TestCreateSubmission()
// // 	- TestAddFiles()
// func TestUploadUserComment(t *testing.T) {
// 	// upload a single user comment to a valid file in a valid submission
// 	t.Run("Upload Single User Comment", func(t *testing.T) {
// 		// the test values added to the db and filesystem (saved here so it can be easily changed)
// 		testFile := testFiles[0]
// 		testSubmission := testSubmissions[0]
// 		testAuthor := testAuthors[0]
// 		testComment := testComments[0]

// 		// Set up server and configures filesystem/db
// 		srv := setupCORSsrv()
// 		go srv.ListenAndServe()
// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")

// 		// adds test values to the db and filesystem
// 		testSubmission.Authors = registerUsers(t, testAuthors[3:4])
// 		testSubmission.Reviewers = registerUsers(t, testReviewers[:1])
// 		submissionId, err := addSubmission(testSubmission)
// 		assert.NoErrorf(t, err, "error occurred while adding testSubmission: %v", err)

// 		// adds the file to the submission
// 		_, err = addFileTo(testFile, submissionId)
// 		assert.NoErrorf(t, err, "error occurred while adding testSubmission: %v", err)

// 		// registers the test author for the comment
// 		testAuthor.Id, err = registerUser(testAuthor)
// 		assert.NoErrorf(t, err, "error occurred while adding testAuthor: %v", err)
// 		testComment.AuthorId = testAuthor.Id // sets test comment author

// 		// formats the request body to send to the server to add a comment
// 		reqBody, err := json.Marshal(map[string]interface{}{
// 			getJsonTag(&File{}, "SubmissionId"):	submissionId,
// 			getJsonTag(&File{}, "Path"):			testFile.Path,
// 			getJsonTag(&Comment{}, "AuthorId"):		testAuthor.Id,
// 			getJsonTag(&Comment{}, "Base64Value"):	testComment.Base64Value,
// 		})
// 		assert.NoErrorf(t, err, "Error formatting request body: %v", err)

// 		// formats and executes the request
// 		req, err := http.NewRequest("POST", fmt.Sprintf("%s:%s%s", TEST_URL, TEST_SERVER_PORT, ENDPOINT_NEWCOMMENT), bytes.NewBuffer(reqBody))
// 		assert.NoErrorf(t, err, "Error creating request: %v", err)

// 		// sends a request to the server to post a user comment
// 		resp, err := sendSecureRequest(req, TEAM_ID)
// 		assert.NoErrorf(t, err, "Error executing request: %v", err)
// 		defer resp.Body.Close()

// 		// tests that the result is as desired
// 		assert.Equalf(t, resp.StatusCode, http.StatusOK, "HTTP request error: %d", resp.StatusCode)

// 		// tests that the comment was added properly
// 		fileDataPath := filepath.Join(
// 			TEST_FILES_DIR,
// 			fmt.Sprint(testSubmission.Id),
// 			DATA_DIR_NAME,
// 			testSubmission.Name,
// 			strings.TrimSuffix(testFile.Path, filepath.Ext(testFile.Path))+".json",
// 		)
// 		fileBytes, err := ioutil.ReadFile(fileDataPath)
// 		assert.NoErrorf(t, err, "failed to read data file: %v", err)

// 		// decodes the json from the file's metadata
// 		codeData := &FileData{}
// 		assert.NoError(t, json.Unmarshal(fileBytes, codeData), "failed to decode file metadata")

// 		// extracts the last comment (most recently added) from the comments and checks for equality with
// 		// the passed in comment
// 		addedComment := codeData.Comments[len(codeData.Comments)-1]
// 		assert.Equalf(t, testComment.AuthorId, addedComment.AuthorId,
// 			"Comment author ID mismatch: %s vs %s", testComment.AuthorId, addedComment.AuthorId)
// 		assert.Equalf(t, testComment.Base64Value, addedComment.Base64Value,
// 			"Comment content mismatch: %s vs %s", testComment.AuthorId, addedComment.AuthorId)

// 		// clears environment
// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 		assert.NoError(t, srv.Shutdown(context.Background()), "HTTP server shutdown error")
// 	})
// }
