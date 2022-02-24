// // ===========================
// // submissions_test.go
// // Authors: 190010425
// // Created: November 18, 2021
// //
// // This file takes care of testing
// // submissions.go
// // ===========================

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm/clause"
)

// data to use in the tests
// NOTE: make sure to use these directly, not as pointers, so that the
// .ID field will not be set in any test
var testSubmissions []Submission = []Submission{
	// valid
	{
		Name:       "TestSubmission1",
		License:    "MIT",
		Authors:    []GlobalUser{},
		Reviewers:  []GlobalUser{},
		Files:      []File{},
		Categories: []Category{{Tag: "testtag"}},
		MetaData: &SubmissionData{
			Abstract: "test abstract",
			Reviews:  []*Comment{},
		},
	},
	{
		Name:       "TestSubmission2",
		License:    "MIT",
		Authors:    []GlobalUser{},
		Reviewers:  []GlobalUser{},
		Files:      []File{},
		Categories: []Category{{Tag: "testtag"}},
		MetaData: &SubmissionData{
			Abstract: "test abstract",
			Reviews:  []*Comment{},
		},
	},
}

// TODO: add comments here
var testSubmissionMetaData = []*SubmissionData{
	{Abstract: "test abstract, this means nothing", Reviews: nil},
}
var testAuthors []User = []User{
	{Email: "paul@test.com", Password: "123456aB$", FirstName: "paul",
		LastName: "test", PhoneNumber: "0574349206"},
	{Email: "john.doe@test.com", Password: "dlbjDs2!", FirstName: "John",
		LastName: "Doe", Organization: "TestOrg"},
	{Email: "author2@test.net", Password: "dlbjDs2!", FirstName: "Jane",
		LastName: "Doe"},
	{Email: "author3@test.net", Password: "dlbjDs2!", FirstName: "Adam",
		LastName: "Doe"},
}
var testReviewers []User = []User{
	{Email: "dave@test.com", Password: "123456aB$", FirstName: "dave",
		LastName: "smith", PhoneNumber: "0574349206"},
	{Email: "Geoff@test.com", Password: "dlbjDs2!", FirstName: "Geoff",
		LastName: "Williams", Organization: "TestOrg"},
	{Email: "reviewer2@test.net", Password: "dlbjDs2!", FirstName: "Jane",
		LastName: "Doe"},
	{Email: "reviewer3@test.net", Password: "dlbjDs2!", FirstName: "Adam",
		LastName: "Doe"},
}

// Initialise mock data in the database for use later on in the testing.
func initMockUsers(t *testing.T) ([]GlobalUser, []GlobalUser, error) {
	// Fill database with users.
	var err error
	globalAuthors := make([]GlobalUser, len(testAuthors))
	for i, user := range testAuthors {
		if globalAuthors[i].ID, err = registerUser(user, USERTYPE_PUBLISHER); err != nil {
			t.Errorf("User registration failed: %v", err)
			return nil, nil, err
		}

	}
	globalReviewers := make([]GlobalUser, len(testReviewers))
	for i, user := range testReviewers {
		if globalReviewers[i].ID, err = registerUser(user, USERTYPE_REVIEWER); err != nil {
			t.Errorf("User registration failed: %v", err)
			return nil, nil, err
		}

	}
	return globalAuthors, globalReviewers, nil
}

// ------------
// Router Function Tests
// ------------

// Tests that submissions.go can upload submissions properly
func TestUploadSubmission(t *testing.T) {
	// Set up server and configures filesystem/db
	testInit()
	defer testEnd()

	// Create mux router
	router := mux.NewRouter()
	router.HandleFunc(ENDPOINT_SUBMISSIONS+ENDPOINT_UPLOAD_SUBMISSION, uploadSubmission)
	route := ENDPOINT_SUBMISSIONS + ENDPOINT_UPLOAD_SUBMISSION

	globalAuthors, globalReviewers, err := initMockUsers(t)
	if err != nil {
		return
	}

	// Check all cases in which an uploaded submission is valid.
	t.Run("Upload valid submissions", func(t *testing.T) {
		testValidUpload := func(submission UploadSubmissionBody, t *testing.T) bool {
			// Get body marshal then send request.
			reqBody, err := json.Marshal(submission)
			if !assert.NoErrorf(t, err, "Marshalling should not error, but got: %v", err) {
				return false
			}
			req, w := httptest.NewRequest(http.MethodPost, route, bytes.NewBuffer(reqBody)), httptest.NewRecorder()
			router.ServeHTTP(w, req.WithContext(context.WithValue(req.Context(), "userId", globalAuthors[0].ID)))
			resp := w.Result()

			// Check success and response.
			var respBody UploadSubmissionResponse
			if err := json.NewDecoder(resp.Body).Decode(&respBody); !assert.NoError(t, err, "Response in invalid format!") {
				return false
			} else if !assert.Equalf(t, http.StatusOK, resp.StatusCode, "Response should succeed, but got: %d - %s", resp.StatusCode, respBody.Message) {
				return false
			} else if !assert.NotEqual(t, "", respBody.SubmissionID, "Returned ID should not be nil!") {
				return false
			}
			return true
		}

		t.Run("Simple submission", func(t *testing.T) {
			// Valid submission - Minimum amount of information.
			testSubmission := UploadSubmissionBody{Name: "Test", Authors: []string{globalAuthors[0].ID}}
			testValidUpload(testSubmission, t)
		})
		t.Run("Full submission", func(t *testing.T) {
			// Valid submission - Minimum amount of information.
			testSubmission := UploadSubmissionBody{
				Name: "Test", Authors: []string{globalAuthors[0].ID},
				Reviewers: []string{globalReviewers[0].ID},
				Files: []File{
					{Name: "test.txt", Path: "test.txt", Base64Value: "test"}, // Check correct file paths.
					{Name: "test.txt", Path: "test/test.txt", Base64Value: "test"},
				},
			}
			testValidUpload(testSubmission, t)
		})
	})

	// Check all invalid upload cases.
	t.Run("Invalid uploads", func(t *testing.T) {

		// Function to check if a request returns valid error.
		testInvalidUpload := func(submission UploadSubmissionBody, status int, authed bool, t *testing.T) bool {
			// Get body marshal then send request.
			reqBody, err := json.Marshal(submission)
			if !assert.NoErrorf(t, err, "Marshalling should not error, but got: %v", err) {
				return false
			}
			req, w := httptest.NewRequest(http.MethodPost, route, bytes.NewBuffer(reqBody)), httptest.NewRecorder()
			if authed {
				router.ServeHTTP(w, req.WithContext(context.WithValue(req.Context(), "userId", globalAuthors[0].ID)))
			} else {
				router.ServeHTTP(w, req)
			}
			resp := w.Result()

			return assert.Equalf(t, status, resp.StatusCode, "Should return code %d but got %d", status, resp.StatusCode)
		}

		// Suite of cases in which submission upload should fail.
		t.Run("No authors", func(t *testing.T) {
			testSubmission := UploadSubmissionBody{
				Name: "Test", License: "MIT",
			}
			testInvalidUpload(testSubmission, http.StatusBadRequest, true, t)
		}) // Return: 400
		t.Run("Unregistered author", func(t *testing.T) {
			testSubmission := UploadSubmissionBody{
				Name: "Test", License: "MIT",
				Authors: []string{"-"},
			}
			testInvalidUpload(testSubmission, http.StatusUnauthorized, true, t)
		}) // Return: 401
		t.Run("No name", func(t *testing.T) {
			testSubmission := UploadSubmissionBody{
				Authors: []string{globalAuthors[0].ID},
			}
			testInvalidUpload(testSubmission, http.StatusBadRequest, true, t)
		}) // Return: 400
		t.Run("Unauthenticated user", func(t *testing.T) {
			testSubmission := UploadSubmissionBody{Name: "Test", Authors: []string{globalAuthors[0].ID}}
			testInvalidUpload(testSubmission, http.StatusUnauthorized, false, t)
		}) // Return: 401

	})
}

// Tests the ability of the submissions file to get a submission from the db
//
// Test Depends On:
// 	- TestCreateSubmissions()
// 	- TestAddFiles()
// 	- TestAddReviewers()
// 	- TestAddAuthors()
func TestRouteGetSubmission(t *testing.T) {
	// Set up server and test environment
	testInit()
	defer testEnd()

	// Create mux router
	router := mux.NewRouter()
	router.HandleFunc(SUBROUTE_SUBMISSION+"/{id}", RouteGetSubmission)

	// Initialise users and created submissions.
	globalAuthors, _, err := initMockUsers(t)
	if err != nil {
		return
	}
	submission := Submission{
		Name:    "Test",
		Authors: []GlobalUser{globalAuthors[0]},
		MetaData: &SubmissionData{
			Abstract: "Test",
		},
	}
	id, err := addSubmission(&submission)
	if !assert.NoError(t, err, "Submission creation shouldn't error!") {
		return
	}

	// tests that a single valid submission with one reviewer and one author can be retrieved
	t.Run("Get Valid Submission", func(t *testing.T) {
		// Create submission, then send request.
		url := fmt.Sprintf("%s/%d", SUBROUTE_SUBMISSION, id)
		r, w := httptest.NewRequest(http.MethodPost, url, nil), httptest.NewRecorder()
		router.ServeHTTP(w, r)
		resp := w.Result()

		// Read result and check success.
		var respBody Submission
		if !assert.Equalf(t, http.StatusOK, resp.StatusCode, "Should succeed, but got error %d", resp.StatusCode) {
			return
		} else if err := json.NewDecoder(resp.Body).Decode(&respBody); !assert.NoError(t, err, "Response schema is invalid.") {
			return
		}
		switch {
		case !assert.Equal(t, id, respBody.ID, "Returned submission should be the same as the one created."),
			!assert.Equal(t, submission.Authors[0].ID, respBody.Authors[0].ID, "Authors should be returned by the request."),
			!assert.Equal(t, submission.MetaData.Abstract, respBody.MetaData.Abstract, "Metadata should be included in the result."):
			return
		}
	})

	t.Run("Get non-existant Submission", func(t *testing.T) {
		// Send request with submission ID that has no submission mapped to it.
		r, w := httptest.NewRequest(http.MethodPost, SUBROUTE_SUBMISSION+"/21474836", nil), httptest.NewRecorder()
		router.ServeHTTP(w, r)
		resp := w.Result()
		assert.Equalf(t, http.StatusNotFound, resp.StatusCode, "Request should return nothing, but instead got %d", resp.StatusCode)
	})
}

// ------------
// Helper Function Tests
// ------------

// test the addSubmission() function in submissions.go
func TestAddSubmission(t *testing.T) {
	testInit()
	defer testEnd()

	// Get authors and reviewers
	globalAuthors, globalReviewers, err := initMockUsers(t)
	if err != nil {
		return
	}

	// Define full testing submission
	FULL_SUBMISSION := Submission{
		Name: "Test", License: "Test", // Basic fields
		Authors:   globalAuthors,   // Check for authors
		Reviewers: globalReviewers, // Check for reviewers
		Files: []File{
			{Name: "test.txt", Path: "test.txt", Base64Value: "test"}, // Check correct file paths.
			{Name: "test.txt", Path: "test/test.txt", Base64Value: "test"},
		},
		MetaData: &SubmissionData{
			Abstract: "test", // Check that metadata is correctly stored.
			Reviews: []*Comment{
				{AuthorID: globalReviewers[0].ID, Base64Value: "test"},
			},
		},
	}

	// Utility function to be re-used for testing adding submissions to the db
	testAddSubmission := func(testSub *Submission) {
		// adds the submission to the db and filesystem
		_, err := addSubmission(testSub)
		assert.NoErrorf(t, err, "Error adding submission: %v", err)

		// retrieve the submission
		queriedSubmission := &Submission{}
		err = gormDb.Model(&Submission{}).First(queriedSubmission, testSub.ID).Error
		assert.NoError(t, err, "Error retrieving submission: %v", err)

		// checks that the filesystem has a proper corresponding entry and metadata file
		submissionData := &SubmissionData{}
		submissionDirPath := getSubmissionDirectoryPath(*testSub)
		fileDataPath := filepath.Join(submissionDirPath, "data.json")
		dataString, err := ioutil.ReadFile(fileDataPath)
		switch {
		case !assert.NoError(t, err, "error reading submission data"),
			assert.NoError(t, json.Unmarshal(dataString, submissionData), "error unmarshalling submission data"):
			return
		}

		// for each file in the submission, checks that it was added to the filesystem and database properly
		for _, file := range testSub.Files {
			// retrieve the file
			queriedFile := &File{}
			if err := gormDb.Model(&File{}).First(queriedFile, file.ID).Error; !assert.NoError(t, err, "Error retrieving file: %v", err) {
				return
			}

			// gets the file content from the filesystem
			fileBytes, err := ioutil.ReadFile(queriedFile.Path)
			if !assert.NoErrorf(t, err, "File read failure after added to filesystem: %v", err) {
				return
			}
			queriedFileContent := string(fileBytes)

			// gets data about the file, and tests it for equality against the added file
			_, err = os.Stat(fileDataPath)
			switch {
			case !assert.NotErrorIs(t, err, os.ErrNotExist, "Data file not generated during file upload"),
				!assert.Equal(t, file.Name, queriedFile.Name, "File names do not match"),
				!assert.Equal(t, file.Path, queriedFile.Path, "File Paths do not match"),
				!assert.Equal(t, file.SubmissionID, queriedFile.SubmissionID, "File SubmissionIDs do not match"),
				!assert.Equal(t, file.Base64Value, queriedFileContent, "file content not written to filesystem properly"),
				!assert.ElementsMatch(t, file.Comments, queriedFile.Comments, "File comments do not match"):
				return
			}
		}

		// tests that the metadata is properly formatted
		assert.Equalf(t, submissionData.Abstract, testSub.MetaData.Abstract,
			"submission metadata not added to filesystem properly. Abstracts %s, %s do not match",
			submissionData.Abstract, testSub.MetaData.Abstract)
		assert.ElementsMatch(t, submissionData.Reviews, testSub.MetaData.Reviews, "Submission Reviews do not match")
	}
	// tests that multiple submissions can be added in a row properly
	t.Run("Add Full Submission", func(t *testing.T) {
		testAddSubmission(&FULL_SUBMISSION)
	})

	// tests that trying to add a nil submission to the db and filesystem will result in an error
	t.Run("Invalid cases do not change the database and filesystem's state", func(t *testing.T) {
		verifyRollback := func(submission *Submission) bool {
			_, err := addSubmission(submission)
			if !assert.Error(t, err, "No error occured while uploading nil submission") {
				return false
			} else if submission != nil {
				_, err := os.Stat(getSubmissionDirectoryPath(*submission))
				switch {
				case !assert.True(t, os.IsNotExist(err), "The submission's directory should not have been created."):
					return false
				}
			}
			return true
		}
		t.Run("Add Nil Submission", func(t *testing.T) {
			verifyRollback(nil)
		})
		t.Run("Duplicate files", func(t *testing.T) {
			BadFilesSubmission := Submission{
				Name: "Test", Authors: globalAuthors,
				Files: []File{
					{Name: "test.txt", Path: "test.txt"},
					{Name: "test.txt", Path: "test.txt"},
				},
			}
			verifyRollback(&BadFilesSubmission)
		})
	})
}

// tests the getSubmission() function, which returns a submission struct
//
// Test Depends On:
// 	- TestAddSubmission
// 	- TestAddFile
func TestGetSubmission(t *testing.T) {
	testInit()
	defer testEnd()

	testSubmission := *testSubmissions[0].getCopy()
	testFile := testFiles[0]
	testAuthor := testAuthors[0]
	testReviewer := testReviewers[0]

	// sets up test environment, and adds a submission with one file to the db and filesystem

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

	testSubmission.Files = []File{testFile}
	submissionID, err := addSubmission(&testSubmission)
	if !assert.NoErrorf(t, err, "Error occurred while adding submission: %v", err) {
		return
	}

	// tests the basic case of getting back a valid submission
	t.Run("Single Valid Submission", func(t *testing.T) {

		// gets the submission back
		queriedSubmission, err := getSubmission(submissionID)
		if !assert.NoErrorf(t, err, "Error occurred while retrieving submission: %v", err) {
			return
		}

		// tests the submission was returned properly
		switch {
		case !assert.Equal(t, testSubmission.Name, queriedSubmission.Name, "Submission names do not match"),
			!assert.Equal(t, testSubmission.License, queriedSubmission.License, "Submission Licenses do not match"),
			!assert.ElementsMatch(t, getTagArray(testSubmission.Categories), getTagArray(queriedSubmission.Categories), "Submission tags do not match"),
			!assert.Equal(t, testSubmission.MetaData.Abstract, queriedSubmission.MetaData.Abstract, "Abstracts do not match"):
			return
		}

		// tests authors
		authorIDs := []string{}
		for _, author := range queriedSubmission.Authors {
			authorIDs = append(authorIDs, author.ID)
		}
		testAuthorIDs := []string{}
		for _, author := range testSubmission.Authors {
			testAuthorIDs = append(testAuthorIDs, author.ID)
		}
		assert.ElementsMatch(t, testAuthorIDs, authorIDs, "author IDs don't match")

		// tests reviewers
		testReviewerIDs := []string{}
		for _, reviewer := range testSubmission.Reviewers {
			testReviewerIDs = append(testReviewerIDs, reviewer.ID)
		}
		reviewerIDs := []string{}
		for _, reviewer := range queriedSubmission.Reviewers {
			reviewerIDs = append(reviewerIDs, reviewer.ID)
		}
		assert.ElementsMatch(t, testReviewerIDs, reviewerIDs, "reviewer IDs don't match")

		// tests files
		testFiles := []File{}
		for _, file := range testSubmission.Files {
			testFiles = append(testFiles, File{Name: file.Name, Path: file.Path})
		}
		files := []File{}
		for _, file := range queriedSubmission.Files {
			files = append(files, File{Name: file.Name, Path: file.Path})
		}
		assert.ElementsMatch(t, testFiles, files, "reviewer IDs don't match")
	})

	// tests trying to get an invalid submission
	t.Run("Invalid Submission", func(t *testing.T) {
		_, err := getSubmission(100)
		assert.Errorf(t, err, "No error was thrown for invalid submission")
	})

	t.Run("Delete Submission", func(t *testing.T) {
		if err := gormDb.Select(clause.Associations).Delete(&testSubmission).Error; !assert.NoError(t, err, "Submission deletion should not error!") {
			return
		}
		_, err := getSubmission(testSubmission.ID)
		assert.Error(t, err, "No error thrown for deleted submission.")
	})
}

// This function tests the getSubmissionMetaData function
//
// This test depends on:
// 	- TestAddSubmission
func TestGetSubmissionMetaData(t *testing.T) {
	testInit()
	defer testEnd()

	testSubmission := *testSubmissions[0].getCopy()
	testAuthor := testAuthors[0]

	authorID, err := registerUser(testAuthor, USERTYPE_PUBLISHER)
	if !assert.NoErrorf(t, err, "Error occurred while registering user: %v", err) {
		return
	}
	testSubmission.Authors = []GlobalUser{{ID: authorID}}

	submissionID, err := addSubmission(&testSubmission)
	if !assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err) {
		return
	}

	// valid metadata file and format
	t.Run("Valid Metadata", func(t *testing.T) {
		// tests that the metadata can be read back properly, and that it matches the uploaded submission
		submissionData, err := getSubmissionMetaData(submissionID)
		switch {
		case !assert.NoErrorf(t, err, "Error getting submission metadata: %v", err),
			!assert.Equalf(t, submissionData.Abstract, testSubmission.MetaData.Abstract,
				"submission metadata not added to filesystem properly. Abstracts %s, %s do not match",
				submissionData.Abstract, testSubmission.MetaData.Abstract),
			!assert.ElementsMatch(t, submissionData.Reviews, testSubmission.MetaData.Reviews, "Submission Reviews do not match"):

		}
	})

	// Tests that getSubmissionMetaData will throw an error if an incorrect submission ID is passed in
	t.Run("Invalid Submission ID", func(t *testing.T) {
		_, err := getSubmissionMetaData(400)
		assert.Errorf(t, err, "No error was thrown for invalid submission")
	})
}
