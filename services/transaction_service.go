package services

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"io"
)

// CSVRecord should contain transaction data from CSV
type CSVRecord []string

// TransactionService manages transactions and balances
type TransactionService struct {
	Database  *sql.DB
	Workers   int
	BatchSize int
}

// BalanceReport general info about the account
type BalanceReport struct {
	AccountID        int64
	TotalCredit      decimal.Decimal
	CountCredit      int64
	TotalDebit       decimal.Decimal
	CountDebit       int64
	TotalBalance     decimal.Decimal
	AvgDebitAmount   decimal.Decimal
	AvgCreditAmount  decimal.Decimal
	TransactionCount map[int]int
}

// ProcessFile start a work group and divides the calculation of transactions
func (s *TransactionService) ProcessFile(ctx context.Context, accountID int64, reader *csv.Reader) (BalanceReport, error) {
	reports := make(chan WorkerReport)
	transactions := make(chan CSVRecord, s.BatchSize)

	// Spin up the workers
	for i := 0; i < s.Workers; i++ {
		worker := TransactionWorker{
			db:           s.Database,
			ctx:          ctx,
			workerID:     i,
			accountID:    accountID,
			transactions: transactions,
			reports:      reports,
		}

		go worker.PullTransactions()
	}

	// Read all transactions from file
	_, err := reader.Read()
	if err != nil {
		return BalanceReport{}, fmt.Errorf("error reading header: %w", err)
	}
	for {
		line, err := reader.Read()
		if errors.Is(err, io.EOF) || errors.Is(err, csv.ErrFieldCount) {
			break
		} else if err != nil {
			return BalanceReport{}, err
		}
		transactions <- line
	}

	// Terminate processing
	close(transactions)

	// Wait for reports
	receivedReports := 0
	balanceReport := BalanceReport{
		AccountID:        accountID,
		TransactionCount: make(map[int]int),
	}
	for workerReport := range reports {
		receivedReports += 1

		// add each result
		balanceReport.TotalDebit.Add(workerReport.TotalDebit)
		balanceReport.CountDebit += workerReport.CountDebit
		balanceReport.TotalCredit.Add(workerReport.TotalCredit)
		balanceReport.CountCredit += workerReport.CountCredit

		// add transaction count for each month
		for month, count := range workerReport.TransactionCount {
			balanceReport.TransactionCount[month] += count
		}

		// interrupt after all workers have reported
		if receivedReports == s.Workers {
			close(reports)
			break
		}
	}

	balanceReport.TotalBalance = balanceReport.TotalCredit.Sub(balanceReport.TotalDebit)
	balanceReport.AvgCreditAmount = balanceReport.TotalCredit.Div(decimal.NewFromInt(balanceReport.CountCredit))
	balanceReport.AvgDebitAmount = balanceReport.TotalDebit.Div(decimal.NewFromInt(balanceReport.CountDebit)).Neg()
	return balanceReport, nil
}
