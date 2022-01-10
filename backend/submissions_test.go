// ===========================
// submissions_test.go
// Authors: 190010425
// Created: November 18, 2021
//
// This file takes care of
// ===========================

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"reflect"
	"testing"
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

// Utility function to be re-used for testing adding submissions to the db
func testAddSubmission(testSubmission *Submission) error {
	submissionId, err := addSubmission(testSubmission)

	// simple error cases
	if err != nil {
		return err
	} else if submissionId < 0 {
		return errors.New(fmt.Sprintf("Invalid Submission ID returned: %d", submissionId))
	}

	// checks manually that the submission was added correctly
	var submissionName string
	authors := []string{}
	reviewers := []string{}
	// builds SQL Queries for testing the added values
	querySubmissionName := fmt.Sprintf(SELECT_ROW, getDbTag(&Submission{}, "Name"),
		TABLE_SUBMISSIONS, getDbTag(&Submission{}, "Id"))
	queryAuthors := fmt.Sprintf(SELECT_ROW, "userId",
		TABLE_AUTHORS, "submissionId")
	queryReviewers := fmt.Sprintf(SELECT_ROW, "userId",
		TABLE_REVIEWERS, "submissionId")

	// tests that the submission name was added correctly
	row := db.QueryRow(querySubmissionName, submissionId)
	if row.Err() != nil {
		return errors.New(fmt.Sprintf("Error in submission name query: %v", row.Err()))
	} else if err = row.Scan(&submissionName); err != nil {
		return err
	} else if testSubmission.Name != submissionName {
		return errors.New(
			fmt.Sprintf("Submission name mismatch. %s vs %s",
				testSubmission.Name, submissionName))
	}

	// tests that the authors were added correctly
	rows, err := db.Query(queryAuthors, submissionId)
	if err != nil {
		return errors.New(fmt.Sprintf("Error querying submission Authors: %v", err))
	}
	var author string
	for rows.Next() {
		rows.Scan(&author)
		authors = append(authors, author)
	}
	if !(reflect.DeepEqual(testSubmission.Authors, authors)) {
		return errors.New("authors arrays do not match")
	}

	// tests that the reviewers were added correctly
	rows, err = db.Query(queryReviewers, submissionId)
	if err != nil {
		return errors.New(fmt.Sprintf("error querying submission Reviewers: %v", err))
	}
	var reviewer string
	for rows.Next() {
		rows.Scan(&reviewer)
		reviewers = append(reviewers, reviewer)
	}
	if !(reflect.DeepEqual(testSubmission.Reviewers, reviewers)) {
		return errors.New("reviewers arrays do not match")
	}

	// checks that the filesystem has a proper corresponding entry and metadata file
	submissionDirPath := filepath.Join(TEST_FILES_DIR, fmt.Sprint(submissionId))
	fileDataPath := filepath.Join(submissionDirPath, DATA_DIR_NAME, submissionName+".json")
	dataString, err := ioutil.ReadFile(fileDataPath)
	if err != nil {
		return err
	}
	// marshalls the string of data into a struct
	submissionData := &CodeSubmissionData{}
	if err := json.Unmarshal(dataString, submissionData); err != nil {
		return err
	}
	// tests that the metadata is properly formatted
	if submissionData.Abstract != testSubmission.MetaData.Abstract {
		return errors.New(fmt.Sprintf(
			"submission metadata not added to filesystem properly. Abstracts %s, %s do not match",
			submissionData.Abstract, testSubmission.MetaData.Abstract))
	} else if !(reflect.DeepEqual(submissionData.Reviews, testSubmission.MetaData.Reviews)) {
		return errors.New("Submission Reviews do not match")
	}
	return nil
}

// tests that a single valid submission can be added to the db and filesystem properly
func TestAddOneSubmission(t *testing.T) {
	submission := testSubmissions[0]

	// initializes the test environment, returning an error if any occurs
	if err := initTestEnvironment(); err != nil {
		t.Errorf("Error in db init: %v", err)
	}
	// adds the submission and tests that it was added properly
	if err := testAddSubmission(submission); err != nil {
		t.Errorf("%v", err) // error already formatted here
	}
	// tears down the test environment
	if err := clearTestEnvironment(); err != nil {
		t.Errorf("Error in db teardown: %v", err)
	}
}

// tests that multiple submissions can be added in a row properly
func TestAddMultipleSubmissions(t *testing.T) {
	submissions := testSubmissions[0:2] // list of submissions to add to the db

	// initializes the test environment, returning an error if any occurs
	if err := initTestEnvironment(); err != nil {
		t.Errorf("Error in db init: %v", err)
	}
	// adds a range all submissions in the submissions slice
	for _, submission := range submissions {
		// adds the submission and tests that it was added properly
		if err := testAddSubmission(submission); err != nil {
			t.Errorf("%v", err)
		}
	}
	// tears down the test environment
	if err := clearTestEnvironment(); err != nil {
		t.Errorf("Error in db teardown: %v", err)
	}
}

// tests that trying to add a nil submission to the db and filesystem will return an error
func TestAddNilSubmission(t *testing.T) {
	// initializes the test environment
	if err := initTestEnvironment(); err != nil {
		t.Errorf("Error in db init: %v", err)
	}
	// tries to add a nil submission
	if _, err := addSubmission(nil); err == nil {
		t.Error("Nil submission added to the db without error")
	}
	// tears down the test environment
	if err := clearTestEnvironment(); err != nil {
		t.Errorf("Error in db teardown: %v", err)
	}
}

// utility function which tests that an author can be added to a valid submission properly
// this test depends on the add submissions tests
func testAddAuthor(submissionId int, author *Credentials) error {
	// declares test variables
	var queriedSubmissionId int // gotten from db after adding author
	var queriedAuthorId string  // gotten from db after adding author

	authorId, err := registerUser(author)
	if err != nil {
		return errors.New(fmt.Sprintf("Error in author registration: %v", err))
	}

	// adds the author to the database
	if err = addAuthor(authorId, submissionId); err != nil {
		return errors.New(fmt.Sprintf("Error adding the author to the db: %v", err))
	}

	// checks the author ID and submission ID for matches
	queryAuthor := fmt.Sprintf(SELECT_ROW, "*", TABLE_AUTHORS, "userId")
	row := db.QueryRow(queryAuthor, authorId)
	if err := row.Scan(&queriedSubmissionId, &queriedAuthorId); err != nil {
		return errors.New(
			fmt.Sprintf("error while querying db for authors: %v", row.Err()))
	}

	// checks data returned from the database
	if submissionId != queriedSubmissionId {
		return errors.New(
			fmt.Sprintf("Author added to the wrong submission: Wanted: %d Got: %d",
				submissionId, queriedSubmissionId))
	} else if authorId != queriedAuthorId {
		return errors.New(
			fmt.Sprintf("Author Ids do not match: Added: %s Gotten Back: %s",
				authorId, queriedAuthorId))
	}
	return nil
}

func TestAddOneAuthor(t *testing.T) {
	testSubmission := testSubmissions[0]
	testAuthor := testAuthors[0]

	// initializes the test environment
	if err := initTestEnvironment(); err != nil {
		t.Errorf("Error in db init: %v", err)
	}
	// adds a valid submission and user to the db and filesystem so that an author can be added
	submissionId, err := addSubmission(testSubmission)
	if err != nil {
		t.Errorf("Error occurred while adding test submission: %v", err)
	}
	// adds the author to the db and filesystem
	if err := testAddAuthor(submissionId, testAuthor); err != nil {
		t.Errorf("Error occurred while adding test author: %v", err)
	}
	// tears down the test environment
	if err := clearTestEnvironment(); err != nil {
		t.Errorf("Error in db teardown: %v", err)
	}
}

// attemps to add an author without the correct permissions, if addAuthor succeeds, an error is thrown
func TestAddInvalidAuthor(t *testing.T) {
	testSubmission := testSubmissions[0]
	testAuthor := testAuthors[1] // user without publisher permissions

	// initializes the test environment
	if err := initTestEnvironment(); err != nil {
		t.Errorf("Error in db init: %v", err)
	}
	// adds a valid submission and user to the db and filesystem so that an author can be added
	submissionId, err := addSubmission(testSubmission)
	if err != nil {
		t.Errorf("Error adding test submission: %v", err)
	}
	// if adding the author is successful, throw an error
	if err = testAddAuthor(submissionId, testAuthor); err == nil {
		t.Error("Incorrect permissions added to authors table.")
	}
	// tears down the test environment
	if err := clearTestEnvironment(); err != nil {
		t.Errorf("Error in db teardown: %v", err)
	}
}

// tests that a user must be registered with the db before being and author
func TestAddNonUserAuthor(t *testing.T) {
	testSubmission := testSubmissions[0]
	authorId := "u881jafjka" // non-user fake id

	// initializes the test environment, returning an error if any occurs
	if err := initTestEnvironment(); err != nil {
		t.Errorf("Error in test environment init: %v", err)
	}
	// adds a valid submission to the db and filesystem
	submissionId, err := addSubmission(testSubmission)
	if err != nil {
		t.Errorf("Error while adding test submission: %v", err)
	}
	// if adding the author is successful, throw an error
	if err = addAuthor(authorId, submissionId); err == nil {
		t.Error("Added unregistered user id as author.")
	}
	// clears the test environment
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("Error on db teardown: %v", err)
	}
}

// utility function which tests that a reviewer can be added to a valid submission properly
// this test depends on the add submissions tests
func testAddReviewer(submissionId int, reviewer *Credentials) error {
	var queriedSubmissionId int  // gotten from db after adding reviewer
	var queriedReviewerId string // gotten from db after adding reviewer

	reviewerId, err := registerUser(reviewer)
	if err != nil {
		return errors.New(fmt.Sprintf("Error in reviewer registration: %v", err))
	}

	// adds the reviewer to the database
	if err = addReviewer(reviewerId, submissionId); err != nil {
		return errors.New(fmt.Sprintf("Error adding the reviewer to the db: %v", err))
	}

	// checks the reviewer ID and submission ID for matches
	queryReviewer := fmt.Sprintf(SELECT_ROW, "*", TABLE_REVIEWERS, "userId")
	row := db.QueryRow(queryReviewer, reviewerId)
	if err := row.Scan(&queriedSubmissionId, &queriedReviewerId); err != nil {
		return errors.New(
			fmt.Sprintf("error while querying db for reviewers: %v", row.Err()))
	}

	// checks data returned from the database
	if submissionId != queriedSubmissionId {
		return errors.New(
			fmt.Sprintf("Reviewer added to the wrong submission: Wanted: %d Got: %d",
				submissionId, queriedSubmissionId))
	} else if reviewerId != queriedReviewerId {
		return errors.New(
			fmt.Sprintf("Reviewer Ids do not match: Added: %s Gotten Back: %s",
				reviewerId, queriedReviewerId))
	}
	return nil
}

// tests that a single valid reviewer can be added to the database properly
func TestAddOneReviewer(t *testing.T) {
	testSubmission := testSubmissions[0]
	testReviewer := testReviewers[0]

	// initializes the test environment
	if err := initTestEnvironment(); err != nil {
		t.Errorf("Error in db init: %v", err)
	}
	// adds a valid submission and user to the db and filesystem so that an author can be added
	submissionId, err := addSubmission(testSubmission)
	if err != nil {
		t.Errorf("Error occurred while adding test submission: %v", err)
	}
	// adds the reviewer to the db and filesystem
	if err := testAddReviewer(submissionId, testReviewer); err != nil {
		t.Errorf("Error occurred while adding test reviewer: %v", err)
	}
	// tears down the test environment
	if err := clearTestEnvironment(); err != nil {
		t.Errorf("Error in db teardown: %v", err)
	}
}

// attemps to add a reviewere without the correct permissions, if addReviewer succeeds, an error is thrown
func TestAddInvalidReviewer(t *testing.T) {
	testSubmission := testSubmissions[0]
	testReviewer := testReviewers[1] // reviewer without reviewer permissions

	// initializes the test environment
	if err := initTestEnvironment(); err != nil {
		t.Errorf("Error in db init: %v", err)
	}
	// adds a valid submission and user to the db and filesystem so that an reviewer can be added
	submissionId, err := addSubmission(testSubmission)
	if err != nil {
		t.Errorf("Error adding test submission: %v", err)
	}
	// if adding the reviewer is successful, throw an error
	if err = testAddReviewer(submissionId, testReviewer); err == nil {
		t.Error("Incorrect permissions added to reviewers table.")
	}
	// tears down the test environment
	if err := clearTestEnvironment(); err != nil {
		t.Errorf("Error in db teardown: %v", err)
	}
}

// tests that a user must be registered with the db before being and author
func TestAddNonUserReviewer(t *testing.T) {
	testSubmission := testSubmissions[0]
	reviewerId := "u881jafjka" // non-user fake id

	// initializes the test environment, returning an error if any occurs
	if err := initTestEnvironment(); err != nil {
		t.Errorf("Error in test environment init: %v", err)
	}
	// adds a valid submission to the db and filesystem
	submissionId, err := addSubmission(testSubmission)
	if err != nil {
		t.Errorf("Error while adding test submission: %v", err)
	}
	// if adding the author is successful, throw an error
	if err = addReviewer(reviewerId, submissionId); err == nil {
		t.Error("Added unregistered user id as reviewer.")
	}
	// clears the test environment
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("Error on db teardown: %v", err)
	}
}

// This function tests the getSubmissionMetaData function
//
// This test depends on:
// 	- addSubmission()
func TestGetSubmissionMetaData(t *testing.T) {
	testSubmission := testSubmissions[0]

	// initializes the test environment, returning an error if any occurs
	if err := initTestEnvironment(); err != nil {
		t.Errorf("Error in test environment init: %v", err)
	}
	// adds the test submission to the db
	submissionId, err := addSubmission(testSubmission)
	if err != nil {
		t.Errorf("Error adding submission to the db and filesystem: %v", err)
	}
	// tests that the metadata can be read back properly
	submissionData, err := getSubmissionMetaData(submissionId)
	if err != nil {
		t.Errorf("Error getting submission metadata: %v", err)
	}
	// tests for equality of the added metadata with that which was retrieved
	if submissionData.Abstract != testSubmission.MetaData.Abstract {
		t.Errorf("submission metadata not added to filesystem properly. Abstracts %s, %s do not match",
			submissionData.Abstract, testSubmission.MetaData.Abstract)
	} else if !(reflect.DeepEqual(submissionData.Reviews, testSubmission.MetaData.Reviews)) {
		t.Error("Submission Reviews do not match")
	}
	// clears the test environment
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("Error on db teardown: %v", err)
	}
}

// Tests that getSubmissionMetaData will throw an error if an incorrect submission ID is passed in
func TestGetInvalidSubmissionMetaData(t *testing.T) {
	// initializes the test environment, returning an error if any occurs
	if err := initTestEnvironment(); err != nil {
		t.Errorf("Error in test environment init: %v", err)
	}
	// tests that an error is thrown if a non-existant submission ID is passed to getSubmissionMetaData
	_, err := getSubmissionMetaData(400)
	if err == nil {
		t.Errorf("No error was thrown for invalid submission")
	}
	// clears the test environment
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("Error on db teardown: %v", err)
	}
}

// Tests the ability of the getAllSubmissions() function to get all submissions from the db at once
//
// Test Depends On:
// 	- TestCreateSubmissions()
func TestGetAllSubmissions(t *testing.T) {
	var err error

	// Set up server to listen with the getFile() function.
	srv := setupCORSsrv()

	// Start server.
	go srv.ListenAndServe()

	var submissionId int                    // variable to temporarily store submission ids as they are added to the db
	sentSubmissions := make(map[int]string) // variable to hold the id: submission name mappings which are sent to the db

	// sets up the test environment (db and filesystem)
	if err = initTestEnvironment(); err != nil {
		t.Errorf("Error initializing the test environment %s", err)
	}
	// uses a slice here so that we can add more submissions to testSubmissions without breaking the test
	for _, proj := range testSubmissions[0:2] {
		submissionId, err = addSubmission(proj)
		if err != nil {
			t.Errorf("Error adding submission %s: %v", proj.Name, err)
		}
		// saves the added submission with its id
		sentSubmissions[submissionId] = proj.Name
	}

	// builds and sends and http get request
	req, err := http.NewRequest("GET", fmt.Sprintf("%s:%s%s", TEST_URL, TEST_SERVER_PORT, ENDPOINT_ALL_SUBMISSIONS), nil)
	resp, err := sendSecureRequest(req, TEAM_ID)
	if err != nil {
		t.Errorf("Error occurred while sending get request to the Go server: %v", err)
	}
	defer resp.Body.Close()
	if err != nil {
		t.Error(err)
	}

	// checks the returned list of submissions for equality with the sent list
	returnedSubmissions := make(map[int]string)
	json.NewDecoder(resp.Body).Decode(&returnedSubmissions)

	// tests that the proper values have been returned
	for k, v := range returnedSubmissions {
		if v != sentSubmissions[k] {
			t.Errorf("Submissions of ids: %d do not have matching names. Given: %s, Returned: %s ", k, sentSubmissions[k], v)
		}
	}

	// Close server.
	if err = srv.Shutdown(context.Background()); err != nil {
		t.Errorf("HTTP server shutdown: %v", err)
	}

	// tears down the test environment (makes sure that if a test fails, the env is still cleared)
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("error while tearing down db: %v", err)
	}
}

// test for basic functionality. Adds 2 submissions to the db with different authors, then queries them and tests for equality
// Test Depends On:
// 	- TestCreateSubmissions()
// 	- TestAddReviewers()
// 	- TestAddAuthors()
func TestGetSingleSubmission(t *testing.T) {
	testSubmission1 := testSubmissions[0] // test submission to return on getUserSubmissions()
	testSubmission2 := testSubmissions[1] // test submission to not return on getUserSubmissions()
	testAuthor := testAuthors[0]          // test author of the submission being queried
	testNonAuthor := testAuthors[3]       // test author of submission not being queried

	// sets up the test environment (db and filesystem)
	if err := initTestEnvironment(); err != nil {
		t.Errorf("Error initializing the test environment %s", err)
	}

	// adds two test users to the db
	authorId, err := registerUser(testAuthor)
	if err != nil {
		t.Errorf("Error occurred while registering user: %v", err)
	}
	nonAuthorId, err := registerUser(testNonAuthor)
	if err != nil {
		t.Errorf("Error occurred while registering user: %v", err)
	}

	// adds two test submissions to the db
	testSubmission1.Id, err = addSubmission(testSubmission1)
	if err != nil {
		t.Errorf("Error occurred while adding submission1: %v", err)
	}
	testSubmission2.Id, err = addSubmission(testSubmission2)
	if err != nil {
		t.Errorf("Error occurred while adding submission2: %v", err)
	}

	// adds authors to the test submissions
	if err = addAuthor(authorId, testSubmission1.Id); err != nil {
		t.Errorf("Failed to add author")
	}
	if err = addAuthor(nonAuthorId, testSubmission2.Id); err != nil {
		t.Errorf("Failed to add author")
	}

	// queries all of testAuthor's submissions
	submissions, err := getUserSubmissions(authorId)
	if err != nil {
		t.Errorf("Error getting user submissions: %v", err)
	}

	// tests for equality of submission Id and that testSubmission2.Id is not in the map
	if _, ok := submissions[testSubmission2.Id]; ok {
		t.Errorf("Returned submission where the test author is not an author")
	} else if submissions[testSubmission1.Id] != testSubmission1.Name {
		t.Errorf("Returned incorrect submission name: %s", submissions[testSubmission1.Id])
	}

	// destroys the test environment
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("Error occurred while destroying the database and filesystem: %v", err)
	}
}

// Tests the ability of the CodeFiles module to get a submission from the db
//
// Test Depends On:
// 	- TestCreateSubmissions()
// 	- TestAddFiles()
// 	- TestAddReviewers()
// 	- TestAddAuthors()
func TestGetSubmission(t *testing.T) {
	var err error
	var submissionId int // holds the submission id as returned from the addSubmission() function

	// Set up server to listen with the getFile() function.
	srv := setupCORSsrv()

	// Start server.
	go srv.ListenAndServe()

	testFile := testFiles[0]             // defines the file to use for the test here so that it can be easily changed
	testSubmission := testSubmissions[0] // defines the submission to use for the test here so that it can be easily changed
	testAuthor := testAuthors[0]         // defines the author of the submission
	testReviewer := testReviewers[0]     // defines the reviewer of the submission

	// initializes the filesystem and db
	if err = initTestEnvironment(); err != nil {
		t.Errorf("Error initializing the test environment: %v", err)
	}
	// adds the test submission to the filesystem and database
	submissionId, err = addSubmission(testSubmission)
	if err != nil {
		t.Errorf("Error adding submission %v", err)
	}
	// adds the test file to the filesystem and database
	_, err = addFileTo(testFile, submissionId)
	if err != nil {
		t.Errorf("Error adding file to the submission %v", err)
	}
	// adds an author and reviewer to the submission
	authorId, err := registerUser(testAuthor)
	if err != nil {
		t.Errorf("Error registering author in the db: %v", err)
	}
	reviewerId, err := registerUser(testReviewer)
	if err != nil {
		t.Errorf("Error registering reviewer in the db: %v", err)
	}

	// adds reviewer and author to the submission
	if err = addAuthor(authorId, submissionId); err != nil {
		t.Errorf("Error adding author to the submission: %v", err)
	}
	if err = addReviewer(reviewerId, submissionId); err != nil {
		t.Errorf("Error adding reviewer to the submission: %v", err)
	}

	// creates a request to send to the test server
	urlString := fmt.Sprintf("%s:%s%s?%s=%d", TEST_URL, TEST_SERVER_PORT, 
		ENDPOINT_SUBMISSION, getJsonTag(&Submission{}, "Id"), submissionId)
	req, err := http.NewRequest("GET", urlString, nil)
	resp, err := sendSecureRequest(req, TEAM_ID)
	if err != nil {
		t.Errorf("Error while sending Get request: %v", err)
	}
	defer resp.Body.Close()

	// if an error occurred while getting the file, it is printed out here
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Error: %d", resp.StatusCode)
	}

	// marshals the json response into a Submission struct
	submission := &Submission{}
	err = json.NewDecoder(resp.Body).Decode(&submission)
	if err != nil {
		t.Error("Error while decoding server response: ", err)
	}

	// tests that the submission matches the passed in data
	if testSubmission.Id != submission.Id {
		t.Errorf("Submission IDs do not match. Given: %d != Returned: %d", testSubmission.Id, submission.Id)
	} else if testSubmission.Name != submission.Name {
		t.Errorf("Submission Names do not match. Given: %s != Returned: %s", testSubmission.Name, submission.Name)
		// tests that file paths match (done directly here as there is only one constituent file)
	} else if testSubmission.FilePaths[0] != submission.FilePaths[0] {
		t.Errorf("Submission file path lists do not match. Given: %s != Returned: %s", testSubmission.FilePaths[0], submission.FilePaths[0])
		// tests that the authors lists match (done directly here as there is only one author)
	} else if authorId != submission.Authors[0] {
		t.Errorf("Authors do not match. Expected: %s Given: %s", authorId, testSubmission.Authors[0])
		// tests that the reviewer lists match (done directly here as there is only one reviewer)
	} else if reviewerId != submission.Reviewers[0] {
		t.Errorf("Authors do not match. Expected: %s Given: %s", reviewerId, testSubmission.Reviewers[0])
	}

	// destroys the filesystem and clears the db
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("Error occurred while destroying the database and filesystem: %v", err)
	}

	// Close server.
	if err = srv.Shutdown(context.Background()); err != nil {
		fmt.Printf("HTTP server shutdown: %v", err)
	}
}
