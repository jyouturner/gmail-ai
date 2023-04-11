package integration

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestRejection(t *testing.T) {
	// Create a connection pool with 10 RejectionCheck objects
	cp, err := NewConnectionPool("localhost:50051", 2, time.Second*10)
	if err != nil {
		t.Errorf("Error creating connection pool: %v", err)
	}
	defer cp.Close()

	rc, err := cp.GetRejectionCheck()

	defer cp.ReturnRejectionCheck(rc)
	if err != nil {
		t.Errorf("Error get rejection chjeck: %v", err)
	}
	isRejection := rc.IsRejection(context.Background(), "email text goes here")

	fmt.Printf("res: %v\n", isRejection)
}
