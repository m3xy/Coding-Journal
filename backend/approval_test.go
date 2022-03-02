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
	// "io/ioutil"
	"net/http"
	"net/http/httptest"
	// "os"
	// "path/filepath"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	// "gorm.io/gorm/clause"
)

// ------------
// Router Function Tests
// ------------

func TestUploadReview(t *testing.T) {
	// configures main test environment
	testInit()
	defer testEnd()

	// Create mux router
	router := mux.NewRouter()
	route := ENDPOINT_SUBMISSIONS+"/{id}"+ENPOINT_REVIEW
	router.HandleFunc(route, uploadReview)

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

			ctx := context.WithValue(req.Context(), "userId", globalReviewers[0].ID)
			ctx = context.WithValue(ctx, "userType", globalReviewers[0].UserType)
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

			ctx := context.WithValue(req.Context(), "userId", globalReviewers[0].ID)
			ctx = context.WithValue(ctx, "userType", globalReviewers[0].UserType)
			router.ServeHTTP(w, req.WithContext(ctx))
			resp := w.Result()

			// makes sure the request succeeded
			if !assert.Equalf(t, http.StatusBadRequest, resp.StatusCode, "duplicate review added without proper error code!") {
				return
			}
		})
	})

	// deals with error cases that occur before the review is added to the submission
	t.Run("Request Validation", func(t *testing.T) {
		// uploads a given review and returns the status code
		uploadReview := func(reqStruct *UploadReviewBody, reviewerID string, userType int, submissionID uint) int {
			reqBody, err := json.Marshal(reqStruct)
			if !assert.NoError(t, err, "Error while marshalling review upload body!") {
				return -1
			}
			// sends the request to upload a review
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("%s/%d%s", ENDPOINT_SUBMISSIONS, submissionID, ENPOINT_REVIEW), bytes.NewBuffer(reqBody))
			w := httptest.NewRecorder()

			ctx := context.WithValue(req.Context(), "userId", reviewerID)
			ctx = context.WithValue(ctx, "userType", userType)
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
			if !assert.Equalf(t, http.StatusUnauthorized, uploadReview(reqStruct, "", USERTYPE_REVIEWER, submissionID), 
				"unauthenticated user added review without proper error code!") {
				return
			}
		})

		t.Run("Review no content", func(t *testing.T) {
			reqStruct := &UploadReviewBody{
				Approved: true,
			}
			// makes sure the request failed StatusBadRequest
			if !assert.Equalf(t, http.StatusBadRequest, uploadReview(reqStruct, globalReviewers[0].ID, globalReviewers[0].UserType, submissionID), 
				"empty review added without proper error code!") {
				return
			}
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
		Reviewers: []GlobalUser{globalReviewers[0]},
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
			ReviewerID: globalReviewers[1].ID,
			Approved: true,
			Base64Value: "test",
		}
		if !assert.Error(t, addReview(invalidReview, submissionID), "No error occurred while adding review for non-reviewer user!") {
			return
		}
	})

	t.Run("Non-existant submission", func(t *testing.T) {
		validReview := &Review{
			ReviewerID: globalReviewers[0].ID,
			Approved: true,
			Base64Value: "test",
		}
		if !assert.Error(t, addReview(validReview, 0), "No error adding review to invalid submission!") {
			return
		}
	})
}