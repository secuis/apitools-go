package grpchook

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EndpointConfig map[string]Config
type EndpointOption func(*Config)

// ShouldNotifyFunc determines whether a notification should happen for a given error.
type ShouldNotifyForErrFunc func(ctx context.Context, respError error) bool

type Config struct {
	CustomShouldNotify ShouldNotifyForErrFunc
	SkipErrors         bool
	NotifyOnSuccess    bool
	ErrorCodes         []codes.Code
}

// The default options for an endpoint is
// * Send notifications on error only
// * Send notification on all error codes
func NewEndpointConfig(funcName string, opts ...EndpointOption) EndpointConfig {
	c := Config{
		CustomShouldNotify: nil,
		SkipErrors:         false,
		NotifyOnSuccess:    false,
	}

	for _, opt := range opts {
		opt(&c)
	}

	return EndpointConfig{funcName: c}
}

// If this is set a slack notification will only be sent if any of these gRPC errors occurs.
// If this is not set a slack message will be sent on any gRPC error that occurs.
func NotifyOnlyOn(codes []codes.Code) EndpointOption {
	return func(c *Config) {
		c.ErrorCodes = codes
	}
}

// Set if a message notification should be sent if the request was handled successfully.
func DoNotifyOnSuccess() EndpointOption {
	return func(c *Config) {
		c.NotifyOnSuccess = true
	}
}

// Set to skip message notifications when a gRPC error occurs.
func DoSkipErrors() EndpointOption {
	return func(c *Config) {
		c.SkipErrors = true
	}
}

// UseShouldNotifyForErrFunc will use the provided function to determine whether a
// a slack notification should be sent for a given error.
func UseShouldNotifyForErrFunc(f ShouldNotifyForErrFunc) EndpointOption {
	return func(c *Config) {
		c.CustomShouldNotify = f
	}
}

// ShouldNotify determines whether a notification should be sent for the provided error.
// The default behaviour can be overridden by setting a custom decision function with
// UseShouldNotifyForErrFunc(ShouldNotify).
func (c Config) ShouldNotifyForErr(ctx context.Context, err error) bool {
	if c.CustomShouldNotify != nil {
		return c.CustomShouldNotify(ctx, err)
	}

	if err == nil {
		return c.NotifyOnSuccess
	}

	if c.SkipErrors {
		return false
	}

	// Return true on all errors if no specific error codes have been set
	if len(c.ErrorCodes) == 0 {
		return true
	}

	// Determine whether the error is in the list of error codes
	errStatus, _ := status.FromError(err)
	respErrCode := errStatus.Code()
	for _, errCode := range c.ErrorCodes {
		if respErrCode == errCode {
			return true
		}
	}
	return false
}
