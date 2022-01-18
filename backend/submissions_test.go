// ===========================
// submissions_test.go
// Authors: 190010425
// Created: November 18, 2021
//
// This file takes care of testing
// submissions.go
// ===========================

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// data to use in the tests
var testSubmissions []*Submission = []*Submission{
	{Id: -1, Name: "testSubmission1", Reviewers: []string{},
		Authors: []string{}, FilePaths: []string{"testFile1.txt"}, MetaData: testSubmissionMetaData[0]},
	{Id: -1, Name: "testSubmission2", Reviewers: []string{},
		Authors: []string{}, FilePaths: []string{"testFile2.txt"}, MetaData: testSubmissionMetaData[0]},
}
var testSubmissionMetaData = []*CodeSubmissionData{
	{Abstract: "test abstract, this means nothing", Reviews: nil}, // TODO: add comments here
}
var testAuthors []*Credentials = []*Credentials{
	{Email: "test@test.com", Pw: "123456aB$", Fname: "test",
		Lname: "test", PhoneNumber: "0574349206", Usertype: USERTYPE_PUBLISHER},
	{Email: "john.doe@test.com", Pw: "dlbjDs2!", Fname: "John",
		Lname: "Doe", Organization: "TestOrg", Usertype: USERTYPE_USER},
	{Email: "jane.doe@test.net", Pw: "dlbjDs2!", Fname: "Jane",
		Lname: "Doe", Usertype: USERTYPE_REVIEWER},
	{Email: "adam.doe@test.net", Pw: "dlbjDs2!", Fname: "Adam",
		Lname: "Doe", Usertype: USERTYPE_REVIEWER_PUBLISHER},
}
var testReviewers []*Credentials = []*Credentials{
	{Email: "dave@test.com", Pw: "123456aB$", Fname: "dave",
		Lname: "smith", PhoneNumber: "0574349206", Usertype: USERTYPE_REVIEWER},
	{Email: "Geoff@test.com", Pw: "dlbjDs2!", Fname: "Geoff",
		Lname: "Williams", Organization: "TestOrg", Usertype: USERTYPE_USER},
	{Email: "jane.doe@test.net", Pw: "dlbjDs2!", Fname: "Jane",
		Lname: "Doe", Usertype: USERTYPE_PUBLISHER},
	{Email: "adam.doe@test.net", Pw: "dlbjDs2!", Fname: "Adam",
		Lname: "Doe", Usertype: USERTYPE_REVIEWER_PUBLISHER},
}

// test the addSubmission() function in submissions.go
func TestAddSubmission(t *testing.T) {
	// Utility function to be re-used for testing adding submissions to the db
	testAddSubmission := func(testSubmission *Submission) {
		submissionId, err := addSubmission(testSubmission)
		assert.NoErrorf(t, err, "%v", err)
		assert.Greaterf(t, submissionId, 0, "Invalid Submission ID returned: %d", submissionId)

		// checks manually that the submission was added correctly
		var submissionName string
		authors := []string{}
		reviewers := []string{}

		// builds SQL Queries for retrieving the added values
		querySubmissionName := fmt.Sprintf(SELECT_ROW, getDbTag(&Submission{}, "Name"),
			TABLE_SUBMISSIONS, getDbTag(&Submission{}, "Id"))
		queryAuthors := fmt.Sprintf(SELECT_ROW, "userId",
			TABLE_AUTHORS, "submissionId")
		queryReviewers := fmt.Sprintf(SELECT_ROW, "userId",
			TABLE_REVIEWERS, "submissionId")

		// tests that the submission name was added correctly by querying the sql database
		row := db.QueryRow(querySubmissionName, submissionId)
		assert.NoErrorf(t, row.Err(), "%v", row.Err())
		assert.NoErrorf(t, row.Scan(&submissionName), "Error querying submission: %v", err)
		assert.Equalf(t, testSubmission.Name, submissionName, "Submission name mismatch. %s vs %s", testSubmission.Name, submissionName)

		// tests that the authors were added correctly by querying the sql db and comparing to the known values
		var author string
		rows, err := db.Query(queryAuthors, submissionId)
		assert.NoErrorf(t, err, "Error querying submission Authors: %v", err)
		for rows.Next() {
			rows.Scan(&author)
			authors = append(authors, author)
		}
		assert.ElementsMatch(t, testSubmission.Authors, authors, "authors arrays do not match")

		// tests that the reviewers were added correctly by querying the sql db and comparing to the known values
		var reviewer string
		rows, err = db.Query(queryReviewers, submissionId)
		assert.NoErrorf(t, err, "Error querying submission Authors: %v", err)
		for rows.Next() {
			rows.Scan(&author)
			reviewers = append(reviewers, reviewer)
		}
		assert.ElementsMatch(t, testSubmission.Reviewers, reviewers, "reviewer arrays do not match")

		// checks that the filesystem has a proper corresponding entry and metadata file
		submissionData := &CodeSubmissionData{}
		submissionDirPath := filepath.Join(TEST_FILES_DIR, fmt.Sprint(submissionId))
		fileDataPath := filepath.Join(submissionDirPath, DATA_DIR_NAME, submissionName+".json")
		dataString, err := ioutil.ReadFile(fileDataPath)
		assert.NoError(t, err, "error reading submission data")
		assert.NoError(t, json.Unmarshal(dataString, submissionData), "error unmarshalling submission data")

		// tests that the metadata is properly formatted
		assert.Equalf(t, submissionData.Abstract, testSubmission.MetaData.Abstract,
			"submission metadata not added to filesystem properly. Abstracts %s, %s do not match",
			submissionData.Abstract, testSubmission.MetaData.Abstract)
		assert.ElementsMatch(t, submissionData.Reviews, testSubmission.MetaData.Reviews, "Submission Reviews do not match")
	}

	// tests that a single valid submission can be added to the db and filesystem properly
	t.Run("Add One Submission", func(t *testing.T) {
		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
		submission := testSubmissions[0]
		testAddSubmission(submission)
		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
	})

	// tests that multiple submissions can be added in a row properly
	t.Run("Add Multiple Submissions", func(t *testing.T) {
		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
		submissions := testSubmissions[0:2] // list of submissions to add to the db
		for _, submission := range submissions {
			testAddSubmission(submission)
		}
		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
	})

	// tests that trying to add a nil submission to the db and filesystem will result in an error
	t.Run("Add Nil Submission", func(t *testing.T) {
		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
		_, err := addSubmission(nil)
		assert.Error(t, err, "No error occurred while uploading nil submission")
		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
	})
}

// tests the ability of the submissions.go module to add authors to projects
// Test Depends on:
// 	- TestAddSubmission
func TestAddAuthor(t *testing.T) {
	// utility function which tests that an author can be added to a valid submission properly
	testAddAuthor := func(submissionId int, author *Credentials) {
		// registers the author as a user and then to the given submission as an author
		authorId, err := registerUser(author)
		assert.NoErrorf(t, err, "Error in author registration: %v", err)
		assert.NoErrorf(t, addAuthor(authorId, submissionId), "Error adding the author to the db: %v", err)

		// queries the authors table in the database for author IDs which match that returned from registerUser()
		var queriedSubmissionId int
		var queriedAuthorId string 
		queryAuthor := fmt.Sprintf(SELECT_ROW, "*", TABLE_AUTHORS, "userId")
		row := db.QueryRow(queryAuthor, authorId)
		assert.NoErrorf(t, row.Scan(&queriedSubmissionId, &queriedAuthorId), "error while querying db for authors: %v", row.Err())

		// checks submission Id and author Id for equality with those which were queried
		assert.Equalf(t, submissionId, queriedSubmissionId, 
			"Author added to the wrong submission: Wanted: %d Got: %d", submissionId, queriedSubmissionId)
		assert.Equalf(t, authorId, queriedAuthorId,
			"Author Ids do not match: Added: %s Gotten Back: %s", authorId, queriedAuthorId)
	}

	// tests adding one author to a valid project
	t.Run("Adding One Author", func(t *testing.T) {
		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
		// defines test submission and author to use for this test, and uploads the submission
		testSubmission := testSubmissions[0]
		testAuthor := testAuthors[0]
		submissionId, err := addSubmission(testSubmission)
		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)

		// uses the utility function to add the author, and test that it was done properly
		testAddAuthor(submissionId, testAuthor)
		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
	})

	// attemps to add an author without the correct permissions, if addAuthor succeeds, an error is thrown
	t.Run("Add Invalid Author", func(t *testing.T) {
		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
		testSubmission := testSubmissions[0]
		testAuthor := testAuthors[1] // user without publisher permissions

		// uploads the valid test submission
		submissionId, err := addSubmission(testSubmission)
		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)

		// registers the author as a user and then to the given submission as an author
		authorId, err := registerUser(testAuthor)
		assert.NoErrorf(t, err, "Error in author registration: %v", err)
		assert.Error(t, addAuthor(authorId, submissionId), "Author without publisher permissions registered")

		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
	})

	// tests that a user must be registered with the db before being and author
	t.Run("Add Non-User Author", func(t *testing.T) {
		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
		testSubmission := testSubmissions[0]
		authorId := "u881jafjka" // non-user fake id
		submissionId, err := addSubmission(testSubmission)
		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)
		assert.Error(t, addAuthor(authorId, submissionId), "Non-user added as author")
		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
	})
}

// tests the ability of the submissions.go module to add reviewers to projects
// Test Depends on:
// 	- TestAddSubmission
func TestAddReviewer(t *testing.T) {
	// utility function which tests that a reviewer can be added to a valid submission properly
	testAddReviewer := func(submissionId int, reviewer *Credentials) {
		// registers the reviewer as a user and then to the given submission as an reviewer
		reviewerId, err := registerUser(reviewer)
		assert.NoErrorf(t, err, "Error in reviewer registration: %v", err)
		assert.NoErrorf(t, addReviewer(reviewerId, submissionId), "Error adding the reviewer to the db: %v", err)

		// queries the reviewers table in the database for reviewer IDs which match that returned from registerUser()
		var queriedSubmissionId int
		var queriedReviewerId string 
		queryReviewer := fmt.Sprintf(SELECT_ROW, "*", TABLE_REVIEWERS, "userId")
		row := db.QueryRow(queryReviewer, reviewerId)
		assert.NoErrorf(t, row.Scan(&queriedSubmissionId, &queriedReviewerId), "error while querying db for reviewers: %v", row.Err())

		// checks submission Id and reviewer Id for equality with those which were queried
		assert.Equalf(t, submissionId, queriedSubmissionId, 
			"Reviewer added to the wrong submission: Wanted: %d Got: %d", submissionId, queriedSubmissionId)
		assert.Equalf(t, reviewerId, queriedReviewerId,
			"Reviewer Ids do not match: Added: %s Gotten Back: %s", reviewerId, queriedReviewerId)
	}

	// tests adding one reviewer to a valid project
	t.Run("Adding One Reviewer", func(t *testing.T) {
		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
		// defines test submission and reviewer to use for this test, and uploads the submission
		testSubmission := testSubmissions[0]
		testReviewer := testReviewers[0]
		submissionId, err := addSubmission(testSubmission)
		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)

		// uses the utility function to add the reviewer, and test that it was done properly
		testAddReviewer(submissionId, testReviewer)
		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
	})

	// attemps to add an reviewer without the correct permissions, if addReviewer succeeds, an error is thrown
	t.Run("Add Invalid Reviewer", func(t *testing.T) {
		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
		testSubmission := testSubmissions[0]
		testReviewer := testReviewers[1] // user without publisher permissions

		// uploads the valid test submission
		submissionId, err := addSubmission(testSubmission)
		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)

		// registers the reviewer as a user and then to the given submission as an reviewer
		reviewerId, err := registerUser(testReviewer)
		assert.NoErrorf(t, err, "Error in reviewer registration: %v", err)
		assert.Error(t, addReviewer(reviewerId, submissionId), "Reviewer without reviewer permissions registered")

		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
	})

	// tests that a user must be registered with the db before being and reviewer
	t.Run("Add Non-User Reviewer", func(t *testing.T) {
		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")

		testSubmission := testSubmissions[0]
		reviewerId := "u881jafjka" // non-user fake id
		submissionId, err := addSubmission(testSubmission)
		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)
		assert.Error(t, addReviewer(reviewerId, submissionId), "Non-user added as reviewer")
		
		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
	})
}

// This function tests the getSubmissionMetaData function
//
// This test depends on:
// 	- TestAddSubmission
func TestGetSubmissionMetaData(t *testing.T) {
	// valid metadata file and format
	t.Run("Valid Metadata", func(t *testing.T) {
		// sets up the test environment, and uploads a test submission
		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
		testSubmission := testSubmissions[0]
		submissionId, err := addSubmission(testSubmission)
		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)

		// tests that the metadata can be read back properly, and that it matches the uploaded submission
		submissionData, err := getSubmissionMetaData(submissionId)
		assert.NoErrorf(t, err, "Error getting submission metadata: %v", err)
		assert.Equalf(t, submissionData.Abstract, testSubmission.MetaData.Abstract, 
			"submission metadata not added to filesystem properly. Abstracts %s, %s do not match",
			submissionData.Abstract, testSubmission.MetaData.Abstract)
		assert.ElementsMatch(t, submissionData.Reviews, testSubmission.MetaData.Reviews, "Submission Reviews do not match")

		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
	})

	// Tests that getSubmissionMetaData will throw an error if an incorrect submission ID is passed in
	t.Run("Invalid Submission Id", func(t *testing.T) {
		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
		_, err := getSubmissionMetaData(400)
		assert.Errorf(t, err, "No error was thrown for invalid submission")
		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
	})
}

// Tests the ability of the getAllSubmissions() function to get all submission ids and names from the db
// at once
//
// Test Depends On:
// 	- TestAddSubmission
// 	- TestGetUserSubmissions
func TestGetAllSubmissions(t *testing.T) {
	// tests that multiple valid submissions can be uploaded, then retrieved from the database
	t.Run("Get Multiple Valid submissions", func(t *testing.T) {
		// Set up server and test environment
		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
		srv := setupCORSsrv()
		go srv.ListenAndServe()

		// adds all of the submissions and stores their ids and names
		sentSubmissions := make(map[int]string) // variable to hold the id: submission name mappings which are sent to the db
		for _, sub := range testSubmissions[0:2] {
			submissionId, err := addSubmission(sub)
			assert.NoErrorf(t, err, "Error adding submission %s: %v", sub.Name, err)
			sentSubmissions[submissionId] = sub.Name
		}

		// builds and sends and http request to get the names and Ids of all submissions
		req, err := http.NewRequest("GET", fmt.Sprintf("%s:%s%s", TEST_URL, TEST_SERVER_PORT, ENDPOINT_ALL_SUBMISSIONS), nil)
		resp, err := sendSecureRequest(req, TEAM_ID)
		assert.NoErrorf(t, err, "Error occurred while sending get request to the Go server: %v", err)
		defer resp.Body.Close()

		// checks the returned list of submissions for equality with the sent list
		returnedSubmissions := make(map[int]string)
		json.NewDecoder(resp.Body).Decode(&returnedSubmissions)
		for k, v := range returnedSubmissions {
			assert.Equalf(t, v,  sentSubmissions[k], 
				"Submissions of ids: %d do not have matching names. Given: %s, Returned: %s ", k, sentSubmissions[k], v)
		}

		// clears test env and shuts down the test server
		assert.NoError(t, srv.Shutdown(context.Background()), "failed to shut down server")
		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
	})
}

// test for basic functionality. Adds 2 submissions to the db with different authors, then queries them and tests for equality
// Test Depends On:
// 	- TestAddSubmission
// 	- TestAddAuthors
func TestGetUserSubmissions(t *testing.T) {
	// adds two submissions each with different authors to the db and then queries one author's submissions
	t.Run("Get Single Submission from an Author", func(t *testing.T) {
		testSubmission1 := testSubmissions[0] // test submission to return on getUserSubmissions()
		testSubmission2 := testSubmissions[1] // test submission to not return on getUserSubmissions()
		testAuthor := testAuthors[0]          // test author of the submission being queried
		testNonAuthor := testAuthors[3]       // test author of submission not being queried

		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")

		// adds two test users to the db
		authorId, err := registerUser(testAuthor)
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
		nonAuthorId, err := registerUser(testNonAuthor) // author of the submission we are not interested in
		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
	
		// adds two test submissions to the db
		testSubmission1.Id, err = addSubmission(testSubmission1)
		assert.NoErrorf(t, err, "Error occurred while adding submission1: %v", err)
		testSubmission2.Id, err = addSubmission(testSubmission2)
		assert.NoErrorf(t, err, "Error occurred while adding submission2: %v", err)
	
		// adds authors to the test submissions
		assert.NoError(t, addAuthor(authorId, testSubmission1.Id), "Failed to add author")
		assert.NoError(t, addAuthor(nonAuthorId, testSubmission2.Id), "Failed to add author")
	
		// queries all of testAuthor's submissions
		submissions, err := getUserSubmissions(authorId)
		assert.NoErrorf(t, err, "Error getting user submissions: %v", err)
	
		// tests for equality of submission Id and that testSubmission2.Id is not in the map
		_, ok := submissions[testSubmission2.Id]
		assert.False(t, ok, "Returned submission where the test author is not an author")
		assert.Equalf(t, submissions[testSubmission1.Id], testSubmission1.Name,
			"Returned incorrect submission name: %s", submissions[testSubmission1.Id])
		
		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
	})
}

// Tests the ability of the CodeFiles module to get a submission from the db
//
// Test Depends On:
// 	- TestCreateSubmissions()
// 	- TestAddFiles()
// 	- TestAddReviewers()
// 	- TestAddAuthors()
func TestGetSubmission(t *testing.T) {
	// tests that a single valid submission with one reviewer and one author can be retrieved
	t.Run("Get Valid Submission", func(t *testing.T) {
		testFile := testFiles[0]             // defines the file to use for the test here so that it can be easily changed
		testSubmission := testSubmissions[0] // defines the submission to use for the test here so that it can be easily changed
		testAuthor := testAuthors[0]         // defines the author of the submission
		testReviewer := testReviewers[0]     // defines the reviewer of the submission	

		// Set up server and test environment
		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
		srv := setupCORSsrv()
		go srv.ListenAndServe()

		// uploads the test submission and adds a file to it
		submissionId, err := addSubmission(testSubmission)
		assert.NoErrorf(t, err, "Error adding submission %v", err)
		_, err = addFileTo(testFile, submissionId)
		assert.NoErrorf(t, err, "Error adding file to the submission %v", err)

		// registers an author and reviewer as users, then adds them to the database
		authorId, err := registerUser(testAuthor)
		assert.NoErrorf(t, err, "Error registering author in the db: %v", err)
		assert.NoErrorf(t, addAuthor(authorId, submissionId), "Error adding the author to the db: %v", err)
		reviewerId, err := registerUser(testReviewer)
		assert.NoErrorf(t, err, "Error registering reviewer in the db: %v", err)
		assert.NoErrorf(t, addReviewer(reviewerId, submissionId), "Error adding the reviewer to the db: %v", err)

		// creates a request to send to the test server
		urlString := fmt.Sprintf("%s:%s%s?%s=%d", TEST_URL, TEST_SERVER_PORT, 
			ENDPOINT_SUBMISSION, getJsonTag(&Submission{}, "Id"), submissionId)
		req, _ := http.NewRequest("GET", urlString, nil)
		resp, err := sendSecureRequest(req, TEAM_ID)
		assert.NoErrorf(t, err, "Error while sending Get request: %v", err)
		defer resp.Body.Close()
		assert.Equalf(t, resp.StatusCode, http.StatusOK, "Non-OK status returned from GET request: %d", resp.StatusCode)

		// decodes the json response into a Submission struct
		submission := &Submission{}
		assert.NoErrorf(t, json.NewDecoder(resp.Body).Decode(submission), "Error while decoding server response: %v", err)

		// tests that the returned submission matches the passed in data
		assert.Equalf(t, testSubmission.Id, submission.Id, 
			"Submission IDs do not match. Given: %d != Returned: %d", testSubmission.Id, submission.Id)
		assert.Equalf(t, testSubmission.Name, submission.Name,
			"Submission Names do not match. Given: %s != Returned: %s", testSubmission.Name, submission.Name)
		assert.ElementsMatch(t, []string{authorId}, submission.Authors, "Authors do not match")
		assert.ElementsMatch(t, []string{reviewerId}, submission.Reviewers, "Reviewers do not match")
		assert.ElementsMatch(t, testSubmission.FilePaths, submission.FilePaths, "Submission file path lists do not match.")

		// clears test env and shuts down the test server
		assert.NoError(t, srv.Shutdown(context.Background()), "failed to shut down server")
		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
	})
}
