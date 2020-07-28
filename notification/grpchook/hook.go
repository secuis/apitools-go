package grpchook

import (
	"context"
	"fmt"
	"strings"

	"github.com/SecuritasCrimePrediction/apitools-go/notification"
	"google.golang.org/grpc"
)

type interceptor struct {
	configs    EndpointConfig
	recipients []notification.Sender
}

// The hook for streaming endpoints
// Assembles all endpoint configurations to one EndpointConfig and returns the hook function
func StreamNotificationInterceptor(recipients []notification.Sender, configs ...EndpointConfig) grpc.StreamServerInterceptor {
	endpointConfigs := EndpointConfig{}
	for _, conf := range configs {
		for k, v := range conf {
			endpointConfigs[k] = v
		}
	}
	i := interceptor{
		recipients: recipients,
		configs:    endpointConfigs,
	}
	return i.streamHook
}

// The hook function for streaming endpoints
// It will call the deciding function to decide if a notification should be sent or not.
func (i interceptor) streamHook(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	endpointParts := strings.Split(info.FullMethod, "/")
	endpointName := endpointParts[len(endpointParts)-1]

	// handle request
	err = handler(srv, stream)

	// if no configuration exist for this endpoint, just return the response
	conf, exists := i.configs[endpointName]
	if !exists {
		return
	}

	// Determine whether a notification should be sent
	if conf.ShouldNotifyForErr(stream.Context(), err) {
		switch err {
		case nil:
			i.sendInfoMsg(endpointName, "")
		default:
			i.sendErrorMsg(endpointName, err.Error())
		}
	}
	return
}

// The hook for unary endpoints
// Assembles all endpoint configurations to one EndpointConfig and returns the hook function
func UnaryNotificationInterceptor(recipients []notification.Sender, configs ...EndpointConfig) grpc.UnaryServerInterceptor {
	endpointConfigs := EndpointConfig{}
	for _, conf := range configs {
		for k, v := range conf {
			endpointConfigs[k] = v
		}
	}
	i := interceptor{
		recipients: recipients,
		configs:    endpointConfigs,
	}
	return i.unaryHook
}

// The hook function for unary endpoints
// It will call the deciding function to decide if a notification should be sent or not.
func (i interceptor) unaryHook(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	endpointParts := strings.Split(info.FullMethod, "/")
	endpointName := endpointParts[len(endpointParts)-1]

	// handle request
	resp, err = handler(ctx, req)

	// if no configuration exist for this endpoint, just return the response
	conf, exist := i.configs[endpointName]
	if !exist {
		return
	}

	// Determine whether a notification should be sent
	if conf.ShouldNotifyForErr(ctx, err) {
		switch err {
		case nil:
			i.sendInfoMsg(endpointName, resp)
		default:
			i.sendErrorMsg(endpointName, err.Error())
		}
	}
	return
}

// Send an info message on all notification channels
func (i interceptor) sendInfoMsg(methodName string, info interface{}) {
	for _, recipient := range i.recipients {
		recipient.Info(fmt.Sprintf("Request to endpoint %s recieved\nExtra info: %+v", methodName, info))
	}
}

// Send an error message on all notification channels
func (i interceptor) sendErrorMsg(methodName string, errStr string) {
	for _, recipient := range i.recipients {
		recipient.Alert(fmt.Sprintf("Error occurred in a call to %s\nError: %s", methodName, errStr))
	}
}
