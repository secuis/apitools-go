package grpchook

import (
	"context"
	"fmt"
	"github.com/SecuritasCrimePrediction/apitools-go/notification"
	"google.golang.org/grpc"
	"strings"
)

type interceptor struct {
	configs    EndpointConfig
	channels []notification.Sender
}

// The hook for streaming endpoints
// Assembles all endpoint configurations to one EndpointConfig and returns the hook function
func StreamNotificationInterceptor(ch []notification.Sender, configs ...EndpointConfig) grpc.StreamServerInterceptor {
	endpointConfigs := EndpointConfig{}
	for _, conf := range configs {
		for k, v := range conf {
			endpointConfigs[k] = v
		}
	}
	i := interceptor{
		channels: ch,
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

	conf, exist := i.configs[endpointName]
	// if no configuration exist for this endpoint, just return the response
	if !exist {
		// no notification configuration set for this endpoint,
		// return response without sending any notifications
		return
	}

	// if a custom decider func has been set, use it
	if conf.DeciderFunc != nil && conf.DeciderFunc(stream.Context(), err) {
		switch err {
		case nil:
			i.sendInfoMsg(endpointName, "")
		default:
			i.sendErrorMsg(endpointName, err.Error())
		}
		return
	}

	// use the default decider func which is basing the decision on the endpoint configuration.
	if conf.ShouldSendNotification(stream.Context(), err) {
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
func UnaryNotificationInterceptor(ch []notification.Sender, configs ...EndpointConfig) grpc.UnaryServerInterceptor {
	endpointConfigs := EndpointConfig{}
	for _, conf := range configs {
		for k, v := range conf {
			endpointConfigs[k] = v
		}
	}
	i := interceptor{
		channels: ch,
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

	conf, exist := i.configs[endpointName]
	// if no configuration exist for this endpoint, just return the response
	if !exist {
		// no notification configuration set for this endpoint,
		// return response without sending any notifications
		return
	}

	// if a custom decider func has been set, use it
	if conf.DeciderFunc != nil && conf.DeciderFunc(ctx, err) {
		switch err {
		case nil:
			i.sendInfoMsg(endpointName, resp)
		default:
			i.sendErrorMsg(endpointName, err.Error())
		}
		return
	}

	// use the default decider func which is basing the decision on the endpoint configuration.
	if conf.ShouldSendNotification(ctx, err) {
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
	for _, ch := range i.channels {
		ch.Info(fmt.Sprintf("Request to endpoint %s recieved\nExtra info: %+v", methodName, info))
	}
}

// Send an error message on all notification channels
func (i interceptor) sendErrorMsg(methodName string, errStr string) {
	for _, ch := range i.channels {
		ch.Alert(fmt.Sprintf("Error occurred in a call to %s\nError: %s", methodName, errStr))
	}
}
