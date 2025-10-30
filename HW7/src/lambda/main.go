package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type OrderMessage struct {
	OrderID   string `json:"orderId"`
	Timestamp string `json:"timestamp"`
	// Add other fields from your order
}

func handleSNSEvent(ctx context.Context, snsEvent events.SNSEvent) error {
	for _, record := range snsEvent.Records {
		snsRecord := record.SNS

		// Parse the message
		var order OrderMessage
		if err := json.Unmarshal([]byte(snsRecord.Message), &order); err != nil {
			fmt.Printf("Error parsing message: %v\n", err)
			return err
		}

		fmt.Printf("Processing order: %s\n", order.OrderID)

		// Simulate 3-second payment processing
		time.Sleep(3 * time.Second)

		fmt.Printf("Order %s processed successfully\n", order.OrderID)
	}

	return nil
}

func main() {
	lambda.Start(handleSNSEvent)
}
