// ========================================
// comments_test.go
// Authors: 190010425
// Created: February 28, 2022
//
// This file tests comments.go
// ========================================

package main

import (
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
			LineNumber:  0,
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
		assert.Equal(t, testComment.LineNumber, addedComment.LineNumber, "line numbers do not match")
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
			LineNumber:  0,
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
