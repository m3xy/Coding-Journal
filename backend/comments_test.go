// ========================================
// comments_test.go
// Authors: 190010425
// Created: February 28, 2022
//
// This file tests comments.go
// ========================================

package main

import (
	"gorm.io/gorm/clause"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

var testComments []*Comment = []*Comment{
	{
		AuthorID:    "",
		Base64Value: "Hello World",
		Comments:    []Comment{},
		LineNumber: 0,
	},
	{
		AuthorID:    "",
		Base64Value: "Goodbye World",
		Comments:    []Comment{},
		LineNumber: 0,
	},
}

// -------------
// Router Function Tests
// -------------

// Tests the basic ability of the CodeFiles module to add a comment to a file
// given file path and submission id
func TestUploadUserComment(t *testing.T) {
	// Set up server and configures filesystem/db
	testInit()
	defer testEnd()

	router := mux.NewRouter()
	router.HandleFunc(SUBROUTE_FILE+"/{id}"+ENDPOINT_NEWCOMMENT, PostUploadUserComment)

	// the test values added to the db and filesystem (saved here so it can be easily changed)
	testFile := testFiles[0]
	testSubmission := testSubmissions[0]

	// Register test users.
	globalAuthors, globalReviewers, err := initMockUsers(t)
	if !assert.NoError(t, err, "error registering mock users") {
		return
	}

	// Add submission, and test file linked to submission.
	testSubmission.Authors = globalAuthors[:1]
	testSubmission.Files = []File{testFile}
	_, err = addSubmission(&testSubmission)
	if !assert.NoErrorf(t, err, "error occurred while adding test submission: %v", err) { return }
	// gets the file ID from the files table
	if !assert.NoError(t, gormDb.Model(&File{}).Find(&testFile).Error, "error occurred while getting file ID") { return }
	fileID := testFile.ID

	// clears the comments for the test file
	clearComments := func() {
		var comments []Comment
		assert.NoError(t, gormDb.Find(&comments).Error, "error while clearing comments table")
		for _, comment := range comments {
			gormDb.Select(clause.Associations).Unscoped().Delete(&comment)
		}
	}

	// tests that two given comment structs are equal
	testCommentEquality := func(c1 *Comment, c2 *Comment) {
		switch {
		case !assert.Equal(t, c1.FileID, c2.FileID, "file IDs do not match"),
			!assert.Equal(t, c1.AuthorID, c2.AuthorID, "Comment author ID mismatch"),
			!assert.Equal(t, c1.Base64Value, c2.Base64Value, "Comment content does not match"),
			!assert.Equal(t, c1.LineNumber, c2.LineNumber, "line numbers do not match"):
			return
		}
	}

	t.Run("valid upload", func(t *testing.T) {
		defer clearComments()

		// configures sub-test values
		commentAuthor := globalReviewers[1]
		testComment := testComments[0]
		testComment.FileID = fileID
		testComment.AuthorID = commentAuthor.ID
		testReply := testComments[1]
		testReply.FileID = fileID
		testReply.AuthorID = commentAuthor.ID

		// sends request to upload a comment
		handleRequest := func(queryRoute string, fileID uint, reqStruct *NewCommentPostBody) *http.Response {
			reqBody, err := json.Marshal(reqStruct)
			assert.NoErrorf(t, err, "Error formatting request body: %v", err)
			// formats and executes the request
			req, w := httptest.NewRequest("POST", fmt.Sprintf("%s/%d%s", SUBROUTE_FILE, fileID, ENDPOINT_NEWCOMMENT),
				bytes.NewBuffer(reqBody)), httptest.NewRecorder()
			// sends a request to the server to post a user comment
			ctx := context.WithValue(req.Context(), "data", &RequestContext{
				ID: commentAuthor.ID,
				UserType: USERTYPE_NIL,
			})
			router.ServeHTTP(w, req.WithContext(ctx))
			return w.Result()
		}

		// upload a single user comment to a valid file in a valid submission
		t.Run("Upload Single User Comment", func(t *testing.T) {
			// formats the request body to send to the server to add a comment
			queryRoute := fmt.Sprintf("%s/%d%s", SUBROUTE_FILE, fileID, ENDPOINT_NEWCOMMENT)
			resp := handleRequest(queryRoute, fileID, &NewCommentPostBody{
				ParentID: nil,
				Base64Value: testComment.Base64Value,
				LineNumber:  testComment.LineNumber,
			})
			defer resp.Body.Close()
			assert.Equalf(t, http.StatusOK, resp.StatusCode, "HTTP request error: %d", resp.StatusCode)

			// gets the comment from the db
			respBody := &NewCommentResponse{}
			addedComment := &Comment{}
			assert.NoError(t, json.NewDecoder(resp.Body).Decode(respBody), "Error decoding JSON in server response")
			assert.NoError(t, gormDb.Model(&Comment{}).Find(addedComment, respBody.ID).Error, "Could not query added comment")
			testCommentEquality(testComment, addedComment)

			testComment.ID = respBody.ID // this is important for the reply test
		})

		// upload a single user comment to a valid file in a valid submission
		t.Run("Upload Single Comment Reply", func(t *testing.T) {
			// formats the request body to send to the server to add a comment
			queryRoute := fmt.Sprintf("%s/%d%s", SUBROUTE_FILE, fileID, ENDPOINT_NEWCOMMENT)
			resp := handleRequest(queryRoute, fileID, &NewCommentPostBody{
				ParentID:    &testComment.ID,
				LineNumber:  0,
				Base64Value: testReply.Base64Value,
			})
			defer resp.Body.Close()
			assert.Equalf(t, http.StatusOK, resp.StatusCode, "HTTP request error: %d", resp.StatusCode)

			// gets the added comment via its file to verify the parent -> child structure is correct
			file, err := getFileData(fileID)
			assert.NoError(t, err, "error retrieving test file")
			assert.Equal(t, 1, len(file.Comments), "comment array is incorrect length. Child comment returned on top level of comment tree structure")
			addedReply := file.Comments[0].Comments[0]

			testCommentEquality(testReply, &addedReply)
		})
	})

	t.Run("Request Validation", func(t *testing.T) {
		t.Run("Not logged in", func(t *testing.T) {
			defer clearComments()
			reqStruct := &NewCommentPostBody{
				ParentID: nil,
				Base64Value: testComments[0].Base64Value,
				LineNumber: 0,
			}
			reqBody, err := json.Marshal(reqStruct)
			assert.NoErrorf(t, err, "Error formatting request body: %v", err)
			// formats and executes the request
			req, w := httptest.NewRequest("POST", fmt.Sprintf("%s/%d%s", SUBROUTE_FILE, fileID, ENDPOINT_NEWCOMMENT),
				bytes.NewBuffer(reqBody)), httptest.NewRecorder()
			router.ServeHTTP(w, req)
			resp := w.Result()
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "status code incorrect")
		})

		t.Run("Bad Request Body", func(t *testing.T) {
			defer clearComments()
			reqStruct := &NewCommentPostBody{
				ParentID: nil,
				LineNumber: 0,
			}
			reqBody, err := json.Marshal(reqStruct)
			assert.NoErrorf(t, err, "Error formatting request body: %v", err)
			// formats and executes the request
			req, w := httptest.NewRequest("POST", fmt.Sprintf("%s/%d%s", SUBROUTE_FILE, fileID, ENDPOINT_NEWCOMMENT),
				bytes.NewBuffer(reqBody)), httptest.NewRecorder()
			ctx := context.WithValue(req.Context(), "data", &RequestContext{
				ID: globalReviewers[0].ID,
				UserType: USERTYPE_NIL,
			})
			router.ServeHTTP(w, req.WithContext(ctx))
			resp := w.Result()
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "status code incorrect")
		})
	})
}

// -------------
// Helper Function Tests
// -------------

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

	// Add submission, and test file linked to submission.
	testSubmission.Files = []File{testFile}
	_, err = addSubmission(&testSubmission)
	if !assert.NoErrorf(t, err, "error occurred while adding test submission: %v", err) { return }
	// gets the file ID from the files table
	if !assert.NoError(t, gormDb.Model(&File{}).Find(&testFile).Error, "error occurred while getting file ID") { return }
	fileID := testFile.ID

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
		case !assert.Equal(t, fileID, queriedComment.FileID, "file IDs do not match"),
			!assert.Equal(t, fileID, queriedReply.FileID, "file IDs do not match"),
			!assert.Equal(t, testComment.AuthorID, queriedComment.AuthorID, "Comment author ID mismatch"),
			!assert.Equal(t, testReply.AuthorID, queriedReply.AuthorID, "Reply author ID mismatch"),
			!assert.Equal(t, testComment.Base64Value, queriedComment.Base64Value, "Comment content does not match"),
			!assert.Equal(t, testReply.Base64Value, queriedReply.Base64Value, "Reply content does not match"),
			!assert.Empty(t, testComment.ParentID, "ParentID for parent comment is not nil"),
			!assert.Equal(t, queriedComment.ID, *queriedReply.ParentID, "ParentID of child comment does not match its parent's ID"):
			return
		}
	})
}
