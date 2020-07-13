# API Tools

A collection of tools we use in our Golang APIs

# The gRPC notification hook
The gRPC notification hook package can be used to send messages on different channels when an endpoint is called. It can be restricted to only send notifications when an error, or only when specific errors, occurred.

The default for endpoints is to send notifications on all errors, and not on successful requests.
```
notificationChannels := []notification.Sender{} // Add the channels where you want to send notifications
grpcServer := grpc.NewServer(
    grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
        grpchook.UnaryNotificationInterceptor(notificationChannels, grpchook.Endpoint("gRPCEndpointName", ...), ...),
    )),
    grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
        grpchook.StreamNotificationInterceptor(notificationChannels, grpchook.Endpoint("gRPCStreamEndpointName", ...), ...),
    )),
)
```

It is possible to add options for each endpoint to change the default behaviour.
Available options for an endpoint:
`NotifyOnlyOn(codeList []codes.Code)` \
`DoNotifyOnSuccess(b bool)` \
`DoNotifyOnError(b bool)` \
`UseCustomDecisionFunction(f DecisionFunc)` 

The function that decides if a notification should be sent or not is called DecisionFunc. You can add your own for each endpoint if you want to make the decision based on something else than the configurations available.
The decision function signature looks like this.
```
type DecisionFunc func(ctx context.Context, respError error) bool
```

### Examples
Add options to an endpoint:
```
grpchook.UnaryNotificationInterceptor(notificationChannels, grpchook.Endpoint("gRPCEndpointName", grpchook.NotifyOnlyOn([]codes.Code{codes.Internal, codes.InvalidArgument}), ...))
```

Add more endpoint configurations:
```
grpchook.UnaryNotificationInterceptor(notificationChannels, grpchook.Endpoint("oneEndpoint", ...), grpchook.Endpoint("anotherEndpoint", ...), ...)
```