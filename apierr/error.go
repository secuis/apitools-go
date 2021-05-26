package apierr

import (
	"errors"
	"net/http"

	"google.golang.org/grpc/codes"
)

var ErrNotFound = errors.New("not found")
var ErrInvalidFile = errors.New("invalid file")
var ErrNotImplemented = errors.New("not implemented")
var ErrAlreadyExists = errors.New("already exists")
var ErrUnexpected = errors.New("unexpected")
var ErrBadRequest = errors.New("bad request")
var ErrUnauthenticated = errors.New("unauthenticated")
var ErrForbidden = errors.New("forbidden")
var ErrInvalidRequest = errors.New("invalid request")
var ErrValidationFailed = errors.New("validation failed")
var ErrInvalidPassword = errors.New("invalid password")
var ErrFailedToGenerateCredentials = errors.New("failed to generate credentials")

// HTTPStatusCode returns the HTTP status code for the given error
func HTTPStatusCode(err error) int {
	if errors.Is(err, ErrNotFound) {
		return http.StatusNotFound
	}

	// Protobuf storage errors
	// These are possible to remedy by the user so they are marked as 400
	if errors.Is(err, ErrInvalidFile) ||
		errors.Is(err, ErrValidationFailed) ||
		errors.Is(err, ErrInvalidRequest) ||
		errors.Is(err, ErrBadRequest) {
		return http.StatusBadRequest
	}

	// Authentication failures
	if errors.Is(err, ErrInvalidPassword) ||
		errors.Is(err, ErrFailedToGenerateCredentials) ||
		errors.Is(err, ErrUnauthenticated) {
		return http.StatusUnauthorized
	}

	// Authorization failure
	if errors.Is(err, ErrForbidden) {
		return http.StatusForbidden
	}

	// Not implemented error
	if errors.Is(err, ErrNotImplemented) {
		return http.StatusNotImplemented
	}

	// CONFLICT
	if errors.Is(err, ErrAlreadyExists) {
		return http.StatusConflict
	}

	// unexpected
	if errors.Is(err, ErrUnexpected) {
		return http.StatusInternalServerError
	}

	// still unexpected
	return http.StatusInternalServerError
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
		errors.Is(err, ErrBadRequest) ||
		errors.Is(err, ErrInvalidRequest) {
		return codes.InvalidArgument
	}

	// Authentication failures
	if errors.Is(err, ErrInvalidPassword) ||
		errors.Is(err, ErrUnauthenticated) ||
		errors.Is(err, ErrFailedToGenerateCredentials) {
		return codes.Unauthenticated
	}

	// Authorization failure
	if errors.Is(err, ErrForbidden) {
		return codes.PermissionDenied
	}

	// Not implemented error
	if errors.Is(err, ErrNotImplemented) {
		return codes.Unimplemented
	}

	// CONFLICT
	if errors.Is(err, ErrAlreadyExists) {
		return codes.AlreadyExists
	}

	// unexpected
	if errors.Is(err, ErrUnexpected) {
		return codes.Unknown
	}

	// still unexpected
	return codes.Unknown
}
