/*
CodeFiles_test.go
author: 190010425
created: November 2, 2021

Test file for the CodeFiles module.

Note that the tests are written dependency wise from top to bottom. This means
that if a test breaks, then most of the tests below it will also break. Hence if
a test breaks, fix the top one first and then re-run
*/

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
	"reflect"
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

// Constants for testing
var testProjects []*Project = []*Project{
	{Id: -1, Name: "testProject1", Reviewers: []string{},
		Authors: []string{}, FilePaths: []string{"testFile1.txt"}},
	{Id: -1, Name: "testProject2", Reviewers: []string{},
		Authors: []string{}, FilePaths: []string{"testFile2.txt"}},
}
var testFiles []*File = []*File{
	{Id: -1, ProjectId: -1, ProjectName: "testProject1", Path: "testFile1.txt",
		Name: "testFile1.txt", Content: "hello world", Comments: nil},
	{Id: -1, ProjectId: -1, ProjectName: "testProject1", Path: "testFile2.txt",
		Name: "testFile2.txt", Content: "hello world", Comments: nil},
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

/*
  Tests the functionality to create projects in the database and filesystem from a
  Project struct
*/
func TestCreateProjects(t *testing.T) {
	var err error

	// tests basic functionality with a valid test project. re-used in many tests
	testAddProject := func(testProject *Project) error {
		projectId, err := addProject(testProject)

		// simple error cases
		if err != nil {
			return err
		} else if projectId < 0 {
			return errors.New(fmt.Sprintf("Invalid Project ID returned: %d", projectId))
		}

		// checks manually that the project was added correctly
		var projectName string
		authors := []string{}
		reviewers := []string{}
		// builds SQL Queries for testing the added values
		queryProjectName := fmt.Sprintf(SELECT_ROW, getDbTag(&Project{}, "Name"),
			TABLE_PROJECTS, getDbTag(&Project{}, "Id"))
		queryAuthors := fmt.Sprintf(SELECT_ROW, "userId",
			TABLE_AUTHORS, "projectId")
		queryReviewers := fmt.Sprintf(SELECT_ROW, "userId",
			TABLE_REVIEWERS, "projectId")

		// tests that the project name was added correctly
		row := db.QueryRow(queryProjectName, projectId)
		if row.Err() != nil {
			return errors.New(fmt.Sprintf("Error in project name query: %v", row.Err()))
		} else if err = row.Scan(&projectName); err != nil {
			return err
		} else if testProject.Name != projectName {
			return errors.New(
				fmt.Sprintf("Project name mismatch. %s vs %s",
					testProject.Name, projectName))
		}

		// tests that the authors were added correctly
		rows, err := db.Query(queryAuthors, projectId)
		if err != nil {
			return errors.New(fmt.Sprintf("Error querying project Authors: %v", err))
		}
		var author string
		for rows.Next() {
			rows.Scan(&author)
			authors = append(authors, author)
		}
		if !(reflect.DeepEqual(testProject.Authors, authors)) {
			return errors.New("authors arrays do not match")
		}

		// tests that the reviewers were added correctly
		rows, err = db.Query(queryReviewers, projectId)
		if err != nil {
			return errors.New(fmt.Sprintf("error querying project Reviewers: %v", err))
		}
		var reviewer string
		for rows.Next() {
			rows.Scan(&reviewer)
			reviewers = append(reviewers, reviewer)
		}
		if !(reflect.DeepEqual(testProject.Reviewers, reviewers)) {
			return errors.New("reviewers arrays do not match")
		}
		return nil
	}

	// wrapper to init and teardown test environment to test adding a single project
	testAddSingleProj := func(project *Project) error {
		// initializes the test environment, returning an error if any occurs
		if err = initTestEnvironment(); err != nil {
			return errors.New(fmt.Sprintf("Error in testdb init: %v", err))
		}
		// adds the project and tests that it was added properly
		if err = testAddProject(project); err != nil {
			return err // error already formatted here
		}
		// tears down the test environment
		if err = clearTestEnvironment(); err != nil {
			return errors.New(fmt.Sprintf("Error in db teardown: %v", err))
		}
		return nil
	}

	// test to add n projects in a row
	testAddNProjects := func(projects []*Project) error {
		// initializes the test environment, returning an error if any occurs
		if err = initTestEnvironment(); err != nil {
			return errors.New(
				fmt.Sprintf("error while initializing the test environment db: %v", err))
		}
		for _, project := range projects {
			if err = testAddProject(project); err != nil {
				return err
			}
		}
		// tears down the test environment
		if err = clearTestEnvironment(); err != nil {
			return errors.New(fmt.Sprintf("error while tearing down db: %v", err))
		}
		return nil
	}

	testAddNilProject := func() error {
		// initializes the test environment, returning an error if any occurs
		if err = initTestEnvironment(); err != nil {
			return errors.New(
				fmt.Sprintf("error while initializing the test environment db: %v", err))
		}
		// tries to add a nil project
		if _, err = addProject(nil); err == nil {
			return errors.New("nil project added without error")
		}
		// tears down the test environment
		if err = clearTestEnvironment(); err != nil {
			return errors.New(fmt.Sprintf("error while tearing down db: %v", err))
		}
		return nil
	}

	// runs tests
	if err = testAddSingleProj(testProjects[0]); err != nil {
		t.Errorf("testAddSingleProj failed for testProjects[0]: %v", err)
	} else if err = testAddNProjects(testProjects[0:2]); err != nil {
		t.Errorf("testAddNProjects failed for testProjects[0:2]: %v", err)
	} else if err = testAddNilProject(); err != nil {
		t.Errorf("testAddNilProject failed: %v", err)
	}

	// tears down the test environment (makes sure that if a test fails, the env is still cleared)
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("error while tearing down db: %v", err)
	}
}

// This function tests adding authors to a given project.
//
// Test Depends on:
//	- TestCreateProject()
//	- TestRegisterUser() (in authentication.go)
func TestAddAuthors(t *testing.T) {
	var err error
	testProject := testProjects[0]

	// test to add a single valid author
	testSingleValidAuthor := func(author *Credentials) error {
		// initializes the test environment, returning an error if any occurs
		if err := initTestEnvironment(); err != nil {
			return errors.New(fmt.Sprintf("Error in testdb init: %v", err))
		}

		// declares test variables
		var queriedProjectId int   // gotten from db after adding author
		var queriedAuthorId string // gotten from db after adding author

		// adds a valid project and user to the db and filesystem so that an author can be added
		projectId, err := addProject(testProject)
		if err != nil {
			return err
		}
		authorId, err := registerUser(author)
		if err != nil {
			return errors.New(fmt.Sprintf("Error in author registration: %v", err))
		}

		// adds the author to the database
		if err = addAuthor(authorId, projectId); err != nil {
			return errors.New(fmt.Sprintf("Error adding the author to the db: %v", err))
		}

		// checks the author ID and project ID for matches
		queryAuthor := fmt.Sprintf(SELECT_ROW, "*", TABLE_AUTHORS, "userId")
		// executes query
		row := db.QueryRow(queryAuthor, authorId)
		if row.Err() != nil {
			return errors.New(
				fmt.Sprintf("error while querying db for authors: %v", row.Err()))
		} else if err = row.Scan(&queriedProjectId, &queriedAuthorId); err != nil {
			return errors.New(
				fmt.Sprintf("error while querying db for authors: %v", row.Err()))
		}

		// checks data returned from the database
		if projectId != queriedProjectId {
			return errors.New(
				fmt.Sprintf("Author added to the wrong project: Wanted: %d Got: %d",
					projectId, queriedProjectId))
		} else if authorId != queriedAuthorId {
			return errors.New(
				fmt.Sprintf("Author Ids do not match: Added: %s Gotten Back: %s",
					authorId, queriedAuthorId))
		}

		// clears the test environment and returns nil because the test has passed
		if err = clearTestEnvironment(); err != nil {
			return errors.New(fmt.Sprintf("error while tearing down db: %v", err))
		}
		return nil
	}

	// attemps to add an author without the correct permissions, if addAuthor succeeds, an error is thrown
	testAddInvalidAuthor := func(author *Credentials) error {
		// initializes the test environment, returning an error if any occurs
		if err := initTestEnvironment(); err != nil {
			return errors.New(fmt.Sprintf("Error in test environment init: %v", err))
		}

		// declares test variables
		var projectId int
		var authorId string

		// adds a valid project and user to the db and filesystem so that an author can be added
		projectId, err := addProject(testProject)
		if err != nil {
			return err
		}
		authorId, err = registerUser(author)
		if err != nil {
			return errors.New(fmt.Sprintf("Error registering author: %v", err))
		}

		// if adding the author is successful, throw an error
		if err = addAuthor(authorId, projectId); err == nil {
			return errors.New("Incorrect permissions added to authors table.")
		}

		// clears the test environment and returns nil because the test has passed
		if err = clearTestEnvironment(); err != nil {
			return errors.New(fmt.Sprintf("Error on db teardown: %v", err))
		}
		return nil
	}

	// tests that a user must be registered with the db before being and author
	testAddNonUserAuthor := func() error {
		// initializes the test environment, returning an error if any occurs
		if err := initTestEnvironment(); err != nil {
			return errors.New(fmt.Sprintf("Error in test environment init: %v", err))
		}

		// declares test variables
		var projectId int
		authorId := "u881jafjka" // non-user fake id

		// adds a valid project and user to the db and filesystem so that an author can be added
		projectId, err := addProject(testProject)
		if err != nil {
			return err
		}
		// if adding the author is successful, throw an error
		if err = addAuthor(authorId, projectId); err == nil {
			return errors.New("Added unregistered user id as author.")
		}

		// clears the test environment and returns nil because the test has passed
		if err = clearTestEnvironment(); err != nil {
			return errors.New(fmt.Sprintf("Error on db teardown: %v", err))
		}
		return nil
	}

	// runs tests
	if err := testSingleValidAuthor(testAuthors[0]); err != nil {
		t.Errorf("Failure on testAuthors[0]: %v", err)
	} else if err = testSingleValidAuthor(testAuthors[3]); err != nil {
		t.Errorf("Failure on testAuthors[3]: %v", err)
	} else if err = testAddInvalidAuthor(testAuthors[1]); err != nil {
		t.Errorf("Added invalid author testAuthors[1]: %v", err)
	} else if err = testAddNonUserAuthor(); err != nil {
		t.Errorf("Added non-user author: %v", err)
	}

	// tears down the test environment (makes sure that if a test fails, the env is still cleared)
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("error while tearing down db: %v", err)
	}
}

/*
 This function tests adding reviewers to a given project. Note that this test uses the
 addProject functionality, and as such will fail if it fails

 Test Depends on:
	- TestCreateProject()
	- TestRegisterUser() (in authentication.go)
*/
func TestAddReviewers(t *testing.T) {
	var err error
	testProject := testProjects[0]

	// test to add a single valid reviewer
	testSingleValidReviewer := func(reviewer *Credentials) error {
		// initializes the test environment, returning an error if any occurs
		if err = initTestEnvironment(); err != nil {
			return errors.New(fmt.Sprintf("error while initializing the test environment db: %v", err))
		}

		// declares test variables
		var projectId int
		var reviewerId string
		var queriedProjectId int     // gotten from db after adding reviewer
		var queriedReviewerId string // gotten from db after adding reviewer

		// adds a valid project and user to the db and filesystem so that an reviewer can be added
		projectId, err = addProject(testProject)
		if err != nil {
			return err
		}
		reviewerId, err = registerUser(reviewer)
		if err != nil {
			return errors.New(fmt.Sprintf("error registering the reviewer: %v", err))
		}

		// adds the reviewer to the database
		if err = addReviewer(reviewerId, projectId); err != nil {
			return errors.New(fmt.Sprintf("error adding the reviewer to the db: %v", err))
		}

		// checks the reviewer ID and project ID for matches
		queryReviewers := fmt.Sprintf(
			SELECT_ROW,
			"*",
			TABLE_REVIEWERS,
			"userId",
		)
		// executes query
		row := db.QueryRow(queryReviewers, reviewerId)
		if row.Err() != nil {
			return errors.New(fmt.Sprintf("Error on reviewer query: %v", row.Err()))
		} else if err = row.Scan(&queriedProjectId, &queriedReviewerId); err != nil {
			return errors.New(fmt.Sprintf("Error on reviewer query: %v", row.Err()))
		}
		// checks data returned from the database
		if projectId != queriedProjectId {
			return errors.New(fmt.Sprintf("Reviewer added to wrong project: %d vs %d", projectId, queriedProjectId))
		} else if reviewerId != queriedReviewerId {
			return errors.New(fmt.Sprintf("Reviewer ID mismatch: %s vs %s", reviewerId, queriedReviewerId))
		}

		// clears the test environment and returns nil because the test has passed
		if err = clearTestEnvironment(); err != nil {
			return errors.New(fmt.Sprintf("error while tearing down db: %v", err))
		}
		return nil
	}

	// attemps to add an reviewer without the correct permissions, if addReviewer succeeds, an error is thrown
	testAddInvalidReviewer := func(reviewer *Credentials) error {
		// initializes the test environment, returning an error if any occurs
		if err = initTestEnvironment(); err != nil {
			return errors.New(fmt.Sprintf("error while initializing the test environment db: %v", err))
		}

		// declares test variables
		var projectId int
		var reviewerId string

		// adds a valid project and user to the db and filesystem so that an reviewer can be added
		projectId, err = addProject(testProject)
		if err != nil {
			return err
		}
		reviewerId, err = registerUser(reviewer)
		if err != nil {
			return errors.New(fmt.Sprintf("error registering the reviewer: %v", err))
		}

		// if adding the reviewer is successful, throw an error
		if err = addReviewer(reviewerId, projectId); err == nil {
			return errors.New("reviewer with permissions incorrect permissions added to reviewers table")
		}

		// clears the test environment and returns nil because the test has passed
		if err = clearTestEnvironment(); err != nil {
			return errors.New(fmt.Sprintf("error while tearing down db: %v", err))
		}
		return nil
	}

	// tests that a user must be registered with the db before being and reviewer
	testAddNonUserReviewer := func() error {
		// initializes the test environment, returning an error if any occurs
		if err := initTestEnvironment(); err != nil {
			return errors.New(fmt.Sprintf("Error in test environment init: %v", err))
		}

		// declares test variables
		var projectId int
		authorId := "u881jafjka" // non-user fake id

		// adds a valid project and user to the db and filesystem so that an author can be added
		projectId, err := addProject(testProject)
		if err != nil {
			return err
		}
		// if adding the author is successful, throw an error
		if err = addReviewer(authorId, projectId); err == nil {
			return errors.New("Added unregistered user id as reviewer.")
		}

		// clears the test environment and returns nil because the test has passed
		if err = clearTestEnvironment(); err != nil {
			return errors.New(fmt.Sprintf("Error on db teardown: %v", err))
		}
		return nil
	}

	// runs tests
	if err = testSingleValidReviewer(testReviewers[0]); err != nil {
		t.Errorf("testSingleValidAuthor failed for testAuthors[0]: %v", err)
	} else if err = testSingleValidReviewer(testReviewers[3]); err != nil {
		t.Errorf("testSingleValidAuthor failed for testAuthors[3]: %v", err)
	} else if err = testAddInvalidReviewer(testReviewers[1]); err != nil {
		t.Errorf("testAddInvalidAuthor failed for testAuthors[1]: %v", err)
	} else if err = testAddNonUserReviewer(); err != nil {
		t.Errorf("Added non-user reviewer: %v", err)
	}

	// tears down the test environment (makes sure that if a test fails, the env is still cleared)
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("error while tearing down db: %v", err)
	}
}

/*
 This function tests adding authors to a given project. Note that this test uses the
 addProject functionality, and as such will fail if it fails

 Test Depends on:
	- TestCreateProject()
*/
func TestAddFiles(t *testing.T) {
	var err error
	testProject := testProjects[0] // test project to add files to

	// test function to add a single file. This function is not called directly as a test, but is a utility method for other tests
	testAddSingleFile := func(file *File, projectId int) error {
		// instantiates test variables
		var projectName string        // name of the project as queried from the SQL db
		var fileId int                // id of the file as returned from addFileTo()
		var queriedFileContent string // the content of the file
		var queriedProjectId int      // the id of the project as gotten from the files table
		var queriedFilePath string    // the file path as queried from the files table

		// adds file to the already instantiated project
		fileId, err := addFileTo(file, projectId)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to add file to the given project"))
		}

		// gets the project name from the db
		queryProjectName := fmt.Sprintf(
			SELECT_ROW,
			getDbTag(&Project{}, "Name"),
			TABLE_PROJECTS,
			getDbTag(&Project{}, "Id"),
		)
		// executes the query
		row := db.QueryRow(queryProjectName, projectId)
		if row.Err() != nil {
			return errors.New(
				fmt.Sprintf("Query failure on project name: %v", row.Err()))
		} else if err = row.Scan(&projectName); err != nil {
			return errors.New(fmt.Sprintf("Query failure on project name: %v", err))
		}

		// gets the file data from the db
		queryFileData := fmt.Sprintf(
			SELECT_ROW,
			fmt.Sprintf("%s, %s", getDbTag(&File{}, "ProjectId"), getDbTag(&File{}, "Path")),
			TABLE_FILES,
			getDbTag(&File{}, "Id"),
		)
		// executes query
		row = db.QueryRow(queryFileData, fileId)
		if row.Err() != nil {
			return errors.New(
				fmt.Sprintf("Failed query for project name : %v", row.Err()))
		} else if err = row.Scan(&queriedProjectId, &queriedFilePath); err != nil {
			return errors.New(
				fmt.Sprintf("Failed to query project name after db: %v", err))
		}

		// gets the file content from the filesystem
		filePath := filepath.Join(TEST_FILES_DIR, fmt.Sprint(projectId), projectName, queriedFilePath)
		fileBytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			return errors.New(
				fmt.Sprintf("File read failure after added to filesystem: %v", err))
		}
		queriedFileContent = string(fileBytes)

		// checks that a data file has been generated for the uploaded file
		fileDataPath := filepath.Join(
			TEST_FILES_DIR,
			fmt.Sprint(projectId),
			DATA_DIR_NAME,
			projectName,
			strings.TrimSuffix(queriedFilePath, filepath.Ext(queriedFilePath))+".json",
		)
		_, err = os.Stat(fileDataPath)
		if err != nil && errors.Is(err, os.ErrNotExist) {
			return errors.New("Data file not generated during file upload")
		} else if projectId != queriedProjectId { // Compare  test values.
			return errors.New(fmt.Sprintf("Project ID mismatch: %d vs %d",
				projectId, queriedProjectId))
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
	testAddSingleValidFile := func(file *File) error {
		if err = initTestEnvironment(); err != nil {
			return errors.New(fmt.Sprintf("error while initializing the test environment db: %v", err))
		}
		projectId, err := addProject(testProject)
		if err != nil {
			return err
		} else if err = testAddSingleFile(file, projectId); err != nil {
			return err
		} else if err = clearTestEnvironment(); err != nil {
			return errors.New(
				fmt.Sprintf("error while tearing down test environment: %v", err))
		}
		return nil
	}

	testAddNValidFiles := func(files []*File) error {
		if err = initTestEnvironment(); err != nil {
			return errors.New(fmt.Sprintf("error while initializing the test environment db: %v", err))
		}

		projectId, err := addProject(testProject)
		if err != nil {
			return err
		}
		// Test adding file for every file in array.
		for _, file := range files {
			if err = testAddSingleFile(file, projectId); err != nil {
				return err
			}
		}
		if err = clearTestEnvironment(); err != nil {
			return errors.New(fmt.Sprintf("error while tearing down test environment: %v", err))
		}
		return nil
	}

	// runs tests
	if err = testAddSingleValidFile(testFiles[0]); err != nil {
		t.Errorf("testAddSingleValidFile failed for testFiles[0]: %v", err)
	} else if err = testAddNValidFiles(testFiles[0:2]); err != nil {
		t.Errorf("testAddNValidFiles failed for testFiles[0:2]")
	}

	// tears down the test environment (makes sure that if a test fails, the env is still cleared)
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("error while tearing down db: %v", err)
	}
}

/*
function to test that the CodeFiles.go module can add comments to code files

Test Depends on:
	- TestAddFiles()
	- TestCreateProjects()
*/
func TestAddComment(t *testing.T) {
	var err error
	testProject := testProjects[0] // test project to add testFile to
	testFile := testFiles[0]       // test file to add comments to
	testAuthor := testAuthors[0]   // test author of comment

	testAddComment := func(comment *Comment, fileId int) error {
		// adds a comment to the file
		if err = addComment(comment, fileId); err != nil {
			return errors.New(fmt.Sprintf("failed to add comment to the project: %v", err))
		}

		// reads the data file into a CodeDataFile struct
		fileDataPath := filepath.Join(
			TEST_FILES_DIR,
			fmt.Sprint(testProject.Id),
			DATA_DIR_NAME,
			testProject.Name,
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

	testAddSingleValidComment := func(comment *Comment) error {
		// initializes the test environment, returning an error if any occurs
		if err = initTestEnvironment(); err != nil {
			return errors.New(fmt.Sprintf("error while initializing the test environment db: %v", err))
		}

		// creates a project, adds a file to it, and adds a test user to the system
		projectId, err := addProject(testProject)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to add project: %v", err))
		}
		fileId, err := addFileTo(testFile, projectId)
		if err != nil {
			return errors.New(
				fmt.Sprintf("failed to add a file to the project: %v", err))
		}
		authorId, err := registerUser(testAuthor)
		if err != nil {
			return errors.New(
				fmt.Sprintf("failed to add user to the database: %v", err))
		}
		testProject.Id = projectId
		testFile.Id = fileId
		comment.AuthorId = authorId

		// adds a comment to the file and tests that it was added properly
		if err = testAddComment(comment, fileId); err != nil {
			return err
		}

		// clears the test environment and returns nil because the test has passed
		if err = clearTestEnvironment(); err != nil {
			return errors.New(
				fmt.Sprintf("error while tearing down test environment: %v", err))
		}
		return nil
	}

	// runs tests
	if err = testAddSingleValidComment(testComments[0]); err != nil {
		t.Errorf("testAddSingleValidComment failed for testComments[0]: %v", err)
	}

	// tears down the test environment (makes sure that if a test fails, the env is still cleared)
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("error while tearing down db: %v", err)
	}
}

/*
Tests the ability of the getAllProjects() function to get all projects from the db at once

Test Depends On:
	- TestCreateProjects()
*/
func TestGetAllProjects(t *testing.T) {
	var err error

	// Set up server to listen with the getFile() function.
	srv := setupCORSsrv()

	// Start server.
	go srv.ListenAndServe()

	/*
		test for basic functionality. Adds 2 projects to the db, then queries them and tests for equality
	*/
	testGetTwoProjects := func() {
		var projectId int                    // variable to temporarily store project ids as they are added to the db
		sentProjects := make(map[int]string) // variable to hold the id: project name mappings which are sent to the db

		// sets up the test environment (db and filesystem)
		if err = initTestEnvironment(); err != nil {
			t.Errorf("Error initializing the test environment %s", err)
		}
		// uses a slice here so that we can add more projects to testProjects without breaking the test
		for _, proj := range testProjects[0:2] {
			projectId, err = addProject(proj)
			if err != nil {
				t.Errorf("Error adding project %s: %v", proj.Name, err)
			}
			// saves the added project with its id
			sentProjects[projectId] = proj.Name
		}

		// builds and sends and http get request
		req, err := http.NewRequest("GET", fmt.Sprintf("%s:%s/projects", TEST_URL, TEST_SERVER_PORT), nil)
		resp, err := sendSecureRequest(req, TEAM_ID)
		if err != nil {
			t.Errorf("Error occurred while sending get request to the Go server: %v", err)
		}
		defer resp.Body.Close()
		if err != nil {
			t.Error(err)
		}

		// checks the returned list of projects for equality with the sent list
		returnedProjects := make(map[int]string)
		json.NewDecoder(resp.Body).Decode(&returnedProjects)

		// tests that the proper values have been returned
		for k, v := range returnedProjects {
			if v != sentProjects[k] {
				t.Errorf("Projects of ids: %d do not have matching names. Given: %s, Returned: %s ", k, sentProjects[k], v)
			}
		}

		// destroys the test environment
		if err = clearTestEnvironment(); err != nil {
			t.Errorf("Error occurred while destroying the database and filesystem: %v", err)
		}
	}

	// runs tests
	testGetTwoProjects()

	// Close server.
	if err = srv.Shutdown(context.Background()); err != nil {
		t.Errorf("HTTP server shutdown: %v", err)
	}

	// tears down the test environment (makes sure that if a test fails, the env is still cleared)
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("error while tearing down db: %v", err)
	}
}

/*
Tests the getUserProjects() function to get

Test Depends On:
	- TestCreateProjects()
	- TestAddReviewers()
	- TestAddAuthors()
*/
func TestGetUserProjects(t *testing.T) {
	var err error
	testProject1 := testProjects[0] // test project to return on getUserProjects()
	testProject2 := testProjects[1] // test project to not return on getUserProjects()
	testAuthor := testAuthors[0]    // test author of the project being queried
	testNonAuthor := testAuthors[3] // test author of project not being queried

	/*
		test for basic functionality. Adds 2 projects to the db with different authors, then queries them and tests for equality
	*/
	testGetSingleProject := func() {
		// sets up the test environment (db and filesystem)
		if err = initTestEnvironment(); err != nil {
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

		// adds two test projects to the db
		testProject1.Id, err = addProject(testProject1)
		if err != nil {
			t.Errorf("Error occurred while adding project1: %v", err)
		}
		testProject2.Id, err = addProject(testProject2)
		if err != nil {
			t.Errorf("Error occurred while adding project2: %v", err)
		}

		// adds authors to the test projects
		if err = addAuthor(authorId, testProject1.Id); err != nil {
			t.Errorf("Failed to add author")
		}
		if err = addAuthor(nonAuthorId, testProject2.Id); err != nil {
			t.Errorf("Failed to add author")
		}

		// queries all of testAuthor's projects
		projects, err := getUserProjects(authorId)
		if err != nil {
			t.Errorf("Error getting user projects: %v", err)
		}

		// tests for equality of project Id and that testProject2.Id is not in the map
		if _, ok := projects[testProject2.Id]; ok {
			t.Errorf("Returned project where the test author is not an author")
		} else if projects[testProject1.Id] != testProject1.Name {
			t.Errorf("Returned incorrect project name: %s", projects[testProject1.Id])
		}

		// destroys the test environment
		if err = clearTestEnvironment(); err != nil {
			t.Errorf("Error occurred while destroying the database and filesystem: %v", err)
		}
	}

	// runs tests
	testGetSingleProject()

	// tears down the test environment (makes sure that if a test fails, the env is still cleared)
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("error while tearing down db: %v", err)
	}
}

/*
Tests the ability of the CodeFiles module to get a project from the db

Test Depends On:
	- TestCreateProjects()
	- TestAddFiles()
	- TestAddReviewers()
	- TestAddAuthors()
*/
func TestGetProject(t *testing.T) {
	var err error

	// Set up server to listen with the getFile() function.
	srv := setupCORSsrv()

	// Start server.
	go srv.ListenAndServe()

	/*
		Tests the basic ability of the CodeFiles module to load a project from the
		db and filesystem
	*/
	testGetValidProject := func() {
		var projectId int // holds the project id as returned from the addProject() function

		testFile := testFiles[0]         // defines the file to use for the test here so that it can be easily changed
		testProject := testProjects[0]   // defines the project to use for the test here so that it can be easily changed
		testAuthor := testAuthors[0]     // defines the author of the project
		testReviewer := testReviewers[0] // defines the reviewer of the project

		// initializes the filesystem and db
		if err = initTestEnvironment(); err != nil {
			t.Errorf("Error initializing the test environment: %v", err)
		}
		// adds the test project to the filesystem and database
		projectId, err = addProject(testProject)
		if err != nil {
			t.Errorf("Error adding project %v", err)
		}
		// adds the test file to the filesystem and database
		_, err = addFileTo(testFile, projectId)
		if err != nil {
			t.Errorf("Error adding file to the project %v", err)
		}
		// adds an author and reviewer to the project
		authorId, err := registerUser(testAuthor)
		if err != nil {
			t.Errorf("Error registering author in the db: %v", err)
		}
		reviewerId, err := registerUser(testReviewer)
		if err != nil {
			t.Errorf("Error registering reviewer in the db: %v", err)
		}

		// adds reviewer and author to the project
		if err = addAuthor(authorId, projectId); err != nil {
			t.Errorf("Error adding author to the project: %v", err)
		}
		if err = addReviewer(reviewerId, projectId); err != nil {
			t.Errorf("Error adding reviewer to the project: %v", err)
		}

		// creates a request to get a project of a given id
		reqBody, err := json.Marshal(map[string]interface{}{
			getJsonTag(&Project{}, "Id"): projectId,
		})
		if err != nil {
			t.Errorf("Error Retrieving Project: %v", err)
		}
		// sets a custom header of "project":id to query the specific project id
		req, err := http.NewRequest("POST", fmt.Sprintf("%s:%s%s", TEST_URL, TEST_SERVER_PORT, ENDPOINT_PROJECT), bytes.NewBuffer(reqBody))
		resp, err := sendSecureRequest(req, TEAM_ID)
		if err != nil {
			t.Errorf("Error while sending Get request: %v", err)
		}
		defer resp.Body.Close()

		// if an error occurred while getting the file, it is printed out here
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Error: %d", resp.StatusCode)
		}

		// marshals the json response into a Project struct
		project := &Project{}
		err = json.NewDecoder(resp.Body).Decode(&project)
		if err != nil {
			t.Error("Error while decoding server response: ", err)
		}

		// tests that the project matches the passed in data
		if testProject.Id != project.Id {
			t.Errorf("Project IDs do not match. Given: %d != Returned: %d", testProject.Id, project.Id)
		} else if testProject.Name != project.Name {
			t.Errorf("Project Names do not match. Given: %s != Returned: %s", testProject.Name, project.Name)
			// tests that file paths match (done directly here as there is only one constituent file)
		} else if testProject.FilePaths[0] != project.FilePaths[0] {
			t.Errorf("Project file path lists do not match. Given: %s != Returned: %s", testProject.FilePaths[0], project.FilePaths[0])
			// tests that the authors lists match (done directly here as there is only one author)
		} else if authorId != project.Authors[0] {
			t.Errorf("Authors do not match. Expected: %s Given: %s", authorId, testProject.Authors[0])
			// tests that the reviewer lists match (done directly here as there is only one reviewer)
		} else if reviewerId != project.Reviewers[0] {
			t.Errorf("Authors do not match. Expected: %s Given: %s", reviewerId, testProject.Reviewers[0])
		}

		// destroys the filesystem and clears the db
		if err = clearTestEnvironment(); err != nil {
			t.Errorf("Error occurred while destroying the database and filesystem: %v", err)
		}
	}

	// runs tests
	testGetValidProject()

	// Close server.
	if err = srv.Shutdown(context.Background()); err != nil {
		fmt.Printf("HTTP server shutdown: %v", err)
	}

	// tears down the test environment (makes sure that if a test fails, the env is still cleared)
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("error while tearing down db: %v", err)
	}
}

/*
function to test querying files

Test Depends On:
	- TestCreateProject()
	- TestAddFiles()
*/
func TestGetFile(t *testing.T) {
	var err error

	// Set up server to listen with the getFile() function.
	srv := setupCORSsrv()

	// Start server.
	go srv.ListenAndServe()

	// Tests the basic ability of the CodeFiles module to load the data from a
	// valid file path passed to it. Simple valid one code file project
	testGetSingleFile := func() {
		var projectId int              // stores project id returned by addProject()
		testFile := testFiles[0]       // the test file to be added to the db and filesystem (saved here so it can be easily changed)
		testProject := testProjects[0] // the test project to be added to the db and filesystem (saved here so it can be easily changed)

		// initializes the filesystem and db
		if err = initTestEnvironment(); err != nil {
			t.Errorf("Error initializing the test environment %s", err)
		}
		// adds a project to the database and filesystem
		projectId, err = addProject(testProject)
		if err != nil {
			t.Errorf("Error adding project %s: %v", testProject.Name, err)
		}
		// adds a file to the database and filesystem
		_, err = addFileTo(testFile, projectId)
		if err != nil {
			t.Errorf("Error adding file %s: %v", testFile.Name, err)
		}
		// sets the project id of the added file to link it with the project on this end (just in case. This should happen in addFileTo)
		testFile.ProjectId = projectId

		// sets a custom header "file": file path and "project": projectId to indicate which file is being queried to the server
		reqBody, err := json.Marshal(map[string]interface{}{
			getJsonTag(&File{}, "Path"):      testFile.Path,
			getJsonTag(&File{}, "ProjectId"): testFile.ProjectId,
		})
		if err != nil {
			t.Errorf("Error formatting request body: %v", err)
		}
		req, err := http.NewRequest("POST", fmt.Sprintf("%s:%s%s", TEST_URL, TEST_SERVER_PORT, ENDPOINT_FILE), bytes.NewBuffer(reqBody))
		if err != nil {
			t.Errorf("Error creating request: %v\n", err)
		}
		// send POST request
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
			// tests for project id correctness
		} else if testFile.ProjectId != file.ProjectId {
			t.Errorf("File Project Id %d != %d", file.ProjectId, testFile.ProjectId)
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
	}

	// runs tests
	testGetSingleFile()

	// Close server.
	if err = srv.Shutdown(context.Background()); err != nil {
		fmt.Printf("HTTP server shutdown: %v", err)
	}

	// tears down the test environment (makes sure that if a test fails, the env is still cleared)
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("error while tearing down db: %v", err)
	}
}

/*
tests uploading single files via HTTP

Test Depends on:
	- TestCreateProject()
	- TestAddFile()
*/
func TestUploadSingleFile(t *testing.T) {
	var err error

	// the test values added to the db and filesystem (saved here so it can be easily changed)
	testFile := testFiles[0]
	testAuthor := testAuthors[0]

	// Set up server to listen with the getFile() function.
	srv := setupCORSsrv()

	// Start server.
	go srv.ListenAndServe()

	// Tests the basic ability of the CodeFiles module to upload a single file
	// code project
	testAddSingleFile := func() {
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
	}

	// runs tests
	testAddSingleFile()

	// Close server.
	if err = srv.Shutdown(context.Background()); err != nil {
		fmt.Printf("HTTP server shutdown: %v", err)
	}

	// tears down the test environment (makes sure that if a test fails, the env is still cleared)
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("error while tearing down db: %v", err)
	}
}

/*
Tests router function to upload comments to a given file

Test Depends On:
	- TestAddComment()
	- TestCreateProject()
	- TestAddFiles()
*/
func TestUploadUserComment(t *testing.T) {
	var err error

	// the test values added to the db and filesystem (saved here so it can be easily changed)
	testFile := testFiles[0]
	testProject := testProjects[0]
	testAuthor := testAuthors[0]
	testComment := testComments[0]

	// Set up server to listen with the getFile() function.
	srv := setupCORSsrv()

	// Start server.
	go srv.ListenAndServe()

	// adds a comment to a test project

	// Tests the basic ability of the CodeFiles module to add a comment to a file
	// given file path and project id
	testAddSingleComment := func() {
		var projectId int // stores project id returned by addProject()

		// initializes the filesystem and db
		if err = initTestEnvironment(); err != nil {
			t.Errorf("Error initializing the test environment %s", err)
		}

		// adds test values to the db and filesystem
		projectId, err = addProject(testProject)
		if err != nil {
			t.Errorf("error occurred while adding testProject: %v", err)
		}
		_, err = addFileTo(testFile, projectId)
		if err != nil {
			t.Errorf("error occurred while adding testProject: %v", err)
		}
		testAuthor.Id, err = registerUser(testAuthor)
		if err != nil {
			t.Errorf("error occurred while adding testAuthor: %v", err)
		}
		testComment.AuthorId = testAuthor.Id // sets test comment author

		// formats the request body to send to the server to add a comment
		reqBody, err := json.Marshal(map[string]interface{}{
			getJsonTag(&File{}, "ProjectId"):   projectId,
			getJsonTag(&File{}, "Path"):        testFile.Path,
			getJsonTag(&Comment{}, "AuthorId"): testAuthor.Id,
			getJsonTag(&Comment{}, "Content"):  testComment.Content,
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
			fmt.Sprint(testProject.Id),
			DATA_DIR_NAME,
			testProject.Name,
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
	}

	// runs tests
	testAddSingleComment()

	// Close server.
	if err = srv.Shutdown(context.Background()); err != nil {
		fmt.Printf("HTTP server shutdown: %v", err)
	}

	// tears down the test environment (makes sure that if a test fails, the env is still cleared)
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("error while tearing down db: %v", err)
	}
}
