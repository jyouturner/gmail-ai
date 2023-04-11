package integration

import (
	"context"
	"fmt"
	"log"
	sync "sync"
	"time"

	"github.com/jyouturer/gmail-ai/internal/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type RejectionCheck struct {
	client ClassifierClient
}

type ConnectionPool struct {
	pool      chan *RejectionCheck
	address   string
	size      int
	timeout   time.Duration
	mutex     sync.Mutex
	waitGroup sync.WaitGroup
}

func NewConnectionPool(address string, size int, timeout time.Duration) (*ConnectionPool, error) {
	pool := make(chan *RejectionCheck, size)

	for i := 0; i < size; i++ {
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			return nil, err
		}

		client := NewClassifierClient(conn)

		rc := &RejectionCheck{
			client: client,
		}
		logging.Logger.Info("Adding RejectionCheck object to pool", zap.String("address", address), zap.Int("size", size), zap.Int("number", i))
		pool <- rc
	}

	return &ConnectionPool{
		pool:    pool,
		address: address,
		size:    size,
		timeout: timeout,
	}, nil
}

func (cp *ConnectionPool) GetRejectionCheck() (*RejectionCheck, error) {
	logging.Logger.Info("Getting RejectionCheck object from pool")
	select {
	case rc := <-cp.pool:
		return rc, nil
	case <-time.After(cp.timeout):
		return nil, fmt.Errorf("timed out while waiting for RejectionCheck object")
	}
}

func (cp *ConnectionPool) ReturnRejectionCheck(rc *RejectionCheck) {
	logging.Logger.Info("Returning RejectionCheck object to pool")
	select {
	case cp.pool <- rc:
	default:
		// The pool is full, discard the RejectionCheck object
		cp.waitGroup.Add(1)
		go func() {
			defer cp.waitGroup.Done()
			rc.client = nil
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

func (rc *RejectionCheck) IsRejection(ctx context.Context, text string) bool {
	req := &ClassifyRequest{
		EmailText: text,
	}

	resp, err := rc.client.ClassifyEmail(ctx, req)
	if err != nil {
		log.Printf("failed to classify email: %v", err)
		return false
	}

	return resp.IsRejection
}
