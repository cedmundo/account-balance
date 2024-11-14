package services

import (
	"common/dao"
	"context"
	"database/sql/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestTransactionWorker_ValidateRecord(t *testing.T) {
	tests := []struct {
		name      string
		record    CSVRecord
		expectErr bool
	}{
		{
			name:      "Valid Record",
			record:    []string{"422202", "01/31", "+10.50"},
			expectErr: false,
		},
		{
			name:      "Invalid ID",
			record:    []string{"422AA2", "29/10", "+10.50"},
			expectErr: true,
		},
		{
			name:      "Invalid Date",
			record:    []string{"422202", "13-20", "+10.50"},
			expectErr: true,
		},
		{
			name:      "Invalid Operation type",
			record:    []string{"422202", "29/10", "~10.50"},
			expectErr: true,
		},
		{
			name:      "Invalid Amount",
			record:    []string{"422202", "29/10", "+1A.50"},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			worker := TransactionWorker{}
			date, op, amount, err := worker.ValidateRecord(tc.record)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.IsType(t, time.Time{}, date)
				assert.IsType(t, dao.TxOperationType(""), op)
				assert.IsType(t, decimal.Decimal{}, amount)
			}
		})
	}
}

func TestTransactionWorker_ParseRecord(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	accountID := int64(20)
	csvRow := []string{"10", "01/31", "+10.50"}
	args := []driver.Value{accountID, "credit", decimal.NewFromFloat(10.50), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()}
	mock.ExpectQuery(`INSERT INTO transactions`).WithArgs(args...).
		WillReturnRows(sqlmock.NewRows([]string{"transaction_id"}).AddRow(1))

	transactions := make(chan CSVRecord)
	reports := make(chan WorkerReport)

	worker := TransactionWorker{
		db:           db,
		ctx:          context.Background(),
		workerID:     0,
		accountID:    accountID,
		transactions: transactions,
		reports:      reports,
	}

	go worker.PullTransactions()
	transactions <- csvRow
	close(transactions)

	report := <-reports
	close(reports)

	expectedReport := WorkerReport{
		TotalDebit:       decimal.Zero,
		CountDebit:       0,
		TotalCredit:      decimal.NewFromFloat(10.50),
		CountCredit:      1,
		TransactionCount: map[int]int{1: 1},
		Errors:           0,
	}

	require.NoError(t, mock.ExpectationsWereMet())
	require.Equal(t, expectedReport.TotalDebit.StringFixed(2), report.TotalDebit.StringFixed(2))
	require.Equal(t, expectedReport.TotalCredit.StringFixed(2), report.TotalCredit.StringFixed(2))
	require.Equal(t, expectedReport.CountDebit, report.CountDebit)
	require.Equal(t, expectedReport.CountCredit, report.CountCredit)
	require.Equal(t, expectedReport.TransactionCount, report.TransactionCount)
	require.Equal(t, expectedReport.Errors, report.Errors)
}
