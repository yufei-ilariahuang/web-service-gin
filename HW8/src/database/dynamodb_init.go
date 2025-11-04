package database

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// DDB is the DynamoDB client used by the DynamoDB store implementation.
var DDB *dynamodb.Client

// InitDynamoDB initialises the DynamoDB client. If endpoint is non-empty it will
// be used (useful for local DynamoDB testing with DynamoDB Local).
func InitDynamoDB(ctx context.Context, region, endpoint string) error {
	var err error
	var cfg aws.Config

	if region == "" {
		region = "us-west-2"
	}

	cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("failed to load aws config: %w", err)
	}

	if endpoint != "" {
		// Override endpoint for local testing
		cfg.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{URL: endpoint, SigningRegion: region}, nil
		})
	}

	DDB = dynamodb.NewFromConfig(cfg)
	return nil
}
