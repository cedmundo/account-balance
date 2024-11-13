package lambda

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// CSVProcessRequest request a process of a CSV file within a S3 disk.
type CSVProcessRequest struct {
	ObjectKey        string `json:"object_key"`
	Bucket           string `json:"bucket"`
	Workers          int    `json:"workers"`
	BatchSize        int    `json:"batch_size"`
	AccountEmail     string `json:"account_email"`
	AccountFirstName string `json:"account_first_name"`
	AccountLastName  string `json:"account_last_name"`
}

var (
	s3Client *s3.Client
	dsn      string
)

func init() {
	// Initialize the S3 client outside the handler, during the init phase
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Load database environment variables
	dbName := os.Getenv("DATABASE_NAME")
	dbUser := os.Getenv("DATABASE_USER")
	dbHost := os.Getenv("DATABASE_HOST") // Add hostname without https
	dbPort := os.Getenv("DATABASE_PORT") // Add port number
	dbEndpoint := dbHost + ":" + dbPort
	region := os.Getenv("AWS_REGION")

	// Initialize S3 from config
	s3Client = s3.NewFromConfig(cfg)

	// Initialize DSN connection
	authenticationToken, err := auth.BuildAuthToken(context.TODO(), dbEndpoint, region, dbUser, cfg.Credentials)
	if err != nil {
		panic("failed to create authentication token: " + err.Error())
	}

	dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s?tls=true&allowCleartextPasswords=true",
		dbUser, authenticationToken, dbEndpoint, dbName,
	)
}

func getFile(ctx context.Context, bucket, key string) ([]byte, error) {
	obj, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}, nil)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	n, err := buf.ReadFrom(obj.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("Read %d bytes from S3 object", n)
	return buf.Bytes(), nil
}

func handleRequest(ctx context.Context, event json.RawMessage) (map[string]any, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}

	defer func(db *sql.DB) {
		if err := db.Close(); err != nil {
			log.Fatalf("Failed to close database connection: %v", err)
		}
	}(db)

	var req CSVProcessRequest
	if err := json.Unmarshal(event, &req); err != nil {
		log.Fatalf("Failed to unmarshal event: %v", err)
	}

	csvContent, err := getFile(ctx, req.Bucket, req.ObjectKey)
	if err != nil {
		log.Fatalf("Failed to get file from S3: %v", err)
	}

	reader := csv.NewReader(bytes.NewReader(csvContent))
	// TODO: skip header
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to parse CSV file: %v", err)
	}

	return nil, nil
}

func main() {
	lambda.Start(handleRequest)
}
