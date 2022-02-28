// =======================================================================
// approval.go
// Authors: 190010425
// Created: February 28, 2022
//
// This file takes care of everything having to do w/ submission approval
// This includes review upload, editor approval, etc.
// =======================================================================

package main

// ------------
// Router Functions
// ------------

// ------------
// Helper Functions
// ------------

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

	// checks that the reviewer is assigned to the given submission
	isReviewer := false
	for _, reviewer := range submission.Reviewers {
		if reviewer.ID == review.ReviewerID {
			isReviewer = true
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