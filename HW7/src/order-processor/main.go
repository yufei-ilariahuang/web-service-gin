package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/labstack/echo/v4"
)

// Item represents an item in an order
type Item struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float32 `json:"price"`
}

// Order defines the complete order structure
type Order struct {
	OrderID    string    `json:"order_id"`
	CustomerID int       `json:"customer_id"`
	Status     string    `json:"status"`
	Items      []Item    `json:"items"`
	CreatedAt  time.Time `json:"created_at"`
}

// SNSMessage wraps the actual order message from SNS
type SNSMessage struct {
	Type             string `json:"Type"`
	MessageId        string `json:"MessageId"`
	TopicArn         string `json:"TopicArn"`
	Message          string `json:"Message"`
	Timestamp        string `json:"Timestamp"`
	SignatureVersion string `json:"SignatureVersion"`
	Signature        string `json:"Signature"`
	SigningCertURL   string `json:"SigningCertURL"`
	UnsubscribeURL   string `json:"UnsubscribeURL"`
}

var (
	sqsClient *sqs.SQS
	queueURL  string
)

func main() {
	// Initialize AWS SQS client
	initAWS()

	// Start worker goroutine
	go startWorker()

	// Health check endpoint for ECS
	e := echo.New()
	e.GET("/health", healthCheck)

	fmt.Println("🚀 Order Processor starting on :8080")
	fmt.Println("👷 Worker polling SQS queue for messages...")
	fmt.Println("💳 Payment processing: 3 seconds per order\n")

	e.Logger.Fatal(e.Start("0.0.0.0:8080"))
}

func initAWS() {
	queueURL = os.Getenv("SQS_QUEUE_URL")
	if queueURL == "" {
		log.Fatal("❌ SQS_QUEUE_URL environment variable not set")
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-west-2"
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		log.Fatalf("❌ Failed to create AWS session: %v", err)
	}

	sqsClient = sqs.New(sess)
	fmt.Printf("✅ AWS SQS client initialized\n")
	fmt.Printf("📥 Queue URL: %s\n\n", queueURL)
}

func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "order-processor",
	})
}

func startWorker() {
	log.Println("👷 Worker started, polling for messages...")

	for {
		// Long polling with 20 second wait time
		result, err := sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(queueURL),
			MaxNumberOfMessages: aws.Int64(1),
			WaitTimeSeconds:     aws.Int64(20), // Long polling
			VisibilityTimeout:   aws.Int64(30),
		})

		if err != nil {
			log.Printf("❌ Error receiving messages: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Process each message
		for _, message := range result.Messages {
			processMessage(message)
		}
	}
}

func processMessage(message *sqs.Message) {
	startTime := time.Now()

	// Parse SNS message wrapper
	var snsMsg SNSMessage
	if err := json.Unmarshal([]byte(*message.Body), &snsMsg); err != nil {
		log.Printf("❌ Error parsing SNS message: %v", err)
		return
	}

	// Parse order from SNS message
	var order Order
	if err := json.Unmarshal([]byte(snsMsg.Message), &order); err != nil {
		log.Printf("❌ Error parsing order from message: %v", err)
		return
	}

	log.Printf("📦 Processing order: %s for customer: %d", order.OrderID, order.CustomerID)

	// Simulate payment processing (3 seconds)
	time.Sleep(3 * time.Second)

	// Update order status
	order.Status = "completed"

	processingTime := time.Since(startTime)
	log.Printf("✅ Order %s completed successfully in %v", order.OrderID, processingTime.Round(time.Millisecond))

	// Delete message from queue after successful processing
	_, err := sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: message.ReceiptHandle,
	})

	if err != nil {
		log.Printf("❌ Error deleting message: %v", err)
	} else {
		log.Printf("🗑️  Message deleted from queue")
	}
}
