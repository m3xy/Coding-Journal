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
	"os"
	"strings"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// data to use in the tests
var testSubmissions []Submission = []Submission{
	// valid
	Submission{
		Name:       "TestSubmission1",
		License:    "MIT",
		Authors:    []GlobalUser{},
		Reviewers:  []GlobalUser{},
		Files:      []File{*testFiles[0]},
		Categories: []string{"testtag"},
		MetaData: &SubmissionData{
			Abstract: "test abstract",
			Reviews:  []*Comment{},
		},
	},
	Submission{
		Name:       "TestSubmission2",
		License:    "MIT",
		Authors:    []GlobalUser{},
		Reviewers:  []GlobalUser{},
		Files:      []File{*testFiles[1]},
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

// // TODO: move this function to somewhere more sensible
// // utility function to register an array of users and return an array of their Ids
// func registerUsers(t *testing.T, users []*Credentials) ([]string) {
// 	var id string
// 	var err error
// 	userIds := []string{}
// 	for _, user := range users {
// 		id, err = registerUser(user)
// 		assert.NoError(t, err, "error while registering user")
// 		userIds = append(userIds, id)
// 	}
// 	return userIds
// }

// test the addSubmission() function in submissions.go
func TestAddSubmission(t *testing.T) {
	// Utility function to be re-used for testing adding submissions to the db
	testAddSubmission := func(testSub *Submission) {
		// adds the submission to the db and filesystem
		submissionId, err := addSubmission(testSub)
		assert.NoErrorf(t, err, "Error adding submission: %v", err)

		// retrieve the submission
		queriedSubmission := &Submission{}
		err = gormDb.Model(&Submission{}).First(queriedSubmission, testSub.ID).Error
		assert.NoError(t, err, "Error retrieving submission: %v", err)

		// checks that the filesystem has a proper corresponding entry and metadata file
		submissionData := &SubmissionData{}
		submissionDirPath := filepath.Join(TEST_FILES_DIR, fmt.Sprint(submissionId))
		fileDataPath := filepath.Join(submissionDirPath, DATA_DIR_NAME, testSub.Name+".json")
		dataString, err := ioutil.ReadFile(fileDataPath)
		assert.NoError(t, err, "error reading submission data")
		assert.NoError(t, json.Unmarshal(dataString, submissionData), "error unmarshalling submission data")

		// for each file in the submission, checks that it was added to the filesystem properly
		for _, file := range queriedSubmission.Files {
			// retrieve the submission
			queriedFile := &File{}
			err = gormDb.Model(&File{}).First(queriedFile, file.ID).Error
			assert.NoError(t, err, "Error retrieving file: %v", err)

			// gets the file content from the filesystem
			fileBytes, err := ioutil.ReadFile(queriedFile.Path)
			assert.NoErrorf(t, err, "File read failure after added to filesystem: %v", err)
			queriedFileContent := string(fileBytes)

			// checks that a data file has been generated for the uploaded file
			var fileData *FileData
			fileDataPath := filepath.Join(
				TEST_FILES_DIR,
				fmt.Sprint(queriedSubmission.ID),
				DATA_DIR_NAME,
				queriedSubmission.Name,
				strings.TrimSuffix(queriedFile.Path, filepath.Ext(queriedFile.Path))+".json",
			)
			dataString, err := ioutil.ReadFile(fileDataPath)
			assert.NoError(t, err, "error reading submission data")
			assert.NoError(t, json.Unmarshal(dataString, fileData), "error unmarshalling submission data")
	
			// gets data about the file, and tests it for equality against the added file
			_, err = os.Stat(fileDataPath)
			assert.NotErrorIs(t, err, os.ErrNotExist, "Data file not generated during file upload")
			assert.Equal(t, file.Name, queriedFile.Name, "File names do not match")
			assert.Equal(t, file.Path, queriedFile.Path, "File Paths do not match")
			assert.Equal(t, file.SubmissionID, queriedFile.SubmissionID, "File SubmissionIDs do not match")
			assert.Equal(t, file.Base64Value, queriedFileContent, "file content not written to filesystem properly")
			assert.ElementsMatch(t, file.MetaData.Comments, fileData.Comments, "File comments do not match")
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

// // tests the ability of the submissions.go module to add authors to projects
// // Test Depends on:
// // 	- TestAddSubmission
// func TestAddAuthor(t *testing.T) {
// 	// utility function to add submissions just to the database, without checking for correctness
// 	testAddSubmission := func(submission *Submission) (int) {
// 		// formats query to insert the submission into the db
// 		insertSubmission := fmt.Sprintf(
// 			INSERT_SUBMISSION,
// 			TABLE_SUBMISSIONS,
// 			getDbTag(&Submission{}, "Name"),
// 			getDbTag(&Submission{}, "License"),
// 		)

// 		// executes the query and gets the submission id
// 		var submissionId int
// 		row := db.QueryRow(insertSubmission, submission.Name, submission.License)
// 		assert.NoError(t, row.Scan(&submissionId), "failed to add submission to the db")
// 		return submissionId
// 	}

// 	// utility function which tests that an author can be added to a valid submission properly
// 	testAddAuthor := func(submissionId int, author *Credentials) {
// 		// registers the author as a user and then to the given submission as an author
// 		authorId, err := registerUser(author)
// 		assert.NoErrorf(t, err, "Error in author registration: %v", err)
// 		assert.NoErrorf(t, addAuthor(authorId, submissionId), "Error adding the author to the db: %v", err)

// 		// queries the authors table in the database for author IDs which match that returned from registerUser()
// 		var queriedAuthors []string
// 		var queriedSubmissionId int
// 		var queriedAuthorId string
// 		queryAuthor := fmt.Sprintf(SELECT_ROW, "*", TABLE_AUTHORS, "userId")
// 		assert.NoErrorf(t, err, "Error querying submission Authors: %v", err)
// 		rows, err:= db.Query(queryAuthor, authorId)
// 		assert.NoErrorf(t, err, "Error querying submission Authors: %v", err)
// 		for rows.Next() {
// 			rows.Scan(&queriedSubmissionId, &queriedAuthorId)
// 			queriedAuthors = append(queriedAuthors, queriedAuthorId)
// 		}

// 		// checks submission Id and author Id for equality with those which were queried
// 		assert.Equalf(t, submissionId, queriedSubmissionId,
// 			"Author added to the wrong submission: Wanted: %d Got: %d", submissionId, queriedSubmissionId)
// 		assert.Contains(t, queriedAuthors, authorId,
// 			"Author not added to the submission properly")
// 	}

// 	// tests adding one author to a valid project
// 	t.Run("Adding One Author", func(t *testing.T) {
// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
// 		// defines test submission and author to use for this test, and uploads the submission
// 		testSubmission := testSubmissions[0]
// 		testAuthor := testAuthors[0]
// 		submissionId := testAddSubmission(testSubmission)

// 		// uses the utility function to add the author, and test that it was done properly
// 		testAddAuthor(submissionId, testAuthor)
// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 	})

// 	// attemps to add an author without the correct permissions, if addAuthor succeeds, an error is thrown
// 	t.Run("Add Invalid Author", func(t *testing.T) {
// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
// 		testSubmission := testSubmissions[0]
// 		testAuthor := testAuthors[1] // user without publisher permissions

// 		// uploads a test submission
// 		submissionId := testAddSubmission(testSubmission)

// 		// registers the author as a user and then to the given submission as an author
// 		authorId, err := registerUser(testAuthor)
// 		assert.NoErrorf(t, err, "Error in author registration: %v", err)
// 		assert.Error(t, addAuthor(authorId, submissionId), "Author without publisher permissions registered")

// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 	})

// 	// tests that a user must be registered with the db before being and author
// 	t.Run("Add Non-User Author", func(t *testing.T) {
// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
// 		testSubmission := testSubmissions[0]
// 		authorId := "u881jafjka" // non-user fake id
// 		submissionId := testAddSubmission(testSubmission)
// 		assert.Error(t, addAuthor(authorId, submissionId), "Non-user added as author")
// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 	})
// }

// // tests the ability of the submissions.go module to add reviewers to projects
// // Test Depends on:
// // 	- TestAddSubmission
// func TestAddReviewer(t *testing.T) {
// 	// utility function to add submissions just to the database, without checking for correctness
// 	testAddSubmission := func(submission *Submission) (int) {
// 		// formats query to insert the submission into the db
// 		insertSubmission := fmt.Sprintf(
// 			INSERT_SUBMISSION,
// 			TABLE_SUBMISSIONS,
// 			getDbTag(&Submission{}, "Name"),
// 			getDbTag(&Submission{}, "License"),
// 		)

// 		// executes the query and gets the submission id
// 		var submissionId int
// 		row := db.QueryRow(insertSubmission, submission.Name, submission.License)
// 		assert.NoError(t, row.Scan(&submissionId), "failed to add submission to the db")
// 		return submissionId
// 	}

// 	// utility function which tests that a reviewer can be added to a valid submission properly
// 	testAddReviewer := func(submissionId int, reviewer *Credentials) {
// 		// registers the reviewer as a user and then to the given submission as an reviewer
// 		reviewerId, err := registerUser(reviewer)
// 		assert.NoErrorf(t, err, "Error in reviewer registration: %v", err)
// 		assert.NoErrorf(t, addReviewer(reviewerId, submissionId), "Error adding the reviewer to the db: %v", err)

// 		// queries the authors table in the database for author IDs which match that returned from registerUser()
// 		var queriedReviewers []string
// 		var queriedSubmissionId int
// 		var queriedReviewerId string
// 		queryReviewer := fmt.Sprintf(SELECT_ROW, "*", TABLE_REVIEWERS, "userId")
// 		rows, err:= db.Query(queryReviewer, reviewerId)
// 		assert.NoErrorf(t, err, "Error querying submission Reviewers: %v", err)
// 		for rows.Next() {
// 			rows.Scan(&queriedSubmissionId, &queriedReviewerId)
// 			queriedReviewers = append(queriedReviewers, queriedReviewerId)
// 		}

// 		// checks submission Id and reviewer Id for equality with those which were queried
// 		assert.Equalf(t, submissionId, queriedSubmissionId,
// 			"Reviewer added to the wrong submission: Wanted: %d Got: %d", submissionId, queriedSubmissionId)
// 		assert.Contains(t, queriedReviewers, reviewerId,
// 			"Reviewer not added to the submission properly")
// 	}

// 	// tests adding one reviewer to a valid project
// 	t.Run("Adding One Reviewer", func(t *testing.T) {
// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
// 		// defines test submission and reviewer to use for this test, and uploads the submission
// 		testSubmission := testSubmissions[0]
// 		testReviewer := testReviewers[0]
// 		submissionId := testAddSubmission(testSubmission)

// 		// uses the utility function to add the reviewer, and test that it was done properly
// 		testAddReviewer(submissionId, testReviewer)
// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 	})

// 	// attemps to add an reviewer without the correct permissions, if addReviewer succeeds, an error is thrown
// 	t.Run("Add Invalid Reviewer", func(t *testing.T) {
// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
// 		testSubmission := testSubmissions[0]
// 		testReviewer := testReviewers[1] // user without publisher permissions
// 		submissionId := testAddSubmission(testSubmission)

// 		// registers the reviewer as a user and then to the given submission as an reviewer
// 		reviewerId, err := registerUser(testReviewer)
// 		assert.NoErrorf(t, err, "Error in reviewer registration: %v", err)
// 		assert.Error(t, addReviewer(reviewerId, submissionId), "Reviewer without reviewer permissions registered")

// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 	})

// 	// tests that a user must be registered with the db before being and reviewer
// 	t.Run("Add Non-User Reviewer", func(t *testing.T) {
// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")

// 		testSubmission := testSubmissions[0]
// 		reviewerId := "u881jafjka" // non-user fake id
// 		submissionId := testAddSubmission(testSubmission)
// 		assert.Error(t, addReviewer(reviewerId, submissionId), "Non-user added as reviewer")

// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 	})
// }

// // This function tests the addTag function
// //
// // This test depends on:
// // 	- TestAddSubmission
// func TestAddTag(t *testing.T) {
// 	// standard use case
// 	t.Run("Add Valid Tag", func(t *testing.T) {
// 		testTag := "TEST"

// 		// sets up the test environment, and uploads a test submission
// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
// 		testSubmission := testSubmissions[0]
// 		testSubmission.Authors = registerUsers(t, testAuthors[:1])
// 		testSubmission.Reviewers = registerUsers(t, testReviewers[:1])
// 		submissionId, err := addSubmission(testSubmission)
// 		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)

// 		// adds a tag to the submission
// 		addTag(testTag, submissionId)

// 		// queries the db to make sure the tag was added properly
// 		queryTagExists := fmt.Sprintf(
// 			SELECT_ROW_TWO_CONDITION,
// 			"*",
// 			TABLE_CATEGORIES,
// 			getDbTag(&Categories{}, "SubmissionId"),
// 			getDbTag(&Categories{}, "Tag"),
// 		)
// 		row := db.QueryRow(queryTagExists, submissionId, testTag)
// 		assert.NoError(t, row.Err(), "Tag not added to the database properly")

// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 	})

// 	// duplicate tag case (no error thrown, tag just not added)
// 	t.Run("Add Duplicate Tag", func(t *testing.T) {
// 		testTag := "TEST"

// 		// sets up the test environment, and uploads a test submission
// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
// 		testSubmission := testSubmissions[0]
// 		testSubmission.Authors = registerUsers(t, testAuthors[:1])
// 		testSubmission.Reviewers = registerUsers(t, testReviewers[:1])
// 		submissionId, err := addSubmission(testSubmission)
// 		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)

// 		// adds a tag to the submission twice
// 		addTag(testTag, submissionId)
// 		assert.Error(t, addTag(testTag, submissionId), "attempting to add duplicate tag does not return an error")

// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 	})

// 	// duplicate tag case (no error thrown, tag just not added)
// 	t.Run("Add Same Tag to Different Submissions", func(t *testing.T) {
// 		testTag := "TEST"

// 		// sets up the test environment, and uploads a test submission
// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
// 		testSubmission1 := testSubmissions[0]
// 		testSubmission2 := testSubmissions[1]
// 		authors := registerUsers(t, testAuthors[:1])
// 		reviewers := registerUsers(t, testReviewers[:1])
// 		testSubmission1.Authors = authors
// 		testSubmission1.Reviewers = reviewers
// 		testSubmission2.Authors = authors
// 		testSubmission2.Reviewers = reviewers

// 		// adds the submissions
// 		submissionId1, err := addSubmission(testSubmission1)
// 		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)
// 		submissionId2, err := addSubmission(testSubmission2)
// 		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)

// 		// adds the tags
// 		assert.NoError(t, addTag(testTag, submissionId1), "Error occurred while adding 1st tag")
// 		assert.NoError(t, addTag(testTag, submissionId2), "Error occurred while adding 2nd tag")

// 		// queries the db to make sure the tag was added properly
// 		queryTag := fmt.Sprintf(
// 			SELECT_ROW,
// 			getDbTag(&Categories{}, "Tag"),
// 			TABLE_CATEGORIES,
// 			getDbTag(&Categories{}, "SubmissionId"),
// 		)
// 		// adds the first tag
// 		rows, err := db.Query(queryTag, submissionId1)
// 		rows.Close()
// 		assert.NoError(t, err, "Tag 1 not added to the database properly")

// 		// adds the second tag
// 		rows, err = db.Query(queryTag, submissionId2)
// 		rows.Close()
// 		assert.NoError(t, err, "Tag 2 not added to the database properly")

// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 	})

// 	// invalid tag case
// 	t.Run("Add Empty Tag", func(t *testing.T) {
// 		// sets up the test environment, and uploads a test submission
// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
// 		testSubmission := testSubmissions[0]
// 		testSubmission.Authors = registerUsers(t, testAuthors[:1])
// 		testSubmission.Reviewers = registerUsers(t, testReviewers[:1])
// 		submissionId, err := addSubmission(testSubmission)
// 		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)

// 		// adds an invalid tag to an existing submission
// 		assert.Error(t, addTag("", submissionId), "empty tag was able ot be added")

// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 	})

// 	// add tag to non-existant project (foreign key constraint fails in db)
// 	t.Run("Add Tag Invalid Project", func(t *testing.T) {
// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
// 		assert.Error(t, addTag("INVALID_PROJECT", 10), "Error not thrown when tag added to a non-existant submission")
// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 	})
// }

// // test for basic functionality. Adds 2 submissions to the db with different authors, then queries them and tests for equality
// // Test Depends On:
// // 	- TestAddSubmission
// // 	- TestAddAuthors
// func TestGetUserSubmissions(t *testing.T) {
// 	// adds two submissions each with different authors to the db and then queries one author's submissions
// 	t.Run("Get Single Submission from an Author", func(t *testing.T) {
// 		testSubmission1 := testSubmissions[0] // test submission to return on getUserSubmissions()
// 		testSubmission2 := testSubmissions[1] // test submission to not return on getUserSubmissions()
// 		testAuthor := testAuthors[0]          // test author of the submission being queried
// 		testNonAuthor := testAuthors[3]       // test author of submission not being queried

// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")

// 		// adds two test users to the db as authors
// 		authorId, err := registerUser(testAuthor)
// 		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
// 		testSubmission1.Authors = []string{authorId}

// 		nonAuthorId, err := registerUser(testNonAuthor) // author of the submission we are not interested in
// 		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
// 		testSubmission2.Authors = []string{nonAuthorId}

// 		// adds dummy reviewers
// 		reviewerId, err := registerUser(testReviewers[0])
// 		assert.NoErrorf(t, err, "Error occurred while registering user: %v", err)
// 		testSubmission1.Reviewers = []string{reviewerId}
// 		testSubmission2.Reviewers = []string{reviewerId}

// 		// adds two test submissions to the db
// 		testSubmission1.Id, err = addSubmission(testSubmission1)
// 		assert.NoErrorf(t, err, "Error occurred while adding submission1: %v", err)
// 		testSubmission2.Id, err = addSubmission(testSubmission2)
// 		assert.NoErrorf(t, err, "Error occurred while adding submission2: %v", err)

// 		// queries all of testAuthor's submissions
// 		submissions, err := getUserSubmissions(authorId)
// 		assert.NoErrorf(t, err, "Error getting user submissions: %v", err)

// 		// tests for equality of submission Id and that testSubmission2.Id is not in the map
// 		_, ok := submissions[testSubmission2.Id]
// 		assert.False(t, ok, "Returned submission where the test author is not an author")
// 		assert.Equalf(t, submissions[testSubmission1.Id], testSubmission1.Name,
// 			"Returned incorrect submission name: %s", submissions[testSubmission1.Id])

// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 	})
// }

// // tests the getSubmission() function, which returns a submission struct
// //
// // Test Depends On:
// // 	- TestAddSubmission
// // 	- TestAddFile
// func TestGetSubmission(t *testing.T) {
// 	// tests the basic case of getting back a valid submission
// 	t.Run("Single Valid Submission", func(t *testing.T) {
// 		testSubmission := testSubmissions[0]
// 		testFile := testFiles[0]

// 		// sets up test environment, and adds a submission with one file to the db and filesystem
// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
// 		testSubmission.Authors = registerUsers(t, testAuthors[:1])
// 		testSubmission.Reviewers = registerUsers(t, testReviewers[:1])
// 		submissionId, err := addSubmission(testSubmission)
// 		assert.NoErrorf(t, err, "Error occurred while adding submission: %v", err)
// 		_, err = addFileTo(testFile, submissionId)
// 		assert.NoErrorf(t, err, "Error occurred while adding test file: %v", err)

// 		// gets the submission back
// 		queriedSubmission, err := getSubmission(submissionId)

// 		assert.NoErrorf(t, err, "Error occurred while retrieving submission: %v", err)
// 		// tests the submission was returned properly
// 		assert.Equal(t, testSubmission.Name, queriedSubmission.Name, "Submission names do not match")
// 		assert.Equal(t, testSubmission.License, queriedSubmission.License, "Submission Licenses do not match")
// 		assert.ElementsMatch(t, testSubmission.Authors, queriedSubmission.Authors, "Author lists do not match")
// 		assert.ElementsMatch(t, testSubmission.Reviewers, queriedSubmission.Reviewers, "Reviewer arrays do not match")
// 		assert.ElementsMatch(t, testSubmission.Reviewers, queriedSubmission.Reviewers, "Reviewer arrays do not match")
// 		assert.ElementsMatch(t, testSubmission.FilePaths, queriedSubmission.FilePaths, "File path arrays do not match")
// 		assert.ElementsMatch(t, testSubmission.Categories, queriedSubmission.Categories, "Submission tags do not match")
// 		assert.Equal(t, testSubmission.MetaData.Abstract, queriedSubmission.MetaData.Abstract, "Abstracts do not match")

// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 	})

// 	// tests trying to get an invalid submission
// 	t.Run("Invalid Submission", func(t *testing.T) {
// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
// 		_, err := getSubmission(100)
// 		assert.Errorf(t, err, "No error was thrown for invalid submission")
// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 	})
// }

// // This function tests the getSubmissionAuthors function
// //
// // This test depends on:
// // 	- TestAddSubmission
// func TestGetSubmissionAuthors(t *testing.T) {
// 	// valid metadata file and format
// 	t.Run("Valid Author List", func(t *testing.T) {
// 		// sets up the test environment, and uploads a test submission
// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
// 		testSubmission := testSubmissions[0]
// 		testSubmission.Authors = registerUsers(t, testAuthors[:1])
// 		testSubmission.Reviewers = registerUsers(t, testReviewers[:1])
// 		submissionId, err := addSubmission(testSubmission)
// 		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)

// 		// tests that the metadata can be read back properly, and that it matches the uploaded submission
// 		authors, err := getSubmissionAuthors(submissionId)
// 		assert.NoErrorf(t, err, "Error getting submission authors: %v", err)
// 		assert.ElementsMatch(t, testSubmission.Authors, authors, "authors lists do not match")

// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 	})
// }

// // This function tests the getSubmissionReviewers function
// //
// // This test depends on:
// // 	- TestAddSubmission
// func TestGetSubmissionReviewers(t *testing.T) {
// 	// valid metadata file and format
// 	t.Run("Valid Reviewer List", func(t *testing.T) {
// 		// sets up the test environment, and uploads a test submission
// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
// 		testSubmission := testSubmissions[0]
// 		testSubmission.Authors = registerUsers(t, testAuthors[:1])
// 		testSubmission.Reviewers = registerUsers(t, testReviewers[:1])
// 		submissionId, err := addSubmission(testSubmission)
// 		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)

// 		// tests that the metadata can be read back properly, and that it matches the uploaded submission
// 		reviewers, err := getSubmissionReviewers(submissionId)
// 		assert.NoErrorf(t, err, "Error getting submission reviewers: %v", err)
// 		assert.ElementsMatch(t, testSubmission.Reviewers, reviewers, "reviewers lists do not match")

// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 	})
// }

// // This function tests the getSubmissionCategories function
// //
// // This test depends on:
// // 	- TestAddSubmission
// func TestGetSubmissionCategories(t *testing.T) {
// 	// valid metadata file and format
// 	t.Run("Valid Categories", func(t *testing.T) {
// 		// sets up the test environment, and uploads a test submission
// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
// 		testSubmission := testSubmissions[0]
// 		testSubmission.Authors = registerUsers(t, testAuthors[:1])
// 		testSubmission.Reviewers = registerUsers(t, testReviewers[:1])
// 		submissionId, err := addSubmission(testSubmission)
// 		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)

// 		// tests that the metadata can be read back properly, and that it matches the uploaded submission
// 		categories, err := getSubmissionCategories(submissionId)
// 		assert.NoErrorf(t, err, "Error getting submission tags: %v", err)
// 		assert.ElementsMatch(t, testSubmission.Categories, categories, "Submission tags do not match")

// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 	})
// }

// // This function tests the getSubmissionFilePaths function
// //
// // This test depends on:
// // 	- TestAddSubmission
// func TestGetSubmissionFilePaths(t *testing.T) {
// 	// valid metadata file and format
// 	t.Run("Valid Submission", func(t *testing.T) {
// 		// sets up the test environment, and uploads a test submission
// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
// 		testSubmission := testSubmissions[0]
// 		testFile := testFiles[0]
// 		testSubmission.Authors = registerUsers(t, testAuthors[:1])
// 		testSubmission.Reviewers = registerUsers(t, testReviewers[:1])
// 		submissionId, err := addSubmission(testSubmission)
// 		_, err = addFileTo(testFile, submissionId)
// 		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)

// 		// tests that the metadata can be read back properly, and that it matches the uploaded submission
// 		filePaths, err := getSubmissionFilePaths(submissionId)
// 		// note that testSubmission.FilePaths is not used here, as this array is hard-coded and thus not useful here
// 		assert.ElementsMatch(t, []string{testFile.Path}, filePaths, "Filepath arrays do not match")

// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 	})
// }

// // This function tests the getSubmissionMetaData function
// //
// // This test depends on:
// // 	- TestAddSubmission
// func TestGetSubmissionMetaData(t *testing.T) {
// 	// valid metadata file and format
// 	t.Run("Valid Metadata", func(t *testing.T) {
// 		// sets up the test environment, and uploads a test submission
// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
// 		testSubmission := testSubmissions[0]
// 		testSubmission.Authors = registerUsers(t, testAuthors[:1])
// 		testSubmission.Reviewers = registerUsers(t, testReviewers[:1])
// 		submissionId, err := addSubmission(testSubmission)
// 		assert.NoErrorf(t, err, "Error occurred while adding test submission: %v", err)

// 		// tests that the metadata can be read back properly, and that it matches the uploaded submission
// 		submissionData, err := getSubmissionMetaData(submissionId)
// 		assert.NoErrorf(t, err, "Error getting submission metadata: %v", err)
// 		assert.Equalf(t, submissionData.Abstract, testSubmission.MetaData.Abstract,
// 			"submission metadata not added to filesystem properly. Abstracts %s, %s do not match",
// 			submissionData.Abstract, testSubmission.MetaData.Abstract)
// 		assert.ElementsMatch(t, submissionData.Reviews, testSubmission.MetaData.Reviews, "Submission Reviews do not match")

// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 	})

// 	// Tests that getSubmissionMetaData will throw an error if an incorrect submission ID is passed in
// 	t.Run("Invalid Submission Id", func(t *testing.T) {
// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
// 		_, err := getSubmissionMetaData(400)
// 		assert.Errorf(t, err, "No error was thrown for invalid submission")
// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 	})
// }

// // This tests converting from the local submission data format to the supergroup specified format
// func TestLocalToGlobal(t *testing.T) {
// 	// tests valid submission struct
// 	t.Run("Valid Submission", func(t *testing.T) {
// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")

// 		// adds the submission and a file to the system
// 		testSubmission := testSubmissions[0]
// 		testFile := testFiles[0]
// 		testAuthor := testAuthors[0]
// 		testSubmission.Authors = registerUsers(t, []*Credentials{testAuthor})
// 		testSubmission.Reviewers = registerUsers(t, testReviewers[:1])
// 		submissionId, err := addSubmission(testSubmission)
// 		assert.NoErrorf(t, err, "Error occurred while adding submission: %v", err)
// 		_, err = addFileTo(testFile, submissionId)
// 		assert.NoErrorf(t, err, "Error occurred while adding file to submission: %v", err)

// 		// gets the supergroup compliant submission
// 		globalSubmission, err := localToGlobal(submissionId)
// 		assert.NoErrorf(t, err, "Error occurred while converting submission format: %v", err)

// 		// compares submission fields
// 		assert.Equal(t, testSubmission.Name, globalSubmission.Name, "Names do not match")
// 		assert.Equal(t, testSubmission.License, globalSubmission.MetaData.License,
// 			"Licenses do not match")
// 		assert.Equal(t, testAuthor.Fname+" "+testAuthor.Lname, globalSubmission.MetaData.AuthorNames[0],
// 			"Authors do not match")
// 		assert.Equal(t, testSubmission.Categories, globalSubmission.MetaData.Categories,
// 			"Tags do not match")
// 		assert.Equal(t, testSubmission.MetaData.Abstract, globalSubmission.MetaData.Abstract,
// 			"Abstracts do not match")
// 		// compares files
// 		assert.Equal(t, testFile.Name, globalSubmission.Files[0].Name, "File names do not match")
// 		assert.Equal(t, testFile.Base64Value, globalSubmission.Files[0].Base64Value, "File content does not match")

// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 	})
// }

// // Tests the ability of the getAllSubmissions() function to get all submission ids and names from the db
// // at once
// //
// // Test Depends On:
// // 	- TestAddSubmission
// // 	- TestGetUserSubmissions
// func TestGetAllSubmissions(t *testing.T) {
// 	// tests that multiple valid submissions can be uploaded, then retrieved from the database
// 	t.Run("Get Multiple Valid submissions", func(t *testing.T) {
// 		// Set up server and test environment
// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
// 		srv := setupCORSsrv()
// 		go srv.ListenAndServe()

// 		// registers authors and reviewers of the submissions (same for all submissions here)
// 		authors := registerUsers(t, testAuthors[:1])
// 		reviewers := registerUsers(t, testReviewers[:1])

// 		// adds all of the submissions and stores their ids and names
// 		sentSubmissions := make(map[int]string) // variable to hold the id: submission name mappings which are sent to the db
// 		for _, sub := range testSubmissions[0:2] {
// 			sub.Authors = authors
// 			sub.Reviewers = reviewers
// 			submissionId, err := addSubmission(sub)
// 			assert.NoErrorf(t, err, "Error adding submission %s: %v", sub.Name, err)
// 			sentSubmissions[submissionId] = sub.Name
// 		}

// 		// builds and sends and http request to get the names and Ids of all submissions
// 		req, err := http.NewRequest("GET", fmt.Sprintf("%s:%s%s", TEST_URL, TEST_SERVER_PORT, ENDPOINT_ALL_SUBMISSIONS), nil)
// 		resp, err := sendSecureRequest(req, TEAM_ID)
// 		assert.NoErrorf(t, err, "Error occurred while sending get request to the Go server: %v", err)
// 		defer resp.Body.Close()

// 		// checks the returned list of submissions for equality with the sent list
// 		returnedSubmissions := make(map[int]string)
// 		json.NewDecoder(resp.Body).Decode(&returnedSubmissions)
// 		for k, v := range returnedSubmissions {
// 			assert.Equalf(t, v,  sentSubmissions[k],
// 				"Submissions of ids: %d do not have matching names. Given: %s, Returned: %s ", k, sentSubmissions[k], v)
// 		}

// 		// clears test env and shuts down the test server
// 		assert.NoError(t, srv.Shutdown(context.Background()), "failed to shut down server")
// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 	})
// }

// // Tests the ability of the CodeFiles module to get a submission from the db
// //
// // Test Depends On:
// // 	- TestCreateSubmissions()
// // 	- TestAddFiles()
// // 	- TestAddReviewers()
// // 	- TestAddAuthors()
// func TestSendSubmission(t *testing.T) {
// 	// tests that a single valid submission with one reviewer and one author can be retrieved
// 	t.Run("Get Valid Submission", func(t *testing.T) {
// 		testFile := testFiles[0]             // defines the file to use for the test here so that it can be easily changed
// 		testSubmission := testSubmissions[0] // defines the submission to use for the test here so that it can be easily changed
// 		testAuthor := testAuthors[0]         // defines the author of the submission
// 		testReviewer := testReviewers[0]     // defines the reviewer of the submission

// 		// Set up server and test environment
// 		assert.NoError(t, initTestEnvironment(), "failed to initialise test environment")
// 		srv := setupCORSsrv()
// 		go srv.ListenAndServe()

// 		// registers author and reviewer
// 		authorId, err := registerUser(testAuthor)
// 		assert.NoErrorf(t, err, "Error registering author in the db: %v", err)
// 		testSubmission.Authors = []string{authorId}
// 		reviewerId, err := registerUser(testReviewer)
// 		assert.NoErrorf(t, err, "Error registering reviewer in the db: %v", err)
// 		testSubmission.Reviewers = []string{reviewerId}

// 		// uploads the test submission and adds a file to it
// 		submissionId, err := addSubmission(testSubmission)
// 		assert.NoErrorf(t, err, "Error adding submission %v", err)
// 		_, err = addFileTo(testFile, submissionId)
// 		assert.NoErrorf(t, err, "Error adding file to the submission %v", err)

// 		// creates a request to send to the test server
// 		urlString := fmt.Sprintf("%s:%s%s?%s=%d", TEST_URL, TEST_SERVER_PORT,
// 			ENDPOINT_SUBMISSION, getJsonTag(&Submission{}, "Id"), submissionId)
// 		req, _ := http.NewRequest("GET", urlString, nil)
// 		resp, err := sendSecureRequest(req, TEAM_ID)
// 		assert.NoErrorf(t, err, "Error while sending Get request: %v", err)
// 		defer resp.Body.Close()
// 		assert.Equalf(t, http.StatusOK, resp.StatusCode, "Non-OK status returned from GET request: %d", resp.StatusCode)

// 		// decodes the json response into a Submission struct
// 		submission := &Submission{}
// 		assert.NoErrorf(t, json.NewDecoder(resp.Body).Decode(submission), "Error while decoding server response: %v", err)

// 		// tests that the returned submission matches the passed in data
// 		assert.Equalf(t, testSubmission.Id, submission.Id,
// 			"Submission IDs do not match. Given: %d != Returned: %d", testSubmission.Id, submission.Id)
// 		assert.Equalf(t, testSubmission.Name, submission.Name,
// 			"Submission Names do not match. Given: %s != Returned: %s", testSubmission.Name, submission.Name)
// 		assert.ElementsMatch(t, []string{authorId}, submission.Authors, "Authors do not match")
// 		assert.ElementsMatch(t, []string{reviewerId}, submission.Reviewers, "Reviewers do not match")
// 		assert.ElementsMatch(t, testSubmission.FilePaths, submission.FilePaths, "Submission file path lists do not match.")

// 		// clears test env and shuts down the test server
// 		assert.NoError(t, srv.Shutdown(context.Background()), "failed to shut down server")
// 		assert.NoError(t, clearTestEnvironment(), "failed to tear down test environment")
// 	})
// }
