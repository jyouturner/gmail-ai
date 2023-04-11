package integration

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
)

type RejectionCheck struct {
	url string
}

func NewRejectionCheck(url string) *RejectionCheck {
	return &RejectionCheck{
		url: url,
	}
}

func (c *RejectionCheck) IsRejection(ctx context.Context, text string) bool {
	conn, err := grpc.Dial(c.url, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := NewClassifierClient(conn)

	response, err := client.ClassifyEmail(ctx, &ClassifyRequest{EmailText: text})
	if err != nil {
		log.Fatalf("Error calling ClassifyEmail: %v", err)
	}

	fmt.Printf("Is rejection email: %v\n", response.IsRejection)

	return response.IsRejection
}
