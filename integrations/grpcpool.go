package integration

import (
	"fmt"
	sync "sync"
	"time"

	"github.com/jyouturer/gmail-ai/internal/logging"
	"go.uber.org/zap"
	grpc "google.golang.org/grpc"
)

type GRPCClient struct {
	Client ClassifierClient
}

type ConnectionPool struct {
	pool      chan *GRPCClient
	address   string
	size      int
	timeout   time.Duration
	mutex     sync.Mutex
	waitGroup sync.WaitGroup
}

func NewConnectionPool(address string, size int, timeout time.Duration) (*ConnectionPool, error) {
	pool := make(chan *GRPCClient, size)

	for i := 0; i < size; i++ {
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			return nil, err
		}

		client := NewClassifierClient(conn)

		rc := &GRPCClient{
			Client: client,
		}
		logging.Logger.Info("Adding GRPCClient object to pool", zap.String("address", address), zap.Int("size", size), zap.Int("number", i))
		pool <- rc
	}

	return &ConnectionPool{
		pool:    pool,
		address: address,
		size:    size,
		timeout: timeout,
	}, nil
}

func (cp *ConnectionPool) GetGRPCClient() (*GRPCClient, error) {
	logging.Logger.Info("Getting GRPCClient object from pool")
	select {
	case rc := <-cp.pool:
		return rc, nil
	case <-time.After(cp.timeout):
		return nil, fmt.Errorf("timed out while waiting for GRPCClient object")
	}
}

func (cp *ConnectionPool) ReturnGRPCClient(rc *GRPCClient) {
	logging.Logger.Info("Returning GRPCClient object to pool")
	select {
	case cp.pool <- rc:
	default:
		// The pool is full, discard the GRPCClient object
		cp.waitGroup.Add(1)
		go func() {
			defer cp.waitGroup.Done()
			rc.Client = nil
		}()
	}
}

func (cp *ConnectionPool) Close() {
	logging.Logger.Info("Closing connection pool")
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	close(cp.pool)

	// Wait for all connections to be returned and closed
	cp.waitGroup.Wait()
}
