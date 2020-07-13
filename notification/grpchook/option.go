package grpchook

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EndpointConfig map[string]Config

type EndpointOption func(*Config)
type DecisionFunc func(ctx context.Context, respError error) bool

// The decision func decides if a notification should be sent or not.
// The default func decides this based on the options for each endpoint.
// Set a custom func to decide if a notification should be sent based on other information from the request or
type Config struct {
	DeciderFunc     DecisionFunc
	NotifyOnError   bool
	NotifyOnSuccess bool
	ErrorCodes      []codes.Code
}


// The default options for an endpoint is
// * Send notifications on error only
// * Send notification on all error codes
func Endpoint(funcName string, opts ...EndpointOption) EndpointConfig {
	c := &Config{
		DeciderFunc:     nil,
		NotifyOnError:   true,
		NotifyOnSuccess: false,
	}

	for _, opt := range opts {
		opt(c)
	}

	return EndpointConfig{funcName: *c}
}

// If this is set a slack notification will only be sent if any of these gRPC errors occurs.
// If this is not set a slack message will be sent on any gRPC error that occurs.
func NotifyOnlyOn(codeList []codes.Code) EndpointOption {
	return func(c *Config) {
		c.ErrorCodes = codeList
	}
}

// Set if a message notification should be sent if the request was handled successfully.
func DoNotifyOnSuccess(b bool) EndpointOption {
	return func(c *Config) {
		c.NotifyOnSuccess = b
	}
}

// Set if a message notification should be sent if a gRPC error occurs.
func DoNotifyOnError(b bool) EndpointOption {
	return func(c *Config) {
		c.NotifyOnError = b
	}
}

// This function decides if a slack notification should be sent.
// A default function will be used based on the sendOnError and sendOnInfo values if no customer function is set.
func UseCustomDecisionFunction(f DecisionFunc) EndpointOption {
	return func(c *Config) {
		c.DeciderFunc = f
	}
}

// The default decider func
func (c Config) ShouldSendNotification(ctx context.Context, respError error) bool {
	if respError != nil {
		if c.NotifyOnError {
			if len(c.ErrorCodes) > 0 {
				errStatus, _ := status.FromError(respError)
				respErrCode := errStatus.Code()
				for _, errCode := range c.ErrorCodes {
					if respErrCode == errCode {
						return true
					}
				}
				return false
			}
			return true
		}
		return false
	}

	if c.NotifyOnSuccess {
		return true
	}

	return false
}
