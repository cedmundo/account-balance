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
	AccountID        int64           `json:"account_id"`
	TotalCredit      decimal.Decimal `json:"total_credit"`
	CountCredit      int64           `json:"count_credit"`
	TotalDebit       decimal.Decimal `json:"total_debit"`
	CountDebit       int64           `json:"count_debit"`
	TotalBalance     decimal.Decimal `json:"total_balance"`
	AvgDebitAmount   decimal.Decimal `json:"avg_debit_amount"`
	AvgCreditAmount  decimal.Decimal `json:"avg_credit_amount"`
	TransactionCount map[int]int     `json:"transaction_count"`
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
	balanceReport := &BalanceReport{
		AccountID:        accountID,
		TotalCredit:      decimal.Zero,
		TotalDebit:       decimal.Zero,
		TotalBalance:     decimal.Zero,
		AvgCreditAmount:  decimal.Zero,
		AvgDebitAmount:   decimal.Zero,
		TransactionCount: make(map[int]int),
	}
	for workerReport := range reports {
		receivedReports += 1

		// add each result
		balanceReport.TotalDebit = balanceReport.TotalDebit.Add(workerReport.TotalDebit)
		balanceReport.TotalCredit = balanceReport.TotalCredit.Add(workerReport.TotalCredit)
		balanceReport.CountDebit += workerReport.CountDebit
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
	if balanceReport.CountCredit != 0 {
		balanceReport.AvgCreditAmount = balanceReport.TotalCredit.Div(decimal.NewFromInt(balanceReport.CountCredit))
	}
	if balanceReport.CountDebit != 0 {
		balanceReport.AvgDebitAmount = balanceReport.TotalDebit.Div(decimal.NewFromInt(balanceReport.CountDebit))
	}
	return *balanceReport, nil
}
