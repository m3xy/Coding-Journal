package main

import "fmt"

// ------------------------------------------------------------------- //
// Collection of error types for route and controller error handling.
// Author(s): 190014935
// ------------------------------------------------------------------- //

// Handle non-registered users.
type BadUserError struct {
	userID string
}

func (e *BadUserError) Error() string { return "User " + e.userID + " doesn't exist!" }

// Handle users with incorrect permissions.
type WrongPermissionsError struct {
	userID string
}

func (e *WrongPermissionsError) Error() string {
	return "User" + e.userID + "does not have required permissions!"
}

// Handle non-existant submissions.
type NoSubmissionError struct {
	ID uint
}

func (e *NoSubmissionError) Error() string { return fmt.Sprintf("Submission %d doesn't exist!", e.ID) }

// handle the case where a user is not assigned as a reviewer for a given submission.
type NotReviewerError struct {
	UserID string
	SubmissionID uint
}

func (e *NotReviewerError) Error() string {
	return fmt.Sprintf("User: %s is not assigned as reviewer to submission %d", e.UserID, e.SubmissionID)
}

// Handle duplicate reviews.
type DuplicateReviewError struct {
	UserID string
	SubmissionID uint
}

func (e *DuplicateReviewError) Error() string {
	return fmt.Sprintf("Reviewer %s submitted multiple reviews for submission %d", e.UserID, e.SubmissionID)
}

// Handle duplicate files.
type DuplicateFileError struct {
	Path string
}

func (e *DuplicateFileError) Error() string {
	return fmt.Sprintf("Path %s appears more than once!", e.Path)
}
