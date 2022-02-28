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
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const (
	// BE VERY CAREFUL WITH THIS PATH!! IT GETS RECURSIVELY REMOVED!!
	TEST_FILES_DIR = "../filesystem/" // environment variable set to this value
)

// NOTE: ID gets set upon file insertion, so these should not be used as pointers in tests
// as to prevent adding a file with the same SubmissionID twice
var testFiles []File = []File{
	{SubmissionID: 0, Path: "testFile1.txt",
		Name: "testFile1.txt", Base64Value: "hello world"},
	{SubmissionID: 0, Path: "testFile2.txt",
		Name: "testFile2.txt", Base64Value: "hello world"},
}

var testComments []*Comment = []*Comment{
	{
		AuthorID:    "",
		Base64Value: "Hello World",
		Comments:    []Comment{},
	},
	{
		AuthorID:    "",
		Base64Value: "Goodbye World",
		Comments:    []Comment{},
	},
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
	// Set up server and configures filesystem/db
	testInit()
	defer testEnd()
	testFile := testFiles[0]             // the test file to be added to the db and filesystem (saved here so it can be easily changed)
	testSubmission := testSubmissions[0] // the test submission to be added to the db and filesystem (saved here so it can be easily changed)

	router := mux.NewRouter()
	router.HandleFunc(SUBROUTE_FILE+"/{id}", getFile)

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

	submissionID, err := addSubmission(&testSubmission)
	if !assert.NoErrorf(t, err, "Error adding submission %s: %v", testSubmission.Name, err) {
		return
	}

	// tests getting a single valid file without comments
	t.Run("Get One File no comments", func(t *testing.T) {
		// adds a file to the database and filesystem
		fileID, err := addFileTo(&testFile, submissionID)
		if !assert.NoErrorf(t, err, "Error adding file %s: %v", testFile.Name, err) {
			return
		}

		// builds the request url inserting query parameters
		urlString := fmt.Sprintf("%s/%d", SUBROUTE_FILE, fileID)
		req, w := httptest.NewRequest("GET", urlString, nil), httptest.NewRecorder()
		router.ServeHTTP(w, req)
		resp := w.Result()

		if !assert.NoErrorf(t, err, "Error occurred in request: %v", err) {
			return
		}
		defer resp.Body.Close()
		if !assert.Equalf(t, resp.StatusCode, http.StatusOK, "Error: %d", resp.StatusCode) {
			return
		} // fails if status code is not 200

		// marshals the json response into a file struct
		file := &File{}
		if !assert.NoError(t, json.NewDecoder(resp.Body).Decode(&file), "Error decoding JSON in server response") {
			return
		}

		// tests that the file was retrieved with the correct information
		switch {
		case !assert.Equal(t, testFile.Path, file.Path, "file paths do not match"),
			!assert.Equal(t, testFile.Name, file.Name, "file names do not match"),
			!assert.Equal(t, submissionID, file.SubmissionID, "Submission IDs do not match"),
			!assert.Equal(t, testFile.Base64Value, file.Base64Value, "File Content does not match"):
			return
		}

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
	// Set up server and configures filesystem/db
	testInit()
	defer testEnd()

	router := mux.NewRouter()
	router.HandleFunc(SUBROUTE_FILE+"/{id}"+ENDPOINT_NEWCOMMENT, uploadUserComment)

	// the test values added to the db and filesystem (saved here so it can be easily changed)
	testFile := testFiles[0]
	testSubmission := testSubmissions[0]
	testAuthor := testAuthors[1] // author of the comment
	testComment := testComments[0]
	testReply := testComments[1]

	// Register submission author.
	subAuthorID, err := registerUser(testAuthors[0], USERTYPE_NIL)
	if assert.NoErrorf(t, err, "Error occurred while registering user: %v", err) {
		return
	}

	// Add submission, and test file linked to submission.
	testSubmission.Authors = []GlobalUser{{ID: subAuthorID}}
	submissionID, err := addSubmission(&testSubmission)
	assert.NoErrorf(t, err, "error occurred while adding test submission: %v", err)
	fileID, err := addFileTo(&testFile, submissionID)
	if assert.NoErrorf(t, err, "error occurred while adding test file: %v", err) {
		return
	}

	// Register comment author and it's bearer token.
	authorID, err := registerUser(testAuthor, USERTYPE_NIL)
	if assert.NoErrorf(t, err, "error occurred while adding testAuthor: %v", err) {
		return
	}

	// upload a single user comment to a valid file in a valid submission
	t.Run("Upload Single User Comment", func(t *testing.T) {
		// formats the request body to send to the server to add a comment
		reqBody, err := json.Marshal(&NewCommentPostBody{
			Base64Value: testComment.Base64Value,
		})
		assert.NoErrorf(t, err, "Error formatting request body: %v", err)

		// formats and executes the request
		req, w := httptest.NewRequest("POST", fmt.Sprintf("%s/%d%s", SUBROUTE_FILE, fileID, ENDPOINT_NEWCOMMENT),
			bytes.NewBuffer(reqBody)), httptest.NewRecorder()

		// sends a request to the server to post a user comment
		router.ServeHTTP(w, req.WithContext(context.WithValue(req.Context(), "userId", testAuthor.ID)))
		resp := w.Result()
		assert.NoErrorf(t, err, "Error executing request: %v", err)
		defer resp.Body.Close()
		assert.Equalf(t, http.StatusOK, resp.StatusCode, "HTTP request error: %d", resp.StatusCode)

		// gets the comment from the db
		addedComment := &Comment{}
		assert.NoError(t, json.NewDecoder(resp.Body).Decode(addedComment), "Error decoding JSON in server response")
		assert.NoError(t, gormDb.Model(addedComment).Find(addedComment).Error, "Could not query added comment")

		// compares the queried comment to that which was sent
		assert.Equal(t, fileID, addedComment.FileID, "file IDs do not match")
		assert.Equal(t, testComment.AuthorID, addedComment.AuthorID, "Comment author ID mismatch")
		assert.Equal(t, testComment.Base64Value, addedComment.Base64Value, "Comment content does not match")
	})

	// upload a single user comment to a valid file in a valid submission
	t.Run("Upload Single Comment Reply", func(t *testing.T) {
		// adds the initial comment without using the server
		testComment.AuthorID = authorID
		commentID, err := addComment(testComment)
		assert.NoError(t, err, "error occurred while adding parent comment")

		// formats the request body to send to the server to add a comment
		reqBody, err := json.Marshal(&NewCommentPostBody{
			ParentID:    &commentID,
			Base64Value: testReply.Base64Value,
		})
		assert.NoErrorf(t, err, "Error formatting request body: %v", err)

		// formats and executes the request
		req, w := httptest.NewRequest("POST", fmt.Sprintf("%s/%d%s", SUBROUTE_FILE, fileID, ENDPOINT_NEWCOMMENT),
			bytes.NewBuffer(reqBody)), httptest.NewRecorder()
		router.ServeHTTP(w, req.WithContext(context.WithValue(req.Context(), "userId", testAuthor.ID)))

		// sends a request to the server to post a user comment
		resp := w.Result()
		defer resp.Body.Close()
		assert.Equalf(t, http.StatusOK, resp.StatusCode, "HTTP request error: %d", resp.StatusCode)

		// gets the added comment via its file to verify the parent -> child structure is correct
		file, err := getFileData(fileID)
		assert.NoError(t, err, "error retrieving test file")
		assert.Equal(t, 1, len(file.Comments), "comment array is incorrect length. Child comment returned on top level of comment tree structure")
		addedReply := file.Comments[0].Comments[0]

		// compares the queried comment to that which was sent
		assert.Equal(t, fileID, addedReply.FileID, "file IDs do not match")
		assert.Equal(t, testReply.AuthorID, addedReply.AuthorID, "Comment author ID mismatch")
		assert.Equal(t, commentID, *addedReply.ParentID, "Parent ID mismatch")
		assert.Equal(t, testReply.Base64Value, addedReply.Base64Value, "Comment content does not match")
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

// tests the ability of the backend to add comments to a given file.
// Test Depends On:
// 	- TestAddSubmission (in submissions_test.go)
// 	- TestAddFile
func TestAddComment(t *testing.T) {
	testInit()
	defer testEnd()

	testSubmission := testSubmissions[0] // test submission to add testFile to
	testFile := testFiles[0]             // test file to add comments to
	testAuthor := testAuthors[1]         // test author of comment
	testComment := testComments[0]
	testReply := testComments[1]

	// adds a submission for the test file to be added to
	authorID, err := registerUser(testAuthors[0], USERTYPE_PUBLISHER)
	if !assert.NoErrorf(t, err, "Error occurred while registering user: %v", err) {
		return
	}
	testSubmission.Authors = []GlobalUser{{ID: authorID}}

	// Add reviewer to a submission.
	reviewerID, err := registerUser(testReviewers[0], USERTYPE_REVIEWER)
	if !assert.NoErrorf(t, err, "Error occurred while registering user: %v", err) {
		return
	}
	testSubmission.Reviewers = []GlobalUser{{ID: reviewerID}}

	// Create submission with given reviewer and author.
	submissionID, err := addSubmission(&testSubmission)
	if !assert.NoErrorf(t, err, "failed to add submission: %v", err) {
		return
	}

	// Create file and add it to the submission.
	fileID, err := addFileTo(&testFile, submissionID)
	if !assert.NoErrorf(t, err, "failed to add file to submission: %v", err) {
		return
	}

	// adds a test user to author a comment
	commentAuthorID, err := registerUser(testAuthor, USERTYPE_PUBLISHER)
	if !assert.NoErrorf(t, err, "failed to add user to the database: %v", err) {
		return
	}
	testComment.AuthorID = commentAuthorID

	// tests adding one valid comment. Uses the testAddComment() utility method
	t.Run("Add One Comment", func(t *testing.T) {
		// adds a comment to the file
		testComment.FileID = fileID
		commentID, err := addComment(testComment)
		if !assert.NoError(t, err, "failed to add comment to the submission") {
			return
		}

		// gets the comment from the db
		addedComment := &Comment{}
		addedComment.ID = commentID
		if !assert.NoError(t, gormDb.Model(addedComment).Find(addedComment).Error, "Could not query added comment") {
			return
		}

		// compares the queried comment to that which was sent
		switch {
		case !assert.Equal(t, fileID, addedComment.FileID, "file IDs do not match"):
		case !assert.Equal(t, testComment.AuthorID, addedComment.AuthorID, "Comment author ID mismatch"):
		case !assert.Equal(t, testComment.Base64Value, addedComment.Base64Value, "Comment content does not match"):
			return
		}
	})

	// tests adding one valid comment. Uses the testAddComment() utility method
	t.Run("Add Comment Reply", func(t *testing.T) {
		// adds a comment to the file
		testComment.FileID = fileID
		commentID, err := addComment(testComment)
		if !assert.NoError(t, err, "failed to add comment to the submission") {
			return
		}

		// adds a reply to the comment
		testReply.AuthorID = authorID
		testReply.FileID = fileID
		testReply.ParentID = &commentID
		_, err = addComment(testReply)
		if !assert.NoError(t, err, "failed to add comment to the submission") {
			return
		}

		// gets the full file back
		file, err := getFileData(fileID)
		if !assert.NoError(t, err, "unable to retrieve file from db") {
			return
		}
		if !assert.Equal(t, 1, len(file.Comments), "comment array is incorrect length. Child comment returned on top level of comment tree structure") {
			return
		}
		queriedComment := file.Comments[0]
		queriedReply := file.Comments[0].Comments[0]

		// checks for equality with comment structure
		switch {
		case !assert.Equal(t, fileID, queriedComment.FileID, "file IDs do not match"):
		case !assert.Equal(t, fileID, queriedReply.FileID, "file IDs do not match"):
		case !assert.Equal(t, testComment.AuthorID, queriedComment.AuthorID, "Comment author ID mismatch"):
		case !assert.Equal(t, testReply.AuthorID, queriedReply.AuthorID, "Reply author ID mismatch"):
		case !assert.Equal(t, testComment.Base64Value, queriedComment.Base64Value, "Comment content does not match"):
		case !assert.Equal(t, testReply.Base64Value, queriedReply.Base64Value, "Reply content does not match"):
		case !assert.Empty(t, testComment.ParentID, "ParentID for parent comment is not nil"):
		case !assert.Equal(t, queriedComment.ID, *queriedReply.ParentID, "ParentID of child comment does not match its parent's ID"):
			return
		}
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
		case !assert.Equal(t, testFile.Name, queriedFile.Name, "File names do not match"):
		case !assert.Equal(t, testFile.Path, queriedFile.Path, "File paths do not match"):
		case !assert.Equal(t, testFile.Base64Value, queriedFile.Base64Value, "File Content does not match"):
			return
		}
	})
}
