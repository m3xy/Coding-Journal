// =====================================
// approval_test.go
// Authors: 190010425
// Created: February 28, 2022
//
// test file for approval.go
// =====================================

package main 

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// ------------
// Router Function Tests
// ------------

func TestRouteAssignReviewers(t *testing.T) {
	// wipes the database and filesystem
	testInit()
	defer testEnd()

	// Create mux router
	router := mux.NewRouter()
	route := ENDPOINT_SUBMISSIONS+"/{id}"+ENDPOINT_ASSIGN_REVIEWERS
	router.HandleFunc(route, RouteAssignReviewers)

	// adds a submission to the db with authors and reviewers
	globalAuthors, globalReviewers, err := initMockUsers(t)
	if err != nil {
		return
	}

	// adds a test editor
	editorID, err := registerUser(User{Email: "editor@test.net", 
		Password: "dlbjDs2!", FirstName: "Paul", LastName: "Editman"}, USERTYPE_EDITOR)
	if !assert.NoError(t, err, "Error adding test editor") {
		return
	}

	submission := Submission{
		Name:    "Test",
		Authors: []GlobalUser{globalAuthors[0]},
		Reviewers: []GlobalUser{},
		MetaData: &SubmissionData{
			Abstract: "Test",
		},
	}

	submissionID, err := addSubmission(&submission)
	if !assert.NoError(t, err, "Submission creation shouldn't error!") {
		return
	}

	// removes all reviewer associations from the given submission
	resetReviewers := func(submissionID uint) {
		assert.NoError(t, gormDb.Model(&Submission{}).Association("Reviewers").Delete(submissionID),
			"reviewers could not be removed from the submission")
	}

	// does the work of sending a test request and returns the status code
	testAssignReviewers := func(reqStruct *AssignReviewersBody, submissionID uint, ctxStruct *RequestContext) int {
		resetReviewers(submissionID)
		reqBody, err := json.Marshal(reqStruct)
		if !assert.NoError(t, err, "Error while marshalling assign reviewers body!") {
			return -1
		}

		// sends the request to assign reviewers
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("%s/%d%s", ENDPOINT_SUBMISSIONS, 
			submissionID, ENDPOINT_ASSIGN_REVIEWERS), bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), "data", *ctxStruct)
		router.ServeHTTP(w, req.WithContext(ctx))
		resp := w.Result()

		return resp.StatusCode
	}

	t.Run("Add Single Reviewer Valid", func(t *testing.T) {
		reqStruct := &AssignReviewersBody{
			Reviewers: []string{globalReviewers[0].ID},
		}
		ctx := &RequestContext{
			ID: editorID,
			UserType: USERTYPE_EDITOR,
		}
		// makes sure the request succeeded
		status := testAssignReviewers(reqStruct, submissionID, ctx)
		if !assert.Equalf(t, http.StatusOK, status, "request did not succeed!") {
			return
		}
	})

	t.Run("Non-editor Request", func(t *testing.T) {
		reqStruct := &AssignReviewersBody{
			Reviewers: []string{globalReviewers[0].ID},
		}
		ctx := &RequestContext{
			ID: editorID,
			UserType: USERTYPE_REVIEWER,
		}
		// makes sure the request did not succeed
		status := testAssignReviewers(reqStruct, submissionID, ctx)
		if !assert.Equalf(t, http.StatusUnauthorized, status, "request did not succeed!") {
			return
		}
	})

	t.Run("Assign Non-reviewer", func(t *testing.T) {
		reqStruct := &AssignReviewersBody{
			Reviewers: []string{globalAuthors[1].ID},
		}
		ctx := &RequestContext{
			ID: editorID,
			UserType: USERTYPE_EDITOR,
		}
		// makes sure the request did not succeed
		status := testAssignReviewers(reqStruct, submissionID, ctx)
		if !assert.Equalf(t, http.StatusBadRequest, status, "request did not succeed!") {
			return
		}
	})

	t.Run("Assign Non-user as reviewer", func(t *testing.T) {
		reqStruct := &AssignReviewersBody{
			Reviewers: []string{"uuiagief84920348hgakagjdka"},
		}
		ctx := &RequestContext{
			ID: editorID,
			UserType: USERTYPE_EDITOR,
		}
		// makes sure the request did not succeed
		status := testAssignReviewers(reqStruct, submissionID, ctx)
		if !assert.Equalf(t, http.StatusBadRequest, status, "request did not succeed!") {
			return
		}
	})

	// make sure this test runs last
	t.Run("Assign Reviewer to Approved Submission", func(t *testing.T) {
		reqStruct := &AssignReviewersBody{
			Reviewers: []string{globalReviewers[0].ID},
		}
		ctx := &RequestContext{
			ID: editorID,
			UserType: USERTYPE_EDITOR,
		}
		// approves the submission
		if !assert.NoError(t, gormDb.Model(&Submission{}).Where("ID = ?", submissionID).Update("approved", true).Error,
			"changing submission status should not err!") {
			return
		}
		// makes sure the request did not succeed
		status := testAssignReviewers(reqStruct, submissionID, ctx)
		if !assert.Equalf(t, http.StatusConflict, status, "request did not succeed!") {
			return
		}
	})
}

func TestUploadReview(t *testing.T) {
	// wipes the database and filesystem
	testInit()
	defer testEnd()

	// Create mux router
	router := mux.NewRouter()
	route := ENDPOINT_SUBMISSIONS+"/{id}"+ENPOINT_REVIEW
	router.HandleFunc(route, RouteUploadReview)

	// adds a submission to the db with authors and reviewers
	globalAuthors, globalReviewers, err := initMockUsers(t)
	if err != nil {
		return
	}

	submission := Submission{
		Name:    "Test",
		Authors: []GlobalUser{globalAuthors[0]},
		Reviewers: []GlobalUser{globalReviewers[0], globalReviewers[1]},
		MetaData: &SubmissionData{
			Abstract: "Test",
		},
	}

	submissionID, err := addSubmission(&submission)
	if !assert.NoError(t, err, "Submission creation shouldn't error!") {
		return
	}

	// uses globalReviewers[0]
	t.Run("Single Approving Review", func(t *testing.T) {
		// passes on no error
		t.Run("Adds Review", func(t *testing.T) {
			reqStruct := &UploadReviewBody{
				Approved: true,
				Base64Value: "test",
			}
			reqBody, err := json.Marshal(reqStruct)
			if !assert.NoError(t, err, "Error while marshalling review upload body!") {
				return
			}

			// sends the request to upload a review
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("%s/%d%s", ENDPOINT_SUBMISSIONS, submissionID, ENPOINT_REVIEW), bytes.NewBuffer(reqBody))
			w := httptest.NewRecorder()

			ctx := context.WithValue(req.Context(), "data", RequestContext{
				ID: globalReviewers[0].ID,
				UserType: globalReviewers[0].UserType,
			})
			router.ServeHTTP(w, req.WithContext(ctx))
			resp := w.Result()

			// makes sure the request succeeded
			if !assert.Equalf(t, http.StatusOK, resp.StatusCode, "request did not succeed!") {
				return
			}

			// gets the submission metadata and checks that it matches that which was sent
			queriedMetaData, err := getSubmissionMetaData(submissionID)
			if !assert.NoError(t, err, "Error while getting submission metadata!") {
				return
			}
			queriedReview := queriedMetaData.Reviews[0]

			// compares reviews in queried metadata with that which was added
			switch {
			case !assert.Equal(t, globalReviewers[0].ID, queriedReview.ReviewerID, "Reviewer IDs do not match"),
				!assert.Equal(t, reqStruct.Approved, queriedReview.Approved, "Review approval does not match"),
				!assert.Equal(t, reqStruct.Base64Value, queriedReview.Base64Value, "Review content does not match"):
				return
			}
		})

		// uploads the same review as above, should error -> test passes on err
		t.Run("Adds duplicate review", func(t *testing.T) {
			reqStruct := &UploadReviewBody{
				Approved: true,
				Base64Value: "test",
			}
			reqBody, err := json.Marshal(reqStruct)
			if !assert.NoError(t, err, "Error while marshalling review upload body!") {
				return
			}

			// sends the request to upload a review
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("%s/%d%s", ENDPOINT_SUBMISSIONS, submissionID, ENPOINT_REVIEW), bytes.NewBuffer(reqBody))
			w := httptest.NewRecorder()

			ctx := context.WithValue(req.Context(), "data", RequestContext{
				ID: globalReviewers[0].ID,
				UserType: globalReviewers[0].UserType,
			})
			router.ServeHTTP(w, req.WithContext(ctx))
			resp := w.Result()

			// makes sure the request succeeded
			if !assert.Equalf(t, http.StatusBadRequest, resp.StatusCode, "duplicate review added without proper error code!") {
				return
			}
		})
	})

	// deals with error cases that occur before the review is added to the submission (uses globalReviewers[1])
	t.Run("Request Validation", func(t *testing.T) {
		// uploads a given review and returns the status code
		RouteUploadReview := func(reqStruct *UploadReviewBody, reviewerID string, userType int, submissionID uint) int {
			reqBody, err := json.Marshal(reqStruct)
			if !assert.NoError(t, err, "Error while marshalling review upload body!") {
				return -1
			}
			// sends the request to upload a review
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("%s/%d%s", ENDPOINT_SUBMISSIONS, submissionID, ENPOINT_REVIEW), bytes.NewBuffer(reqBody))
			w := httptest.NewRecorder()

			ctx := context.WithValue(req.Context(), "data", RequestContext{
				ID: reviewerID,
				UserType: userType,
			})
			router.ServeHTTP(w, req.WithContext(ctx))
			resp := w.Result()

			return resp.StatusCode
		}

		t.Run("User ID not in context", func(t *testing.T) {
			reqStruct := &UploadReviewBody{
				Approved: true,
				Base64Value: "hello",
			}
			// makes sure the request failed StatusUnauthorized
			if !assert.Equalf(t, http.StatusUnauthorized, RouteUploadReview(reqStruct, "", USERTYPE_REVIEWER, submissionID), 
				"unauthenticated user added review without proper error code!") {
				return
			}
		})

		t.Run("Review no content", func(t *testing.T) {
			reqStruct := &UploadReviewBody{
				Approved: true,
			}
			// makes sure the request failed StatusBadRequest
			if !assert.Equalf(t, http.StatusBadRequest, RouteUploadReview(reqStruct, globalReviewers[1].ID, globalReviewers[1].UserType, submissionID), 
				"empty review added without proper error code!") {
				return
			}
		})
	})
}

func TestRouteUpdateSubmissionStatus(t *testing.T) {
	// wipes the database and filesystem
	testInit()
	defer testEnd()

	// Create mux router
	router := mux.NewRouter()
	route := ENDPOINT_SUBMISSIONS+"/{id}"+ENDPOINT_CHANGE_STATUS
	router.HandleFunc(route, RouteUpdateSubmissionStatus)

	// adds a submission to the db with authors and reviewers
	globalAuthors, globalReviewers, err := initMockUsers(t)
	if err != nil {
		return
	}
	// adds a test editor
	editorID, err := registerUser(User{Email: "editor@test.net", 
		Password: "dlbjDs2!", FirstName: "Paul", LastName: "Editman"}, USERTYPE_EDITOR)
	if !assert.NoError(t, err, "Error adding test editor") {
		return
	}

	submission := Submission{
		Name:    "Test",
		Authors: []GlobalUser{globalAuthors[0]},
		Reviewers: []GlobalUser{globalReviewers[0]},
		MetaData: &SubmissionData{
			Abstract: "Test",
		},
	}

	submissionID, err := addSubmission(&submission)
	if !assert.NoError(t, err, "Submission creation shouldn't error!") {
		return
	}

	// function to format and send test requests
	changeStatus := func(submissionID uint, editorID string, userType int, reqStruct *UpdateSubmissionStatusBody) int {
		reqBody, err := json.Marshal(reqStruct)
		if !assert.NoError(t, err, "Error while marshalling request body!") {
			return -1
		}
		// sends the request to upload a review
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("%s/%d%s", ENDPOINT_SUBMISSIONS, submissionID, ENDPOINT_CHANGE_STATUS), bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), "data", RequestContext{
			ID: editorID,
			UserType: userType,
		})
		router.ServeHTTP(w, req.WithContext(ctx))
		resp := w.Result()
		return resp.StatusCode
	}

	// cases where the helper function gets called
	t.Run("Change Status", func(t *testing.T){
		t.Run("Review not added", func(t *testing.T) {
			reqStruct := &UpdateSubmissionStatusBody{Status:true}
			assert.Equal(t, http.StatusConflict, changeStatus(submissionID, editorID, USERTYPE_EDITOR, reqStruct), "Wrong error code, was expecting 409 Conflict")
		})
		
		// adds a review from the one reviewer to allow for submission approval
		validReview := &Review{
			ReviewerID: globalReviewers[0].ID,
			Approved: true,
			Base64Value: "test",
		}
		if !assert.NoError(t, addReview(validReview, submissionID), "Review Addition shouldn't error!") {
			return
		}

		t.Run("Valid approval", func(t *testing.T) {
			reqStruct := &UpdateSubmissionStatusBody{Status:true}
			assert.Equal(t, http.StatusOK, changeStatus(submissionID, editorID, USERTYPE_EDITOR, reqStruct), "Wrong error code, was expecting 200")
			submission := &Submission{}
			if !assert.NoError(t, gormDb.Model(&Submission{}).Select("submissions.approved").Find(submission, submissionID).Error, "unable to find submission") {
				return
			}
			assert.Equal(t, true, *submission.Approved, "Submission status not updated properly")
		})
	})

	// tests error cases having to do with the request itself
	t.Run("Validate Request", func(t *testing.T) {
		t.Run("Context Not Set", func(t *testing.T) {
			// context invalid
			reqStruct := &UpdateSubmissionStatusBody{Status:true}
			assert.Equal(t, http.StatusUnauthorized, changeStatus(submissionID, "", -1, reqStruct), "Wrong error code, was expecting 401")

			// context not set at all
			reqBody, err := json.Marshal(reqStruct)
			if !assert.NoError(t, err, "Error while marshalling request body!") {
				return
			}
			// sends the request to upload a review
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("%s/%d%s", ENDPOINT_SUBMISSIONS, submissionID, ENDPOINT_CHANGE_STATUS), bytes.NewBuffer(reqBody))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode, "Wrong error code, was expecting 401")
		})

		t.Run("Non-editor user type", func(t *testing.T){
			reqStruct := &UpdateSubmissionStatusBody{Status:true}
			assert.Equal(t, http.StatusUnauthorized, changeStatus(submissionID, globalReviewers[0].ID, USERTYPE_REVIEWER, reqStruct), "Wrong error code, was expecting 401")
		})

		t.Run("Bad Request body", func(t *testing.T){
			reqStruct := &UpdateSubmissionStatusBody{}
			assert.Equal(t, http.StatusBadRequest, changeStatus(submissionID, editorID, USERTYPE_EDITOR, reqStruct), "Wrong error code, was expecting 400")
			assert.Equal(t, http.StatusBadRequest, changeStatus(submissionID, editorID, USERTYPE_EDITOR, nil), "Wrong error code, was expecting 400")
		})
	})
}


// ------------
// Helper Function Tests
// ------------

func TestAddReview(t *testing.T) {
	// configures main test environment
	testInit()
	defer testEnd()

	// adds a submission to the db with authors and reviewers
	globalAuthors, globalReviewers, err := initMockUsers(t)
	if err != nil {
		return
	}

	submission := Submission{
		Name:    "Test",
		Authors: []GlobalUser{globalAuthors[0]},
		Reviewers: []GlobalUser{globalReviewers[0], globalReviewers[1]},
		MetaData: &SubmissionData{
			Abstract: "Test",
		},
	}

	submissionID, err := addSubmission(&submission)
	if !assert.NoError(t, err, "Submission creation shouldn't error!") {
		return
	}

	t.Run("Single Approved Review", func(t *testing.T) {
		// adds review to the submission
		validReview := &Review{
			ReviewerID: globalReviewers[0].ID,
			Approved: true,
			Base64Value: "test",
		}

		t.Run("Adds single Review", func(t *testing.T) {
			if !assert.NoError(t, addReview(validReview, submissionID), "Review Addition shouldn't error!") {
				return
			}

			// gets the submission metadata and checks that it matches that which was sent
			queriedMetaData, err := getSubmissionMetaData(submissionID)
			if !assert.NoError(t, err, "Error while getting submission metadata!") {
				return
			}
			queriedReview := queriedMetaData.Reviews[0]

			// compares reviews in queried metadata with that which was added
			switch {
			case !assert.Equal(t, validReview.ReviewerID, queriedReview.ReviewerID, "Reviewer IDs do not match"),
				!assert.Equal(t, validReview.Approved, queriedReview.Approved, "Review approval does not match"),
				!assert.Equal(t, validReview.Base64Value, queriedReview.Base64Value, "Review content does not match"):
				return
			}
		})

		// succeeds only if the previous test succeeds
		t.Run("Add duplicate Review", func(t *testing.T) {
			if !assert.Equal(t, addReview(validReview, submissionID),
				&DuplicateReviewError{SubmissionID:submissionID, UserID:validReview.ReviewerID}, 
				"Duplicate review addition should cause an error!") {
			}
			// gets the submission metadata and checks that the duplicate review was not added
			queriedMetaData, _ := getSubmissionMetaData(submissionID)
			if !assert.Equal(t, 1, len(queriedMetaData.Reviews), "Duplicate review added to submission!") {
				return
			}
		})
	})

	t.Run("Add non-reviewer review", func(t *testing.T) {
		invalidReview := &Review{
			ReviewerID: globalReviewers[2].ID,
			Approved: true,
			Base64Value: "test",
		}
		assert.Error(t, addReview(invalidReview, submissionID), "No error occurred while adding review for non-reviewer user!")
	})

	t.Run("Non-existant submission", func(t *testing.T) {
		validReview := &Review{
			ReviewerID: globalReviewers[0].ID,
			Approved: true,
			Base64Value: "test",
		}
		assert.Error(t, addReview(validReview, 0), "No error adding review to invalid submission!")
	})

	// uploading reviews to already approved submissions is not allowed behaviour
	t.Run("Submission Approved", func(t *testing.T) {
		// approves submission
		if !assert.NoError(t, gormDb.Model(&Submission{}).Where("ID = ?", submissionID).Update("approved", true).Error, 
			"submission unable to be marked approved") {
			return
		}
		validReview := &Review{
			ReviewerID: globalReviewers[0].ID,
			Approved: true,
			Base64Value: "test",
		}
		assert.Error(t, addReview(validReview, submissionID), "No error adding review to already approved submission!")
	})
}

func TestUpdateSubmissionStatus(t *testing.T) {
	// configures main test environment
	testInit()
	defer testEnd()

	// adds a submission to the db with authors and reviewers
	globalAuthors, globalReviewers, err := initMockUsers(t)
	if err != nil {
		return
	}

	submission := Submission{
		Name:    "Test",
		Authors: []GlobalUser{globalAuthors[0]},
		Reviewers: []GlobalUser{globalReviewers[0]},
		MetaData: &SubmissionData{
			Abstract: "Test",
		},
	}

	submissionID, err := addSubmission(&submission)
	if !assert.NoError(t, err, "Submission creation shouldn't error!") {
		return
	}

	// resets the submission's approval status to nil and uploads a new review with status given as a parameter
	resetSubmission := func(approved bool) {		
		// resets the submission status and reviews
		if !assert.NoError(t, gormDb.Model(&Submission{}).Where("ID = ?", submissionID).Update("approved", nil).Error, 
			"Error while resetting submission status") {
			return
		}
		// resets metadata file and hence reviews
		if !assert.NoError(t, addMetaData(&submission), "could not reset reviews") {
			return
		}
		// adds the reviewers review to the submission
		review := &Review{
			ReviewerID: globalReviewers[0].ID,
			Approved: approved,
			Base64Value: "test",
		}
		if !assert.NoError(t, addReview(review, submissionID), "Review Addition shouldn't error!") {
			return
		}
	}

	// checks that the submission status starts as nil 
	t.Run("Check Status is Nil", func(t *testing.T) {
		testSub := &Submission{}
		if !assert.NoError(t, gormDb.Model(testSub).Select("approved").Find(testSub, submissionID).Error, "submission unable to be retrieved") {
			return 
		}
		assert.Empty(t, testSub.Approved, "submission approval not nil!")
	})

	// runs first so that there are no uploaded reviews
	t.Run("Update status missing reviews", func(t *testing.T) {
		assert.Error(t, updateSubmissionStatus(true, submissionID), "no error for updating submission status of unreviewed submission")
	})

	t.Run("Update status valid", func(t *testing.T) {
		t.Run("Approve Valid", func(t *testing.T) {
			resetSubmission(true)
			assert.NoError(t, updateSubmissionStatus(true, submissionID), "status update should not error")
			submission := &Submission{}
			if !assert.NoError(t, gormDb.Model(&Submission{}).Select("submissions.approved").Find(submission, submissionID).Error, "unable to find submission") {
				return
			}
			assert.Equal(t, true, *submission.Approved, "Submission status not updated properly")
		})

		t.Run("Approve Invalid", func(t *testing.T) {
			resetSubmission(false)
			assert.Error(t, updateSubmissionStatus(true, submissionID), "should not be able to approve submission with disapproving reviews")
		})

		t.Run("Disapprove", func(t *testing.T) {
			// with approving reviews
			// resetSubmission(true)
			assert.NoError(t, updateSubmissionStatus(false, submissionID), "status update should not error")
			submission := &Submission{}
			if !assert.NoError(t, gormDb.Model(&Submission{}).Select("submissions.approved").Find(submission, submissionID).Error, "unable to find submission") {
				return
			}
			assert.Equal(t, false, *submission.Approved, "Submission status not updated properly")

			// with disapproving reviews
			resetSubmission(false) 
			assert.NoError(t, updateSubmissionStatus(false, submissionID), "status update should not error")
			submission = &Submission{}
			if !assert.NoError(t, gormDb.Model(&Submission{}).Select("submissions.approved").Find(submission, submissionID).Error, "unable to find submission") {
				return
			}
			assert.Equal(t, false, *submission.Approved, "Submission status not updated properly")
		})
	})
}
