package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/cedmundo/account-balance/services"
	"github.com/jaswdr/faker/v2"
	"github.com/joho/godotenv"
	"log"
	"math/rand"
	"os"
	"time"
)

import _ "github.com/lib/pq"

const (
	sqlDriver = "postgres"
)

var (
	pFile             = flag.String("file", "", "File to read transactions from")
	pSeed             = flag.Int64("seed", 0, "Seed to use for random number generation")
	pWorkers          = flag.Int("workers", 5, "Number of workers to use when processing transactions")
	pBatchSize        = flag.Int("batch-size", 100, "Number of transactions to process at a time")
	pAccountEmail     = flag.String("account-email", "", "Account Email to use when creating accounts (leave blank to random)")
	pAccountFirstName = flag.String("account-first-name", "", "Account First name to use when creating accounts (leave blank to random)")
	pAccountLastName  = flag.String("account-last-name", "", "Account Last name to use when creating accounts (leave blank to random)")
	pDatabaseURL      = flag.String("database-url", "", "Database to use")
	fake              faker.Faker
)

func flagSeed() int64 {
	if *pSeed == 0 {
		return time.Now().Unix()
	}

	return *pSeed
}

func flagFile() *os.File {
	if *pFile == "" {
		log.Println("File not specified, doing nothing.")
		os.Exit(0)
	}

	file, err := os.Open(*pFile)
	if err != nil {
		log.Fatal("Could not open file:", err)
	}
	return file
}

func flagAccountEmail() string {
	if *pAccountEmail == "" {
		return "fake+" + fake.Internet().Email()
	}

	return *pAccountEmail
}

func flagAccountFirstName() string {
	if *pAccountFirstName == "" {
		return fake.Person().FirstName()
	}

	return *pAccountFirstName
}

func flagAccountLastName() string {
	if *pAccountLastName == "" {
		return fake.Person().LastName()
	}

	return *pAccountLastName
}

func flagDatabase() string {
	if *pDatabaseURL == "" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		host := os.Getenv("POSTGRES_HOST")
		if host == "" {
			host = "database" // from docker-compose
		}

		port := os.Getenv("POSTGRES_PORT")
		if port == "" {
			port = "5432"
		}

		user := os.Getenv("POSTGRES_USER")
		if user == "" {
			log.Fatal("POSTGRES_USER not set")
		}

		pass := os.Getenv("POSTGRES_PASSWORD")
		if pass == "" {
			log.Fatal("POSTGRES_PASSWORD not set")
		}

		db := os.Getenv("POSTGRES_DB")
		if db == "" {
			log.Fatal("POSTGRES_DB not set")
		}

		opts := os.Getenv("POSTGRES_OPTS")
		if opts == "" {
			opts = "?sslmode=disable" // from docker-compose
		}

		url := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s%s", user, pass, host, port, db, opts)
		log.Printf("Using database host: %s", host)
		return url
	}

	return *pDatabaseURL
}

func main() {
	flag.Parse()
	fake = faker.NewWithSeed(rand.NewSource(flagSeed()))

	file := flagFile()
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			log.Fatal("Could not close file:", err)
		}
	}(file)

	backgroundContext := context.Background()
	db, err := sql.Open(sqlDriver, flagDatabase())
	if err != nil {
		log.Fatal("Could not open database:", err)
	}

	accountService := services.AccountService{Database: db}
	transactionService := services.TransactionService{Database: db, Workers: *pWorkers, BatchSize: *pBatchSize}
	emailService := services.EmailService{}
	err = emailService.LoadMessages()
	if err != nil {
		log.Fatal("Could not load email messages:", err)
	}

	account, err := accountService.FetchOrCreateAccount(backgroundContext, flagAccountEmail(), flagAccountFirstName(), flagAccountLastName())
	if err != nil {
		log.Fatal("Could not fetch or create account:", err)
	}

	reader := csv.NewReader(file)
	report, err := transactionService.ProcessFile(backgroundContext, account.AccountID, reader)
	if err != nil {
		log.Fatal("Could not process transactions:", err)
	}

	err = emailService.SendReport(account, report)
	if err != nil {
		log.Fatal("Could not send report:", err)
	}
}
