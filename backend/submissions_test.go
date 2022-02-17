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
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"bytes"

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
		LastName: "test", PhoneNumber: "0574349206", UserType: USERTYPE_PUBLISHER},
	{Email: "john.doe@test.com", Password: "dlbjDs2!", FirstName: "John",
		LastName: "Doe", Organization: "TestOrg", UserType: USERTYPE_USER},
	{Email: "jane.doe@test.net", Password: "dlbjDs2!", FirstName: "Jane",
		LastName: "Doe", UserType: USERTYPE_REVIEWER},
	{Email: "adam.doe@test.net", Password: "dlbjDs2!", FirstName: "Adam",
		LastName: "Doe", UserType: USERTYPE_REVIEWER_PUBLISHER},
}
var testReviewers []User = []User{
	{Email: "dave@test.com", Password: "123456aB$", FirstName: "dave",
		LastName: "smith", PhoneNumber: "0574349206", UserType: USERTYPE_REVIEWER},
	{Email: "Geoff@test.com", Password: "dlbjDs2!", FirstName: "Geoff",
		LastName: "Williams", Organization: "TestOrg", UserType: USERTYPE_USER},
	{Email: "jane.doe@test.net", Password: "dlbjDs2!", FirstName: "Jane",
		LastName: "Doe", UserType: USERTYPE_PUBLISHER},
	{Email: "adam.doe@test.net", Password: "dlbjDs2!", FirstName: "Adam",
		LastName: "Doe", UserType: USERTYPE_REVIEWER_PUBLISHER},
}

// Set up server used for submissions testing.
func submissionServerSetup() *http.Server {
	router := mux.NewRouter()

	router.HandleFunc(ENDPOINT_SUBMISSION+"/{id}", RouteGetSubmission).Methods(http.MethodGet)
	router.HandleFunc("/{id}"+ENDPOINT_ALL_SUBMISSIONS, getAllAuthoredSubmissions).Methods(http.MethodGet)
	router.HandleFunc(ENDPOINT_UPLOAD_SUBMISSION, uploadSubmission).Methods(http.MethodPost, http.MethodOptions)

	return &http.Server{
		Addr:    TEST_PORT_SUBMISSION,
		Handler: router,
	}
}

// ------------
// Router Function Tests
// ------------

// Tests the ability of the getAllSubmissions() function to get all submission ids and names from the db
// at once
//
// Test Depends On:
// 	- TestAddSubmission
// 	- TestGetAuthoredSubmissions
func TestGetAllSubmissions(t *testing.T) {
	// tests that multiple valid submissions can be uploaded, then retrieved from the database
	t.Run("Get Multiple Valid submissions", func(t *testing.T) {
		// Set up server and test environment
		testInit()
		srv := submissionServerSetup()
		go srv.ListenAndServe()

		// registers authors and reviewers of the submissions (same for all submissions here)
		authorID, err := registerUser(testAuthors[0])
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)

		reviewerID, err := registerUser(testReviewers[0])
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)

		// adds all of the submissions and stores their ids and names
		sentSubmissions := make(map[uint]string) // variable to hold the id: submission name mappings which are sent to the db
		for _, sub := range testSubmissions[0:2] {
			sub.Authors = []GlobalUser{{ID: authorID}}
			sub.Reviewers = []GlobalUser{{ID: reviewerID}}
			submissionID, err := addSubmission(&sub)
			assert.NoErrorf(t, err, "Error adding submission %s: %v", sub.Name, err)
			sentSubmissions[submissionID] = sub.Name
		}

		// builds and sends and http request to get the names and IDs of all submissions
		req, err := http.NewRequest("GET", ADDRESS_SUBMISSION+"/"+authorID+ENDPOINT_ALL_SUBMISSIONS, nil)
		resp, err := sendSecureRequest(gormDb, req, TEAM_ID)
		assert.NoErrorf(t, err, "Error occurred while sending get request to the Go server: %v", err)
		defer resp.Body.Close()
		assert.Equalf(t, http.StatusOK, resp.StatusCode, "Non-OK status returned from GET request: %d", resp.StatusCode)

		// checks the returned list of submissions for equality with the sent list
		returnedSubmissions := make(map[uint]string)
		json.NewDecoder(resp.Body).Decode(&returnedSubmissions)
		for k, v := range returnedSubmissions {
			assert.Equalf(t, v, sentSubmissions[k],
				"Submissions of ids: %d do not have matching names. Given: %s, Returned: %s ", k, sentSubmissions[k], v)
		}

		// clears test env and shuts down the test server
		assert.NoError(t, srv.Shutdown(context.Background()), "failed to shut down server")
		testEnd()
	})
}

// Tests that submissions.go can upload submissions properly
func TestUploadSubmission(t *testing.T) {
	// tests that a single valid submission with one reviewer and one author can be retrieved
	t.Run("Upload Single Valid Submission", func(t *testing.T) {
		testSubmission := testSubmissions[0]
		testFile := testFiles[0]
		testSubmission.Files = []File{testFile}
		testAuthor := testAuthors[0]
		testReviewer := testReviewers[0]

		// Set up server and configures filesystem/db
		testInit()
		srv := submissionServerSetup()
		go srv.ListenAndServe()

		// registers author and reviewer
		authorID, err := registerUser(testAuthor)
		assert.NoErrorf(t, err, "Error registering author in the db: %v", err)
		testSubmission.Authors = []GlobalUser{{ID: authorID}}

		reviewerID, err := registerUser(testReviewer)
		assert.NoErrorf(t, err, "Error registering reviewer in the db: %v", err)
		testSubmission.Reviewers = []GlobalUser{{ID: reviewerID}}

		// constructs the request body
		reqBody, err := json.Marshal(testSubmission)
		assert.NoError(t, err, "Error marshalling test submission to Json")

		// creates a request to send to the test server
		urlString := fmt.Sprintf("%s%s", ADDRESS_SUBMISSION, ENDPOINT_UPLOAD_SUBMISSION)
		req, _ := http.NewRequest("POST", urlString, bytes.NewBuffer(reqBody))
		resp, err := sendSecureRequest(gormDb, req, TEAM_ID)
		assert.NoErrorf(t, err, "Error while sending Post request: %v", err)
		defer resp.Body.Close()
		assert.Equalf(t, http.StatusOK, resp.StatusCode, "Non-OK status returned from GET request: %d", resp.StatusCode)

		// gets the added submission back
		respBody := &Submission{}
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		uploadedSubmission, err := getSubmission(respBody.ID)
		assert.NoError(t, err, "Error retrieving uploaded submission")

		// tests that the returned submission matches the passed in data
		assert.Equal(t, respBody.ID, uploadedSubmission.ID, "Submission IDs do not match")
		assert.Equal(t, testSubmission.Name, uploadedSubmission.Name, "Submission Names do not match.")
		assert.Equal(t, authorID, uploadedSubmission.Authors[0].ID, "Author IDs do not match")
		assert.Equal(t, testAuthor.FirstName+" "+testAuthor.LastName, uploadedSubmission.Authors[0].FullName, "Author Names not match")
		assert.Equal(t, reviewerID, uploadedSubmission.Reviewers[0].ID, "Reviewer IDs not match")
		assert.Equal(t, testReviewer.FirstName+" "+testReviewer.LastName, uploadedSubmission.Reviewers[0].FullName, "Reviewer Names not match")

		// clears test env and shuts down the test server
		assert.NoError(t, srv.Shutdown(context.Background()), "failed to shut down server")
		testEnd()
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
	// tests that a single valid submission with one reviewer and one author can be retrieved
	t.Run("Get Valid Submission", func(t *testing.T) {
		testFile := testFiles[0]             // defines the file to use for the test here so that it can be easily changed
		testSubmission := testSubmissions[0] // defines the submission to use for the test here so that it can be easily changed
		testAuthor := testAuthors[0]         // defines the author of the submission
		testReviewer := testReviewers[0]     // defines the reviewer of the submission

		// Set up server and test environment
		testInit()
		srv := submissionServerSetup()
		go srv.ListenAndServe()

		// registers author and reviewer
		authorID, err := registerUser(testAuthor)
		assert.NoErrorf(t, err, "Error registering author in the db: %v", err)
		testSubmission.Authors = []GlobalUser{{ID: authorID}}

		reviewerID, err := registerUser(testReviewer)
		assert.NoErrorf(t, err, "Error registering reviewer in the db: %v", err)
		testSubmission.Reviewers = []GlobalUser{{ID: reviewerID}}

		// uploads the test submission and adds a file to it
		testSubmission.Files = []File{testFile}
		submissionID, err := addSubmission(&testSubmission)
		assert.NoErrorf(t, err, "Error adding submission %v", err)

		// creates a request to send to the test server
		urlString := fmt.Sprintf("%s%s/%d", ADDRESS_SUBMISSION,
			ENDPOINT_SUBMISSION, submissionID)
		fmt.Println(urlString)
		req, _ := http.NewRequest("GET", urlString, nil)
		resp, err := sendSecureRequest(gormDb, req, TEAM_ID)
		assert.NoErrorf(t, err, "Error while sending Get request: %v", err)
		defer resp.Body.Close()
		assert.Equalf(t, http.StatusOK, resp.StatusCode, "Non-OK status returned from GET request: %d", resp.StatusCode)

		// decodes the json response into a Submission struct
		submission := &Submission{}
		assert.NoErrorf(t, json.NewDecoder(resp.Body).Decode(submission), "Error while decoding server response: %v", err)

		// tests that the returned submission matches the passed in data
		assert.Equal(t, testSubmission.ID, submission.ID, "Submission IDs do not match")
		assert.Equal(t, testSubmission.Name, submission.Name, "Submission Names do not match.")
		assert.Equal(t, authorID, submission.Authors[0].ID, "Author IDs do not match")
		assert.Equal(t, testAuthor.FirstName+" "+testAuthor.LastName, submission.Authors[0].FullName, "Author Names not match")
		assert.Equal(t, reviewerID, submission.Reviewers[0].ID, "Reviewer IDs not match")
		assert.Equal(t, testReviewer.FirstName+" "+testReviewer.LastName, submission.Reviewers[0].FullName, "Reviewer Names not match")

		// clears test env and shuts down the test server
		assert.NoError(t, srv.Shutdown(context.Background()), "failed to shut down server")
		testEnd()
	})
}

// ------------
// Helper Function Tests
// ------------

// test the addSubmission() function in submissions.go
func TestAddSubmission(t *testing.T) {
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
		assert.NoError(t, err, "error reading submission data")
		assert.NoError(t, json.Unmarshal(dataString, submissionData), "error unmarshalling submission data")

		// for each file in the submission, checks that it was added to the filesystem and database properly
		for _, file := range testSub.Files {
			// retrieve the file
			queriedFile := &File{}
			err = gormDb.Model(&File{}).First(queriedFile, file.ID).Error
			assert.NoError(t, err, "Error retrieving file: %v", err)

			// gets the file content from the filesystem
			fileBytes, err := ioutil.ReadFile(queriedFile.Path)
			assert.NoErrorf(t, err, "File read failure after added to filesystem: %v", err)
			queriedFileContent := string(fileBytes)

			// // checks that a data file has been generated for the uploaded file
			// var fileData *FileData
			// fileDataPath := filepath.Join(
			// 	TEST_FILES_DIR,
			// 	fmt.Sprint(queriedSubmission.ID),
			// 	DATA_DIR_NAME,
			// 	queriedSubmission.Name,
			// 	strings.TrimSuffix(queriedFile.Path, filepath.Ext(queriedFile.Path))+".json",
			// )
			// dataString, err := ioutil.ReadFile(fileDataPath)
			// assert.NoError(t, err, "error reading submission data")
			// assert.NoError(t, json.Unmarshal(dataString, fileData), "error unmarshalling submission data")

			// gets data about the file, and tests it for equality against the added file
			_, err = os.Stat(fileDataPath)
			assert.NotErrorIs(t, err, os.ErrNotExist, "Data file not generated during file upload")
			assert.Equal(t, file.Name, queriedFile.Name, "File names do not match")
			assert.Equal(t, file.Path, queriedFile.Path, "File Paths do not match")
			assert.Equal(t, file.SubmissionID, queriedFile.SubmissionID, "File SubmissionIDs do not match")
			assert.Equal(t, file.Base64Value, queriedFileContent, "file content not written to filesystem properly")
			assert.ElementsMatch(t, file.Comments, queriedFile.Comments, "File comments do not match")
		}

		// tests that the metadata is properly formatted
		assert.Equalf(t, submissionData.Abstract, testSub.MetaData.Abstract,
			"submission metadata not added to filesystem properly. Abstracts %s, %s do not match",
			submissionData.Abstract, testSub.MetaData.Abstract)
		assert.ElementsMatch(t, submissionData.Reviews, testSub.MetaData.Reviews, "Submission Reviews do not match")
	}

	// tests that a single valid submission can be added to the db and filesystem properly
	t.Run("Add One Submission", func(t *testing.T) {
		testInit()
		submission := testSubmissions[0]
		authorID, err := registerUser(testAuthors[0])
		assert.NoErrorf(t, err, "Error while registering author: %v", err)
		reviewerID, err := registerUser(testReviewers[0])
		assert.NoErrorf(t, err, "Error while registering reviewer: %v", err)
		submission.Authors = []GlobalUser{{ID: authorID}}
		submission.Reviewers = []GlobalUser{{ID: reviewerID}}
		testAddSubmission(&submission)
		testEnd()
	})

	// tests that multiple submissions can be added in a row properly
	t.Run("Add Multiple Submissions", func(t *testing.T) {
		testInit()
		submissions := testSubmissions[0:2] // list of submissions to add to the db
		authorID, err := registerUser(testAuthors[0])
		assert.NoErrorf(t, err, "Error while registering author: %v", err)
		reviewerID, err := registerUser(testReviewers[0])
		assert.NoErrorf(t, err, "Error while registering reviewer: %v", err)
		for _, submission := range submissions {
			submission.Authors = []GlobalUser{{ID: authorID}}
			submission.Reviewers = []GlobalUser{{ID: reviewerID}}
			testAddSubmission(&submission)
		}
		testEnd()
	})

	// tests that trying to add a nil submission to the db and filesystem will result in an error
	t.Run("Add Nil Submission", func(t *testing.T) {
		testInit()
		_, err := addSubmission(nil)
		assert.Error(t, err, "No error occurred while uploading nil submission")
		testEnd()
	})

	// tests that a submission with some fields as nil will not be added
	t.Run("Add Submission No Authors", func(t *testing.T) {
		testInit()
		testSubmission := testSubmissions[0]

		// nil author array
		testSubmission.Authors = nil
		_, err := addSubmission(&testSubmission)
		assert.Error(t, err, "No error occurred while uploading nil Author array")

		// empty author array
		testSubmission.Authors = []GlobalUser{}
		_, err = addSubmission(&testSubmission)
		assert.Error(t, err, "No error occurred while uploading empty Author array")
		testEnd()
	})
}

// tests the ability of the submissions.go module to add authors to projects
// Test Depends on:
// 	- TestAddSubmission
func TestAddAuthor(t *testing.T) {
	// utility function to add submissions just to the database (not filesystem), without checking for correctness
	testAddSubmission := func(submission *Submission) uint {
		assert.NoError(t, gormDb.Create(submission).Error, "Error creating submission")
		return submission.ID
	}

	// utility function which tests that an author can be added to a valid submission properly
	testAddAuthor := func(submissionID uint, author User) {
		// registers the author as a user and then to the given submission as an author
		authorID, err := registerUser(author)
		assert.NoErrorf(t, err, "Error in author registration: %v", err)
		assert.NoErrorf(t, addAuthors([]GlobalUser{{ID: authorID}}, submissionID), "Error adding the author to the db: %v", err)

		// queries the authors table in the database for author IDs which match that returned from registerUser()
		var authors []GlobalUser
		submission := &Submission{}
		submission.ID = submissionID
		err = gormDb.Model(submission).Select("global_users.id", "global_users.full_name").Association("Authors").Find(&authors)
		assert.NoErrorf(t, err, "Error querying submission Authors: %v", err)

		// checks author details for equality with those which were queried
		authorIDs := []string{}
		authorNames := []string{}
		for _, author := range authors {
			authorIDs = append(authorIDs, author.ID)
			authorNames = append(authorNames, author.FullName)
		}
		assert.Contains(t, authorIDs, authorID, "author id not in the queried array")
		assert.Contains(t, authorNames, author.FirstName+" "+author.LastName, "author id not in the queried array")
	}

	// tests adding one author to a valid project
	t.Run("Adding One Author", func(t *testing.T) {
		testInit()
		// defines test submission and author to use for this test, and uploads the submission
		testSubmission := testSubmissions[0]
		testAuthor := testAuthors[0]
		submissionID := testAddSubmission(&testSubmission)

		// uses the utility function to add the author, and test that it was done properly
		testAddAuthor(submissionID, testAuthor)
		testEnd()
	})

	// attemps to add an author without the correct permissions, if addAuthor succeeds, an error is thrown
	t.Run("Add Invalid Author", func(t *testing.T) {
		testInit()
		testSubmission := testSubmissions[0]
		testAuthor := testAuthors[1] // user without publisher permissions

		// uploads a test submission
		submissionID := testAddSubmission(&testSubmission)

		// registers the author as a user and then to the given submission as an author
		authorID, err := registerUser(testAuthor)
		assert.NoErrorf(t, err, "Error in author registration: %v", err)
		assert.Error(t, addAuthors([]GlobalUser{{ID: authorID}}, submissionID), "Author without publisher permissions registered")

		testEnd()
	})

	// tests that a user must be registered with the db before being and author
	t.Run("Add Non-User Author", func(t *testing.T) {
		testInit()
		testSubmission := testSubmissions[0]
		authorID := "u881jafjka" // non-user fake id
		submissionID := testAddSubmission(&testSubmission)
		assert.Error(t, addAuthors([]GlobalUser{{ID: authorID}}, submissionID), "Non-user added as author")
		testEnd()
	})
}

// tests the ability of the submissions.go module to add reviewers to submissions
// Test Depends on:
// 	- TestAddSubmission
func TestAddReviewer(t *testing.T) {
	// utility function to add submissions just to the database (not filesystem), without checking for correctness
	testAddSubmission := func(submission *Submission) uint {
		assert.NoError(t, gormDb.Create(submission).Error, "Error creating submission")
		return submission.ID
	}

	// utility function which tests that an reviewer can be added to a valid submission properly
	testAddReviewer := func(submissionID uint, reviewer User) {
		// registers the reviewer as a user and then to the given submission as an reviewer
		reviewerID, err := registerUser(reviewer)
		assert.NoErrorf(t, err, "Error in reviewer registration: %v", err)
		assert.NoErrorf(t, addReviewers([]GlobalUser{{ID: reviewerID}}, submissionID), "Error adding the reviewer to the db: %v", err)

		// queries the reviewers table in the database for reviewer IDs which match that returned from registerUser()
		var reviewers []GlobalUser
		submission := &Submission{}
		submission.ID = submissionID
		err = gormDb.Model(submission).Select("global_users.id", "global_users.full_name").Association("Reviewers").Find(&reviewers)
		assert.NoErrorf(t, err, "Error querying submission Reviewers: %v", err)

		// checks reviewer details for equality with those which were queried
		reviewerIDs := []string{}
		reviewerNames := []string{}
		for _, reviewer := range reviewers {
			reviewerIDs = append(reviewerIDs, reviewer.ID)
			reviewerNames = append(reviewerNames, reviewer.FullName)
		}
		assert.Contains(t, reviewerIDs, reviewerID, "reviewer id not in the queried array")
		assert.Contains(t, reviewerNames, reviewer.FirstName+" "+reviewer.LastName, "reviewer id not in the queried array")
	}

	// tests adding one reviewer to a valid project
	t.Run("Adding One Reviewer", func(t *testing.T) {
		testInit()
		// defines test submission and reviewer to use for this test, and uploads the submission
		testSubmission := testSubmissions[0]
		testReviewer := testReviewers[0]
		submissionID := testAddSubmission(&testSubmission)

		// uses the utility function to add the reviewer, and test that it was done properly
		testAddReviewer(submissionID, testReviewer)
		testEnd()
	})

	// attemps to add an reviewer without the correct permissions, if addReviewer succeeds, an error is thrown
	t.Run("Add Invalid Reviewer", func(t *testing.T) {
		testInit()
		testSubmission := testSubmissions[0]
		testReviewer := testReviewers[1] // user without publisher permissions
		submissionID := testAddSubmission(&testSubmission)

		// registers the reviewer as a user and then to the given submission as an reviewer
		reviewerID, err := registerUser(testReviewer)
		assert.NoErrorf(t, err, "Error in reviewer registration: %v", err)
		assert.Error(t, addReviewers([]GlobalUser{{ID: reviewerID}}, submissionID), "Reviewer without reviewer permissions registered")
		testEnd()
	})

	// tests that a user must be registered with the db before being and reviewer
	t.Run("Add Non-User Reviewer", func(t *testing.T) {
		testInit()
		testSubmission := testSubmissions[0]
		reviewerID := "u881jafjka" // non-user fake id
		submissionID := testAddSubmission(&testSubmission)
		assert.Error(t, addReviewers([]GlobalUser{{ID: reviewerID}}, submissionID), "Non-user added as reviewer")
		testEnd()
	})
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
		addTags([]string{testTag}, submissionID)

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
		assert.NoError(t, addTags([]string{testTag}, submissionID), "adding first tag caused an error")
		assert.Error(t, addTags([]string{testTag}, submissionID), "attempting to add duplicate tag does not return an error")

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
		assert.NoError(t, addTags([]string{testTag}, submissionID1), "Error occurred while adding 1st tag")
		assert.NoError(t, addTags([]string{testTag}, submissionID2), "Error occurred while adding 2nd tag")

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
		assert.Error(t, addTags([]string{""}, submissionID), "empty tag was able ot be added")

		testEnd()
	})

	// add tag to non-existant project (foreign key constraint fails in db)
	t.Run("Add Tag Invalid Project", func(t *testing.T) {
		testInit()
		assert.Error(t, addTags([]string{"INVALID_PROJECT"}, 10), "Error not thrown when tag added to a non-existant submission")
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
		authorID, err := registerUser(testAuthor)
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission1.Authors = []GlobalUser{{ID: authorID}}

		nonauthorID, err := registerUser(testNonAuthor) // author of the submission we are not interested in
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission2.Authors = []GlobalUser{{ID: nonauthorID}}

		// adds dummy reviewers
		reviewerID, err := registerUser(testReviewers[0])
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
		reviewerID, err := registerUser(testReviewer)
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission1.Reviewers = []GlobalUser{{ID: reviewerID}}

		nonreviewerID, err := registerUser(testNonReviewer) // author of the submission we are not interested in
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission2.Reviewers = []GlobalUser{{ID: nonreviewerID}}

		// adds dummy authors
		authorID, err := registerUser(testAuthors[0])
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
	// tests the basic case of getting back a valid submission
	t.Run("Single Valid Submission", func(t *testing.T) {
		testSubmission := testSubmissions[0]
		testFile := testFiles[0]
		testAuthor := testAuthors[0]
		testReviewer := testReviewers[0]

		// sets up test environment, and adds a submission with one file to the db and filesystem
		testInit()

		authorID, err := registerUser(testAuthor)
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission.Authors = []GlobalUser{{ID: authorID}}

		reviewerID, err := registerUser(testReviewer)
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission.Reviewers = []GlobalUser{{ID: reviewerID}}

		testSubmission.Files = []File{testFile}
		submissionID, err := addSubmission(&testSubmission)
		assert.NoErrorf(t, err, "Error occurred while adding submission: %v", err)

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

		testEnd()
	})

	// tests trying to get an invalid submission
	t.Run("Invalid Submission", func(t *testing.T) {
		testInit()
		_, err := getSubmission(100)
		assert.Errorf(t, err, "No error was thrown for invalid submission")
		testEnd()
	})
}

// This function tests the getSubmissionAuthors function
//
// This test depends on:
// 	- TestAddSubmission
func TestGetSubmissionAuthors(t *testing.T) {
	// valid metadata file and format
	t.Run("Valid Single Author", func(t *testing.T) {
		testSubmission := testSubmissions[0]
		testAuthor := testAuthors[0]

		// sets up the test environment, and uploads a test submission
		testInit()
		authorID, err := registerUser(testAuthor)
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission.Authors = []GlobalUser{{ID: authorID}}

		submissionID, err := addSubmission(&testSubmission)
		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)

		// tests that the metadata can be read back properly, and that it matches the uploaded submission
		authors, err := getSubmissionAuthors(submissionID)
		assert.NoErrorf(t, err, "Error getting submission authors: %v", err)

		// compares the authors
		assert.Equal(t, authorID, authors[0].ID, "author IDs do not match")
		assert.Equal(t, testAuthor.FirstName+" "+testAuthor.LastName, authors[0].FullName, "Author names do not match")

		testEnd()
	})
}

// This function tests the getSubmissionReviewers function
//
// This test depends on:
// 	- TestAddSubmission
func TestGetSubmissionReviewers(t *testing.T) {
	// valid metadata file and format
	t.Run("Valid Single Reviewer", func(t *testing.T) {
		testSubmission := testSubmissions[0]
		testAuthor := testAuthors[0]
		testReviewer := testReviewers[0]

		// sets up the test environment, and uploads a test submission
		testInit()
		authorID, err := registerUser(testAuthor)
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission.Authors = []GlobalUser{{ID: authorID}}

		reviewerID, err := registerUser(testReviewer)
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission.Reviewers = []GlobalUser{{ID: reviewerID}}

		submissionID, err := addSubmission(&testSubmission)
		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)

		// tests that the metadata can be read back properly, and that it matches the uploaded submission
		reviewers, err := getSubmissionReviewers(submissionID)
		assert.NoErrorf(t, err, "Error getting submission reviewers: %v", err)

		// compares the reviewers
		assert.Equal(t, reviewerID, reviewers[0].ID, "author IDs do not match")
		assert.Equal(t, testReviewer.FirstName+" "+testReviewer.LastName, reviewers[0].FullName, "Author names do not match")

		testEnd()
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
		authorID, err := registerUser(testAuthor)
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission.Authors = []GlobalUser{{ID: authorID}}

		submissionID, err := addSubmission(&testSubmission)
		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)

		// tests that the metadata can be read back properly, and that it matches the uploaded submission
		categories, err := getSubmissionCategories(submissionID)
		assert.NoErrorf(t, err, "Error getting submission tags: %v", err)
		assert.ElementsMatch(t, testSubmission.Categories, categories, "Submission tags do not match")
		testEnd()
	})
}

// This function tests the getSubmissionFiles function
func TestGetSubmissionFiles(t *testing.T) {
	// valid metadata file and format
	t.Run("Single Valid File", func(t *testing.T) {
		testSubmission := testSubmissions[0]
		testFile := testFiles[0]
		testAuthor := testAuthors[0]

		// sets up the test environment, and uploads a test submission
		testInit()
		authorID, err := registerUser(testAuthor)
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission.Authors = []GlobalUser{{ID: authorID}}

		submissionID, err := addSubmission(&testSubmission)
		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)
		fileID, err := addFileTo(&testFile, submissionID)
		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)

		// retrieves the files
		files, err := getSubmissionFiles(submissionID)
		assert.NoErrorf(t, err, "Error retrieving submission files: %v", err)

		// compares the retrieved files with those which were added
		assert.Equal(t, fileID, files[0].ID, "File IDs do not match")
		assert.Equal(t, submissionID, files[0].SubmissionID, "Submission IDs do not match")
		assert.Equal(t, testFile.Name, files[0].Name, "File names do not match")
		assert.Equal(t, testFile.Path, files[0].Path, "File paths do not match")

		testEnd()
	})
}

// This function tests the getSubmissionMetaData function
//
// This test depends on:
// 	- TestAddSubmission
func TestGetSubmissionMetaData(t *testing.T) {
	// valid metadata file and format
	t.Run("Valid Metadata", func(t *testing.T) {
		testSubmission := testSubmissions[0]
		testAuthor := testAuthors[0]

		// sets up the test environment, and uploads a test submission
		testInit()
		authorID, err := registerUser(testAuthor)
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission.Authors = []GlobalUser{{ID: authorID}}

		submissionID, err := addSubmission(&testSubmission)
		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)

		// tests that the metadata can be read back properly, and that it matches the uploaded submission
		submissionData, err := getSubmissionMetaData(submissionID)
		assert.NoErrorf(t, err, "Error getting submission metadata: %v", err)
		assert.Equalf(t, submissionData.Abstract, testSubmission.MetaData.Abstract,
			"submission metadata not added to filesystem properly. Abstracts %s, %s do not match",
			submissionData.Abstract, testSubmission.MetaData.Abstract)
		assert.ElementsMatch(t, submissionData.Reviews, testSubmission.MetaData.Reviews, "Submission Reviews do not match")
		testEnd()
	})

	// Tests that getSubmissionMetaData will throw an error if an incorrect submission ID is passed in
	t.Run("Invalid Submission ID", func(t *testing.T) {
		testInit()
		_, err := getSubmissionMetaData(400)
		assert.Errorf(t, err, "No error was thrown for invalid submission")
		testEnd()
	})
}

// This tests converting from the local submission data format to the supergroup specified format
func TestLocalToGlobal(t *testing.T) {
	// tests valid submission struct
	t.Run("Valid Submission", func(t *testing.T) {
		testInit()

		// adds the submission and a file to the system
		testSubmission := testSubmissions[0]
		testFile := testFiles[0]
		testAuthor := testAuthors[0]
		testReviewer := testReviewers[0]

		// registers authors and reviewers, and adds them to the test submission
		authorID, err := registerUser(testAuthor)
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission.Authors = []GlobalUser{{ID: authorID}}

		reviewerID, err := registerUser(testReviewer)
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		testSubmission.Reviewers = []GlobalUser{{ID: reviewerID}}

		testSubmission.Files = []File{testFile}
		submissionID, err := addSubmission(&testSubmission)
		assert.NoErrorf(t, err, "Error occurred while adding submission: %v", err)

		// gets the supergroup compliant submission
		globalSubmission, err := localToGlobal(submissionID)
		assert.NoErrorf(t, err, "Error occurred while converting submission format: %v", err)

		// compares submission fields
		assert.Equal(t, testSubmission.Name, globalSubmission.Name, "Names do not match")
		assert.Equal(t, testSubmission.License, globalSubmission.MetaData.License,
			"Licenses do not match")
		assert.Equal(t, testAuthor.FirstName+" "+testAuthor.LastName, globalSubmission.MetaData.AuthorNames[0],
			"Authors do not match")
		assert.Equal(t, testSubmission.Categories, globalSubmission.MetaData.Categories,
			"Tags do not match")
		assert.Equal(t, testSubmission.MetaData.Abstract, globalSubmission.MetaData.Abstract,
			"Abstracts do not match")
		// compares files
		assert.Equal(t, testFile.Name, globalSubmission.Files[0].Name, "File names do not match")
		assert.Equal(t, testFile.Base64Value, globalSubmission.Files[0].Base64Value, "File content does not match")

		testEnd()
	})
}
