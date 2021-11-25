/*
submissions_test.go
author: 190010425
created: November 18, 2021

This file takes care of 
*/

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"io/ioutil"
	"path/filepath"
)

// data to use in the tests
var testProjects []*Project = []*Project{
	{Id: -1, Name: "testProject1", Reviewers: []string{},
		Authors: []string{}, FilePaths: []string{"testFile1.txt"}, MetaData: testProjectMetaData[0]},
	{Id: -1, Name: "testProject2", Reviewers: []string{},
		Authors: []string{}, FilePaths: []string{"testFile2.txt"}, MetaData: testProjectMetaData[0]},
}
var testProjectMetaData = []*CodeProjectData{
	{Abstract: "test abstract, this means nothing", Reviews:nil}, // TODO: add comments here
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

// Utility function to be re-used for testing adding projects to the db
func testAddProject(testProject *Project) error {
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

	// checks that the filesystem has a proper corresponding entry and metadata file
	projectDirPath := filepath.Join(TEST_FILES_DIR, fmt.Sprint(projectId))
	fileDataPath := filepath.Join(projectDirPath, DATA_DIR_NAME, projectName + ".json")
	dataString, err := ioutil.ReadFile(fileDataPath)
	if err != nil {
		return err
	}
	// marshalls the string of data into a struct
	projectData := &CodeProjectData{}
	if err := json.Unmarshal(dataString, projectData); err != nil {
		return err
	}
	// tests that the metadata is properly formatted
	if projectData.Abstract != testProject.MetaData.Abstract {
		return errors.New(fmt.Sprintf(
			"project metadata not added to filesystem properly. Abstracts %s, %s do not match",
			projectData.Abstract, testProject.MetaData.Abstract))
	} else if !(reflect.DeepEqual(projectData.Reviews, testProject.MetaData.Reviews)) {
		return errors.New("Project Reviews do not match")
	}
	return nil
}

// tests that a single valid project can be added to the db and filesystem properly
func TestAddOneProject(t *testing.T) {
	project := testProjects[0]

	// initializes the test environment, returning an error if any occurs
	if err := initTestEnvironment(); err != nil {
		t.Errorf("Error in db init: %v", err)
	}
	// adds the project and tests that it was added properly
	if err := testAddProject(project); err != nil {
		t.Errorf("%v", err) // error already formatted here
	}
	// tears down the test environment
	if err := clearTestEnvironment(); err != nil {
		t.Errorf("Error in db teardown: %v", err)
	}
}

// tests that multiple projects can be added in a row properly
func TestAddMultipleProjects(t *testing.T) {
	projects := testProjects[0:2] // list of projects to add to the db

	// initializes the test environment, returning an error if any occurs
	if err := initTestEnvironment(); err != nil {
		t.Errorf("Error in db init: %v", err)
	}
	// adds a range all projects in the projects slice
	for _, project := range projects {
		// adds the project and tests that it was added properly
		if err := testAddProject(project); err != nil {
			t.Errorf("%v", err)
		}
	}
	// tears down the test environment
	if err := clearTestEnvironment(); err != nil {
		t.Errorf("Error in db teardown: %v", err)
	}
}

// tests that trying to add a nil project to the db and filesystem will return an error
func TestAddNilProject(t *testing.T) {
	// initializes the test environment
	if err := initTestEnvironment(); err != nil {
		t.Errorf("Error in db init: %v", err)
	}
	// tries to add a nil project
	if _, err := addProject(nil); err == nil {
		t.Error("Nil project added to the db without error")
	}
	// tears down the test environment
	if err := clearTestEnvironment(); err != nil {
		t.Errorf("Error in db teardown: %v", err)
	}
}

// utility function which tests that an author can be added to a valid project properly
// this test depends on the add projects tests
func testAddAuthor(projectId int, author *Credentials) error {
	// declares test variables
	var queriedProjectId int   // gotten from db after adding author
	var queriedAuthorId string // gotten from db after adding author

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
	row := db.QueryRow(queryAuthor, authorId)
	if err := row.Scan(&queriedProjectId, &queriedAuthorId); err != nil {
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
	return nil
}

func TestAddOneAuthor(t *testing.T) {
	testProject := testProjects[0]
	testAuthor := testAuthors[0]

	// initializes the test environment
	if err := initTestEnvironment(); err != nil {
		t.Errorf("Error in db init: %v", err)
	}
	// adds a valid project and user to the db and filesystem so that an author can be added
	projectId, err := addProject(testProject)
	if err != nil {
		t.Errorf("Error occurred while adding test project: %v", err)
	}
	// adds the author to the db and filesystem
	if err := testAddAuthor(projectId, testAuthor); err != nil {
		t.Errorf("Error occurred while adding test author: %v", err)
	}
	// tears down the test environment
	if err := clearTestEnvironment(); err != nil {
		t.Errorf("Error in db teardown: %v", err)
	}
}

// attemps to add an author without the correct permissions, if addAuthor succeeds, an error is thrown
func TestAddInvalidAuthor(t *testing.T) {
	testProject := testProjects[0]
	testAuthor := testAuthors[1] // user without publisher permissions

	// initializes the test environment
	if err := initTestEnvironment(); err != nil {
		t.Errorf("Error in db init: %v", err)
	}
	// adds a valid project and user to the db and filesystem so that an author can be added
	projectId, err := addProject(testProject)
	if err != nil {
		t.Errorf("Error adding test project: %v", err)
	}
	// if adding the author is successful, throw an error
	if err = testAddAuthor(projectId, testAuthor); err == nil {
		t.Error("Incorrect permissions added to authors table.")
	}
	// tears down the test environment
	if err := clearTestEnvironment(); err != nil {
		t.Errorf("Error in db teardown: %v", err)
	}
}

// tests that a user must be registered with the db before being and author
func TestAddNonUserAuthor(t *testing.T) {
	testProject := testProjects[0]
	authorId := "u881jafjka" // non-user fake id

	// initializes the test environment, returning an error if any occurs
	if err := initTestEnvironment(); err != nil {
		t.Errorf("Error in test environment init: %v", err)
	}
	// adds a valid project to the db and filesystem
	projectId, err := addProject(testProject)
	if err != nil {
		t.Errorf("Error while adding test project: %v", err)
	}
	// if adding the author is successful, throw an error
	if err = addAuthor(authorId, projectId); err == nil {
		t.Error("Added unregistered user id as author.")
	}
	// clears the test environment
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("Error on db teardown: %v", err)
	}
}

// utility function which tests that a reviewer can be added to a valid project properly
// this test depends on the add projects tests
func testAddReviewer(projectId int, reviewer *Credentials) error {
	var queriedProjectId int   // gotten from db after adding reviewer
	var queriedReviewerId string // gotten from db after adding reviewer

	reviewerId, err := registerUser(reviewer)
	if err != nil {
		return errors.New(fmt.Sprintf("Error in reviewer registration: %v", err))
	}

	// adds the reviewer to the database
	if err = addReviewer(reviewerId, projectId); err != nil {
		return errors.New(fmt.Sprintf("Error adding the reviewer to the db: %v", err))
	}

	// checks the reviewer ID and project ID for matches
	queryReviewer := fmt.Sprintf(SELECT_ROW, "*", TABLE_REVIEWERS, "userId")
	row := db.QueryRow(queryReviewer, reviewerId)
	if err := row.Scan(&queriedProjectId, &queriedReviewerId); err != nil {
		return errors.New(
			fmt.Sprintf("error while querying db for reviewers: %v", row.Err()))
	}

	// checks data returned from the database
	if projectId != queriedProjectId {
		return errors.New(
			fmt.Sprintf("Reviewer added to the wrong project: Wanted: %d Got: %d",
				projectId, queriedProjectId))
	} else if reviewerId != queriedReviewerId {
		return errors.New(
			fmt.Sprintf("Reviewer Ids do not match: Added: %s Gotten Back: %s",
				reviewerId, queriedReviewerId))
	}
	return nil
}

// tests that a single valid reviewer can be added to the database properly
func TestAddOneReviewer(t *testing.T) {
	testProject := testProjects[0]
	testReviewer := testReviewers[0]

	// initializes the test environment
	if err := initTestEnvironment(); err != nil {
		t.Errorf("Error in db init: %v", err)
	}
	// adds a valid project and user to the db and filesystem so that an author can be added
	projectId, err := addProject(testProject)
	if err != nil {
		t.Errorf("Error occurred while adding test project: %v", err)
	}
	// adds the reviewer to the db and filesystem
	if err := testAddReviewer(projectId, testReviewer); err != nil {
		t.Errorf("Error occurred while adding test reviewer: %v", err)
	}
	// tears down the test environment
	if err := clearTestEnvironment(); err != nil {
		t.Errorf("Error in db teardown: %v", err)
	}
}

// attemps to add a reviewere without the correct permissions, if addReviewer succeeds, an error is thrown
func TestAddInvalidReviewer(t *testing.T) {
	testProject := testProjects[0]
	testReviewer := testReviewers[1] // reviewer without reviewer permissions

	// initializes the test environment
	if err := initTestEnvironment(); err != nil {
		t.Errorf("Error in db init: %v", err)
	}
	// adds a valid project and user to the db and filesystem so that an reviewer can be added
	projectId, err := addProject(testProject)
	if err != nil {
		t.Errorf("Error adding test project: %v", err)
	}
	// if adding the reviewer is successful, throw an error
	if err = testAddReviewer(projectId, testReviewer); err == nil {
		t.Error("Incorrect permissions added to reviewers table.")
	}
	// tears down the test environment
	if err := clearTestEnvironment(); err != nil {
		t.Errorf("Error in db teardown: %v", err)
	}
}

// tests that a user must be registered with the db before being and author
func TestAddNonUserReviewer(t *testing.T) {
	testProject := testProjects[0]
	reviewerId := "u881jafjka" // non-user fake id

	// initializes the test environment, returning an error if any occurs
	if err := initTestEnvironment(); err != nil {
		t.Errorf("Error in test environment init: %v", err)
	}
	// adds a valid project to the db and filesystem
	projectId, err := addProject(testProject)
	if err != nil {
		t.Errorf("Error while adding test project: %v", err)
	}
	// if adding the author is successful, throw an error
	if err = addReviewer(reviewerId, projectId); err == nil {
		t.Error("Added unregistered user id as reviewer.")
	}
	// clears the test environment
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("Error on db teardown: %v", err)
	}
}

// This function tests the getProjectMetaData function
//
// This test depends on:
// 	- addProject()
func TestGetProjectMetaData(t *testing.T) {
	testProject := testProjects[0]

	// initializes the test environment, returning an error if any occurs
	if err := initTestEnvironment(); err != nil {
		t.Errorf("Error in test environment init: %v", err)
	}
	// adds the test project to the db
	projectId, err := addProject(testProject)
	if err != nil {
		t.Errorf("Error adding project to the db and filesystem: %v", err)
	}
	// tests that the metadata can be read back properly
	projectData, err := getProjectMetaData(projectId)
	if err != nil {
		t.Errorf("Error getting project metadata: %v", err)
	}
	// tests for equality of the added metadata with that which was retrieved
	if projectData.Abstract != testProject.MetaData.Abstract {
		t.Errorf("project metadata not added to filesystem properly. Abstracts %s, %s do not match",
			projectData.Abstract, testProject.MetaData.Abstract)
	} else if !(reflect.DeepEqual(projectData.Reviews, testProject.MetaData.Reviews)) {
		t.Error("Project Reviews do not match")
	}
	// clears the test environment
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("Error on db teardown: %v", err)
	}
}

// Tests that getProjectMetaData will throw an error if an incorrect project ID is passed in
func TestGetInvalidProjectMetaData(t *testing.T) {
	// initializes the test environment, returning an error if any occurs
	if err := initTestEnvironment(); err != nil {
		t.Errorf("Error in test environment init: %v", err)
	}
	// tests that an error is thrown if a non-existant project ID is passed to getProjectMetaData
	_, err := getProjectMetaData(400)
	if err == nil {
		t.Errorf("No error was thrown for invalid project")
	}
	// clears the test environment
	if err = clearTestEnvironment(); err != nil {
		t.Errorf("Error on db teardown: %v", err)
	}
}

// Tests the ability of the getAllProjects() function to get all projects from the db at once
//
// Test Depends On:
// 	- TestCreateProjects()
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

	
// test for basic functionality. Adds 2 projects to the db with different authors, then queries them and tests for equality
// Test Depends On:
// 	- TestCreateProjects()
// 	- TestAddReviewers()
// 	- TestAddAuthors()
func TestGetSingleProject(t *testing.T) {
	testProject1 := testProjects[0] // test project to return on getUserProjects()
	testProject2 := testProjects[1] // test project to not return on getUserProjects()
	testAuthor := testAuthors[0]    // test author of the project being queried
	testNonAuthor := testAuthors[3] // test author of project not being queried

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

// Tests the ability of the CodeFiles module to get a project from the db
// 
// Test Depends On:
// 	- TestCreateProjects()
// 	- TestAddFiles()
// 	- TestAddReviewers()
// 	- TestAddAuthors()
func TestGetProject(t *testing.T) {
	var err error
	var projectId int // holds the project id as returned from the addProject() function

	// Set up server to listen with the getFile() function.
	srv := setupCORSsrv()

	// Start server.
	go srv.ListenAndServe()

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

	// Close server.
	if err = srv.Shutdown(context.Background()); err != nil {
		fmt.Printf("HTTP server shutdown: %v", err)
	}
}