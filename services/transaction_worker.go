package services

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/cedmundo/account-balance/dao"
	"github.com/shopspring/decimal"
	"log"
	"regexp"
	"time"
)

const (
	dayMonthLayout = "01/02"
)

var (
	rTxID                 = regexp.MustCompile(`^\d+$`)
	rTxDate               = regexp.MustCompile(`^\d{2}/\d{2}$`)
	rTxOperationAndAmount = regexp.MustCompile(`^[+|-]\d+(\.\d+)?$`)
)

// WorkerReport holds sums for each worker
type WorkerReport struct {
	TotalDebit       decimal.Decimal
	CountDebit       int64
	TotalCredit      decimal.Decimal
	CountCredit      int64
	TransactionCount map[int]int
	Errors           int
}

// TransactionWorker processes CSV records as a work group.
type TransactionWorker struct {
	db           *sql.DB
	ctx          context.Context
	workerID     int
	accountID    int64
	transactions chan CSVRecord
	reports      chan WorkerReport
}

// ValidateRecord validates a CSV record and extracts transaction date, operation type, and amount.
// Returns an error if any of the fields are invalid.
func (w *TransactionWorker) ValidateRecord(record CSVRecord) (txDate time.Time, txOperation dao.TxOperationType, txAmount decimal.Decimal, err error) {
	if !rTxID.MatchString(record[0]) {
		err = fmt.Errorf("invalid transaction id: %s", record[0])
		return
	}

	if !rTxDate.MatchString(record[1]) {
		err = fmt.Errorf("invalid transaction date: %s", record[1])
		return
	}

	txDate, err = time.Parse(dayMonthLayout, record[1])
	if err != nil {
		return
	}
	txDate = txDate.AddDate(time.Now().Year(), 0, 0)

	if !rTxOperationAndAmount.MatchString(record[2]) {
		err = fmt.Errorf("invalid transaction operation and amount: %s", record[2])
		return
	}

	operationAndAmount := record[2]
	txOperation = dao.TxOperationTypeCredit
	if operationAndAmount[0:1] == "-" {
		txOperation = dao.TxOperationTypeDebit
	}

	txAmount, err = decimal.NewFromString(operationAndAmount[1:])
	return
}

// PullTransactions processes transaction records from the worker's transaction channel and inserts them into the database.
func (w *TransactionWorker) PullTransactions() {
	queries := dao.New(w.db)
	now := time.Now()
	inserted := 0
	report := WorkerReport{
		TotalDebit:       decimal.Zero,
		TotalCredit:      decimal.Zero,
		TransactionCount: make(map[int]int),
	}
	for transaction := range w.transactions {
		performedAt, operation, amount, err := w.ValidateRecord(transaction)
		if err != nil {
			log.Printf("worker %d: error validating transaction: %s", w.workerID, err)
			report.Errors += 1
			continue
		}

		// Perform basic calculations
		month := performedAt.Month()
		report.TransactionCount[int(month)] += 1
		if operation == dao.TxOperationTypeDebit {
			report.TotalDebit = report.TotalDebit.Add(amount)
			report.CountDebit += 1
		} else {
			report.TotalCredit = report.TotalCredit.Add(amount)
			report.CountCredit += 1
		}

		// Insert transaction into database
		_, err = queries.InsertTransaction(w.ctx, dao.InsertTransactionParams{
			AccountID:   w.accountID,
			Operation:   operation,
			Amount:      amount,
			PerformedAt: performedAt,
			CreatedAt:   sql.NullTime{Valid: true, Time: now},
			UpdatedAt:   sql.NullTime{Valid: true, Time: now},
		})
		if err != nil {
			log.Printf("worker %d: error inserting transaction: %s", w.workerID, err)
			report.Errors += 1
		}

		inserted += 1
	}

	log.Printf("worker %d: inserted %d transactions in %s with %d errors", w.workerID, inserted, time.Since(now), report.Errors)
	w.reports <- report
}
