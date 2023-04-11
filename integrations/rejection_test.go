package integration

import (
	"context"
	"fmt"
	"testing"
)

func TestRejection(t *testing.T) {
	r, _ := NewRejectionCheck("localhost:50051")
	res := r.IsRejection(context.TODO(), "test")
	fmt.Printf("res: %v\n", res)
}
