package handler

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	integration "github.com/jyouturer/gmail-ai/integrations"
)

// NewMockClassifierServer creates a new mocked gRPC server for the Classifier service
func NewMockClassifierServer() (*grpc.Server, *bufconn.Listener) {
	// Create an in-memory connection for the gRPC server
	lis := bufconn.Listen(1024 * 1024)

	// Create a new gRPC server and register the mocked Classifier service
	s := grpc.NewServer()
	integration.RegisterClassifierServer(s, &mockClassifierServer{})

	// Start the server
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(fmt.Sprintf("failed to serve: %v", err))
		}
	}()

	return s, lis
}

// mockClassifierServer is an implementation of the ClassifierServer for testing purposes
type mockClassifierServer struct {
	integration.UnimplementedClassifierServer
}

// ClassifyEmail is a mocked implementation of the ClassifyEmail RPC
func (m *mockClassifierServer) ClassifyEmail(ctx context.Context, req *integration.ClassifyRequest) (*integration.ClassifyResponse, error) {
	// Do some basic validation of the input request
	if req == nil || req.EmailText == "" {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// Mock the response
	//fmt.Println(req.EmailText)
	if strings.Contains(req.EmailText, "rejected") {
		//fmt.Println("rejected")
		return &integration.ClassifyResponse{IsRejection: true}, nil
	}
	return &integration.ClassifyResponse{IsRejection: false}, nil
}

func NewMockConnectionPool() (*integration.ConnectionPool, error) {
	size := 1
	timeout := time.Second * 10

	// Create a buffer listener
	lis := bufconn.Listen(1024 * 1024)

	// Create a gRPC server
	grpcServer := grpc.NewServer()

	// Register the classifier service with the gRPC server
	integration.RegisterClassifierServer(grpcServer, &mockClassifierServer{})

	// Start the gRPC server
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Create a connection pool with 1 gRPC connection object
	pool := make(chan *integration.GRPCClient, size)

	// Connect to the gRPC server using the buffer listener
	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(func(ctx context.Context, address string) (net.Conn, error) {
		return lis.Dial()
	}), grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to dial bufnet: %v", err)
	}

	// Create a gRPC client object
	client := integration.NewClassifierClient(conn)

	// Add the gRPC client object to the connection pool
	rc := &integration.GRPCClient{
		Client: client,
	}
	pool <- rc

	return &integration.ConnectionPool{
		Pool:    pool,
		Timeout: timeout,
	}, nil
}

func TestRejectionChecker_IsRejection(t *testing.T) {
	// Create a mocked gRPC server for testing
	grpcServer, _ := NewMockClassifierServer()
	defer grpcServer.Stop()

	// Create a mocked gRPC connection pool
	grpcConnectionPool, err := NewMockConnectionPool()
	if err != nil {
		t.Errorf("failed to create mocked connection pool: %v", err)
	}

	// Create a new RejectionChecker with the mocked connection pool
	rc := &RejectionChecker{
		GRPCClientPool: grpcConnectionPool,
	}

	// Test a rejection email
	isRejection, err := rc.IsRejection(context.Background(), "Your application has been rejected")
	assert.NoError(t, err)
	assert.True(t, isRejection)

	// Test a non-rejection email
	isRejection, err = rc.IsRejection(context.Background(), "Thank you for your application")
	assert.NoError(t, err)
	assert.False(t, isRejection)
}
