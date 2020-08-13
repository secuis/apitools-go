package apierr

import (
	"errors"

	"google.golang.org/grpc/codes"
)

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

// GRPCCode returns the gRPC status code for an error.
// Done in accordance with https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto
func GRPCCode(err error) codes.Code {
	if errors.Is(err, ErrNotFound) {
		return codes.NotFound
	}

	// Protobuf storage errors
	// These are possible to remedy by the user so they are marked as 400
	if errors.Is(err, ErrInvalidFile) ||
		errors.Is(err, ErrValidationFailed) ||
		errors.Is(err, ErrInvalidRequest) {
		return codes.InvalidArgument
	}

	// Authentication failures
	if errors.Is(err, ErrInvalidPassword) ||
		errors.Is(err, ErrFailedToGenerateCredentials) {
		return codes.Unauthenticated
	}

	// Not implemented error
	if errors.Is(err, ErrNotImplemented) {
		return codes.Unimplemented
	}

	// CONFLICT
	if errors.Is(err, ErrAlreadyExists) {
		return codes.Aborted
	}

	// unexpected
	if errors.Is(err, ErrUnexpected) {
		return codes.Unknown
	}

	// still unexpected
	return codes.Unknown
}
