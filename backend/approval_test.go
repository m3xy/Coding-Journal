// =====================================
// approval_test.go
// Authors: 190010425
// Created: February 28, 2022
//
// test file for approval.go
// =====================================

package main 

import (
	// "bytes"
	// "context"
	// "encoding/json"
	// "fmt"
	// "io/ioutil"
	// "net/http"
	// "net/http/httptest"
	// "os"
	// "path/filepath"
	"testing"

	// "github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	// "gorm.io/gorm/clause"
)

// ------------
// Router Function Tests
// ------------

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
		if !assert.Error(t, addReview(invalidReview, submissionID), "No error !") {
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