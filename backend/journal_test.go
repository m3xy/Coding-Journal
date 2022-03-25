package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// ---------
// Handler Function Tests
// ---------

// Test user log in.
func TestJournalLogIn(t *testing.T) {
	// Set up test
	testInit()
	defer testEnd()

	// Populate database with valid users.
	trialUsers := make([]GlobalUser, len(testGlobUsers))
	for i, u := range testGlobUsers {
		trialUsers[i] = u
		trialUsers[i].ID, _ = registerTestUser(trialUsers[i])
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

// tests ability of another journal to get all global users from this journal
func TestGetUsers(t *testing.T) {
	// Set up test
	testInit()
	defer testEnd()

	// Populate database with valid users.
	var id string // temp loop variable
	trialUsers := map[string]GlobalUser{}
	for _, u := range testGlobUsers {
		id, _ = registerTestUser(u)
		trialUsers[id] = u
	}

	router := mux.NewRouter()
	router.HandleFunc(SUBROUTE_JOURNAL+ENDPOINT_USER, GetUsers)

	t.Run("Get all users", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, SUBROUTE_JOURNAL+ENDPOINT_USER, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		resp := w.Result()

		// checks status code
		if !assert.Equal(t, http.StatusOK, resp.StatusCode, "status returned not 200") {
			return
		}

		// gets the user list from the json response
		users := []SupergroupUser{}
		if !assert.NoError(t, json.NewDecoder(resp.Body).Decode(&users), "error decoding response body") {
			return
		}
		// tests that the users match
		var testUser GlobalUser
		for _, user := range users {
			testUser = trialUsers[user.ID]
			switch {
			case !assert.Equal(t, testUser.FirstName, user.FirstName, "user names do not match"),
				!assert.Equal(t, testUser.LastName, user.LastName, "user names do not match"),
				!assert.Equal(t, testUser.User.Email, user.Email, "Emails do not match"),
				!assert.Equal(t, testUser.User.PhoneNumber, user.PhoneNumber, "Phone numbers do not match"),
				!assert.Equal(t, testUser.User.Organization, user.Organization, "Phone numbers do not match"):
				return
			}
		}
	})
}

// tests ability of another journal to get all global users from this journal
func TestGetUser(t *testing.T) {
	testInit()
	defer testEnd()

	// Populate database with valid users.
	trialUsers := make([]GlobalUser, len(testUsers))
	for i, u := range testUsers {
		trialUsers[i] = GlobalUser{
			FirstName: fmt.Sprint(i),
			LastName: fmt.Sprint(i),
			UserType: USERTYPE_REVIEWER_PUBLISHER,
			User: u.getCopy(),
		}
		trialUsers[i].ID, _ = registerTestUser(trialUsers[i])
	}

	router := mux.NewRouter()
	router.HandleFunc(SUBROUTE_JOURNAL+ENDPOINT_USER+"/{id}", GetUser)

	t.Run("Get valid user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, SUBROUTE_JOURNAL+ENDPOINT_USER+"/"+trialUsers[0].ID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		resp := w.Result()

		// checks status code
		if !assert.Equal(t, http.StatusOK, resp.StatusCode, "status returned not 200") {
			return
		}

		queriedUser := SupergroupUser{}
		if !assert.NoError(t, json.NewDecoder(resp.Body).Decode(&queriedUser), "error decoding response body") {
			return
		}

		// test that the retrieved user matches that which was expected
		switch {
		case !assert.Equal(t, trialUsers[0].ID, queriedUser.ID, "user IDs do not match"),
			!assert.Equal(t, trialUsers[0].User.Email, queriedUser.Email, "user names do not match"),
			!assert.Equal(t, trialUsers[0].FirstName, queriedUser.FirstName, "user names do not match"),
			!assert.Equal(t, trialUsers[0].LastName, queriedUser.LastName, "user names do not match"),
			!assert.Equal(t, trialUsers[0].User.Organization, queriedUser.Organization, "user organizations do not match"),
			!assert.Equal(t, trialUsers[0].User.PhoneNumber, queriedUser.PhoneNumber, "user phone numbers do not match"):
			return
		}
	})

	t.Run("Get invalid user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, SUBROUTE_JOURNAL+ENDPOINT_USER+"/fausldifvnaunliaekan", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		resp := w.Result()

		// checks status code
		if !assert.Equal(t, http.StatusNotFound, resp.StatusCode, "status returned not 200") {
			return
		}
	})
}

// Tests the ability of this journal to export submissions to another journal
func TestExportSubmission(t *testing.T) {
	// wipes the database and filesystem
	testInit()
	defer testEnd()

	exportGroupNumber := 2 // group number to export to

	// Create local server
	router := mux.NewRouter()
	route := SUBROUTE_SUBMISSIONS + "/{id}" + ENDPOINT_EXPORT_SUBMISSION + "/{groupNumber}"
	router.HandleFunc(route, PostExportSubmission)

	// Create mock global server
	globalRouter := mux.NewRouter()
	globalRoute := journalURLs[exportGroupNumber] + SUBROUTE_JOURNAL + "/submission"
	globalRouterFunc := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	globalRouter.HandleFunc(globalRoute, globalRouterFunc)

	// adds a submission to the db with authors and reviewers
	globalAuthors, globalReviewers, err := initMockUsers(t)
	if err != nil {
		return
	}

	// adds a test editor
	editorID, err := registerTestUser(testEditors[0])
	if !assert.NoError(t, err, "Error adding test editor") {
		return
	}

	submission := Submission{
		Name:      "Test",
		Authors:   []GlobalUser{globalAuthors[0]},
		Reviewers: []GlobalUser{globalReviewers[0]},
		MetaData: &SubmissionData{
			Abstract: "Test",
		},
	}

	submissionID, err := addSubmission(&submission)
	if !assert.NoError(t, err, "Submission creation shouldn't error!") {
		return
	}

	// utility function to send export submission requests to the local server
	testExportSubmission := func(submissionID uint, ctxStruct *RequestContext, groupNb int) *http.Response {
		// mocks out function to contact other servers
		sendSecureRequest = func(db *gorm.DB, req *http.Request, groupNb int) (*http.Response, error) {
			w := httptest.NewRecorder()
			w.WriteHeader(http.StatusOK)
			return w.Result(), nil
		}
		// builds request to the backend
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("%s/%d%s/%d", SUBROUTE_SUBMISSIONS,
			submissionID, ENDPOINT_EXPORT_SUBMISSION, groupNb), nil)
		w := httptest.NewRecorder()
		ctx := context.WithValue(req.Context(), "data", ctxStruct)
		router.ServeHTTP(w, req.WithContext(ctx))
		resp := w.Result()
		return resp
	}

	// exports a valid submission
	t.Run("Export Valid Submission", func(t *testing.T) {
		ctx := &RequestContext{
			ID:       editorID,
			UserType: USERTYPE_EDITOR,
		}
		resp := testExportSubmission(submissionID, ctx, exportGroupNumber)
		assert.Equalf(t, http.StatusOK, resp.StatusCode, "Returned Wrong status code!")
	})

	// makes sure the errors occur in the right places
	t.Run("Request verification", func(t *testing.T) {
		t.Run("Wrong Permissions", func(t *testing.T) {
			ctx := &RequestContext{
				ID:       globalAuthors[1].ID,
				UserType: USERTYPE_PUBLISHER,
			}
			resp := testExportSubmission(submissionID, ctx, exportGroupNumber)
			assert.Equalf(t, http.StatusUnauthorized, resp.StatusCode, "Returned Wrong status code!")
		})

		t.Run("Invalid Group Number", func(t *testing.T) {
			ctx := &RequestContext{
				ID:       editorID,
				UserType: USERTYPE_EDITOR,
			}
			resp := testExportSubmission(submissionID, ctx, -1)
			assert.Equalf(t, http.StatusBadRequest, resp.StatusCode, "Returned Wrong status code!")
		})

		t.Run("Bad Submission ID", func(t *testing.T) {
			// submission does not exist
			ctx := &RequestContext{
				ID:       editorID,
				UserType: USERTYPE_EDITOR,
			}
			resp := testExportSubmission(submissionID+1, ctx, exportGroupNumber)
			assert.Equalf(t, http.StatusBadRequest, resp.StatusCode, "Returned Wrong status code!")
		})
	})
}

// Tests the ability of this journal to import submissions from another journal
func TestImportSubmission(t *testing.T) {
	// wipes the database and filesystem
	testInit()
	defer testEnd()

	// init server
	router := mux.NewRouter()
	route := SUBROUTE_JOURNAL + "/submission"
	router.HandleFunc(route, PostImportSubmission)

	globalAuthors, globalReviewers, err := initMockUsers(t)
	if !assert.NoError(t, err, "error registering test users") {
		return 
	}
	authorID := globalAuthors[0].ID

	// test supergroup submission
	globalMetadata := SupergroupSubmissionData{
		CreationDate: time.Now(),
		Abstract:     "abstract",
		License:      "MIT",
		Categories:   []string{"python", "sorting"},
		Authors: []SuperGroupAuthor{
			{ID: authorID, Journal: "11"},
		},
	}
	globalCodeVersion := SupergroupCodeVersion{
		TimeStamp: time.Now(),
		Files: []SupergroupFile{
			{Name: "hello", Base64Value: "Goodbye"},
		},
	}
	globalSub := &SupergroupSubmission{
		Name:         "test",
		MetaData:     globalMetadata,
		CodeVersions: []SupergroupCodeVersion{globalCodeVersion},
	}

	// utility function to send POST requests to tell the local server to import global submissions
	testImportSubmission := func(globalSub *SupergroupSubmission) *http.Response {
		// builds request to the backend
		reqBody, err := json.Marshal(globalSub)
		if !assert.NoError(t, err, "error occurred while marshalling request body") {
			return nil
		}
		req := httptest.NewRequest(http.MethodPost, SUBROUTE_JOURNAL+"/submission", bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		resp := w.Result()
		return resp
	}

	t.Run("Import Valid Submission", func(t *testing.T) {
		resp := testImportSubmission(globalSub.getCopy())
		assert.Equalf(t, http.StatusOK, resp.StatusCode, "Returned Wrong status code!")
	})

	// makes sure the errors occur in the right places
	t.Run("Request verification", func(t *testing.T) {
		t.Run("Author Wrong Permissions", func(t *testing.T) {
			newGlobalSub := globalSub.getCopy()
			newGlobalSub.MetaData.Authors = []SuperGroupAuthor{
				{ID: globalReviewers[0].ID, Journal: "11"},
			}
			resp := testImportSubmission(newGlobalSub)
			assert.Equalf(t, http.StatusUnauthorized, resp.StatusCode, "Returned Wrong status code!")
		})

		t.Run("Unregistered Author", func(t *testing.T) {
			newGlobalSub := globalSub.getCopy()
			newGlobalSub.MetaData.Authors = []SuperGroupAuthor{
				{ID: "aiargiajradkgj430293", Journal: "11"},
			}
			resp := testImportSubmission(newGlobalSub)
			assert.Equalf(t, http.StatusUnauthorized, resp.StatusCode, "Returned Wrong status code!")
		})

		t.Run("Empty request", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, SUBROUTE_JOURNAL+"/submission", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			resp := w.Result()
			assert.Equalf(t, http.StatusBadRequest, resp.StatusCode, "Returned Wrong status code!")
		})

		t.Run("Duplicate File Error", func(t *testing.T) {
			newGlobalSub := globalSub.getCopy()
			file := SupergroupFile{
				Name:        "name",
				Base64Value: "file",
			}
			newGlobalSub.CodeVersions[0].Files = []SupergroupFile{file, file}
			resp := testImportSubmission(newGlobalSub)
			assert.Equalf(t, http.StatusBadRequest, resp.StatusCode, "Returned Wrong status code!")
		})
	})
}

// ---------
// Helper Function Tests
// ---------

// This tests converting from the local submission data format to the supergroup specified format
func TestLocalToGlobal(t *testing.T) {
	testInit()
	defer testEnd()

	// adds the submission and a file to the system
	testSubmission := *testSubmissions[0].getCopy()
	testFile := testFiles[0]

	// registers authors and reviewers, and adds them to the test submission
	globalAuthors, globalReviewers, err := initMockUsers(t)
	if err != nil {
		return
	}
	testSubmission.Authors = globalAuthors[:1]
	testSubmission.Reviewers = globalReviewers[:1]

	testSubmission.Files = []File{testFile}
	submissionID, err := addSubmission(testSubmission.getCopy())
	if !assert.NoErrorf(t, err, "Error occurred while adding submission: %v", err) {
		return
	}

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
			!assert.Equal(t, testSubmission.Authors[0].ID, globalSubmission.MetaData.Authors[0].ID,
				"Authors do not match"),
			!assert.Equal(t, categories, globalSubmission.MetaData.Categories, "Tags do not match"),
			!assert.Equal(t, testSubmission.MetaData.Abstract, globalSubmission.MetaData.Abstract,
				"Abstracts do not match"),
			// compares files
			!assert.Equal(t, testFile.Path, globalSubmission.CodeVersions[0].Files[0].Name, "File names do not match"),
			!assert.Equal(t, testFile.Base64Value, globalSubmission.CodeVersions[0].Files[0].Base64Value, "File content does not match"):
			return
		}
	})

	t.Run("Non-existant Submission", func(t *testing.T) {
		_, err := localToGlobal(submissionID + 1)
		assert.Error(t, err, "did not err on non-existant submission conversion")
		assert.IsType(t, &NoSubmissionError{}, err, "Incorrect type of error returned")
	})
}

// This tests converting from the supergroup submission data format to the local format
func TestGlobalToLocal(t *testing.T) {
	testInit()
	defer testEnd()

	// adds a submission to the db with authors and reviewers
	globalAuthors, _, err := initMockUsers(t)
	if err != nil {
		return
	}
	authorID := globalAuthors[0].ID

	// test supergroup submission
	globalMetadata := SupergroupSubmissionData{
		CreationDate: time.Now(),
		Abstract:     "abstract",
		License:      "MIT",
		Categories:   []string{"python", "sorting"},
		Authors: []SuperGroupAuthor{
			{ID: authorID, Journal: "11"},
		},
	}
	globalCodeVersion := SupergroupCodeVersion{
		TimeStamp: time.Now(),
		Files: []SupergroupFile{
			{Name: "hello", Base64Value: "Goodbye"},
		},
	}
	globalSub := &SupergroupSubmission{
		Name:         "test",
		MetaData:     globalMetadata,
		CodeVersions: []SupergroupCodeVersion{globalCodeVersion},
	}

	testGlobalToLocal := func(globalSub *SupergroupSubmission) {
		localSubmission, err := globalToLocal(globalSub)
		if !assert.NoErrorf(t, err, "Error occurred while converting submission format!") {
			return
		}
		// builds a comparable array of categories
		categories := []string{}
		for _, category := range localSubmission.Categories {
			categories = append(categories, category.Tag)
		}
		switch {
		case !assert.Equal(t, localSubmission.Name, globalSub.Name, "Names do not match"),
			!assert.Equal(t, localSubmission.License, globalSub.MetaData.License,
				"Licenses do not match"),
			!assert.Equal(t, localSubmission.Authors[0].ID, globalSub.MetaData.Authors[0].ID,
				"Authors do not match"),
			!assert.Equal(t, globalSub.MetaData.Categories, categories, "Tags do not match"),
			!assert.Equal(t, localSubmission.MetaData.Abstract, globalSub.MetaData.Abstract,
				"Abstracts do not match"),
			// compares files
			!assert.Equal(t, globalCodeVersion.Files[0].Name, localSubmission.Files[0].Path,
				"File names do not match"),
			!assert.Equal(t, globalCodeVersion.Files[0].Base64Value, localSubmission.Files[0].Base64Value,
				"File content does not match"):
			return
		}
	}

	t.Run("Valid Submission", func(t *testing.T) {
		testGlobalToLocal(globalSub)
	})

	t.Run("nil submission", func(t *testing.T) {
		_, err := globalToLocal(nil)
		assert.Error(t, err, "did not err on nil submission conversion")
	})
}
