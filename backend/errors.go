package main

import "fmt"

// ------------------------------------------------------------------- //
// Collection of error types for route and controller error handling.
// Author(s): 190014935
// ------------------------------------------------------------------- //

// -----------
// Submission Errors
// -----------

// submission does not exist
type NoSubmissionError struct {
	ID uint
}

func (e *NoSubmissionError) Error() string {
	return fmt.Sprintf("Submission %d doesn't exist!", e.ID)
}

// submission was marked runnable but has no run.sh file
type SubmissionNotRunnableError struct{}

func (e *SubmissionNotRunnableError) Error() string {
	return "Given submission is missing a run file"
}

// -----------
// File Errors
// -----------

// file does not exist in the db or filesystem
type FileNotFoundError struct {
	ID uint
}

func (e *FileNotFoundError) Error() string {
	return fmt.Sprintf("File %d could not be found!", e.ID)
}

// Handle duplicate files.
type DuplicateFileError struct {
	Path string
}

func (e *DuplicateFileError) Error() string {
	return fmt.Sprintf("Path %s appears more than once!", e.Path)
}

// -----------
// Comments Errors
// -----------

type CommentNotFoundError struct {
	ID uint
}

func (e *CommentNotFoundError) Error() string {
	return fmt.Sprintf("Comment %d does not exist!", e.ID)
}

// -----------
// Approval Errors
// -----------

// handle the case where a user is not assigned as a reviewer for a given submission.
type NotReviewerError struct {
	UserID       string
	SubmissionID uint
}

func (e *NotReviewerError) Error() string {
	return fmt.Sprintf("User: %s is not assigned as reviewer to submission %d", e.UserID, e.SubmissionID)
}

// Handle duplicate reviews.
type DuplicateReviewError struct {
	UserID       string
	SubmissionID uint
}

func (e *DuplicateReviewError) Error() string {
	return fmt.Sprintf("Reviewer %s submitted multiple reviews for submission %d", e.UserID, e.SubmissionID)
}

// handles case where a review is uploaded or reviewer is assigned to an already approved submission
type SubmissionStatusFinalisedError struct {
	SubmissionID uint
}

func (e *SubmissionStatusFinalisedError) Error() string {
	return fmt.Sprintf("Cannot perform this action on approved/disapproved submission: %d", e.SubmissionID)
}

// handle case where an editor tries to change status of a submission without all reviews being submitted first
type MissingReviewsError struct {
	SubmissionID uint
}

func (e *MissingReviewsError) Error() string {
	return fmt.Sprintf("Cannot change status of submission: %d as it is missing reviews", e.SubmissionID)
}

// handle case where an editor tries to approve a submission without all reviews being approving reviews
type MissingApprovalError struct {
	SubmissionID uint
}

func (e *MissingApprovalError) Error() string {
	return fmt.Sprintf("Cannot change status of submission: %d as not all reviewers approve", e.SubmissionID)
}

// -----------
// Authentication/User Errors
// -----------

// Handle non-registered users.
type BadUserError struct {
	userID string
}

func (e *BadUserError) Error() string {
	return "User " + e.userID + " doesn't exist!"
}

// Handle repeat emails
type RepeatEmailError struct {
	email string
}

func (e *RepeatEmailError) Error() string {
	return "Email " + e.email + " is already taken!"
}

// Handle users with incorrect permissions.
type WrongPermissionsError struct {
	userID string
}

func (e *WrongPermissionsError) Error() string {
	return "User" + e.userID + "does not have required permissions!"
}

// -----------
// Misc Errors
// -----------

// Handles the case where the result of a given query is empty
type ResultSetEmptyError struct{}

func (e *ResultSetEmptyError) Error() string {
	return fmt.Sprintf("No results returned for the given query")
}

// Handle a badly formatted/illegal query parameter value
type BadQueryParameterError struct {
	ParamName string
	Value     interface{}
}

func (e *BadQueryParameterError) Error() string {
	return fmt.Sprintf("Illegal query parameter. %s cannot take value %v", e.ParamName, e.Value)
}
