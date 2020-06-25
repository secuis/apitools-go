package apiError

import "errors"

var ErrNotFound = errors.New("not found")
var ErrInvalidFile = errors.New("invalid file")
var ErrNotImplemented = errors.New("not implemented")
var ErrAlreadyExists = errors.New("already exists")
var ErrUnexpected = errors.New("unexpected")
var ErrInvalidRequest = errors.New("invalid request")
var ErrValidationFailed = errors.New("validation failed")
var ErrInvalidPassword = errors.New("invalid password")
var ErrFailedToGenerateCredentials = errors.New("failed to generate credentials")

// HTTPStatusCode returns the HTTP status code for the given error
func HTTPStatusCode(err error) int {
	if errors.Is(err, ErrNotFound) {
		return 404
	}

	// Protobuf storage errors
	// These are possible to remedy by the user so they are marked as 400
	if errors.Is(err, ErrInvalidFile) ||
		errors.Is(err, ErrValidationFailed) ||
		errors.Is(err, ErrInvalidRequest) {
		return 400
	}

	// Authentication failures
	if errors.Is(err, ErrInvalidPassword) ||
		errors.Is(err, ErrFailedToGenerateCredentials) {
		return 401
	}

	// Not implemented error
	if errors.Is(err, ErrNotImplemented) {
		return 501
	}

	// CONFLICT
	if errors.Is(err, ErrAlreadyExists) {
		return 409
	}

	// unexpected
	if errors.Is(err, ErrUnexpected) {
		return 500
	}

	// still unexpected
	return 500
}
