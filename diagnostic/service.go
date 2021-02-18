package diagnostic

import (
	"context"
	"go.uber.org/zap"
)

type DiagnosticService struct {
	log *zap.SugaredLogger
	UnimplementedDiagnosticServiceServer
}

func NewDiagnosticService(log *zap.SugaredLogger) DiagnosticService {
	return DiagnosticService{
		log: log,
	}
}

func (s DiagnosticService) Ping(ctx context.Context, request *PingRequest) (*PingResponse, error) {
	return &PingResponse{Answer: "pong"}, nil
}
