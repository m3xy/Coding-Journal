// ===================================================================
// files_test.go
// Authors: 190010425
// Created: November 23, 2021
//
// test file for files.go
// Note that the tests are written dependency wise from top to bottom.
// Hence if a test breaks, fix the top one first and then re-run.
// ===================================================================

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	// constants for filesystem
	// TEST_DB = "testdb" // TEMP: declared in authentication_test.go

	// BE VERY CAREFUL WITH THIS PATH!! IT GETS RECURSIVELY REMOVED!!
	TEST_FILES_DIR = "../filesystem/" // environment variable set to this value

	TEST_URL         = "http://localhost"
	TEST_SERVER_PORT = "3333"
)

var testFiles []*File = []*File{
	{Id: -1, SubmissionId: -1, SubmissionName: "testSubmission1", Path: "testFile1.txt",
		Name: "testFile1.txt", Content: "hello world", Comments: nil},
	{Id: -1, SubmissionId: -1, SubmissionName: "testSubmission1", Path: "testFile2.txt",
		Name: "testFile2.txt", Content: "hello world", Comments: nil},
}

var testComments []*Comment = []*Comment{
	{
		AuthorId: "",
		Time:     fmt.Sprint(time.Now()),
		Content:  "Hello World",
		Replies:  []*Comment{},
	},
	{
		AuthorId: "",
		Time:     fmt.Sprint(time.Now()),
		Content:  "Goodbye World",
		Replies:  []*Comment{},
	},
}
var testFileData []*CodeFileData = []*CodeFileData{
	{Comments: testComments},
}

// TODO: move these to some common place as they are used here and in submissions_test
// Initialise and clear filesystem and database.
func initTestEnvironment() error {
	dbInit(TEST_DB)
	dbClear()
	err := setup()
	if err != nil {
		return err
	}
	if _, err = os.Stat(TEST_FILES_DIR); err == nil {
		os.RemoveAll(TEST_FILES_DIR)
	}
	if err := os.Mkdir(TEST_FILES_DIR, DIR_PERMISSIONS); err != nil {
		return err
	}
	return nil
}

// Clear filesystem and database before closing connections.
func clearTestEnvironment() error {
	if err := os.RemoveAll(TEST_FILES_DIR); err != nil {
		return err
	}
	dbClear()
	db.Close()
	return nil
}

// test function to add a single file. This function is not called directly as a test, but is a utility method for other tests
func testAddFile(file *File, submissionId int) error {
	var submissionName string     // name of the submission as queried from the SQL db
	var fileId int                // id of the file as returned from addFileTo()
	var queriedFileContent string // the content of the file
	var queriedSubmissionId int   // the id of the submission as gotten from the files table
	var queriedFilePath string    // the file path as queried from the files table

	// adds file to the already instantiated submission
	fileId, err := addFileTo(file, submissionId)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to add file to the given submission"))
	}

	// gets the submission name from the db
	querySubmissionName := fmt.Sprintf(
		SELECT_ROW,
		getDbTag(&Submission{}, "Name"),
		TABLE_SUBMISSIONS,
		getDbTag(&Submission{}, "Id"),
	)
	// executes the query
	row := db.QueryRow(querySubmissionName, submissionId)
	if err = row.Scan(&submissionName); err != nil {
		return errors.New(fmt.Sprintf("Query failure on submission name: %v", err))
	}

	// gets the file data from the db
	queryFileData := fmt.Sprintf(
		SELECT_ROW,
		fmt.Sprintf("%s, %s", getDbTag(&File{}, "SubmissionId"), getDbTag(&File{}, "Path")),
		TABLE_FILES,
		getDbTag(&File{}, "Id"),
	)
	// executes query
	row = db.QueryRow(queryFileData, fileId)
	if err = row.Scan(&queriedSubmissionId, &queriedFilePath); err != nil {
		return errors.New(
			fmt.Sprintf("Failed to query submission name after db: %v", err))
	}

	// gets the file content from the filesystem
	filePath := filepath.Join(TEST_FILES_DIR, fmt.Sprint(submissionId), submissionName, queriedFilePath)
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return errors.New(
			fmt.Sprintf("File read failure after added to filesystem: %v", err))
	}
	queriedFileContent = string(fileBytes)

	// checks that a data file has been generated for the uploaded file
	fileDataPath := filepath.Join(
		TEST_FILES_DIR,
		fmt.Sprint(submissionId),
		DATA_DIR_NAME,
		submissionName,
		strings.TrimSuffix(queriedFilePath, filepath.Ext(queriedFilePath))+".json",
	)
	// gets data about the file, and tests it for equality against the added file
	_, err = os.Stat(fileDataPath)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return errors.New("Data file not generated during file upload")
	} else if submissionId != queriedSubmissionId { // Compare  test values.
		return errors.New(fmt.Sprintf("Submission ID mismatch: %d vs %d",
			submissionId, queriedSubmissionId))
	} else if file.Path != queriedFilePath {
		return errors.New(fmt.Sprintf("File path mismatch:  %s vs %s",
			file.Path, queriedFilePath))
	} else if file.Content != queriedFileContent {
		return errors.New(
			fmt.Sprintf("file content not written to filesystem properly"))
	}
	return nil
}

// tests that a single given valid file will be uploaded to the db and filesystem properly
func TestAddOneFile(t *testing.T) {
	testSubmission := testSubmissions[0]
	testFile := testFiles[0]
	// sets up the test environment
	if err := initTestEnvironment(); err != nil {
		t.Errorf("error while initializing the test environment db: %v", err)
	}

	// adds the test submission and file to the db and filesystem
	submissionId, err := addSubmission(testSubmission)
	if err != nil {
		t.Errorf("Error occurred while adding test submission: %v", err)
	} else if err = testAddFile(testFile, submissionId); err != nil {
		t.Errorf("Error occurred while adding file: %v", err)
	}

	// tears down the test environment
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("error while tearing down test environment: %v", err)
	}
}

// tests that multiple files can be successfully added to one code submission
func TestAddMultipleFiles(t *testing.T) {
	testSubmission := testSubmissions[0]
	testFiles := testFiles[0:2]

	// sets up the test environmetn
	if err := initTestEnvironment(); err != nil {
		t.Errorf("error while initializing the test environment db: %v", err)
	}

	// adds a test submission to the db
	submissionId, err := addSubmission(testSubmission)
	if err != nil {
		t.Errorf("Error occurred while adding test submission: %v", err)
	}
	// Test adding file for every file in array.
	for _, file := range testFiles {
		if err = testAddFile(file, submissionId); err != nil {
			t.Errorf("Error occurred while adding file: %v", err)
		}
	}
	// clears the test environment
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("error while tearing down test environment: %v", err)
	}
}

// Utility function to add a comment to a given file
func testAddComment(comment *Comment, testFile *File) error {
	// adds a comment to the file
	if err := addComment(comment, testFile.Id); err != nil {
		return errors.New(fmt.Sprintf("failed to add comment to the submission: %v", err))
	}

	// reads the data file into a CodeDataFile struct
	fileDataPath := filepath.Join(
		TEST_FILES_DIR,
		fmt.Sprint(testFile.SubmissionId),
		DATA_DIR_NAME,
		testFile.SubmissionName,
		strings.TrimSuffix(testFile.Path, filepath.Ext(testFile.Path))+".json",
	)
	fileBytes, err := ioutil.ReadFile(fileDataPath)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to read data file: %v", err))
	}
	codeData := &CodeFileData{}
	err = json.Unmarshal(fileBytes, codeData)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to unmarshal code file data: %v", err))
	}

	// extracts the last comment (most recently added) from the comments and checks for equality with
	// the passed in comment
	addedComment := codeData.Comments[len(codeData.Comments)-1]
	if comment.AuthorId != addedComment.AuthorId {
		return errors.New(fmt.Sprintf("Comment author ID mismatch: %s vs %s",
			comment.AuthorId, addedComment.AuthorId))
	}
	return nil
}

// tests adding one valid comment. Uses the testAddComment() utility method
func TestAddOneComment(t *testing.T) {
	testSubmission := testSubmissions[0] // test submission to add testFile to
	testFile := testFiles[0]             // test file to add comments to
	testAuthor := testAuthors[0]         // test author of comment
	testComment := testComments[0]

	// initializes the test environment, returning an error if any occurs
	if err := initTestEnvironment(); err != nil {
		t.Errorf("error while initializing the test environment db: %v", err)
	}

	// creates a submission, adds a file to it, and adds a test user to the system
	submissionId, err := addSubmission(testSubmission)
	if err != nil {
		t.Errorf("failed to add submission: %v", err)
	}
	fileId, err := addFileTo(testFile, submissionId)
	if err != nil {
		t.Errorf("failed to add a file to the submission: %v", err)
	}
	authorId, err := registerUser(testAuthor)
	if err != nil {
		t.Errorf("failed to add user to the database: %v", err)
	}
	testSubmission.Id = submissionId
	testFile.Id = fileId
	testComment.AuthorId = authorId

	// adds a comment to the file and tests that it was added properly
	if err := testAddComment(testComment, testFile); err != nil {
		t.Errorf("error while adding comment: %v", err)
	}

	// clears the test environment
	if err := clearTestEnvironment(); err != nil {
		t.Errorf("error while tearing down test environment: %v", err)
	}
}

// Tests the basic ability of the CodeFiles module to load the data from a
// valid file path passed to it. Simple valid one code file submission
//
// TODO : test whether having / in file path query param breaks the function
//
// Test Depends On:
// 	- TestCreateSubmission()
// 	- TestAddFiles()
func TestGetOneFile(t *testing.T) {
	var err error

	// Set up server to listen with the getFile() function.
	srv := setupCORSsrv()

	// Start server.
	go srv.ListenAndServe()

	var submissionId int                 // stores submission id returned by addSubmission()
	testFile := testFiles[0]             // the test file to be added to the db and filesystem (saved here so it can be easily changed)
	testSubmission := testSubmissions[0] // the test submission to be added to the db and filesystem (saved here so it can be easily changed)

	// initializes the filesystem and db
	if err = initTestEnvironment(); err != nil {
		t.Errorf("Error initializing the test environment %s", err)
	}
	// adds a submission to the database and filesystem
	submissionId, err = addSubmission(testSubmission)
	if err != nil {
		t.Errorf("Error adding submission %s: %v", testSubmission.Name, err)
	}
	// adds a file to the database and filesystem
	_, err = addFileTo(testFile, submissionId)
	if err != nil {
		t.Errorf("Error adding file %s: %v", testFile.Name, err)
	}
	// sets the submission id of the added file to link it with the submission on this end (just in case. This should happen in addFileTo)
	testFile.SubmissionId = submissionId

	// builds the request url inserting query parameters
	urlString := fmt.Sprintf("%s:%s%s?%s=%d&%s=%s", TEST_URL, TEST_SERVER_PORT,
		ENDPOINT_FILE, getJsonTag(&File{}, "SubmissionId"), testFile.SubmissionId,
		getJsonTag(&File{}, "Path"), testFile.Path)
	req, err := http.NewRequest("GET", urlString, nil)

	// send GET request
	resp, err := sendSecureRequest(req, TEAM_ID)
	if err != nil {
		t.Errorf("Error occurred in request: %v", err)
	}
	defer resp.Body.Close()
	// if an error occurred while querying, it's status code is printed here
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Error: %d", resp.StatusCode)
	}

	// marshals the json response into a file struct
	file := &File{}
	err = json.NewDecoder(resp.Body).Decode(&file)
	if err != nil {
		t.Error(err)
	}

	// tests that the file path
	if testFile.Path != file.Path {
		t.Errorf("File Path %d != %d", file.Id, testFile.Id)
		// tests for submission id correctness
	} else if testFile.SubmissionId != file.SubmissionId {
		t.Errorf("File Submission Id %d != %d", file.SubmissionId, testFile.SubmissionId)
		// tests if the file paths are identical
	} else if testFile.Path != file.Path {
		t.Errorf("File Path %s != %s", file.Path, testFile.Path)
		// tests that the file content is correct
	} else if testFile.Content != file.Content {
		t.Error("File Content does not match")
	}

	// destroys the filesystem and db
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("Error occurred while destroying the database and filesystem: %v", err)
	}

	// Close server.
	if err = srv.Shutdown(context.Background()); err != nil {
		fmt.Printf("HTTP server shutdown: %v", err)
	}
}

// Tests the basic ability of the CodeFiles module to upload a single file
// code submission
//
// Test Depends on:
// 	- TestCreateSubmission()
// 	- TestAddFile()
func TestUploadOneFile(t *testing.T) {
	var err error

	// the test values added to the db and filesystem (saved here so it can be easily changed)
	testFile := testFiles[0]
	testAuthor := testAuthors[0]

	// Set up server to listen with the getFile() function.
	srv := setupCORSsrv()

	// Start server.
	go srv.ListenAndServe()

	// initializes the filesystem and db
	if err = initTestEnvironment(); err != nil {
		t.Errorf("Error initializing the test environment %s", err)
	}

	// registers test author
	testAuthor.Id, err = registerUser(testAuthor)
	if err != nil {
		t.Errorf("failed to add test author: %v", err)
	}

	// formats the request body to send to the server to add a comment
	reqBody, err := json.Marshal(map[string]string{
		"author":                       testAuthor.Id,
		getJsonTag(&File{}, "Name"):    testFile.Name,
		getJsonTag(&File{}, "Content"): testFile.Content,
	})
	if err != nil {
		t.Errorf("Error formatting request body: %v", err)
	}

	// formats and executes the request
	req, err := http.NewRequest("POST", fmt.Sprintf("%s:%s%s", TEST_URL, TEST_SERVER_PORT, ENDPOINT_NEWFILE), bytes.NewBuffer(reqBody))
	if err != nil {
		t.Errorf("Error creating request: %v", err)
	}
	resp, err := sendSecureRequest(req, TEAM_ID)
	if err != nil {
		t.Errorf("Error executing request: %v", err)
	}
	defer resp.Body.Close()

	// tests that the result is as desired
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Error: %d", resp.StatusCode)
	}

	// destroys the filesystem and db
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("Error occurred while destroying the database and filesystem: %v", err)
	}

	// Close server.
	if err = srv.Shutdown(context.Background()); err != nil {
		fmt.Printf("HTTP server shutdown: %v", err)
	}
}

// Tests the basic ability of the CodeFiles module to add a comment to a file
// given file path and submission id
//
// Test Depends On:
// 	- TestAddComment()
// 	- TestCreateSubmission()
// 	- TestAddFiles()
func TestUploadUserComment(t *testing.T) {
	var err error

	// the test values added to the db and filesystem (saved here so it can be easily changed)
	testFile := testFiles[0]
	testSubmission := testSubmissions[0]
	testAuthor := testAuthors[0]
	testComment := testComments[0]

	// Set up server to listen with the getFile() function.
	srv := setupCORSsrv()

	// Start server.
	go srv.ListenAndServe()

	var submissionId int // stores submission id returned by addSubmission()

	// initializes the filesystem and db
	if err = initTestEnvironment(); err != nil {
		t.Errorf("Error initializing the test environment %s", err)
	}

	// adds test values to the db and filesystem
	submissionId, err = addSubmission(testSubmission)
	if err != nil {
		t.Errorf("error occurred while adding testSubmission: %v", err)
	}
	_, err = addFileTo(testFile, submissionId)
	if err != nil {
		t.Errorf("error occurred while adding testSubmission: %v", err)
	}
	testAuthor.Id, err = registerUser(testAuthor)
	if err != nil {
		t.Errorf("error occurred while adding testAuthor: %v", err)
	}
	testComment.AuthorId = testAuthor.Id // sets test comment author

	// formats the request body to send to the server to add a comment
	reqBody, err := json.Marshal(map[string]interface{}{
		getJsonTag(&File{}, "SubmissionId"): submissionId,
		getJsonTag(&File{}, "Path"):         testFile.Path,
		getJsonTag(&Comment{}, "AuthorId"):  testAuthor.Id,
		getJsonTag(&Comment{}, "Content"):   testComment.Content,
	})
	if err != nil {
		t.Errorf("Error formatting request body: %v", err)
	}

	// formats and executes the request
	req, err := http.NewRequest("POST", fmt.Sprintf("%s:%s%s", TEST_URL, TEST_SERVER_PORT, ENDPOINT_NEWCOMMENT), bytes.NewBuffer(reqBody))
	if err != nil {
		t.Errorf("Error creating request: %v", err)
	}
	resp, err := sendSecureRequest(req, TEAM_ID)
	if err != nil {
		t.Errorf("Error executing request: %v", err)
	}
	defer resp.Body.Close()

	// tests that the result is as desired
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Error: %d", resp.StatusCode)
	}

	// tests that the comment was added properly
	fileDataPath := filepath.Join(
		TEST_FILES_DIR,
		fmt.Sprint(testSubmission.Id),
		DATA_DIR_NAME,
		testSubmission.Name,
		strings.TrimSuffix(testFile.Path, filepath.Ext(testFile.Path))+".json",
	)
	fileBytes, err := ioutil.ReadFile(fileDataPath)
	if err != nil {
		t.Errorf("failed to read data file: %v", err)
	}
	codeData := &CodeFileData{}
	err = json.Unmarshal(fileBytes, codeData)
	if err != nil {
		t.Errorf("failed to unmarshal code file data: %v", err)
	}

	// extracts the last comment (most recently added) from the comments and checks for equality with
	// the passed in comment
	addedComment := codeData.Comments[len(codeData.Comments)-1]
	if testComment.AuthorId != addedComment.AuthorId {
		t.Errorf("Comment author ID mismatch: %s vs %s", testComment.AuthorId, addedComment.AuthorId)
	} else if testComment.Content != addedComment.Content {
		t.Errorf("Comment content mismatch: %s vs %s", testComment.AuthorId, addedComment.AuthorId)
	}

	// destroys the filesystem and db
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("Error occurred while destroying the database and filesystem: %v", err)
	}

	// Close server.
	if err = srv.Shutdown(context.Background()); err != nil {
		fmt.Printf("HTTP server shutdown: %v", err)
	}
}
