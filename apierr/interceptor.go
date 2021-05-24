package apierr

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// HandleErrFunc is a function that transforms the error in some way
type HandleErrFunc func(error) error

type Option func(*options)

type options struct {
	handleErr HandleErrFunc
}

func WithErrTranslation(f HandleErrFunc) Option {
	return func(o *options) {
		o.handleErr = f
	}
}

var defaultOptions = options{
	handleErr: func(err error) error {
		return status.Errorf(GRPCCode(err), err.Error())
	},
}

// UnaryServerInterceptor returns a new unary server interceptor for error translation.
func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		resp, err := handler(ctx, req)
		if err != nil {
			err = options.handleErr(err)
		}
		return resp, err
	}
}

// StreamServerInterceptor returns a new streaming server interceptor for error translation.
func StreamServerInterceptor(opts ...Option) grpc.StreamServerInterceptor {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}

	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		err = handler(srv, stream)
		if err != nil {
			err = options.handleErr(err)
		}
		return err
	}
}
