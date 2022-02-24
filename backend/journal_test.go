package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// Test user log in.
func TestJournalLogIn(t *testing.T) {
	// Set up test
	testInit()
	defer testEnd()

	// Populate database with valid users.
	trialUsers := make([]GlobalUser, len(testUsers))
	for i, u := range testUsers {
		trialUsers[i] = GlobalUser{User: *u.getCopy(), UserType: USERTYPE_REVIEWER_PUBLISHER}
		trialUsers[i].ID, _ = registerUser(trialUsers[i].User, USERTYPE_REVIEWER_PUBLISHER)
	}

	router := mux.NewRouter()
	router.HandleFunc(SUBROUTE_JOURNAL+ENDPOINT_LOGIN, logIn)

	// Test valid logins
	t.Run("Valid logins", func(t *testing.T) {
		for i := range testUsers {
			// Create a request for user login.
			loginBody := JournalLoginPostBody{Email: testUsers[i].Email, Password: testUsers[i].Password}
			reqBody, _ := json.Marshal(loginBody)
			req, w := httptest.NewRequest("POST", SUBROUTE_JOURNAL+ENDPOINT_LOGIN, bytes.NewBuffer(reqBody)), httptest.NewRecorder()
			router.ServeHTTP(w, req)
			resp := w.Result()

			assert.Equalf(t, http.StatusOK, resp.StatusCode, "Response status should be %d", http.StatusOK)

			// Get ID from user response.
			var respMap JournalLogInResponse
			if err := json.NewDecoder(resp.Body).Decode(&respMap); !assert.Nil(t, err, "Body unparsing should succeed") {
				return
			}
			// Check if gotten
			assert.Equal(t, trialUsers[i].ID, respMap.ID, "ID must equal registration's ID.")
		}
	})

	// Test invalid password login.
	t.Run("Invalid password logins", func(t *testing.T) {
		for i := 0; i < len(testUsers); i++ {
			loginMap := JournalLoginPostBody{Email: testUsers[i].Email, Password: VALID_PW}

			reqBody, _ := json.Marshal(loginMap)
			req, w := httptest.NewRequest("POST", SUBROUTE_JOURNAL+ENDPOINT_LOGIN, bytes.NewBuffer(reqBody)), httptest.NewRecorder()
			router.ServeHTTP(w, req)
			resp := w.Result()

			assert.Equalf(t, http.StatusUnauthorized, resp.StatusCode, "Response should have status %d", http.StatusUnauthorized)
		}
	})

	// Test invalid email login.
	t.Run("Invalid email logins", func(t *testing.T) {
		for i := 1; i < len(testUsers); i++ {
			loginMap := JournalLoginPostBody{Email: testUsers[0].Email, Password: testUsers[i].Password}

			reqBody, _ := json.Marshal(loginMap)
			req, w := httptest.NewRequest("POST", SUBROUTE_JOURNAL+ENDPOINT_LOGIN, bytes.NewBuffer(reqBody)), httptest.NewRecorder()
			router.ServeHTTP(w, req)
			resp := w.Result()

			assert.Equalf(t, http.StatusUnauthorized, resp.StatusCode, "Response should have status %d", http.StatusUnauthorized)
		}
	})
}

// This tests converting from the local submission data format to the supergroup specified format
func TestLocalToGlobal(t *testing.T) {
	testInit()
	defer testEnd()

	// adds the submission and a file to the system
	testSubmission := *testSubmissions[0].getCopy()
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
	submissionID, err := addSubmission(testSubmission.getCopy())
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
		categories := make([]string, len(testSubmission.Categories))
		for i, category := range testSubmission.Categories {
			categories[i] = category.Tag
		}
		switch {
		case !assert.Equal(t, testSubmission.Name, globalSubmission.Name, "Names do not match"),
			!assert.Equal(t, testSubmission.License, globalSubmission.MetaData.License,
				"Licenses do not match"),
			!assert.Equal(t, testAuthor.FirstName+" "+testAuthor.LastName, globalSubmission.MetaData.AuthorNames[0],
				"Authors do not match"),
			!assert.Equal(t, categories, globalSubmission.MetaData.Categories,
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
