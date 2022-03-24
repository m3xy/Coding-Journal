// =========================================================================
// approval.go
// Authors: 190010425
// Created: February 28, 2022
//
// This file takes care of everything having to do w/ submission approval
// This includes assignment of reviewers, review upload, and editor approval
// =========================================================================

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

const (
	ENDPOINT_ASSIGN_REVIEWERS = "/assignreviewers"
	ENPOINT_REVIEW            = "/review"
	ENDPOINT_CHANGE_STATUS    = "/approve"
)

// ------------
// Router Functions
// ------------
// actual endpoints defined on the submissions sub-router in submissions.go

// router function to allow journal editors to assign reviewers to a given submission
// uses addReviewers in submissions.go
// POST /submission/{id}/assignreviewers
func PostAssignReviewers(w http.ResponseWriter, r *http.Request) {
	resp := &StandardResponse{}
	reqBody := &AssignReviewersBody{}

	// gets the submission ID from the vars and user details from request context
	params := mux.Vars(r)
	submissionID64, err := strconv.ParseUint(params["id"], 10, 32)
	submissionID := uint(submissionID64)
	if err != nil {
		resp = &StandardResponse{Message: "Given Submission ID not a number.", Error: true}
		w.WriteHeader(http.StatusBadRequest)

	} else if ctx, ok := r.Context().Value("data").(*RequestContext); !ok || validate.Struct(ctx) != nil {
		resp = &StandardResponse{Message: "Request Context not set, user not logged in.", Error: true}
		w.WriteHeader(http.StatusUnauthorized)

	} else if ctx.UserType != USERTYPE_EDITOR { // logged in user must be an editor to assign reviewers
		resp = &StandardResponse{Message: "The client must have editor permissions to assign reviewers.", Error: true}
		w.WriteHeader(http.StatusUnauthorized)

	} else if err := json.NewDecoder(r.Body).Decode(reqBody); err != nil || validate.Struct(reqBody) != nil {
		// request body could not be validated or decoded
		resp = &StandardResponse{Message: "Unable to parse request body.", Error: true}
		w.WriteHeader(http.StatusBadRequest)

	} else if err := assignReviewers(reqBody.Reviewers, submissionID); err != nil {
		switch err.(type) {
		// one of the reviewers is not registered as a user, or does not have proper permissions
		case *BadUserError, *WrongPermissionsError:
			resp = &StandardResponse{Message: err.Error(), Error: true}
			w.WriteHeader(http.StatusBadRequest)

		// editors are not allowed to assign reviewers to submissions which are already accepted or rejected
		case *SubmissionStatusFinalisedError:
			resp = &StandardResponse{Message: err.Error(), Error: true}
			w.WriteHeader(http.StatusUnauthorized)

		default: // Unexpected error - error out as server error.
			log.Printf("[ERROR] could not change submission status: %v\n", err)
			resp = &StandardResponse{Message: "Internal Server Error - could not change submission status", Error: true}
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	// Return response body after function successful.
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] error formatting response: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// router function for reviewer review upload
// POST /submission/{id}/review
func PostUploadReview(w http.ResponseWriter, r *http.Request) {
	reqBody := &UploadReviewBody{}
	resp := &StandardResponse{}

	// gets the submission ID from the vars and user details from request context
	params := mux.Vars(r)
	submissionID64, err := strconv.ParseUint(params["id"], 10, 32)
	submissionID := uint(submissionID64)
	if err != nil {
		resp = &StandardResponse{Message: "Given Submission ID not a number.", Error: true}
		w.WriteHeader(http.StatusBadRequest)

	} else if ctx, ok := r.Context().Value("data").(*RequestContext); !ok || validate.Struct(ctx) != nil {
		// no user is logged in, hence not allowed to upload a review
		resp = &StandardResponse{Message: "Request Context not set, user not logged in.", Error: true}
		w.WriteHeader(http.StatusUnauthorized)

	} else if ctx.UserType != USERTYPE_REVIEWER && ctx.UserType != USERTYPE_REVIEWER_PUBLISHER {
		// currently logged in user does not have reviewer permissions
		resp = &StandardResponse{Message: "The client must have reviewer permissions to upload a review.", Error: true}
		w.WriteHeader(http.StatusUnauthorized)

	} else if json.NewDecoder(r.Body).Decode(reqBody) != nil || validate.Struct(reqBody) != nil {
		// request body is improperly formatted
		resp = &StandardResponse{Message: "Unable to parse request body.", Error: true}
		w.WriteHeader(http.StatusBadRequest)

	} else {
		// request is valid -> build review and add it to the submission
		review := &Review{
			ReviewerID:  ctx.ID,
			Approved:    reqBody.Approved,
			Base64Value: reqBody.Base64Value,
		}
		// adds the review and formats response based upon error type if one occurs
		if err := addReview(review, submissionID); err != nil {
			switch err.(type) {
			case *NotReviewerError: // not a reviewer on the given submission
				resp = &StandardResponse{Message: err.Error(), Error: true}
				w.WriteHeader(http.StatusUnauthorized)

			// each reviewer can only upload a review once for each submission,
			// before the submission has been accepted or rejected
			case *DuplicateReviewError, *SubmissionStatusFinalisedError:
				resp = &StandardResponse{Message: err.Error(), Error: true}
				w.WriteHeader(http.StatusBadRequest)

			default: // Unexpected error - error out as server error.
				log.Printf("[ERROR] could not upload review: %v\n", err)
				resp = &StandardResponse{Message: "Internal Server Error - could not upload review", Error: true}
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	}

	// Return response body after function successful.
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] error formatting response: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// router function for updating submission status (i.e. accepting or rejecting)
// POST /submission/{id}/approve
func PostUpdateSubmissionStatus(w http.ResponseWriter, r *http.Request) {
	resp := &StandardResponse{}
	reqBody := &UpdateSubmissionStatusBody{}

	// gets the submission ID from the vars and user details from request context
	params := mux.Vars(r)
	submissionID64, err := strconv.ParseUint(params["id"], 10, 32)
	submissionID := uint(submissionID64)
	if err != nil {
		resp = &StandardResponse{Message: "Given Submission ID not a number.", Error: true}
		w.WriteHeader(http.StatusBadRequest)

	} else if ctx, ok := r.Context().Value("data").(*RequestContext); !ok || validate.Struct(ctx) != nil {
		// no user logged in, request not allowed
		resp = &StandardResponse{Message: "Request Context not set, user not logged in.", Error: true}
		w.WriteHeader(http.StatusUnauthorized)

	} else if ctx.UserType != USERTYPE_EDITOR {
		// current user is not an editor, cannot accept or reject submission
		resp = &StandardResponse{Message: "The client must have editor permissions to update submission status.", Error: true}
		w.WriteHeader(http.StatusUnauthorized)

	} else if err := json.NewDecoder(r.Body).Decode(reqBody); err != nil || validate.Struct(reqBody) != nil {
		// request body improperly formatted
		resp = &StandardResponse{Message: "Unable to parse request body.", Error: true}
		w.WriteHeader(http.StatusBadRequest)

	} else {
		// changes the submission status. If an error occurs responds according to the type
		if err := updateSubmissionStatus(reqBody.Status, submissionID); err != nil {
			switch err.(type) {
			case *MissingReviewsError, *MissingApprovalError:
				resp = &StandardResponse{Message: err.Error(), Error: true}
				w.WriteHeader(http.StatusUnauthorized)

			default: // Unexpected error - error out as server error.
				log.Printf("[ERROR] could not change submission status: %v\n", err)
				resp = &StandardResponse{Message: "Internal Server Error - could not change submission status", Error: true}
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	}

	// Return response body after function successful.
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[ERROR] error formatting response: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// ------------
// Helper Functions
// ------------

// assigns reviewers to a given submission
//
// Params:
// 	reviewerIDs ([]string) : a list of reviewers IDs to be added to the submission
// 	submissionID (uint) : the ID of the submission which reviewer IDs are to be added to
// Returns:
// 	(error) : an error if one occurs
func assignReviewers(reviewerIDs []string, submissionID uint) error {
	// builds array of reviewers
	reviewers := make([]GlobalUser, len(reviewerIDs))
	for i, reviewerID := range reviewerIDs {
		reviewers[i] = GlobalUser{ID: reviewerID}
	}
	// begins a transaction and checks that the submission has not been approved or disapproved yet
	if err := gormDb.Transaction(func(tx *gorm.DB) error {
		submission := &Submission{}
		if err := tx.Model(&Submission{}).Select("approved").Find(&submission, submissionID).Error; err != nil {
			return err
		} else if submission.Approved != nil {
			return &SubmissionStatusFinalisedError{SubmissionID: submissionID}
		}
		return addReviewers(tx, reviewers, submissionID)
	}); err != nil {
		return err
	}
	return nil
}

// adds a review to a given submission's metadata
//
// Params:
// 	review (*Review) : the review to be added
// 	submissionID (uint) : the id of the submission for the review to be added to
// Return:
// 	(error) : an error if one occurs, nil otherwise
func addReview(review *Review, submissionID uint) error {
	submission, err := getSubmission(submissionID)
	if err != nil {
		return err
	}

	// checks that the given submission has not been approved yet (as approved submissions cannot have new reviews submitted)
	if submission.Approved != nil && *submission.Approved == true {
		return &SubmissionStatusFinalisedError{SubmissionID: submissionID}
	}
	// checks that the reviewer is assigned to the given submission (implicitly checks usertype)
	isReviewer := false
	for _, reviewer := range submission.Reviewers {
		if reviewer.ID == review.ReviewerID {
			isReviewer = true
			break
		}
	}
	if !isReviewer {
		return &NotReviewerError{UserID: review.ReviewerID, SubmissionID: submissionID}
	}
	// checks that the reviewer has not already uploaded a review
	for _, currReview := range submission.MetaData.Reviews {
		if review.ReviewerID == currReview.ReviewerID {
			return &DuplicateReviewError{UserID: review.ReviewerID, SubmissionID: submissionID}
		}
	}

	// adds the review to the given submission
	submission.MetaData.Reviews = append(submission.MetaData.Reviews, review)
	return addMetaData(submission)
}

// approves or dissaproves a given submission by ID
//
// Params:
//	status (bool) : indicates whether submission should be marked accepted or rejected
// 	submissionID (uint) : the submissions unique ID
// Return:
// 	(error) : an error if one occurs, nil otherwise
func updateSubmissionStatus(status bool, submissionID uint) error {
	submission, err := getSubmission(submissionID)
	if err != nil {
		return err
	}

	// maps reviewer ID to review approval status
	reviews := make(map[string]bool)
	if len(submission.Reviewers) == 0 && status { // no reviews uploaded, cannot accept submission
		return &MissingReviewsError{SubmissionID: submissionID}
	}
	for _, review := range submission.MetaData.Reviews {
		reviews[review.ReviewerID] = review.Approved
	}
	// checks that each reviewer has submitted a review
	for _, reviewer := range submission.Reviewers {
		if approved, ok := reviews[reviewer.ID]; !ok {
			// a review is missing -> cannot change submission status
			return &MissingReviewsError{SubmissionID: submissionID}
		} else if !approved && status {
			// a reviewer does not approve, hence the editor cannot accept the submission (but can still reject)
			return &MissingApprovalError{SubmissionID: submissionID}
		}
	}
	// updates the submission to be accepted/rejected
	if err := gormDb.Model(&Submission{}).Where("ID = ?", submissionID).Update("approved", status).Error; err != nil {
		return err
	}
	return nil
}
