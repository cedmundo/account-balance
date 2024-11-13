package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
)

// var (
// s3Client *s3.Client
// )

// type Request struct {
// 	Name string `json:"name"`
// }

func init() {
	// Initialize the S3 client outside the handler, during the init phase
	// cfg, err := config.LoadDefaultConfig(context.TODO())
	// if err != nil {
	// 	log.Printf("unable to load SDK config, %v", err)
	// }

	// Initialize S3 from config
	// s3Client = s3.NewFromConfig(cfg)
}

func handleHelloWorld(ctx context.Context, event json.RawMessage) (map[string]any, error) {
	// var req Request
	// if err := json.Unmarshal(event, &req); err != nil {
	// 	return nil, err
	// }

	// buckets, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	// if err != nil {
	// 	return nil, err
	// }

	log.Printf("Event: %s!", string(event))
	// log.Printf("Buckets: %v", buckets)
	return map[string]any{
		"statusCode": 200,
		"headers": map[string]string{
			"Content-Type": "text/plain",
		},
		"body": "Hello World!",
	}, nil
}

func main() {
	lambda.Start(handleHelloWorld)
}
