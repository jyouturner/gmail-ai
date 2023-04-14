package integration

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
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

// NewMockClassifierServer creates a new mocked gRPC server for the Classifier service
func NewMockClassifierServer() (*grpc.Server, *bufconn.Listener) {
	// Create an in-memory connection for the gRPC server
	lis := bufconn.Listen(1024 * 1024)

	// Create a new gRPC server and register the mocked Classifier service
	s := grpc.NewServer()
	RegisterClassifierServer(s, &mockClassifierServer{})

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
	UnimplementedClassifierServer
}

// ClassifyEmail is a mocked implementation of the ClassifyEmail RPC
func (m *mockClassifierServer) ClassifyEmail(ctx context.Context, req *ClassifyRequest) (*ClassifyResponse, error) {
	// Do some basic validation of the input request
	if req == nil || req.EmailText == "" {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// Mock the response
	if strings.Contains(req.EmailText, "rejection") {
		return &ClassifyResponse{IsRejection: true}, nil
	}
	return &ClassifyResponse{IsRejection: false}, nil
}

func NewMockConnectionPool() (*ConnectionPool, error) {
	size := 1
	timeout := time.Second * 10

	// Create a buffer listener
	lis := bufconn.Listen(1024 * 1024)

	// Create a gRPC server
	grpcServer := grpc.NewServer()

	// Register the classifier service with the gRPC server
	RegisterClassifierServer(grpcServer, &mockClassifierServer{})

	// Start the gRPC server
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Create a connection pool with 1 gRPC connection object
	pool := make(chan *GRPCClient, size)

	// Connect to the gRPC server using the buffer listener
	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(func(ctx context.Context, address string) (net.Conn, error) {
		return lis.Dial()
	}), grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to dial bufnet: %v", err)
	}

	// Create a gRPC client object
	client := NewClassifierClient(conn)

	// Add the gRPC client object to the connection pool
	rc := &GRPCClient{
		Client: client,
	}
	pool <- rc

	return &ConnectionPool{
		Pool:    pool,
		Timeout: timeout,
	}, nil
}

func TestGetGrpcClient(t *testing.T) {
	// Create a new connection pool
	pool, err := NewMockConnectionPool()
	if err != nil {
		t.Fatalf("failed to create connection pool: %v", err)
	}

	// Get a gRPC client object from the pool
	client, err := pool.GetGRPCClient()
	if err != nil {
		t.Fatalf("failed to get gRPC client object from pool: %v", err)
	}
	assert.Equal(t, 0, len(pool.Pool))
	// Return the gRPC client object to the pool
	pool.ReturnGRPCClient(client)
	assert.Equal(t, 1, len(pool.Pool))
}
