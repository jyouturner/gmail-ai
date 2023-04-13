package process

import (
	"context"
	"fmt"
	"time"

	integration "github.com/jyouturer/gmail-ai/integrations"
	"github.com/jyouturer/gmail-ai/internal/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// RejectionChecker check whether the given text is rejection or not by calling gRPC service
type RejectionChecker struct {
	GRPCClientPool *integration.ConnectionPool
}

// NewRejectionChecker creates a new RejectionChecker, it will initiate the gRPC connection pool, and return a function to close the connection pool
func NewRejectionChecker(grpcUrl string, grpcConnectionNumber int, grpcTimeoutSeconds int) (*RejectionChecker, func() error, error) {
	// Create a connection pool with 10 grpc connection objects
	cp, err := integration.NewConnectionPool(grpcUrl, grpcConnectionNumber, time.Duration(grpcTimeoutSeconds))
	if err != nil {
		logging.Logger.Error("Error creating connection pool: %v", zap.Error(err))
		return nil, nil, err
	}

	return &RejectionChecker{
			GRPCClientPool: cp,
		}, func() error {
			cp.Close()
			return nil
		}, nil
}

// IsRejection check whether the given text is rejection or not
func (h *RejectionChecker) IsRejection(ctx context.Context, text string) (bool, error) {
	req := &integration.ClassifyRequest{EmailText: text}
	res, err := h.checkRejectionGrpc(ctx, req)
	if err != nil {
		return false, fmt.Errorf("error calling IsRejection gRPC: %v", err)
	}
	return res.IsRejection, nil
}

// checkRejectionGrpc check whether the given text is rejection or not by calling gRPC service
func (h *RejectionChecker) checkRejectionGrpc(ctx context.Context, req *integration.ClassifyRequest, opts ...grpc.CallOption) (*integration.ClassifyResponse, error) {
	rc, err := h.GRPCClientPool.GetGRPCClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get rejection check from pool %v", err)
	}
	defer h.GRPCClientPool.ReturnGRPCClient(rc)

	res, err := rc.Client.ClassifyEmail(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error calling IsRejection gRPC: %v", err)
	}
	return res, nil
}
