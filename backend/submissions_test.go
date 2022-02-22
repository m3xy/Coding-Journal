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
)

const (
	TEST_PORT_SUBMISSION = ":59217"
	ADDRESS_SUBMISSION   = "http://localhost:59217"
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
		Categories: []string{"testtag"},
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
		Categories: []string{"testtag"},
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

// Set up server used for submissions testing.
func submissionServerSetup() *http.Server {
	router := mux.NewRouter()

	getSubmissionsSubRoutes(router)
	router.HandleFunc(SUBROUTE_USER+"/{id}"+ENDPOINT_SUBMISSIONS, getAllAuthoredSubmissions).Methods(http.MethodGet)

	return &http.Server{
		Addr:    TEST_PORT_SUBMISSION,
		Handler: router,
	}
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
		submissionID, err := addSubmission(testSub)
		assert.NoErrorf(t, err, "Error adding submission: %v", err)

		// retrieve the submission
		queriedSubmission := &Submission{}
		err = gormDb.Model(&Submission{}).First(queriedSubmission, testSub.ID).Error
		assert.NoError(t, err, "Error retrieving submission: %v", err)

		// checks that the filesystem has a proper corresponding entry and metadata file
		submissionData := &SubmissionData{}
		submissionDirPath := filepath.Join(TEST_FILES_DIR, fmt.Sprint(submissionID))
		fileDataPath := filepath.Join(submissionDirPath, DATA_DIR_NAME, testSub.Name+".json")
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
	t.Run("Add Nil Submission", func(t *testing.T) {
		_, err := addSubmission(nil)
		assert.Error(t, err, "No error occurred while uploading nil submission")
	})
}

// tests the ability of the submissions.go module to add reviewers to submissions
// Test Depends on:
// 	- TestAddSubmission
func TestAppendUsers(t *testing.T) {
}

// This function tests the addTags function
//
// This test depends on:
// 	- TestAddSubmission
func TestAddTags(t *testing.T) {
	// utility function to add a submission to the db (no authors required here)
	testAddSubmission := func(submission *Submission) uint {
		assert.NoError(t, gormDb.Create(submission).Error, "Error creating submission")
		return submission.ID
	}

	// standard use case
	t.Run("Add Valid Tag", func(t *testing.T) {
		testTag := "TEST"

		// sets up the test environment, and uploads a test submission
		testInit()
		testSubmission := testSubmissions[0]
		submissionID := testAddSubmission(&testSubmission)

		// adds a tag to the submission
		addTags(gormDb, []string{testTag}, submissionID)

		// queries the db to make sure the tag was added properly
		cat := &Category{}
		assert.NoError(t, gormDb.Model(&Category{SubmissionID: submissionID, Tag: testTag}).Find(cat).Error, "Tag unable to be retrieved")
		testEnd()
	})

	// duplicate tag case
	t.Run("Add Duplicate Tag", func(t *testing.T) {
		testTag := "TEST"

		// sets up the test environment, and uploads a test submission
		testInit()
		testSubmission := testSubmissions[0]
		submissionID := testAddSubmission(&testSubmission)

		// adds a tag to the submission twice (second prints to log as it violates a key constraint)
		assert.NoError(t, addTags(gormDb, []string{testTag}, submissionID), "adding first tag caused an error")
		assert.Error(t, addTags(gormDb, []string{testTag}, submissionID), "attempting to add duplicate tag does not return an error")

		testEnd()
	})

	// 2 identical tags on different submissions
	t.Run("Add Same Tag to Different Submissions", func(t *testing.T) {
		testTag := "TEST"

		// sets up the test environment, and uploads test submissions
		testInit()
		testSubmission1 := testSubmissions[0]
		testSubmission2 := testSubmissions[1]
		submissionID1 := testAddSubmission(&testSubmission1)
		submissionID2 := testAddSubmission(&testSubmission2)

		// adds the tags
		assert.NoError(t, addTags(gormDb, []string{testTag}, submissionID1), "Error occurred while adding 1st tag")
		assert.NoError(t, addTags(gormDb, []string{testTag}, submissionID2), "Error occurred while adding 2nd tag")

		// queries the db to make sure the tags were added properly
		cat := &Category{}
		assert.NoError(t, gormDb.Model(&Category{SubmissionID: submissionID1, Tag: testTag}).Find(cat).Error, "Tag 1 unable to be retrieved")
		assert.NoError(t, gormDb.Model(&Category{SubmissionID: submissionID2, Tag: testTag}).Find(cat).Error, "Tag 2 unable to be retrieved")

		testEnd()
	})

	// invalid tag case
	t.Run("Add Empty Tag", func(t *testing.T) {
		// sets up the test environment, and uploads a test submission
		testInit()
		testSubmission := testSubmissions[0]
		submissionID := testAddSubmission(&testSubmission)

		// adds an invalid tag to an existing submission
		assert.Error(t, addTags(gormDb, []string{""}, submissionID), "empty tag was able ot be added")

		testEnd()
	})

	// add tag to non-existant project (foreign key constraint fails in db)
	t.Run("Add Tag Invalid Project", func(t *testing.T) {
		testInit()
		assert.Error(t, addTags(gormDb, []string{"INVALID_PROJECT"}, 10), "Error not thrown when tag added to a non-existant submission")
		testEnd()
	})
}

// test for basic functionality. Adds 2 submissions to the db with different authors, then queries them and tests for equality
// Test Depends On:
// 	- TestAddSubmission
// 	- TestAddAuthors
func TestGetAuthoredSubmissions(t *testing.T) {
	// adds two submissions each with different authors to the db and then queries one author's submissions
	t.Run("Get Single Submission from an Author", func(t *testing.T) {
		testSubmission1 := testSubmissions[0] // test submission to return on getAuthoredSubmissions()
		testSubmission2 := testSubmissions[1] // test submission to not return on getAuthoredSubmissions()
		testAuthor := testAuthors[0]          // test author of the submission being queried
		testNonAuthor := testAuthors[3]       // test author of submission not being queried

		testInit()

		// adds two test users to the db as authors
		authorID, err := registerUser(testAuthor, USERTYPE_PUBLISHER)
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission1.Authors = []GlobalUser{{ID: authorID}}

		nonauthorID, err := registerUser(testNonAuthor, USERTYPE_PUBLISHER) // author of the submission we are not interested in
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission2.Authors = []GlobalUser{{ID: nonauthorID}}

		// adds dummy reviewers
		reviewerID, err := registerUser(testReviewers[0], USERTYPE_REVIEWER)
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission1.Reviewers = []GlobalUser{{ID: reviewerID}}
		testSubmission2.Reviewers = []GlobalUser{{ID: reviewerID}}

		// adds two test submissions to the db
		testSubmission1.ID, err = addSubmission(&testSubmission1)
		assert.NoErrorf(t, err, "Error occurred while adding submission1: %v", err)
		testSubmission2.ID, err = addSubmission(&testSubmission2)
		assert.NoErrorf(t, err, "Error occurred while adding submission2: %v", err)

		// queries all of testAuthor's submissions
		submissions, err := getAuthoredSubmissions(authorID)
		assert.NoErrorf(t, err, "Error getting user submissions: %v", err)

		// tests for equality of submission ID and that testSubmission2.ID is not in the map
		_, ok := submissions[testSubmission2.ID]
		assert.False(t, ok, "Returned submission where the test author is not an author")
		assert.Equalf(t, submissions[testSubmission1.ID], testSubmission1.Name,
			"Returned incorrect submission name: %s", submissions[testSubmission1.ID])

		testEnd()
	})
}

// test for basic functionality. Adds 2 submissions to the db with different authors, then queries them and tests for equality
// Test Depends On:
// 	- TestAddSubmission
// 	- TestAddAuthors
// 	- TestAddReviewers
func TestGetReviewedSubmissions(t *testing.T) {
	// adds two submissions each with different authors to the db and then queries one author's submissions
	t.Run("Get Single Submission from a Reviewer", func(t *testing.T) {
		testSubmission1 := testSubmissions[0] // test submission to return on getAuthoredSubmissions()
		testSubmission2 := testSubmissions[1] // test submission to not return on getAuthoredSubmissions()
		testReviewer := testReviewers[0]      // test author of the submission being queried
		testNonReviewer := testReviewers[3]   // test author of submission not being queried

		testInit()

		// adds two test users to the db as authors
		reviewerID, err := registerUser(testReviewer, USERTYPE_REVIEWER)
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission1.Reviewers = []GlobalUser{{ID: reviewerID}}

		nonreviewerID, err := registerUser(testNonReviewer, USERTYPE_REVIEWER) // author of the submission we are not interested in
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission2.Reviewers = []GlobalUser{{ID: nonreviewerID}}

		// adds dummy authors
		authorID, err := registerUser(testAuthors[0], USERTYPE_PUBLISHER)
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission1.Authors = []GlobalUser{{ID: authorID}}
		testSubmission2.Authors = []GlobalUser{{ID: authorID}}

		// adds two test submissions to the db
		testSubmission1.ID, err = addSubmission(&testSubmission1)
		assert.NoErrorf(t, err, "Error occurred while adding submission1: %v", err)
		testSubmission2.ID, err = addSubmission(&testSubmission2)
		assert.NoErrorf(t, err, "Error occurred while adding submission2: %v", err)

		// queries all of testAuthor's submissions
		submissions, err := getReviewedSubmissions(reviewerID)
		assert.NoErrorf(t, err, "Error getting user reviewed submissions: %v", err)

		// tests for equality of submission ID and that testSubmission2.ID is not in the map
		_, ok := submissions[testSubmission2.ID]
		assert.False(t, ok, "Returned submission where the test reviewer is not a reviewer")
		assert.Equalf(t, submissions[testSubmission1.ID], testSubmission1.Name,
			"Returned incorrect submission name: %s", submissions[testSubmission1.ID])

		testEnd()
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

	testSubmission := testSubmissions[0]
	testFile := testFiles[0]
	testAuthor := testAuthors[0]
	testReviewer := testReviewers[0]

	// sets up test environment, and adds a submission with one file to the db and filesystem

	authorID, err := registerUser(testAuthor, USERTYPE_PUBLISHER)
	if !assert.NoErrorf(t, err, "Error occurred while registering user: %v", err) {

	}
	testSubmission.Authors = []GlobalUser{{ID: authorID}}

	reviewerID, err := registerUser(testReviewer, USERTYPE_REVIEWER)
	if !assert.NoErrorf(t, err, "Error occurred while registering user: %v", err) {

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
		assert.NoErrorf(t, err, "Error occurred while retrieving submission: %v", err)

		// tests the submission was returned properly
		assert.Equal(t, testSubmission.Name, queriedSubmission.Name, "Submission names do not match")
		assert.Equal(t, testSubmission.License, queriedSubmission.License, "Submission Licenses do not match")
		assert.ElementsMatch(t, testSubmission.Categories, queriedSubmission.Categories, "Submission tags do not match")
		assert.Equal(t, testSubmission.MetaData.Abstract, queriedSubmission.MetaData.Abstract, "Abstracts do not match")

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
}

// This function tests the getSubmissionCategories function
//
// This test depends on:
// 	- TestAddSubmission
func TestGetSubmissionCategories(t *testing.T) {
	// valid metadata file and format
	t.Run("Valid Categories", func(t *testing.T) {
		testSubmission := testSubmissions[0]
		testAuthor := testAuthors[0]

		// sets up the test environment, and uploads a test submission
		testInit()
		authorID, err := registerUser(testAuthor, USERTYPE_PUBLISHER)
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission.Authors = []GlobalUser{{ID: authorID}}

		submissionID, err := addSubmission(&testSubmission)
		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)

		// tests that the metadata can be read back properly, and that it matches the uploaded submission
		categories, err := getSubmissionCategories(gormDb, submissionID)
		assert.NoErrorf(t, err, "Error getting submission tags: %v", err)
		assert.ElementsMatch(t, testSubmission.Categories, categories, "Submission tags do not match")
		testEnd()
	})
}

// This function tests the getSubmissionMetaData function
//
// This test depends on:
// 	- TestAddSubmission
func TestGetSubmissionMetaData(t *testing.T) {
	testInit()
	defer testEnd()

	testSubmission := testSubmissions[0]
	testAuthor := testAuthors[0]

	authorID, err := registerUser(testAuthor, USERTYPE_PUBLISHER)
	assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
	testSubmission.Authors = []GlobalUser{{ID: authorID}}

	submissionID, err := addSubmission(&testSubmission)
	assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)

	// valid metadata file and format
	t.Run("Valid Metadata", func(t *testing.T) {
		// tests that the metadata can be read back properly, and that it matches the uploaded submission
		submissionData, err := getSubmissionMetaData(submissionID)
		assert.NoErrorf(t, err, "Error getting submission metadata: %v", err)
		assert.Equalf(t, submissionData.Abstract, testSubmission.MetaData.Abstract,
			"submission metadata not added to filesystem properly. Abstracts %s, %s do not match",
			submissionData.Abstract, testSubmission.MetaData.Abstract)
		assert.ElementsMatch(t, submissionData.Reviews, testSubmission.MetaData.Reviews, "Submission Reviews do not match")
	})

	// Tests that getSubmissionMetaData will throw an error if an incorrect submission ID is passed in
	t.Run("Invalid Submission ID", func(t *testing.T) {
		_, err := getSubmissionMetaData(400)
		assert.Errorf(t, err, "No error was thrown for invalid submission")
	})
}

// This tests converting from the local submission data format to the supergroup specified format
func TestLocalToGlobal(t *testing.T) {
	testInit()
	defer testEnd()

	// adds the submission and a file to the system
	testSubmission := testSubmissions[0]
	testFile := testFiles[0]
	testAuthor := testAuthors[0]
	testReviewer := testReviewers[0]

	// registers authors and reviewers, and adds them to the test submission
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

	// tests valid submission struct
	t.Run("Valid Submission", func(t *testing.T) {
		// gets the supergroup compliant submission
		globalSubmission, err := localToGlobal(submissionID)
		if !assert.NoErrorf(t, err, "Error occurred while converting submission format: %v", err) {
			return
		}

		// compares submission fields
		switch {
		case !assert.Equal(t, testSubmission.Name, globalSubmission.Name, "Names do not match"),
			!assert.Equal(t, testSubmission.License, globalSubmission.MetaData.License,
				"Licenses do not match"),
			!assert.Equal(t, testAuthor.FirstName+" "+testAuthor.LastName, globalSubmission.MetaData.AuthorNames[0],
				"Authors do not match"),
			!assert.Equal(t, testSubmission.Categories, globalSubmission.MetaData.Categories,
				"Tags do not match"),
			!assert.Equal(t, testSubmission.MetaData.Abstract, globalSubmission.MetaData.Abstract,
				"Abstracts do not match"),
			// compares files
			!assert.Equal(t, testFile.Name, globalSubmission.Files[0].Name, "File names do not match"),
			!assert.Equal(t, testFile.Base64Value, globalSubmission.Files[0].Base64Value, "File content does not match"):
			return
		}
	})
}
