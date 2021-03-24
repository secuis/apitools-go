package fieldmaskx

import (
	"context"
	"google.golang.org/grpc/metadata"

	"google.golang.org/protobuf/proto"

	"github.com/mennanov/fmutils"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// Check if a request has a field mask variable
type FieldMaskable interface {
	GetFieldMask() *fieldmaskpb.FieldMask
}

// UnaryServerInterceptor returns a new unary server interceptor for applying request field mask on response.
// The validity of the field mask is not checked in this method, the check need to be implemented by the user.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		// get the response
		resp, err := handler(ctx, req)
		if err != nil {
			return resp, err
		}

		if sub, ok := req.(FieldMaskable); ok {
			// cast to proto message if possible
			protoResp, isProtoResponse := resp.(proto.Message)
			if !isProtoResponse {
				return resp, err
			}

			// filter the response
			fmutils.Filter(protoResp, sub.GetFieldMask().GetPaths())

			// set the filtered response
			resp = protoResp
		}

		return resp, err
	}
}

// StreamServerInterceptor returns a new streaming server interceptor for applying request field mask on response.
// The validity of the field mask is not checked in this method, the check need to be implemented by the user.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		s := &fieldMaskStream{
			wrappedStream: stream,
		}

		return handler(srv, s)
	}
}

// Wraps a StreamServer to filter values with a requested field mask
type fieldMaskStream struct {
	wrappedStream grpc.ServerStream
	mask          *fieldmaskpb.FieldMask
}

func (w *fieldMaskStream) RecvMsg(m interface{}) error {
	if err := w.wrappedStream.RecvMsg(m); err != nil {
		return err
	}

	if sub, ok := m.(FieldMaskable); ok {
		w.mask = sub.GetFieldMask()
	}

	return nil
}

func (w *fieldMaskStream) SendMsg(m interface{}) error {
	protoMsg, isProto := m.(proto.Message)
	if !isProto {
		return w.wrappedStream.SendMsg(m)
	}

	// filter the response
	fmutils.Filter(protoMsg, w.mask.GetPaths())

	// send the filtered response
	return w.wrappedStream.SendMsg(protoMsg)
}

func (w *fieldMaskStream) SetHeader(md metadata.MD) error {
	return w.wrappedStream.SetHeader(md)
}

func (w *fieldMaskStream) SendHeader(md metadata.MD) error {
	return w.wrappedStream.SendHeader(md)
}

func (w *fieldMaskStream) SetTrailer(md metadata.MD) {
	w.wrappedStream.SetTrailer(md)
}

func (w *fieldMaskStream) Context() context.Context {
	return w.wrappedStream.Context()
}
