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
	"bytes"
	"context"
	"encoding/base64"
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
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	TEST_ZIP_PATH = "../testing/test.zip"
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
			Reviews:  []*Review{},
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
			Reviews:  []*Review{},
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
		globalAuthors[i].UserType = USERTYPE_PUBLISHER
	}
	globalReviewers := make([]GlobalUser, len(testReviewers))
	for i, user := range testReviewers {
		if globalReviewers[i].ID, err = registerUser(user, USERTYPE_REVIEWER); err != nil {
			t.Errorf("User registration failed: %v", err)
			return nil, nil, err
		}
		globalReviewers[i].UserType = USERTYPE_REVIEWER
	}
	return globalAuthors, globalReviewers, nil
}

// ------------
// Router Function Tests
// ------------

func TestGetAvailableTags(t *testing.T) {
	testInit()
	defer testEnd()

	// Create mux router
	router := mux.NewRouter()
	router.HandleFunc(SUBROUTE_SUBMISSIONS+ENDPOINT_GET_TAGS, GetAvailableTags)

	addTag := func(tag string) {
		assert.NoError(t, gormDb.Model(&Category{}).Create(&Category{Tag: tag}).Error, "failed to add tag")
	}
	clearTags := func() {
		gormDb.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&Category{})
	}
	handleQuery := func() (int, *GetAvailableTagsResponse) {
		req, w := httptest.NewRequest(http.MethodGet, SUBROUTE_SUBMISSIONS+ENDPOINT_GET_TAGS, nil), httptest.NewRecorder()
		router.ServeHTTP(w, req)
		resp := w.Result()

		respData := &GetAvailableTagsResponse{}
		if !assert.NoError(t, json.NewDecoder(resp.Body).Decode(respData), "error occurred while parsing response") {
			return 0, respData
		}
		return resp.StatusCode, respData
	}

	t.Run("No tags returned", func(t *testing.T) {
		defer clearTags()
		status, resp := handleQuery()
		switch {
		case !assert.False(t, resp.Error, "error field should be false in response"),
			!assert.Equal(t, http.StatusNoContent, status, "returned incorrect status code"),
			!assert.Equal(t, 0, len(resp.Tags), "returned non-empty tag array"):
			return
		}
	})

	t.Run("One tag", func(t *testing.T) {
		addTag("python")
		defer clearTags()
		status, resp := handleQuery()
		switch {
		case !assert.False(t, resp.Error, "error field should be false in response"),
			!assert.Equal(t, http.StatusOK, status, "returned incorrect status code"),
			!assert.ElementsMatch(t, []string{"python"}, resp.Tags, "returned incorrect tag array"):
			return
		}
	})

	t.Run("Many tags", func(t *testing.T) {
		defer clearTags()
		tags := []string{"python", "java", "c", "go", "javascript"}
		for _, tag := range tags {
			addTag(tag)
		}
		status, resp := handleQuery()
		switch {
		case !assert.False(t, resp.Error, "error field should be false in response"),
			!assert.Equal(t, http.StatusOK, status, "returned incorrect status code"),
			!assert.ElementsMatch(t, tags, resp.Tags, "returned incorrect tag array"):
			return
		}
	})
}

// Tests that GetQuerySubmissions works properly
func TestQuerySubmissions(t *testing.T) {
	var err error
	// Set up server and configures filesystem/db
	testInit()
	defer testEnd()

	// Create mux router
	router := mux.NewRouter()
	router.HandleFunc(SUBROUTE_SUBMISSIONS+ENDPOINT_QUERY_SUBMISSIONS, GetQuerySubmissions)

	globalAuthors, globalReviewers, err := initMockUsers(t)
	if err != nil {
		return
	}

	// adds a test submission to the db and filesystem
	addTestSubmission := func(name string, approved *bool, tags []string, authors []GlobalUser, reviewers []GlobalUser) uint {
		categories := []Category{}
		for _, tag := range tags {
			categories = append(categories, Category{Tag: tag})
		}
		submission := &Submission{
			Name:       name,
			License:    "MIT",
			Approved:   approved,
			Categories: categories,
			Authors:    authors,
			Reviewers:  reviewers,
		}
		submissionID, err := addSubmission(submission)
		if !assert.NoError(t, err, "Error while adding test submission") {
			return 0
		}
		return submissionID
	}

	// wipe the db and filesystem submission tables
	clearSubmissions := func() {
		// deletes submissions w/ associations
		var submissions []Submission
		if !assert.NoError(t, gormDb.Find(&submissions).Error) {
			return
		}
		for _, submission := range submissions {
			gormDb.Select(clause.Associations).Unscoped().Delete(&submission)
		}
	}

	t.Run("Valid Query", func(t *testing.T) {
		// handles sending the request and parsing the response
		handleQuery := func(queryRoute string) *QuerySubmissionsResponse {
			req, w := httptest.NewRequest(http.MethodGet, queryRoute, nil), httptest.NewRecorder()
			router.ServeHTTP(w, req)
			resp := w.Result()

			respData := &QuerySubmissionsResponse{}
			if !assert.NoError(t, json.NewDecoder(resp.Body).Decode(respData), "Error decoding request body") {
				return nil
			} else if !assert.Falsef(t, respData.StandardResponse.Error,
				"Error returned on query - %v", respData.StandardResponse.Message) {
				return nil
			}
			return respData
		}

		t.Run("no query parameters", func(t *testing.T) {
			defer clearSubmissions()
			submissionIDs := make([]uint, 2)
			submissionIDs[0] = addTestSubmission("test1", nil, []string{"python", "sorting"}, globalAuthors[:1], globalReviewers[:1])
			submissionIDs[1] = addTestSubmission("test2", nil, []string{"go", "sorting"}, globalAuthors[:1], globalReviewers[:1])

			// test that the response is as expected
			resp := handleQuery(SUBROUTE_SUBMISSIONS + ENDPOINT_QUERY_SUBMISSIONS)
			switch {
			case !assert.NotEmpty(t, resp, "request response is nil"),
				!assert.Contains(t, submissionIDs, resp.Submissions[0].ID, "Missing submission 1 ID"),
				!assert.Contains(t, submissionIDs, resp.Submissions[1].ID, "Missing submission 2 ID"):
				return
			}
		})

		t.Run("order by date", func(t *testing.T) {
			defer clearSubmissions()
			submissionIDs := make([]uint, 2)
			submissionIDs[0] = addTestSubmission("test1", nil, []string{"python", "sorting"}, globalAuthors[:1], globalReviewers[:1])
			submissionIDs[1] = addTestSubmission("test2", nil, []string{"go", "sorting"}, globalAuthors[:1], globalReviewers[:1])

			// test that the response is as expected
			queryRoute := fmt.Sprintf("%s%s?orderBy=newest", SUBROUTE_SUBMISSIONS, ENDPOINT_QUERY_SUBMISSIONS)
			resp := handleQuery(queryRoute)
			switch {
			case !assert.NotEmpty(t, resp, "request response is nil"),
				!assert.Equal(t, submissionIDs[0], resp.Submissions[0].ID, "Submissions not in correct order"),
				!assert.Equal(t, submissionIDs[1], resp.Submissions[1].ID, "Submissions not in correct order"):
				return
			}

			// test that the response is as expected
			queryRoute = fmt.Sprintf("%s%s?orderBy=oldest", SUBROUTE_SUBMISSIONS, ENDPOINT_QUERY_SUBMISSIONS)
			resp = handleQuery(queryRoute)
			switch {
			case !assert.NotEmpty(t, resp, "request response is nil"),
				!assert.Equal(t, submissionIDs[0], resp.Submissions[1].ID, "Submissions not in correct order"),
				!assert.Equal(t, submissionIDs[1], resp.Submissions[0].ID, "Submissions not in correct order"):
				return
			}
		})

		t.Run("order alphabetically", func(t *testing.T) {
			defer clearSubmissions()
			submissionIDs := make([]uint, 2)
			submissionIDs[0] = addTestSubmission("btest", nil, []string{"python", "sorting"}, globalAuthors[:1], globalReviewers[:1])
			submissionIDs[1] = addTestSubmission("atest", nil, []string{"go", "sorting"}, globalAuthors[:1], globalReviewers[:1])

			// test that the response is as expected
			queryRoute := fmt.Sprintf("%s%s?orderBy=alphabetical", SUBROUTE_SUBMISSIONS, ENDPOINT_QUERY_SUBMISSIONS)
			resp := handleQuery(queryRoute)
			switch {
			case !assert.NotEmpty(t, resp, "request response is nil"),
				!assert.Equal(t, submissionIDs[0], resp.Submissions[1].ID, "Submissions not in correct order"),
				!assert.Equal(t, submissionIDs[1], resp.Submissions[0].ID, "Submissions not in correct order"):
				return
			}
		})

		t.Run("query by single tag", func(t *testing.T) {
			defer clearSubmissions()
			submissionIDs := make([]uint, 2)
			submissionIDs[0] = addTestSubmission("test1", nil, []string{"python", "sorting"}, globalAuthors[:1], globalReviewers[:1])
			submissionIDs[1] = addTestSubmission("test2", nil, []string{"go", "sorting"}, globalAuthors[:1], globalReviewers[:1])

			queryRoute := fmt.Sprintf("%s%s?tags=python", SUBROUTE_SUBMISSIONS, ENDPOINT_QUERY_SUBMISSIONS)
			respData := handleQuery(queryRoute)
			if !assert.Equal(t, 1, len(respData.Submissions), "too many submissions returned") {
				return
			}
			assert.Equal(t, submissionIDs[0], respData.Submissions[0].ID, "Submissions id incorrect")
		})

		t.Run("query by multiple tags", func(t *testing.T) {
			defer clearSubmissions()
			submissionIDs := make([]uint, 3)
			submissionIDs[0] = addTestSubmission("test1", nil, []string{"python", "sorting"}, globalAuthors[:1], globalReviewers[:1])
			submissionIDs[1] = addTestSubmission("test2", nil, []string{"go", "sorting"}, globalAuthors[:1], globalReviewers[:1])
			submissionIDs[2] = addTestSubmission("test3", nil, []string{"java", "sorting"}, globalAuthors[:1], globalReviewers[:1])

			queryRoute := fmt.Sprintf("%s%s?tags=python&tags=go", SUBROUTE_SUBMISSIONS, ENDPOINT_QUERY_SUBMISSIONS)
			respData := handleQuery(queryRoute)
			switch {
			case !assert.Equal(t, 2, len(respData.Submissions), "too many submissions returned"),
				!assert.Contains(t, submissionIDs, respData.Submissions[0].ID, "Submission id incorrect"),
				!assert.Contains(t, submissionIDs, respData.Submissions[1].ID, "Submission id incorrect"):
				return
			}
		})

		t.Run("query by author", func(t *testing.T) {
			defer clearSubmissions()
			submissionIDs := make([]uint, 2)
			submissionIDs[0] = addTestSubmission("test1", nil, []string{"python", "sorting"}, globalAuthors[:1], globalReviewers[:1])
			submissionIDs[1] = addTestSubmission("test2", nil, []string{"go", "sorting"}, globalAuthors[1:2], globalReviewers[:1])

			queryRoute := fmt.Sprintf("%s%s?authors=%s", SUBROUTE_SUBMISSIONS, ENDPOINT_QUERY_SUBMISSIONS, globalAuthors[0].ID)
			respData := handleQuery(queryRoute)
			switch {
			case !assert.Equal(t, 1, len(respData.Submissions), "too many submissions returned"),
				!assert.Equal(t, submissionIDs[0], respData.Submissions[0].ID, "Submission id incorrect"):
				return
			}
		})

		t.Run("query by reviewer", func(t *testing.T) {
			defer clearSubmissions()
			submissionIDs := make([]uint, 2)
			submissionIDs[0] = addTestSubmission("test1", nil, []string{"python", "sorting"}, globalAuthors[:1], globalReviewers[:1])
			submissionIDs[1] = addTestSubmission("test2", nil, []string{"go", "sorting"}, globalAuthors[:1], globalReviewers[1:2])

			queryRoute := fmt.Sprintf("%s%s?reviewers=%s", SUBROUTE_SUBMISSIONS, ENDPOINT_QUERY_SUBMISSIONS, globalReviewers[0].ID)
			respData := handleQuery(queryRoute)
			switch {
			case !assert.Equal(t, 1, len(respData.Submissions), "too many submissions returned"),
				!assert.Equal(t, submissionIDs[0], respData.Submissions[0].ID, "Submission id incorrect"):
				return
			}
		})

		// note when querying for names, the query returns all names which exactly match or match 1 word from the name
		t.Run("query by name", func(t *testing.T) {
			defer clearSubmissions()
			submissionIDs := make([]uint, 5)
			submissionIDs[0] = addTestSubmission("unique", nil, []string{"python", "sorting"}, globalAuthors[:1], globalReviewers[:1])
			submissionIDs[1] = addTestSubmission("test2", nil, []string{"go", "sorting"}, globalAuthors[:1], globalReviewers[:1])
			submissionIDs[2] = addTestSubmission("test python", nil, []string{"go", "sorting"}, globalAuthors[:1], globalReviewers[:1])
			submissionIDs[3] = addTestSubmission("testpython", nil, []string{"go", "sorting"}, globalAuthors[:1], globalReviewers[:1])
			submissionIDs[4] = addTestSubmission("[$]python", nil, []string{"go", "sorting"}, globalAuthors[:1], globalReviewers[:1])

			t.Run("full name", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?name=unique", SUBROUTE_SUBMISSIONS, ENDPOINT_QUERY_SUBMISSIONS)
				respData := handleQuery(queryRoute)
				switch {
				case !assert.Equal(t, 1, len(respData.Submissions), "too many submissions returned"),
					!assert.Equal(t, submissionIDs[0], respData.Submissions[0].ID, "Submission id incorrect"):
					return
				}
			})

			t.Run("part of name", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?name=test", SUBROUTE_SUBMISSIONS, ENDPOINT_QUERY_SUBMISSIONS)
				respData := handleQuery(queryRoute)
				switch {
				case !assert.Equal(t, 3, len(respData.Submissions), "too many submissions returned"),
					!assert.Contains(t, submissionIDs[1:4], respData.Submissions[0].ID, "Submission id incorrect"),
					!assert.Contains(t, submissionIDs[1:4], respData.Submissions[1].ID, "Submission id incorrect"),
					!assert.Contains(t, submissionIDs[1:4], respData.Submissions[2].ID, "Submission id incorrect"):
					return
				}
			})

			t.Run("regex with spaces", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?name=unique+test", SUBROUTE_SUBMISSIONS, ENDPOINT_QUERY_SUBMISSIONS)
				respData := handleQuery(queryRoute)
				switch {
				case !assert.Equal(t, 4, len(respData.Submissions), "too many submissions returned"),
					!assert.Contains(t, submissionIDs[:4], respData.Submissions[0].ID, "Submission id incorrect"),
					!assert.Contains(t, submissionIDs[:4], respData.Submissions[1].ID, "Submission id incorrect"),
					!assert.Contains(t, submissionIDs[:4], respData.Submissions[2].ID, "Submission id incorrect"),
					!assert.Contains(t, submissionIDs[:4], respData.Submissions[3].ID, "Submission id incorrect"):
					return
				}
			})

			t.Run("include regex chars", func(t *testing.T) {
				queryRoute := fmt.Sprintf("%s%s?name=[$]", SUBROUTE_SUBMISSIONS, ENDPOINT_QUERY_SUBMISSIONS)
				respData := handleQuery(queryRoute)
				switch {
				case !assert.Equal(t, 1, len(respData.Submissions), "too many submissions returned"),
					!assert.Equal(t, submissionIDs[4], respData.Submissions[0].ID, "Submission id incorrect"):
					return
				}
			})
		})

		t.Run("query by approval", func(t *testing.T) {
			defer clearSubmissions()
			submissionIDs := make([]uint, 3)
			tval := true
			fval := false
			submissionIDs[0] = addTestSubmission("test1", &tval, []string{"python", "sorting"}, globalAuthors[:1], globalReviewers[:1])
			submissionIDs[1] = addTestSubmission("test2", &fval, []string{"go", "sorting"}, globalAuthors[:1], globalReviewers[1:2])
			submissionIDs[2] = addTestSubmission("test3", nil, []string{"go", "sorting"}, globalAuthors[:1], globalReviewers[1:2])

			// just here for easy swapping of approval type to reduce code duplication
			test := func(approved string, index int) {
				queryRoute := fmt.Sprintf("%s%s?approved=%s", SUBROUTE_SUBMISSIONS, ENDPOINT_QUERY_SUBMISSIONS, approved)
				respData := handleQuery(queryRoute)
				switch {
				case !assert.Equal(t, 1, len(respData.Submissions), "too many submissions returned"),
					!assert.Equal(t, submissionIDs[index], respData.Submissions[0].ID, "Submission id incorrect"):
					return
				}
			}
			test("accepted", 0)
			test("rejected", 1)
			test("unapproved", 2)
		})
	})

	t.Run("Error Handling Validation", func(t *testing.T) {
		// handles sending the request and returns the response
		handleQuery := func(queryRoute string) *http.Response {
			req, w := httptest.NewRequest(http.MethodGet, queryRoute, nil), httptest.NewRecorder()
			router.ServeHTTP(w, req)
			return w.Result()
		}

		t.Run("bad value given for query parameter", func(t *testing.T) {
			defer clearSubmissions()
			submissionIDs := make([]uint, 2)
			submissionIDs[0] = addTestSubmission("test1", nil, []string{"python", "sorting"}, globalAuthors[:1], globalReviewers[:1])
			submissionIDs[1] = addTestSubmission("test2", nil, []string{"go", "sorting"}, globalAuthors[:1], globalReviewers[1:2])
			queryRoute := fmt.Sprintf("%s%s?orderBy=blub", SUBROUTE_SUBMISSIONS, ENDPOINT_QUERY_SUBMISSIONS)
			resp := handleQuery(queryRoute)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Incorrect status code returned")
		})

		t.Run("query empty result set", func(t *testing.T) {
			defer clearSubmissions()
			submissionIDs := make([]uint, 2)
			submissionIDs[0] = addTestSubmission("test1", nil, []string{"python", "sorting"}, globalAuthors[:1], globalReviewers[:1])
			submissionIDs[1] = addTestSubmission("test2", nil, []string{"go", "sorting"}, globalAuthors[:1], globalReviewers[1:2])
			queryRoute := fmt.Sprintf("%s%s?tags=blub", SUBROUTE_SUBMISSIONS, ENDPOINT_QUERY_SUBMISSIONS)
			resp := handleQuery(queryRoute)
			assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Incorrect status code returned")
		})
	})
}

// Tests that submissions.go can upload submissions properly
func TestUploadSubmission(t *testing.T) {
	// Set up server and configures filesystem/db
	testInit()
	defer testEnd()

	// Create mux router
	router := mux.NewRouter()
	router.HandleFunc(SUBROUTE_SUBMISSIONS+ENDPOINT_UPLOAD_SUBMISSION, PostUploadSubmission)
	route := SUBROUTE_SUBMISSIONS + ENDPOINT_UPLOAD_SUBMISSION

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
			ctx := context.WithValue(req.Context(), "data", RequestContext{
				ID:       globalAuthors[0].ID,
				UserType: globalAuthors[0].UserType,
			})
			router.ServeHTTP(w, req.WithContext(ctx))
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
					{Path: "test.txt", Base64Value: "test"}, // Check correct file paths.
					{Path: "test/test.txt", Base64Value: "test"},
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
				ctx := context.WithValue(req.Context(), "data", RequestContext{
					ID:       globalAuthors[0].ID,
					UserType: globalAuthors[0].UserType,
				})
				router.ServeHTTP(w, req.WithContext(ctx))
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

// Tests the ability of the backend to download submissions as zip files
func TestDownloadSubmission(t *testing.T) {
	// Set up server and test environment
	testInit()
	defer testEnd()

	// Create mux router
	router := mux.NewRouter()
	router.HandleFunc(SUBROUTE_SUBMISSION+"/{id}"+ENDPOINT_DOWNLOAD_SUBMISSION, GetDownloadSubmission)

	// Initialise users and created submissions.
	globalAuthors, _, err := initMockUsers(t)
	if err != nil {
		return
	}
	// here we encode the file input to mimic a real submission
	submission := Submission{
		Name:    "Test",
		Authors: []GlobalUser{globalAuthors[0]},
		Files: []File{
			{Path: "test.txt", Base64Value: base64.StdEncoding.EncodeToString([]byte("test"))},
			{Path: "test/test.txt", Base64Value: base64.StdEncoding.EncodeToString([]byte("test"))},
		},
		MetaData: &SubmissionData{
			Abstract: "Test",
		},
	}
	submissionID, err := addSubmission(&submission)
	if !assert.NoError(t, err, "Submission creation shouldn't error!") {
		return
	}

	// valid requests
	t.Run("Valid Request", func(t *testing.T) {
		// this function takes care of sending a request and parsing the submission from the downloaded zip
		downloadSubmission := func(submissionID uint) []File {
			// sends the download request
			url := fmt.Sprintf("%s/%d%s", SUBROUTE_SUBMISSION, submissionID, ENDPOINT_DOWNLOAD_SUBMISSION)
			r, w := httptest.NewRequest(http.MethodGet, url, nil), httptest.NewRecorder()
			router.ServeHTTP(w, r)
			resp := w.Result()

			// decodes the response body from base64 and unzips it
			encodedBytes := make([]byte, 1000)
			n, err := resp.Body.Read(encodedBytes)
			encodedBytes = encodedBytes[0:n]
			if !assert.NoError(t, err, "error occurred while getting download response body") {
				return nil
			}
			files, err := getFileArrayFromZipBase64(string(encodedBytes)) // this function will base64 encode file contents to be stored
			if !assert.NoError(t, err, "error occurred while getting the file array from the zip base 64") {
				return nil
			}
			return files
		}

		t.Run("download valid submission", func(t *testing.T) {
			filesVerified := 0 // counts the number of files from the submission we have compared
			files := downloadSubmission(submissionID)

			// Note that the files in the zip archive are not base64 encoded though they are here (done by getFileArrayFromZipBase64)
			// and the zip files are wrapped with a directory of the same name as the submission
			for _, downloadedFile := range files {
				for _, addedFile := range submission.Files {
					if downloadedFile.Path == (submission.Name+"/"+addedFile.Path) &&
						downloadedFile.Base64Value == addedFile.Base64Value {
						filesVerified += 1
					}
				}
			}
			assert.Equal(t, len(submission.Files), filesVerified, "zip is missing files")
		})
	})

	t.Run("Request Validation", func(t *testing.T) {
		t.Run("non-existant submission", func(t *testing.T) {
			// sends the download request
			url := fmt.Sprintf("%s/%d%s", SUBROUTE_SUBMISSION, submissionID+1, ENDPOINT_DOWNLOAD_SUBMISSION)
			r, w := httptest.NewRequest(http.MethodGet, url, nil), httptest.NewRecorder()
			router.ServeHTTP(w, r)
			resp := w.Result()
			assert.Equal(t, http.StatusNotFound, resp.StatusCode, "incorrect status returned for non-existant submission")
		})

		t.Run("submission id as string", func(t *testing.T) {
			// sends the download request
			url := fmt.Sprintf("%s/%s%s", SUBROUTE_SUBMISSION, "nonid", ENDPOINT_DOWNLOAD_SUBMISSION)
			r, w := httptest.NewRequest(http.MethodGet, url, nil), httptest.NewRecorder()
			router.ServeHTTP(w, r)
			resp := w.Result()
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "incorrect status returned for non-existant submission")
		})
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
			{Path: "test.txt", Base64Value: "test"}, // Check correct file paths.
			{Path: "test/test.txt", Base64Value: "test"},
		},
		MetaData: &SubmissionData{
			Abstract: "test", // Check that metadata is correctly stored.
			Reviews: []*Review{
				{ReviewerID: globalReviewers[0].ID, Approved: true, Base64Value: "test"},
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
					{Path: "test.txt"},
					{Path: "test.txt"},
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
			testFiles = append(testFiles, File{Path: file.Path})
		}
		files := []File{}
		for _, file := range queriedSubmission.Files {
			files = append(files, File{Path: file.Path})
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

// Test RouteUploadSubmissionByZip - executes similarly to UploadSubmission,
// with zip file instead of file array.
func TestRouteUploadSubmissionByZip(t *testing.T) {
	testInit()
	defer testEnd()

	authors, _, err := initMockUsers(t)
	if err != nil {
		return
	}

	// Create mux router, give handler, and valid context.
	route := SUBROUTE_SUBMISSIONS + ENDPOINT_UPLOAD_SUBMISSION
	router := mux.NewRouter()
	router.HandleFunc(route, PostUploadSubmissionByZip)
	reqCtx := RequestContext{
		ID:       authors[0].ID,
		UserType: authors[0].UserType,
	}

	t.Run("Valid decoded zip file", func(t *testing.T) {
		// Get test Zip file's base 64 value.
		content, err := ioutil.ReadFile(TEST_ZIP_PATH)
		if !assert.NoErrorf(t, err, "Zip file failed to open: %v", err) {
			return
		}

		// Valid Zip file for a submission
		testFileZipSubmission := UploadSubmissionByZipBody{
			Name:           "Test",
			Authors:        []string{authors[0].ID},
			ZipBase64Value: base64.URLEncoding.EncodeToString(content),
		}

		// Send request body and get response.
		reqBody, err := json.Marshal(testFileZipSubmission)
		if !assert.NoErrorf(t, err, "JSON marshalling shouldn't error.") {
			return
		}
		req := httptest.NewRequest(http.MethodPost, route, bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()
		ctx := context.WithValue(req.Context(), "data", reqCtx)
		RequestLoggerMiddleware(router).ServeHTTP(w, req.WithContext(ctx))
		res := w.Result()

		// Check result
		var body UploadSubmissionResponse
		switch {
		case !assert.NoError(t, json.NewDecoder(res.Body).Decode(&body), "Response is not under the right format"):
			fallthrough
		case !assert.Equalf(t, http.StatusOK, res.StatusCode, "Should succeed, but got \"%s\"", body.Message):
			return
		}

		// Check submission exists in database
		var submission Submission
		err = gormDb.Preload(clause.Associations).First(&submission, body.SubmissionID).Error
		if !assert.NoErrorf(t, err, "Submission fetch should not fail!") {
			return
		}

		// Check if files are in the filesystem.
		path := getSubmissionDirectoryPath(submission)
		for _, file := range submission.Files {
			_, err := os.Stat(filepath.Join(path, fmt.Sprintf("%d", file.ID)))
			if !assert.NoErrorf(t, err, "File stat shouldn't fail/ file should exist!") {
				return
			}
		}
	})

	t.Run("Empty Zip content", func(t *testing.T) {
		// Valid Zip file for a submission
		emptyZipSubmission := UploadSubmissionByZipBody{
			Name:    "Test",
			Authors: []string{authors[0].ID},
		}

		// Send request body and get response.
		reqBody, err := json.Marshal(emptyZipSubmission)
		if !assert.NoErrorf(t, err, "JSON marshalling shouldn't error.") {
			return
		}
		req := httptest.NewRequest(http.MethodPost, route, bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()
		ctx := context.WithValue(req.Context(), "data", reqCtx)
		RequestLoggerMiddleware(router).ServeHTTP(w, req.WithContext(ctx))
		res := w.Result()

		if !assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Empty zip uploads should be rejected") {
			return
		}
	})

	t.Run("Invalid zip", func(t *testing.T) {
		// Valid Zip file for a submission
		emptyZipSubmission := UploadSubmissionByZipBody{
			Name:           "Test",
			Authors:        []string{authors[0].ID},
			ZipBase64Value: "ADbasdflADA==",
		}

		// Send request body and get response.
		reqBody, err := json.Marshal(emptyZipSubmission)
		if !assert.NoErrorf(t, err, "JSON marshalling shouldn't error.") {
			return
		}
		req := httptest.NewRequest(http.MethodPost, route, bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()
		ctx := context.WithValue(req.Context(), "data", reqCtx)
		RequestLoggerMiddleware(router).ServeHTTP(w, req.WithContext(ctx))
		res := w.Result()

		if !assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Empty zip uploads should be rejected") {
			return
		}
	})
}
