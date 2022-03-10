// =======================================================================
// approval.go
// Authors: 190010425
// Created: February 28, 2022
//
// This file takes care of everything having to do w/ submission approval
// This includes review upload, editor approval, etc.
// =======================================================================

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"gorm.io/gorm"
	"github.com/gorilla/mux"
)

const (
	ENDPOINT_ASSIGN_REVIEWERS = "/assignreviewers"
	ENPOINT_REVIEW = "/review"
	ENDPOINT_APPROVE = "/approve"
)

// ------------
// Router Functions
// ------------

// router function to allow journal editors to assign reviewers to a given submission
// uses addReviewers in submissions.go
func RouteAssignReviewers(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] AssignReviewers request received from %v", r.RemoteAddr)
	resp := &StandardResponse{}
	reqBody := &AssignReviewersBody{}

	// gets the submission ID from the vars and user details from request context
	params := mux.Vars(r)
	submissionID64, err := strconv.ParseUint(params["id"], 10, 32)
	submissionID := uint(submissionID64)
	if err != nil {
		resp = &StandardResponse{Message: "Given Submission ID not a number.", Error: true}
		w.WriteHeader(http.StatusBadRequest)

	// gets context struct and validates it
	} else if ctx, ok := r.Context().Value("data").(RequestContext); !ok || validate.Struct(ctx) != nil {
		resp = &StandardResponse{Message: "Request Context not set, user not logged in.", Error: true}
		w.WriteHeader(http.StatusUnauthorized)

	// checks that the client has the proper permisssions (i.e. is an editor)
	} else if ctx.UserType != USERTYPE_EDITOR {
		resp = &StandardResponse{Message: "The client must have editor permissions to assign reviewers.", Error: true}
		w.WriteHeader(http.StatusUnauthorized)

	// decodes request body and validates it
	} else if err := json.NewDecoder(r.Body).Decode(reqBody); err != nil || validate.Struct(reqBody) != nil {
		resp = &StandardResponse{Message: "Unable to parse request body.", Error: true}
		w.WriteHeader(http.StatusBadRequest)

	// adds reviewers and handles error cases
	} else {
		if err := assignReviewers(reqBody.Reviewers, submissionID); err != nil {
			switch err.(type) {
			// one of the reviewers is not registered as a user, or does not have proper permissions
			case *BadUserError, *WrongPermissionsError: 
				resp = &StandardResponse{Message: err.Error(), Error: true}
				w.WriteHeader(http.StatusBadRequest)

			case *SubmissionStatusFinalisedError:
				resp = &StandardResponse{Message: err.Error(), Error: true}
				w.WriteHeader(http.StatusConflict)

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
	} else if !resp.Error {
		log.Print("[INFO] AssignReviewers request successful\n")
	}
}


// router function for reviewer review upload
func RouteUploadReview(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] RouteUploadReview request received from %v", r.RemoteAddr)
	resp := &StandardResponse{}
	reqBody := &UploadReviewBody{}

	// gets the submission ID from the vars and user details from request context
	params := mux.Vars(r)
	submissionID64, err := strconv.ParseUint(params["id"], 10, 32)
	submissionID := uint(submissionID64)
	if err != nil {
		resp = &StandardResponse{Message: "Given Submission ID not a number.", Error: true}
		w.WriteHeader(http.StatusBadRequest)

	// gets context struct and validates it
	} else if ctx, ok := r.Context().Value("data").(RequestContext); !ok || validate.Struct(ctx) != nil {
		resp = &StandardResponse{Message: "Request Context not set, user not logged in.", Error: true}
		w.WriteHeader(http.StatusUnauthorized)

	// checks that the client has the proper permisssions
	} else if ctx.UserType != USERTYPE_REVIEWER && ctx.UserType != USERTYPE_REVIEWER_PUBLISHER {
		resp = &StandardResponse{Message: "The client must have reviewer permissions to upload a review.", Error: true}
		w.WriteHeader(http.StatusUnauthorized)

	// decodes request body and validates it
	} else if json.NewDecoder(r.Body).Decode(reqBody) != nil || validate.Struct(reqBody) != nil {
		resp = &StandardResponse{Message: "Unable to parse request body.", Error: true}
		w.WriteHeader(http.StatusBadRequest)

	// adds review and handles error cases
	} else {
		// builds review to add now that all fields have been validated
		review := &Review{
			ReviewerID: ctx.ID,
			Approved: reqBody.Approved,
			Base64Value: reqBody.Base64Value,
		}
		// adds the review and formats response based upon error type if one occurs
		if err := addReview(review, submissionID); err != nil {
			switch err.(type) {
			case *NotReviewerError:
				resp = &StandardResponse{Message: err.Error(), Error: true}
				w.WriteHeader(http.StatusUnauthorized)

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
	} else if !resp.Error {
		log.Print("[INFO] RouteUploadReview request successful\n")
	}
}


// router function for updating submission status
func RouteUpdateSubmissionStatus(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] updateSubmissionStatus request received from %v", r.RemoteAddr)
	resp := &StandardResponse{}
	reqBody := &UpdateSubmissionStatusBody{}

	// gets the submission ID from the vars and user details from request context
	params := mux.Vars(r)
	submissionID64, err := strconv.ParseUint(params["id"], 10, 32)
	submissionID := uint(submissionID64)
	if err != nil {
		resp = &StandardResponse{Message: "Given Submission ID not a number.", Error: true}
		w.WriteHeader(http.StatusBadRequest)

	// gets context struct and validates it
	} else if ctx, ok := r.Context().Value("data").(RequestContext); !ok || validate.Struct(ctx) != nil {
		resp = &StandardResponse{Message: "Request Context not set, user not logged in.", Error: true}
		w.WriteHeader(http.StatusUnauthorized)

	// checks that the client has the proper permisssions
	} else if ctx.UserType != USERTYPE_EDITOR {
		resp = &StandardResponse{Message: "The client must have editor permissions to update submission status.", Error: true}
		w.WriteHeader(http.StatusUnauthorized)

	// decodes request body and validates it
	} else if err := json.NewDecoder(r.Body).Decode(reqBody); err != nil || validate.Struct(reqBody) != nil {
		resp = &StandardResponse{Message: "Unable to parse request body.", Error: true}
		w.WriteHeader(http.StatusBadRequest)

	// adds review and handles error cases
	} else {
		// changes the submission status
		if err := updateSubmissionStatus(reqBody.Status, submissionID); err != nil {
			switch err.(type) {
			case *MissingReviewsError, *MissingApprovalError:
				resp = &StandardResponse{Message: err.Error(), Error: true}
				w.WriteHeader(http.StatusConflict)

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
	} else if !resp.Error {
		log.Print("[INFO] updateSubmissionStatus request successful\n")
	}
}


// ------------
// Helper Functions
// ------------

func assignReviewers(reviewerIDs []string, submissionID uint) error {
	// builds array of reviewers
	reviewers := make([]GlobalUser, len(reviewerIDs))
	for i, reviewerID := range reviewerIDs {
		reviewers[i] = GlobalUser{ ID: reviewerID }
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
	// gets a given submission, and adds the review to it
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
	// checks that the reviewer has not already uploaded a review (loop is ok here because the number of reviews is small)
	for _, currReview := range submission.MetaData.Reviews {
		if review.ReviewerID == currReview.ReviewerID {
			return &DuplicateReviewError{ UserID:review.ReviewerID, SubmissionID: submissionID }
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
	// checks that all of the submission's reviewers have uploaded reviews
	submission, err := getSubmission(submissionID)
	if err != nil {
		return err
	}

	// maps reviewer ID to review approval status
	reviews := make(map[string]bool)
	for _, review := range submission.MetaData.Reviews {
		reviews[review.ReviewerID] = review.Approved
	}
	// checks that each reviewer has submitted a review
	for _, reviewer := range submission.Reviewers {
		// all reviewers must submit reviews before the submission status is changed
		if approved, ok := reviews[reviewer.ID]; !ok {
			return &MissingReviewsError{SubmissionID: submissionID}
		// cannot approve a submission if any reviewer has not yet approved it
		} else if !approved && status {
			return &MissingApprovalError{SubmissionID: submissionID}
		}
	}
	// updates the submission to be approved/dissaproved
	if err := gormDb.Model(&Submission{}).Where("ID = ?", submissionID).Update("approved", status).Error; err != nil {
		return err
	}
	return nil
}