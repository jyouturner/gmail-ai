package integration

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
)

type RejectionCheck struct {
	url        string
	connection *grpc.ClientConn
}

func NewRejectionCheck(url string) (*RejectionCheck, error) {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to dial gRPC server: %v", err)
	}

	return &RejectionCheck{
		url:        url,
		connection: conn,
	}, nil
}

func (c *RejectionCheck) IsRejection(ctx context.Context, text string) bool {
	client := NewClassifierClient(c.connection)

	response, err := client.ClassifyEmail(ctx, &ClassifyRequest{EmailText: text})
	if err != nil {
		log.Fatalf("Error calling ClassifyEmail: %v", err)
	}

	fmt.Printf("Is rejection email: %v\n", response.IsRejection)

	return response.IsRejection
}

func (c *RejectionCheck) Close() error {
	if c.connection != nil {
		return c.connection.Close()
	}
	return nil
}
